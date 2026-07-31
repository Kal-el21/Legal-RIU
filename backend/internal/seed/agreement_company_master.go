package seed

import (
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
