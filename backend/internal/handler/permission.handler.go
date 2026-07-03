package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionSvc service.PermissionService
}

func NewPermissionHandler(permissionSvc service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionSvc: permissionSvc}
}

func (h *PermissionHandler) GetCatalog(c *gin.Context) {
	permissions, err := h.permissionSvc.GetCatalog()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", permissions)
}

func (h *PermissionHandler) GetUserAccess(c *gin.Context) {
	access, err := h.permissionSvc.GetUserAccess(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Success", access)
}

func (h *PermissionHandler) UpdateUserAccess(c *gin.Context) {
	var req dto.UpdateUserPermissionOverridesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	access, err := h.permissionSvc.UpdateUserOverrides(c.Param("id"), req, middleware.GetUserID(c))
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	middleware.SetAuditContext(c, entity.ActionPermissionUpdate, "user", c.Param("id"))
	c.Set("audit_description", "User permission updated")
	utils.OK(c, "Permission user berhasil diupdate", access)
}
