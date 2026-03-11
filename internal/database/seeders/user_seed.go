package seeders

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type UserSeeder struct {
	User *models.User
}

func SeedUsers(db *gorm.DB) error {
	users := []UserSeeder{
		{
			User: &models.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: utils.HashPassword("password123"),
			},
		},
		{
			User: &models.User{
				Name:     "Jane Smith",
				Email:    "jane@example.com",
				Password: utils.HashPassword("password123"),
			},
		},
	}

	for _, userData := range users {
		// Create new user
		if err := db.Create(&userData.User).Error; err != nil {
			logger.Errorf("Error creating user %s: %v", userData.User.Name, err)
			continue
		}
	}

	return nil
}
