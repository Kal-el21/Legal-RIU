package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LegalMaterialRepository interface {
	FindAll() ([]entity.LegalMaterial, error)
	FindByID(id uuid.UUID) (*entity.LegalMaterial, error)
	Create(material *entity.LegalMaterial) error
	Update(material *entity.LegalMaterial) error
	Delete(id uuid.UUID) error
}

type legalMaterialRepository struct {
	db *gorm.DB
}

func NewLegalMaterialRepository(db *gorm.DB) LegalMaterialRepository {
	return &legalMaterialRepository{db: db}
}

func (r *legalMaterialRepository) FindAll() ([]entity.LegalMaterial, error) {
	var items []entity.LegalMaterial
	err := r.db.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *legalMaterialRepository) FindByID(id uuid.UUID) (*entity.LegalMaterial, error) {
	var material entity.LegalMaterial
	err := r.db.Where("id = ?", id).First(&material).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
}

func (r *legalMaterialRepository) Create(material *entity.LegalMaterial) error {
	return r.db.Create(material).Error
}

func (r *legalMaterialRepository) Update(material *entity.LegalMaterial) error {
	return r.db.Save(material).Error
}

func (r *legalMaterialRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.LegalMaterial{}, "id = ?", id).Error
}
