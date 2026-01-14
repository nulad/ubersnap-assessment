package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nulad/ubersnap-assessment/internal/config"
	"github.com/nulad/ubersnap-assessment/internal/database"
	"github.com/nulad/ubersnap-assessment/internal/handler"
	"github.com/nulad/ubersnap-assessment/internal/repository"
	"github.com/nulad/ubersnap-assessment/internal/service"
)

func main() {
	// 1. Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Starting server...")

	// 2. Initialize database connection
	// We use the database package's config loader for DB specifics
	dbConfig := database.LoadConfigFromEnv()
	
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// 3. Create repository instances
	// db.DB access the underlying *sql.DB from the database.DB wrapper
	couponRepo := repository.NewCouponRepository(db.DB)
	claimRepo := database.NewClaimRepository(db.DB)

	// 4. Create service instances
	couponService := service.NewCouponService(couponRepo, claimRepo, db.DB)

	// 5. Create handler instances
	couponHandler := handler.NewCouponHandler(couponService)

	// 6. Setup Gin router with routes
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/coupons", couponHandler.CreateCoupon)
		api.POST("/coupons/claim", couponHandler.ClaimCoupon)
		api.GET("/coupons/:name", couponHandler.GetCoupon)
	}

	// 7. Start HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// Start server in a separate goroutine
	go func() {
		log.Printf("Server listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 8. Handle graceful shutdown (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
