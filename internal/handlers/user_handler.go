package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

type IUserhandler interface {
	CreateUser(c *gin.Context)
	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	Login(c *gin.Context)
	GetUser(c *gin.Context)
	GetUsers(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
}

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (handler *UserHandler) CreateUser(c *gin.Context) {

	var input struct {
		Email    string  `json:"email" binding:"required,email"`
		Password string  `json:"password" binding:"required,min=6,max=255"`
		Name     string  `json:"name" binding:"required,min=1,max=45"`
		Birthday *string `json:"birthday" binding:"required,datetime=2006-01-02"` // Assumes YYYY-MM-DD format
		Address  *string `json:"address" binding:"required,min=1,max=255"`
		Gender   int16   `json:"gender" binding:"required,oneof=0 1 2"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
		Birthday: input.Birthday,
		Address:  input.Address,
		Gender:   input.Gender,
	}

	user.Password = utils.HashPassword(user.Password)

	if err := handler.userService.CreateUser(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Create user successfully"})

}

func (handle *UserHandler) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	// Bind and validate JSON request body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by email from database
	user, err := handle.userService.GetUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send password reset email to user
	if err := services.SendMailForgotPassword(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("Email sent successfully!")

	c.JSON(http.StatusOK, gin.H{"message": "Forgot password successfully"})
}

func (handler *UserHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		Password    string `json:"password" binding:"required,min=6,max=255"`
		NewPassword string `json:"new_password" binding:"required,min=6,max=255"`
	}
	// Bind and validate JSON request body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by token from database
	user, err := handler.userService.GetUserByToken(input.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if token is expired
	if time.Now().Unix() > *user.ExpiredAt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token expired"})
		return
	}

	// Check if new password is the same as old password
	if isValid := utils.CheckPasswordHash(input.Password, user.Password); !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be different from old password"})
		return
	}

	// Update user password
	user.Password = utils.HashPassword(input.NewPassword)
	user.Token = nil
	user.ExpiredAt = nil

	// Update user in database
	if err := handler.userService.UpdateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reset password successfully"})
}

func (handler *UserHandler) ChangePassword(c *gin.Context) {
	// Get user ID from the context
	// If user ID is 0 or not found, return bad request error
	userId := c.GetUint("UserID")
	if userId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UserID"})
		return
	}

	var input struct {
		OldPassword     string `json:"old_password" binding:"required,min=6,max=255"`
		NewPassword     string `json:"new_password" binding:"required,min=6,max=255"`
		ConfirmPassword string `json:"confirm_password" binding:"required,min=6,max=255"`
	}
	// Bind and validate JSON request body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by ID from database
	user, err := handler.userService.GetUser(uint(userId))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if old password is correct
	if isValid := utils.CheckPasswordHash(input.OldPassword, user.Password); !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Old password is incorrect"})
		return
	}

	// Check if new password is the same as old password
	if input.OldPassword == input.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be different from old password"})
		return
	}

	// Check if new password and confirm password match
	if input.NewPassword != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password and confirm password do not match"})
		return
	}

	// Update user password
	user.Password = utils.HashPassword(input.NewPassword)

	// Update user in database
	if err := handler.userService.UpdateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Change password successfully"})
}

func (handler *UserHandler) DeleteUser(c *gin.Context) {
	// Get user ID from the context
	id := c.Param("id")
	userId, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delete user from database
	if err := handler.userService.DeleteUser(uint(userId)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delete user successfully"})
}

func (handler *UserHandler) UpdateUser(c *gin.Context) {
	// Get user ID from the context
	id := c.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Define input struct with validation tags
	var input struct {
		Name     string `json:"name" validate:"min=1,max=45"`           // Name must be between 1-45 chars
		Birthday string `json:"birthday" validate:"valid_birthday"`     // Birthday must be valid date
		Address  string `json:"address" validate:"min=1,max=255"`       // Address must be between 1-255 chars
		Gender   int16  `json:"gender" validate:"required,oneof=0 1 2"` // Gender must be 0, 1 or 2
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing user from database
	user, err := handler.userService.GetUser(uint(userId))

	// Return error if user not found
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user fields with input values
	user.Name = input.Name
	user.Birthday = &input.Birthday
	user.Address = &input.Address
	user.Gender = input.Gender

	// Save updated user to database
	if err := handler.userService.UpdateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Update user successfully"})
}

func (handler *UserHandler) GetUser(c *gin.Context) {
	// Get user ID from the context
	id := c.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from database
	user, err := handler.userService.GetUser(uint(userId))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (handler *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from the context
	userId := c.GetUint("UserID")
	if userId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UserID"})
		return
	}

	// Get user from database
	user, err := handler.userService.GetUser(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (handler *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context and validate
	userId := c.GetUint("UserID")
	if userId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UserID"})
		return
	}

	// Define input struct for profile update with validation rules
	var input struct {
		Name     *string `json:"name" binding:"omitempty,min=1,max=45"`            // Name must be between 1 and 45 characters if provided
		Birthday *string `json:"birthday" binding:"omitempty,datetime=2006-01-02"` // Birthday must be a valid date (YYYY-MM-DD) if provided
		Address  *string `json:"address" binding:"omitempty,min=1,max=255"`        // Address must be between 1 and 255 characters if provided
		Gender   *int16  `json:"gender" binding:"omitempty,oneof=0 1 2"`           // Gender must be 0, 1, or 2 if provided
	}

	// Bind and validate JSON request body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing user from database
	user, err := handler.userService.GetUser(userId)

	// Return error if user not found
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user fields if provided in input
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Birthday != nil {
		user.Birthday = input.Birthday
	}
	if input.Address != nil {
		user.Address = input.Address
	}
	if input.Gender != nil {
		user.Gender = *input.Gender
	}

	// Save updated user to database
	if err := handler.userService.UpdateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Update profile successfully"})

}
