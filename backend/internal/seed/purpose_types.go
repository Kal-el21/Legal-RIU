package seed

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DEFAULT_PURPOSE_TYPES = []entity.PurposeType{
	{Name: "Konsultan Hukum", Description: "Konsultan hukum eksternal", IsActive: true},
	{Name: "Pengacara", Description: "Pengacara litigasi/non-litigasi", IsActive: true},
	{Name: "Jaksa Pengacara Negara", Description: "JPU dari kejaksaan", IsActive: true},
}

func SeedPurposeTypes(db *gorm.DB) error {
	items := make([]entity.PurposeType, 0, len(DEFAULT_PURPOSE_TYPES))
	for _, pt := range DEFAULT_PURPOSE_TYPES {
		items = append(items, entity.PurposeType{
			Base: entity.Base{
				ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("purpose_type:"+pt.Name)),
			},
			Name:        pt.Name,
			Description: pt.Description,
			IsActive:    pt.IsActive,
		})
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).CreateInBatches(items, 100).Error
}
