package handler

import (
	"strconv"

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

func parseReminderPagination(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit
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
	page, limit := parseReminderPagination(c)

	result, err := h.svc.GetReminders(userID, role, page, limit)
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
		"none":                 result.None,
		"items":                result.Items,
		"total":                result.Total,
		"unread_total":         result.UnreadTotal,
		"page":                 result.Page,
		"limit":                result.Limit,
		"total_pages":          result.TotalPages,
	}
	utils.OK(c, "Success", summary)
}

// GET /api/v1/dashboard/reminders
func (h *NotificationSettingHandler) GetReminders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	page, limit := parseReminderPagination(c)

	result, err := h.svc.GetReminders(userID, role, page, limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", result)
}

// PATCH /api/v1/dashboard/reminders/read
func (h *NotificationSettingHandler) MarkReminderRead(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.MarkReminderReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	if err := h.svc.MarkReminderRead(userID, req.SubmissionType, req.SubmissionID); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	utils.OK(c, "Notifikasi berhasil ditandai terbaca", nil)
}

// PATCH /api/v1/dashboard/reminders/read-all
func (h *NotificationSettingHandler) MarkAllRemindersRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	if err := h.svc.MarkAllRemindersRead(userID, role); err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.OK(c, "Semua notifikasi berhasil ditandai terbaca", nil)
}

// GET /api/v1/legal/reminders
func (h *NotificationSettingHandler) GetLegalReminders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	page, limit := parseReminderPagination(c)

	result, err := h.svc.GetReminders(userID, role, page, limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", result)
}
