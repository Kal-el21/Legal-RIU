package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type CaseTypeHandler struct {
	svc service.CaseTypeService
}

func NewCaseTypeHandler(svc service.CaseTypeService) *CaseTypeHandler {
	return &CaseTypeHandler{svc: svc}
}

func toCaseTypeResponse(ct *entity.CaseType) dto.CaseTypeResponse {
	return dto.CaseTypeResponse{
		ID:        ct.ID.String(),
		Code:      ct.Code,
		Label:     ct.Label,
		IsActive:  ct.IsActive,
		CreatedAt: ct.CreatedAt,
		UpdatedAt: ct.UpdatedAt,
	}
}

func (h *CaseTypeHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	resp := make([]dto.CaseTypeResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toCaseTypeResponse(&items[i]))
	}
	utils.OK(c, "Success", resp)
}

func (h *CaseTypeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.svc.GetByID(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", toCaseTypeResponse(item))
}

func (h *CaseTypeHandler) Create(c *gin.Context) {
	var req struct {
		Code     string `json:"code" binding:"required"`
		Label    string `json:"label" binding:"required"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Create(req.Code, req.Label)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "case_type", item.ID)
	c.Set("audit_description", "Case type created")
	utils.Created(c, "Jenis kasus berhasil dibuat", toCaseTypeResponse(item))
}

func (h *CaseTypeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Code     string `json:"code" binding:"required"`
		Label    string `json:"label" binding:"required"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.svc.Update(id, req.Code, req.Label, req.IsActive)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "case_type", item.ID)
	c.Set("audit_description", "Case type updated")
	utils.OK(c, "Jenis kasus berhasil diupdate", toCaseTypeResponse(item))
}

func (h *CaseTypeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "case_type", id)
	c.Set("audit_description", "Case type deleted")
	utils.OK(c, "Jenis kasus berhasil dihapus", nil)
}
