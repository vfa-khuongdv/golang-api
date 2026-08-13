package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
	"gorm.io/gorm"
)

type AuthServiceTestSuite struct {
	suite.Suite
	repo                *mocks.MockUserRepository
	refreshTokenService *mocks.MockRefreshTokenService
	service             services.AuthService
	jwtService          *mocks.MockJWTService
}

func (s *AuthServiceTestSuite) SetupTest() {
	s.repo = new(mocks.MockUserRepository)
	s.refreshTokenService = new(mocks.MockRefreshTokenService)
	s.jwtService = new(mocks.MockJWTService)

	s.service = services.NewAuthService(
		s.repo,
		s.refreshTokenService,
		s.jwtService,
	)
}

// ------------------------ LOGIN TESTS ------------------------
func (s *AuthServiceTestSuite) TestLogin() {
	email := "test@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	tests := []struct {
		name       string
		setupMocks func()
		expectErr  bool
		errCode    int
		errMsg     string
	}{
		{
			name: "Success",
			setupMocks: func() {
				hashedPassword, _ := utils.HashPassword(password)
				user := &models.User{ID: 1, Email: email, Password: hashedPassword}
				s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)
				s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{
					Token:     "mocked-access-token",
					ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				}, nil)
				s.refreshTokenService.On("Create", mock.Anything, user, ipAddress).Return(&dto.JwtResult{
					Token:     "mocked-refresh-token",
					ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
				}, nil)
			},
		},
		{
			name: "UserNotFound",
			setupMocks: func() {
				s.repo.On("FindByField", mock.Anything, "email", email).Return((*models.User)(nil), gorm.ErrRecordNotFound)
			},
			expectErr: true,
			errCode:   apperror.ErrInvalidPassword,
		},
		{
			name: "InvalidPassword",
			setupMocks: func() {
				user := &models.User{ID: 1, Email: email, Password: "wrong-hashed-password"}
				s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)
				s.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expectErr: true,
			errCode:   apperror.ErrInvalidPassword,
		},
		{
			name: "JwtError",
			setupMocks: func() {
				hashedPassword, _ := utils.HashPassword(password)
				user := &models.User{ID: 1, Email: email, Password: hashedPassword}
				s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)
				s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{}, errors.New("Failed to generate JWT token"))
			},
			expectErr: true,
			errCode:   apperror.ErrInternalServer,
		},
		{
			name: "RefreshTokenCreateError",
			setupMocks: func() {
				hashedPassword, _ := utils.HashPassword(password)
				user := &models.User{ID: 1, Email: email, Password: hashedPassword}
				s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)
				s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{
					Token:     "mocked-access-token",
					ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				}, nil)
				s.refreshTokenService.On("Create", mock.Anything, user, ipAddress).Return((*dto.JwtResult)(nil), errors.New("refresh create failed"))
			},
			expectErr: true,
			errMsg:    "refresh create failed",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			// reset mocks for each subtest
			s.SetupTest()
			tt.setupMocks()

			resp, err := s.service.Login(context.Background(), email, password, ipAddress)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
				if appErr, ok := err.(*apperror.AppError); ok {
					assert.Equal(t, tt.errCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "mocked-access-token", resp.AccessToken.Token)
				assert.NotZero(t, resp.AccessToken.ExpiresAt)
				assert.Equal(t, "mocked-refresh-token", resp.RefreshToken.Token)
				assert.NotZero(t, resp.RefreshToken.ExpiresAt)
				s.repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			}
			s.repo.AssertExpectations(t)
			s.refreshTokenService.AssertExpectations(t)
			s.jwtService.AssertExpectations(t)
		})
	}
}

