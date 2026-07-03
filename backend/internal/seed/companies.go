package seed

import (
	"fmt"
	"legal-riu-portal/internal/entity"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DEFAULT_COMPANIES = []entity.Company{
	{Name: "PT Reasuransi Indonesia Utama", EmailDomain: "indonesiare.co.id", IsInternal: true},
	{Name: "PT Asuransi ASEI Indonesia", EmailDomain: "asei.co.id", IsInternal: true},
	{Name: "PT Reasuransi Syariah Indonesia", EmailDomain: "resyariah.co.id", IsInternal: true},
}

func SeedCompanies(db *gorm.DB) error {
	companies := make([]entity.Company, 0, len(DEFAULT_COMPANIES))
	for _, c := range DEFAULT_COMPANIES {
		companies = append(companies, entity.Company{
			Base: entity.Base{
				ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("company:"+c.EmailDomain)),
			},
			Name:        c.Name,
			EmailDomain: c.EmailDomain,
			IsInternal:  c.IsInternal,
		})
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).CreateInBatches(companies, 100).Error
}

func FindCompanyIDByDomain(db *gorm.DB, email string) (uuid.UUID, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return uuid.Nil, fmt.Errorf("invalid email format")
	}
	domain := parts[1]

	var company entity.Company
	if err := db.Where("email_domain = ?", domain).First(&company).Error; err != nil {
		return uuid.Nil, err
	}
	return company.ID, nil
}

func FindFirstCompanyID(db *gorm.DB) (uuid.UUID, error) {
	var company entity.Company
	if err := db.First(&company).Error; err != nil {
		return uuid.Nil, err
	}
	return company.ID, nil
}
