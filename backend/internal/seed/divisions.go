package seed

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DEFAULT_DIVISIONS = []entity.Division{
	{Name: "Finance division", Description: "Finance division"},
	{Name: "Aktuaris perusahaan", Description: "Aktuaris perusahaan"},
	{Name: "Accounting division", Description: "Accounting division"},
	{Name: "Client market & treaty P&C division", Description: "Client market & treaty P&C division"},
	{Name: "Business management division", Description: "Business management division"},
	{Name: "Reinsurance & product underwriting P&C division", Description: "Reinsurance & product underwriting P&C division"},
	{Name: "Client market & pricing actuary L&H division", Description: "Client market & pricing actuary L&H division"},
	{Name: "Underwriting center & risk engineering dept", Description: "Underwriting center & risk engineering dept"},
	{Name: "Legal, compliance and risk management division", Description: "Legal, compliance and risk management division"},
	{Name: "Corporate secretary division", Description: "Corporate secretary division"},
	{Name: "TJSL & ESG division", Description: "TJSL & ESG division"},
	{Name: "Human capital & general affair division", Description: "Human capital & general affair division"},
	{Name: "Information technology division", Description: "Information technology division"},
	{Name: "Strategic development division", Description: "Strategic development division"},
	{Name: "Indonesia Re Institute", Description: "Indonesia Re Institute"},
	{Name: "Internal auditor", Description: "Internal auditor"},
	{Name: "Corporate transformation management office", Description: "Corporate transformation management office"},
}

func SeedDivisions(db *gorm.DB) error {
	divisions := make([]entity.Division, 0, len(DEFAULT_DIVISIONS))
	for _, d := range DEFAULT_DIVISIONS {
		divisions = append(divisions, entity.Division{
			Base: entity.Base{
				ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("division:"+d.Name)),
			},
			Name:        d.Name,
			Description: d.Description,
		})
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).CreateInBatches(divisions, 100).Error
}

func FindDivisionIDByName(db *gorm.DB, name string) (uuid.UUID, error) {
	var div entity.Division
	if err := db.Where("LOWER(name) = LOWER(?)", name).First(&div).Error; err != nil {
		return uuid.Nil, err
	}
	return div.ID, nil
}

func BackfillUserDivisionIDs(db *gorm.DB) error {
	var users []entity.User
	if err := db.Where("division_id IS NULL AND division <> ''").Find(&users).Error; err != nil {
		return err
	}

	for i := range users {
		var division entity.Division
		if err := db.Where("LOWER(name) = LOWER(?)", users[i].Division).First(&division).Error; err != nil {
			continue
		}
		if division.ID == uuid.Nil {
			continue
		}
		if err := db.Model(&entity.User{}).Where("id = ?", users[i].ID).Updates(map[string]interface{}{
			"division":    division.Name,
			"division_id": division.ID,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
