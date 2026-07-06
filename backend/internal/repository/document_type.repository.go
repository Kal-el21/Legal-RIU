package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentTypeRepository interface {
	FindAll() ([]entity.DocumentType, error)
	FindByID(id uuid.UUID) (*entity.DocumentType, error)
	FindByName(name string) (*entity.DocumentType, error)
	Create(documentType *entity.DocumentType) error
	Update(documentType *entity.DocumentType) error
	Delete(id uuid.UUID) error
}

type documentTypeRepository struct {
	db *gorm.DB
}

func NewDocumentTypeRepository(db *gorm.DB) DocumentTypeRepository {
	return &documentTypeRepository{db: db}
}

func (r *documentTypeRepository) FindAll() ([]entity.DocumentType, error) {
	var items []entity.DocumentType
	err := r.db.Order("label ASC").Find(&items).Error
	return items, err
}

func (r *documentTypeRepository) FindByID(id uuid.UUID) (*entity.DocumentType, error) {
	var dt entity.DocumentType
	err := r.db.Where("id = ?", id).First(&dt).Error
	if err != nil {
		return nil, err
	}
	return &dt, nil
}

func (r *documentTypeRepository) FindByName(name string) (*entity.DocumentType, error) {
	var dt entity.DocumentType
	err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&dt).Error
	if err != nil {
		return nil, err
	}
	return &dt, nil
}

func (r *documentTypeRepository) Create(documentType *entity.DocumentType) error {
	return r.db.Create(documentType).Error
}

func (r *documentTypeRepository) Update(documentType *entity.DocumentType) error {
	return r.db.Save(documentType).Error
}

func (r *documentTypeRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.DocumentType{}, "id = ?", id).Error
}