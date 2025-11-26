package services_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
	"gorm.io/gorm"
)

// TestSuite is a struct that holds the mock repositories and the service under test
type AuthServiceTestSuite struct {
	suite.Suite
	repo                *mocks.MockUserRepository
	refreshTokenService *mocks.MockRefreshTokenService
	service             services.IAuthService
	bcryptService       *mocks.MockBcryptService
	jwtService          *mocks.MockJWTService
}

func (s *AuthServiceTestSuite) SetupTest() {
	s.repo = new(mocks.MockUserRepository)
	s.refreshTokenService = new(mocks.MockRefreshTokenService)
	s.bcryptService = new(mocks.MockBcryptService)
	s.jwtService = new(mocks.MockJWTService)

	s.service = services.NewAuthService(
		s.repo,
		s.refreshTokenService,
		s.bcryptService,
		s.jwtService,
	)
}

func (s *AuthServiceTestSuite) TestLoginSuccess() {
	// Set up the expected user and mock repository behavior
	email := "test@example.com"
	password := "password123"

	user := &models.User{
		ID:       1,
		Email:    email,
		Password: "hashed_password",
	}
	ip := "127.0.0.1"

	// Mock the methods of the dependencies
	s.repo.On("FindByField", "email", email).Return(user, nil)
	s.bcryptService.On("CheckPasswordHash", password, user.Password).Return(true)
	s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{
		Token:     "mocked-access-token",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}, nil)

	// Mock Create to return a valid JWT result
	s.refreshTokenService.On("Create", user, ip).Return(&dto.JwtResult{
		Token:     "mocked-refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil)

	ginCtx, _ := gin.CreateTestContext(nil)
	ginCtx.Request = &http.Request{RemoteAddr: ip + ":12345"}

	// Call the Login method
	resp, _ := s.service.Login(email, password, ginCtx)

	// Assert response is LoginResponse
	loginResp := resp
	assert.Equal(s.T(), "mocked-refresh-token", loginResp.RefreshToken.Token)

}

func (s *AuthServiceTestSuite) TestLogin_UserNotFound() {
	email := "nonexistent@example.com"
	password := "password123"

	s.repo.On("FindByField", "email", email).Return((*models.User)(nil), gorm.ErrRecordNotFound)

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	resp, err := s.service.Login(email, password, ginCtx)

	// Check if the error is of type AppError
	if appError, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrNotFound, appError.Code)
		assert.Equal(s.T(), "record not found", appError.Message)
	} else {
		s.Fail("Expected AppError with ErrNotFound code")
	}
	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)

	s.repo.AssertExpectations(s.T())

}

func (s *AuthServiceTestSuite) TestLogin_InvalidPassword() {
	email := "test@example.com"
	wrongPassword := "wrongpass"
	user := &models.User{
		ID:       1,
		Email:    email,
		Password: "hashed_password", // Assume this is a invalid hashed password
	}

	s.repo.On("FindByField", "email", email).Return(user, nil)
	s.bcryptService.On("CheckPasswordHash", wrongPassword, user.Password).Return(false).Once()

	ginCtx, _ := gin.CreateTestContext(nil)

	resp, err := s.service.Login(email, wrongPassword, ginCtx)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)

	// Assert that the error is of type AppError with ErrInvalidPassword code
	if appError, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrInvalidPassword, appError.Code)
		assert.Equal(s.T(), "Invalid credentials", appError.Message)
	} else {
		s.Fail("Expected AppError with ErrInvalidPassword code")
	}

	s.repo.AssertExpectations(s.T())

}

