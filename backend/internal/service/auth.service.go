package service

import (
	"errors"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
	GetUserByID(id string) (*dto.UserResponse, error)
	ChangePassword(userID string, req dto.ChangePasswordRequest) error
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{userRepo: userRepo, cfg: cfg}
}

func (s *authService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
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

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), s.cfg.JWT.Secret, s.cfg.JWT.ExpiresHours)
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	return &dto.LoginResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
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

// ─── Helpers ──────────────────────────────────────────────────────────────────

func toUserResponse(user *entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       user.ID.String(),
		FullName: user.FullName,
		Email:    user.Email,
		Position: user.Position,
		Division: user.Division,
		Role:     string(user.Role),
		Status:   string(user.Status),
	}
}
