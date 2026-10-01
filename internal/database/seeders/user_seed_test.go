package seeders_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/database/seeders"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSeederDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.UserRole{}))
	return db
}

func rolesOf(t *testing.T, db *gorm.DB, email string) []string {
	var names []string
	require.NoError(t, db.Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Joins("JOIN users ON users.id = user_roles.user_id").
		Where("users.email = ?", email).Pluck("roles.name", &names).Error)
	return names
}

func TestSeedUsers_GrantsAdminRoleToFirstUser(t *testing.T) {
	db := setupSeederDB(t)
	require.NoError(t, db.Create(&models.Role{Name: models.RoleAdmin}).Error)

	require.NoError(t, seeders.SeedUsers(db))

	assert.Equal(t, []string{models.RoleAdmin}, rolesOf(t, db, "john@example.com"))
	assert.Empty(t, rolesOf(t, db, "jane@example.com"))
}

func TestSeedUsers_WithoutAdminRoleStillSeedsUsers(t *testing.T) {
	db := setupSeederDB(t)

	require.NoError(t, seeders.SeedUsers(db))

	var count int64
	db.Model(&models.User{}).Count(&count)
	assert.EqualValues(t, 2, count)
}
