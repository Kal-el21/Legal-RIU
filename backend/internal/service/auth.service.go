package service

import (
	"errors"
	"time"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(req dto.LoginRequest, ipAddress, userAgent *string) (*dto.LoginResponse, error)
	RefreshToken(req dto.RefreshTokenRequest) (*dto.LoginResponse, error)
	Logout(req dto.LogoutRequest) error
	GetUserByID(id string) (*dto.UserResponse, error)
	ChangePassword(userID string, req dto.ChangePasswordRequest) error
	UpdateProfile(userID string, req dto.UpdateProfileRequest) (*dto.UserResponse, error)
	UpdateNotification(userID string, req dto.UpdateNotificationRequest) error
	Toggle2FA(userID string, req dto.Toggle2FARequest) error
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	cfg              *config.Config
}

func NewAuthService(userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, cfg *config.Config) AuthService {
	return &authService{userRepo: userRepo, refreshTokenRepo: refreshTokenRepo, cfg: cfg}
}

func (s *authService) Login(req dto.LoginRequest, ipAddress, userAgent *string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("email atau password salah")
	}

	if user.Status == entity.UserInactive {
		return nil, errors.New("akun anda tidak aktif, hubungi administrator")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	return s.issueTokens(user, ipAddress, userAgent)
}

func (s *authService) RefreshToken(req dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	tokenHash := utils.HashToken(req.RefreshToken)
	storedToken, err := s.refreshTokenRepo.FindActiveByHash(tokenHash)
	if err != nil {
		return nil, errors.New("refresh token tidak valid")
	}

	if storedToken.User.Status == entity.UserInactive {
		_ = s.refreshTokenRepo.RevokeAllByUserID(storedToken.UserID)
		return nil, errors.New("akun anda tidak aktif, hubungi administrator")
	}

	if err := s.refreshTokenRepo.RevokeByHash(tokenHash); err != nil {
		return nil, errors.New("gagal memperbarui sesi")
	}

	return s.issueTokens(&storedToken.User, nil, nil)
}

func (s *authService) Logout(req dto.LogoutRequest) error {
	if req.RefreshToken == "" {
		return nil
	}
	return s.refreshTokenRepo.RevokeByHash(utils.HashToken(req.RefreshToken))
}

func (s *authService) GetUserByID(id string) (*dto.UserResponse, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	res := toUserResponse(user)
	return &res, nil
}

func (s *authService) ChangePassword(userID string, req dto.ChangePasswordRequest) error {
	uid, err := parseUUID(userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("password saat ini tidak sesuai")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal mengubah password")
	}

	return s.userRepo.UpdatePassword(uid, string(hash))
}

// UpdateProfile — user update nama/jabatan/divisi sendiri
func (s *authService) UpdateProfile(userID string, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}
	if err := s.userRepo.UpdateProfile(uid, req.FullName, req.Position, req.Division); err != nil {
		return nil, errors.New("gagal mengupdate profil")
	}
	user, _ := s.userRepo.FindByID(uid)
	res := toUserResponse(user)
	return &res, nil
}

// UpdateNotification — user set email notification preference
func (s *authService) UpdateNotification(userID string, req dto.UpdateNotificationRequest) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user tidak valid")
	}
	return s.userRepo.UpdateNotificationPref(uid, req.EmailNotifications)
}

// Toggle2FA — user enable/disable 2FA (validasi password dulu)
func (s *authService) Toggle2FA(userID string, req dto.Toggle2FARequest) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user tidak valid")
	}
	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return errors.New("password tidak sesuai")
	}
	return s.userRepo.UpdateTwoFA(uid, req.Enabled)
}

func (s *authService) issueTokens(user *entity.User, ipAddress, userAgent *string) (*dto.LoginResponse, error) {
	accessToken, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), s.cfg.JWT.Secret, s.cfg.JWT.ExpiresHours)
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	refreshToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return nil, errors.New("gagal membuat refresh token")
	}

	tokenRecord := &entity.RefreshToken{
		UserID:    user.ID,
		TokenHash: utils.HashToken(refreshToken),
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.JWT.RefreshExpiresHours) * time.Hour),
		IPAddress: func() string {
			if ipAddress != nil {
				return *ipAddress
			}
			return ""
		}(),
		UserAgent: func() string {
			if userAgent != nil {
				return *userAgent
			}
			return ""
		}(),
		LastUsedAt: time.Now(),
	}
	if err := s.refreshTokenRepo.Create(tokenRecord); err != nil {
		return nil, errors.New("gagal menyimpan sesi")
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(user),
	}, nil
}