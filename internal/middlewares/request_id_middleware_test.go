package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates new request ID when not provided", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.RequestIDMiddleware())

		var capturedRequestID string
		router.GET("/test", func(c *gin.Context) {
			capturedRequestID = middlewares.GetRequestID(c)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.NotEmpty(t, capturedRequestID)

		_, err := uuid.Parse(capturedRequestID)
		require.NoError(t, err, "Generated request ID should be a valid UUID")

		assert.Equal(t, capturedRequestID, resp.Header().Get(middlewares.RequestIDHeader))
	})

	t.Run("uses client-provided request ID", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.RequestIDMiddleware())

		clientRequestID := "custom-request-id-123"
		var capturedRequestID string

		router.GET("/test", func(c *gin.Context) {
			capturedRequestID = middlewares.GetRequestID(c)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(middlewares.RequestIDHeader, clientRequestID)
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, clientRequestID, capturedRequestID)
		assert.Equal(t, clientRequestID, resp.Header().Get(middlewares.RequestIDHeader))
	})

	t.Run("request ID accessible in context", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.RequestIDMiddleware())

		router.GET("/test", func(c *gin.Context) {
			requestID, exists := c.Get(middlewares.RequestIDKey)
			assert.True(t, exists, "Request ID should exist in context")
			assert.NotEmpty(t, requestID, "Request ID should not be empty")

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

	t.Run("each request gets unique ID", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.RequestIDMiddleware())

		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Act - First request
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp1 := httptest.NewRecorder()
		router.ServeHTTP(resp1, req1)
		requestID1 := resp1.Header().Get(middlewares.RequestIDHeader)

		// Act - Second request
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)
		requestID2 := resp2.Header().Get(middlewares.RequestIDHeader)

		// Assert
		assert.NotEmpty(t, requestID1)
		assert.NotEmpty(t, requestID2)
		assert.NotEqual(t, requestID1, requestID2, "Each request should have a unique ID")
	})

	t.Run("works with multiple routes", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.RequestIDMiddleware())

		router.GET("/route1", func(c *gin.Context) {
			assert.NotEmpty(t, middlewares.GetRequestID(c))
			c.Status(http.StatusOK)
		})

		router.POST("/route2", func(c *gin.Context) {
			assert.NotEmpty(t, middlewares.GetRequestID(c))
			c.Status(http.StatusCreated)
		})

		// Act & Assert - Route 1
		req1 := httptest.NewRequest(http.MethodGet, "/route1", nil)
		resp1 := httptest.NewRecorder()
		router.ServeHTTP(resp1, req1)
		assert.Equal(t, http.StatusOK, resp1.Code)
		assert.NotEmpty(t, resp1.Header().Get(middlewares.RequestIDHeader))

		// Act & Assert - Route 2
		req2 := httptest.NewRequest(http.MethodPost, "/route2", nil)
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)
		assert.Equal(t, http.StatusCreated, resp2.Code)
		assert.NotEmpty(t, resp2.Header().Get(middlewares.RequestIDHeader))
	})
}

func TestGetRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns request ID when present", func(t *testing.T) {
		// Arrange
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		expectedID := "test-request-id-456"
		c.Set(middlewares.RequestIDKey, expectedID)

		// Act
		actualID := middlewares.GetRequestID(c)

		// Assert
		assert.Equal(t, expectedID, actualID)
	})

	t.Run("returns empty string when request ID not present", func(t *testing.T) {
		// Arrange
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		// Act
		actualID := middlewares.GetRequestID(c)

		// Assert
		assert.Empty(t, actualID)
	})

	t.Run("returns empty string when request ID is wrong type", func(t *testing.T) {
		// Arrange
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(middlewares.RequestIDKey, 12345)

		// Act
		actualID := middlewares.GetRequestID(c)

		// Assert
		assert.Empty(t, actualID)
	})

	t.Run("returns empty string when request ID is nil", func(t *testing.T) {
		// Arrange
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(middlewares.RequestIDKey, nil)

		// Act
		actualID := middlewares.GetRequestID(c)

		// Assert
		assert.Empty(t, actualID)
	})
}