// --------------------- REFRESH TOKEN TESTS ---------------------
func (s *AuthServiceTestSuite) TestRefreshToken() {
	oldRefreshToken := "old-refresh-token"
	oldAccessToken := "old-access-token"
	ipAddress := "127.0.0.1"
	userID := uint(1)

	tests := []struct {
		name       string
		setupMocks func()
		expectErr  bool
		errCode    int
	}{
		{
			name: "Success",
			setupMocks: func() {
				mockRefreshToken := &dto.JwtResult{Token: "new-refresh-token", ExpiresAt: time.Now().Add(24 * time.Hour).Unix()}
				mockRes := &dto.RefreshTokenResult{UserId: userID, Token: mockRefreshToken}
				user := &models.User{ID: userID, Email: "user@example.com"}
				claims := &services.CustomClaims{ID: userID, Scope: services.TokenScopeAccess}

				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress).Return(mockRes, nil)
				s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(claims, nil)
				s.repo.On("GetByID", mock.Anything, userID).Return(user, nil)
				s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{
					Token:     "new-access-token",
					ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				}, nil)
			},
		},
		{
			name: "UpdateError",
			setupMocks: func() {
				claims := &services.CustomClaims{ID: userID, Scope: services.TokenScopeAccess}
				s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(claims, nil)
				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress).Return(nil, apperror.NewUnauthorizedError("Invalid refresh token"))
			},
			expectErr: true,
			errCode:   apperror.ErrUnauthorized,
		},
		{
			name: "GetByIDError",
			setupMocks: func() {
				mockRefreshToken := &dto.JwtResult{Token: "new-refresh-token", ExpiresAt: time.Now().Add(24 * time.Hour).Unix()}
				mockRes := &dto.RefreshTokenResult{UserId: userID, Token: mockRefreshToken}
				claims := &services.CustomClaims{ID: userID, Scope: services.TokenScopeAccess}

				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress).Return(mockRes, nil)
				s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(claims, nil)
				s.repo.On("GetByID", mock.Anything, userID).Return((*models.User)(nil), gorm.ErrRecordNotFound)
			},
			expectErr: true,
			errCode:   apperror.ErrNotFound,
		},
		{
			name: "JwtError",
			setupMocks: func() {
				mockRefreshToken := &dto.JwtResult{Token: "new-refresh-token", ExpiresAt: time.Now().Add(24 * time.Hour).Unix()}
				mockRes := &dto.RefreshTokenResult{UserId: userID, Token: mockRefreshToken}
				user := &models.User{ID: userID, Email: "user@example.com"}
				claims := &services.CustomClaims{ID: userID, Scope: services.TokenScopeAccess}

				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress).Return(mockRes, nil)
				s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(claims, nil)
				s.repo.On("GetByID", mock.Anything, userID).Return(user, nil)
				s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{}, errors.New("Failed to generate JWT token"))
			},
			expectErr: true,
			errCode:   apperror.ErrInternalServer,
		},
		{
			name: "InvalidAccessToken",
			setupMocks: func() {
				// The invalid access token must fail BEFORE the refresh token is
				// touched, so no rotation happens and no Update is expected.
				s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(nil, errors.New("Invalid token signature"))
			},
			expectErr: true,
			errCode:   apperror.ErrUnauthorized,
		},
		{
			name: "TokenMismatch",
			setupMocks: func() {
				refreshUserID := userID
				accessUserID := uint(2)
				mockRefreshToken := &dto.JwtResult{Token: "new-refresh-token", ExpiresAt: time.Now().Add(24 * time.Hour).Unix()}
				mockRes := &dto.RefreshTokenResult{UserId: refreshUserID, Token: mockRefreshToken}
				claims := &services.CustomClaims{ID: accessUserID, Scope: services.TokenScopeAccess}

				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress).Return(mockRes, nil)
				s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(claims, nil)
			},
			expectErr: true,
			errCode:   apperror.ErrUnauthorized,
		},
		{
			name: "InvalidAccessTokenScope",
			setupMocks: func() {
				claims := &services.CustomClaims{ID: userID, Scope: "other-scope"}
				s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(claims, nil)
			},
			expectErr: true,
			errCode:   apperror.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			// reset mocks per subtest
			s.SetupTest()
			tt.setupMocks()

			result, err := s.service.RefreshToken(context.Background(), oldRefreshToken, oldAccessToken, ipAddress)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if appErr, ok := err.(*apperror.AppError); ok {
					assert.Equal(t, tt.errCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "new-access-token", result.AccessToken.Token)
				assert.NotZero(t, result.AccessToken.ExpiresAt)
				assert.Equal(t, "new-refresh-token", result.RefreshToken.Token)
				assert.NotZero(t, result.RefreshToken.ExpiresAt)
			}
			s.repo.AssertExpectations(t)
			s.refreshTokenService.AssertExpectations(t)
			s.jwtService.AssertExpectations(t)
		})
	}
}

// Regression: an invalid access token must be rejected BEFORE the refresh
// token is rotated, so an attacker cannot burn a stolen refresh token by
// sending it with a garbage access token.
func (s *AuthServiceTestSuite) TestRefreshTokenSkipsRotationOnInvalidAccessToken() {
	s.SetupTest()

	oldRefreshToken := "old-refresh-token"
	oldAccessToken := "garbage-access-token"
	ipAddress := "127.0.0.1"

	s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(nil, errors.New("Invalid token signature"))

	result, err := s.service.RefreshToken(context.Background(), oldRefreshToken, oldAccessToken, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	if appErr, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrUnauthorized, appErr.Code)
	}

	s.jwtService.AssertExpectations(s.T())
	s.refreshTokenService.AssertNotCalled(s.T(), "Update", mock.Anything, oldRefreshToken, ipAddress)
	s.refreshTokenService.AssertExpectations(s.T())
	s.repo.AssertExpectations(s.T())
}

