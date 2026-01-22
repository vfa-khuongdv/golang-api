package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetUserIDFromContext - Success", func(t *testing.T) {
		// Arrange
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		var expectedUserID uint = 1
		c.Set("UserID", expectedUserID)

		// Act
		userID, err := GetUserIDFromContext(c)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedUserID, userID)
	})

	t.Run("GetUserIDFromContext - UserID Not Found", func(t *testing.T) {
		// Arrange
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Act
		userID, err := GetUserIDFromContext(c)

		// Assert
		require.Error(t, err)
		assert.Equal(t, uint(0), userID)
		assert.EqualError(t, err, "User ID not found in context")
	})

	t.Run("GetUserIDFromContext - Invalid Type", func(t *testing.T) {
		// Arrange
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("UserID", "1") // wrong type

		// Act
		userID, err := GetUserIDFromContext(c)

		// Assert
		require.Error(t, err)
		assert.Equal(t, uint(0), userID)
		assert.EqualError(t, err, "User ID in context has invalid type")
	})
}
