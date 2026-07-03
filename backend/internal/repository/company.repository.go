package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyRepository interface {
	FindAll() ([]entity.Company, error)
	FindByID(id uuid.UUID) (*entity.Company, error)
	FindByDomain(domain string) (*entity.Company, error)
	Create(company *entity.Company) error
	Update(company *entity.Company) error
	Delete(id uuid.UUID) error
}

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) FindAll() ([]entity.Company, error) {
	var items []entity.Company
	err := r.db.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *companyRepository) FindByID(id uuid.UUID) (*entity.Company, error) {
	var company entity.Company
	err := r.db.Where("id = ?", id).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) FindByDomain(domain string) (*entity.Company, error) {
	var company entity.Company
	err := r.db.Where("LOWER(email_domain) = LOWER(?)", domain).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) Create(company *entity.Company) error {
	return r.db.Create(company).Error
}

func (r *companyRepository) Update(company *entity.Company) error {
	return r.db.Save(company).Error
}

func (r *companyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.Company{}, "id = ?", id).Error
}
