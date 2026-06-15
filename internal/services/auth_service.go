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

	user, err := service.repo.FindByField(ctx, "email", email)
	if err != nil {
		logger.WithEvent(ctx, logger.EventLoginFailed).Warnf("Login failed - user not found: %s", utils.MaskWithPrefix(email, 4))
		return nil, apperror.NewInvalidPasswordError("Invalid credentials")
	}

	if isValid := utils.CheckPasswordHash(password, user.Password); !isValid {
		logger.WithEvent(ctx, logger.EventLoginFailed).Warnf("Login failed - invalid password for email: %s", utils.MaskWithPrefix(email, 4))
		return nil, apperror.NewInvalidPasswordError("Invalid credentials")
	}

	accessToken, err := service.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		logger.WithEvent(ctx, logger.EventLoginFailed).Errorf("Failed to generate access token for user ID %d: %v", user.ID, err)
		return nil, apperror.NewInternalServerError("Failed to generate access token")
	}

	refreshToken, errToken := service.refreshTokenService.Create(ctx, user, ipAddress)

	if errToken != nil {
		logger.WithEvent(ctx, logger.EventLoginFailed).Errorf("Failed to create refresh token for user ID %d: %v", user.ID, errToken)
		return nil, errToken
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

	refreshResult, err := service.refreshTokenService.Update(ctx, refreshToken, ipAddress)
	if err != nil {
		logger.WithEvent(ctx, logger.EventTokenRefreshFailed).Warnf("Token refresh failed - invalid refresh token")
		return nil, apperror.NewUnauthorizedError("Invalid refresh token")
	}

	claims, err := service.jwtService.ValidateTokenIgnoreExpiration(accessToken)
	if err != nil {
		logger.WithEvent(ctx, logger.EventTokenRefreshFailed).Warnf("Token refresh failed - invalid access token")
		return nil, apperror.NewUnauthorizedError("Invalid access token")
	}

	if claims.Scope != TokenScopeAccess {
		logger.WithEvent(ctx, logger.EventTokenRefreshFailed).Warnf("Token refresh failed - invalid scope")
		return nil, apperror.NewUnauthorizedError("Invalid access token scope")
	}

	if claims.ID != refreshResult.UserId {
		logger.WithEvent(ctx, logger.EventTokenRefreshFailed).Warnf("Token refresh failed - token mismatch")
		return nil, apperror.NewUnauthorizedError("Token mismatch: refresh and access tokens belong to different users")
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
