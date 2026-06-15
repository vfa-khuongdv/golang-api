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

func initializeDatabase() *gorm.DB {
	config := configs.DatabaseConfig{
		Host:     utils.GetEnv("DB_HOST", "127.0.0.1"),
		Port:     utils.GetEnv("DB_PORT", "3306"),
		User:     utils.GetEnv("DB_USERNAME", ""),
		Password: utils.GetEnv("DB_PASSWORD", ""),
		DBName:   utils.GetEnv("DB_DATABASE", ""),
	}
	return configs.InitDB(config)
}

func runMigrations() {
	sqlConfig := migrator.MySQLConfig{
		Host:     utils.GetEnv("DB_HOST", "127.0.0.1"),
		Port:     utils.GetEnv("DB_PORT", "3306"),
		User:     utils.GetEnv("DB_USERNAME", ""),
		Password: utils.GetEnv("DB_PASSWORD", ""),
		DBName:   utils.GetEnv("DB_DATABASE", ""),
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
	// Load environment variables
	configs.LoadEnv()

	// Initialize logger
	logger.Init()

	// Initialize database
	db := initializeDatabase()

	// Run migrations
	isRunMigrate := utils.GetEnv("RUN_MIGRATE", "false")
	if isRunMigrate == "true" {
		runMigrations()
	}

	// Setup routes
	router := routes.SetupRouter(db)

	// Initialize custom validator
	utils.InitValidator()

	// Start server
	port := fmt.Sprintf(":%s", utils.GetEnv("PORT", "3000"))
	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}

	go func() {
		logger.Infof("Server starting on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
	logger.Infof("Server exited")
}
