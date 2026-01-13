package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	Database DatabaseConfig

	// NATS
	NatsURL string

	// Service
	ServicePort int
	LogLevel    string
	Environment string
	JWTSecret   string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

func Load() *Config {
	// Load .env file if exists
	_ = godotenv.Load()

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "ecommerce_platform"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		ServicePort: getEnvAsInt("APP_PORT", 8009),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("APP_ENV", "development"),
		JWTSecret:   getEnv("JWT_SECRET", "default-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
