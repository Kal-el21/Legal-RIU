package handler

import (
	"strconv"

	"legal-riu-portal/internal/dto"
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

func toDivisionResponse(d *entity.Division) dto.DivisionResponse {
	return dto.DivisionResponse{
		ID:          d.ID.String(),
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func (h *DivisionHandler) GetAll(c *gin.Context) {
	search := c.Query("search")
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "200"))
	if err != nil {
		utils.BadRequest(c, "Limit tidak valid", err.Error())
		return
	}
	items, err := h.divisionService.GetAll(search, limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	response := make([]dto.DivisionResponse, 0, len(items))
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

func (h *DivisionHandler) ImportDivisions(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "File Excel wajib diupload", err.Error())
		return
	}
	result, err := h.divisionService.ImportFromExcel(file)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Impor divisi selesai", result)
}

func (h *DivisionHandler) DownloadDivisionTemplate(c *gin.Context) {
	buf, err := h.divisionService.GenerateImportTemplate()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	c.DataFromReader(-1, -1, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf, map[string]string{
		"Content-Disposition": `attachment; filename="division-template.xlsx"`,
	})
}
