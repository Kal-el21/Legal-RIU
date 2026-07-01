package service

import (
	"errors"
	"log"
	"strings"
	"time"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	if err == nil {
		if user.Status == entity.UserInactive {
			return nil, errors.New("akun anda tidak aktif, hubungi administrator")
		}

		switch user.AuthType {
		case "", entity.AuthTypeLocal:
			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
				return nil, errors.New("email atau password salah")
			}
		case entity.AuthTypeLDAP:
			if err := s.authenticateLDAPUser(user, req); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("tipe autentikasi tidak dikenal")
		}

		return s.issueTokens(user, ipAddress, userAgent)
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[AuthService] failed to find user by email %s: %v", req.Email, err)
		return nil, errors.New("email atau password salah")
	}

	return s.loginWithLDAPAutoCreate(req, ipAddress, userAgent)
}

func (s *authService) authenticateLDAPUser(user *entity.User, req dto.LoginRequest) error {
	ldapUsername := extractLDAPUsername(req.Email)
	ldapInfo, err := utils.LDAPAuthenticate(s.cfg.LDAP, ldapUsername, req.Password)
	if err != nil {
		log.Printf("[AuthService] LDAP authenticate failed for %s: %v", req.Email, err)
		return errors.New("email atau password salah")
	}

	s.syncUserFromLDAP(user, ldapInfo)
	return nil
}

func (s *authService) loginWithLDAPAutoCreate(req dto.LoginRequest, ipAddress, userAgent *string) (*dto.LoginResponse, error) {
	if s.cfg.LDAP.Host == "" {
		return nil, errors.New("email atau password salah")
	}

	ldapUsername := extractLDAPUsername(req.Email)
	ldapInfo, err := utils.LDAPAuthenticate(s.cfg.LDAP, ldapUsername, req.Password)
	if err != nil {
		log.Printf("[AuthService] LDAP auto-create failed for %s: %v", req.Email, err)
		return nil, errors.New("email atau password salah")
	}

	email := ldapInfo.Email
	if email == "" {
		email = req.Email
	}

	if existing, err := s.userRepo.FindByEmail(email); err == nil {
		if existing.Status == entity.UserInactive {
			return nil, errors.New("akun anda tidak aktif, hubungi administrator")
		}
		if !existing.IsLDAP() {
			return nil, errors.New("email atau password salah")
		}
		s.syncUserFromLDAP(existing, ldapInfo)
		return s.issueTokens(existing, ipAddress, userAgent)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[AuthService] failed to check LDAP email %s: %v", email, err)
		return nil, errors.New("email atau password salah")
	}

	newUser := &entity.User{
		FullName:     ldapInfo.FullName,
		Email:        email,
		PasswordHash: "",
		AuthType:     entity.AuthTypeLDAP,
		Position:     ldapInfo.Position,
		Division:     ldapInfo.Division,
		Role:         entity.RoleUser,
		Status:       entity.UserActive,
	}

	if newUser.FullName == "" {
		newUser.FullName = ldapUsername
	}
	if newUser.Position == "" {
		newUser.Position = s.cfg.LDAP.DefaultPosition
	}
	if newUser.Division == "" {
		newUser.Division = s.cfg.LDAP.DefaultDivision
	}

	if err := s.userRepo.Create(newUser); err != nil {
		if isDuplicateKeyError(err) {
			existing, fetchErr := s.userRepo.FindByEmail(email)
			if fetchErr == nil && existing.IsLDAP() && existing.Status == entity.UserActive {
				s.syncUserFromLDAP(existing, ldapInfo)
				return s.issueTokens(existing, ipAddress, userAgent)
			}
		}
		log.Printf("[AuthService] failed to auto-create LDAP user %s: %v", email, err)
		return nil, errors.New("gagal membuat akun secara otomatis")
	}

	return s.issueTokens(newUser, ipAddress, userAgent)
}

func (s *authService) syncUserFromLDAP(user *entity.User, ldapInfo *utils.LDAPUserInfo) {
	changed := false

	if ldapInfo.FullName != "" && user.FullName != ldapInfo.FullName {
		user.FullName = ldapInfo.FullName
		changed = true
	}
	if ldapInfo.Position != "" && user.Position != ldapInfo.Position {
		user.Position = ldapInfo.Position
		changed = true
	}
	if ldapInfo.Division != "" && user.Division != ldapInfo.Division {
		user.Division = ldapInfo.Division
		changed = true
	}

	if ldapInfo.Email != "" && user.Email != ldapInfo.Email {
		existing, err := s.userRepo.FindByEmail(ldapInfo.Email)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user.Email = ldapInfo.Email
			changed = true
		} else if err != nil {
			log.Printf("[AuthService] failed to check LDAP email sync %s: %v", ldapInfo.Email, err)
		} else if existing.ID == user.ID {
			user.Email = ldapInfo.Email
			changed = true
		} else {
			log.Printf("[AuthService] skipped LDAP email sync for user %s; email %s is already used", user.ID, ldapInfo.Email)
		}
	}

	if changed {
		if err := s.userRepo.Update(user); err != nil {
			log.Printf("[AuthService] failed to sync LDAP user %s: %v", user.ID, err)
		}
	}
}

func extractLDAPUsername(email string) string {
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return email
}

func isDuplicateKeyError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
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

	if user.IsLDAP() {
		return errors.New("password dikelola oleh Active Directory dan tidak dapat diubah melalui aplikasi ini")
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

	divisionName, divisionID, err := resolveDivisionSelection(s.userRepo, req.Division)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateProfile(uid, req.FullName, req.Position, divisionName, divisionID); err != nil {
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
	if user.IsLDAP() {
		return errors.New("pengaturan two-step login berbasis password lokal tidak tersedia untuk akun LDAP")
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
