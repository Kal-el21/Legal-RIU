package repository

import (
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgreementDocumentRepository interface {
	Create(*entity.AgreementDocument) error
	FindByID(uuid.UUID) (*entity.AgreementDocument, error)
	FindAll(*uuid.UUID, string, int, int) ([]entity.AgreementDocument, int64, error)
	Save(*entity.AgreementDocument) error
	Delete(uuid.UUID) error
	AddAttachment(*entity.AgreementAttachment) error
	FindAttachment(uuid.UUID, uuid.UUID) (*entity.AgreementAttachment, error)
	LatestUploadRound(uuid.UUID) (int, error)
	NextAgreementSequence(int) (int64, error)
	Complete(uuid.UUID, entity.SubmissionStatus, map[string]interface{}) (bool, error)
	GetActiveMaster() (*entity.AgreementCompanyMaster, error)
	SaveMaster(*entity.AgreementCompanyMaster) error
}

type agreementDocumentRepository struct{ db *gorm.DB }

func NewAgreementDocumentRepository(db *gorm.DB) AgreementDocumentRepository {
	return &agreementDocumentRepository{db}
}
func (r *agreementDocumentRepository) Create(v *entity.AgreementDocument) error {
	return r.db.Create(v).Error
}
func (r *agreementDocumentRepository) FindByID(id uuid.UUID) (*entity.AgreementDocument, error) {
	var v entity.AgreementDocument
	err := r.db.Preload("Requester").Preload("Approver").Preload("Attachments", func(db *gorm.DB) *gorm.DB { return db.Order("upload_round, created_at") }).First(&v, "id = ?", id).Error
	return &v, err
}
func (r *agreementDocumentRepository) FindAll(owner *uuid.UUID, status string, page, limit int) ([]entity.AgreementDocument, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	q := r.db.Model(&entity.AgreementDocument{})
	if owner != nil {
		q = q.Where("requester_id = ?", *owner)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []entity.AgreementDocument
	err := q.Preload("Requester").Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}
func (r *agreementDocumentRepository) Save(v *entity.AgreementDocument) error {
	return r.db.Save(v).Error
}
func (r *agreementDocumentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.AgreementDocument{}, "id = ?", id).Error
}
func (r *agreementDocumentRepository) AddAttachment(v *entity.AgreementAttachment) error {
	return r.db.Create(v).Error
}
func (r *agreementDocumentRepository) FindAttachment(docID, id uuid.UUID) (*entity.AgreementAttachment, error) {
	var v entity.AgreementAttachment
	err := r.db.First(&v, "id = ? AND agreement_document_id = ?", id, docID).Error
	return &v, err
}
func (r *agreementDocumentRepository) LatestUploadRound(id uuid.UUID) (int, error) {
	var n int
	err := r.db.Model(&entity.AgreementAttachment{}).Where("agreement_document_id = ?", id).Select("COALESCE(MAX(upload_round), 0)").Scan(&n).Error
	return n, err
}
func (r *agreementDocumentRepository) NextAgreementSequence(year int) (int64, error) {
	var n int64
	err := r.db.Model(&entity.AgreementDocument{}).Where("agreement_number LIKE ?", "%/"+time.Date(year, 1, 1, 0, 0, 0, 0, time.Local).Format("2006")).Count(&n).Error
	return n + 1, err
}
func (r *agreementDocumentRepository) Complete(id uuid.UUID, expected entity.SubmissionStatus, updates map[string]interface{}) (bool, error) {
	tx := r.db.Model(&entity.AgreementDocument{}).Where("id = ? AND status = ?", id, expected).Updates(updates)
	return tx.RowsAffected == 1, tx.Error
}
func (r *agreementDocumentRepository) GetActiveMaster() (*entity.AgreementCompanyMaster, error) {
	var v entity.AgreementCompanyMaster
	err := r.db.Where("is_active = ?", true).Order("created_at").First(&v).Error
	return &v, err
}
func (r *agreementDocumentRepository) SaveMaster(v *entity.AgreementCompanyMaster) error {
	return r.db.Save(v).Error
}
