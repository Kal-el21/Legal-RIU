package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type PurposeTypeHandler struct {
	svc service.PurposeTypeService
}

func NewPurposeTypeHandler(svc service.PurposeTypeService) *PurposeTypeHandler {
	return &PurposeTypeHandler{svc: svc}
}

func toPurposeTypeResponse(pt *entity.PurposeType) dto.PurposeTypeResponse {
	return dto.PurposeTypeResponse{
		ID:          pt.ID.String(),
		Name:        pt.Name,
		Description: pt.Description,
		IsActive:    pt.IsActive,
		CreatedAt:   pt.CreatedAt,
		UpdatedAt:   pt.UpdatedAt,
	}
}

func (h *PurposeTypeHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	resp := make([]dto.PurposeTypeResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toPurposeTypeResponse(&items[i]))
	}
	utils.OK(c, "Success", resp)
}

func (h *PurposeTypeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.svc.GetByID(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", toPurposeTypeResponse(item))
}

func (h *PurposeTypeHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Create(req.Name, req.Description)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "purpose_type", item.ID)
	c.Set("audit_description", "Purpose type created")
	utils.Created(c, "Tujuan pembuatan berhasil dibuat", toPurposeTypeResponse(item))
}

func (h *PurposeTypeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Update(id, req.Name, req.Description, req.IsActive)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "purpose_type", item.ID)
	c.Set("audit_description", "Purpose type updated")
	utils.OK(c, "Tujuan pembuatan berhasil diupdate", toPurposeTypeResponse(item))
}

func (h *PurposeTypeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "purpose_type", id)
	c.Set("audit_description", "Purpose type deleted")
	utils.OK(c, "Tujuan pembuatan berhasil dihapus", nil)
}
