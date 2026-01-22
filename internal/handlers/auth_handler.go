package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

type AuthHandler interface {
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
}

type authHandlerImpl struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) AuthHandler {
	return &authHandlerImpl{
		authService: authService,
	}
}

func (handler *authHandlerImpl) Login(ctx *gin.Context) {
	// Bind and validate JSON request body
	var credentials dto.LoginInput
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		validateErr := utils.TranslateValidationErrors(err, credentials)
		utils.RespondWithError(
			ctx,
			validateErr,
		)
		return
	}

	// login handler
	res, err := handler.authService.Login(credentials.Email, credentials.Password, ctx)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, res)
}

func (handler *authHandlerImpl) RefreshToken(ctx *gin.Context) {
	// Bind and validate JSON request body
	var input dto.RefreshTokenInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validationErr := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(
			ctx,
			validationErr,
		)
		return
	}

	// Call auth service to refresh the token
	res, err := handler.authService.RefreshToken(input.RefreshToken, input.AccessToken, ctx)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, res)
}
