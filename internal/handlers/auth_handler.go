package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type authHandlerImpl struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *authHandlerImpl {
	return &authHandlerImpl{
		authService: authService,
	}
}

func (handler *authHandlerImpl) Login(ctx *gin.Context) {
	var credentials dto.LoginInput
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		validateErr := utils.TranslateValidationErrors(err, credentials)
		utils.RespondWithError(ctx, validateErr)
		return
	}

	res, err := handler.authService.Login(ctx.Request.Context(), credentials.Email, credentials.Password, ctx.ClientIP())
	if err != nil {
		logger.WithEvent(ctx.Request.Context(), logger.EventLoginFailed).Errorf("Login failed for email %s: %v", utils.MaskWithPrefix(credentials.Email, 4), err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, res)
}

func (handler *authHandlerImpl) Logout(ctx *gin.Context) {
	userID, exists := ctx.Get("UserID")
	if !exists {
		utils.RespondWithError(ctx, apperror.NewUnauthorizedError("Unauthorized"))
		return
	}

	if err := handler.authService.Logout(ctx.Request.Context(), userID.(uint)); err != nil {
		logger.WithEvent(ctx.Request.Context(), logger.EventLogout).Errorf("Logout failed for user ID %d: %v", userID, err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (handler *authHandlerImpl) RefreshToken(ctx *gin.Context) {
	var input dto.RefreshTokenInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validationErr := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validationErr)
		return
	}

	res, err := handler.authService.RefreshToken(ctx.Request.Context(), input.RefreshToken, input.AccessToken, ctx.ClientIP())
	if err != nil {
		logger.WithEvent(ctx.Request.Context(), logger.EventTokenRefreshFailed).Errorf("Token refresh failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, res)
}
