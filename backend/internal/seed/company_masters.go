package seed

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DEFAULT_COMPANY_MASTERS = []entity.CompanyMaster{
	{
		Name:              "PT Reasuransi Indonesia Utama (Persero)",
		Address:           "Jalan Salemba Raya Nomor 30, Jakarta Pusat",
		NPWP:              "01.000.000.0-000.000",
		Phone:             "021-3920101",
		Email:             "info@indonesiare.co.id",
		DefaultPejabat:    "",
		DefaultJabatan:    "",
		DefaultTempatTtd:  "Jakarta Pusat",
		IsActive:          true,
	},
}

func SeedCompanyMasters(db *gorm.DB) error {
	for _, m := range DEFAULT_COMPANY_MASTERS {
		expectedID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("company_master:"+m.Email))

		var existing entity.CompanyMaster
		if err := db.Where("id = ?", expectedID).First(&existing).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			if err := db.Where("email = ?", m.Email).First(&existing).Error; err == nil {
				continue
			}
			if err := db.Create(&entity.CompanyMaster{
				Base: entity.Base{
					ID: expectedID,
				},
				Name:     m.Name,
				Address:  m.Address,
				NPWP:     m.NPWP,
				Phone:    m.Phone,
				Email:    m.Email,
				IsActive: m.IsActive,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
