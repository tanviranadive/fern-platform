// Fern Platform - Unified platform entry point
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/guidewire-oss/fern-platform/pkg/config"
	"github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"github.com/guidewire-oss/fern-platform/pkg/middleware"
	"github.com/guidewire-oss/fern-platform/internal/reporter/api"
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
)

func main() {
	var configPath = flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	configManager := config.NewManager()
	if err := configManager.Load(*configPath); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cfg := config.GetConfig()

	// Initialize logging
	if err := logging.Initialize(&cfg.Logging); err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}

	logger := logging.GetLogger()
	logger.WithService("fern-platform").Info("Starting Fern Platform")

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		logger.WithService("fern-platform").WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate("migrations"); err != nil {
		logger.WithService("fern-platform").WithError(err).Fatal("Failed to run database migrations")
	}

	// Initialize repositories
	testRunRepo := repository.NewTestRunRepository(db.DB)
	suiteRunRepo := repository.NewSuiteRunRepository(db.DB)
	specRunRepo := repository.NewSpecRunRepository(db.DB)
	tagRepo := repository.NewTagRepository(db.DB)
	projectRepo := repository.NewProjectRepository(db.DB)

	// Initialize services
	testRunService := service.NewTestRunService(testRunRepo, suiteRunRepo, specRunRepo, logger)
	projectService := service.NewProjectService(projectRepo, logger)
	tagService := service.NewTagService(tagRepo, logger)

	// Initialize GraphQL resolver
	resolver := graphql.NewResolver(testRunService, projectService, tagService, db.DB, logger)

	// Initialize HTTP server
	if cfg.Server.Host == "0.0.0.0" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.HealthCheckMiddleware())

	// CORS configuration
	if gin.Mode() == gin.DebugMode {
		router.Use(middleware.DevCORSMiddleware())
	} else {
		corsConfig := middleware.DefaultCORSConfig()
		router.Use(middleware.NewCORSMiddleware(corsConfig))
	}

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(&cfg.Auth, logger)
	oauthMiddleware := middleware.NewOAuthMiddleware(&cfg.Auth, db.DB, logger)

	// REST API routes
	apiHandler := api.NewHandler(testRunService, projectService, tagService, authMiddleware, oauthMiddleware, logger)
	apiHandler.RegisterRoutes(router)

	// GraphQL routes with role group names from config
	roleGroupNames := &graphql.RoleGroupNames{
		AdminGroup:   cfg.Auth.OAuth.AdminGroupName,
		ManagerGroup: cfg.Auth.OAuth.ManagerGroupName,
		UserGroup:    cfg.Auth.OAuth.UserGroupName,
	}
	gqlHandler := graphql.NewHandler(resolver, roleGroupNames)
	gqlHandler.RegisterRoutes(router, oauthMiddleware)

	// Note: Static file serving is handled by the API handler

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.WithService("fern-platform").
			WithFields(map[string]interface{}{
				"host": cfg.Server.Host,
				"port": cfg.Server.Port,
			}).Info("Starting Fern Platform HTTP server")

		if cfg.Server.TLS.Enabled {
			if err := srv.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				logger.WithService("fern-platform").WithError(err).Fatal("Failed to start HTTPS server")
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.WithService("fern-platform").WithError(err).Fatal("Failed to start HTTP server")
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.WithService("fern-platform").Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithService("fern-platform").WithError(err).Fatal("Server forced to shutdown")
	}

	logger.WithService("fern-platform").Info("Server exited")
}