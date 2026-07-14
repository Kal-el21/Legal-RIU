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
		log.Printf("Admin user %s already exists, skipping seed", email)
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

	if err := db.Exec(`
		UPDATE agreement_documents 
		SET requester_id = ?
		WHERE requester_id IS NULL
	`, admin.ID).Error; err != nil {
		log.Fatalf("Failed to backfill agreement_documents.requester_id: %v", err)
	}

	if err := seed.EnforceAgreementDocumentRequesterIDNotNull(db); err != nil {
		log.Fatalf("Failed to enforce agreement_documents.requester_id NOT NULL: %v", err)
	}

	if err := seed.EnforceAgreementDocumentTicketNumberNotNull(db); err != nil {
		log.Fatalf("Failed to enforce agreement_documents.ticket_number NOT NULL: %v", err)
	}

	if err := seed.BackfillAgreementDocumentDocumentTypeCode(db); err != nil {
		log.Fatalf("Failed to backfill agreement_documents.document_type_code: %v", err)
	}
	if err := seed.EnforceAgreementDocumentDocumentTypeCodeNotNull(db); err != nil {
		log.Fatalf("Failed to enforce agreement_documents.document_type_code NOT NULL: %v", err)
	}

	if err := seed.EnforceAgreementDocumentFormDataNotNull(db); err != nil {
		log.Fatalf("Failed to enforce agreement_documents.form_data NOT NULL: %v", err)
	}

	if err := seed.BackfillAgreementAttachmentColumns(db); err != nil {
		log.Fatalf("Failed to backfill agreement_attachments: %v", err)
	}
	if err := seed.EnforceAgreementAttachmentColumnsNotNull(db); err != nil {
		log.Fatalf("Failed to enforce agreement_attachments NOT NULL: %v", err)
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
