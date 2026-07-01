package handler

import (
	"strconv"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	auditLogSvc  service.AuditLogService
	auditLogRepo repository.AuditLogRepository
}

func NewAuditLogHandler(auditLogSvc service.AuditLogService, auditLogRepo repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{auditLogSvc: auditLogSvc, auditLogRepo: auditLogRepo}
}

type AuditLogResponse struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	User        *dto.UserResponse  `json:"user,omitempty"`
	Action      entity.AuditAction `json:"action"`
	EntityType  string             `json:"entity_type"`
	EntityID    string             `json:"entity_id"`
	OldValue    *string            `json:"old_value,omitempty"`
	NewValue    *string            `json:"new_value,omitempty"`
	Description *string            `json:"description,omitempty"`
	IPAddress   string             `json:"ip_address"`
	UserAgent   string             `json:"user_agent"`
	CreatedAt   time.Time          `json:"created_at"`
}

// GET /api/v1/admin/audit-logs
func (h *AuditLogHandler) GetAll(c *gin.Context) {
	var filters dto.AuditLogFilters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	filters.Page = page
	filters.Limit = limit

	if action := c.Query("action"); action != "" {
		filters.Action = &action
	}
	if entityType := c.Query("entity_type"); entityType != "" {
		filters.EntityType = &entityType
	}
	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters.DateFrom = &t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			dateToEnd := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filters.DateTo = &dateToEnd
		}
	}

	items, total, err := h.auditLogRepo.GetAll(repository.AuditLogFilters{
		Page:       filters.Page,
		Limit:      filters.Limit,
		Action:     filters.Action,
		EntityType: filters.EntityType,
		EntityID:   filters.EntityID,
		UserID:     filters.UserID,
		DateFrom:   filters.DateFrom,
		DateTo:     filters.DateTo,
		Search:     filters.Search,
	})
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	response := map[string]interface{}{
		"items":       items,
		"total":       total,
		"page":        filters.Page,
		"limit":       filters.Limit,
		"total_pages": totalPages,
	}

	utils.OK(c, "Success", response)
}
