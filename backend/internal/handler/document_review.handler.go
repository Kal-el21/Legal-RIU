package handler

import (
	"mime/multipart"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type DocumentReviewHandler struct {
	svc service.DocumentReviewService
}

func NewDocumentReviewHandler(svc service.DocumentReviewService) *DocumentReviewHandler {
	return &DocumentReviewHandler{svc: svc}
}

// GET /api/v1/review-documents
func (h *DocumentReviewHandler) GetAll(c *gin.Context) {
	var query dto.DocumentReviewListQuery
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

// GET /api/v1/review-documents/:id
func (h *DocumentReviewHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	dr, err := h.svc.GetByID(id, userID, role)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", dr)
}

// POST /api/v1/review-documents
func (h *DocumentReviewHandler) Create(c *gin.Context) {
	_, _ = c.MultipartForm()

	var req dto.CreateDocumentReviewRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	form, _ := c.MultipartForm()
	var files []*multipart.FileHeader
	if form != nil {
		files = form.File["attachments"]
	}
	userID := middleware.GetUserID(c)
	dr, err := h.svc.Create(userID, req, files)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.Created(c, "Pengajuan berhasil dibuat", dr)
}

// PUT /api/v1/review-documents/:id
func (h *DocumentReviewHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateDocumentReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	dr, err := h.svc.Update(id, userID, req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Pengajuan berhasil diupdate", dr)
}

// DELETE /api/v1/review-documents/:id
func (h *DocumentReviewHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	if err := h.svc.Delete(id, userID); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Pengajuan berhasil dihapus", nil)
}

// POST /api/v1/review-documents/:id/resubmit
func (h *DocumentReviewHandler) Resubmit(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	form, _ := c.MultipartForm()
	var files []*multipart.FileHeader
	if form != nil {
		files = form.File["attachments"]
	}
	dr, err := h.svc.Resubmit(id, userID, files)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Pengajuan berhasil diajukan ulang", dr)
}

// GET /api/v1/review-documents/presign?path=xxx
func (h *DocumentReviewHandler) GetPresignedURL(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		utils.BadRequest(c, "Path file diperlukan", nil)
		return
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	url, err := h.svc.GetPresignedURL(filePath, userID, role)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", gin.H{"url": url})
}

// PATCH /api/v1/admin/review-documents/:id/status
func (h *DocumentReviewHandler) AdminUpdateStatus(c *gin.Context) {
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

// POST /api/v1/admin/review-documents/:id/result
func (h *DocumentReviewHandler) AdminUploadResult(c *gin.Context) {
	id := c.Param("id")
	adminID := middleware.GetUserID(c)
	file, err := c.FormFile("result")
	if err != nil {
		utils.BadRequest(c, "File hasil review diperlukan", nil)
		return
	}
	req := dto.UploadResultRequest{Notes: c.PostForm("notes")}
	if err := h.svc.UploadResult(id, adminID, req, file); err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Hasil review berhasil diupload", nil)
}
