package middleware

import (
	"errors"
	"strings"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthUserFinder membaca user pemilik token. Dipenuhi repository.UserRepository.
type AuthUserFinder interface {
	FindForAuth(id uuid.UUID) (*entity.User, error)
}

const (
	ContextUserID = "userID"
	ContextEmail  = "userEmail"
	ContextRole   = "userRole"
)

// AuthMiddleware validates JWT from Authorization header OR cookie. Token hanya
// dipercaya untuk identitas; status dan peran dibaca dari database di setiap
// permintaan, supaya user yang dinonaktifkan, dihapus, atau diganti perannya
// tidak perlu menunggu tokennya kedaluwarsa.
func AuthMiddleware(cfg *config.Config, users AuthUserFinder) gin.HandlerFunc {
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

		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			utils.Unauthorized(c, "Token tidak valid atau sudah expired")
			c.Abort()
			return
		}
		user, err := users.FindForAuth(uid)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Unauthorized(c, "Akun tidak ditemukan")
			c.Abort()
			return
		}
		if err != nil {
			// Gangguan database bukan alasan mengeluarkan user.
			utils.InternalError(c, "Gagal memeriksa akun")
			c.Abort()
			return
		}
		if user.Status == entity.UserInactive {
			utils.Unauthorized(c, "Akun anda tidak aktif, hubungi administrator")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, user.Email)
		c.Set(ContextRole, string(user.Role))
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
