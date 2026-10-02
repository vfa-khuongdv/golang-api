package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestAuthLogin(t *testing.T) {
	router, db := setupTestRouter()

	// Helper to create a user directly in DB
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)
	user := models.User{
		Name:     "Test User",
		Email:    "test_login@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	t.Run("Login - Success", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "test_login@example.com",
			"password": password,
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AccessToken.Token)
		assert.NotEmpty(t, response.RefreshToken.Token)
	})

	t.Run("Login - Invalid Credentials", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "test_login@example.com",
			"password": "wrongpassword",
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrInvalidPassword, errResp.Code)
	})

	t.Run("Login - Missing Fields", func(t *testing.T) {
		loginPayload := map[string]string{
			"email": "test_login@example.com",
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Login - User Not Found", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "nonexistent@example.com",
			"password": password,
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrInvalidPassword, errResp.Code)
	})

	t.Run("Login - Empty Email", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "",
			"password": password,
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Login - Empty Password", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "test_login@example.com",
			"password": "",
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Login - Invalid Email Format", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "not-an-email",
			"password": password,
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})
}

func TestAuthLoginLockout(t *testing.T) {
	// Separate test function so it gets its own router (and thus its own
	// rate limiter), avoiding interference with TestAuthLogin's request count.
	router, db := setupTestRouter()

	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)
	user := models.User{
		Name:     "Test User Lockout",
		Email:    "test_lockout@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	loginPayload := map[string]string{
		"email":    "test_lockout@example.com",
		"password": "wrongpassword",
	}
	payloadBytes, _ := json.Marshal(loginPayload)

	// MaxFailedAttempts is 5; each failed attempt increments FailedAttempts.
	for i := 0; i < services.MaxFailedAttempts; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	login := func(email, password string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]string{"email": email, "password": password})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		return w
	}

	// The account is now locked. Whatever the password, the answer is the same
	// as for an unknown email: a distinct one would reveal the account exists,
	// and one that differs for the right password would let guessing go on.
	unknown := login("nobody-lockout@example.com", "wrongpassword")
	lockedWrong := login("test_lockout@example.com", "wrongpassword")
	lockedRight := login("test_lockout@example.com", password)

	assert.Equal(t, http.StatusBadRequest, unknown.Code)
	assert.Equal(t, unknown.Code, lockedWrong.Code)
	assert.Equal(t, unknown.Body.String(), lockedWrong.Body.String())
	assert.Equal(t, unknown.Code, lockedRight.Code)
	assert.Equal(t, unknown.Body.String(), lockedRight.Body.String())
}

func TestAuthLogin_ExpiredLockRestartsTheCount(t *testing.T) {
	router, db := setupTestRouter()
	hashed, err := utils.HashPassword("password123")
	require.NoError(t, err)
	expired := time.Now().Add(-time.Minute).Unix()
	user := models.User{Name: "Locked", Email: "expired-lock@example.com", Password: hashed, Gender: 1,
		FailedAttempts: services.MaxFailedAttempts, LockedUntil: &expired}
	require.NoError(t, db.Create(&user).Error)

	login := func(password string) int {
		body, _ := json.Marshal(map[string]string{"email": user.Email, "password": password})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.0.2.10:1234" // its own rate limit bucket
		router.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, http.StatusBadRequest, login("wrong-password"), "one wrong password after the lock ends is not a new lock")

	var got models.User
	require.NoError(t, db.First(&got, user.ID).Error)
	assert.Equal(t, 1, got.FailedAttempts)
	assert.Nil(t, got.LockedUntil)
	assert.Equal(t, http.StatusOK, login("password123"))
}
