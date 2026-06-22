package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	MinIO    MinIOConfig
	Security SecurityConfig
	LDAP     LDAPConfig
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
	Secret              string
	ExpiresHours        int
	RefreshExpiresHours int
}

type MinIOConfig struct {
	Endpoint              string
	AccessKey             string
	SecretKey             string
	Bucket                string
	UseSSL                bool
	PresignExpiresMinutes int
	PublicEndpoint        string
}

type SecurityConfig struct {
	AllowedOrigins         []string
	LoginRateLimit         int
	LoginRateWindowMinutes int
}

type LDAPConfig struct {
	Host               string
	Port               int
	UseSSL             bool
	InsecureSkipVerify bool
	BindDN             string
	BindPassword       string
	BaseDN             string
	UserFilter         string
	AttrName           string
	AttrEmail          string
	AttrPosition       string
	AttrDivision       string
	DefaultEmailDomain string
	DefaultPosition    string
	DefaultDivision    string
}

var AppConfig_ *Config

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	jwtExpires, _ := strconv.Atoi(getEnv("JWT_EXPIRES_HOURS", "24"))
	refreshExpires, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRES_HOURS", "168"))
	minioSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))
	presignExpires, _ := strconv.Atoi(getEnv("MINIO_PRESIGN_EXPIRES_MINUTES", "15"))
	loginRateLimit, _ := strconv.Atoi(getEnv("LOGIN_RATE_LIMIT", "5"))
	loginRateWindow, _ := strconv.Atoi(getEnv("LOGIN_RATE_WINDOW_MINUTES", "15"))

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
			Secret:              getEnv("JWT_SECRET", "secret"),
			ExpiresHours:        jwtExpires,
			RefreshExpiresHours: refreshExpires,
		},
		MinIO: MinIOConfig{
			Endpoint:              getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:             getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:             getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:                getEnv("MINIO_BUCKET", "legal-riu"),
			UseSSL:                minioSSL,
			PresignExpiresMinutes: presignExpires,
			PublicEndpoint:        getEnv("MINIO_PUBLIC_ENDPOINT", getEnv("MINIO_ENDPOINT", "localhost:9000")),
		},
		Security: SecurityConfig{
			AllowedOrigins: splitCSV(getEnv(
				"CORS_ALLOWED_ORIGINS",
				"http://localhost:89,http://localhost:5173,http://127.0.0.1:89,http://127.0.0.1:5173",
			)),
			LoginRateLimit:         loginRateLimit,
			LoginRateWindowMinutes: loginRateWindow,
		},
		LDAP: LDAPConfig{
			Host:               getEnv("LDAP_HOST", ""),
			Port:               getEnvAsInt("LDAP_PORT", 389),
			UseSSL:             getEnvAsBool("LDAP_USE_SSL", false),
			InsecureSkipVerify: getEnvAsBool("LDAP_INSECURE_SKIP_VERIFY", false),
			BindDN:             getEnv("LDAP_BIND_DN", ""),
			BindPassword:       getEnv("LDAP_BIND_PASSWORD", ""),
			BaseDN:             getEnv("LDAP_BASE_DN", ""),
			UserFilter:         getEnv("LDAP_USER_FILTER", "(sAMAccountName=%s)"),
			AttrName:           getEnv("LDAP_ATTR_NAME", "displayName"),
			AttrEmail:          getEnv("LDAP_ATTR_EMAIL", "mail"),
			AttrPosition:       getEnv("LDAP_ATTR_POSITION", "title"),
			AttrDivision:       getEnv("LDAP_ATTR_DIVISION", "department"),
			DefaultEmailDomain: getEnv("LDAP_DEFAULT_EMAIL_DOMAIN", ""),
			DefaultPosition:    getEnv("LDAP_DEFAULT_POSITION", "Staff"),
			DefaultDivision:    getEnv("LDAP_DEFAULT_DIVISION", "General"),
		},
	}

	if cfg.App.Env == "production" && (cfg.JWT.Secret == "secret" || cfg.JWT.Secret == "change-me" || len(cfg.JWT.Secret) < 32) {
		log.Fatal("JWT_SECRET must be changed to a strong value in production")
	}

	if cfg.App.Env == "production" {
		if cfg.LDAP.Host == "" {
			log.Println("Warning: LDAP_HOST is not set in production; only local users can log in")
		}
		if cfg.LDAP.Host != "" && (cfg.LDAP.BindDN == "" || cfg.LDAP.BaseDN == "") {
			log.Println("Warning: LDAP_BIND_DN and LDAP_BASE_DN should be set when LDAP is enabled")
		}
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

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
