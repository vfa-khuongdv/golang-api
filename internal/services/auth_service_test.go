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

// expectFailedLogin expects the failed attempt to be recorded atomically, with
// a lock of LockoutDurationMinutes from now, and returns err from the repository.
func (s *AuthServiceTestSuite) expectFailedLogin(userID uint, err error) {
	expectedLock := time.Now().Add(services.LockoutDurationMinutes * time.Minute).Unix()
	s.repo.On("RecordFailedLogin", mock.Anything, userID, services.MaxFailedAttempts, mock.MatchedBy(func(lockUntil int64) bool {
		return lockUntil >= expectedLock-5 && lockUntil <= expectedLock+5
	})).Return(err).Once()
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
				s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
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
				s.repo.On("FindByEmail", mock.Anything, email).Return((*models.User)(nil), gorm.ErrRecordNotFound)
			},
			expectErr: true,
			errCode:   apperror.ErrInvalidPassword,
		},
		{
			name: "InvalidPassword",
			setupMocks: func() {
				user := &models.User{ID: 1, Email: email, Password: "wrong-hashed-password"}
				s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
				s.expectFailedLogin(user.ID, nil)
			},
			expectErr: true,
			errCode:   apperror.ErrInvalidPassword,
		},
		{
			name: "JwtError",
			setupMocks: func() {
				hashedPassword, _ := utils.HashPassword(password)
				user := &models.User{ID: 1, Email: email, Password: hashedPassword}
				s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
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
				s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
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
				s.repo.AssertNotCalled(t, "ResetFailedLogins", mock.Anything, mock.Anything)
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

				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress, userID).Return(mockRes, nil)
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
				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress, userID).Return(nil, apperror.NewUnauthorizedError("Invalid refresh token"))
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

				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress, userID).Return(mockRes, nil)
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

				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress, userID).Return(mockRes, nil)
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
				// The refresh token belongs to userID but the access token to
				// another user: the rotation must be refused, not done then rejected.
				accessUserID := uint(2)
				claims := &services.CustomClaims{ID: accessUserID, Scope: services.TokenScopeAccess}

				s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress, accessUserID).Return(nil, apperror.NewUnauthorizedError("Invalid refresh token"))
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

func (s *AuthServiceTestSuite) TestRefreshToken_EmptyAccessToken() {
	oldRefreshToken := "old-refresh-token"
	oldAccessToken := ""
	ipAddress := "127.0.0.1"

	// Edge: an empty access token cannot be validated, so refresh must be
	// rejected before the refresh token is touched (no rotation).
	s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(nil, errors.New("Invalid token signature"))

	result, err := s.service.RefreshToken(context.Background(), oldRefreshToken, oldAccessToken, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	if appErr, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrUnauthorized, appErr.Code)
	}
	s.refreshTokenService.AssertNotCalled(s.T(), "Update", mock.Anything, oldRefreshToken, ipAddress, mock.Anything)
}

func (s *AuthServiceTestSuite) TestRefreshToken_EmptyRefreshToken() {
	oldRefreshToken := ""
	oldAccessToken := "old-access-token"
	ipAddress := "127.0.0.1"
	userID := uint(1)

	// Edge: an empty refresh token still reaches Update (after the access token
	// is validated first), which must fail without rotating a new access token.
	claims := &services.CustomClaims{ID: userID, Scope: services.TokenScopeAccess}
	s.jwtService.On("ValidateTokenIgnoreExpiration", oldAccessToken).Return(claims, nil)
	s.refreshTokenService.On("Update", mock.Anything, oldRefreshToken, ipAddress, userID).Return(nil, apperror.NewUnauthorizedError("Invalid refresh token"))

	result, err := s.service.RefreshToken(context.Background(), oldRefreshToken, oldAccessToken, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	if appErr, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrUnauthorized, appErr.Code)
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
	s.refreshTokenService.AssertNotCalled(s.T(), "Update", mock.Anything, oldRefreshToken, ipAddress, mock.Anything)
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
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)

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
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	s.repo.On("ResetFailedLogins", mock.Anything, user.ID).Return(nil).Once()
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
	s.repo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogin_ExpiredLockResetOnSuccess() {
	email := "expiredlock@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	hashedPassword, _ := utils.HashPassword(password)
	expiredLock := time.Now().Add(-30 * time.Minute).Unix()
	user := &models.User{ID: 1, Email: email, Password: hashedPassword, FailedAttempts: 3, LockedUntil: &expiredLock}
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	s.repo.On("ResetFailedLogins", mock.Anything, user.ID).Return(nil).Once()
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
	s.repo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogin_InvalidPasswordUpdateError() {
	email := "updatefail@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	user := &models.User{ID: 1, Email: email, Password: "wrong-hashed-password"}
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	s.expectFailedLogin(user.ID, errors.New("update error"))

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)
}