// --------------------- LOGOUT TESTS ---------------------
func (s *AuthServiceTestSuite) TestLogout() {
	userID := uint(1)

	s.refreshTokenService.On("DeleteByUserID", mock.Anything, userID).Return(nil)

	err := s.service.Logout(context.Background(), userID)

	assert.NoError(s.T(), err)
	s.refreshTokenService.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogout_Error() {
	userID := uint(1)

	s.refreshTokenService.On("DeleteByUserID", mock.Anything, userID).Return(errors.New("delete failed"))

	err := s.service.Logout(context.Background(), userID)

	assert.Error(s.T(), err)
	s.refreshTokenService.AssertExpectations(s.T())
}

// --------------------- LOCKOUT TESTS ---------------------
func (s *AuthServiceTestSuite) TestLogin_AccountLocked() {
	email := "locked@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	lockedUntil := time.Now().Add(30 * time.Minute).Unix()
	user := &models.User{ID: 1, Email: email, Password: "irrelevant", LockedUntil: &lockedUntil}
	s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)
	if appErr, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrAccountLocked, appErr.Code)
	}
}

func (s *AuthServiceTestSuite) TestLogin_WithFailedAttemptsResetOnSuccess() {
	email := "reset@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	hashedPassword, _ := utils.HashPassword(password)
	user := &models.User{ID: 1, Email: email, Password: hashedPassword, FailedAttempts: 3}
	s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)
	s.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{
		Token:     "token",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}, nil)
	s.refreshTokenService.On("Create", mock.Anything, user, ipAddress).Return(&dto.JwtResult{
		Token:     "refresh",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil)

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), 0, user.FailedAttempts)
	assert.Nil(s.T(), user.LockedUntil)
}

func (s *AuthServiceTestSuite) TestLogin_ExpiredLockResetOnSuccess() {
	email := "expiredlock@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	hashedPassword, _ := utils.HashPassword(password)
	expiredLock := time.Now().Add(-30 * time.Minute).Unix()
	user := &models.User{ID: 1, Email: email, Password: hashedPassword, FailedAttempts: 3, LockedUntil: &expiredLock}
	s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)
	s.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{
		Token:     "token",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}, nil)
	s.refreshTokenService.On("Create", mock.Anything, user, ipAddress).Return(&dto.JwtResult{
		Token:     "refresh",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil)

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), 0, user.FailedAttempts)
	assert.Nil(s.T(), user.LockedUntil)
}

func (s *AuthServiceTestSuite) TestLogin_InvalidPasswordUpdateError() {
	email := "updatefail@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	user := &models.User{ID: 1, Email: email, Password: "wrong-hashed-password"}
	s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)
	s.repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)
}

func (s *AuthServiceTestSuite) TestLogin_LockoutAfterMaxFailedAttempts() {
	email := "lockout@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	user := &models.User{ID: 1, Email: email, Password: "wrong-hashed", FailedAttempts: 4}
	s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)
	s.repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)
	assert.Equal(s.T(), 5, user.FailedAttempts)
	if s.NotNil(user.LockedUntil) {
		// LockedUntil should be ~ now + LockoutDurationMinutes
		expected := time.Now().Add(time.Duration(services.LockoutDurationMinutes) * time.Minute).Unix()
		assert.InDelta(s.T(), expected, *user.LockedUntil, 60)
	}
}

func (s *AuthServiceTestSuite) TestLogin_ResetFailedAttemptsUpdateError() {
	email := "resetfail@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	hashedPassword, _ := utils.HashPassword(password)
	user := &models.User{ID: 1, Email: email, Password: hashedPassword, FailedAttempts: 2}
	s.repo.On("FindByField", mock.Anything, "email", email).Return(user, nil)
	s.repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))
	s.jwtService.On("GenerateAccessToken", user.ID).Return(&dto.JwtResult{
		Token:     "token",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}, nil)
	s.refreshTokenService.On("Create", mock.Anything, user, ipAddress).Return(&dto.JwtResult{
		Token:     "refresh",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil)

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
}

// --------------------- RUN TEST SUITE ---------------------
func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
