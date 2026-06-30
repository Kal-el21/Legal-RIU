package handler

import (
	"time"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService   service.AuthService
	cfg           *config.Config
	auditLogSvc   service.AuditLogService
}

func NewAuthHandler(authService service.AuthService, cfg *config.Config, auditLogSvc service.AuditLogService) *AuthHandler {
	return &AuthHandler{authService: authService, cfg: cfg, auditLogSvc: auditLogSvc}
}

func getClientIP(c *gin.Context) string {
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return c.ClientIP()
}

func getUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

// setAuthCookies sets httpOnly cookies for access and refresh tokens
func (h *AuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	isProduction := h.cfg.App.Env == "production"

	// Access token - convert hours to seconds
	accessMaxAge := h.cfg.JWT.ExpiresHours * int(time.Hour / time.Second)
	if accessMaxAge <= 0 {
		accessMaxAge = 3600 // 1 hour default
	}

	// Refresh token
	refreshMaxAge := h.cfg.JWT.RefreshExpiresHours * int(time.Hour / time.Second)
	if refreshMaxAge <= 0 {
		refreshMaxAge = 604800 // 7 days default
	}

	c.SetCookie("access_token", accessToken, accessMaxAge, "/", "", isProduction, true)
	c.SetCookie("refresh_token", refreshToken, refreshMaxAge, "/", "", isProduction, true)
}

// clearAuthCookies clears auth cookies
func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	ip := getClientIP(c)
	ua := getUserAgent(c)
	res, err := h.authService.Login(req, &ip, &ua)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	h.setAuthCookies(c, res.AccessToken, res.RefreshToken)
	userUUID, _ := uuid.Parse(res.User.ID)
	_ = h.auditLogSvc.Log(userUUID, entity.ActionLogin, "auth", userUUID, nil, nil, nil, ip, ua)
	utils.OK(c, "Login berhasil", res)
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Try to get refresh token from cookie first
	refreshToken, _ := c.Cookie("refresh_token")
	if refreshToken == "" {
		var req dto.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		refreshToken = req.RefreshToken
	}

	req := dto.RefreshTokenRequest{RefreshToken: refreshToken}

	res, err := h.authService.RefreshToken(req)
	if err != nil {
		h.clearAuthCookies(c)
		utils.Unauthorized(c, err.Error())
		return
	}

	h.setAuthCookies(c, res.AccessToken, res.RefreshToken)
	utils.OK(c, "Sesi berhasil diperbarui", res)
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie("refresh_token")

	var req dto.LogoutRequest
	if refreshToken != "" {
		req.RefreshToken = refreshToken
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			h.clearAuthCookies(c)
			utils.OK(c, "Logout berhasil", nil)
			return
		}
	}

	_ = h.authService.Logout(req)
	h.clearAuthCookies(c)
	uid, _ := uuid.Parse(userIDFromToken(c))
	if uid != uuid.Nil {
		desc := "User logout"
		_ = h.auditLogSvc.Log(uid, entity.ActionLogout, "auth", uid, nil, nil, &desc, c.ClientIP(), c.GetHeader("User-Agent"))
	}
	utils.OK(c, "Logout berhasil", nil)
}

func userIDFromToken(c *gin.Context) string {
	if uid := c.GetString(middleware.ContextUserID); uid != "" {
		return uid
	}
	return ""
}

func stringPtr(s string) *string {
	return &s
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		utils.NotFound(c, "User tidak ditemukan")
		return
	}
	utils.OK(c, "Success", user)
}

// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.authService.ChangePassword(userID, req); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "auth", userID)
	c.Set("audit_description", "Password changed")
	utils.OK(c, "Password berhasil diubah", nil)
}

// PUT /api/v1/settings/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	user, err := h.authService.UpdateProfile(userID, req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "auth", userID)
	utils.OK(c, "Profil berhasil diupdate", user)
}

// PUT /api/v1/settings/notifications
func (h *AuthHandler) UpdateNotification(c *gin.Context) {
	var req dto.UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.authService.UpdateNotification(userID, req); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "auth", userID)
	c.Set("audit_description", "Notification preference updated")
	utils.OK(c, "Preferensi notifikasi diperbarui", nil)
}

// PUT /api/v1/settings/two-fa
func (h *AuthHandler) Toggle2FA(c *gin.Context) {
	var req dto.Toggle2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.authService.Toggle2FA(userID, req); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "auth", userID)
	status := "diaktifkan"
	if !req.Enabled {
		status = "dinonaktifkan"
	}
	c.Set("audit_description", "Two-factor authentication " + status)
	utils.OK(c, "Two-step login berhasil "+status, nil)
}
