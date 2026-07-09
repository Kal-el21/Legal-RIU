package seed

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DEFAULT_CASE_TYPES = []entity.CaseType{
	{Code: "NON_LITIGASI", Label: "Non-Litigasi", IsActive: true},
	{Code: "PERDATA", Label: "Perdata", IsActive: true},
	{Code: "PIDANA", Label: "Pidana", IsActive: true},
	{Code: "TIPEKOR", Label: "Tipikor", IsActive: true},
	{Code: "ARBITRASE", Label: "Arbitrase", IsActive: true},
	{Code: "TUN", Label: "TUN", IsActive: true},
}

func SeedCaseTypes(db *gorm.DB) error {
	for _, ct := range DEFAULT_CASE_TYPES {
		expectedID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("case_type:"+ct.Code))

	var existing entity.CaseType
	if err := db.Where("id = ?", expectedID).First(&existing).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.Where("code = ?", ct.Code).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&entity.CaseType{
			Base: entity.Base{
				ID: expectedID,
			},
			Code:     ct.Code,
			Label:    ct.Label,
			IsActive: ct.IsActive,
		}).Error; err != nil {
			return err
		}
	}
	}
	return nil
}
