package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type NotificationSettingHandler struct {
	svc service.NotificationSettingService
}

func NewNotificationSettingHandler(svc service.NotificationSettingService) *NotificationSettingHandler {
	return &NotificationSettingHandler{svc: svc}
}

// GET /api/v1/admin/notification-settings
func (h *NotificationSettingHandler) GetAll(c *gin.Context) {
	settings, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", settings)
}

// PUT /api/v1/admin/notification-settings/:id
func (h *NotificationSettingHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateNotificationSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	item, err := h.svc.Update(id, req)
	if err != nil {
		if err.Error() == "Setting tidak ditemukan" {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	_ = userID

	utils.OK(c, "Setting notifikasi berhasil diperbarui", item)
}

// GET /api/v1/admin/reminders-dashboard
func (h *NotificationSettingHandler) GetRemindersDashboard(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	result, err := h.svc.GetReminders(userID, role)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	totalPendingYellow := len(result.Yellow)
	totalPendingRed := len(result.Red)
	totalOverdue := totalPendingRed

	summary := map[string]interface{}{
		"total_pending_yellow": totalPendingYellow,
		"total_pending_red":    totalPendingRed,
		"total_overdue":        totalOverdue,
		"yellow":               result.Yellow,
		"red":                  result.Red,
	}
	utils.OK(c, "Success", summary)
}

// GET /api/v1/dashboard/reminders
func (h *NotificationSettingHandler) GetReminders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	result, err := h.svc.GetReminders(userID, role)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", result)
}

// GET /api/v1/legal/reminders
func (h *NotificationSettingHandler) GetLegalReminders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	result, err := h.svc.GetReminders(userID, role)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", result)
}
