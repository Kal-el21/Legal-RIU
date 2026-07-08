package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LegalMaterialHandler struct {
	svc service.LegalMaterialService
}

func NewLegalMaterialHandler(svc service.LegalMaterialService) *LegalMaterialHandler {
	return &LegalMaterialHandler{svc: svc}
}

func toLegalMaterialResponse(material *entity.LegalMaterial) dto.LegalMaterialResponse {
	return dto.LegalMaterialResponse{
		ID:        material.ID.String(),
		Title:     material.Title,
		Excerpt:   material.Excerpt,
		Content:   material.Content,
		CreatedBy: material.CreatedBy.String(),
		UpdatedBy: material.UpdatedBy.String(),
		CreatedAt: material.CreatedAt,
		UpdatedAt: material.UpdatedAt,
	}
}

func (h *LegalMaterialHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	resp := make([]dto.LegalMaterialResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toLegalMaterialResponse(&items[i]))
	}
	utils.OK(c, "Success", resp)
}

func (h *LegalMaterialHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.svc.GetByID(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", toLegalMaterialResponse(item))
}

func (h *LegalMaterialHandler) Create(c *gin.Context) {
	var req dto.CreateLegalMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	createdBy, err := uuid.Parse(userID)
	if err != nil {
		utils.BadRequest(c, "User tidak valid", nil)
		return
	}

	item, err := h.svc.Create(req.Title, req.Excerpt, req.Content, createdBy)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_material", item.ID)
	c.Set("audit_description", "Legal material created")
	utils.Created(c, "Materi berhasil dibuat", toLegalMaterialResponse(item))
}

func (h *LegalMaterialHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateLegalMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	updatedBy, err := uuid.Parse(userID)
	if err != nil {
		utils.BadRequest(c, "User tidak valid", nil)
		return
	}

	item, err := h.svc.Update(id, req.Title, req.Excerpt, req.Content, updatedBy)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionUserUpdate, "legal_material", item.ID)
	c.Set("audit_description", "Legal material updated")
	utils.OK(c, "Materi berhasil diupdate", toLegalMaterialResponse(item))
}

func (h *LegalMaterialHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "legal_material", id)
	c.Set("audit_description", "Legal material deleted")
	utils.OK(c, "Materi berhasil dihapus", nil)
}

func (h *LegalMaterialHandler) ImportLegalMaterials(c *gin.Context) {
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
	middleware.SetAuditContext(c, entity.ActionFileUpload, "legal_material", "import")
	c.Set("audit_description", "Legal materials imported")
	utils.OK(c, "Impor materi selesai", result)
}

func (h *LegalMaterialHandler) DownloadLegalMaterialTemplate(c *gin.Context) {
	buf, err := h.svc.GenerateImportTemplate()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	c.DataFromReader(-1, -1, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf, map[string]string{
		"Content-Disposition": `attachment; filename="legal-material-template.xlsx"`,
	})
}
