package seeders

import (
	"errors"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

// DevPassword is the well-known password of the sample users outside prod.
const DevPassword = "password123"

// SeedPassword returns the password for the sample users: configured (the
// SEED_USER_PASSWORD variable) if set, else DevPassword. In prod the first
// sample user becomes an admin, so a well-known password is refused there.
func SeedPassword(stage, configured string) (string, error) {
	if configured != "" {
		return configured, nil
	}
	if stage == "prod" {
		return "", errors.New("SEED_USER_PASSWORD must be set to seed users when STAGE is prod")
	}
	return DevPassword, nil
}

type UserSeeder struct {
	User *models.User
}

// SeedUsers creates two sample users with the given password; the first one
// gets the admin role when it exists.
func SeedUsers(db *gorm.DB, password string) error {
	hashed, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	users := []UserSeeder{
		{
			User: &models.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: hashed,
				Gender:   models.GenderMale,
			},
		},
		{
			User: &models.User{
				Name:     "Jane Smith",
				Email:    "jane@example.com",
				Password: hashed,
				Gender:   models.GenderFemale,
			},
		},
	}

	// The first seeded user becomes the admin so the admin-only endpoints are reachable
	var adminRole models.Role
	hasAdminRole := db.Where("name = ?", models.RoleAdmin).First(&adminRole).Error == nil

	for i, userData := range users {
		// Create new user
		if err := db.Create(&userData.User).Error; err != nil {
			logger.Errorf("Error creating user %s: %v", userData.User.Name, err)
			continue
		}

		if i == 0 && hasAdminRole {
			userRole := models.UserRole{UserID: userData.User.ID, RoleID: adminRole.ID}
			if err := db.Create(&userRole).Error; err != nil {
				logger.Errorf("Error assigning admin role to %s: %v", userData.User.Email, err)
			}
		}
	}

	return nil
}
