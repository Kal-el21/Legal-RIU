package repository

import (
	"errors"
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RepositoryDocumentRepository interface {
	FindAll(featureCode string, search string) ([]entity.RepositoryDocument, error)
	FindByID(id uuid.UUID) (*entity.RepositoryDocument, error)
	BatchUpsert(docs []entity.RepositoryDocument) error
	MarkInactiveBySource(sourceType string, sourceIDs []string) error
	UpsertBySource(doc entity.RepositoryDocument) error
}

type repositoryDocumentRepository struct {
	db *gorm.DB
}

func NewRepositoryDocumentRepository(db *gorm.DB) RepositoryDocumentRepository {
	return &repositoryDocumentRepository{db: db}
}

func (r *repositoryDocumentRepository) FindAll(featureCode string, search string) ([]entity.RepositoryDocument, error) {
	var docs []entity.RepositoryDocument
	query := r.db.Where("is_active = ?", true)
	if featureCode != "" {
		query = query.Where("feature_code = ?", featureCode)
	}
	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}
	err := query.Order("created_at DESC").Find(&docs).Error
	return docs, err
}

func (r *repositoryDocumentRepository) FindByID(id uuid.UUID) (*entity.RepositoryDocument, error) {
	var doc entity.RepositoryDocument
	err := r.db.First(&doc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *repositoryDocumentRepository) BatchUpsert(docs []entity.RepositoryDocument) error {
	if len(docs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, doc := range docs {
			if err := tx.Save(&doc).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
	existing.IsActive = true
	return r.db.Save(&existing).Error
}