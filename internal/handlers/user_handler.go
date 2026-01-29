package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type UserHandler interface {
	// Authentication and Password Management
	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	ChangePassword(c *gin.Context)

	// Profile Management for the authenticated user
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
	requestID := middlewares.GetRequestID(ctx)

	// Bind and validate JSON request body
	var input dto.ForgotPasswordInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	logger.InfofWithRequestID(requestID, "Processing forgot password request for email: %s", input.Email)

	// Handle forgot password logic
	user, err := handler.userService.ForgotPassword(&input)

	if err != nil {
		logger.ErrorfWithRequestID(requestID, "Forgot password failed for email %s: %v", input.Email, err)
		utils.RespondWithError(ctx, err)
		return
	}

	// Send password reset email to user
	if user != nil {
		if err := handler.mailerService.SendMailForgotPassword(user); err != nil {
			logger.ErrorfWithRequestID(requestID, "Failed to send password reset email to %s: %v", user.Email, err)
			utils.RespondWithError(ctx, err)
			return
		}
		logger.InfofWithRequestID(requestID, "Password reset email sent successfully to %s", user.Email)
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "If your email is in our system, you will receive instructions to reset your password"})
}

func (handler *userHandlerImpl) ResetPassword(ctx *gin.Context) {
	// Bind and validate JSON request body
	var input dto.ResetPasswordInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	// Handle reset password logic
	_, err := handler.userService.ResetPassword(&input)
	if err != nil {
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
	// Bind and validate JSON request body
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	// Handle change password logic
	_, err = handler.userService.ChangePassword(userId, &input)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Change password successfully"})
}

func (handler *userHandlerImpl) GetProfile(ctx *gin.Context) {
	// Get user ID from context and validate
	userId, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.RespondWithError(
			ctx,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	// Handle get profile logic
	dbUser, err := handler.userService.GetProfile(userId)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, dbUser)
}

func (handler *userHandlerImpl) UpdateProfile(ctx *gin.Context) {
	// Get user ID from context and validate
	userId, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.RespondWithError(
			ctx,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	// Bind and validate JSON request body
	var input dto.UpdateProfileInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	// Handle update profile logic
	err = handler.userService.UpdateProfile(userId, &input)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Update profile successfully"})
}