func (s *AuthServiceTestSuite) TestLogin_LockoutAfterMaxFailedAttempts() {
	email := "lockout@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	user := &models.User{ID: 1, Email: email, Password: "wrong-hashed", FailedAttempts: 4}
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	s.expectFailedLogin(user.ID, nil)

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)
	s.repo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogin_LockedUntilExactlyNow() {
	email := "boundary@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	// Boundary: lock expiry exactly at "now". The check is
	// `time.Now().Unix() < *user.LockedUntil`, so equality must be treated as
	// UNLOCKED and login proceeds normally.
	// NOTE: timestamps are second-resolution, so by the time Login runs the lock
	// is effectively "just expired"; the assertion still guards the equality
	// boundary (now == lockedUntil must not be locked).
	hashedPassword, _ := utils.HashPassword(password)
	lockedUntil := time.Now().Unix()
	user := &models.User{ID: 1, Email: email, Password: hashedPassword, FailedAttempts: 0, LockedUntil: &lockedUntil}
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	s.repo.On("ResetFailedLogins", mock.Anything, user.ID).Return(nil).Once()
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

func (s *AuthServiceTestSuite) TestLogin_EmptyEmail() {
	email := ""
	password := "password123"
	ipAddress := "127.0.0.1"

	// Edge: an empty email maps to no user, so login must fail with the same
	// generic invalid-credentials error as a wrong email (no user enumeration).
	s.repo.On("FindByEmail", mock.Anything, email).Return((*models.User)(nil), gorm.ErrRecordNotFound)

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)
	if appErr, ok := err.(*apperror.AppError); ok {
		assert.Equal(s.T(), apperror.ErrInvalidPassword, appErr.Code)
	}
}

func (s *AuthServiceTestSuite) TestLogin_ValidLoginAtMaxFailedAttemptsResets() {
	email := "maxreset@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	// Boundary: the user already sits exactly at MaxFailedAttempts but supplies
	// a correct password. FailedAttempts must reset to 0 and LockedUntil cleared.
	hashedPassword, _ := utils.HashPassword(password)
	user := &models.User{ID: 1, Email: email, Password: hashedPassword, FailedAttempts: services.MaxFailedAttempts}
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	s.repo.On("ResetFailedLogins", mock.Anything, user.ID).Return(nil).Once()
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
	s.repo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogin_FailedAttemptsAtMaxRelocks() {
	email := "maxrelock@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	// Boundary: the user is already at MaxFailedAttempts and fails again. The
	// account must stay locked (FailedAttempts stays >= max, LockedUntil set).
	user := &models.User{ID: 1, Email: email, Password: "wrong-hashed", FailedAttempts: services.MaxFailedAttempts}
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	s.expectFailedLogin(user.ID, nil)

	resp, err := s.service.Login(context.Background(), email, password, ipAddress)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), resp)
	s.repo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogin_LockedUntilOnlyResetOnSuccess() {
	email := "lockonly@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	// The reset block triggers on `FailedAttempts > 0 || LockedUntil != nil`.
	// Here only LockedUntil is set (FailedAttempts == 0), exercising the
	// lock-only branch: a valid login must still clear the stale lock.
	hashedPassword, _ := utils.HashPassword(password)
	expiredLock := time.Now().Add(-10 * time.Minute).Unix()
	user := &models.User{ID: 1, Email: email, Password: hashedPassword, FailedAttempts: 0, LockedUntil: &expiredLock}
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	s.repo.On("ResetFailedLogins", mock.Anything, user.ID).Return(nil).Once()
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
	s.repo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogin_ResetFailedAttemptsUpdateError() {
	email := "resetfail@example.com"
	password := "password123"
	ipAddress := "127.0.0.1"

	hashedPassword, _ := utils.HashPassword(password)
	user := &models.User{ID: 1, Email: email, Password: hashedPassword, FailedAttempts: 2}
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	s.repo.On("ResetFailedLogins", mock.Anything, user.ID).Return(errors.New("update error")).Once()
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

func (s *AuthServiceTestSuite) TestLogin_ExpiredLockRestartsTheCount() {
	email := "expired-relock@example.com"
	ipAddress := "127.0.0.1"

	// The lock has expired but the counter is still at the maximum. One wrong
	// password must count as the first attempt again, not lock the account.
	expiredLock := time.Now().Add(-time.Minute).Unix()
	user := &models.User{ID: 1, Email: email, Password: "wrong-hashed", FailedAttempts: services.MaxFailedAttempts, LockedUntil: &expiredLock}
	s.repo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	var calls []string
	s.repo.On("ResetFailedLogins", mock.Anything, user.ID).Return(nil).Once().Run(func(mock.Arguments) { calls = append(calls, "reset") })
	expectedLock := time.Now().Add(services.LockoutDurationMinutes * time.Minute).Unix()
	s.repo.On("RecordFailedLogin", mock.Anything, user.ID, services.MaxFailedAttempts, mock.MatchedBy(func(lockUntil int64) bool {
		return lockUntil >= expectedLock-5 && lockUntil <= expectedLock+5
	})).Return(nil).Once().Run(func(mock.Arguments) { calls = append(calls, "record") })

	resp, err := s.service.Login(context.Background(), email, "password123", ipAddress)

	assert.Nil(s.T(), resp)
	if appErr, ok := err.(*apperror.AppError); s.True(ok) {
		assert.Equal(s.T(), apperror.ErrInvalidPassword, appErr.Code, "a wrong password, not a lock")
	}
	assert.Equal(s.T(), []string{"reset", "record"}, calls)
	s.repo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogin_UnknownEmailStillChecksAPassword() {
	// Skipping bcrypt for an unknown email makes that answer much faster, which
	// tells registered emails apart.
	var checked []string
	original := utils.CheckPasswordHash
	utils.CheckPasswordHash = func(password, hash string) bool {
		checked = append(checked, password)
		return original(password, hash)
	}
	defer func() { utils.CheckPasswordHash = original }()
	s.repo.On("FindByEmail", mock.Anything, "nobody@example.com").Return((*models.User)(nil), gorm.ErrRecordNotFound)

	resp, err := s.service.Login(context.Background(), "nobody@example.com", "password123", "127.0.0.1")

	assert.Nil(s.T(), resp)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), []string{"password123"}, checked)
}

// --------------------- RUN TEST SUITE ---------------------
func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
