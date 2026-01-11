// This is the main entry point for the Budget Planner API server.
package main

import (
	// Go standard libraries
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Internal packages
	"budget-planner/internal/api/rest/router"
	"budget-planner/internal/config"
	"budget-planner/internal/infrastructure/database/postgres"
	"budget-planner/pkg/logger"

	// External packages
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	log.Info("Starting Budget Planner API Server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	log.SetLevel(cfg.Environment.LogLevel)

	// Connect to PostgreSQL with connection pooling
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL", "error", err)
	}

	// Ensure database connection is closed when the application exits
	defer func() {
		db.Close()
		log.Info("PostgreSQL connection pool closed")
	}()

	// Initialize Gin router with recommended middlewares
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Set Gin mode based on environment
	if cfg.Environment.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		ExposeHeaders:    cfg.CORS.ExposeHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// Health check endpoint with database connectivity check
	r.GET("/health", func(c *gin.Context) {
		// Check database connectivity
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "database unavailable", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Register all routes
	router.RegisterRoutes(r, db, log, cfg)

	// Configure server with timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSeconds) * time.Second,
	}

	// Create a server context for graceful shutdown
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Set up graceful shutdown channel
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Info("Server starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", "error", err)
		}
		serverStopCtx()
	}()

	// Wait for shutdown signal
	select {
	case <-quit:
		log.Info("Shutdown signal received...")
	case <-serverCtx.Done():
		log.Info("Server stopped...")
	}

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Shutdown the server
	log.Info("Shutting down server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited properly")
}
