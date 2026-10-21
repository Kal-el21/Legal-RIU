package handler

import (
	"bytes"
	"errors"
	"mime/multipart"
	"path/filepath"
	"strconv"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LegalCaseHandler struct {
	svc         service.LegalCaseService
	auditLogSvc service.AuditLogService
	userRepo    repository.UserRepository
}

func NewLegalCaseHandler(svc service.LegalCaseService, auditLogSvc service.AuditLogService, userRepo repository.UserRepository) *LegalCaseHandler {
	return &LegalCaseHandler{svc: svc, auditLogSvc: auditLogSvc, userRepo: userRepo}
}

// companyScope mengembalikan perusahaan yang membatasi si pemanggil, atau nil
// kalau ia tidak dibatasi (ia ADMIN, perusahaannya internal, atau ia tidak
// terikat perusahaan). Kalau data usernya tidak bisa dibaca, permintaan
// ditolak; tidak dianggap "tidak dibatasi".
func (h *LegalCaseHandler) companyScope(c *gin.Context) (*uuid.UUID, bool) {
	return companyScope(c, h.userRepo)
}

// Dipakai juga oleh Repositori Dokumen, yang menampilkan dokumen perkara.
func companyScope(c *gin.Context, userRepo repository.UserRepository) (*uuid.UUID, bool) {
	user, err := userRepo.FindByID(mustParseUUID(middleware.GetUserID(c)))
	if err != nil {
		utils.Forbidden(c, "Data pengguna tidak dapat dibaca")
		return nil, false
	}
	// ADMIN mengelola seluruh sistem, di perusahaan mana pun ia terdaftar.
	if user.Role == entity.RoleAdmin || user.CompanyID == nil || user.Company.IsInternal {
		return nil, true
	}
	return user.CompanyID, true
}

func (h *LegalCaseHandler) GetAll(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	var query dto.LegalCaseListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, "Query tidak valid", err.Error())
		return
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}

	items, total, err := h.svc.GetAll(scope, query)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	utils.OK(c, "Success", gin.H{
		"items":       items,
		"total":       total,
		"page":        query.Page,
		"limit":       query.Limit,
		"total_pages": (total + int64(query.Limit) - 1) / int64(query.Limit),
	})
}

func (h *LegalCaseHandler) GetLatest(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	legalCase, err := h.svc.GetLatest(scope)
	if err != nil {
		utils.OK(c, "Success", nil)
		return
	}
	utils.OK(c, "Success", legalCase)
}

func (h *LegalCaseHandler) GetByID(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	legalCase, err := h.svc.GetByID(scope, c.Param("id"))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", legalCase)
}

func (h *LegalCaseHandler) Create(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	var req dto.CreateLegalCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	legalCase, err := h.svc.Create(scope, req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_case", legalCase.ID)
	c.Set("audit_description", "Legal case created")
	utils.Created(c, "Kasus hukum berhasil dibuat", legalCase)
}

func (h *LegalCaseHandler) Update(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	var req dto.UpdateLegalCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	legalCase, err := h.svc.Update(scope, c.Param("id"), req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_case", legalCase.ID)
	c.Set("audit_description", "Legal case updated")
	utils.OK(c, "Kasus hukum berhasil diupdate", legalCase)
}

func (h *LegalCaseHandler) Delete(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	id := c.Param("id")
	if err := h.svc.Delete(scope, id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionDelete, "legal_case", id)
	utils.OK(c, "Kasus hukum berhasil dihapus", nil)
}

func (h *LegalCaseHandler) UploadDocument(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	id := c.Param("id")

	file, err := c.FormFile("document")
	if err != nil {
		utils.BadRequest(c, "Dokumen wajib diupload", err.Error())
		return
	}

	legalCase, err := h.svc.UploadDocument(scope, id, file)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionFileUpload, "legal_case", id)
	c.Set("audit_description", "Legal case document uploaded")
	utils.OK(c, "Dokumen berhasil diupload", legalCase)
}

func (h *LegalCaseHandler) UploadPhoto(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	id := c.Param("id")

	file, err := c.FormFile("photo")
	if err != nil {
		utils.BadRequest(c, "Foto wajib diupload", err.Error())
		return
	}

	legalCase, err := h.svc.UploadPhoto(scope, id, file)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionFileUpload, "legal_case", id)
	c.Set("audit_description", "Legal case photo uploaded")
	utils.OK(c, "Foto berhasil diupload", legalCase)
}

func (h *LegalCaseHandler) DeletePhoto(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	id := c.Param("id")

	legalCase, err := h.svc.DeletePhoto(scope, id)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionFileDelete, "legal_case", id)
	c.Set("audit_description", "Legal case photo deleted")
	utils.OK(c, "Foto berhasil dihapus", legalCase)
}

