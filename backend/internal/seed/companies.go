package seed

import (
	"errors"
	"fmt"
	"legal-riu-portal/internal/entity"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Hanya induk yang internal. User perusahaan internal melihat perkara semua
// perusahaan; user anak usaha hanya perkara perusahaannya sendiri. Aturannya ada
// di docs/case_company_access.md.
var DEFAULT_COMPANIES = []entity.Company{
	{Name: "PT Reasuransi Indonesia Utama", EmailDomain: entity.ParentCompanyDomain, IsInternal: true, IsParent: true},
	{Name: "PT Asuransi ASEI Indonesia", EmailDomain: "asei.co.id", IsInternal: false},
	{Name: "PT Reasuransi Syariah Indonesia", EmailDomain: "resyariah.co.id", IsInternal: false},
}

// SeedCompanies membuat perusahaan bawaan yang belum ada. Perusahaan yang sudah
// ada tidak disentuh: tanda internalnya diatur ADMIN lewat menu, jadi di
// database lama anak usaha tetap internal sampai centangnya dicabut di sana.
func SeedCompanies(db *gorm.DB) error {
	for _, c := range DEFAULT_COMPANIES {
		expectedID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("company:"+c.EmailDomain))

		// Unscoped: perusahaan bawaan yang sudah dihapus lewat menu tetap dianggap
		// ada. Kalau tidak, seed mencoba membuatnya lagi dengan ID yang sama dan
		// server gagal start.
		var existing entity.Company
		err := db.Unscoped().Where("id = ? OR email_domain = ?", expectedID, c.EmailDomain).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := db.Create(&entity.Company{
			Base:        entity.Base{ID: expectedID},
			Name:        c.Name,
			EmailDomain: c.EmailDomain,
			IsInternal:  c.IsInternal,
			IsParent:    c.IsParent,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// BackfillParentCompany menandai induk bawaan di database yang dibuat sebelum
// kolom is_parent ada. Hanya berjalan kalau belum ada induk, supaya induk yang
// sudah diganti admin tidak dikembalikan setiap server dinyalakan ulang.
func BackfillParentCompany(db *gorm.DB) error {
	return db.Exec(`
		UPDATE companies SET is_parent = true
		WHERE email_domain = ? AND deleted_at IS NULL
			AND NOT EXISTS (SELECT 1 FROM companies WHERE is_parent)
	`, entity.ParentCompanyDomain).Error
}

func FindCompanyIDByDomain(db *gorm.DB, email string) (uuid.UUID, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return uuid.Nil, fmt.Errorf("invalid email format")
	}
	domain := parts[1]

	var company entity.Company
	if err := db.Where("email_domain = ?", domain).First(&company).Error; err != nil {
		return uuid.Nil, err
	}
	return company.ID, nil
}

// FindFirstCompanyID mengembalikan perusahaan cadangan untuk user yang surelnya
// tidak cocok dengan perusahaan mana pun. Induk didahulukan, lalu perusahaan
// internal, supaya admin bawaan tidak terdaftar di anak usaha.
func FindFirstCompanyID(db *gorm.DB) (uuid.UUID, error) {
	var company entity.Company
	if err := db.Order("is_parent DESC, is_internal DESC").First(&company).Error; err != nil {
		return uuid.Nil, err
	}
	return company.ID, nil
}
