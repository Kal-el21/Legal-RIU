package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseTypeRepository interface {
	FindAll() ([]entity.CaseType, error)
	FindByID(id uuid.UUID) (*entity.CaseType, error)
	FindByCode(code string) (*entity.CaseType, error)
	Create(caseType *entity.CaseType) error
	Update(caseType *entity.CaseType) error
	Delete(id uuid.UUID) error
}

type caseTypeRepository struct {
	db *gorm.DB
}

func NewCaseTypeRepository(db *gorm.DB) CaseTypeRepository {
	return &caseTypeRepository{db: db}
}

func (r *caseTypeRepository) FindAll() ([]entity.CaseType, error) {
	var items []entity.CaseType
	err := r.db.Order("label ASC").Find(&items).Error
	return items, err
}

func (r *caseTypeRepository) FindByID(id uuid.UUID) (*entity.CaseType, error) {
	var ct entity.CaseType
	err := r.db.Where("id = ?", id).First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *caseTypeRepository) FindByCode(code string) (*entity.CaseType, error) {
	var ct entity.CaseType
	err := r.db.Where("code = ?", code).First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *caseTypeRepository) Create(caseType *entity.CaseType) error {
	return r.db.Create(caseType).Error
}

func (r *caseTypeRepository) Update(caseType *entity.CaseType) error {
	return r.db.Save(caseType).Error
}

func (r *caseTypeRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.CaseType{}, "id = ?", id).Error
}
