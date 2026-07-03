package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	FindAll() ([]entity.Permission, error)
	FindByCodes(codes []string) ([]entity.Permission, error)
	FindRolePermissionCodes(role entity.UserRole) ([]string, error)
	FindUserOverrides(userID uuid.UUID) ([]entity.UserPermissionOverride, error)
	ReplaceUserOverrides(userID uuid.UUID, overrides []entity.UserPermissionOverride) error
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) FindAll() ([]entity.Permission, error) {
	var permissions []entity.Permission
	err := r.db.
		Where("is_active = ?", true).
		Order("feature ASC, action ASC, scope ASC, code ASC").
		Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) FindByCodes(codes []string) ([]entity.Permission, error) {
	if len(codes) == 0 {
		return []entity.Permission{}, nil
	}

	var permissions []entity.Permission
	err := r.db.Where("code IN ? AND is_active = ?", codes, true).Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) FindRolePermissionCodes(role entity.UserRole) ([]string, error) {
	var codes []string
	err := r.db.
		Table("role_permissions").
		Select("permissions.code").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role = ? AND role_permissions.deleted_at IS NULL AND permissions.is_active = ?", role, true).
		Order("permissions.code ASC").
		Scan(&codes).Error
	return codes, err
}

func (r *permissionRepository) FindUserOverrides(userID uuid.UUID) ([]entity.UserPermissionOverride, error) {
	var overrides []entity.UserPermissionOverride
	err := r.db.
		Preload("Permission").
		Where("user_id = ?", userID).
		Find(&overrides).Error
	return overrides, err
}

func (r *permissionRepository) ReplaceUserOverrides(userID uuid.UUID, overrides []entity.UserPermissionOverride) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entity.UserPermissionOverride{}).Error; err != nil {
			return err
		}

		if len(overrides) == 0 {
			return nil
		}

		for i := range overrides {
			overrides[i].UserID = userID
		}
		return tx.Create(&overrides).Error
	})
}
