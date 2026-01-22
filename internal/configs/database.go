package configs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

var DB *gorm.DB

// Default connection pool settings
// Note: When hight traffic is expected, consider increasing these values
// Example:
// - For 1000 concurrent connections, set MaxOpenConns to 500 and MaxIdleConns to 100
// - Monitor database performance and adjust accordingly
// - Ensure the database server can handle the configured number of connections
const (
	DEFAULT_MAX_OPEN_CONNS     = 50
	DEFAULT_MAX_IDLE_CONNS     = 10
	DEFAULT_CONN_MAX_IDLE_TIME = 5 * time.Minute
	DEFAULT_CONN_MAX_LIFETIME  = 30 * time.Minute
)

// InitDB initializes MySQL with GORM and configures a resilient connection pool
func InitDB(config DatabaseConfig) *gorm.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	// Open GORM connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		logger.Fatalf("Failed to connect to MySQL: %+v", err)
	}

	// Get underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatalf("Failed to get sql.DB: %+v", err)
	}

	// =========================
	// Connection Pool Settings
	// =========================
	setDefaults(&config)

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Validate connection
	if err := pingDB(sqlDB); err != nil {
		logger.Fatalf("Database ping failed: %+v", err)
	}

	logger.Infof(
		"MySQL connected | open=%d idle=%d lifetime=%s idleTime=%s",
		config.MaxOpenConns,
		config.MaxIdleConns,
		config.ConnMaxLifetime,
		config.ConnMaxIdleTime,
	)

	DB = db
	return db
}

// setDefaults applies safe defaults if values are not provided
func setDefaults(config *DatabaseConfig) {
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = DEFAULT_MAX_OPEN_CONNS
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = DEFAULT_MAX_IDLE_CONNS
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = DEFAULT_CONN_MAX_LIFETIME
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = DEFAULT_CONN_MAX_IDLE_TIME
	}
}

// pingDB verifies DB connectivity with timeout
func pingDB(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}
