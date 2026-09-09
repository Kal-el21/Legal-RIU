package seed

import (
	"context"
	"errors"
	"time"

	"legal-riu-portal/internal/assets"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	agreementTemplateCode = "PKS"
	agreementTemplateMIME = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

// SeedAgreementTemplate memindahkan template PKS yang selama ini di-embed ke
// dalam binary menjadi versi 1 di database dan MinIO. Template embed tetap
// dipertahankan sebagai sumber seed ini, bukan dihapus.
func SeedAgreementTemplate(db *gorm.DB, store *storage.MinIOClient) error {
	var count int64
	if err := db.Model(&entity.AgreementTemplate{}).Where("code = ?", agreementTemplateCode).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return backfillAgreementDocumentTemplate(db)
	}

	data, err := assets.Files.ReadFile("templates/pks-template/Draft-PKS-RIU.docx")
	if err != nil {
		return err
	}

	now := time.Now()
	tpl := &entity.AgreementTemplate{
		Code:        agreementTemplateCode,
		Version:     1,
		Name:        "PKS - Template Bawaan",
		FileName:    "Draft-PKS-RIU.docx",
		Checksum:    service.TemplateChecksum(data),
		Status:      entity.TemplateStatusActive,
		IsLegacy:    true,
		Note:        "Template bawaan hasil migrasi dari binary aplikasi.",
		ActivatedAt: &now,
	}
	tpl.ID = uuid.New()
	tpl.FilePath = "agreement-templates/" + agreementTemplateCode + "/" + tpl.ID.String() + ".docx"

	if err := store.UploadBytes(context.Background(), tpl.FilePath, data, agreementTemplateMIME); err != nil {
		return err
	}
	if err := db.Create(tpl).Error; err != nil {
		return err
	}
	return backfillAgreementDocumentTemplate(db)
}

// backfillAgreementDocumentTemplate mengarahkan dokumen lama ke versi 1, karena
// itulah template yang sesungguhnya dipakai saat dokumen tersebut dibuat.
func backfillAgreementDocumentTemplate(db *gorm.DB) error {
	var tpl entity.AgreementTemplate
	err := db.Where("code = ? AND version = ?", agreementTemplateCode, 1).First(&tpl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return db.Model(&entity.AgreementDocument{}).
		Where("template_id IS NULL").
		Update("template_id", tpl.ID).Error
}
