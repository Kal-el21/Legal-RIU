package main

import (
	"errors"
	"log"
	"os"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/seed"

	"github.com/google/uuid"
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

	if err := seed.RunAllMigrationsAndSeeds(db); err != nil {
		log.Fatalf("Migration and seed failed: %v", err)
	}

	email := getEnv("ADMIN_EMAIL", "admin@example.com")
	password := getEnv("ADMIN_PASSWORD", "12345678")
	adminDivision := getEnv("ADMIN_DIVISION", "Legal, compliance and risk management division")

	companyID, err := seed.FindCompanyIDByDomain(db, email)
	if err != nil {
		log.Printf("Admin email domain does not match any company, using first available company: %v", err)
		companyID, err = seed.FindFirstCompanyID(db)
		if err != nil {
			log.Fatalf("Failed to find fallback company for admin: %v", err)
		}
	}

	divisionID, err := seed.FindDivisionIDByName(db, adminDivision)
	if err != nil {
		log.Printf("Admin division not found, leaving division_id empty: %v", err)
	}

	admin := entity.User{
		FullName:           getEnv("ADMIN_FULL_NAME", "Super Admin"),
		Email:              email,
		Position:           getEnv("ADMIN_POSITION", "Administrator"),
		Division:           adminDivision,
		Role:               entity.RoleAdmin,
		Status:             entity.UserActive,
		EmailNotifications: true,
		TwoFAEnabled:       false,
		CompanyID:          &companyID,
	}
	if divisionID != uuid.Nil {
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
		if existing.EmailNotifications != true {
			updates["email_notifications"] = true
		}
		if existing.TwoFAEnabled != false {
			updates["two_fa_enabled"] = false
		}
		if companyID != uuid.Nil {
			if existing.CompanyID == nil || *existing.CompanyID != companyID {
				updates["company_id"] = companyID
			}
		} else {
			if existing.CompanyID != nil {
				updates["company_id"] = nil
			}
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

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
