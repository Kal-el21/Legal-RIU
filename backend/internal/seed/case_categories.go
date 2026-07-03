package seed

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DEFAULT_CASE_CATEGORIES = []entity.CaseCategory{
	{Code: "Life", Label: "Life", IsActive: true},
	{Code: "BPPDAN", Label: "BPPDAN", IsActive: true},
	{Code: "Property", Label: "Property", IsActive: true},
	{Code: "COB", Label: "COB (IFRS)", IsActive: true},
}

func SeedCaseCategories(db *gorm.DB) error {
	for _, cc := range DEFAULT_CASE_CATEGORIES {
		var existing entity.CaseCategory
		if err := db.Where("code = ?", cc.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&entity.CaseCategory{
					Base: entity.Base{
						ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("case_category:"+cc.Code)),
					},
					Code:     cc.Code,
					Label:    cc.Label,
					IsActive: cc.IsActive,
				}).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}
