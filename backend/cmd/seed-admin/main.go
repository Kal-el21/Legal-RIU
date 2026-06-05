package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/entity"

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

	if err := db.AutoMigrate(&entity.User{}); err != nil {
		log.Fatalf("Failed to migrate users table: %v", err)
	}

	email := requiredEnv("ADMIN_EMAIL")
	password := requiredEnv("ADMIN_PASSWORD")

	admin := entity.User{
		FullName: getEnv("ADMIN_FULL_NAME", "Super Admin"),
		Email:    email,
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
