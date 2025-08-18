package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Set environment variable for CORS_ALLOWED_ORIGINS for this test
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")
	defer func() {
		if err := os.Unsetenv("CORS_ALLOWED_ORIGINS"); err != nil {
			panic(err) // or t.Fatalf("failed to unset env: %v", err)
		}
	}()

	router := gin.New()
	router.Use(middlewares.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Test normal GET request (should set CORS headers and pass)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "https://example.com", resp.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With", resp.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "POST, OPTIONS, GET, PUT, PATCH, DELETE", resp.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "86400", resp.Header().Get("Access-Control-Max-Age"))

	// Test OPTIONS request (should abort with 204)
	reqOptions := httptest.NewRequest(http.MethodOptions, "/test", nil)
	respOptions := httptest.NewRecorder()
	router.ServeHTTP(respOptions, reqOptions)

	assert.Equal(t, http.StatusNoContent, respOptions.Code) // 204
	// Headers should also be set on OPTIONS
	assert.Equal(t, "https://example.com", respOptions.Header().Get("Access-Control-Allow-Origin"))
}
