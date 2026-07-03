package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(email string) (*entity.User, error)
	FindByID(id uuid.UUID) (*entity.User, error)
	Create(user *entity.User) error
	Update(user *entity.User) error
	FindAll(page, limit int, search string) ([]entity.User, int64, error)
	UpdateStatus(id uuid.UUID, status entity.UserStatus) error
	Delete(id uuid.UUID) error
	UpdatePassword(id uuid.UUID, passwordHash string) error
	UpdateProfile(id uuid.UUID, fullName, position, division string, divisionID *uuid.UUID) error
	UpdateNotificationPref(id uuid.UUID, emailNotif bool) error
	UpdateTwoFA(id uuid.UUID, enabled bool) error
	FindDivisionByID(id uuid.UUID) (*entity.Division, error)
	FindDivisionByName(name string) (*entity.Division, error)
	FindCompanyByID(id uuid.UUID) (*entity.Company, error)
	FindCompanyByDomain(domain string) (*entity.Company, error)
	FindFirstCompany() (*entity.Company, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.
		Preload("DivisionRef").
		Preload("Company").
		Preload("PurposeType").
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.
		Preload("DivisionRef").
		Preload("Company").
		Preload("PurposeType").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *entity.User) error {
	return r.db.Omit("DivisionRef").Save(user).Error
}

func (r *userRepository) FindAll(page, limit int, search string) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := r.db.Preload("DivisionRef").Preload("Company").Preload("PurposeType").Model(&entity.User{})
	if search != "" {
		query = query.Where("full_name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	err := query.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

func (r *userRepository) UpdateStatus(id uuid.UUID, status entity.UserStatus) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("status", status).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.User{}, id).Error
}

func (r *userRepository) UpdatePassword(id uuid.UUID, passwordHash string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("password_hash", passwordHash).Error
}

func (r *userRepository) UpdateProfile(id uuid.UUID, fullName, position, division string, divisionID *uuid.UUID) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"full_name":   fullName,
		"position":    position,
		"division":    division,
		"division_id": divisionID,
	}).Error
}

func (r *userRepository) UpdateNotificationPref(id uuid.UUID, emailNotif bool) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("email_notifications", emailNotif).Error
}

func (r *userRepository) UpdateTwoFA(id uuid.UUID, enabled bool) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("two_fa_enabled", enabled).Error
}

func (r *userRepository) FindDivisionByID(id uuid.UUID) (*entity.Division, error) {
	var division entity.Division
	if err := r.db.Where("id = ?", id).First(&division).Error; err != nil {
		return nil, err
	}
	return &division, nil
}

func (r *userRepository) FindDivisionByName(name string) (*entity.Division, error) {
	var division entity.Division
	if err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&division).Error; err != nil {
		return nil, err
	}
	return &division, nil
}

func (r *userRepository) FindCompanyByID(id uuid.UUID) (*entity.Company, error) {
	var company entity.Company
	if err := r.db.Where("id = ?", id).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *userRepository) FindCompanyByDomain(domain string) (*entity.Company, error) {
	var company entity.Company
	if err := r.db.Where("LOWER(email_domain) = LOWER(?)", domain).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *userRepository) FindFirstCompany() (*entity.Company, error) {
	var company entity.Company
	if err := r.db.First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}
