package handler

import (
	"bytes"
	"fmt"
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

func (h *LegalCaseHandler) getCompanyID(c *gin.Context) *uuid.UUID {
	userID := middleware.GetUserID(c)
	user, err := h.userRepo.FindByID(mustParseUUID(userID))
	if err != nil {
		return nil
	}
	if user.CompanyID == nil {
		return nil
	}
	if user.Company.IsInternal {
		return nil
	}
	return user.CompanyID
}

func (h *LegalCaseHandler) GetAll(c *gin.Context) {
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

	companyID := h.getCompanyID(c)
	items, total, err := h.svc.GetAll(companyID, query)
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
	companyID := h.getCompanyID(c)
	legalCase, err := h.svc.GetLatest(companyID)
	if err != nil {
		utils.OK(c, "Success", nil)
		return
	}
	utils.OK(c, "Success", legalCase)
}

func (h *LegalCaseHandler) GetByID(c *gin.Context) {
	companyID := h.getCompanyID(c)
	legalCase, err := h.svc.GetByID(companyID, c.Param("id"))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", legalCase)
}

func (h *LegalCaseHandler) Create(c *gin.Context) {
	var req dto.CreateLegalCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	user, err := h.userRepo.FindByID(mustParseUUID(userID))
	if err != nil {
		utils.BadRequest(c, "User tidak ditemukan", nil)
		return
	}
	if user.CompanyID != nil && user.Company.IsInternal {
		req.CompanyID = user.CompanyID.String()
	}

	legalCase, err := h.svc.Create(user.CompanyID, req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_case", legalCase.ID)
	c.Set("audit_description", "Legal case created")
	utils.Created(c, "Kasus hukum berhasil dibuat", legalCase)
}

func (h *LegalCaseHandler) Update(c *gin.Context) {
	var req dto.UpdateLegalCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	legalCase, err := h.svc.Update(c.Param("id"), req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_case", legalCase.ID)
	c.Set("audit_description", "Legal case updated")
	utils.OK(c, "Kasus hukum berhasil diupdate", legalCase)
}

func (h *LegalCaseHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionDelete, "legal_case", id)
	utils.OK(c, "Kasus hukum berhasil dihapus", nil)
}

func (h *LegalCaseHandler) UploadDocument(c *gin.Context) {
	id := c.Param("id")

	file, err := c.FormFile("document")
	if err != nil {
		utils.BadRequest(c, "Dokumen wajib diupload", err.Error())
		return
	}

	legalCase, err := h.svc.UploadDocument(id, file)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionFileUpload, "legal_case", id)
	c.Set("audit_description", "Legal case document uploaded")
	utils.OK(c, "Dokumen berhasil diupload", legalCase)
}

func (h *LegalCaseHandler) UploadPhoto(c *gin.Context) {
	id := c.Param("id")

	file, err := c.FormFile("photo")
	if err != nil {
		utils.BadRequest(c, "Foto wajib diupload", err.Error())
		return
	}

	legalCase, err := h.svc.UploadPhoto(id, file)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionFileUpload, "legal_case", id)
	c.Set("audit_description", "Legal case photo uploaded")
	utils.OK(c, "Foto berhasil diupload", legalCase)
}

func (h *LegalCaseHandler) DeletePhoto(c *gin.Context) {
	id := c.Param("id")

	legalCase, err := h.svc.DeletePhoto(id)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionFileDelete, "legal_case", id)
	c.Set("audit_description", "Legal case photo deleted")
	utils.OK(c, "Foto berhasil dihapus", legalCase)
}

func (h *LegalCaseHandler) ImportChronologies(c *gin.Context) {
	id := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "File Excel wajib diupload", err.Error())
		return
	}

	result, err := h.svc.ImportChronologies(id, file)
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
	filePath := c.Query("path")
	if filePath == "" {
		utils.BadRequest(c, "Path file diperlukan", nil)
		return
	}

	obj, contentType, err := h.svc.ViewFile(filePath)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	defer obj.Close()

	c.DataFromReader(-1, -1, contentType, obj, map[string]string{
		"Content-Disposition": fmt.Sprintf(`inline; filename="%s"`, filepath.Base(filePath)),
	})
}

func (h *LegalCaseHandler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")

	legalCase, err := h.svc.DeleteDocument(id)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionFileDelete, "legal_case", id)
	c.Set("audit_description", "Legal case document deleted")
	utils.OK(c, "Dokumen berhasil dihapus", legalCase)
}

func (h *LegalCaseHandler) ListChronologies(c *gin.Context) {
	items, err := h.svc.ListChronologies(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Success", items)
}

func (h *LegalCaseHandler) CreateChronology(c *gin.Context) {
	req, files, ok := bindChronologyRequest(c)
	if !ok {
		return
	}

	chronology, err := h.svc.CreateChronology(c.Param("id"), req, files)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_case", c.Param("id"))
	c.Set("audit_description", "Legal case chronology added")
	utils.Created(c, "Kronologi kasus berhasil ditambahkan", chronology)
}

func (h *LegalCaseHandler) UpdateChronology(c *gin.Context) {
	req, files, ok := bindChronologyRequest(c)
	if !ok {
		return
	}

	chronology, err := h.svc.UpdateChronology(c.Param("id"), c.Param("chronId"), req, files)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_case", c.Param("id"))
	c.Set("audit_description", "Legal case chronology updated")
	utils.OK(c, "Kronologi kasus berhasil diupdate", chronology)
}

func (h *LegalCaseHandler) DeleteChronology(c *gin.Context) {
	if err := h.svc.DeleteChronology(c.Param("id"), c.Param("chronId")); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionDelete, "legal_case", c.Param("id"))
	c.Set("audit_description", "Legal case chronology deleted")
	utils.OK(c, "Kronologi kasus berhasil dihapus", nil)
}

func (h *LegalCaseHandler) Download(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		utils.BadRequest(c, "Path file diperlukan", nil)
		return
	}

	obj, err := h.svc.DownloadFile(filePath)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	defer obj.Close()

	c.DataFromReader(-1, -1, "application/octet-stream", obj, map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filePath)),
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

func (h *LegalCaseHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"current_status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	companyID := h.getCompanyID(c)
	legalCase, err := h.svc.GetByID(companyID, id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	if companyID != nil && legalCase.CompanyID != companyID.String() {
		utils.Forbidden(c, "Anda tidak memiliki akses ke kasus ini")
		return
	}

	if err := h.svc.UpdateStatus(id, body.Status); err != nil {
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
	id := c.Param("id")
	pdfData, err := h.svc.GeneratePDF(id)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	lc, _ := h.svc.GetByID(h.getCompanyID(c), id)

	c.DataFromReader(-1, -1, "application/pdf", bytes.NewReader(pdfData), map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="legal-case-%s.pdf"`, lc.ID),
	})
}
