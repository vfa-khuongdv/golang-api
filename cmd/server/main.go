package main

import (
	"fmt"

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
	if err := router.Run(port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
