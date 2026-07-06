package repository

import (
	"fmt"
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentReviewRepository interface {
	Create(dr *entity.DocumentReview) error
	FindByID(id uuid.UUID) (*entity.DocumentReview, error)
	FindAll(userID *uuid.UUID, status string, page, limit int) ([]entity.DocumentReview, int64, error)
	Update(dr *entity.DocumentReview) error
	Delete(id uuid.UUID) error
	UpdateStatus(id uuid.UUID, status entity.SubmissionStatus, adminNote string) error
	CountByMonthAndPrefix(prefix string) (int64, error)
	AddAttachment(att *entity.DocumentReviewAttachment) error
	GetLatestUploadRound(drID uuid.UUID) (int, error)
	AddResult(result *entity.DocumentReviewResult) error
}

type documentReviewRepository struct {
	db *gorm.DB
}

func NewDocumentReviewRepository(db *gorm.DB) DocumentReviewRepository {
	return &documentReviewRepository{db: db}
}

func (r *documentReviewRepository) Create(dr *entity.DocumentReview) error {
	return r.db.Create(dr).Error
}

func (r *documentReviewRepository) FindByID(id uuid.UUID) (*entity.DocumentReview, error) {
	var dr entity.DocumentReview
	err := r.db.
		Preload("Attachments", func(db *gorm.DB) *gorm.DB {
			return db.Order("upload_round ASC, created_at ASC")
		}).
		Preload("Results", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("User").
		Where("id = ?", id).
		First(&dr).Error
	if err != nil {
		return nil, err
	}
	return &dr, nil
}

func (r *documentReviewRepository) FindAll(userID *uuid.UUID, status string, page, limit int) ([]entity.DocumentReview, int64, error) {
	var items []entity.DocumentReview
	var total int64

	query := r.db.Model(&entity.DocumentReview{})
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

func (r *documentReviewRepository) Update(dr *entity.DocumentReview) error {
	return r.db.Save(dr).Error
}

func (r *documentReviewRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.DocumentReview{}, "id = ?", id).Error
}

func (r *documentReviewRepository) UpdateStatus(id uuid.UUID, status entity.SubmissionStatus, adminNote string) error {
	now := time.Now()
	updates := map[string]interface{}{"status": status, "status_updated_at": now}
	if adminNote != "" {
		updates["admin_note"] = adminNote
	}
	return r.db.Model(&entity.DocumentReview{}).Where("id = ?", id).Updates(updates).Error
}

func (r *documentReviewRepository) CountByMonthAndPrefix(prefix string) (int64, error) {
	var count int64
	now := time.Now()
	start := fmt.Sprintf("%s-%s-", prefix, now.Format("200601"))
	err := r.db.Model(&entity.DocumentReview{}).
		Where("ticket_number LIKE ? AND created_at >= ?", start+"%",
			time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())).
		Count(&count).Error
	return count, err
}

func (r *documentReviewRepository) AddAttachment(att *entity.DocumentReviewAttachment) error {
	return r.db.Create(att).Error
}

func (r *documentReviewRepository) GetLatestUploadRound(drID uuid.UUID) (int, error) {
	var maxRound int
	err := r.db.Model(&entity.DocumentReviewAttachment{}).
		Where("document_review_id = ?", drID).
		Select("COALESCE(MAX(upload_round), 0)").
		Scan(&maxRound).Error
	return maxRound, err
}

func (r *documentReviewRepository) AddResult(result *entity.DocumentReviewResult) error {
	return r.db.Create(result).Error
}
