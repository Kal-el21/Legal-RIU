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

// reminderAccess menghitung hak user di tiap modul pengajuan dari permission
// efektifnya. Nama modul sama dengan submission_type di lonceng.
func reminderAccess(c *gin.Context) service.ReminderAccess {
	access := service.ReminderAccess{}
	for _, feature := range []string{"legal_opinion", "document_review"} {
		access[feature] = service.ReminderScope{
			Owner:     middleware.HasAnyPermission(c, feature+".view.own", feature+".view.all"),
			Processor: middleware.HasPermission(c, feature+".update_status.all"),
		}
	}
	return access
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

// GET /api/v1/notifications
func (h *NotificationSettingHandler) GetReminders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, limit := parseReminderPagination(c)

	result, err := h.svc.GetReminders(userID, reminderAccess(c), page, limit, c.Query("warning_only") == "true")
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", result)
}

// PATCH /api/v1/notifications/read
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

// PATCH /api/v1/notifications/read-all
func (h *NotificationSettingHandler) MarkAllRemindersRead(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if err := h.svc.MarkAllRemindersRead(userID, reminderAccess(c)); err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.OK(c, "Semua notifikasi berhasil ditandai terbaca", nil)
}
