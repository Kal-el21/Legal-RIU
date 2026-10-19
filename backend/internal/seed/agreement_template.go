package seed

import (
	"context"
	"errors"
	"log"
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

// SeedAgreementTemplate memasang template PKS yang di-embed di aplikasi sebagai
// versi di database dan MinIO, lalu memindahkan pengajuan yang belum selesai dari
// template lama ke template itu. Aman dijalankan setiap server start: template
// bawaan dikenali dari checksum-nya, jadi hanya dipasang sekali per isi file.
func SeedAgreementTemplate(db *gorm.DB, store *storage.MinIOClient) error {
	data, err := assets.Files.ReadFile("templates/pks-template/Draft-PKS-RIU.docx")
	if err != nil {
		return err
	}
	checksum := service.TemplateChecksum(data)

	var terpasang int64
	if err := db.Model(&entity.AgreementTemplate{}).
		Where("code = ? AND checksum = ?", agreementTemplateCode, checksum).
		Count(&terpasang).Error; err != nil {
		return err
	}
	if terpasang == 0 {
		if err := pasangTemplateBawaan(db, store, data, checksum); err != nil {
			return err
		}
	}
	if err := backfillAgreementDocumentTemplate(db); err != nil {
		return err
	}
	return pindahkanDariTemplateLama(db, checksum)
}

// pasangTemplateBawaan menambah template bawaan sebagai versi berikutnya. Versi
// itu langsung aktif kalau yang aktif juga template bawaan (atau belum ada yang
// aktif); template unggahan Admin tidak pernah diganti.
func pasangTemplateBawaan(db *gorm.DB, store *storage.MinIOClient, data []byte, checksum string) error {
	var aktif entity.AgreementTemplate
	err := db.Where("code = ? AND status = ?", agreementTemplateCode, entity.TemplateStatusActive).First(&aktif).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	aktifkan := errors.Is(err, gorm.ErrRecordNotFound) || aktif.UploadedBy == nil

	var versi int
	if err := db.Model(&entity.AgreementTemplate{}).
		Where("code = ?", agreementTemplateCode).
		Select("COALESCE(MAX(version), 0) + 1").
		Scan(&versi).Error; err != nil {
		return err
	}

	tpl := &entity.AgreementTemplate{
		Code:     agreementTemplateCode,
		Version:  versi,
		Name:     "PKS - Template Bawaan",
		FileName: "Draft-PKS-RIU.docx",
		Checksum: checksum,
		Status:   entity.TemplateStatusDraft,
		Note:     "Template bawaan aplikasi.",
	}
	if aktifkan {
		now := time.Now()
		tpl.Status, tpl.ActivatedAt = entity.TemplateStatusActive, &now
	}
	tpl.ID = uuid.New()
	tpl.FilePath = "agreement-templates/" + agreementTemplateCode + "/" + tpl.ID.String() + ".docx"

	if err := store.UploadBytes(context.Background(), tpl.FilePath, data, agreementTemplateMIME); err != nil {
		return err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if aktifkan {
			if err := tx.Model(&entity.AgreementTemplate{}).
				Where("code = ? AND status = ?", agreementTemplateCode, entity.TemplateStatusActive).
				Update("status", entity.TemplateStatusArchived).Error; err != nil {
				return err
			}
		}
		return tx.Create(tpl).Error
	})
	if err != nil {
		_ = store.DeleteFile(context.Background(), tpl.FilePath)
	}
	return err
}

// pindahkanDariTemplateLama mengarahkan pengajuan yang belum selesai dari
// template lama (diisi lewat pencocokan kalimat) ke template bawaan bercheksum
// itu. Isi keduanya sama (dijaga TestIsiSamaDenganTemplateLama), dan template
// lama tidak bisa dipakai lagi setelah kode pengisiannya dihapus. Pengajuan
// Selesai tidak disentuh: dokumen finalnya sudah tersimpan.
func pindahkanDariTemplateLama(db *gorm.DB, checksum string) error {
	res := db.Exec(`
		UPDATE agreement_documents SET
			template_id = (
				SELECT id FROM agreement_templates
				WHERE code = ? AND checksum = ? AND deleted_at IS NULL
				ORDER BY version DESC LIMIT 1
			),
			template_checksum = ?
		WHERE status <> ?
			AND template_id IN (SELECT id FROM agreement_templates WHERE is_legacy)
	`, agreementTemplateCode, checksum, checksum, entity.StatusCompleted)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		log.Printf("%d pengajuan perjanjian dipindah dari template lama ke template bawaan versi baru", res.RowsAffected)
	}
	return nil
}

// backfillAgreementDocumentTemplate mengarahkan dokumen tanpa template ke versi 1,
// karena itulah template yang sesungguhnya dipakai saat dokumen tersebut dibuat.
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
