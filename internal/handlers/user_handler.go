package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type UserHandler interface {
	// User Management
	CreateUser(c *gin.Context)
	GetUser(c *gin.Context)
	GetUsers(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)

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
	bcryptService services.BcryptService
	mallerService services.MailerService
}

func NewUserHandler(
	userService services.UserService,
	bcryptService services.BcryptService,
	mailerService services.MailerService,
) UserHandler {
	return &userHandlerImpl{
		userService:   userService,
		bcryptService: bcryptService,
		mallerService: mailerService,
	}
}

func (handler *userHandlerImpl) GetUsers(ctx *gin.Context) {
	// Parse pagination query parameters with default values
	page, limit := utils.ParsePageAndLimit(ctx)

	// Retrieve paginated list of users from the service using PaginateUser
	pagination, err := handler.userService.GetUsers(page, limit)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	// Respond with pagination data
	utils.RespondWithOK(ctx, http.StatusOK, pagination)
}

func (handler *userHandlerImpl) CreateUser(ctx *gin.Context) {

	var input dto.CreateUserInput

	// Bind and validate the JSON request body to the input struct
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	hashpassword, err := handler.bcryptService.HashPassword(input.Password)
	if err != nil {
		utils.RespondWithError(
			ctx,
			apperror.NewPasswordHashFailedError("Failed to hash password"))
		return
	}

	// Create a new User model instance with the validated input data
	// Password is stored as the hashed value
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashpassword,
		Address:  input.Address,
		Gender:   input.Gender,
	}
	if input.Birthday != nil {
		birthdayDate, err := utils.ParseDateStringYYYYMMDD(*input.Birthday)
		if err != nil {
			utils.RespondWithError(
				ctx,
				apperror.NewParseError("Invalid birthday format, expected YYYY-MM-DD"),
			)
			return
		}
		user.Birthday = birthdayDate
	}

	// Attempt to create the user in the database
	// Return 400 Bad Request if creation fails
	if err := handler.userService.CreateUser(&user); err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusCreated, gin.H{"message": "Create user successfully"})
}

func (handle *userHandlerImpl) ForgotPassword(ctx *gin.Context) {
	var input dto.ForgotPasswordInput
	// Bind and validate JSON request body
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	// Get user by email from database
	user, err := handle.userService.GetUserByEmail(input.Email)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	// Generate random token string for password reset
	newToken := utils.GenerateRandomString(60)

	expiredAt := time.Now().Add(time.Hour).Unix()

	// Set new token on user
	user.Token = &newToken
	user.ExpiredAt = &expiredAt

	// Update user in database with new token
	if err := handle.userService.UpdateUser(user); err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	// Send password reset email to user
	if err := handle.mallerService.SendMailForgotPassword(user); err != nil {
		utils.RespondWithError(ctx, err)
		return
	}
	logger.Info("Email sent successfully!")

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Forgot password successfully"})
}

func (handler *userHandlerImpl) ResetPassword(ctx *gin.Context) {
	var input dto.ResetPasswordInput
	// Bind and validate JSON request body
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	// Get user by token from database
	user, err := handler.userService.GetUserByToken(input.Token)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	// Check if token is expired
	if time.Now().Unix() > *user.ExpiredAt {
		utils.RespondWithError(ctx, apperror.NewTokenExpiredError("Token is expired"))
		return
	}

	// Hash the password using the utils.HashPassword function
	// If hashing fails (returns empty string), return a 400 error
	hashpassword, err := handler.bcryptService.HashPassword(input.NewPassword)
	if err != nil {
		utils.RespondWithError(ctx, apperror.NewPasswordHashFailedError("Failed to hash password"))
		return
	}

	// Update user password
	user.Password = hashpassword
	user.Token = nil
	user.ExpiredAt = nil

	// Update user in database
	if err := handler.userService.UpdateUser(user); err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Reset password successfully"})
}

