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

type UserHandler interface {
	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	ChangePassword(c *gin.Context)
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
}

type userHandlerImpl struct {
	userService   services.UserService
	mailerService services.MailerService
}

func NewUserHandler(
	userService services.UserService,
	mailerService services.MailerService,
) UserHandler {
	return &userHandlerImpl{
		userService:   userService,
		mailerService: mailerService,
	}
}

func (handler *userHandlerImpl) ForgotPassword(ctx *gin.Context) {
	var input dto.ForgotPasswordInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	err := handler.userService.ForgotPassword(ctx.Request.Context(), &input)
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Forgot password failed for email %s: %v", input.Email, err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "If your email is in our system, you will receive instructions to reset your password"})
}

func (handler *userHandlerImpl) ResetPassword(ctx *gin.Context) {
	var input dto.ResetPasswordInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	_, err := handler.userService.ResetPassword(ctx.Request.Context(), &input)
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Reset password failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Reset password successfully"})
}

func (handler *userHandlerImpl) ChangePassword(ctx *gin.Context) {
	userId, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, apperror.NewParseError("Invalid UserID"))
		return
	}

	var input dto.ChangePasswordInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	_, err = handler.userService.ChangePassword(ctx.Request.Context(), userId, &input)
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Change password failed for user %d: %v", userId, err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Change password successfully"})
}

func (handler *userHandlerImpl) GetProfile(ctx *gin.Context) {
	userId, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, apperror.NewParseError("Invalid UserID"))
		return
	}

	dbUser, err := handler.userService.GetProfile(ctx.Request.Context(), userId)
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Get profile failed for user %d: %v", userId, err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, dbUser)
}

func (handler *userHandlerImpl) UpdateProfile(ctx *gin.Context) {
	userId, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, apperror.NewParseError("Invalid UserID"))
		return
	}

	var input dto.UpdateProfileInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	err = handler.userService.UpdateProfile(ctx.Request.Context(), userId, &input)
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Update profile failed for user %d: %v", userId, err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Update profile successfully"})
}