func (h *LegalCaseHandler) ImportChronologies(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	id := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "File Excel wajib diupload", err.Error())
		return
	}

	result, err := h.svc.ImportChronologies(scope, id, file)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionFileUpload, "legal_case", id)
	c.Set("audit_description", "Legal case chronologies imported")
	utils.OK(c, "Impor kronologi selesai", result)
}

func (h *LegalCaseHandler) DownloadChronologyTemplate(c *gin.Context) {
	buf, err := h.svc.GenerateChronologyTemplate()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	c.DataFromReader(-1, -1, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf, map[string]string{
		"Content-Disposition": `attachment; filename="chronology-template.xlsx"`,
	})
}

func (h *LegalCaseHandler) ViewFile(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	filePath := c.Query("path")
	if filePath == "" {
		utils.BadRequest(c, "Path file diperlukan", nil)
		return
	}

	obj, contentType, err := h.svc.ViewFile(scope, filePath)
	if errors.Is(err, service.ErrFileNotFound) {
		utils.NotFound(c, "File tidak ditemukan")
		return
	}
	if err != nil {
		utils.InternalError(c, "Gagal mengambil file")
		return
	}
	defer obj.Close()

	c.DataFromReader(-1, -1, contentType, obj, map[string]string{
		"Content-Disposition": utils.HeaderUnduhan("inline", filepath.Base(filePath)),
	})
}

func (h *LegalCaseHandler) DeleteDocument(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	id := c.Param("id")

	legalCase, err := h.svc.DeleteDocument(scope, id)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionFileDelete, "legal_case", id)
	c.Set("audit_description", "Legal case document deleted")
	utils.OK(c, "Dokumen berhasil dihapus", legalCase)
}

func (h *LegalCaseHandler) ListChronologies(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	items, err := h.svc.ListChronologies(scope, c.Param("id"))
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Success", items)
}

func (h *LegalCaseHandler) CreateChronology(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	req, files, ok := bindChronologyRequest(c)
	if !ok {
		return
	}

	chronology, err := h.svc.CreateChronology(scope, c.Param("id"), req, files)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_case", c.Param("id"))
	c.Set("audit_description", "Legal case chronology added")
	utils.Created(c, "Kronologi kasus berhasil ditambahkan", chronology)
}

func (h *LegalCaseHandler) UpdateChronology(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	req, files, ok := bindChronologyRequest(c)
	if !ok {
		return
	}

	chronology, err := h.svc.UpdateChronology(scope, c.Param("id"), c.Param("chronId"), req, files)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_case", c.Param("id"))
	c.Set("audit_description", "Legal case chronology updated")
	utils.OK(c, "Kronologi kasus berhasil diupdate", chronology)
}

func (h *LegalCaseHandler) DeleteChronology(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteChronology(scope, c.Param("id"), c.Param("chronId")); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionDelete, "legal_case", c.Param("id"))
	c.Set("audit_description", "Legal case chronology deleted")
	utils.OK(c, "Kronologi kasus berhasil dihapus", nil)
}

func (h *LegalCaseHandler) Download(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	filePath := c.Query("path")
	if filePath == "" {
		utils.BadRequest(c, "Path file diperlukan", nil)
		return
	}

	obj, err := h.svc.DownloadFile(scope, filePath)
	if errors.Is(err, service.ErrFileNotFound) {
		utils.NotFound(c, "File tidak ditemukan")
		return
	}
	if err != nil {
		utils.InternalError(c, "Gagal mengambil file")
		return
	}
	defer obj.Close()

	c.DataFromReader(-1, -1, "application/octet-stream", obj, map[string]string{
		"Content-Disposition": utils.HeaderUnduhan("attachment", filepath.Base(filePath)),
	})
}

func (h *LegalCaseHandler) ListRegencies(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "500"))
	items, err := h.svc.ListRegencies(c.Query("search"), limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", items)
}

