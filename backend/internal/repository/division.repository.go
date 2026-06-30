package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DivisionRepository interface {
	FindAll(search string, limit int) ([]entity.Division, error)
	FindByID(id uuid.UUID) (*entity.Division, error)
	FindByName(name string) (*entity.Division, error)
	Create(division *entity.Division) error
	Update(division *entity.Division) error
	Delete(id uuid.UUID) error
}

type divisionRepository struct {
	db *gorm.DB
}

func NewDivisionRepository(db *gorm.DB) DivisionRepository {
	return &divisionRepository{db: db}
}

func (r *divisionRepository) FindAll(search string, limit int) ([]entity.Division, error) {
	var items []entity.Division
	query := r.db.Model(&entity.Division{})
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *divisionRepository) FindByID(id uuid.UUID) (*entity.Division, error) {
	var div entity.Division
	err := r.db.Where("id = ?", id).First(&div).Error
	if err != nil {
		return nil, err
	}
	return &div, nil
}

func (r *divisionRepository) FindByName(name string) (*entity.Division, error) {
	var div entity.Division
	err := r.db.Where("name = ?", name).First(&div).Error
	if err != nil {
		return nil, err
	}
	return &div, nil
}

func (r *divisionRepository) Create(division *entity.Division) error {
	return r.db.Create(division).Error
}

func (r *divisionRepository) Update(division *entity.Division) error {
	return r.db.Save(division).Error
}

func (r *divisionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.Division{}, "id = ?", id).Error
}
