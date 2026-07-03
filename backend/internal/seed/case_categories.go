package seed

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DEFAULT_CASE_CATEGORIES = []entity.CaseCategory{
	{Code: "Life", Label: "Life", IsActive: true},
	{Code: "BPPDAN", Label: "BPPDAN", IsActive: true},
	{Code: "Property", Label: "Property", IsActive: true},
	{Code: "COB", Label: "COB (IFRS)", IsActive: true},
}

func SeedCaseCategories(db *gorm.DB) error {
	items := make([]entity.CaseCategory, 0, len(DEFAULT_CASE_CATEGORIES))
	for _, cc := range DEFAULT_CASE_CATEGORIES {
		items = append(items, entity.CaseCategory{
			Base: entity.Base{
				ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("case_category:"+cc.Code)),
			},
			Code:     cc.Code,
			Label:    cc.Label,
			IsActive: cc.IsActive,
		})
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}, {Name: "label"}},
		DoNothing: true,
	}).CreateInBatches(items, 100).Error
}
