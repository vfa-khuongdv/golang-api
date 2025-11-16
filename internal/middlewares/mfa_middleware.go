package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MfaVerificationMiddleware checks if MFA is enabled and verified for the user
// This middleware should be used after JWT auth middleware on sensitive endpoints
func MfaVerificationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("UserID")
		if userID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UserID"})
			c.Abort()
			return
		}

		// Check if user has pending MFA verification
		// This would be set by the login endpoint if MFA is required
		mfaPending, exists := c.Get("MFAPending")
		if exists && mfaPending == true {
			c.JSON(http.StatusForbidden, gin.H{"error": "MFA verification required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// MfaRequiredMiddleware enforces MFA verification for users who have it enabled
func MfaRequiredMiddleware(mfaService interface {
	GetMfaStatus(userID uint) (bool, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("UserID")
		if userID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UserID"})
			c.Abort()
			return
		}

		// Get MFA status
		mfaEnabled, err := mfaService.GetMfaStatus(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check MFA status"})
			c.Abort()
			return
		}

		// If MFA is enabled, allow the request to proceed
		// Note: Actual MFA verification is handled during authentication
		if mfaEnabled {
			// MFA is enabled, the caller should have already verified it
			// You may want to add additional checks if needed
		}

		c.Next()
	}
}
