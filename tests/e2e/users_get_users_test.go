package e2e

import (
	"encoding/json"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

func TestUsersGetUsers(t *testing.T) {
	router, db := setupTestRouter()

	// Create test users
	password := "password123"
	hashedPassword := utils.HashPassword(password)

	users := []models.User{
		{
			Name:     "User One",
			Email:    "user1@example.com",
			Password: hashedPassword,
			Gender:   1,
		},
		{
			Name:     "User Two",
			Email:    "user2@example.com",
			Password: hashedPassword,
			Gender:   2,
		},
		{
			Name:     "User Three",
			Email:    "user3@example.com",
			Password: hashedPassword,
			Gender:   3,
		},
	}

	for i := range users {
		result := db.Create(&users[i])
		require.NoError(t, result.Error)
	}

	// Create an authenticated user and generate token
	authUser := models.User{
		Name:     "Auth User",
		Email:    "auth@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	db.Create(&authUser)

	// Generate access token
	jwtService := services.NewJWTService()
	tokenResult, err := jwtService.GenerateAccessToken(authUser.ID)
	require.NoError(t, err)
	accessToken := tokenResult.Token

	t.Run("Get Users - Success with Default Pagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.Pagination
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check pagination structure
		assert.NotNil(t, response.Data)
		assert.Equal(t, 1, response.Page)
		assert.GreaterOrEqual(t, response.TotalItems, 4) // At least 4 users (3 test + 1 auth)
	})

	t.Run("Get Users - Success with Custom Pagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users?page=1&limit=2", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.Pagination
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 2, response.Limit)

		data := response.Data.([]interface{})
		assert.LessOrEqual(t, len(data), 2)
	})

	t.Run("Get Users - Unauthorized without Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Get Users - Success with Page 2", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users?page=2&limit=2", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.Pagination
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.Page)
	})
}

func TestUsersGetUsersEmpty(t *testing.T) {
	router, db := setupTestRouter()

	// Create only an authenticated user
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	authUser := models.User{
		Name:     "Auth User",
		Email:    "auth@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	db.Create(&authUser)

	// Generate access token
	jwtService := services.NewJWTService()
	tokenResult, err := jwtService.GenerateAccessToken(authUser.ID)
	require.NoError(t, err)
	accessToken := tokenResult.Token

	t.Run("Get Users - Success with Single User", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.Pagination
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response.Data.([]interface{})
		assert.Equal(t, 1, len(data))
	})
}
