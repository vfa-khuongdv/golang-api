package repositories_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupUserTestDB creates an in-memory SQLite database for testing
func setupUserTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NotNil(t, db)

	// Auto-migrate the models
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	return db
}

func TestUserRepository(t *testing.T) {
	t.Run("GetAll - Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUsers := []*models.User{
			{Name: "User1", Email: "email1@example.com", Password: "password1", Gender: 1},
			{Name: "User2", Email: "email2@example.com", Password: "password2", Gender: 1},
		}
		for _, user := range mockUsers {
			_, err := repo.Create(user)
			require.NoError(t, err)
		}

		// Act
		users, err := repo.GetAll()

		// Assert
		require.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("GetAll - Database Error", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		err = sqlDB.Close()
		require.NoError(t, err)

		// Act
		users, err := repo.GetAll()

		// Assert
		assert.Error(t, err)
		assert.Nil(t, users)
	})

	t.Run("GetByID - Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUser := &models.User{
			Name:     "User1",
			Email:    "email1@example.com",
			Password: "password1",
			Gender:   1,
		}
		createdUser, err := repo.Create(mockUser)
		require.NoError(t, err)

		// Act
		user, err := repo.GetByID(createdUser.ID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "User1", user.Name)
	})

	t.Run("GetByID - Not Found Error", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)

		// Act
		user, err := repo.GetByID(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("Create - Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUser := &models.User{
			Name:     "New User",
			Email:    "email@example.com",
			Password: "password",
			Gender:   1,
		}

		// Act
		createdUser, err := repo.Create(mockUser)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, createdUser)
		assert.Equal(t, "New User", createdUser.Name)
		assert.NotEqual(t, uint(0), createdUser.ID)
	})

	t.Run("Create - Duplicate Email Error", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		user1 := &models.User{
			Email:    "test@example.com",
			Name:     "testuser",
			Password: "pass",
		}
		user2 := &models.User{
			Email:    "test@example.com",
			Name:     "anotheruser",
			Password: "pass",
		}
		createdUser, err := repo.Create(user1)
		require.NoError(t, err)
		require.NotNil(t, createdUser)

		// Act
		createdUser2, err := repo.Create(user2)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, createdUser2)
	})

	t.Run("Delete - Database Error", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		err = sqlDB.Close()
		require.NoError(t, err)

		// Act
		err = repo.Delete(999)

		// Assert
		assert.Error(t, err)
	})

	t.Run("FindByField - Find By Email Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUser := &models.User{
			Name:     "Find User",
			Email:    "email@example.com",
			Password: "password",
			Gender:   1,
		}
		_, err := repo.Create(mockUser)
		require.NoError(t, err)

		// Act
		foundUser, err := repo.FindByField("email", "email@example.com")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, "Find User", foundUser.Name)
	})

	t.Run("FindByField - Find By Name Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUser := &models.User{
			Name:     "Another User",
			Email:    "another@example.com",
			Password: "password",
			Gender:   1,
		}
		_, err := repo.Create(mockUser)
		require.NoError(t, err)

		// Act
		foundUser, err := repo.FindByField("name", "Another User")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, "Another User", foundUser.Name)
	})

	t.Run("FindByField - Find By Token Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUser := &models.User{
			Name:     "Token User",
			Email:    "token@example.com",
			Password: "password",
			Token:    utils.StringToPtr("token123"),
			Gender:   1,
		}
		_, err := repo.Create(mockUser)
		require.NoError(t, err)

		// Act
		foundUser, err := repo.FindByField("token", "token123")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, "Token User", foundUser.Name)
	})

	t.Run("FindByField - Not Found Error", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)

		// Act
		user, err := repo.FindByField("email", "notfound@example.com")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("FindByField - Invalid Field Error", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)

		// Act
		user, err := repo.FindByField("sql;", "Invalid")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("Update - Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUser := &models.User{
			Name:     "Update User",
			Email:    "update@example.com",
			Password: "password",
			Gender:   1,
		}
		createdUser, err := repo.Create(mockUser)
		require.NoError(t, err)

		// Update fields
		createdUser.Name = "Updated User"
		createdUser.Password = "newpassword"

		// Act
		err = repo.Update(createdUser)

		// Assert
		require.NoError(t, err)

		// Verify update
		updatedUser, err := repo.GetByID(createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated User", updatedUser.Name)
		assert.Equal(t, "newpassword", updatedUser.Password)
	})

	t.Run("CreateWithTx - Duplicate Email Error", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		user1 := &models.User{
			Email:    "duplicate@example.com",
			Name:     "user1",
			Password: "pass",
		}
		user2 := &models.User{
			Email:    "duplicate@example.com",
			Name:     "user2",
			Password: "pass",
		}
		err := db.Create(user1).Error
		require.NoError(t, err)

		tx := db.Begin()
		require.NoError(t, tx.Error)

		// Act
		createdUser, err := repo.CreateWithTx(tx, user2)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, createdUser)

		tx.Rollback()
	})

	t.Run("CreateWithTx - Success", func(t *testing.T) {
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)

		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		user := &models.User{
			Email:    "tx-success@example.com",
			Name:     "tx-user",
			Password: "pass",
			Gender:   1,
		}

		createdUser, err := repo.CreateWithTx(tx, user)
		require.NoError(t, err)
		require.NotNil(t, createdUser)
		assert.NotZero(t, createdUser.ID)
	})

	t.Run("GetDB - Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)

		// Act
		returnedDB := repo.GetDB()

		// Assert
		assert.NotNil(t, returnedDB)
	})

	t.Run("GetUsers - Pagination Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUsers := []*models.User{
			{Name: "User1", Email: "user1@example.com", Password: "password1", Gender: 1},
			{Name: "User2", Email: "user2@example.com", Password: "password2", Gender: 2},
			{Name: "User3", Email: "user3@example.com", Password: "password3", Gender: 1},
			{Name: "User4", Email: "user4@example.com", Password: "password4", Gender: 2},
			{Name: "User5", Email: "user5@example.com", Password: "password5", Gender: 1},
		}
		for _, user := range mockUsers {
			_, err := repo.Create(user)
			require.NoError(t, err)
		}

		// Act - First page
		pagination, err := repo.GetUsers(1, 2)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, pagination)
		assert.Equal(t, 1, pagination.Page)
		assert.Equal(t, 2, pagination.Limit)
		assert.Equal(t, 5, pagination.TotalItems)
		assert.Equal(t, 3, pagination.TotalPages)
		assert.Len(t, pagination.Data, 2)
	})

	t.Run("GetUsers - Second Page Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUsers := []*models.User{
			{Name: "User1", Email: "user1@example.com", Password: "password1", Gender: 1},
			{Name: "User2", Email: "user2@example.com", Password: "password2", Gender: 2},
			{Name: "User3", Email: "user3@example.com", Password: "password3", Gender: 1},
			{Name: "User4", Email: "user4@example.com", Password: "password4", Gender: 2},
			{Name: "User5", Email: "user5@example.com", Password: "password5", Gender: 1},
		}
		for _, user := range mockUsers {
			_, err := repo.Create(user)
			require.NoError(t, err)
		}

		// Act - Second page
		pagination, err := repo.GetUsers(2, 2)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, pagination)
		assert.Equal(t, 2, pagination.Page)
		assert.Equal(t, 2, pagination.Limit)
		assert.Len(t, pagination.Data, 2)
	})

	t.Run("GetUsers - Query Error On Find", func(t *testing.T) {
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)

		_ = db.Callback().Query().Before("gorm:query").Register("force_find_error_only", func(tx *gorm.DB) {
			if tx.Statement != nil {
				_, hasOrderBy := tx.Statement.Clauses["ORDER BY"]
				if hasOrderBy {
				_ = tx.AddError(assert.AnError)
				}
			}
		})
		defer db.Callback().Query().Remove("force_find_error_only")

		_, err := repo.GetUsers(1, 10)
		assert.Error(t, err)
	})

	t.Run("GetUsers - Last Page Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUsers := []*models.User{
			{Name: "User1", Email: "user1@example.com", Password: "password1", Gender: 1},
			{Name: "User2", Email: "user2@example.com", Password: "password2", Gender: 2},
			{Name: "User3", Email: "user3@example.com", Password: "password3", Gender: 1},
			{Name: "User4", Email: "user4@example.com", Password: "password4", Gender: 2},
			{Name: "User5", Email: "user5@example.com", Password: "password5", Gender: 1},
		}
		for _, user := range mockUsers {
			_, err := repo.Create(user)
			require.NoError(t, err)
		}

		// Act - Last page
		pagination, err := repo.GetUsers(3, 2)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, pagination)
		assert.Equal(t, 3, pagination.Page)
		assert.Len(t, pagination.Data, 1)
	})

	t.Run("GetUsers - Out of Range Page", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUsers := []*models.User{
			{Name: "User1", Email: "user1@example.com", Password: "password1", Gender: 1},
			{Name: "User2", Email: "user2@example.com", Password: "password2", Gender: 2},
		}
		for _, user := range mockUsers {
			_, err := repo.Create(user)
			require.NoError(t, err)
		}

		// Act
		pagination, err := repo.GetUsers(5, 2)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, pagination)
		assert.Equal(t, 5, pagination.Page)
		assert.Len(t, pagination.Data, 0)
	})

	t.Run("GetUsers - Large Limit Success", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		mockUsers := []*models.User{
			{Name: "User1", Email: "user1@example.com", Password: "password1", Gender: 1},
			{Name: "User2", Email: "user2@example.com", Password: "password2", Gender: 2},
			{Name: "User3", Email: "user3@example.com", Password: "password3", Gender: 1},
		}
		for _, user := range mockUsers {
			_, err := repo.Create(user)
			require.NoError(t, err)
		}

		// Act
		pagination, err := repo.GetUsers(1, 10)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, pagination)
		assert.Equal(t, 1, pagination.Page)
		assert.Equal(t, 10, pagination.Limit)
		assert.Equal(t, 3, pagination.TotalItems)
		assert.Equal(t, 1, pagination.TotalPages)
		assert.Len(t, pagination.Data, 3)
	})

	t.Run("GetUsers - Database Error", func(t *testing.T) {
		// Arrange
		db := setupUserTestDB(t)
		repo := repositories.NewUserRepository(db)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		err = sqlDB.Close()
		require.NoError(t, err)

		// Act
		pagination, err := repo.GetUsers(1, 10)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, pagination)
	})
}
