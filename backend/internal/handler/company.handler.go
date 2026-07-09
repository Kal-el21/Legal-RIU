package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	svc service.CompanyService
}

func NewCompanyHandler(svc service.CompanyService) *CompanyHandler {
	return &CompanyHandler{svc: svc}
}

func toCompanyResponse(c *entity.Company) dto.CompanyResponse {
	return dto.CompanyResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		EmailDomain: c.EmailDomain,
		IsInternal:  c.IsInternal,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func (h *CompanyHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	resp := make([]dto.CompanyResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toCompanyResponse(&items[i]))
	}
	utils.OK(c, "Success", resp)
}

func (h *CompanyHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.svc.GetByID(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", toCompanyResponse(item))
}

func (h *CompanyHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		EmailDomain string `json:"email_domain" binding:"required"`
		IsInternal  bool   `json:"is_internal"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Create(req.Name, req.EmailDomain, req.IsInternal)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "company", item.ID)
	c.Set("audit_description", "Company created")
	utils.Created(c, "Perusahaan berhasil dibuat", toCompanyResponse(item))
}

func (h *CompanyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name" binding:"required"`
		EmailDomain string `json:"email_domain" binding:"required"`
		IsInternal  bool   `json:"is_internal"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Update(id, req.Name, req.EmailDomain, req.IsInternal)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "company", item.ID)
	c.Set("audit_description", "Company updated")
	utils.OK(c, "Perusahaan berhasil diupdate", toCompanyResponse(item))
}

func (h *CompanyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "company", id)
	c.Set("audit_description", "Company deleted")
	utils.OK(c, "Perusahaan berhasil dihapus", nil)
}

func (h *CompanyHandler) ImportCompanies(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "File Excel wajib diupload", err.Error())
		return
	}
	result, err := h.svc.ImportFromExcel(file)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionFileUpload, "company", "import")
	c.Set("audit_description", "Companies imported")
	utils.OK(c, "Impor perusahaan selesai", result)
}

func (h *CompanyHandler) DownloadCompanyTemplate(c *gin.Context) {
	buf, err := h.svc.GenerateImportTemplate()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	c.DataFromReader(-1, -1, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf, map[string]string{
		"Content-Disposition": `attachment; filename="company-template.xlsx"`,
	})
}
