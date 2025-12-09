package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestUsersGetUser(t *testing.T) {
	router, db := setupTestRouter()

	// Create test user
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	testUser := models.User{
		Name:     "Test User",
		Email:    "testuser@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	db.Create(&testUser)

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

	t.Run("Get User - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", testUser.ID), nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, response.ID)
		assert.Equal(t, testUser.Email, response.Email)
		assert.Equal(t, testUser.Name, response.Name)
	})

	t.Run("Get User - Not Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users/99999", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrNotFound, errResp.Code)
	})

	t.Run("Get User - Invalid ID Format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users/invalid", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrParseError, errResp.Code)
	})

	t.Run("Get User - Unauthorized without Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", testUser.ID), nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
