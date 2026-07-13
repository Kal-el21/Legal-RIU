package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	tmplerrors "legal-riu-portal/internal/errors"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type CompanyMasterHandler struct {
	svc         service.CompanyMasterService
	auditLogSvc service.AuditLogService
}

func NewCompanyMasterHandler(svc service.CompanyMasterService, auditLogSvc service.AuditLogService) *CompanyMasterHandler {
	return &CompanyMasterHandler{svc: svc, auditLogSvc: auditLogSvc}
}

func (h *CompanyMasterHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", items)
}

func (h *CompanyMasterHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.svc.GetByID(id)
	if err != nil {
		utils.NotFound(c, "Data tidak ditemukan")
		return
	}
	utils.OK(c, "Success", item)
}

func (h *CompanyMasterHandler) Create(c *gin.Context) {
	var req dto.CreateCompanyMasterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Create(entity.CompanyMaster{
		Name:             req.Name,
		Address:          req.Address,
		NPWP:             req.NPWP,
		Phone:            req.Phone,
		Email:            req.Email,
		DefaultPejabat:   req.DefaultPejabat,
		DefaultJabatan:   req.DefaultJabatan,
		DefaultTempatTtd: req.DefaultTempatTtd,
		IsActive:         req.IsActive,
	})
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserCreate, "company_master", item.ID.String())
	utils.Created(c, "Data berhasil dibuat", item)
}

func (h *CompanyMasterHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.CreateCompanyMasterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Update(id, entity.CompanyMaster{
		Name:             req.Name,
		Address:          req.Address,
		NPWP:             req.NPWP,
		Phone:            req.Phone,
		Email:            req.Email,
		DefaultPejabat:   req.DefaultPejabat,
		DefaultJabatan:   req.DefaultJabatan,
		DefaultTempatTtd: req.DefaultTempatTtd,
		IsActive:         req.IsActive,
	})
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "company_master", id)
	utils.OK(c, "Data berhasil diupdate", item)
}

func (h *CompanyMasterHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "company_master", id)
	utils.OK(c, "Data berhasil dihapus", nil)
}

func (h *CompanyMasterHandler) UploadTemplate(c *gin.Context) {
	version := c.PostForm("version")
	if version == "" {
		version = "1"
	}

	file, err := c.FormFile("template")
	if err != nil {
		utils.BadRequest(c, "File template wajib diupload", nil)
		return
	}

	if !isValidDocxFile(file.Filename) {
		utils.BadRequest(c, "File template harus berformat .docx", nil)
		return
	}

	f, err := file.Open()
	if err != nil {
		utils.InternalError(c, "Gagal membuka file")
		return
	}
	defer f.Close()

	buf := make([]byte, file.Size)
	if _, err := f.Read(buf); err != nil {
		utils.InternalError(c, "Gagal membaca file")
		return
	}

	tmpl, err := h.svc.UploadTemplate(c.Request.Context(), version, buf)
	if err != nil {
		if errors.Is(err, tmplerrors.ErrInvalidVersion) {
			utils.BadRequest(c, "Versi template tidak valid", nil)
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserCreate, "company_master_template", version)
	utils.Created(c, "Template berhasil diupload", tmpl)
}

func (h *CompanyMasterHandler) GetActiveTemplate(c *gin.Context) {
	tmpl, err := h.svc.GetActiveTemplate()
	if err != nil {
		utils.NotFound(c, "Template belum diupload")
		return
	}
	utils.OK(c, "Success", tmpl)
}

func (h *CompanyMasterHandler) GetTemplateByVersion(c *gin.Context) {
	version := c.Param("version")
	tmpl, err := h.svc.GetTemplate(version)
	if err != nil {
		utils.NotFound(c, "Template tidak ditemukan")
		return
	}
	utils.OK(c, "Success", tmpl)
}

func (h *CompanyMasterHandler) DeleteTemplate(c *gin.Context) {
	version := c.Param("version")
	if err := h.svc.DeleteTemplate(c.Request.Context(), version); err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "company_master_template", version)
	utils.OK(c, "Template berhasil dihapus", nil)
}

// GET /api/v1/admin/templates/:version/field-positions
func (h *CompanyMasterHandler) GetFieldPositions(c *gin.Context) {
	version := c.Param("version")
	positions, err := h.svc.GetFieldPositions(version)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", positions)
}

// PUT /api/v1/admin/templates/:version/field-positions
func (h *CompanyMasterHandler) SaveFieldPositions(c *gin.Context) {
	version := c.Param("version")
	var positions []entity.TemplateFieldPosition
	if err := c.ShouldBindJSON(&positions); err != nil {
		utils.BadRequest(c, "Invalid payload", err.Error())
		return
	}

	if err := h.svc.SaveFieldPositions(version, positions); err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Field positions saved", nil)
}

// GET /api/v1/admin/templates/:version/preview?page=N
// Returns the base PDF page N of the template version as a PNG so the
// calibration UI can overlay clickable field markers on top of it.
func (h *CompanyMasterHandler) GetTemplatePreview(c *gin.Context) {
	version := c.Param("version")
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	data, err := h.svc.GetTemplateBaseImage(version, page)
	if err != nil {
		log.Printf("[PREVIEW][v%s] Failed to generate preview for page %d: %v", version, page, err)
		utils.InternalError(c, fmt.Sprintf("Gagal memuat preview: %v", err))
		return
	}
	c.Data(http.StatusOK, "image/png", data)
}

func isValidDocxFile(filename string) bool {
	lastDot := -1
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			lastDot = i
			break
		}
	}
	if lastDot == -1 {
		return false
	}
	switch filename[lastDot:] {
	case ".docx", ".DOCX", ".doc", ".DOC":
		return true
	}
	return false
}
