package seed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"slices"
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
	berkasTemplateBawaan  = "templates/pks-template/Draft-PKS-RIU.docx"
	berkasFieldBawaan     = "templates/pks-template/form-schema.json"
	// Definisi field PR 9 (termin 1 dan 2) untuk versi sebelum PR 9 yang belum
	// punya field; semuanya memakai template PR 8 yang disimpan di pr9/.
	berkasFieldPR9 = "templates/pks-template/pr9/form-schema.json"
)

// bacaSkema membaca nama jenis dan definisi field dari form-schema.json yang di-embed.
func bacaSkema(path string) (string, json.RawMessage, error) {
	raw, err := assets.Files.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	var schema struct {
		Name     string          `json:"name"`
		Sections json.RawMessage `json:"sections"`
	}
	err = json.Unmarshal(raw, &schema)
	return schema.Name, schema.Sections, err
}

// SeedAgreementTemplate memasang template PKS yang di-embed di aplikasi sebagai
// versi di database dan MinIO, lalu memindahkan pengajuan yang belum selesai dari
// template lama ke template itu. Aman dijalankan setiap server start: template
// bawaan dikenali dari checksum-nya, jadi hanya dipasang sekali per isi file.
func SeedAgreementTemplate(db *gorm.DB, store *storage.MinIOClient) error {
	data, err := assets.Files.ReadFile(berkasTemplateBawaan)
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

// pasangTemplateBawaan menambah template bawaan sebagai versi berikutnya, dengan
// definisi field bawaannya sendiri. Field kerja jenis PKS ikut diganti ke definisi
// itu hanya kalau belum diubah Admin, yaitu masih sama dengan field template bawaan
// yang sudah terpasang (keputusan D3). Versi baru langsung aktif kalau yang aktif juga
// template bawaan (atau belum ada yang aktif) dan field kerja tidak diubah Admin;
// template unggahan Admin tidak pernah diganti.
func pasangTemplateBawaan(db *gorm.DB, store *storage.MinIOClient, data []byte, checksum string) error {
	_, fields, err := bacaSkema(berkasFieldBawaan)
	if err != nil {
		return err
	}
	var aktif entity.AgreementTemplate
	err = db.Where("code = ? AND status = ?", agreementTemplateCode, entity.TemplateStatusActive).First(&aktif).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	aktifkan := errors.Is(err, gorm.ErrRecordNotFound) || aktif.UploadedBy == nil

	// Field kerja belum diubah Admin kalau masih sama dengan field salah satu
	// template bawaan. Jenis PKS yang belum terdaftar (database baru) juga belum.
	var jenis entity.AgreementType
	gantiKerja := true
	if err := db.Where("code = ?", agreementTemplateCode).First(&jenis).Error; err == nil && len(jenis.Fields) > 0 {
		var bawaan []entity.AgreementTemplate
		if err := db.Select("fields").Where("code = ? AND uploaded_by IS NULL", agreementTemplateCode).Find(&bawaan).Error; err != nil {
			return err
		}
		// Kedua kolom jsonb, jadi isi yang sama terbaca sebagai teks yang sama persis.
		gantiKerja = slices.ContainsFunc(bawaan, func(v entity.AgreementTemplate) bool { return bytes.Equal(jenis.Fields, v.Fields) })
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if !gantiKerja {
		aktifkan = false
		log.Printf("Isian kerja jenis %s sudah diubah Admin: template bawaan baru dipasang sebagai Draft, isian kerja tidak diganti", agreementTemplateCode)
	}

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
		Fields:   fields,
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
		if gantiKerja {
			if err := tx.Model(&entity.AgreementType{}).Where("code = ?", agreementTemplateCode).Update("fields", fields).Error; err != nil {
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
// Selesai tidak disentuh: dokumen finalnya sudah tersimpan. Sejak PR 10 template
// bawaan memakai termin berulang, tetapi Deploy 1 (PR 8 dan 9) sudah memindah
// semua pengajuan dari template lama, jadi di server tidak ada lagi yang dipindah.
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

// SeedAgreementTypes mendaftarkan jenis dokumen PKS dan mengisi definisi field
// yang belum ada. Field kerja Admin dimulai dari field bawaan. Versi template
// tanpa field adalah versi sebelum PR 9 (termasuk versi lama yang masih dikunci
// pengajuan berjalan), yang memakai template PR 8 dengan termin 1 dan 2, jadi
// diisi definisi field PR 9. Template bawaan membawa field-nya sejak dipasang.
// Yang sudah punya field tidak disentuh, jadi aman dijalankan berulang.
func SeedAgreementTypes(db *gorm.DB) error {
	name, bawaan, err := bacaSkema(berkasFieldBawaan)
	if err != nil {
		return err
	}
	_, pr9, err := bacaSkema(berkasFieldPR9)
	if err != nil {
		return err
	}
	jenis := entity.AgreementType{Code: agreementTemplateCode, Name: name, IsActive: true}
	if err := db.Where("code = ?", jenis.Code).FirstOrCreate(&jenis).Error; err != nil {
		return err
	}
	if err := db.Model(&entity.AgreementType{}).
		Where("code = ? AND fields IS NULL", agreementTemplateCode).
		Update("fields", bawaan).Error; err != nil {
		return err
	}
	return db.Model(&entity.AgreementTemplate{}).
		Where("code = ? AND fields IS NULL", agreementTemplateCode).
		Update("fields", pr9).Error
}

// TandaiNilaiUtamaPKS menandai nilai_kontrak sebagai nilai utama (PR 11, keputusan P1)
// di isian kerja PKS dan setiap versi template PKS yang belum punya tanda. Hanya tanda
// itu yang ditambahkan; isian pengajuan dan isi dokumen tidak berubah. Definisi yang
// sudah bertanda (di field mana pun) atau tanpa nilai_kontrak bertipe uang dibiarkan,
// jadi aman dijalankan berulang.
func TandaiNilaiUtamaPKS(db *gorm.DB) error {
	tandai := func(raw json.RawMessage) (json.RawMessage, bool) {
		var bagian []map[string]any
		if len(raw) == 0 || json.Unmarshal(raw, &bagian) != nil {
			return nil, false
		}
		var sasaran map[string]any
		for _, b := range bagian {
			fields, _ := b["fields"].([]any)
			for _, x := range fields {
				f, _ := x.(map[string]any)
				if f["primary_value"] == true {
					return nil, false
				}
				if f["name"] == "nilai_kontrak" && f["type"] == "money" {
					sasaran = f
				}
			}
		}
		if sasaran == nil {
			return nil, false
		}
		sasaran["primary_value"] = true
		out, e := json.Marshal(bagian)
		return out, e == nil
	}

	n := 0
	var jenis []entity.AgreementType
	if e := db.Where("code = ?", agreementTemplateCode).Find(&jenis).Error; e != nil {
		return e
	}
	for _, j := range jenis {
		if out, ok := tandai(j.Fields); ok {
			if e := db.Model(&entity.AgreementType{}).Where("id = ?", j.ID).UpdateColumn("fields", out).Error; e != nil {
				return e
			}
			n++
		}
	}
	var versi []entity.AgreementTemplate
	if e := db.Select("id", "fields").Where("code = ?", agreementTemplateCode).Find(&versi).Error; e != nil {
		return e
	}
	for _, v := range versi {
		// UpdateColumn: isi versi lain tidak berubah, jadi waktu ubahnya juga tidak.
		if out, ok := tandai(v.Fields); ok {
			if e := db.Model(&entity.AgreementTemplate{}).Where("id = ?", v.ID).UpdateColumn("fields", out).Error; e != nil {
				return e
			}
			n++
		}
	}
	if n > 0 {
		log.Printf("Nilai utama ditandai pada nilai_kontrak di %d definisi isian PKS (isian kerja dan versi template)", n)
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
