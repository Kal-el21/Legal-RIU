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
	Pelaku(cari string, limit int) ([]entity.User, error)
}

// pelakuTermasukTerhapus: pelaku dimuat juga kalau akunnya sudah dihapus, karena jejak
// audit justru paling dibutuhkan untuk akun yang sudah tidak ada (PR Audit Log tahap L.1).
func pelakuTermasukTerhapus(db *gorm.DB) *gorm.DB { return db.Unscoped() }

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

	query := r.db.Model(&entity.AuditLog{}).Preload("User", pelakuTermasukTerhapus)
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
		// Termasuk nama dan surel pelaku.
		query = query.Where(`(description ILIKE ? OR old_value::text ILIKE ? OR new_value::text ILIKE ?
			OR user_id IN (SELECT id FROM users WHERE full_name ILIKE ? OR email ILIKE ?))`,
			search, search, search, search, search)
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
		Preload("User", pelakuTermasukTerhapus).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

// Pelaku: user yang pernah tercatat di audit log dan nama atau surelnya cocok, termasuk
// yang sudah dihapus. Untuk saringan pelaku di halaman Audit Log, supaya LEGAL tidak
// butuh akses ke Kelola User.
func (r *auditLogRepository) Pelaku(cari string, limit int) ([]entity.User, error) {
	var users []entity.User
	query := r.db.Unscoped().Select("id", "full_name", "email", "role").
		Where("id IN (SELECT DISTINCT user_id FROM audit_logs)")
	if cari != "" {
		like := "%" + cari + "%"
		query = query.Where("(full_name ILIKE ? OR email ILIKE ?)", like, like)
	}
	err := query.Order("full_name").Limit(limit).Find(&users).Error
	return users, err
}