func (handler *userHandlerImpl) ChangePassword(ctx *gin.Context) {
	// Get user ID from the context
	// If user ID is 0 or not found, return bad request error
	userId := ctx.GetUint("UserID")
	if userId == 0 {
		utils.RespondWithError(
			ctx,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	var input dto.ChangePasswordInput
	// Bind and validate JSON request body
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	// Get user by ID from database
	user, err := handler.userService.GetUser(uint(userId))
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	// Check if old password is correct
	if isValid := handler.bcryptService.CheckPasswordHash(input.OldPassword, user.Password); !isValid {
		utils.RespondWithError(
			ctx,
			apperror.NewInvalidPasswordError("Old password is incorrect"),
		)
		return
	}

	// Check if new password is the same as old password
	if input.OldPassword == input.NewPassword {
		utils.RespondWithError(
			ctx,
			apperror.NewPasswordMismatchError("New password must be different from old password"),
		)
		return
	}

	// Check if new password and confirm password match
	if input.NewPassword != input.ConfirmPassword {
		utils.RespondWithError(
			ctx,
			apperror.NewPasswordMismatchError("New password and confirm password do not match"),
		)
		return
	}

	// Hash the password using the utils.HashPassword function
	// If hashing fails (returns empty string), return a 500 error
	hashpassword, err := handler.bcryptService.HashPassword(input.NewPassword)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	// Update user password
	user.Password = hashpassword

	// Update user in database
	if err := handler.userService.UpdateUser(user); err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Change password successfully"})
}

func (handler *userHandlerImpl) DeleteUser(ctx *gin.Context) {
	// Get user ID from the context
	id := ctx.Param("id")
	userId, err := strconv.Atoi(id)

	if err != nil {
		utils.RespondWithError(
			ctx,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	// Get user from database
	user, userErr := handler.userService.GetUser(uint(userId))
	if userErr != nil {
		utils.RespondWithError(ctx, userErr)
		return
	}

	// Delete user from database
	if err := handler.userService.DeleteUser(uint(user.ID)); err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Delete user successfully"})
}

func (handler *userHandlerImpl) UpdateUser(ctx *gin.Context) {
	// Get user ID from the context
	id := ctx.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		utils.RespondWithError(
			ctx,
			apperror.NewParseError("Invalid UserID"),
		)

		return
	}

	// Define input struct with validation tags
	var input dto.UpdateUserInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	// Get existing user from database
	user, userErr := handler.userService.GetUser(uint(userId))
	// Return error if user not found
	if userErr != nil {
		utils.RespondWithError(ctx, userErr)
		return
	}

	// Update user fields with input values
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Birthday != nil {
		birthdayDate, err := utils.ParseDateStringYYYYMMDD(*input.Birthday)
		if err != nil {
			utils.RespondWithError(
				ctx,
				apperror.NewParseError("Invalid birthday format, expected YYYY-MM-DD"),
			)
			return
		}
		user.Birthday = birthdayDate
	}
	if input.Address != nil {
		user.Address = input.Address
	}
	if input.Gender != nil {
		user.Gender = *input.Gender
	}

	// Save updated user to database
	if err := handler.userService.UpdateUser(user); err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Update user successfully"})
}

func (handler *userHandlerImpl) GetUser(ctx *gin.Context) {
	// Get user ID from the context
	id := ctx.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		utils.RespondWithError(
			ctx,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	// Get user from database
	user, userErr := handler.userService.GetUser(uint(userId))
	if userErr != nil {
		utils.RespondWithError(ctx, userErr)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, user)
}

func (handler *userHandlerImpl) GetProfile(ctx *gin.Context) {
	// Get user ID from context and validate
	userId := ctx.GetUint("UserID")
	if userId == 0 {
		utils.RespondWithError(
			ctx,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	// Get user from database
	logger.Info("User retrieved from DB")
	dbUser, err := handler.userService.GetProfile(userId)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, dbUser)
}

func (handler *userHandlerImpl) UpdateProfile(ctx *gin.Context) {
	// Get user ID from context and validate
	userId := ctx.GetUint("UserID")
	if userId == 0 {
		utils.RespondWithError(
			ctx,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	// Define input struct for profile update with validation rules
	var input dto.UpdateProfileInput

	// Bind and validate JSON request body
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	// Get existing user from database
	user, err := handler.userService.GetUser(userId)

	// Return error if user not found
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	// Update user fields if provided in input
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Birthday != nil {
		birthdayDate, err := utils.ParseDateStringYYYYMMDD(*input.Birthday)
		if err != nil {
			utils.RespondWithError(
				ctx,
				apperror.NewParseError("Invalid birthday format, expected YYYY-MM-DD"),
			)
			return
		}
		user.Birthday = birthdayDate
	}
	if input.Address != nil {
		user.Address = input.Address
	}
	if input.Gender != nil {
		user.Gender = *input.Gender
	}

	// Save updated user to database
	if err := handler.userService.UpdateUser(user); err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Update profile successfully"})
}
