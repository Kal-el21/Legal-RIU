package middleware

import (
	"strings"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "userID"
	ContextEmail  = "userEmail"
	ContextRole   = "userRole"
)

// AuthMiddleware validates JWT from Authorization header OR cookie
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := getTokenFromRequest(c)

		if tokenStr == "" {
			utils.Unauthorized(c, "Token tidak ditemukan")
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenStr, cfg.JWT.Secret)
		if err != nil {
			utils.Unauthorized(c, "Token tidak valid atau sudah expired")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

// getTokenFromRequest extracts token from Authorization header or cookie
func getTokenFromRequest(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Fall back to access_token cookie
	token, _ := c.Cookie("access_token")
	return token
}

// RoleMiddleware restricts access to specific roles
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString(ContextRole)
		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}
		utils.Forbidden(c, "Anda tidak memiliki akses ke halaman ini")
		c.Abort()
	}
}

// GetUserID extracts userID string from gin context
func GetUserID(c *gin.Context) string {
	return c.GetString(ContextUserID)
}

// GetUserRole extracts role from gin context
func GetUserRole(c *gin.Context) string {
	return c.GetString(ContextRole)
}
