package database

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromEnv(t *testing.T) {
	// Store original env vars
	originalHost := os.Getenv("DB_HOST")
	originalPort := os.Getenv("DB_PORT")
	originalUser := os.Getenv("DB_USER")
	originalPassword := os.Getenv("DB_PASSWORD")
	originalDBName := os.Getenv("DB_NAME")
	originalSSLMode := os.Getenv("DB_SSLMODE")
	originalMaxOpen := os.Getenv("DB_MAX_OPEN_CONNS")
	originalMaxIdle := os.Getenv("DB_MAX_IDLE_CONNS")
	originalMaxLifetime := os.Getenv("DB_CONN_MAX_LIFETIME")

	// Clean up after test
	defer func() {
		os.Setenv("DB_HOST", originalHost)
		os.Setenv("DB_PORT", originalPort)
		os.Setenv("DB_USER", originalUser)
		os.Setenv("DB_PASSWORD", originalPassword)
		os.Setenv("DB_NAME", originalDBName)
		os.Setenv("DB_SSLMODE", originalSSLMode)
		os.Setenv("DB_MAX_OPEN_CONNS", originalMaxOpen)
		os.Setenv("DB_MAX_IDLE_CONNS", originalMaxIdle)
		os.Setenv("DB_CONN_MAX_LIFETIME", originalMaxLifetime)
	}()

	// Test with custom values
	os.Setenv("DB_HOST", "custom-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "custom-user")
	os.Setenv("DB_PASSWORD", "custom-pass")
	os.Setenv("DB_NAME", "custom-db")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("DB_MAX_IDLE_CONNS", "20")
	os.Setenv("DB_CONN_MAX_LIFETIME", "10m")

	config := LoadConfigFromEnv()

	assert.Equal(t, "custom-host", config.Host)
	assert.Equal(t, 5433, config.Port)
	assert.Equal(t, "custom-user", config.User)
	assert.Equal(t, "custom-pass", config.Password)
	assert.Equal(t, "custom-db", config.DBName)
	assert.Equal(t, "require", config.SSLMode)
	assert.Equal(t, 50, config.MaxOpenConns)
	assert.Equal(t, 20, config.MaxIdleConns)
	assert.Equal(t, 10*time.Minute, config.ConnMaxLifetime)
}

func TestLoadConfigFromEnvDefaults(t *testing.T) {
	// Clear env vars
	envVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"DB_SSLMODE", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}

	config := LoadConfigFromEnv()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "postgres", config.User)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, "ubersnap", config.DBName)
	assert.Equal(t, "disable", config.SSLMode)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
}

func TestConfigConnectionString(t *testing.T) {
	config := &Config{
		Host:     "test-host",
		Port:     5432,
		User:     "test-user",
		Password: "test-pass",
		DBName:   "test-db",
		SSLMode:  "disable",
	}

	expected := "postgres://test-user:test-pass@test-host:5432/test-db?sslmode=disable"
	assert.Equal(t, expected, config.ConnectionString())
}

func TestConfigConnectionStringWithSpecialChars(t *testing.T) {
	config := &Config{
		Host:     "test-host",
		Port:     5432,
		User:     "user@domain.com",
		Password: "p@ssw0rd!",
		DBName:   "test-db",
		SSLMode:  "require",
	}

	// The connection string should be properly formatted
	connStr := config.ConnectionString()
	require.Contains(t, connStr, "postgres://")
	require.Contains(t, connStr, "user@domain.com")
	require.Contains(t, connStr, "p@ssw0rd!")
	require.Contains(t, connStr, "sslmode=require")
}
