package seeders

import (
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

// Run executes all seed functions to populate the database with initial data.
// password is the password of the sample users (see SeedPassword).
func Run(db *gorm.DB, password string) {
	// SeedUsers seeds the users table
	if err := SeedUsers(db, password); err != nil {
		logger.Errorf("Failed to seed users: %+v", err)
	}

}
