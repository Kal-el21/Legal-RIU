package handler

import (
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// GET /api/v1/dashboard/stats
func (h *DashboardHandler) UserStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stats, err := h.svc.GetUserStats(userID)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", stats)
}

// GET /api/v1/dashboard/recent
func (h *DashboardHandler) UserRecent(c *gin.Context) {
	userID := middleware.GetUserID(c)
	data, err := h.svc.GetUserRecentActivity(userID)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", data)
}

// GET /api/v1/admin/dashboard/stats
func (h *DashboardHandler) AdminStats(c *gin.Context) {
	stats, err := h.svc.GetAdminStats()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", stats)
}

// GET /api/v1/admin/dashboard/recent
func (h *DashboardHandler) AdminRecent(c *gin.Context) {
	data, err := h.svc.GetAdminRecentActivity()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", data)
}
