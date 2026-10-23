package middleware

import (
	"fmt"
	"log"

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
		if entityType == "" {
			return
		}
		raw, _ := c.Get("audit_entity_id")
		entityID := idEntitas(raw)

		oldValue := getStringPtr(c, "audit_old_value")
		newValue := getStringPtr(c, "audit_new_value")
		description := getStringPtr(c, "audit_description")

		// Gagal mencatat tidak menggagalkan aksinya, tapi harus terlihat di log server.
		if err := auditLogSvc.LogFromContext(c, entity.AuditAction(action), entityType, entityID, oldValue, newValue, description); err != nil {
			log.Printf("audit log tidak tersimpan (%s %s %s): %v", action, entityType, entityID, err)
		}
	}
}

// idEntitas membaca ID entitas yang dipasang SetAuditContext. Pemanggil mengirim string,
// uuid.UUID (yang punya String()), atau nilai lain yang punya String(); semuanya diterima. Dulu nilai ini dibaca
// dengan c.GetString, yang hanya menerima string, sehingga ID bertipe uuid.UUID menjadi
// kosong dan barisnya diam-diam tidak tersimpan. ID yang bukan UUID (mis. "import" untuk
// impor Excel) dicatat sebagai uuid.Nil, supaya barisnya tetap ada.
func idEntitas(v any) uuid.UUID {
	var s string
	switch x := v.(type) {
	case *uuid.UUID: // pointer kosong tidak boleh sampai ke String()
		if x != nil {
			return *x
		}
	case string:
		s = x
	case fmt.Stringer:
		s = x.String()
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
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
