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
)

// newHTTPServer builds the server with timeouts. ReadHeaderTimeout stops
// slow-header (slowloris) clients from holding connections; IdleTimeout must be
// longer than the load balancer's idle timeout (ALB default 60s), otherwise the
// load balancer reuses a connection Go has already closed and returns 502.
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       75 * time.Second,
	}
}

func runMigrations(database configs.DatabaseConfig) {
	sqlConfig := migrator.MySQLConfig{
		Host:     database.Host,
		Port:     database.Port,
		User:     database.User,
		Password: database.Password,
		DBName:   database.DBName,
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
	cfg, err := configs.Load()
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
	db := configs.InitDB(cfg.Database)

	// Run migrations
	if cfg.App.RunMigrate {
		runMigrations(cfg.Database)
	}

	// Setup routes
	router := routes.SetupRouter(db, cfg)

	// Initialize custom validator
	utils.InitValidator()

	// Start server
	port := fmt.Sprintf(":%s", cfg.Server.Port)
	server := newHTTPServer(port, router)

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

	// Close the pool once the in-flight requests are done with it.
	sqlDB, err := db.DB()
	if err == nil {
		err = sqlDB.Close()
	}
	if err != nil {
		logger.Errorf("Failed to close the database: %v", err)
	}
	logger.Infof("Server exited")
}
