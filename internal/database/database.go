package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps the sql.DB with additional functionality
type DB struct {
	*sql.DB
	config *Config
}

// NewConnection creates a new database connection with retry logic and pooling
func NewConnection(config *Config) (*DB, error) {
	db := &DB{
		config: config,
	}

	// Retry logic for initial connection
	var err error
	var sqlDB *sql.DB

	maxRetries := 10
	backoff := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		sqlDB, err = sql.Open("postgres", config.ConnectionString())
		if err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("failed to open database after %d attempts: %w", maxRetries, err)
			}

			log.Printf("Attempt %d/%d: failed to open database, retrying in %v... Error: %v",
				attempt, maxRetries, backoff, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		// Test the connection
		err = sqlDB.Ping()
		if err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("failed to ping database after %d attempts: %w", maxRetries, err)
			}

			log.Printf("Attempt %d/%d: failed to ping database, retrying in %v... Error: %v",
				attempt, maxRetries, backoff, err)
			sqlDB.Close()
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		// Connection successful
		break
	}

	db.DB = sqlDB

	// Configure connection pool
	db.DB.SetMaxOpenConns(config.MaxOpenConns)
	db.DB.SetMaxIdleConns(config.MaxIdleConns)
	db.DB.SetConnMaxLifetime(config.ConnMaxLifetime)

	log.Printf("Database connection established successfully")
	log.Printf("Connection pool: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		config.MaxOpenConns, config.MaxIdleConns, config.ConnMaxLifetime)

	return db, nil
}

// RunMigrations executes the database schema migrations
func (db *DB) RunMigrations() error {
	log.Println("Running database migrations...")

	// Read schema file
	schemaSQL, err := ioutil.ReadFile("database/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute schema
	_, err = db.Exec(string(schemaSQL))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Close gracefully closes the database connection
func (db *DB) Close() error {
	log.Println("Closing database connection...")
	return db.DB.Close()
}

// Stats returns connection pool statistics
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// Health checks the database connection
func (db *DB) Health() error {
	return db.DB.Ping()
}
