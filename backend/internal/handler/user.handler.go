package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GET /api/v1/admin/users
func (h *UserHandler) GetAll(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, "Query tidak valid", err.Error())
		return
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 10
	}

	users, total, err := h.svc.GetAll(query.Page, query.Limit, query.Search)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.OK(c, "Success", gin.H{
		"items":       users,
		"total":       total,
		"page":        query.Page,
		"limit":       query.Limit,
		"total_pages": (total + int64(query.Limit) - 1) / int64(query.Limit),
	})
}

// POST /api/v1/admin/users
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	user, err := h.svc.Create(req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.Created(c, "User berhasil dibuat", user)
}

// PUT /api/v1/admin/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	user, err := h.svc.Update(id, req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "User berhasil diupdate", user)
}

// DELETE /api/v1/admin/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "User berhasil dihapus", nil)
}

// PATCH /api/v1/admin/users/:id/status
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	if err := h.svc.UpdateStatus(id, body.Status); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Status user berhasil diubah", nil)
}

// POST /api/v1/admin/users/:id/reset-password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	if err := h.svc.ResetPassword(id, req); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Password berhasil direset", nil)
}