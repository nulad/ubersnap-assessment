package main

import (
	"log"

	"github.com/nulad/ubersnap-assessment/internal/database"
)

func main() {
	log.Println("Starting Ubersnap server...")

	// Initialize database connection
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Check database health
	if err := db.Health(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	// Print connection pool stats
	stats := db.Stats()
	log.Printf("Database connection pool stats: %+v", stats)

	log.Println("Server started successfully")
	
	// Keep the server running
	select {}
}
