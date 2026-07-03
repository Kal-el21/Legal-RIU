package middleware

import (
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

const ContextPermissions = "permissions"

type PermissionChecker interface {
	GetEffectivePermissionCodes(userID string, role string) ([]string, error)
	HasAnyPermission(userID string, role string, codes ...string) bool
}

func PermissionContextMiddleware(checker PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		role := GetUserRole(c)
		permissions, err := checker.GetEffectivePermissionCodes(userID, role)
		if err != nil {
			utils.Forbidden(c, "Gagal memuat permission user")
			c.Abort()
			return
		}

		c.Set(ContextPermissions, permissions)
		c.Next()
	}
}

func RequirePermission(checker PermissionChecker, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if HasAnyPermission(c, codes...) {
			c.Next()
			return
		}

		userID := GetUserID(c)
		role := GetUserRole(c)
		if checker.HasAnyPermission(userID, role, codes...) {
			c.Next()
			return
		}

		utils.Forbidden(c, "Anda tidak memiliki permission untuk fitur ini")
		c.Abort()
	}
}

func HasPermission(c *gin.Context, code string) bool {
	return HasAnyPermission(c, code)
}

func HasAnyPermission(c *gin.Context, codes ...string) bool {
	if len(codes) == 0 {
		return true
	}

	value, ok := c.Get(ContextPermissions)
	if !ok {
		return false
	}

	permissions, ok := value.([]string)
	if !ok {
		return false
	}

	allowed := make(map[string]bool, len(permissions))
	for _, permission := range permissions {
		allowed[permission] = true
	}
	for _, code := range codes {
		if allowed[code] {
			return true
		}
	}
	return false
}

func RoleWithAllAccess(c *gin.Context, allAccessPermission string) string {
	role := GetUserRole(c)
	if HasPermission(c, allAccessPermission) {
		return "ADMIN"
	}
	return role
}
