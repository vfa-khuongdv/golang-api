package seeders

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

func hashPassword(password string) string {
	hashed, _ := utils.HashPassword(password)
	return hashed
}

type UserSeeder struct {
	User *models.User
}

func SeedUsers(db *gorm.DB) error {
	users := []UserSeeder{
		{
			User: &models.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: hashPassword("password123"),
			},
		},
		{
			User: &models.User{
				Name:     "Jane Smith",
				Email:    "jane@example.com",
				Password: hashPassword("password123"),
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
