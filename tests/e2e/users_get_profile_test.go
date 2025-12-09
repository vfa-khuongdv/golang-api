package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

func TestUsersGetProfile(t *testing.T) {
	router, db := setupTestRouter()

	// Create test user
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	birthday := "1990-05-15"
	address := "123 Test Street"
	testUser := models.User{
		Name:     "Test User",
		Email:    "testuser@example.com",
		Password: hashedPassword,
		Birthday: &birthday,
		Address:  &address,
		Gender:   1,
	}
	db.Create(&testUser)

	// Generate access token for test user
	jwtService := services.NewJWTService()
	tokenResult, err := jwtService.GenerateAccessToken(testUser.ID)
	require.NoError(t, err)
	accessToken := tokenResult.Token

	t.Run("Get Profile - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, response.ID)
		assert.Equal(t, testUser.Email, response.Email)
		assert.Equal(t, testUser.Name, response.Name)
		// Birthday may be returned as ISO timestamp, so check if it starts with the expected date
		assert.True(t, strings.HasPrefix(*response.Birthday, birthday), "Expected birthday to start with %s, got %s", birthday, *response.Birthday)
		assert.Equal(t, address, *response.Address)
		assert.Equal(t, int16(1), response.Gender)
	})

	t.Run("Get Profile - Unauthorized without Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Get Profile - Invalid Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer invalid_token_here")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
