package repository

import (
	"fmt"
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LegalOpinionRepository interface {
	Create(lo *entity.LegalOpinion) error
	FindByID(id uuid.UUID) (*entity.LegalOpinion, error)
	FindAll(userID *uuid.UUID, status string, page, limit int) ([]entity.LegalOpinion, int64, error)
	Update(lo *entity.LegalOpinion) error
	Delete(id uuid.UUID) error
	UpdateStatus(id uuid.UUID, status entity.SubmissionStatus, adminNote string) error
	CountByMonthAndPrefix(prefix string) (int64, error)
	AddAttachment(att *entity.LegalOpinionAttachment) error
	GetLatestUploadRound(loID uuid.UUID) (int, error)
	AddResult(result *entity.LegalOpinionResult) error
}

type legalOpinionRepository struct {
	db *gorm.DB
}

func NewLegalOpinionRepository(db *gorm.DB) LegalOpinionRepository {
	return &legalOpinionRepository{db: db}
}

func (r *legalOpinionRepository) Create(lo *entity.LegalOpinion) error {
	return r.db.Create(lo).Error
}

func (r *legalOpinionRepository) FindByID(id uuid.UUID) (*entity.LegalOpinion, error) {
	var lo entity.LegalOpinion
	err := r.db.
		Preload("Attachments", func(db *gorm.DB) *gorm.DB {
			return db.Order("upload_round ASC, created_at ASC")
		}).
		Preload("Results", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("User").
		Where("id = ?", id).
		First(&lo).Error
	if err != nil {
		return nil, err
	}
	return &lo, nil
}

func (r *legalOpinionRepository) FindAll(userID *uuid.UUID, status string, page, limit int) ([]entity.LegalOpinion, int64, error) {
	var items []entity.LegalOpinion
	var total int64

	query := r.db.Model(&entity.LegalOpinion{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.
		Preload("User").
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&items).Error

	return items, total, err
}

func (r *legalOpinionRepository) Update(lo *entity.LegalOpinion) error {
	return r.db.Save(lo).Error
}

func (r *legalOpinionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.LegalOpinion{}, "id = ?", id).Error
}

func (r *legalOpinionRepository) UpdateStatus(id uuid.UUID, status entity.SubmissionStatus, adminNote string) error {
	updates := map[string]interface{}{"status": status}
	if adminNote != "" {
		updates["admin_note"] = adminNote
	}
	return r.db.Model(&entity.LegalOpinion{}).Where("id = ?", id).Updates(updates).Error
}

func (r *legalOpinionRepository) CountByMonthAndPrefix(prefix string) (int64, error) {
	var count int64
	now := time.Now()
	start := fmt.Sprintf("%s-%s-", prefix, now.Format("200601"))
	err := r.db.Model(&entity.LegalOpinion{}).
		Where("ticket_number LIKE ? AND created_at >= ?", start+"%",
			time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())).
		Count(&count).Error
	return count, err
}

func (r *legalOpinionRepository) AddAttachment(att *entity.LegalOpinionAttachment) error {
	return r.db.Create(att).Error
}

func (r *legalOpinionRepository) GetLatestUploadRound(loID uuid.UUID) (int, error) {
	var maxRound int
	err := r.db.Model(&entity.LegalOpinionAttachment{}).
		Where("legal_opinion_id = ?", loID).
		Select("COALESCE(MAX(upload_round), 0)").
		Scan(&maxRound).Error
	return maxRound, err
}

func (r *legalOpinionRepository) AddResult(result *entity.LegalOpinionResult) error {
	return r.db.Create(result).Error
}
