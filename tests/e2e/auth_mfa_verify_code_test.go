package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

func TestAuthMfaVerifyCode(t *testing.T) {
	router, db := setupTestRouter()

	password := "password123"
	hashedPassword := utils.HashPassword(password)

	// Generate a real TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GolangCMS",
		AccountName: "test_mfa@example.com",
	})
	require.NoError(t, err)
	secret := key.Secret()

	user := models.User{
		Name:     "Test User MFA",
		Email:    "test_mfa@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	// Enable MFA for the user
	mfaSettings := models.MfaSettings{
		UserID:     user.ID,
		MfaEnabled: true,
		TotpSecret: &secret,
	}
	db.Create(&mfaSettings)

	// Login to get temporary token
	loginPayload := map[string]string{
		"email":    "test_mfa@example.com",
		"password": password,
	}
	payloadBytes, _ := json.Marshal(loginPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var mfaResponse services.MfaRequiredResponse
	err = json.Unmarshal(w.Body.Bytes(), &mfaResponse)
	require.NoError(t, err)
	require.True(t, mfaResponse.MfaRequired)
	tempToken := mfaResponse.TemporaryToken

	t.Run("Verify MFA Code - Success", func(t *testing.T) {
		// Generate valid code
		code, err := totp.GenerateCode(secret, time.Now())
		require.NoError(t, err)

		verifyPayload := map[string]string{
			"code": code,
		}
		payloadBytes, _ := json.Marshal(verifyPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/mfa/verify-code", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tempToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response services.LoginResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AccessToken.Token)
		assert.NotEmpty(t, response.RefreshToken.Token)
	})

	t.Run("Verify MFA Code - Invalid Code", func(t *testing.T) {
		verifyPayload := map[string]string{
			"code": "000000",
		}
		payloadBytes, _ := json.Marshal(verifyPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/mfa/verify-code", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tempToken)

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusOK, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, 5003, errResp.Code) // ErrMfaInvalidCode
	})

	t.Run("Verify MFA Code - Missing Token", func(t *testing.T) {
		// Generate valid code
		code, err := totp.GenerateCode(secret, time.Now())
		require.NoError(t, err)

		verifyPayload := map[string]string{
			"code": code,
		}
		payloadBytes, _ := json.Marshal(verifyPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/mfa/verify-code", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errResp ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, 3000, errResp.Code) // ErrUnauthorized
	})
}
