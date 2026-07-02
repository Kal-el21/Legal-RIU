package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type CaseCategoryHandler struct {
	svc service.CaseCategoryService
}

func NewCaseCategoryHandler(svc service.CaseCategoryService) *CaseCategoryHandler {
	return &CaseCategoryHandler{svc: svc}
}

func toCaseCategoryResponse(cc *entity.CaseCategory) dto.CaseCategoryResponse {
	return dto.CaseCategoryResponse{
		ID:        cc.ID.String(),
		Code:      cc.Code,
		Label:     cc.Label,
		IsActive:  cc.IsActive,
		CreatedAt: cc.CreatedAt,
		UpdatedAt: cc.UpdatedAt,
	}
}

func (h *CaseCategoryHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	resp := make([]dto.CaseCategoryResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toCaseCategoryResponse(&items[i]))
	}
	utils.OK(c, "Success", resp)
}

func (h *CaseCategoryHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.svc.GetByID(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", toCaseCategoryResponse(item))
}

func (h *CaseCategoryHandler) Create(c *gin.Context) {
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
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "case_category", item.ID)
	c.Set("audit_description", "Case category created")
	utils.Created(c, "Kategori berhasil dibuat", toCaseCategoryResponse(item))
}

func (h *CaseCategoryHandler) Update(c *gin.Context) {
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
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "case_category", item.ID)
	c.Set("audit_description", "Case category updated")
	utils.OK(c, "Kategori berhasil diupdate", toCaseCategoryResponse(item))
}

func (h *CaseCategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "case_category", id)
	c.Set("audit_description", "Case category deleted")
	utils.OK(c, "Kategori berhasil dihapus", nil)
}
