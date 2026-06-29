package middleware

import (
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuditMiddleware(auditLogSvc service.AuditLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		action := c.GetString("audit_action")
		if action == "" {
			return
		}

		entityType := c.GetString("audit_entity_type")
		entityIDStr := c.GetString("audit_entity_id")
		if entityType == "" || entityIDStr == "" {
			return
		}

		oldValue := getStringPtr(c, "audit_old_value")
		newValue := getStringPtr(c, "audit_new_value")
		description := getStringPtr(c, "audit_description")

		entityID, err := uuid.Parse(entityIDStr)
		if err != nil {
			return
		}

		_ = auditLogSvc.LogFromContext(c, entity.AuditAction(action), entityType, entityID, oldValue, newValue, description)
	}
}

func SetAuditContext(c *gin.Context, action entity.AuditAction, entityType string, entityID interface{}) {
	c.Set("audit_action", string(action))
	c.Set("audit_entity_type", entityType)
	c.Set("audit_entity_id", entityID)
}

func SetAuditContextWithValues(c *gin.Context, action entity.AuditAction, entityType string, entityID interface{}, oldValue, newValue, description *string) {
	SetAuditContext(c, action, entityType, entityID)
	if oldValue != nil {
		c.Set("audit_old_value", *oldValue)
	}
	if newValue != nil {
		c.Set("audit_new_value", *newValue)
	}
	if description != nil {
		c.Set("audit_description", *description)
	}
}

func getStringPtr(c *gin.Context, key string) *string {
	val, exists := c.Get(key)
	if !exists {
		return nil
	}
	str, ok := val.(string)
	if !ok {
		return nil
	}
	return &str
}
