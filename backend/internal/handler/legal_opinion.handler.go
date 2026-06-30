package handler

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type LegalOpinionHandler struct {
	svc         service.LegalOpinionService
	auditLogSvc service.AuditLogService
}

func NewLegalOpinionHandler(svc service.LegalOpinionService, auditLogSvc service.AuditLogService) *LegalOpinionHandler {
	return &LegalOpinionHandler{svc: svc, auditLogSvc: auditLogSvc}
}

// GET /api/v1/legal-opinions
func (h *LegalOpinionHandler) GetAll(c *gin.Context) {
	var query dto.LegalOpinionListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, "Query tidak valid", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	items, total, err := h.svc.GetAll(userID, role, query)
	if err != nil {
		utils.InternalError(c, err.Error())
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

// GET /api/v1/legal-opinions/:id
func (h *LegalOpinionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	lo, err := h.svc.GetByID(id, userID, role)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", lo)
}

// POST /api/v1/legal-opinions
// Frontend mengirim multipart/form-data (fields + optional files)
func (h *LegalOpinionHandler) Create(c *gin.Context) {
	// Parse multipart form (max 110MB sudah diset di main)
	if err := c.Request.ParseMultipartForm(110 << 20); err != nil {
		// fallback: coba parse sebagai JSON biasa (tanpa file)
		var req dto.CreateLegalOpinionRequest
		if jsonErr := c.ShouldBindJSON(&req); jsonErr != nil {
			utils.BadRequest(c, "Gagal memparse request", err.Error())
			return
		}
		userID := middleware.GetUserID(c)
		lo, svcErr := h.svc.Create(userID, req, nil)
		if svcErr != nil {
			utils.InternalError(c, svcErr.Error())
			return
		}
		utils.Created(c, "Pengajuan berhasil dibuat", lo)
		middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_opinion", lo.ID.String())
		c.Set("audit_description", "Legal opinion created")
		return
	}

	// Bind dari form fields
	req := dto.CreateLegalOpinionRequest{
		RequestorName:     c.PostForm("requestor_name"),
		RequestorPosition: c.PostForm("requestor_position"),
		RequestorDivision: c.PostForm("requestor_division"),
		RequestorEmail:    c.PostForm("requestor_email"),
		RequestorPhone:    c.PostForm("requestor_phone"),
		LegalType:         c.PostForm("legal_type"),
		LegalTypeOther:    c.PostForm("legal_type_other"),
		Title:             c.PostForm("title"),
		Chronology:        c.PostForm("chronology"),
		Question:          c.PostForm("question"),
	}

	// Manual validation
	if req.RequestorName == "" || req.RequestorPosition == "" || req.RequestorDivision == "" ||
		req.RequestorEmail == "" || req.RequestorPhone == "" || req.LegalType == "" ||
		req.Title == "" || req.Chronology == "" || req.Question == "" {
		utils.BadRequest(c, "Semua field wajib diisi", nil)
		return
	}

	// Get files
	var files []*multipart.FileHeader
	if form := c.Request.MultipartForm; form != nil {
		files = form.File["attachments"]
	}

	userID := middleware.GetUserID(c)
	lo, err := h.svc.Create(userID, req, files)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_opinion", lo.ID.String())
	desc := "Legal opinion created"
	if len(files) > 0 {
		desc = "Legal opinion created with " + strconv.Itoa(len(files)) + " file(s)"
	}
	c.Set("audit_description", desc)
	utils.Created(c, "Pengajuan berhasil dibuat", lo)
}

// PUT /api/v1/legal-opinions/:id
func (h *LegalOpinionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateLegalOpinionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	lo, err := h.svc.Update(id, userID, req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_opinion", id)
	utils.OK(c, "Pengajuan berhasil diupdate", lo)
}

// DELETE /api/v1/legal-opinions/:id
func (h *LegalOpinionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	if err := h.svc.Delete(id, userID); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "legal_opinion", id)
	utils.OK(c, "Pengajuan berhasil dihapus", nil)
}

// POST /api/v1/legal-opinions/:id/resubmit
func (h *LegalOpinionHandler) Resubmit(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	var files []*multipart.FileHeader
	if form, _ := c.MultipartForm(); form != nil {
		files = form.File["attachments"]
	}
	lo, err := h.svc.Resubmit(id, userID, files)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_opinion", id)
	desc := "Legal opinion resubmitted"
	if len(files) > 0 {
		desc = "Legal opinion resubmitted with " + strconv.Itoa(len(files)) + " file(s)"
	}
	c.Set("audit_description", desc)
	utils.OK(c, "Pengajuan berhasil diajukan ulang", lo)
}

// GET /api/v1/legal-opinions/presign?path=xxx
func (h *LegalOpinionHandler) GetPresignedURL(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		utils.BadRequest(c, "Path file diperlukan", nil)
		return
	}
	url, err := h.svc.GetPresignedURL(filePath)
	if err != nil {
		utils.InternalError(c, "Gagal membuat URL")
		return
	}
	utils.OK(c, "Success", gin.H{"url": url})
}

// GET /api/v1/legal-opinions/download?path=xxx
func (h *LegalOpinionHandler) Download(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		utils.BadRequest(c, "Path file diperlukan", nil)
		return
	}

	obj, err := h.svc.DownloadFile(filePath)
	if err != nil {
		utils.InternalError(c, "Gagal mengambil file")
		return
	}
	defer obj.Close()

	parts := strings.Split(filePath, "/")
	fileName := parts[len(parts)-1]

	c.DataFromReader(-1, -1, "application/octet-stream", obj, map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, fileName),
	})
}

// PATCH /api/v1/admin/legal-opinions/:id/status
func (h *LegalOpinionHandler) AdminUpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	if err := h.svc.UpdateStatus(id, req); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	desc := "Status changed to " + req.Status
	middleware.SetAuditContextWithValues(c, entity.ActionStatusChange, "legal_opinion", id, nil, &req.Status, &desc)
	utils.OK(c, "Status berhasil diubah", nil)
}

// POST /api/v1/admin/legal-opinions/:id/result
func (h *LegalOpinionHandler) AdminUploadResult(c *gin.Context) {
	id := c.Param("id")
	adminID := middleware.GetUserID(c)
	file, err := c.FormFile("result")
	if err != nil {
		utils.BadRequest(c, "File hasil kajian diperlukan", nil)
		return
	}
	req := dto.UploadResultRequest{Notes: c.PostForm("notes")}
	if err := h.svc.UploadResult(id, adminID, req, file); err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	middleware.SetAuditContext(c, entity.ActionFileUpload, "legal_opinion", id)
	utils.OK(c, "Hasil kajian berhasil diupload", nil)
}

// GET /api/v1/legal-opinions/:id/pdf
func (h *LegalOpinionHandler) GeneratePDF(c *gin.Context) {
	id := c.Param("id")
	pdfData, err := h.svc.GeneratePDF(id)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	lo, _ := h.svc.GetByID(id, middleware.GetUserID(c), middleware.GetUserRole(c))

	c.DataFromReader(-1, -1, "application/pdf", bytes.NewReader(pdfData), map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="legal-opinion-%s.pdf"`, lo.TicketNumber),
	})
}


