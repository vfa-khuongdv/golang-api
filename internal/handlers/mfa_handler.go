package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

// IMfaHandler defines operations for MFA endpoints
type IMfaHandler interface {
	InitMfaSetup(c *gin.Context)
	VerifyMfaSetup(c *gin.Context)
	VerifyMfaCode(c *gin.Context)
	DisableMfa(c *gin.Context)
	GetMfaStatus(c *gin.Context)
}

type MfaHandler struct {
	mfaService          services.IMfaService
	userRepository      repositories.IUserRepository
	jwtService          services.IJWTService
	refreshTokenService services.IRefreshTokenService
}

// NewMfaHandler creates a new instance of MfaHandler
func NewMfaHandler(mfaService services.IMfaService, userRepository repositories.IUserRepository, jwtService services.IJWTService, refreshTokenService services.IRefreshTokenService) IMfaHandler {
	return &MfaHandler{
		mfaService:          mfaService,
		userRepository:      userRepository,
		jwtService:          jwtService,
		refreshTokenService: refreshTokenService,
	}
}

// InitMfaSetup initiates the MFA setup process for the authenticated user
func (h *MfaHandler) InitMfaSetup(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("UserID")
	if userID == 0 {
		utils.RespondWithError(
			c,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	// Get user email from context or body
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(c, validateError)
		return
	}

	// Setup MFA - generates secret and QR code
	secret, qrCodeBytes, backupCodes, err := h.mfaService.SetupMfa(userID, input.Email)
	if err != nil {
		logrus.Errorf("Failed to setup MFA for user %d: %v", userID, err)
		utils.RespondWithError(c, err)
		return
	}

	// Encode QR code to base64 for embedding in response
	qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCodeBytes)

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"secret":       secret,
		"qr_code":      fmt.Sprintf("data:image/png;base64,%s", qrCodeBase64),
		"backup_codes": backupCodes,
	})
}

// VerifyMfaSetup verifies the TOTP code provided during MFA setup
func (h *MfaHandler) VerifyMfaSetup(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("UserID")
	if userID == 0 {
		utils.RespondWithError(
			c,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	var input struct {
		Code string `json:"code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(c, validateError)
		return
	}

	// Verify the TOTP code
	backupCodes, err := h.mfaService.VerifyMfaSetup(userID, input.Code)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"message":      "MFA setup verified successfully",
		"backup_codes": backupCodes,
	})
}

// VerifyMfaCode verifies a TOTP code during login and returns access/refresh tokens
func (h *MfaHandler) VerifyMfaCode(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("UserID")
	if userID == 0 {
		utils.RespondWithError(
			c,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	var input struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(c, validateError)
		return
	}

	// Verify the TOTP code
	valid, err := h.mfaService.VerifyMfaCode(userID, input.Code)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	if !valid {
		utils.RespondWithError(c, apperror.NewMfaInvalidCodeError("Invalid MFA code"))
		return
	}

	// MFA verification successful, now generate access and refresh tokens
	// Get user details
	user, err := h.userRepository.GetByID(userID)
	if err != nil {
		logrus.Errorf("Failed to get user details for ID %d: %v", userID, err)
		utils.RespondWithError(c, apperror.NewNotFoundError("User not found"))
		return
	}

	// Generate access token
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		logrus.Errorf("Failed to generate access token: %v", err)
		utils.RespondWithError(c, apperror.NewInternalError("Failed to generate tokens after MFA verification"))
		return
	}

	// Create new refresh token
	ipAddress := c.ClientIP()
	refreshToken, err := h.refreshTokenService.Create(user, ipAddress)
	if err != nil {
		logrus.Errorf("Failed to create refresh token: %v", err)
		utils.RespondWithError(c, apperror.NewInternalError("Failed to create refresh token"))
		return
	}

	result := &services.LoginResponse{
		AccessToken: services.JwtResult{
			Token:     accessToken.Token,
			ExpiresAt: accessToken.ExpiresAt,
		},
		RefreshToken: services.JwtResult{
			Token:     refreshToken.Token,
			ExpiresAt: refreshToken.ExpiresAt,
		},
	}

	// Return tokens to client
	utils.RespondWithOK(c, http.StatusOK, result)
}

// DisableMfa disables MFA for the authenticated user
func (h *MfaHandler) DisableMfa(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("UserID")
	if userID == 0 {
		utils.RespondWithError(
			c,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	// Disable MFA
	if err := h.mfaService.DisableMfa(userID); err != nil {
		logrus.Errorf("Failed to disable MFA for user %d: %v", userID, err)
		utils.RespondWithError(c, apperror.NewInternalError("Failed to disable MFA - please try again"))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"message": "MFA disabled successfully",
	})
}

// GetMfaStatus retrieves the MFA status for the authenticated user
func (h *MfaHandler) GetMfaStatus(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("UserID")
	if userID == 0 {
		utils.RespondWithError(
			c,
			apperror.NewParseError("Invalid UserID"),
		)
		return
	}

	// Get MFA status
	enabled, err := h.mfaService.GetMfaStatus(userID)
	if err != nil {
		logrus.Errorf("Failed to get MFA status for user %d: %v", userID, err)
		utils.RespondWithError(c, apperror.NewInternalError("Failed to retrieve MFA status"))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"mfa_enabled": enabled,
	})
}
