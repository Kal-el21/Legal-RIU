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

	if err := db.Migrator().DropTable(&entity.CaseChronology{}, &entity.LegalCase{}); err != nil {
		log.Fatalf("Failed to drop legal_case tables: %v", err)
	}
	log.Println("Legal case tables recreated to accommodate PIC -> uuid migration")
	if err := db.AutoMigrate(
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
		&entity.Division{},
		&entity.LegalCase{},
		&entity.CaseChronology{},
		&entity.AuditLog{},
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

	email := getEnv("ADMIN_EMAIL", "admin@example.com")
	password := getEnv("ADMIN_PASSWORD", "12345678")

	admin := entity.User{
		FullName: getEnv("ADMIN_FULL_NAME", "Super Admin"),
		Email:    email,
		AuthType: entity.AuthTypeLocal,
		Position: getEnv("ADMIN_POSITION", "Administrator"),
		Division: getEnv("ADMIN_DIVISION", "Legal"),
		Role:     entity.RoleAdmin,
		Status:   entity.UserActive,
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
