package handler

import (
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

type DocumentReviewHandler struct {
	svc         service.DocumentReviewService
	auditLogSvc service.AuditLogService
}

func NewDocumentReviewHandler(svc service.DocumentReviewService, auditLogSvc service.AuditLogService) *DocumentReviewHandler {
	return &DocumentReviewHandler{svc: svc, auditLogSvc: auditLogSvc}
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
	if err := c.Request.ParseMultipartForm(110 << 20); err != nil {
		var req dto.CreateDocumentReviewRequest
		if jsonErr := c.ShouldBindJSON(&req); jsonErr != nil {
			utils.BadRequest(c, "Gagal memparse request", err.Error())
			return
		}
		userID := middleware.GetUserID(c)
		dr, svcErr := h.svc.Create(userID, req, nil)
		if svcErr != nil {
			utils.InternalError(c, svcErr.Error())
			return
		}
		middleware.SetAuditContext(c, entity.ActionUserUpdate, "document_review", dr.ID.String())
		c.Set("audit_description", "Document review created")
		utils.Created(c, "Pengajuan berhasil dibuat", dr)
		return
	}

	req := dto.CreateDocumentReviewRequest{
		RequestorName:     c.PostForm("requestor_name"),
		RequestorPosition: c.PostForm("requestor_position"),
		RequestorDivision: c.PostForm("requestor_division"),
		RequestorEmail:    c.PostForm("requestor_email"),
		RequestorPhone:    c.PostForm("requestor_phone"),
		DocumentName:      c.PostForm("document_name"),
		SecondParty:       c.PostForm("second_party"),
		ThirdParty:        c.PostForm("third_party"),
		DocumentType:      c.PostForm("document_type"),
		DocumentTypeOther: c.PostForm("document_type_other"),
		AdditionalNote:    c.PostForm("additional_note"),
	}

	if req.RequestorName == "" || req.RequestorPosition == "" || req.RequestorDivision == "" ||
		req.RequestorEmail == "" || req.RequestorPhone == "" || req.DocumentName == "" ||
		req.SecondParty == "" || req.DocumentType == "" {
		utils.BadRequest(c, "Semua field wajib diisi", nil)
		return
	}

	var files []*multipart.FileHeader
	if form := c.Request.MultipartForm; form != nil {
		files = form.File["attachments"]
	}

	userID := middleware.GetUserID(c)
	dr, err := h.svc.Create(userID, req, files)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "document_review", dr.ID.String())
	desc := "Document review created"
	if len(files) > 0 {
		desc = "Document review created with " + strconv.Itoa(len(files)) + " file(s)"
	}
	c.Set("audit_description", desc)
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
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "document_review", id)
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
	middleware.SetAuditContext(c, entity.ActionDelete, "document_review", id)
	utils.OK(c, "Pengajuan berhasil dihapus", nil)
}

// POST /api/v1/review-documents/:id/resubmit
func (h *DocumentReviewHandler) Resubmit(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	var files []*multipart.FileHeader
	if form, _ := c.MultipartForm(); form != nil {
		files = form.File["attachments"]
	}
	dr, err := h.svc.Resubmit(id, userID, files)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "document_review", id)
	desc := "Document review resubmitted"
	if len(files) > 0 {
		desc = "Document review resubmitted with " + strconv.Itoa(len(files)) + " file(s)"
	}
	c.Set("audit_description", desc)
	utils.OK(c, "Pengajuan berhasil diajukan ulang", dr)
}

// GET /api/v1/review-documents/presign?path=xxx
func (h *DocumentReviewHandler) GetPresignedURL(c *gin.Context) {
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

// GET /api/v1/review-documents/download?path=xxx
func (h *DocumentReviewHandler) Download(c *gin.Context) {
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

	// Extract filename from path
	parts := strings.Split(filePath, "/")
	fileName := parts[len(parts)-1]

	c.DataFromReader(-1, -1, "application/octet-stream", obj, map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, fileName),
	})
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
	desc := "Status changed to " + req.Status
	middleware.SetAuditContextWithValues(c, entity.ActionStatusChange, "document_review", id, nil, &req.Status, &desc)
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
	middleware.SetAuditContext(c, entity.ActionFileUpload, "document_review", id)
	utils.OK(c, "Hasil review berhasil diupload", nil)
}
