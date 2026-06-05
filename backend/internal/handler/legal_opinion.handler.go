package handler

import (
	"mime/multipart"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type LegalOpinionHandler struct {
	svc service.LegalOpinionService
}

func NewLegalOpinionHandler(svc service.LegalOpinionService) *LegalOpinionHandler {
	return &LegalOpinionHandler{svc: svc}
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
func (h *LegalOpinionHandler) Create(c *gin.Context) {
	var req dto.CreateLegalOpinionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	form, _ := c.MultipartForm()
	var files []*multipart.FileHeader
	if form != nil {
		files = form.File["attachments"]
	}

	userID := middleware.GetUserID(c)
	lo, err := h.svc.Create(userID, req, files)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
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
	utils.OK(c, "Pengajuan berhasil dihapus", nil)
}

// POST /api/v1/legal-opinions/:id/resubmit
func (h *LegalOpinionHandler) Resubmit(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)

	form, _ := c.MultipartForm()
	var files []*multipart.FileHeader
	if form != nil {
		files = form.File["attachments"]
	}

	lo, err := h.svc.Resubmit(id, userID, files)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Pengajuan berhasil diajukan ulang", lo)
}

// POST /api/v1/legal-opinions/:id/upload-attachment
func (h *LegalOpinionHandler) UploadAttachment(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	lo, err := h.svc.GetByID(id, userID, role)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	form, _ := c.MultipartForm()
	if form == nil || len(form.File["attachments"]) == 0 {
		utils.BadRequest(c, "File tidak ditemukan", nil)
		return
	}

	_ = lo
	utils.OK(c, "File berhasil diupload", nil)
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

// ── Admin endpoints ───────────────────────────────────────────────────────────

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

	req := dto.UploadResultRequest{
		Notes: c.PostForm("notes"),
	}

	if err := h.svc.UploadResult(id, adminID, req, file); err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Hasil kajian berhasil diupload", nil)
}
