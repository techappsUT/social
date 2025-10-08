// path: backend/internal/config/config.go
package config

import (
	"os"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Twitter  TwitterConfig
	Facebook FacebookConfig
	LinkedIn LinkedInConfig
	Security SecurityConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Port string
	Host string
}

type TwitterConfig struct {
	ClientID     string
	ClientSecret string
}

type FacebookConfig struct {
	AppID              string
	AppSecret          string
	WebhookVerifyToken string
}

type LinkedInConfig struct {
	ClientID     string
	ClientSecret string
}

type SecurityConfig struct {
	EncryptionKey string
	JWTSecret     string
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "socialqueue"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Twitter: TwitterConfig{
			ClientID:     getEnv("TWITTER_CLIENT_ID", ""),
			ClientSecret: getEnv("TWITTER_CLIENT_SECRET", ""),
		},
		Facebook: FacebookConfig{
			AppID:              getEnv("FACEBOOK_APP_ID", ""),
			AppSecret:          getEnv("FACEBOOK_APP_SECRET", ""),
			WebhookVerifyToken: getEnv("FACEBOOK_WEBHOOK_VERIFY_TOKEN", ""),
		},
		LinkedIn: LinkedInConfig{
			ClientID:     getEnv("LINKEDIN_CLIENT_ID", ""),
			ClientSecret: getEnv("LINKEDIN_CLIENT_SECRET", ""),
		},
		Security: SecurityConfig{
			EncryptionKey: getEnv("ENCRYPTION_KEY", ""),
			JWTSecret:     getEnv("JWT_SECRET", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
