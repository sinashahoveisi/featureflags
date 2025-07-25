package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"featureflags/config"
	"featureflags/controller"
	"featureflags/handler"
	"featureflags/migrations"
	"featureflags/pkg/logger"
	"featureflags/repository"
	"featureflags/service"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)



func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logger.Level, cfg.Logger.Mode)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Infow("Starting FeatureFlags service",
		"version", "1.0.0",
		"port", cfg.HTTPServer.Port,
		"log_level", cfg.Logger.Level,
		"log_mode", cfg.Logger.Mode,
	)

	// Connect to database
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalw("Failed to connect to database", "error", err)
	}
	defer db.Close()

	log.Infow("Database connected successfully",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"database", cfg.Database.Name,
	)

	// Run migrations
	if err := migrations.RunMigrations(db.DB, "./migrations"); err != nil {
		log.Fatalw("Failed to run database migrations", "error", err)
	}

	log.Infow("Database migrations completed successfully")

	// Initialize repositories
	flagRepo := repository.NewFlagRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Initialize services
	flagService := service.NewFlagService(flagRepo, auditRepo, log)

	// Initialize controllers
	flagController := controller.NewFlagController(flagService, log)

	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true

	// Register routes
	handler.RegisterRoutes(e, flagController, cfg, log)

	// Start server in a goroutine
	serverAddr := fmt.Sprintf(":%d", cfg.HTTPServer.Port)
	go func() {
		log.Infow("Starting HTTP server", "address", serverAddr)
		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			log.Fatalw("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Infow("Shutting down server gracefully...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Application.GracefulShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := e.Shutdown(ctx); err != nil {
		log.Errorw("Failed to shutdown server gracefully", "error", err)
		os.Exit(1)
	}

	log.Infow("Server shutdown completed successfully")
}

func connectDB(cfg *config.Config) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
} 