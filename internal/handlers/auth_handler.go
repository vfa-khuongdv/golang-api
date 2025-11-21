package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

type IAuthHandler interface {
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
}

type AuthHandler struct {
	authService services.IAuthService
}

func NewAuthHandler(authService services.IAuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (handler *AuthHandler) Login(ctx *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6,max=255"`
	}

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

func (handler *AuthHandler) RefreshToken(ctx *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
		AccessToken  string `json:"access_token" binding:"required"`
	}

	// Bind JSON request body to token struct
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
