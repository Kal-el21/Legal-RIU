package seed

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DEFAULT_DOCUMENT_TYPES = []entity.DocumentType{
	{Name: "surat-perintah-kerja", Label: "Surat Perintah Kerja", IsActive: true},
	{Name: "perjanjian-kerjasama-non-teknik", Label: "Perjanjian Kerjasama Non Teknik", IsActive: true},
	{Name: "kontrak-treaty", Label: "Kontrak Treaty", IsActive: true},
	{Name: "kontrak-retro", Label: "Kontrak Retro", IsActive: true},
	{Name: "pembatalan-perjanjian", Label: "Pembatalan Perjanjian", IsActive: true},
	{Name: "nota-kesepahaman", Label: "Nota Kesepahaman", IsActive: true},
	{Name: "surat", Label: "Surat", IsActive: true},
	{Name: "lain-lain", Label: "Lain-Lain", IsActive: true},
}

func SeedDocumentTypes(db *gorm.DB) error {
	for _, dt := range DEFAULT_DOCUMENT_TYPES {
		expectedID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("document_type:"+dt.Name))

	var existing entity.DocumentType
	if err := db.Where("id = ?", expectedID).First(&existing).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.Where("name = ?", dt.Name).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&entity.DocumentType{
			Base: entity.Base{
				ID: expectedID,
			},
			Name:     dt.Name,
			Label:    dt.Label,
			IsActive: dt.IsActive,
		}).Error; err != nil {
			return err
		}
	}
	}
	return nil
}