package repository

import (
	"errors"
	"legal-riu-portal/internal/entity"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentAccess menyatakan dokumen apa saja yang boleh dilihat si pemanggil di
// repositori. Repositori mengumpulkan file dari modul lain, jadi sebuah dokumen
// hanya terlihat oleh yang boleh membukanya di modul asalnya. Nilai kosong
// berarti tidak melihat apa pun.
type DocumentAccess struct {
	UserID uuid.UUID

	Cases       bool       // dokumen, foto, dan lampiran kronologi perkara
	CaseCompany *uuid.UUID // nil: perkara semua perusahaan

	// Jangkauan per jenis dokumen pengajuan: "" (tidak ada), ReachOwn, atau ReachAll.
	Opinions, Reviews                   string
	AgreementAttachments, AgreementPDFs string
	AgreementDOCX                       string
}

const (
	ReachOwn = "own" // hanya dari pengajuan miliknya
	ReachAll = "all"
)

const pdfMIME = "application/pdf"

// where menyusun syarat SQL untuk dokumen yang terlihat. Dokumen dari sumber
// yang tidak dikenali di sini tidak cocok dengan syarat mana pun, jadi tidak
// ditampilkan ke siapa pun.
func (a DocumentAccess) where() (string, []any) {
	conds, args := []string{"1 = 0"}, []any{}
	add := func(sql string, values ...any) {
		conds = append(conds, "("+sql+")")
		args = append(args, values...)
	}
	// kind memilih jenis dokumennya; owner adalah tabel pengajuan pemiliknya.
	reach := func(level, kind, owner string) {
		switch level {
		case ReachAll:
			add(kind)
		case ReachOwn:
			add(kind+" AND source_id IN (SELECT id::text FROM "+owner+" WHERE user_id = ?)", a.UserID)
		}
	}

	switch {
	case !a.Cases:
	case a.CaseCompany == nil:
		add("feature_code = 'case_management'")
	default:
		add(`feature_code = 'case_management' AND (
			(source_type <> 'case_chronologies' AND source_id IN (SELECT id::text FROM legal_cases WHERE company_id = ?))
			OR (source_type = 'case_chronologies' AND source_id IN (
				SELECT c.id::text FROM case_chronologies c JOIN legal_cases l ON l.id = c.case_id WHERE l.company_id = ?))
		)`, a.CaseCompany, a.CaseCompany)
	}
	reach(a.Opinions, "feature_code = 'legal_opinion'", "legal_opinions")
	reach(a.Reviews, "feature_code = 'document_review'", "document_reviews")
	reach(a.AgreementAttachments, "source_type = 'agreement_attachments'", "agreement_documents")
	reach(a.AgreementPDFs, "source_type = 'agreement_documents' AND mime_type = '"+pdfMIME+"'", "agreement_documents")
	reach(a.AgreementDOCX, "source_type = 'agreement_documents' AND mime_type <> '"+pdfMIME+"'", "agreement_documents")

	return "(" + strings.Join(conds, " OR ") + ")", args
}

type RepositoryDocumentRepository interface {
	FindAll(access DocumentAccess, featureCode string, search string) ([]entity.RepositoryDocument, error)
	FindByID(access DocumentAccess, id uuid.UUID) (*entity.RepositoryDocument, error)
	BatchUpsert(docs []entity.RepositoryDocument) error
	UpdateCaseMetadata(caseID string, sourceNumber string, sourceTitle string) error
	MarkInactiveBySource(sourceType string, sourceIDs []string) error
	UpsertBySource(doc entity.RepositoryDocument) error
}

type repositoryDocumentRepository struct {
	db *gorm.DB
}

func NewRepositoryDocumentRepository(db *gorm.DB) RepositoryDocumentRepository {
	return &repositoryDocumentRepository{db: db}
}

func (r *repositoryDocumentRepository) FindAll(access DocumentAccess, featureCode string, search string) ([]entity.RepositoryDocument, error) {
	var docs []entity.RepositoryDocument
	visible, args := access.where()
	query := r.db.Where("is_active = ?", true).Where(visible, args...)
	if featureCode != "" {
		query = query.Where("feature_code = ?", featureCode)
	}
	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}
	err := query.Order("created_at DESC").Find(&docs).Error
	return docs, err
}

func (r *repositoryDocumentRepository) FindByID(access DocumentAccess, id uuid.UUID) (*entity.RepositoryDocument, error) {
	var doc entity.RepositoryDocument
	visible, args := access.where()
	err := r.db.Where(visible, args...).First(&doc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// BatchUpsert refreshes source context without overwriting file metadata captured at upload.
func (r *repositoryDocumentRepository) BatchUpsert(docs []entity.RepositoryDocument) error {
	if len(docs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, doc := range docs {
			var existing entity.RepositoryDocument
			err := tx.Unscoped().
				Where("source_type = ? AND source_id = ? AND file_path = ?", doc.SourceType, doc.SourceID, doc.FilePath).
				First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if doc.CreatedAt.IsZero() {
					doc.CreatedAt = time.Now()
				}
				if doc.UploadedAt == nil {
					doc.UploadedAt = &doc.CreatedAt
				}
				if err := tx.Create(&doc).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
			uploadedAt := existing.UploadedAt
			if doc.UploadedAt != nil {
				uploadedAt = doc.UploadedAt
			} else if uploadedAt == nil {
				uploadedAt = &existing.CreatedAt
			}
			if err := tx.Unscoped().Model(&existing).Updates(map[string]interface{}{
				"source_number": doc.SourceNumber,
				"source_title":  doc.SourceTitle,
				"uploaded_at":   uploadedAt,
				"is_active":     true,
				"deleted_at":    nil,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repositoryDocumentRepository) UpdateCaseMetadata(caseID string, sourceNumber string, sourceTitle string) error {
	return r.db.Model(&entity.RepositoryDocument{}).
		Where(`(source_type IN ? AND source_id = ?)
			OR (source_type = 'case_chronologies' AND source_id IN (
				SELECT id::text FROM case_chronologies WHERE case_id = ?
			))`, []string{"legal_cases", "legal_cases_photo"}, caseID, caseID).
		Updates(map[string]interface{}{
			"source_number": sourceNumber,
			"source_title":  sourceTitle,
		}).Error
}

func (r *repositoryDocumentRepository) MarkInactiveBySource(sourceType string, sourceIDs []string) error {
	if sourceType == "all" {
		return r.db.Model(&entity.RepositoryDocument{}).Where("1 = 1").Update("is_active", false).Error
	}
	if len(sourceIDs) == 0 {
		return nil
	}
	return r.db.Model(&entity.RepositoryDocument{}).
		Where("source_type = ? AND source_id IN ?", sourceType, sourceIDs).
		Update("is_active", false).Error
}

func (r *repositoryDocumentRepository) UpsertBySource(doc entity.RepositoryDocument) error {
	var existing entity.RepositoryDocument
	err := r.db.Where("source_type = ? AND source_id = ? AND file_path = ?", doc.SourceType, doc.SourceID, doc.FilePath).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.Create(&doc).Error
		}
		return err
	}
	// Update existing record
	existing.Title = doc.Title
	existing.FileName = doc.FileName
	existing.MIMEType = doc.MIMEType
	existing.FileSize = doc.FileSize
	existing.UploaderType = doc.UploaderType
	existing.SourceNumber = doc.SourceNumber
	existing.SourceTitle = doc.SourceTitle
	if doc.UploadedAt != nil {
		existing.UploadedAt = doc.UploadedAt
	}
	existing.IsActive = true
	return r.db.Save(&existing).Error
}
