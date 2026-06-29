package service

import (
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

type AuditLogService interface {
	Log(userID uuid.UUID, action entity.AuditAction, entityType string, entityID uuid.UUID,
		oldValue, newValue, description *string, ipAddress, userAgent string) error
	LogFromContext(c *gin.Context, action entity.AuditAction, entityType string, entityID uuid.UUID,
		oldValue, newValue, description *string) error
}

type auditLogService struct {
	repo repository.AuditLogRepository
}

func NewAuditLogService(repo repository.AuditLogRepository) AuditLogService {
	return &auditLogService{repo: repo}
}

func (s *auditLogService) Log(userID uuid.UUID, action entity.AuditAction, entityType string, entityID uuid.UUID,
	oldValue, newValue, description *string, ipAddress, userAgent string) error {
	log := &entity.AuditLog{
		UserID:      userID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		OldValue:    oldValue,
		NewValue:    newValue,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}
	return s.repo.Create(log)
}

func (s *auditLogService) LogFromContext(c *gin.Context, action entity.AuditAction, entityType string, entityID uuid.UUID,
	oldValue, newValue, description *string) error {
	userIDStr := c.GetString("userID")
	if userIDStr == "" {
		return nil
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	return s.Log(userID, action, entityType, entityID, oldValue, newValue, description, ipAddress, userAgent)
}

func GetIPAddress(c *gin.Context) string {
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return c.ClientIP()
}
