package seed

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	items := make([]entity.CaseType, 0, len(DEFAULT_CASE_TYPES))
	for _, ct := range DEFAULT_CASE_TYPES {
		items = append(items, entity.CaseType{
			Base: entity.Base{
				ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("case_type:"+ct.Code)),
			},
			Code:     ct.Code,
			Label:    ct.Label,
			IsActive: ct.IsActive,
		})
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}, {Name: "label"}},
		DoNothing: true,
	}).CreateInBatches(items, 100).Error
}
