package repository

import (
	"errors"
	"strings"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ErrAgreementNumberTaken: nomor perjanjian sudah dipakai dokumen lain. Indeks
// unik di database juga mencakup dokumen yang sudah dihapus.
var ErrAgreementNumberTaken = errors.New("nomor perjanjian sudah dipakai dokumen lain")

type AgreementDocumentRepository interface {
	Create(*entity.AgreementDocument) error
	FindByID(uuid.UUID) (*entity.AgreementDocument, error)
	FindAll(*uuid.UUID, string, string, string, int, int) ([]entity.AgreementDocument, int64, error)
	Save(*entity.AgreementDocument) error
	AgreementNumberTaken(number string, except uuid.UUID) (bool, error)
	Delete(uuid.UUID) error
	AddAttachment(*entity.AgreementAttachment) error
	FindAttachment(uuid.UUID, uuid.UUID) (*entity.AgreementAttachment, error)
	LatestUploadRound(uuid.UUID) (int, error)
	NextTicketSequence(prefix string) (int, error)
	Complete(uuid.UUID, entity.SubmissionStatus, map[string]interface{}) (bool, error)
	GetActiveMaster() (*entity.AgreementCompanyMaster, error)
	SaveMaster(*entity.AgreementCompanyMaster) error
	ListSignatories() ([]entity.AgreementSignatory, error)
	FindSignatory(uuid.UUID) (*entity.AgreementSignatory, error)
	SaveSignatory(*entity.AgreementSignatory) error
	DeleteSignatory(uuid.UUID) error
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
	err := r.db.Preload("User").Preload("Approver").Preload("Attachments", func(db *gorm.DB) *gorm.DB { return db.Order("upload_round, created_at") }).First(&v, "id = ?", id).Error
	return &v, err
}
func (r *agreementDocumentRepository) FindAll(owner *uuid.UUID, status, dateFrom, search string, page, limit int) ([]entity.AgreementDocument, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	q := r.db.Model(&entity.AgreementDocument{})
	if owner != nil {
		q = q.Where("user_id = ?", *owner)
	}
	// Kolom disebut dengan tabelnya: saat dicari per nama, tabel users (yang juga punya
	// status dan created_at) ikut di-join.
	if status != "" {
		q = q.Where("agreement_documents.status = ?", status)
	} else {
		q = q.Scopes(bukanDraf("agreement_documents"))
	}
	if dateFrom != "" {
		q = q.Where("agreement_documents.created_at >= ?", dateFrom)
	}
	if search != "" {
		q = q.Joins("JOIN users u ON u.id = agreement_documents.user_id").Where("u.full_name ILIKE ?", "%"+search+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []entity.AgreementDocument
	err := q.Preload("User").Order("agreement_documents.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}
// Save menerjemahkan penolakan indeks unik nomor perjanjian, yang tetap bisa
// terjadi kalau dua orang menyimpan nomor yang sama hampir bersamaan.
func (r *agreementDocumentRepository) Save(v *entity.AgreementDocument) error {
	err := r.db.Save(v).Error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "agreement_number") {
		return ErrAgreementNumberTaken
	}
	return err
}

// AgreementNumberTaken memeriksa nomor sebelum disimpan, termasuk milik
// dokumen yang sudah dihapus, sama seperti indeks unik di database.
func (r *agreementDocumentRepository) AgreementNumberTaken(number string, except uuid.UUID) (bool, error) {
	var n int64
	err := r.db.Unscoped().Model(&entity.AgreementDocument{}).Where("agreement_number = ? AND id <> ?", number, except).Count(&n).Error
	return n > 0, err
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
func (r *agreementDocumentRepository) NextTicketSequence(prefix string) (int, error) {
	return nextTicketSequence(r.db, &entity.AgreementDocument{}, prefix)
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

// ListSignatories: Daftar Penandatangan, urut dari batas terkecil; tak terbatas di akhir.
func (r *agreementDocumentRepository) ListSignatories() ([]entity.AgreementSignatory, error) {
	var v []entity.AgreementSignatory
	err := r.db.Order("max_value ASC NULLS LAST, created_at").Find(&v).Error
	return v, err
}
func (r *agreementDocumentRepository) FindSignatory(id uuid.UUID) (*entity.AgreementSignatory, error) {
	var v entity.AgreementSignatory
	err := r.db.First(&v, "id = ?", id).Error
	return &v, err
}

// SaveSignatory menyimpan satu penandatangan. Penandatangan default hanya satu, jadi
// menjadikan satu baris default melepas default baris lain dalam transaksi yang sama.
func (r *agreementDocumentRepository) SaveSignatory(v *entity.AgreementSignatory) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if v.IsDefault {
			if err := tx.Model(&entity.AgreementSignatory{}).Where("id <> ? AND is_default", v.ID).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Save(v).Error
	})
}
func (r *agreementDocumentRepository) DeleteSignatory(id uuid.UUID) error {
	return r.db.Delete(&entity.AgreementSignatory{}, "id = ?", id).Error
}
