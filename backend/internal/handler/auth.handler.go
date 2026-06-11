package handler

import (
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
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

	utils.OK(c, "Login berhasil", res)
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	res, err := h.authService.RefreshToken(req)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.OK(c, "Sesi berhasil diperbarui", res)
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	if err := h.authService.Logout(req); err != nil {
		utils.BadRequest(c, "Gagal logout", nil)
		return
	}

	utils.OK(c, "Logout berhasil", nil)
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
	status := "diaktifkan"
	if !req.Enabled {
		status = "dinonaktifkan"
	}
	utils.OK(c, "Two-step login berhasil "+status, nil)
}