func (s *AuthServiceTestSuite) TestLogin_CreateTokenError() {
	email := "test@example.com"
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	user := &models.User{
		ID:       1,
		Email:    email,
		Password: hashedPassword,
	}
	ipAddress := "127.0.0.1"

	// Mock user repository and bcrypt service
	s.repo.On("FindByField", "email", email).Return(user, nil)
	s.bcryptService.On("CheckPasswordHash", password, user.Password).Return(true).Once()
	s.jwtService.On("GenerateAccessToken", user.ID).
		Return(&dto.JwtResult{
			Token:     "mocked-access-token",
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		}, nil).Once()
	s.refreshTokenService.On("Create", user, ipAddress).
		Return(nil, apperror.NewInternalError("Failed to create refresh token")).
		Once()

	// Create a proper gin.Context with ResponseWriter
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = &http.Request{
		RemoteAddr: ipAddress + ":12345",
	}

	resp, err := s.service.Login(email, password, ginCtx)

	// Assert that an error
	if appError, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrInternal, appError.Code)
		assert.Equal(s.T(), "Failed to create refresh token", appError.Message)
	} else {
		s.Fail("Expected AppError with ErrInternal code")
	}
	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)

	s.repo.AssertExpectations(s.T())
	s.refreshTokenService.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogin_JwtError() {
	email := "test@example.com"
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	user := &models.User{
		ID:       1,
		Email:    email,
		Password: hashedPassword,
	}
	ipAddress := "127.0.0.1"

	// Mock user repository and bcrypt service
	s.repo.On("FindByField", "email", email).Return(user, nil)
	s.bcryptService.On("CheckPasswordHash", password, user.Password).Return(true).Once()
	s.jwtService.On("GenerateAccessToken", user.ID).
		Return(&dto.JwtResult{}, errors.New("Failed to generate JWT token")).Once()

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = &http.Request{
		RemoteAddr: ipAddress + ":12345",
	}

	resp, err := s.service.Login(email, password, ginCtx)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)

	// Assert that the error is of type AppError with ErrInternal code
	if appError, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrInternal, appError.Code)
		assert.Equal(s.T(), "Failed to generate JWT token", appError.Message)
	} else {
		s.Fail("Expected AppError with ErrInternal code")
	}

	// Assert mocks
	s.repo.AssertExpectations(s.T())
	s.refreshTokenService.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshToken_Success() {
	// Test input values
	oldRefreshToken := "valid-refresh-token"
	oldAccessToken := "valid-access-token"
	ipAddress := "127.0.0.1"
	userID := uint(1)

	// Mock new refresh token that would be returned by refresh token service
	mockRefreshToken := &dto.JwtResult{
		Token:     "new-refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour * 30).Unix(), // 30 days
	}
	mockRes := &services.RefreshTokenResult{
		UserId: userID,
		Token:  mockRefreshToken,
	}

	// Mock user that would be returned by user repository
	mockUser := &models.User{
		ID:    userID,
		Email: "user@example.com",
	}

	// Mock claims from access token
	mockClaims := &services.CustomClaims{
		ID:    userID,
		Scope: services.TokenScopeAccess,
	}

	// Should update refresh token with correct old token and IP
	s.refreshTokenService.On("Update", oldRefreshToken, ipAddress).Return(mockRes, nil).Once()

	// Should validate access token (even if expired)
	s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(mockClaims, nil).Once()

	s.repo.On("GetByID", mockRes.UserId).Return(mockUser, nil).Once()
	s.jwtService.On("GenerateAccessToken", mockUser.ID).Return(&dto.JwtResult{
		Token:     "new-access-token",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}, nil).Once()

	// Setup gin test context with IP
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = &http.Request{RemoteAddr: ipAddress + ":12345"}

	// Execute the refresh token flow
	result, err := s.service.RefreshToken(oldRefreshToken, oldAccessToken, ginCtx)

	// Assert no errors occurred
	s.NoError(err, "Expected no error from RefreshToken")
	s.NotNil(result, "Expected result not to be nil")

	// Verify response structure and values
	s.NotEmpty(result.AccessToken.Token, "Expected access token to be set")
	s.True(result.AccessToken.ExpiresAt > time.Now().Unix(), "Expected access token to expire in the future")

	// Verify refresh token matches mock
	s.Equal(mockRefreshToken.Token, result.RefreshToken.Token, "Refresh token should match mock")
	s.Equal(mockRefreshToken.ExpiresAt, result.RefreshToken.ExpiresAt, "Refresh token expiry should match mock")

	// Validate mock expectations
	s.refreshTokenService.AssertExpectations(s.T())
	s.repo.AssertExpectations(s.T())
	s.jwtService.AssertExpectations(s.T())
	s.bcryptService.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshToken_UpdateError() {
	// Test input values
	invalidToken := "invalid-refresh-token"
	accessToken := "valid-access-token"
	ipAddress := "127.0.0.1"

	// Mock refresh token service to return error for invalid token
	mockError := apperror.NewNotFoundError("Refresh token not found")
	s.refreshTokenService.On("Update", invalidToken, ipAddress).Return(nil, mockError).Once()

	// Setup gin test context with IP
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = &http.Request{RemoteAddr: ipAddress + ":12345"}

	// Execute the refresh token flow
	result, err := s.service.RefreshToken(invalidToken, accessToken, ginCtx) // Assert error was returned
	s.Error(err, "Expected error for invalid refresh token")
	s.Nil(result, "Expected nil result for error case")

	if appError, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrUnauthorized, appError.Code)
		assert.Equal(s.T(), "Invalid refresh token", appError.Message)
	}

	// Assert mocks
	s.refreshTokenService.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshToken_GetByIDError() {
	oldRefreshToken := "old-refresh-token"
	oldAccessToken := "old-access-token"
	ipAddress := "127.0.0.1"
	userID := uint(1)

	// Mock new refresh token that would be returned by refresh token service
	mockRefreshToken := &dto.JwtResult{
		Token:     "new-refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour * 30).Unix(), // 30 days
	}
	mockRes := &services.RefreshTokenResult{
		UserId: userID,
		Token:  mockRefreshToken,
	}

	// Mock claims from access token
	mockClaims := &services.CustomClaims{
		ID:    userID,
		Scope: services.TokenScopeAccess,
	}

	// Should update refresh token with correct old token and IP
	s.refreshTokenService.On("Update", oldRefreshToken, ipAddress).Return(mockRes, nil).Once()

	// Should validate access token
	s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(mockClaims, nil).Once()

	// Should fetch user with ID from refresh token but fail
	s.repo.On("GetByID", mockRes.UserId).Return((*models.User)(nil), gorm.ErrRecordNotFound).Once()

	// Setup gin test context with IP
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = &http.Request{RemoteAddr: ipAddress + ":12345"}

	// Execute the refresh token flow
	result, err := s.service.RefreshToken(oldRefreshToken, oldAccessToken, ginCtx)
	s.Error(err, "Expected error for valid refresh token")
	s.Nil(result, "Expected nil result for error case")

	// Assert that the error is of type AppError with ErrNotFound code
	if appError, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrNotFound, appError.Code)
		assert.Equal(s.T(), "User not found", appError.Message)
	} else {
		s.Fail("Expected AppError with ErrNotFound code")
	}

	// Assert mocks
	s.repo.AssertExpectations(s.T())
	s.bcryptService.AssertExpectations(s.T())
	s.refreshTokenService.AssertExpectations(s.T())
	s.jwtService.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshToken_JwtError() {
	oldRefreshToken := "old-refresh-token"
	oldAccessToken := "old-access-token"
	userID := uint(1)

	user := &models.User{
		ID:    userID,
		Email: "email@example.com",
	}

	ipAddress := "127.0.0.1"
	// Mock new refresh token that would be returned by refresh token service
	mockRefreshToken := &dto.JwtResult{
		Token:     "new-refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour * 30).Unix(), // 30 days
	}
	mockRes := &services.RefreshTokenResult{
		UserId: userID,
		Token:  mockRefreshToken,
	}

	// Mock claims from access token
	mockClaims := &services.CustomClaims{
		ID:    userID,
		Scope: services.TokenScopeAccess,
	}

	// Should update refresh token with correct old token and IP
	s.refreshTokenService.On("Update", oldRefreshToken, ipAddress).Return(mockRes, nil).Once()

	// Should validate access token
	s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(mockClaims, nil).Once()

	// Should fetch user with ID from refresh token
	s.repo.On("GetByID", mockRes.UserId).Return(user, nil).Once()
	// Should generate new access token for user
	s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{}, errors.New("Failed to generate JWT token")).Once()

	// Setup gin test context with IP
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = &http.Request{RemoteAddr: ipAddress + ":12345"}

	// Execute the refresh token flow
	result, err := s.service.RefreshToken(oldRefreshToken, oldAccessToken, ginCtx)
	s.Error(err, "Expected error for valid refresh token")
	s.Nil(result, "Expected nil result for error case")

	if appError, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrInternal, appError.Code)
		assert.Equal(s.T(), "Failed to generate JWT token", appError.Message)
	} else {
		s.Fail("Expected AppError with ErrInternal code")
	}

	// Assert mocks
	s.refreshTokenService.AssertExpectations(s.T())
	s.repo.AssertExpectations(s.T())
	s.jwtService.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshToken_InvalidAccessToken() {
	// Test input values
	oldRefreshToken := "valid-refresh-token"
	oldAccessToken := "invalid-access-token"
	ipAddress := "127.0.0.1"
	userID := uint(1)

	// Mock new refresh token that would be returned by refresh token service
	mockRefreshToken := &dto.JwtResult{
		Token:     "new-refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour * 30).Unix(),
	}
	mockRes := &services.RefreshTokenResult{
		UserId: userID,
		Token:  mockRefreshToken,
	}

	// Should update refresh token
	s.refreshTokenService.On("Update", oldRefreshToken, ipAddress).Return(mockRes, nil).Once()

	// Should fail to validate access token
	s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(nil, errors.New("Invalid token signature")).Once()

	// Setup gin test context with IP
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = &http.Request{RemoteAddr: ipAddress + ":12345"}

	// Execute the refresh token flow
	result, err := s.service.RefreshToken(oldRefreshToken, oldAccessToken, ginCtx)
	s.Error(err, "Expected error for invalid access token")
	s.Nil(result, "Expected nil result for error case")

	if appError, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrUnauthorized, appError.Code)
	}

	// Assert mocks
	s.refreshTokenService.AssertExpectations(s.T())
	s.jwtService.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshToken_TokenMismatch() {
	// Test input values
	oldRefreshToken := "valid-refresh-token"
	oldAccessToken := "valid-access-token"
	ipAddress := "127.0.0.1"
	refreshUserID := uint(1)
	accessUserID := uint(2)

	// Mock new refresh token that would be returned by refresh token service
	mockRefreshToken := &dto.JwtResult{
		Token:     "new-refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour * 30).Unix(),
	}
	mockRes := &services.RefreshTokenResult{
		UserId: refreshUserID,
		Token:  mockRefreshToken,
	}

	// Mock claims from access token with different user ID
	mockClaims := &services.CustomClaims{
		ID:    accessUserID, // Different from refresh token user ID
		Scope: services.TokenScopeAccess,
	}

	// Should update refresh token
	s.refreshTokenService.On("Update", oldRefreshToken, ipAddress).Return(mockRes, nil).Once()

	// Should validate access token but find user mismatch
	s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(mockClaims, nil).Once()

	// Setup gin test context with IP
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = &http.Request{RemoteAddr: ipAddress + ":12345"}

	// Execute the refresh token flow
	result, err := s.service.RefreshToken(oldRefreshToken, oldAccessToken, ginCtx)
	s.Error(err, "Expected error for token mismatch")
	s.Nil(result, "Expected nil result for error case")

	if appError, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrUnauthorized, appError.Code)
		assert.Contains(s.T(), appError.Message, "Token mismatch")
	}

	// Assert mocks
	s.refreshTokenService.AssertExpectations(s.T())
	s.jwtService.AssertExpectations(s.T())
}

func TestAuthServiceTestSuite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	suite.Run(t, new(AuthServiceTestSuite))
}
