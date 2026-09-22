package repository

import (
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgreementTemplateRepository interface {
	Create(*entity.AgreementTemplate) error
	Save(*entity.AgreementTemplate) error
	FindByID(uuid.UUID) (*entity.AgreementTemplate, error)
	FindByCode(string) ([]entity.AgreementTemplate, error)
	FindActive(string) (*entity.AgreementTemplate, error)
	NextVersion(string) (int, error)
	Activate(uuid.UUID, string, uuid.UUID) error
	Delete(uuid.UUID) error
}

type agreementTemplateRepository struct{ db *gorm.DB }

func NewAgreementTemplateRepository(db *gorm.DB) AgreementTemplateRepository {
	return &agreementTemplateRepository{db}
}

func (r *agreementTemplateRepository) Create(v *entity.AgreementTemplate) error {
	return r.db.Create(v).Error
}

func (r *agreementTemplateRepository) Save(v *entity.AgreementTemplate) error {
	return r.db.Save(v).Error
}

func (r *agreementTemplateRepository) FindByID(id uuid.UUID) (*entity.AgreementTemplate, error) {
	var v entity.AgreementTemplate
	err := r.db.Preload("Uploader").First(&v, "id = ?", id).Error
	return &v, err
}

func (r *agreementTemplateRepository) FindByCode(code string) ([]entity.AgreementTemplate, error) {
	var items []entity.AgreementTemplate
	err := r.db.Preload("Uploader").Where("code = ?", code).Order("version DESC").Find(&items).Error
	return items, err
}

func (r *agreementTemplateRepository) FindActive(code string) (*entity.AgreementTemplate, error) {
	var v entity.AgreementTemplate
	err := r.db.Where("code = ? AND status = ?", code, entity.TemplateStatusActive).Order("version DESC").First(&v).Error
	return &v, err
}

func (r *agreementTemplateRepository) NextVersion(code string) (int, error) {
	var max *int
	err := r.db.Model(&entity.AgreementTemplate{}).Where("code = ?", code).Select("MAX(version)").Scan(&max).Error
	if err != nil {
		return 0, err
	}
	if max == nil {
		return 1, nil
	}
	return *max + 1, nil
}

// Activate memindahkan versi aktif sebelumnya ke ARCHIVED dan mengaktifkan versi
// yang diminta dalam satu transaksi, sehingga tidak pernah ada dua versi aktif
// untuk satu kode dokumen.
func (r *agreementTemplateRepository) Activate(id uuid.UUID, code string, activatedBy uuid.UUID) error {
	now := time.Now()
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.AgreementTemplate{}).
			Where("code = ? AND status = ? AND id <> ?", code, entity.TemplateStatusActive, id).
			Update("status", entity.TemplateStatusArchived).Error; err != nil {
			return err
		}
		return tx.Model(&entity.AgreementTemplate{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":       entity.TemplateStatusActive,
				"activated_by": activatedBy,
				"activated_at": now,
			}).Error
	})
}

func (r *agreementTemplateRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.AgreementTemplate{}, "id = ?", id).Error
}
