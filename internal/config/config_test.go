package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear env vars to ensure we test defaults
	// Note: This changes process env, which might affect other tests if run in parallel.
	// Ideally we should restore them.
	keys := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "SERVER_PORT"}
	originalEnv := make(map[string]string)
	for _, key := range keys {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, val := range originalEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.Host != "postgres" {
		t.Errorf("expected DB_HOST default 'postgres', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("expected DB_PORT default '5432', got '%s'", cfg.Database.Port)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("expected SERVER_PORT default '8080', got '%s'", cfg.Server.Port)
	}
}

func TestLoad_EnvVars(t *testing.T) {
	// Set env vars
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "9999")
	os.Setenv("SERVER_PORT", "3000")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("SERVER_PORT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.Host != "testhost" {
		t.Errorf("expected DB_HOST 'testhost', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.Port != "9999" {
		t.Errorf("expected DB_PORT '9999', got '%s'", cfg.Database.Port)
	}
	if cfg.Server.Port != "3000" {
		t.Errorf("expected SERVER_PORT '3000', got '%s'", cfg.Server.Port)
	}
}
