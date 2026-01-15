package database

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// InitDatabase initializes the database connection and runs migrations
func InitDatabase() (*DB, error) {
	// Load configuration from environment
	config := LoadConfigFromEnv()

	// Create connection
	db, err := NewConnection(config)
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		db.Close()
		return nil, err
	}

	// Setup graceful shutdown
	setupGracefulShutdown(db)

	return db, nil
}

// setupGracefulShutdown ensures database connections are closed on shutdown
func setupGracefulShutdown(db *DB) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Received shutdown signal, closing database connection...")
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
		os.Exit(0)
	}()
}

// WithContext provides context-aware database operations
func (db *DB) WithContext(ctx context.Context) *sqlCtx {
	return &sqlCtx{
		db:  db,
		ctx: ctx,
	}
}

// sqlCtx wraps DB with a context for context-aware operations
type sqlCtx struct {
	db  *DB
	ctx context.Context
}

// ExecContext executes a query with context
func (s *sqlCtx) ExecContext(query string, args ...interface{}) (sql.Result, error) {
	return s.db.ExecContext(s.ctx, query, args...)
}

// QueryContext executes a query that returns rows with context
func (s *sqlCtx) QueryContext(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(s.ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row with context
func (s *sqlCtx) QueryRowContext(query string, args ...interface{}) *sql.Row {
	return s.db.QueryRowContext(s.ctx, query, args...)
}
