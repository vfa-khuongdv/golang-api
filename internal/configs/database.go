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

var (
	openGormConnection = func(dsn string) (*gorm.DB, error) {
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			PrepareStmt: false,
		})
	}
	getSQLDBConnection = func(db *gorm.DB) (*sql.DB, error) {
		return db.DB()
	}
	logFatalf = logger.Fatalf
	logInfof  = logger.Infof
	pingDBFn  = pingDB
)

// Default connection pool settings, sized for a small deployment.
// Total connections to the database = tasks x MaxOpenConns, so size the pool
// from the database limit, not from the app's concurrency:
//
//	MaxOpenConns <= (max_connections * 0.8) / max number of tasks
//
// e.g. RDS db.t3.small (max_connections ~150) with 5 tasks: (150 * 0.8) / 5 = 24.
// Override with DB_MAX_OPEN_CONNS / DB_MAX_IDLE_CONNS per environment.
const (
	DEFAULT_MAX_OPEN_CONNS     = 20
	DEFAULT_MAX_IDLE_CONNS     = 5
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
	db, err := openGormConnection(dsn)
	if err != nil {
		logFatalf("Failed to connect to MySQL: %+v", err)
	}

	// Get underlying sql.DB
	sqlDB, err := getSQLDBConnection(db)
	if err != nil {
		logFatalf("Failed to get sql.DB: %+v", err)
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
	if err := pingDBFn(sqlDB); err != nil {
		logFatalf("Database ping failed: %+v", err)
	}

	logInfof(
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
