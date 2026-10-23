package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditLogHandler struct {
	auditLogSvc  service.AuditLogService
	auditLogRepo repository.AuditLogRepository
}

func NewAuditLogHandler(auditLogSvc service.AuditLogService, auditLogRepo repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{auditLogSvc: auditLogSvc, auditLogRepo: auditLogRepo}
}

// wib: tanggal di saringan Audit Log dibaca sebagai tanggal WIB, bukan UTC. Zona tetap,
// supaya tidak bergantung pada data zona waktu di image container.
var wib = time.FixedZone("WIB", 7*60*60)

func toPelakuAudit(u *entity.User) *dto.PelakuAudit {
	if u == nil || u.ID == uuid.Nil {
		return nil
	}
	return &dto.PelakuAudit{ID: u.ID.String(), FullName: u.FullName, Email: u.Email, Role: string(u.Role)}
}

func toAuditLogResponse(l *entity.AuditLog) dto.AuditLogResponse {
	return dto.AuditLogResponse{
		ID: l.ID.String(), UserID: l.UserID.String(), User: toPelakuAudit(&l.User),
		Action: string(l.Action), EntityType: l.EntityType, EntityID: l.EntityID.String(),
		OldValue: l.OldValue, NewValue: l.NewValue, Description: l.Description,
		IPAddress: l.IPAddress, UserAgent: l.UserAgent, CreatedAt: l.CreatedAt,
	}
}

// denganCatatan menambahkan catatan approver ke keterangan audit dengan format yang dipisah
// halaman Audit Log (pisahCatatan di frontend/src/lib/audit.ts).
func denganCatatan(catatan string) string {
	if c := strings.TrimSpace(catatan); c != "" {
		return ". Catatan: " + c
	}
	return ""
}

// keteranganImpor: keterangan audit impor Excel dengan jumlah baris yang masuk dan dilewati
// (PR Audit Log tahap L.3).
func keteranganImpor(jenis string, hasil *dto.ImportResult) string {
	if hasil == nil {
		return "Impor " + jenis
	}
	return fmt.Sprintf("Impor %s: %d baris masuk, %d dilewati", jenis, hasil.Imported, hasil.Skipped)
}

// uuidQuery membaca saringan ID dari query; kosong berarti tidak disaring.
func uuidQuery(c *gin.Context, nama string) (*uuid.UUID, bool) {
	v := c.Query(nama)
	if v == "" {
		return nil, true
	}
	id, err := uuid.Parse(v)
	if err != nil {
		utils.BadRequest(c, nama+" tidak valid", nil)
		return nil, false
	}
	return &id, true
}

// GET /api/v1/admin/audit-logs
func (h *AuditLogHandler) GetAll(c *gin.Context) {
	var filters dto.AuditLogFilters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	// Nilai yang tidak masuk akal (termasuk bukan angka) diganti nilai bawaan;
	// limit dipakai membagi, dan dibatasi 100 supaya satu permintaan tidak memuat semua baris.
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	filters.Page = page
	filters.Limit = limit

	if action := c.Query("action"); action != "" {
		filters.Action = &action
	}
	if entityType := c.Query("entity_type"); entityType != "" {
		filters.EntityType = &entityType
	}
	var ok bool
	if filters.UserID, ok = uuidQuery(c, "user_id"); !ok {
		return
	}
	if filters.EntityID, ok = uuidQuery(c, "entity_id"); !ok {
		return
	}
	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.ParseInLocation("2006-01-02", dateFrom, wib); err == nil {
			filters.DateFrom = &t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.ParseInLocation("2006-01-02", dateTo, wib); err == nil {
			dateToEnd := t.Add(24*time.Hour - time.Nanosecond)
			filters.DateTo = &dateToEnd
		}
	}

	items, total, err := h.auditLogRepo.GetAll(repository.AuditLogFilters{
		Page:       filters.Page,
		Limit:      filters.Limit,
		Action:     filters.Action,
		EntityType: filters.EntityType,
		EntityID:   filters.EntityID,
		UserID:     filters.UserID,
		DateFrom:   filters.DateFrom,
		DateTo:     filters.DateTo,
		Search:     filters.Search,
	})
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	hasil := make([]dto.AuditLogResponse, 0, len(items))
	for i := range items {
		hasil = append(hasil, toAuditLogResponse(&items[i]))
	}
	response := map[string]interface{}{
		"items":       hasil,
		"total":       total,
		"page":        filters.Page,
		"limit":       filters.Limit,
		"total_pages": totalPages,
	}

	utils.OK(c, "Success", response)
}

// batasRiwayat: baris terbanyak di riwayat satu data.
const batasRiwayat = 200

// Riwayat: GET /<modul>/:id/history, riwayat aktivitas satu data, terbaru di atas (PR Audit Log
// tahap L.4). bolehBuka memeriksa akses dengan aturan halaman detail modulnya dan sudah menjawab
// kalau tidak boleh, jadi siapa pun yang tidak boleh membuka datanya mendapat jawaban yang sama
// dengan detail (keputusan L3). IP dan user agent hanya untuk yang punya audit_log.view (L7).
func (h *AuditLogHandler) Riwayat(jenis string, bolehBuka func(*gin.Context) (uuid.UUID, bool)) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := bolehBuka(c)
		if !ok {
			return
		}
		items, err := h.auditLogRepo.GetByEntity(jenis, id, batasRiwayat)
		if err != nil {
			utils.InternalError(c, "gagal memuat riwayat")
			return
		}
		auditor := middleware.HasPermission(c, "audit_log.view")
		hasil := make([]dto.AuditLogResponse, 0, len(items))
		for i := range items {
			r := toAuditLogResponse(&items[i])
			if !auditor {
				r.IPAddress, r.UserAgent = "", ""
			}
			hasil = append(hasil, r)
		}
		utils.OK(c, "Success", hasil)
	}
}

// GET /api/v1/admin/audit-logs/actors?q= dan /legal/...: pelaku untuk saringan Audit Log,
// maks. 10, termasuk akun yang sudah dihapus (PR Audit Log tahap L.1).
func (h *AuditLogHandler) GetActors(c *gin.Context) {
	users, err := h.auditLogRepo.Pelaku(c.Query("q"), 10)
	if err != nil {
		utils.InternalError(c, "gagal memuat daftar pelaku")
		return
	}
	hasil := make([]*dto.PelakuAudit, 0, len(users))
	for i := range users {
		hasil = append(hasil, toPelakuAudit(&users[i]))
	}
	utils.OK(c, "Success", hasil)
}
