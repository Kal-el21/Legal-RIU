package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	MinIO    MinIOConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret       string
	ExpiresHours int
}

type MinIOConfig struct {
	Endpoint              string
	AccessKey             string
	SecretKey             string
	Bucket                string
	UseSSL                bool
	PresignExpiresMinutes int
}

var AppConfig_ *Config

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	jwtExpires, _ := strconv.Atoi(getEnv("JWT_EXPIRES_HOURS", "24"))
	minioSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))
	presignExpires, _ := strconv.Atoi(getEnv("MINIO_PRESIGN_EXPIRES_MINUTES", "15"))

	cfg := &Config{
		App: AppConfig{
			Port: getEnv("APP_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "legal_riu_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:       getEnv("JWT_SECRET", "secret"),
			ExpiresHours: jwtExpires,
		},
		MinIO: MinIOConfig{
			Endpoint:              getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:             getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:             getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:                getEnv("MINIO_BUCKET", "legal-riu"),
			UseSSL:                minioSSL,
			PresignExpiresMinutes: presignExpires,
		},
	}

	AppConfig_ = cfg
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
