package repository

import (
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLogFilters struct {
	Page       int
	Limit      int
	Action     *string
	EntityType *string
	EntityID   *uuid.UUID
	UserID     *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
	Search     *string
}

type AuditLogRepository interface {
	Create(log *entity.AuditLog) error
	GetAll(filters AuditLogFilters) ([]entity.AuditLog, int64, error)
	GetByEntity(entityType string, entityID uuid.UUID, limit int) ([]entity.AuditLog, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *entity.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) GetAll(filters AuditLogFilters) ([]entity.AuditLog, int64, error) {
	var items []entity.AuditLog
	var total int64

	query := r.db.Model(&entity.AuditLog{}).Preload("User")
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.Action != nil && *filters.Action != "" {
		query = query.Where("action = ?", *filters.Action)
	}
	if filters.EntityType != nil && *filters.EntityType != "" {
		query = query.Where("entity_type = ?", *filters.EntityType)
	}
	if filters.EntityID != nil {
		query = query.Where("entity_id = ?", *filters.EntityID)
	}
	if filters.DateFrom != nil {
		query = query.Where("created_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		query = query.Where("created_at <= ?", *filters.DateTo)
	}
	if filters.Search != nil && *filters.Search != "" {
		search := "%" + *filters.Search + "%"
		query = query.Where("description ILIKE ? OR old_value::text ILIKE ? OR new_value::text ILIKE ?", search, search, search)
	}

	query.Count(&total)
	err := query.
		Order("created_at DESC").
		Offset((filters.Page - 1) * filters.Limit).
		Limit(filters.Limit).
		Find(&items).Error

	return items, total, err
}

func (r *auditLogRepository) GetByEntity(entityType string, entityID uuid.UUID, limit int) ([]entity.AuditLog, error) {
	var items []entity.AuditLog
	err := r.db.
		Preload("User").
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}
