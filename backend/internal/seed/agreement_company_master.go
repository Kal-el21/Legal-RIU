package seed

import (
	"errors"
	"log"

	"gorm.io/gorm"
	"legal-riu-portal/internal/entity"
)

func SeedAgreementCompanyMaster(db *gorm.DB) error {
	var n int64
	if e := db.Model(&entity.AgreementCompanyMaster{}).Where("is_active = ?", true).Count(&n).Error; e != nil {
		return e
	}
	if n > 0 {
		return nil
	}
	return db.Create(&entity.AgreementCompanyMaster{Name: "PT Reasuransi Indonesia Utama (Persero)", Address: "Jl. Salemba Raya No. 30, Kenari, Senen, Jakarta Pusat", Phone: "021 3920101", DefaultSignatoryName: "-", DefaultSignatoryPosition: "-", DefaultSigningPlace: "Jakarta", IsActive: true}).Error
}

// SeedAgreementSignatories membuat Daftar Penandatangan (PR 11) dari penandatangan
// tunggal master Pihak Pertama: satu baris tak terbatas dan default, jadi pengajuan
// tetap memakai penandatangan yang sama sampai Admin mengisi daftarnya. Hanya kalau
// tabelnya belum pernah berisi (termasuk baris terhapus), jadi aman dijalankan
// berulang dan tidak mengembalikan baris yang dihapus Admin.
func SeedAgreementSignatories(db *gorm.DB) error {
	var n int64
	if e := db.Unscoped().Model(&entity.AgreementSignatory{}).Count(&n).Error; e != nil || n > 0 {
		return e
	}
	var m entity.AgreementCompanyMaster
	if e := db.Where("is_active = ?", true).Order("created_at").First(&m).Error; e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil
		}
		return e
	}
	p := entity.AgreementSignatory{Name: m.DefaultSignatoryName, Position: m.DefaultSignatoryPosition, IsDefault: true, IsActive: true}
	if e := db.Create(&p).Error; e != nil {
		return e
	}
	log.Printf("Daftar Penandatangan dibuat dari penandatangan master Pihak Pertama: %s, %s (tak terbatas, default)", p.Name, p.Position)
	return nil
}
