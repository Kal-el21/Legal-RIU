package repository

import (
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LegalCaseFilter struct {
	Search    string
	Status    string
	CaseType  string
	Level     string
	CompanyID *uuid.UUID
	DateFrom  *time.Time
	DateTo    *time.Time
	Page      int
	Limit     int
}

type LegalCaseRepository interface {
	Create(legalCase *entity.LegalCase) error
	FindAll(filter LegalCaseFilter) ([]entity.LegalCase, int64, error)
	FindLatest() (*entity.LegalCase, error)
	FindByID(id uuid.UUID) (*entity.LegalCase, error)
	Update(legalCase *entity.LegalCase) error
	UpdateStatus(id uuid.UUID, status string, statusUpdatedAt *time.Time) error
	Delete(id uuid.UUID) error

	ListChronologies(caseID uuid.UUID) ([]entity.CaseChronology, error)
	FindChronology(caseID uuid.UUID, chronologyID uuid.UUID) (*entity.CaseChronology, error)
	CreateChronology(chronology *entity.CaseChronology) error
	UpdateChronology(chronology *entity.CaseChronology) error
	DeleteChronology(caseID uuid.UUID, chronologyID uuid.UUID) error

	ListRegencies(search string, limit int) ([]entity.Regency, error)
	FindRegencyByID(id uuid.UUID) (*entity.Regency, error)

	ListCedants(search string, limit int) ([]entity.Cedant, error)
	FindCedantByID(id uuid.UUID) (*entity.Cedant, error)
	CreateCedant(cedant *entity.Cedant) error
	UpdateCedant(cedant *entity.Cedant) error
	DeleteCedant(id uuid.UUID) error

	FindDivisionByID(id uuid.UUID) (*entity.Division, error)
	FindCaseTypeByID(id uuid.UUID) (*entity.CaseType, error)
	FindCaseCategoryByID(id uuid.UUID) (*entity.CaseCategory, error)
}

type legalCaseRepository struct {
	db *gorm.DB
}

func NewLegalCaseRepository(db *gorm.DB) LegalCaseRepository {
	return &legalCaseRepository{db: db}
}

func (r *legalCaseRepository) Create(legalCase *entity.LegalCase) error {
	return r.db.Create(legalCase).Error
}

func (r *legalCaseRepository) FindAll(filter LegalCaseFilter) ([]entity.LegalCase, int64, error) {
	var items []entity.LegalCase
	var total int64

	query := r.db.Model(&entity.LegalCase{})
	query = applyLegalCaseFilter(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, limit := normalizePageLimit(filter.Page, filter.Limit)
	err := query.
		Preload("RelatedParty").
		Preload("LocationRegency").
		Preload("PICDivision").
		Preload("CaseTypeRef").
		Preload("CategoryRef").
		Preload("Company").
		Order("case_date DESC, updated_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&items).Error

	return items, total, err
}

func (r *legalCaseRepository) FindLatest() (*entity.LegalCase, error) {
	var legalCase entity.LegalCase
	err := r.db.
		Preload("RelatedParty").
		Preload("LocationRegency").
		Preload("PICDivision").
		Preload("CaseTypeRef").
		Preload("CategoryRef").
		Preload("Company").
		Order("case_date DESC, updated_at DESC").
		First(&legalCase).Error
	if err != nil {
		return nil, err
	}
	return &legalCase, nil
}

func (r *legalCaseRepository) FindByID(id uuid.UUID) (*entity.LegalCase, error) {
	var legalCase entity.LegalCase
	err := r.db.
		Preload("RelatedParty").
		Preload("LocationRegency").
		Preload("PICDivision").
		Preload("CaseTypeRef").
		Preload("CategoryRef").
		Preload("Company").
		Preload("Chronologies", func(db *gorm.DB) *gorm.DB {
			return db.Order("agenda_date DESC, updated_at DESC")
		}).
		Where("id = ?", id).
		First(&legalCase).Error
	if err != nil {
		return nil, err
	}
	return &legalCase, nil
}

func (r *legalCaseRepository) Update(legalCase *entity.LegalCase) error {
	return r.db.Save(legalCase).Error
}

func (r *legalCaseRepository) UpdateStatus(id uuid.UUID, status string, statusUpdatedAt *time.Time) error {
	updates := map[string]interface{}{"current_status": status}
	if statusUpdatedAt != nil {
		updates["status_updated_at"] = statusUpdatedAt
	}
	return r.db.Model(&entity.LegalCase{}).Where("id = ?", id).Updates(updates).Error
}

func (r *legalCaseRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.LegalCase{}, "id = ?", id).Error
}

func (r *legalCaseRepository) ListChronologies(caseID uuid.UUID) ([]entity.CaseChronology, error) {
	var items []entity.CaseChronology
	err := r.db.
		Where("case_id = ?", caseID).
		Order("agenda_date DESC, updated_at DESC").
		Find(&items).Error
	return items, err
}

func (r *legalCaseRepository) FindChronology(caseID uuid.UUID, chronologyID uuid.UUID) (*entity.CaseChronology, error) {
	var chronology entity.CaseChronology
	err := r.db.
		Where("case_id = ? AND id = ?", caseID, chronologyID).
		First(&chronology).Error
	if err != nil {
		return nil, err
	}
	return &chronology, nil
}

func (r *legalCaseRepository) CreateChronology(chronology *entity.CaseChronology) error {
	return r.db.Omit("LegalCase").Create(chronology).Error
}

func (r *legalCaseRepository) UpdateChronology(chronology *entity.CaseChronology) error {
	return r.db.Omit("LegalCase").Save(chronology).Error
}

func (r *legalCaseRepository) DeleteChronology(caseID uuid.UUID, chronologyID uuid.UUID) error {
	return r.db.Delete(&entity.CaseChronology{}, "case_id = ? AND id = ?", caseID, chronologyID).Error
}

func (r *legalCaseRepository) ListRegencies(search string, limit int) ([]entity.Regency, error) {
	var items []entity.Regency
	query := r.db.Model(&entity.Regency{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR province ILIKE ?", like, like)
	}
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	err := query.Order("province ASC, name ASC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *legalCaseRepository) FindRegencyByID(id uuid.UUID) (*entity.Regency, error) {
	var regency entity.Regency
	err := r.db.Where("id = ?", id).First(&regency).Error
	if err != nil {
		return nil, err
	}
	return &regency, nil
}

func (r *legalCaseRepository) ListCedants(search string, limit int) ([]entity.Cedant, error) {
	var items []entity.Cedant
	query := r.db.Model(&entity.Cedant{})
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	err := query.Order("name ASC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *legalCaseRepository) FindCedantByID(id uuid.UUID) (*entity.Cedant, error) {
	var cedant entity.Cedant
	err := r.db.Where("id = ?", id).First(&cedant).Error
	if err != nil {
		return nil, err
	}
	return &cedant, nil
}

func (r *legalCaseRepository) CreateCedant(cedant *entity.Cedant) error {
	return r.db.Create(cedant).Error
}

func (r *legalCaseRepository) UpdateCedant(cedant *entity.Cedant) error {
	return r.db.Save(cedant).Error
}

func (r *legalCaseRepository) DeleteCedant(id uuid.UUID) error {
	return r.db.Delete(&entity.Cedant{}, "id = ?", id).Error
}

func (r *legalCaseRepository) FindDivisionByID(id uuid.UUID) (*entity.Division, error) {
	var division entity.Division
	err := r.db.Where("id = ?", id).First(&division).Error
	if err != nil {
		return nil, err
	}
	return &division, nil
}

func (r *legalCaseRepository) FindCaseTypeByID(id uuid.UUID) (*entity.CaseType, error) {
	var ct entity.CaseType
	err := r.db.Where("id = ?", id).First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *legalCaseRepository) FindCaseCategoryByID(id uuid.UUID) (*entity.CaseCategory, error) {
	var cc entity.CaseCategory
	err := r.db.Where("id = ?", id).First(&cc).Error
	if err != nil {
		return nil, err
	}
	return &cc, nil
}

func applyLegalCaseFilter(query *gorm.DB, filter LegalCaseFilter) *gorm.DB {
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("case_name ILIKE ? OR case_summary ILIKE ?", like, like)
	}
	if filter.Status != "" {
		query = query.Where("current_status = ?", filter.Status)
	}
	if filter.CompanyID != nil {
		query = query.Where("company_id = ?", *filter.CompanyID)
	}
	if filter.CaseType != "" {
		query = query.Joins("JOIN case_types ON case_types.id = legal_cases.case_type_id").Where("case_types.code = ?", filter.CaseType)
	}
	if filter.Level != "" {
		query = query.Where("level = ?", filter.Level)
	}
	if filter.DateFrom != nil {
		query = query.Where("case_date >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("case_date < ?", filter.DateTo.AddDate(0, 0, 1))
	}
	return query
}

func normalizePageLimit(page int, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
