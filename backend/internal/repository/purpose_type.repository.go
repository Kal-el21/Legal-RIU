package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurposeTypeRepository interface {
	FindAll() ([]entity.PurposeType, error)
	FindByID(id uuid.UUID) (*entity.PurposeType, error)
	FindByName(name string) (*entity.PurposeType, error)
	Create(purposeType *entity.PurposeType) error
	Update(purposeType *entity.PurposeType) error
	Delete(id uuid.UUID) error
}

type purposeTypeRepository struct {
	db *gorm.DB
}

func NewPurposeTypeRepository(db *gorm.DB) PurposeTypeRepository {
	return &purposeTypeRepository{db: db}
}

func (r *purposeTypeRepository) FindAll() ([]entity.PurposeType, error) {
	var items []entity.PurposeType
	err := r.db.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *purposeTypeRepository) FindByID(id uuid.UUID) (*entity.PurposeType, error) {
	var pt entity.PurposeType
	err := r.db.Where("id = ?", id).First(&pt).Error
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

func (r *purposeTypeRepository) FindByName(name string) (*entity.PurposeType, error) {
	var pt entity.PurposeType
	err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&pt).Error
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

func (r *purposeTypeRepository) Create(purposeType *entity.PurposeType) error {
	return r.db.Create(purposeType).Error
}

func (r *purposeTypeRepository) Update(purposeType *entity.PurposeType) error {
	return r.db.Save(purposeType).Error
}

func (r *purposeTypeRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.PurposeType{}, "id = ?", id).Error
}
