package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/config"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/database"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/handler"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/handler/channel"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/handler/middleware"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/repository"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Logger.Level, cfg.Logger.Format)

	// Connect to database
	if err := database.Connect(&cfg.Database); err != nil {
		logger.Log.WithError(err).Fatal("Failed to connect to database")
	}
	defer database.Close()

	// Initialize sequences
	ctx := context.Background()
	if err := database.InitializeSequences(ctx); err != nil {
		logger.Log.WithError(err).Warn("Failed to initialize sequences (may already exist)")
	}

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(database.GetDB())
	profileRepo := repository.NewCustomerProfileRepository(database.GetDB())
	lookupRepo := repository.NewCustomerProfileLookupRepository(database.GetDB())
	activityLogRepo := repository.NewActivityLogRepository(database.GetDB())
	sectorRepo := repository.NewSectorRepository(database.GetDB())
	industryRepo := repository.NewIndustryRepository(database.GetDB())
	targetRepo := repository.NewTargetRepository(database.GetDB())

	// Initialize services
	idGen := service.NewIDGeneratorService()
	validationSvc := service.NewValidationService(idGen)
	customerService := service.NewCustomerService(
		database.GetDB(),
		customerRepo,
		profileRepo,
		lookupRepo,
		activityLogRepo,
		sectorRepo,
		industryRepo,
		targetRepo,
		validationSvc,
		idGen,
	)

	// Initialize handlers
	customerHandler := channel.NewCustomerHandler(customerService)
	healthHandler := handler.NewHealthHandler()

	// Setup router
	router := setupRouter(cfg, customerHandler, healthHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Log.Infof("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Log.Info("Server exited gracefully")
	}
}

func setupRouter(cfg *config.Config, customerHandler *channel.CustomerHandler, healthHandler *handler.HealthHandler) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RequestIDMiddleware())

	// Health check endpoints (no auth required, no versioning)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Channel endpoints (API key auth required)
		channelGroup := v1.Group("/channel")
		channelGroup.Use(middleware.APIKeyAuthMiddleware(cfg.API.Key))
		{
			channelGroup.POST("/directCreateCustomer", customerHandler.DirectCreateCustomer)
		}

		// Customer endpoints (for future dashboard)
		_ = v1.Group("/customer")
		// Future endpoints can be added here
	}

	return router
}
