package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type DocumentTypeHandler struct {
	svc service.DocumentTypeService
}

func NewDocumentTypeHandler(svc service.DocumentTypeService) *DocumentTypeHandler {
	return &DocumentTypeHandler{svc: svc}
}

func toDocumentTypeResponse(dt *entity.DocumentType) dto.DocumentTypeResponse {
	return dto.DocumentTypeResponse{
		ID:        dt.ID.String(),
		Name:      dt.Name,
		Label:     dt.Label,
		IsActive:  dt.IsActive,
		CreatedAt: dt.CreatedAt,
		UpdatedAt: dt.UpdatedAt,
	}
}

func (h *DocumentTypeHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	resp := make([]dto.DocumentTypeResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toDocumentTypeResponse(&items[i]))
	}
	utils.OK(c, "Success", resp)
}

func (h *DocumentTypeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.svc.GetByID(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", toDocumentTypeResponse(item))
}

func (h *DocumentTypeHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Label string `json:"label" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Create(req.Name, req.Label)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "document_type", item.ID)
	c.Set("audit_description", "Document type created")
	utils.Created(c, "Jenis dokumen berhasil dibuat", toDocumentTypeResponse(item))
}

func (h *DocumentTypeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name     string `json:"name" binding:"required"`
		Label    string `json:"label" binding:"required"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Update(id, req.Name, req.Label, req.IsActive)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "document_type", item.ID)
	c.Set("audit_description", "Document type updated")
	utils.OK(c, "Jenis dokumen berhasil diupdate", toDocumentTypeResponse(item))
}

func (h *DocumentTypeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "document_type", id)
	c.Set("audit_description", "Document type deleted")
	utils.OK(c, "Jenis dokumen berhasil dihapus", nil)
}