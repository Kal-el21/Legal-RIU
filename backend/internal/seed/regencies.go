package seed

import (
	_ "embed"
	"strings"
	"unicode"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:embed data/kabupaten.md
var regencyMarkdown string

func SeedRegencies(db *gorm.DB) error {
	regencies := parseRegencies(regencyMarkdown)
	if len(regencies) == 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).CreateInBatches(regencies, 100).Error
}

func parseRegencies(markdown string) []entity.Regency {
	var regencies []entity.Regency
	currentProvince := ""

	for _, rawLine := range strings.Split(markdown, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		if strings.Contains(line, "Daftar Nama Kabupaten") && strings.Contains(line, "Provinsi ") {
			currentProvince = extractProvince(line)
			continue
		}

		if currentProvince == "" {
			continue
		}

		name, ok := extractNumberedName(line)
		if !ok {
			continue
		}
		if !strings.HasPrefix(name, "Kabupaten ") && !strings.HasPrefix(name, "Kota ") {
			continue
		}

		regencyType := "KABUPATEN"
		if strings.HasPrefix(name, "Kota ") {
			regencyType = "KOTA"
		}

		regencies = append(regencies, entity.Regency{
			Base: entity.Base{
				ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("regency:"+currentProvince+":"+name)),
			},
			Name:     name,
			Province: currentProvince,
			Type:     regencyType,
		})
	}

	return regencies
}

func extractProvince(line string) string {
	idx := strings.Index(line, "Provinsi ")
	if idx < 0 {
		return ""
	}

	province := strings.TrimSpace(line[idx+len("Provinsi "):])
	province = strings.TrimSuffix(province, ":")
	if colon := strings.Index(province, ":"); colon >= 0 {
		province = province[:colon]
	}
	if paren := strings.Index(province, "("); paren >= 0 {
		province = province[:paren]
	}

	return strings.TrimSpace(province)
}

func extractNumberedName(line string) (string, bool) {
	dot := strings.Index(line, ".")
	if dot <= 0 {
		return "", false
	}

	for _, r := range line[:dot] {
		if !unicode.IsDigit(r) {
			return "", false
		}
	}

	name := strings.TrimSpace(line[dot+1:])
	return name, name != ""
}
