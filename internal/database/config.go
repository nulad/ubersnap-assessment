package database

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the database configuration
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLifetime time.Duration
}

// LoadConfigFromEnv loads database configuration from environment variables
func LoadConfigFromEnv() *Config {
	cfg := &Config{
		Host:             getEnvOrDefault("DB_HOST", "localhost"),
		Port:             getEnvIntOrDefault("DB_PORT", 5432),
		User:             getEnvOrDefault("DB_USER", "postgres"),
		Password:         getEnvOrDefault("DB_PASSWORD", ""),
		DBName:           getEnvOrDefault("DB_NAME", "ubersnap"),
		SSLMode:          getEnvOrDefault("DB_SSLMODE", "disable"),
		MaxOpenConns:     getEnvIntOrDefault("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:     getEnvIntOrDefault("DB_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime:  getEnvDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}
	
	return cfg
}

// ConnectionString returns the PostgreSQL connection string
func (c *Config) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
