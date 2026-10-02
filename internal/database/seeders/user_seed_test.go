package seeders_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/database/seeders"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
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

	require.NoError(t, seeders.SeedUsers(db, "password123"))

	assert.Equal(t, []string{models.RoleAdmin}, rolesOf(t, db, "john@example.com"))
	assert.Empty(t, rolesOf(t, db, "jane@example.com"))
}

func TestSeedUsers_WithoutAdminRoleStillSeedsUsers(t *testing.T) {
	db := setupSeederDB(t)

	require.NoError(t, seeders.SeedUsers(db, "password123"))

	var count int64
	db.Model(&models.User{}).Count(&count)
	assert.EqualValues(t, 2, count)
}

func TestSeedUsers_UsesTheGivenPasswordAndAValidGender(t *testing.T) {
	db := setupSeederDB(t)

	require.NoError(t, seeders.SeedUsers(db, "Chosen@Secret1"))

	var users []models.User
	require.NoError(t, db.Find(&users).Error)
	require.Len(t, users, 2)
	for _, u := range users {
		assert.True(t, utils.CheckPasswordHash("Chosen@Secret1", u.Password), u.Email)
		assert.Contains(t, []models.Gender{models.GenderMale, models.GenderFemale, models.GenderOther}, u.Gender, u.Email)
	}
}

func TestSeedPassword(t *testing.T) {
	t.Run("Uses The Configured Password", func(t *testing.T) {
		for _, stage := range []string{"dev", "prod"} {
			password, err := seeders.SeedPassword(stage, "Chosen@Secret1")

			require.NoError(t, err)
			assert.Equal(t, "Chosen@Secret1", password, stage)
		}
	})

	t.Run("Falls Back To The Development Password Outside Prod", func(t *testing.T) {
		password, err := seeders.SeedPassword("dev", "")

		require.NoError(t, err)
		assert.Equal(t, seeders.DevPassword, password)
	})

	t.Run("Refuses The Well-Known Development Password In Prod", func(t *testing.T) {
		_, err := seeders.SeedPassword("prod", "")

		assert.ErrorContains(t, err, "SEED_USER_PASSWORD")
	})
}
