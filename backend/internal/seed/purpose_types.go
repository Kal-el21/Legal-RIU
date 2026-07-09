package seed

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DEFAULT_PURPOSE_TYPES = []entity.PurposeType{
	{Name: "Konsultan Hukum", Description: "Konsultan hukum eksternal", IsActive: true},
	{Name: "Pengacara", Description: "Pengacara litigasi/non-litigasi", IsActive: true},
	{Name: "Jaksa Pengacara Negara", Description: "JPU dari kejaksaan", IsActive: true},
}

func SeedPurposeTypes(db *gorm.DB) error {
	for _, pt := range DEFAULT_PURPOSE_TYPES {
		expectedID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("purpose_type:"+pt.Name))

	var existing entity.PurposeType
	if err := db.Where("id = ?", expectedID).First(&existing).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.Where("name = ?", pt.Name).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&entity.PurposeType{
			Base: entity.Base{
				ID: expectedID,
			},
			Name:        pt.Name,
			Description: pt.Description,
			IsActive:    pt.IsActive,
		}).Error; err != nil {
			return err
		}
	}
	}
	return nil
}
