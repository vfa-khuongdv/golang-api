package seeders

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

func SeedPermissions(db *gorm.DB) error {
	permissions := []models.Permission{
		// User resource permissions
		{
			ID:       1,
			Resource: "users", // Resource: User management
			Action:   "index", // Action: List all users
		},
		{
			ID:       2,
			Resource: "users",  // Resource: User management
			Action:   "create", // Action: Create new user
		},
		{
			ID:       3,
			Resource: "users",  // Resource: User management
			Action:   "update", // Action: Update existing user
		},
		{
			ID:       4,
			Resource: "users", // Resource: User management
			Action:   "view",  // Action: View user details
		},
		{
			ID:       5,
			Resource: "users",  // Resource: User management
			Action:   "delete", // Action: Delete user
		},
		// Role resource permissions
		{
			ID:       6,
			Resource: "roles", // Resource: Role management
			Action:   "index", // Action: List all roles
		},
		{
			ID:       7,
			Resource: "roles",  // Resource: Role management
			Action:   "create", // Action: Create new role
		},
		{
			ID:       8,
			Resource: "roles",  // Resource: Role management
			Action:   "update", // Action: Update existing role
		},
		{
			ID:       9,
			Resource: "roles", // Resource: Role management
			Action:   "view",  // Action: View role details
		},
		{
			ID:       10,
			Resource: "roles",  // Resource: Role management
			Action:   "delete", // Action: Delete role
		},
		// Settings resource permissions
		{
			ID:       11,
			Resource: "settings", // Resource: System settings
			Action:   "view",     // Action: View settings
		},
		{
			ID:       12,
			Resource: "settings", // Resource: System settings
			Action:   "update",   // Action: Update settings
		},
	}

	for _, permission := range permissions {
		if err := db.Create(&permission).Error; err != nil {
			logger.Infof("The permission %v, action %v was run before\n", permission.Resource, permission.Action)
		}
	}

	return nil
}
