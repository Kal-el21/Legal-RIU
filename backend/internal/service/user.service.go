package service

import (
	"errors"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetAll(page, limit int, search string) ([]dto.UserResponse, int64, error)
	Create(req dto.CreateUserRequest) (*dto.UserResponse, error)
	Update(id string, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	Delete(id string) error
	UpdateStatus(id string, status string) error
	ResetPassword(id string, req dto.ResetPasswordRequest) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetAll(page, limit int, search string) ([]dto.UserResponse, int64, error) {
	users, total, err := s.userRepo.FindAll(page, limit, search)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.UserResponse, len(users))
	for i, u := range users {
		result[i] = toUserResponse(&u)
	}
	return result, total, nil
}

func (s *userService) Create(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal membuat user")
	}

	divisionName, divisionID, err := resolveDivisionSelection(s.userRepo, req.Division)
	if err != nil {
		return nil, err
	}

	role := entity.RoleUser
	switch req.Role {
	case "ADMIN":
		role = entity.RoleAdmin
	case "LEGAL":
		role = entity.RoleLegal
	case "EXTERNAL":
		role = entity.RoleExternal
	}

	user := &entity.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hash),
		Position:     req.Position,
		Division:     divisionName,
		DivisionID:   divisionID,
		Role:         role,
		Status:       entity.UserActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("gagal membuat user")
	}

	created, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		res := toUserResponse(user)
		return &res, nil
	}

	res := toUserResponse(created)
	return &res, nil
}

func (s *userService) Update(id string, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	divisionName, divisionID, err := resolveDivisionSelection(s.userRepo, req.Division)
	if err != nil {
		return nil, err
	}

	user.FullName = req.FullName
	user.Position = req.Position
	user.Division = divisionName
	user.DivisionID = divisionID

	if req.Role != "" {
		switch req.Role {
		case "ADMIN":
			user.Role = entity.RoleAdmin
		case "LEGAL":
			user.Role = entity.RoleLegal
		case "EXTERNAL":
			user.Role = entity.RoleExternal
		default:
			user.Role = entity.RoleUser
		}
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("gagal mengupdate user")
	}

	updated, err := s.userRepo.FindByID(uid)
	if err != nil {
		res := toUserResponse(user)
		return &res, nil
	}

	res := toUserResponse(updated)
	return &res, nil
}

func (s *userService) Delete(id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("user tidak valid")
	}

	if err := s.userRepo.Delete(uid); err != nil {
		return errors.New("gagal menghapus user")
	}
	return nil
}

func (s *userService) UpdateStatus(id string, statusStr string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("user tidak valid")
	}

	var status entity.UserStatus
	switch statusStr {
	case string(entity.UserActive):
		status = entity.UserActive
	case string(entity.UserInactive):
		status = entity.UserInactive
	default:
		return errors.New("status tidak valid")
	}

	if err := s.userRepo.UpdateStatus(uid, status); err != nil {
		return errors.New("gagal mengubah status user")
	}
	return nil
}

func (s *userService) ResetPassword(id string, req dto.ResetPasswordRequest) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("user tidak valid")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal reset password")
	}

	if err := s.userRepo.UpdatePassword(uid, string(hash)); err != nil {
		return errors.New("gagal reset password")
	}
	return nil
}
