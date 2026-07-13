package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"legal-riu-portal/internal/entity"
)

type TemplateFieldPositionRepository interface {
	GetByVersion(version string) ([]entity.TemplateFieldPosition, error)
	Upsert(version string, fields []entity.TemplateFieldPosition) error
	DeleteByVersion(version string) error
}

type templateFieldPositionRepository struct {
	db *gorm.DB
}

func NewTemplateFieldPositionRepository(db *gorm.DB) TemplateFieldPositionRepository {
	return &templateFieldPositionRepository{db: db}
}

func (r *templateFieldPositionRepository) GetByVersion(version string) ([]entity.TemplateFieldPosition, error) {
	var fields []entity.TemplateFieldPosition
	err := r.db.
		Where("template_version = ? AND deleted_at IS NULL", version).
		Order("page_number ASC, field_name ASC, occurrence_index ASC").
		Find(&fields).Error
	return fields, err
}

// Upsert replaces all field positions for a template version. The previous rows
// are hard-deleted (Unscoped) so the table stays clean and historical records
// are not preserved.
func (r *templateFieldPositionRepository) Upsert(version string, fields []entity.TemplateFieldPosition) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("template_version = ?", version).
			Delete(&entity.TemplateFieldPosition{}).Error; err != nil {
			return err
		}

		for i := range fields {
			fields[i].TemplateVersion = version
			if fields[i].ID == uuid.Nil {
				fields[i].ID = uuid.New()
			}
			fields[i].DeletedAt = gorm.DeletedAt{}
		}
		return tx.Create(&fields).Error
	})
}

func (r *templateFieldPositionRepository) DeleteByVersion(version string) error {
	return r.db.Unscoped().
		Where("template_version = ?", version).
		Delete(&entity.TemplateFieldPosition{}).Error
}
