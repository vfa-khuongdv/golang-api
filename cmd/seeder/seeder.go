package main

import (
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/database/seeders"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

func main() {
	// Load env package
	configs.LoadEnv()

	// Init logger
	logger.Init()

	// MySQL database configuration
	config := configs.DatabaseConfig{
		Host:     utils.GetEnv("DB_HOST", "127.0.0.1"),
		Port:     utils.GetEnv("DB_PORT", "3306"),
		User:     utils.GetEnv("DB_USERNAME", ""),
		Password: utils.GetEnv("DB_PASSWORD", ""),
		DBName:   utils.GetEnv("DB_DATABASE", ""),
	}

	// Initialize database connection
	db := configs.InitDB(config)

	// Run seeder
	seeders.Run(db)
}
