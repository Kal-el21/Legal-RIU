package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyMasterRepository interface {
	GetAll() ([]entity.CompanyMaster, error)
	GetByID(id uuid.UUID) (*entity.CompanyMaster, error)
	GetFirstActive() (*entity.CompanyMaster, error)
	Create(m *entity.CompanyMaster) error
	Update(m *entity.CompanyMaster) error
	Delete(id uuid.UUID) error
}

type companyMasterRepository struct {
	db *gorm.DB
}

func NewCompanyMasterRepository(db *gorm.DB) CompanyMasterRepository {
	return &companyMasterRepository{db: db}
}

func (r *companyMasterRepository) GetAll() ([]entity.CompanyMaster, error) {
	var items []entity.CompanyMaster
	err := r.db.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *companyMasterRepository) GetByID(id uuid.UUID) (*entity.CompanyMaster, error) {
	var m entity.CompanyMaster
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *companyMasterRepository) GetFirstActive() (*entity.CompanyMaster, error) {
	var m entity.CompanyMaster
	if err := r.db.Where("is_active = ?", true).Order("created_at ASC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *companyMasterRepository) Create(m *entity.CompanyMaster) error {
	return r.db.Create(m).Error
}

func (r *companyMasterRepository) Update(m *entity.CompanyMaster) error {
	return r.db.Save(m).Error
}

func (r *companyMasterRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.CompanyMaster{}, "id = ?", id).Error
}
