package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/routes"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"github.com/vfa-khuongdv/golang-cms/pkg/migrator"
	"gorm.io/gorm"
)

var (
	cfg    *configs.Config
	server *http.Server
)

func initializeDatabase() *gorm.DB {
	return configs.InitDB(cfg.Database)
}

func runMigrations() {
	sqlConfig := migrator.MySQLConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
	}
	dsn := migrator.NewMySQLDSN(sqlConfig)

	m, err := migrator.NewMigrator("internal/database/migrations", dsn)
	if err != nil {
		logger.Fatalf("Migration initialization failed: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		logger.Fatalf("Migration failed: %v", err)
	} else {
		logger.Infof("MySQL migrations applied successfully!")
	}
}

func main() {
	var err error
	cfg, err = configs.Load()
	if err != nil {
		logger.Fatalf("Config validation failed: %v", err)
	}

	// Initialize logger
	logger.Init(logger.LogConfig{
		ServiceName: cfg.App.ServiceName,
		Stage:       cfg.Server.Stage,
		Version:     cfg.App.Version,
	})

	// Initialize database
	db := initializeDatabase()

	// Run migrations
	if cfg.App.RunMigrate {
		runMigrations()
	}

	// Setup routes
	router := routes.SetupRouter(db)

	// Initialize custom validator
	utils.InitValidator()

	// Start server
	port := fmt.Sprintf(":%s", cfg.Server.Port)
	server = &http.Server{
		Addr:    port,
		Handler: router,
	}

	go func() {
		logger.Infof("Server starting on %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
	logger.Infof("Server exited")
}
