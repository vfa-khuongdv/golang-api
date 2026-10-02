package main

import (
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/database/seeders"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		logger.Fatalf("Config validation failed: %v", err)
	}

	// Init logger
	logger.Init(logger.LogConfig{
		ServiceName: "golang-cms",
		Stage:       cfg.Server.Stage,
	})

	password, err := seeders.SeedPassword(cfg.Server.Stage, configs.GetEnv("SEED_USER_PASSWORD", ""))
	if err != nil {
		logger.Fatalf("Seeding refused: %v", err)
	}

	// Initialize database connection
	db := configs.InitDB(cfg.Database)

	// Run seeder
	seeders.Run(db, password)
}
