package seed

import (
	"log"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedNotificationSettings(db *gorm.DB) error {
	var count int64
	if err := db.Model(&entity.NotificationSetting{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Println("Notification settings already seeded")
		return nil
	}

	settings := []entity.NotificationSetting{
		{SubmissionType: "legal_opinion", WarningLevel: "YELLOW", DaysThreshold: 3, IsActive: true},
		{SubmissionType: "legal_opinion", WarningLevel: "RED", DaysThreshold: 14, IsActive: true},
		{SubmissionType: "document_review", WarningLevel: "YELLOW", DaysThreshold: 3, IsActive: true},
		{SubmissionType: "document_review", WarningLevel: "RED", DaysThreshold: 14, IsActive: true},
	}

	for i := range settings {
		settings[i].ID = uuid.New()
	}

	if err := db.Create(&settings).Error; err != nil {
		return err
	}

	log.Println("Notification settings seeded successfully")
	return nil
}
