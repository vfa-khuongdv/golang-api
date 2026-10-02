package services

import (
	"context"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type AuthService interface {
	Login(ctx context.Context, email, password string, ipAddress string) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken, accessToken string, ipAddress string) (*dto.LoginResponse, error)
	Logout(ctx context.Context, userID uint) error
}

type authServiceImpl struct {
	repo                repositories.UserRepository
	refreshTokenService RefreshTokenService
	jwtService          JWTService
}

func NewAuthService(repo repositories.UserRepository, refreshTokenService RefreshTokenService, jwtService JWTService) AuthService {
	return &authServiceImpl{
		repo:                repo,
		refreshTokenService: refreshTokenService,
		jwtService:          jwtService,
	}
}

func (service *authServiceImpl) Login(ctx context.Context, email, password string, ipAddress string) (*dto.LoginResponse, error) {
	start := time.Now()
	logger.WithEvent(ctx, logger.EventLoginAttempt).Infof("Login attempt for email: %s", utils.MaskWithPrefix(email, 4))

	user, err := service.repo.FindByEmail(ctx, email)
	if err != nil {
		logger.WithEvent(ctx, logger.EventLoginFailed).Warnf("Login failed - user not found: %s", utils.MaskWithPrefix(email, 4))
		return nil, apperror.NewInvalidPasswordError("Invalid credentials")
	}

	if user.LockedUntil != nil && time.Now().Unix() < *user.LockedUntil {
		logger.WithEvent(ctx, logger.EventLoginFailed).Warnf("Login failed - account locked for email: %s", utils.MaskWithPrefix(email, 4))
		return nil, apperror.NewAccountLockedError("Account is temporarily locked due to too many failed attempts. Try again later.")
	}

	// An expired lock starts a fresh count, otherwise one wrong password right
	// after the lock ends would lock the account again.
	if user.LockedUntil != nil {
		if updateErr := service.repo.ResetFailedLogins(ctx, user.ID); updateErr != nil {
			logger.WithEvent(ctx, logger.EventLoginFailed).Errorf("Failed to reset expired lock for user ID %d: %v", user.ID, updateErr)
		}
		user.FailedAttempts = 0
		user.LockedUntil = nil
	}

	// The counters are updated with targeted, atomic queries rather than saving the
	// whole user: concurrent failures must all be counted, and a stale copy of the
	// user must never overwrite a password changed while bcrypt was running.
	if isValid := utils.CheckPasswordHash(password, user.Password); !isValid {
		lockUntil := time.Now().Add(time.Minute * LockoutDurationMinutes).Unix()
		if updateErr := service.repo.RecordFailedLogin(ctx, user.ID, MaxFailedAttempts, lockUntil); updateErr != nil {
			logger.WithEvent(ctx, logger.EventLoginFailed).Errorf("Failed to update user after failed login: %v", updateErr)
		}
		logger.WithEvent(ctx, logger.EventLoginFailed).Warnf("Login failed - invalid password for email: %s (attempt %d/%d)", utils.MaskWithPrefix(email, 4), user.FailedAttempts+1, MaxFailedAttempts)
		return nil, apperror.NewInvalidPasswordError("Invalid credentials")
	}

	if user.FailedAttempts > 0 || user.LockedUntil != nil {
		if updateErr := service.repo.ResetFailedLogins(ctx, user.ID); updateErr != nil {
			logger.WithEvent(ctx, logger.EventLoginFailed).Errorf("Failed to reset failed attempts for user ID %d: %v", user.ID, updateErr)
		}
	}

	accessToken, err := service.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		logger.WithEvent(ctx, logger.EventLoginFailed).Errorf("Failed to generate access token for user ID %d: %v", user.ID, err)
		return nil, apperror.NewInternalServerError("Failed to generate access token")
	}

	refreshToken, err := service.refreshTokenService.Create(ctx, user, ipAddress)

	if err != nil {
		logger.WithEvent(ctx, logger.EventLoginFailed).Errorf("Failed to create refresh token for user ID %d: %v", user.ID, err)
		return nil, err
	}

	logger.WithEvent(ctx, logger.EventLoginSuccess).
		WithField("latency_ms", time.Since(start).Milliseconds()).
		Infof("Login successful for user ID %d", user.ID)

	return &dto.LoginResponse{
		AccessToken: dto.JwtResult{
			Token:     accessToken.Token,
			ExpiresAt: accessToken.ExpiresAt,
		},
		RefreshToken: dto.JwtResult{
			Token:     refreshToken.Token,
			ExpiresAt: refreshToken.ExpiresAt,
		},
	}, nil
}

func (service *authServiceImpl) RefreshToken(ctx context.Context, refreshToken, accessToken string, ipAddress string) (*dto.LoginResponse, error) {
	start := time.Now()
	logger.WithEvent(ctx, logger.EventTokenRefresh).Infof("Token refresh attempt")

	// Validate the access token FIRST, before touching (rotating) the refresh
	// token. If an invalid access token were allowed to reach the rotation
	// step, an attacker who stole a refresh token (but not a valid access
	// token) could burn it, force-logging-out the legitimate user.
	claims, err := service.jwtService.ValidateTokenIgnoreExpiration(accessToken)
	if err != nil {
		logger.WithEvent(ctx, logger.EventTokenRefreshFailed).Warnf("Token refresh failed - invalid access token")
		return nil, apperror.NewUnauthorizedError("Invalid access token")
	}

	if claims.Scope != TokenScopeAccess {
		logger.WithEvent(ctx, logger.EventTokenRefreshFailed).Warnf("Token refresh failed - invalid scope")
		return nil, apperror.NewUnauthorizedError("Invalid access token scope")
	}

	// Update also refuses a refresh token that does not belong to claims.ID.
	refreshResult, err := service.refreshTokenService.Update(ctx, refreshToken, ipAddress, claims.ID)
	if err != nil {
		logger.WithEvent(ctx, logger.EventTokenRefreshFailed).Warnf("Token refresh failed - invalid refresh token: %v", err)
		return nil, apperror.NewUnauthorizedError("Invalid refresh token")
	}

	user, err := service.repo.GetByID(ctx, refreshResult.UserId)
	if err != nil {
		logger.WithEvent(ctx, logger.EventTokenRefreshFailed).Warnf("Token refresh failed - user not found: %d", refreshResult.UserId)
		return nil, apperror.NewNotFoundError("User not found")
	}

	newAccessToken, err := service.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		logger.WithEvent(ctx, logger.EventTokenRefreshFailed).Errorf("Failed to generate new access token for user ID %d: %v", user.ID, err)
		return nil, apperror.NewInternalServerError("Failed to generate access token")
	}

	logger.WithEvent(ctx, logger.EventTokenRefreshSuccess).
		WithField("latency_ms", time.Since(start).Milliseconds()).
		Infof("Token refresh successful for user ID %d", user.ID)

	return &dto.LoginResponse{
		AccessToken: dto.JwtResult{
			Token:     newAccessToken.Token,
			ExpiresAt: newAccessToken.ExpiresAt,
		},
		RefreshToken: dto.JwtResult{
			Token:     refreshResult.Token.Token,
			ExpiresAt: refreshResult.Token.ExpiresAt,
		},
	}, nil
}

func (service *authServiceImpl) Logout(ctx context.Context, userID uint) error {
	logger.WithEvent(ctx, logger.EventLogout).Infof("Logout for user ID %d", userID)
	return service.refreshTokenService.DeleteByUserID(ctx, userID)
}
