// path: backend/cmd/api/config.go
package main

import "os"

// ============================================================================
// CONFIGURATION STRUCTURE
// ============================================================================

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	Security    SecurityConfig
	BaseURL     string
	CORS        CORSConfig
	Social      SocialConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
}

type SecurityConfig struct {
	EncryptionKey string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type SocialConfig struct {
	Twitter  PlatformConfig
	Facebook PlatformConfig
	LinkedIn PlatformConfig
}

type PlatformConfig struct {
	ClientID     string
	ClientSecret string
}

// ============================================================================
// CONFIGURATION LOADER
// ============================================================================

func LoadConfig() *Config {
	env := getEnv("ENVIRONMENT", "development")

	return &Config{
		Environment: env,

		Server: ServerConfig{
			Port: getEnv("PORT", "8000"),
			Host: getEnv("HOST", "0.0.0.0"),
		},

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "socialqueue"),
			Password: getEnv("DB_PASSWORD", "socialqueue_dev_password"),
			DBName:   getEnv("DB_NAME", "socialqueue_dev"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-super-secret-access-key-change-in-production"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-super-secret-refresh-key-change-in-production"),
		},

		Security: SecurityConfig{
			EncryptionKey: getEnv("ENCRYPTION_KEY", ""),
		},

		BaseURL: getEnv("BASE_URL", "http://localhost:3000"),

		CORS: CORSConfig{
			AllowedOrigins: []string{
				getEnv("FRONTEND_URL", "http://localhost:3000"),
				"http://localhost:3001",
			},
		},

		Social: SocialConfig{
			Twitter: PlatformConfig{
				ClientID:     getEnv("TWITTER_CLIENT_ID", ""),
				ClientSecret: getEnv("TWITTER_CLIENT_SECRET", ""),
			},
			Facebook: PlatformConfig{
				ClientID:     getEnv("FACEBOOK_APP_ID", ""),
				ClientSecret: getEnv("FACEBOOK_APP_SECRET", ""),
			},
			LinkedIn: PlatformConfig{
				ClientID:     getEnv("LINKEDIN_CLIENT_ID", ""),
				ClientSecret: getEnv("LINKEDIN_CLIENT_SECRET", ""),
			},
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
