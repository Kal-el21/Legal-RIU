package handler

import (
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type DivisionHandler struct {
	divisionService service.DivisionService
}

func NewDivisionHandler(divisionService service.DivisionService) *DivisionHandler {
	return &DivisionHandler{divisionService: divisionService}
}

type DivisionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toDivisionResponse(d *entity.Division) DivisionResponse {
	return DivisionResponse{
		ID:          d.ID.String(),
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *DivisionHandler) GetAll(c *gin.Context) {
	search := c.Query("search")
	limit := 0
	if l := c.Query("limit"); l != "" {
		// parse limit if needed, for now just use 0 = unlimited
	}
	items, err := h.divisionService.GetAll(search, limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	response := make([]DivisionResponse, 0, len(items))
	for i := range items {
		response = append(response, toDivisionResponse(&items[i]))
	}
	utils.OK(c, "Success", response)
}

func (h *DivisionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.divisionService.GetByID(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", toDivisionResponse(item))
}

func (h *DivisionHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.divisionService.Create(req.Name, req.Description)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.Created(c, "Divisi berhasil dibuat", toDivisionResponse(item))
}

func (h *DivisionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	item, err := h.divisionService.Update(id, req.Name, req.Description)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Divisi berhasil diupdate", toDivisionResponse(item))
}

func (h *DivisionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.divisionService.Delete(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Divisi berhasil dihapus", nil)
}
