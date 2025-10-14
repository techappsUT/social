// backend/cmd/api/config.go
// FIXED: Changed default port from 8080 to 8000

package main

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	Environment string
	BaseURL     string

	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
	Security SecurityConfig
	CORS     CORSConfig
	Social   SocialConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	AccessSecret       string
	RefreshSecret      string
	AccessTokenExpiry  string
	RefreshTokenExpiry string
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Provider    string // "sendgrid", "smtp", "mock"
	APIKey      string // For SendGrid
	FromAddress string
	FromName    string

	// SMTP specific
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	EncryptionKey string
	BcryptCost    int
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// SocialConfig holds social platform configuration
type SocialConfig struct {
	Twitter  PlatformConfig
	Facebook PlatformConfig
	LinkedIn PlatformConfig
}

// PlatformConfig holds individual platform configuration
type PlatformConfig struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8000"),

		Server: ServerConfig{
			Port: getEnv("PORT", "8000"), // ✅ FIXED: Changed from 8080 to 8000
			Host: getEnv("HOST", "0.0.0.0"),
		},

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "socialqueue"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		JWT: JWTConfig{
			AccessSecret:       getEnv("JWT_ACCESS_SECRET", "your-access-secret-key-change-this"),
			RefreshSecret:      getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-this"),
			AccessTokenExpiry:  getEnv("JWT_ACCESS_EXPIRY", "15m"),
			RefreshTokenExpiry: getEnv("JWT_REFRESH_EXPIRY", "30d"),
		},

		Email: EmailConfig{
			Provider:     getEnv("EMAIL_PROVIDER", "mock"),
			APIKey:       getEnv("SENDGRID_API_KEY", ""),
			FromAddress:  getEnv("EMAIL_FROM", "noreply@socialqueue.com"),
			FromName:     getEnv("EMAIL_FROM_NAME", "SocialQueue"),
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		},

		Security: SecurityConfig{
			EncryptionKey: getEnv("ENCRYPTION_KEY", ""),
			BcryptCost:    getEnvAsInt("BCRYPT_COST", 10),
		},

		CORS: CORSConfig{
			AllowedOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8000"), ","),
			AllowedMethods: strings.Split(getEnv("CORS_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"), ","),
			AllowedHeaders: strings.Split(getEnv("CORS_HEADERS", "Accept,Authorization,Content-Type,X-Request-ID"), ","),
		},

		Social: SocialConfig{
			Twitter: PlatformConfig{
				ClientID:     getEnv("TWITTER_CLIENT_ID", ""),
				ClientSecret: getEnv("TWITTER_CLIENT_SECRET", ""),
				CallbackURL:  getEnv("TWITTER_CALLBACK_URL", "http://localhost:8000/api/social/auth/twitter/callback"),
			},
			Facebook: PlatformConfig{
				ClientID:     getEnv("FACEBOOK_CLIENT_ID", ""),
				ClientSecret: getEnv("FACEBOOK_CLIENT_SECRET", ""),
				CallbackURL:  getEnv("FACEBOOK_CALLBACK_URL", "http://localhost:8000/api/social/auth/facebook/callback"),
			},
			LinkedIn: PlatformConfig{
				ClientID:     getEnv("LINKEDIN_CLIENT_ID", ""),
				ClientSecret: getEnv("LINKEDIN_CLIENT_SECRET", ""),
				CallbackURL:  getEnv("LINKEDIN_CALLBACK_URL", "http://localhost:8000/api/social/auth/linkedin/callback"),
			},
		},
	}
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if intValue, err := strconv.Atoi(strValue); err == nil {
		return intValue
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	strValue := getEnv(key, "")
	if boolValue, err := strconv.ParseBool(strValue); err == nil {
		return boolValue
	}
	return defaultValue
}

func stringsFromEnv(key string, defaultValue []string) []string {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	return strings.Split(value, ",")
}
