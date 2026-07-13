package repository

import (
	"fmt"
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgreementDocumentRepository interface {
	Create(doc *entity.AgreementDocument) error
	FindByID(id uuid.UUID) (*entity.AgreementDocument, error)
	FindAll(userID *uuid.UUID, status string, page, limit int) ([]entity.AgreementDocument, int64, error)
	Update(doc *entity.AgreementDocument) error
	Delete(id uuid.UUID) error
	UpdateStatus(id uuid.UUID, status entity.SubmissionStatus, adminNote string) error
	CountByMonthAndPrefix(prefix string) (int64, error)
	AddAttachment(att *entity.AgreementAttachment) error
	GetLatestUploadRound(docID uuid.UUID) (int, error)
	GetCompanyMaster(id uuid.UUID) (*entity.CompanyMaster, error)
	GetFirstActiveCompanyMaster() (*entity.CompanyMaster, error)
	UpdateFields(id uuid.UUID, fields map[string]interface{}) error
}

type agreementDocumentRepository struct {
	db *gorm.DB
}

func NewAgreementDocumentRepository(db *gorm.DB) AgreementDocumentRepository {
	return &agreementDocumentRepository{db: db}
}

func (r *agreementDocumentRepository) Create(doc *entity.AgreementDocument) error {
	return r.db.Create(doc).Error
}

func (r *agreementDocumentRepository) FindByID(id uuid.UUID) (*entity.AgreementDocument, error) {
	var doc entity.AgreementDocument
	err := r.db.
		Preload("Attachments", func(db *gorm.DB) *gorm.DB {
			return db.Order("upload_round ASC, created_at ASC")
		}).
		Preload("User").
		Preload("PihakPertama").
		Where("id = ?", id).
		First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *agreementDocumentRepository) FindAll(userID *uuid.UUID, status string, page, limit int) ([]entity.AgreementDocument, int64, error) {
	var items []entity.AgreementDocument
	var total int64

	query := r.db.Model(&entity.AgreementDocument{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.
		Preload("User").
		Preload("PihakPertama").
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&items).Error

	return items, total, err
}

func (r *agreementDocumentRepository) Update(doc *entity.AgreementDocument) error {
	return r.db.Save(doc).Error
}

func (r *agreementDocumentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.AgreementDocument{}, "id = ?", id).Error
}

func (r *agreementDocumentRepository) UpdateStatus(id uuid.UUID, status entity.SubmissionStatus, adminNote string) error {
	now := time.Now()
	updates := map[string]interface{}{"status": status, "status_updated_at": now}
	if adminNote != "" {
		updates["admin_note"] = adminNote
	}
	return r.db.Model(&entity.AgreementDocument{}).Where("id = ?", id).Updates(updates).Error
}

func (r *agreementDocumentRepository) CountByMonthAndPrefix(prefix string) (int64, error) {
	var count int64
	now := time.Now()
	start := fmt.Sprintf("%s-%s-", prefix, now.Format("200601"))
	err := r.db.Model(&entity.AgreementDocument{}).
		Where("ticket_number LIKE ? AND created_at >= ?", start+"%",
			time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())).
		Count(&count).Error
	return count, err
}

func (r *agreementDocumentRepository) AddAttachment(att *entity.AgreementAttachment) error {
	return r.db.Create(att).Error
}

func (r *agreementDocumentRepository) GetLatestUploadRound(docID uuid.UUID) (int, error) {
	var maxRound int
	err := r.db.Model(&entity.AgreementAttachment{}).
		Where("agreement_id = ?", docID).
		Select("COALESCE(MAX(upload_round), 0)").
		Scan(&maxRound).Error
	return maxRound, err
}

func (r *agreementDocumentRepository) GetCompanyMaster(id uuid.UUID) (*entity.CompanyMaster, error) {
	var m entity.CompanyMaster
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *agreementDocumentRepository) GetFirstActiveCompanyMaster() (*entity.CompanyMaster, error) {
	var m entity.CompanyMaster
	if err := r.db.Where("is_active = ?", true).Order("created_at ASC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *agreementDocumentRepository) UpdateFields(id uuid.UUID, fields map[string]interface{}) error {
	return r.db.Model(&entity.AgreementDocument{}).Where("id = ?", id).Updates(fields).Error
}
