package repository

import (
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
	NextTicketSequence(prefix string) (int, error)
	AddAttachment(att *entity.LegalOpinionAttachment) error
	GetLatestUploadRound(loID uuid.UUID) (int, error)
	AddResult(result *entity.LegalOpinionResult) error
	FindFileByPath(path string) (id uuid.UUID, fileName string, isResult bool, err error)
	HardDelete(id uuid.UUID) error
	DeleteAttachment(id uuid.UUID) error
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
	} else {
		query = query.Scopes(bukanDraf("legal_opinions"))
	}

	// Draf diurutkan menurut terakhir diubah.
	urutan := "created_at DESC"
	if status == string(entity.StatusDraft) {
		urutan = "updated_at DESC"
	}
	query.Count(&total)
	err := query.
		Preload("User").
		Order(urutan).
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
	now := time.Now()
	updates := map[string]interface{}{"status": status, "status_updated_at": now}
	if adminNote != "" {
		updates["admin_note"] = adminNote
	}
	return r.db.Model(&entity.LegalOpinion{}).Where("id = ?", id).Updates(updates).Error
}

func (r *legalOpinionRepository) NextTicketSequence(prefix string) (int, error) {
	return nextTicketSequence(r.db, &entity.LegalOpinion{}, prefix)
}

// HardDelete menghapus permanen pengajuan beserta baris lampiran dan hasilnya. Hanya untuk
// draf (PR Draf); pengajuan yang sudah dikirim tetap soft delete lewat Delete.
func (r *legalOpinionRepository) HardDelete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("legal_opinion_id = ?", id).Delete(&entity.LegalOpinionAttachment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("legal_opinion_id = ?", id).Delete(&entity.LegalOpinionResult{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&entity.LegalOpinion{}, "id = ?", id).Error
	})
}

func (r *legalOpinionRepository) DeleteAttachment(id uuid.UUID) error {
	return r.db.Delete(&entity.LegalOpinionAttachment{}, "id = ?", id).Error
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

// FindFileByPath mengembalikan pengajuan pemilik sebuah file beserta nama aslinya, dari
// lampiran (isResult false) maupun hasil kajiannya (isResult true). File yang tidak
// terdaftar di modul ini tidak ditemukan.
func (r *legalOpinionRepository) FindFileByPath(path string) (uuid.UUID, string, bool, error) {
	var att entity.LegalOpinionAttachment
	if err := r.db.Where("file_path = ?", path).First(&att).Error; err == nil {
		return att.LegalOpinionID, att.FileName, false, nil
	}
	var res entity.LegalOpinionResult
	if err := r.db.Where("file_path = ?", path).First(&res).Error; err != nil {
		return uuid.Nil, "", false, err
	}
	return res.LegalOpinionID, res.FileName, true, nil
}