func (h *LegalCaseHandler) ListCedants(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	items, err := h.svc.ListCedants(c.Query("search"), limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", items)
}

func (h *LegalCaseHandler) CreateRegency(c *gin.Context) {
	var req dto.CreateRegencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	regency, err := h.svc.CreateRegency(req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "regency", regency.ID)
	c.Set("audit_description", "Regency created")
	utils.Created(c, "Kabupaten/Kota berhasil dibuat", regency)
}

func (h *LegalCaseHandler) UpdateRegency(c *gin.Context) {
	var req dto.UpdateRegencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	regency, err := h.svc.UpdateRegency(c.Param("id"), req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "regency", regency.ID)
	c.Set("audit_description", "Regency updated")
	utils.OK(c, "Kabupaten/Kota berhasil diupdate", regency)
}

func (h *LegalCaseHandler) DeleteRegency(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteRegency(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionDelete, "regency", id)
	c.Set("audit_description", "Regency deleted")
	utils.OK(c, "Kabupaten/Kota berhasil dihapus", nil)
}

func (h *LegalCaseHandler) CreateCedant(c *gin.Context) {
	var req dto.CreateCedantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	cedant, err := h.svc.CreateCedant(req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "cedant", cedant.ID)
	c.Set("audit_description", "Cedant created")
	utils.Created(c, "Cedant berhasil dibuat", cedant)
}

func (h *LegalCaseHandler) UpdateCedant(c *gin.Context) {
	var req dto.UpdateCedantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	cedant, err := h.svc.UpdateCedant(c.Param("id"), req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "cedant", cedant.ID)
	c.Set("audit_description", "Cedant updated")
	utils.OK(c, "Cedant berhasil diupdate", cedant)
}

func (h *LegalCaseHandler) DeleteCedant(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteCedant(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionDelete, "cedant", id)
	c.Set("audit_description", "Cedant deleted")
	utils.OK(c, "Cedant berhasil dihapus", nil)
}

func (h *LegalCaseHandler) ImportCedants(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "File Excel wajib diupload", err.Error())
		return
	}
	result, err := h.svc.ImportCedants(file)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionFileUpload, "cedant", "import")
	c.Set("audit_description", "Cedants imported")
	utils.OK(c, "Impor cedant selesai", result)
}

func (h *LegalCaseHandler) DownloadCedantTemplate(c *gin.Context) {
	buf, err := h.svc.GenerateCedantTemplate()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	c.DataFromReader(-1, -1, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf, map[string]string{
		"Content-Disposition": `attachment; filename="cedant-template.xlsx"`,
	})
}

func (h *LegalCaseHandler) ImportRegencies(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "File Excel wajib diupload", err.Error())
		return
	}
	result, err := h.svc.ImportRegencies(file)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionFileUpload, "regency", "import")
	c.Set("audit_description", "Regencies imported")
	utils.OK(c, "Impor regensi selesai", result)
}

func (h *LegalCaseHandler) DownloadRegencyTemplate(c *gin.Context) {
	buf, err := h.svc.GenerateRegencyTemplate()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	c.DataFromReader(-1, -1, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf, map[string]string{
		"Content-Disposition": `attachment; filename="regency-template.xlsx"`,
	})
}

func (h *LegalCaseHandler) UpdateStatus(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	id := c.Param("id")
	var body struct {
		Status string `json:"current_status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	// Perkara yang tidak ada dan perkara perusahaan lain dijawab sama.
	if _, err := h.svc.GetByID(scope, id); err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	if err := h.svc.UpdateStatus(scope, id, body.Status); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionStatusChange, "legal_case", id)
	c.Set("audit_description", "Legal case status updated")
	utils.OK(c, "Status kasus berhasil diupdate", nil)
}

func bindChronologyRequest(c *gin.Context) (dto.CreateCaseChronologyRequest, []*multipart.FileHeader, bool) {
	if err := c.Request.ParseMultipartForm(110 << 20); err != nil {
		var req dto.CreateCaseChronologyRequest
		if jsonErr := c.ShouldBindJSON(&req); jsonErr != nil {
			utils.BadRequest(c, "Gagal memparse request", err.Error())
			return req, nil, false
		}
		return req, nil, true
	}

	req := dto.CreateCaseChronologyRequest{
		AgendaDate:  c.PostForm("agenda_date"),
		Agenda:      c.PostForm("agenda"),
		Description: c.PostForm("description"),
		Documents:   append(c.PostFormArray("documents"), c.PostFormArray("document_paths")...),
	}

	var files []*multipart.FileHeader
	if form := c.Request.MultipartForm; form != nil {
		files = append(files, form.File["documents"]...)
		files = append(files, form.File["files"]...)
	}

	if req.AgendaDate == "" || req.Agenda == "" {
		utils.BadRequest(c, "Tanggal dan agenda wajib diisi", nil)
		return req, files, false
	}

	return req, files, true
}

func mustParseUUID(value string) uuid.UUID {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return parsed
}

func (h *LegalCaseHandler) GeneratePDF(c *gin.Context) {
	scope, ok := h.companyScope(c)
	if !ok {
		return
	}
	id := c.Param("id")
	// Dicek dulu supaya perkara yang tidak ada atau milik perusahaan lain
	// dijawab 404, bukan 500.
	lc, err := h.svc.GetByID(scope, id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	pdfData, err := h.svc.GeneratePDF(scope, id)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	c.DataFromReader(-1, -1, "application/pdf", bytes.NewReader(pdfData), map[string]string{
		"Content-Disposition": utils.HeaderUnduhan("attachment", "legal-case-"+lc.ID+".pdf"),
	})
}
