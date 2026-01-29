package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Generates new request ID when not provided", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(RequestIDMiddleware())

		var capturedRequestID string
		router.GET("/test", func(c *gin.Context) {
			capturedRequestID = GetRequestID(c)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.NotEmpty(t, capturedRequestID)

		// Verify it's a valid UUID
		_, err := uuid.Parse(capturedRequestID)
		require.NoError(t, err, "Generated request ID should be a valid UUID")

		// Verify request ID is in response header
		assert.Equal(t, capturedRequestID, resp.Header().Get(RequestIDHeader))
	})

	t.Run("Uses client-provided request ID", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(RequestIDMiddleware())

		clientRequestID := "custom-request-id-123"
		var capturedRequestID string

		router.GET("/test", func(c *gin.Context) {
			capturedRequestID = GetRequestID(c)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(RequestIDHeader, clientRequestID)
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, clientRequestID, capturedRequestID)
		assert.Equal(t, clientRequestID, resp.Header().Get(RequestIDHeader))
	})

	t.Run("Request ID accessible in context", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(RequestIDMiddleware())

		router.GET("/test", func(c *gin.Context) {
			requestID, exists := c.Get(RequestIDKey)
			assert.True(t, exists, "Request ID should exist in context")
			assert.NotEmpty(t, requestID, "Request ID should not be empty")

			// Verify it's a string
			_, ok := requestID.(string)
			assert.True(t, ok, "Request ID should be a string")

			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("Each request gets unique ID", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(RequestIDMiddleware())

		var requestID1, requestID2 string

		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Act - First request
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp1 := httptest.NewRecorder()
		router.ServeHTTP(resp1, req1)
		requestID1 = resp1.Header().Get(RequestIDHeader)

		// Act - Second request
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)
		requestID2 = resp2.Header().Get(RequestIDHeader)

		// Assert
		assert.NotEmpty(t, requestID1)
		assert.NotEmpty(t, requestID2)
		assert.NotEqual(t, requestID1, requestID2, "Each request should have a unique ID")
	})

	t.Run("Middleware works with multiple routes", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(RequestIDMiddleware())

		router.GET("/route1", func(c *gin.Context) {
			assert.NotEmpty(t, GetRequestID(c))
			c.Status(http.StatusOK)
		})

		router.POST("/route2", func(c *gin.Context) {
			assert.NotEmpty(t, GetRequestID(c))
			c.Status(http.StatusCreated)
		})

		// Act & Assert - Route 1
		req1 := httptest.NewRequest(http.MethodGet, "/route1", nil)
		resp1 := httptest.NewRecorder()
		router.ServeHTTP(resp1, req1)
		assert.Equal(t, http.StatusOK, resp1.Code)
		assert.NotEmpty(t, resp1.Header().Get(RequestIDHeader))

		// Act & Assert - Route 2
		req2 := httptest.NewRequest(http.MethodPost, "/route2", nil)
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)
		assert.Equal(t, http.StatusCreated, resp2.Code)
		assert.NotEmpty(t, resp2.Header().Get(RequestIDHeader))
	})
}

func TestGetRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Returns request ID when present", func(t *testing.T) {
		// Arrange
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		expectedID := "test-request-id-456"
		c.Set(RequestIDKey, expectedID)

		// Act
		actualID := GetRequestID(c)

		// Assert
		assert.Equal(t, expectedID, actualID)
	})

	t.Run("Returns empty string when request ID not present", func(t *testing.T) {
		// Arrange
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		// Act
		actualID := GetRequestID(c)

		// Assert
		assert.Empty(t, actualID)
	})

	t.Run("Returns empty string when request ID is wrong type", func(t *testing.T) {
		// Arrange
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(RequestIDKey, 12345) // Set as int instead of string

		// Act
		actualID := GetRequestID(c)

		// Assert
		assert.Empty(t, actualID)
	})

	t.Run("Returns empty string when request ID is nil", func(t *testing.T) {
		// Arrange
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(RequestIDKey, nil)

		// Act
		actualID := GetRequestID(c)

		// Assert
		assert.Empty(t, actualID)
	})
}
