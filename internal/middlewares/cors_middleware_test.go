package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	setupRouter := func() *gin.Engine {
		router := gin.New()
		router.Use(middlewares.CORSMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})
		router.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"message": "created"})
		})
		return router
	}

	t.Run("Single Allowed Origin - Success", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "https://example.com", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", resp.Header().Get("Vary"))
		assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With", resp.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "POST, OPTIONS, GET, PUT, PATCH, DELETE", resp.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "86400", resp.Header().Get("Access-Control-Max-Age"))
		assert.Equal(t, "Content-Length, Authorization", resp.Header().Get("Access-Control-Expose-Headers"))
	})

	t.Run("Multiple Allowed Origins - First Origin Success", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "https://app1.com,https://app2.com,https://app3.com"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://app1.com")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "https://app1.com", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", resp.Header().Get("Vary"))
	})

	t.Run("Multiple Allowed Origins - Middle Origin Success", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "https://app1.com,https://app2.com,https://app3.com"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://app2.com")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "https://app2.com", resp.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Multiple Allowed Origins - With Spaces", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "https://app1.com, https://app2.com , https://app3.com"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://app2.com")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "https://app2.com", resp.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Wildcard Origin - Allows Any Origin", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "*"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://unknown-origin.com")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "https://unknown-origin.com", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", resp.Header().Get("Vary"))
	})

	t.Run("Rejected Origin - Not In Allowed List", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://malicious.com")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Empty(t, resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, resp.Header().Get("Vary"))
		// Other CORS headers should still be set
		assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "POST, OPTIONS, GET, PUT, PATCH, DELETE", resp.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("No Origin Header - Rejected", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// No Origin header set
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Empty(t, resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, resp.Header().Get("Vary"))
	})

	t.Run("OPTIONS Preflight Request - Success", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, resp.Code)
		assert.Equal(t, "https://example.com", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", resp.Header().Get("Vary"))
		assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "POST, OPTIONS, GET, PUT, PATCH, DELETE", resp.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "86400", resp.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("OPTIONS Preflight Request - Rejected Origin", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://malicious.com")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, resp.Code)
		assert.Empty(t, resp.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Default Localhost Origin", func(t *testing.T) {
		// Arrange - No environment variable set
		os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "http://localhost:3000", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", resp.Header().Get("Vary"))
	})

	t.Run("POST Request - CORS Headers Applied", func(t *testing.T) {
		// Arrange
		require.NoError(t, os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com"))
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		router := setupRouter()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusCreated, resp.Code)
		assert.Equal(t, "https://example.com", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", resp.Header().Get("Vary"))
		assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
	})
}
