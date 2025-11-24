package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/handlers"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

func TestInitMfaSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("InitMfaSetup - Success", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		// Mock the service method
		mockMfaService.On("SetupMfa", uint(1), "test@example.com").Return(
			"JBSWY3DPEBLW64TMMQ======",
			[]byte("mock_qr_code"),
			[]string{"BACKUP1", "BACKUP2", "BACKUP3"},
			nil,
		)

		requestBody := map[string]string{
			"email": "test@example.com",
		}
		reqBody, _ := json.Marshal(requestBody)

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/setup", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		// Call the handler
		handler.InitMfaSetup(c)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseBody))

		assert.Equal(t, "JBSWY3DPEBLW64TMMQ======", responseBody["secret"])
		assert.NotEmpty(t, responseBody["qr_code"])
		assert.NotEmpty(t, responseBody["backup_codes"])

		// Assert that the mock service method was called
		mockMfaService.AssertExpectations(t)
	})

	t.Run("InitMfaSetup - Missing UserID", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		requestBody := map[string]string{
			"email": "test@example.com",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/setup", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		// UserID not set

		handler.InitMfaSetup(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var responseBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseBody))
		assert.NotEmpty(t, responseBody["message"])
	})

	t.Run("InitMfaSetup - Invalid Email", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		requestBody := map[string]string{
			"email": "invalid-email",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/setup", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.InitMfaSetup(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InitMfaSetup - Service Error", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("SetupMfa", uint(1), "test@example.com").Return(
			"",
			[]byte{},
			[]string{},
			apperror.NewInternalError("Database error"),
		)

		requestBody := map[string]string{
			"email": "test@example.com",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/setup", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.InitMfaSetup(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockMfaService.AssertExpectations(t)
	})

	t.Run("InitMfaSetup - MFA Already Enabled", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		// Mock the service to return error when MFA is already enabled
		mockMfaService.On("SetupMfa", uint(1), "test@example.com").Return(
			"",
			[]byte{},
			[]string{},
			apperror.NewBadRequestError("MFA is already enabled for this user"),
		)

		requestBody := map[string]string{
			"email": "test@example.com",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/setup", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.InitMfaSetup(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockMfaService.AssertExpectations(t)
	})
}

func TestVerifyMfaSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("VerifyMfaSetup - Success", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("VerifyMfaSetup", uint(1), "123456").Return(
			[]string{"BACKUP1", "BACKUP2", "BACKUP3"},
			nil,
		)

		requestBody := map[string]string{
			"code": "123456",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify-setup", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.VerifyMfaSetup(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseBody))

		assert.Equal(t, "MFA setup verified successfully", responseBody["message"])
		assert.NotEmpty(t, responseBody["backup_codes"])

		mockMfaService.AssertExpectations(t)
	})

	t.Run("VerifyMfaSetup - Invalid Code Format", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		requestBody := map[string]string{
			"code": "12345", // Invalid - should be 6 digits
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify-setup", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.VerifyMfaSetup(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("VerifyMfaSetup - Invalid Code", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("VerifyMfaSetup", uint(1), "000000").Return(
			[]string{},
			apperror.NewInvalidPasswordError("Invalid MFA code"),
		)

		requestBody := map[string]string{
			"code": "000000",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify-setup", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.VerifyMfaSetup(c)

		// When a service error is returned from RespondWithError, it will determine the status code based on the error type
		// InvalidPasswordError is a 401 error
		var responseBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseBody))
		assert.NotNil(t, responseBody["code"])
		mockMfaService.AssertExpectations(t)
	})

	t.Run("VerifyMfaSetup - Missing UserID", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		requestBody := map[string]string{
			"code": "123456",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify-setup", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		// UserID not set

		handler.VerifyMfaSetup(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestVerifyMfaCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("VerifyMfaCode - Success", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		// Mock the service methods
		mockMfaService.On("VerifyMfaCode", uint(1), "123456").Return(true, nil)

		mockUserRepo.On("GetByID", uint(1)).Return(&models.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
		}, nil)

		mockJwtService.On("GenerateAccessToken", uint(1)).Return(&services.JwtResult{
			Token:     "access_token",
			ExpiresAt: 1000000,
		}, nil)

		mockRefreshTokenService.On("Create", mock.MatchedBy(func(u *models.User) bool {
			return u.ID == 1
		}), mock.Anything).Return(&services.JwtResult{
			Token:     "refresh_token",
			ExpiresAt: 2000000,
		}, nil)

		requestBody := map[string]string{
			"code": "123456",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))
		c.Request.RemoteAddr = "192.168.1.1:8080"

		handler.VerifyMfaCode(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseBody))

		assert.NotEmpty(t, responseBody["access_token"])
		assert.NotEmpty(t, responseBody["refresh_token"])

		mockMfaService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockJwtService.AssertExpectations(t)
		mockRefreshTokenService.AssertExpectations(t)
	})

	t.Run("VerifyMfaCode - Invalid Code", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("VerifyMfaCode", uint(1), "000000").Return(false, nil)

		requestBody := map[string]string{
			"code": "000000",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.VerifyMfaCode(c)

		// When VerifyMfaCode returns false, the handler returns an InvalidPasswordError which is 401
		var responseBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseBody))
		assert.NotNil(t, responseBody["code"])
		mockMfaService.AssertExpectations(t)
	})

	t.Run("VerifyMfaCode - Missing Code", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		requestBody := map[string]string{}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.VerifyMfaCode(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("VerifyMfaCode - User Not Found", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("VerifyMfaCode", uint(1), "123456").Return(true, nil)
		mockUserRepo.On("GetByID", uint(1)).Return(nil, apperror.NewNotFoundError("User not found"))

		requestBody := map[string]string{
			"code": "123456",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.VerifyMfaCode(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockMfaService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("VerifyMfaCode - JWT Generation Error", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("VerifyMfaCode", uint(1), "123456").Return(true, nil)
		mockUserRepo.On("GetByID", uint(1)).Return(&models.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
		}, nil)
		mockJwtService.On("GenerateAccessToken", uint(1)).Return(nil, apperror.NewInternalError("JWT error"))

		requestBody := map[string]string{
			"code": "123456",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.VerifyMfaCode(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockMfaService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockJwtService.AssertExpectations(t)
	})

	t.Run("VerifyMfaCode - Refresh Token Creation Error", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("VerifyMfaCode", uint(1), "123456").Return(true, nil)
		mockUserRepo.On("GetByID", uint(1)).Return(&models.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
		}, nil)
		mockJwtService.On("GenerateAccessToken", uint(1)).Return(&services.JwtResult{
			Token:     "access_token",
			ExpiresAt: 1000000,
		}, nil)
		mockRefreshTokenService.On("Create", mock.MatchedBy(func(u *models.User) bool {
			return u.ID == 1
		}), mock.Anything).Return(nil, apperror.NewInternalError("Refresh token error"))

		requestBody := map[string]string{
			"code": "123456",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))
		c.Request.RemoteAddr = "192.168.1.1:8080"

		handler.VerifyMfaCode(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockMfaService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockJwtService.AssertExpectations(t)
		mockRefreshTokenService.AssertExpectations(t)
	})

	t.Run("VerifyMfaCode - Missing UserID", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		requestBody := map[string]string{
			"code": "123456",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/verify", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		// UserID not set

		handler.VerifyMfaCode(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDisableMfa(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("DisableMfa - Success", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockUserRepo.On("GetByID", uint(1)).Return(&models.User{
			ID:       1,
			Email:    "test@example.com",
			Password: "hashed_password",
		}, nil)

		mockBcryptService.On("CheckPasswordHash", "password123", "hashed_password").Return(true)

		mockMfaService.On("DisableMfa", uint(1)).Return(nil)

		requestBody := map[string]string{
			"password": "password123",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/disable", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.DisableMfa(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseBody))

		assert.Equal(t, "MFA disabled successfully", responseBody["message"])
		mockMfaService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockBcryptService.AssertExpectations(t)
	})

	t.Run("DisableMfa - Service Error", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockUserRepo.On("GetByID", uint(1)).Return(&models.User{
			ID:       1,
			Email:    "test@example.com",
			Password: "hashed_password",
		}, nil)

		mockBcryptService.On("CheckPasswordHash", "password123", "hashed_password").Return(true)

		mockMfaService.On("DisableMfa", uint(1)).Return(apperror.NewInternalError("Database error"))

		requestBody := map[string]string{
			"password": "password123",
		}
		reqBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/disable", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.DisableMfa(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockMfaService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockBcryptService.AssertExpectations(t)
	})

	t.Run("DisableMfa - Missing UserID", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/mfa/disable", bytes.NewReader([]byte{}))
		c.Request.Header.Set("Content-Type", "application/json")
		// UserID not set

		handler.DisableMfa(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetMfaStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetMfaStatus - MFA Enabled", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("GetMfaStatus", uint(1)).Return(true, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/mfa/status", nil)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.GetMfaStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseBody))

		assert.Equal(t, true, responseBody["mfa_enabled"])
		mockMfaService.AssertExpectations(t)
	})

	t.Run("GetMfaStatus - MFA Disabled", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("GetMfaStatus", uint(1)).Return(false, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/mfa/status", nil)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.GetMfaStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseBody))

		assert.Equal(t, false, responseBody["mfa_enabled"])
		mockMfaService.AssertExpectations(t)
	})

	t.Run("GetMfaStatus - Service Error", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		mockMfaService.On("GetMfaStatus", uint(1)).Return(false, apperror.NewInternalError("Database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/mfa/status", nil)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("UserID", uint(1))

		handler.GetMfaStatus(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockMfaService.AssertExpectations(t)
	})

	t.Run("GetMfaStatus - Missing UserID", func(t *testing.T) {
		mockMfaService := new(mocks.MockMfaService)
		mockUserRepo := new(mocks.MockUserRepository)
		mockJwtService := new(mocks.MockJWTService)
		mockRefreshTokenService := new(mocks.MockRefreshTokenService)
		mockBcryptService := new(mocks.MockBcryptService)

		handler := handlers.NewMfaHandler(mockMfaService, mockUserRepo, mockJwtService, mockRefreshTokenService, mockBcryptService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/mfa/status", nil)
		c.Request.Header.Set("Content-Type", "application/json")
		// UserID not set

		handler.GetMfaStatus(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
