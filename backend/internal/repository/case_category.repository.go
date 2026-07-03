package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseCategoryRepository interface {
	FindAll() ([]entity.CaseCategory, error)
	FindByID(id uuid.UUID) (*entity.CaseCategory, error)
	FindByCode(code string) (*entity.CaseCategory, error)
	Create(caseCategory *entity.CaseCategory) error
	Update(caseCategory *entity.CaseCategory) error
	Delete(id uuid.UUID) error
}

type caseCategoryRepository struct {
	db *gorm.DB
}

func NewCaseCategoryRepository(db *gorm.DB) CaseCategoryRepository {
	return &caseCategoryRepository{db: db}
}

func (r *caseCategoryRepository) FindAll() ([]entity.CaseCategory, error) {
	var items []entity.CaseCategory
	err := r.db.Order("label ASC").Find(&items).Error
	return items, err
}

func (r *caseCategoryRepository) FindByID(id uuid.UUID) (*entity.CaseCategory, error) {
	var cc entity.CaseCategory
	err := r.db.Where("id = ?", id).First(&cc).Error
	if err != nil {
		return nil, err
	}
	return &cc, nil
}

func (r *caseCategoryRepository) FindByCode(code string) (*entity.CaseCategory, error) {
	var cc entity.CaseCategory
	err := r.db.Where("code = ?", code).First(&cc).Error
	if err != nil {
		return nil, err
	}
	return &cc, nil
}

func (r *caseCategoryRepository) Create(caseCategory *entity.CaseCategory) error {
	return r.db.Create(caseCategory).Error
}

func (r *caseCategoryRepository) Update(caseCategory *entity.CaseCategory) error {
	return r.db.Save(caseCategory).Error
}

func (r *caseCategoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.CaseCategory{}, "id = ?", id).Error
}
