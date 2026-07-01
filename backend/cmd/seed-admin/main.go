package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/seed"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db := config.InitDatabase(cfg)

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	if err := seed.PrepareLegalCasePICMigration(db); err != nil {
		log.Fatalf("Legal case PIC migration preparation failed: %v", err)
	}
	if err := db.AutoMigrate(
		&entity.Division{},
		&entity.User{},
		&entity.RefreshToken{},
		&entity.LegalOpinion{},
		&entity.LegalOpinionAttachment{},
		&entity.LegalOpinionResult{},
		&entity.DocumentReview{},
		&entity.DocumentReviewAttachment{},
		&entity.DocumentReviewResult{},
		&entity.Regency{},
		&entity.Cedant{},
		&entity.LegalCase{},
		&entity.CaseChronology{},
		&entity.AuditLog{},
		&entity.NotificationSetting{},
		&entity.UserSettings{},
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migrated successfully")

	if err := seed.SeedRegencies(db); err != nil {
		log.Fatalf("Regency seed failed: %v", err)
	}
	log.Println("Regencies seeded successfully")

	if err := seed.SeedDivisions(db); err != nil {
		log.Fatalf("Division seed failed: %v", err)
	}
	log.Println("Divisions seeded successfully")

	if err := seed.BackfillUserDivisionIDs(db); err != nil {
		log.Fatalf("User division backfill failed: %v", err)
	}

	if err := seed.SeedNotificationSettings(db); err != nil {
		log.Fatalf("Notification settings seed failed: %v", err)
	}
	log.Println("Notification settings seeded successfully")

	email := getEnv("ADMIN_EMAIL", "admin@example.com")
	password := getEnv("ADMIN_PASSWORD", "12345678")
	adminDivision := getEnv("ADMIN_DIVISION", "Legal, compliance and risk management division")

	admin := entity.User{
		FullName: getEnv("ADMIN_FULL_NAME", "Super Admin"),
		Email:    email,
		AuthType: entity.AuthTypeLocal,
		Position: getEnv("ADMIN_POSITION", "Administrator"),
		Division: adminDivision,
		Role:     entity.RoleAdmin,
		Status:   entity.UserActive,
	}
	if divisionID, err := seed.FindDivisionIDByName(db, adminDivision); err == nil {
		admin.DivisionID = &divisionID
	}

	var existing entity.User
	err = db.Where("email = ?", email).First(&existing).Error
	if err == nil {
		updates := map[string]interface{}{}
		if existing.Role != entity.RoleAdmin {
			updates["role"] = entity.RoleAdmin
		}
		if existing.Status != entity.UserActive {
			updates["status"] = entity.UserActive
		}
		if existing.AuthType != entity.AuthTypeLocal {
			updates["auth_type"] = entity.AuthTypeLocal
		}
		if admin.DivisionID != nil && (existing.DivisionID == nil || *existing.DivisionID != *admin.DivisionID) {
			updates["division"] = admin.Division
			updates["division_id"] = admin.DivisionID
		}

		if len(updates) > 0 {
			if err := db.Model(&existing).Updates(updates).Error; err != nil {
				log.Fatalf("Failed to update existing admin user: %v", err)
			}
			log.Printf("Admin user %s already exists and was updated", email)
			return
		}

		log.Printf("Admin user %s already exists", email)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("Failed to check admin user: %v", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}
	admin.PasswordHash = string(passwordHash)

	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("Admin user %s created successfully", email)
}

func requiredEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
