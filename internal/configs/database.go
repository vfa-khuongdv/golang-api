package configs

import (
	"fmt"

	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

var DB *gorm.DB

// InitDB initializes a MySQL database connection using GORM
// Parameters:
//   - config: DatabaseConfig struct containing database connection parameters
//
// Returns:
//   - *gorm.DB: Database connection instance
//
// Note: Also sets the global DB variable with the connection instance
func InitDB(config DatabaseConfig) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to connect to the MySQL database: %+v", err)
	} else {
		logger.Info("MySQL database connection established successfully")
	}
	DB = db
	return db
}
