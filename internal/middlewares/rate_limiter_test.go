package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Allows requests within limit", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.RateLimiter(5, time.Second))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		// Act & Assert
		expectedRemaining := []string{"4", "3", "2", "1", "0"}
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
			assert.Equal(t, expectedRemaining[i], w.Header().Get("X-RateLimit-Remaining"))
			assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
		}
	})

	t.Run("Blocks requests over limit", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.RateLimiter(2, time.Second))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		// Act
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w3, req3)

		// Assert
		assert.Equal(t, http.StatusTooManyRequests, w3.Code)
		assert.Equal(t, "2", w3.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "0", w3.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w3.Header().Get("X-RateLimit-Reset"))
	})

	t.Run("Different IPs have separate limits", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.RateLimiter(1, time.Second))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		// Act
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.1:1234"
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.2:1234"
		router.ServeHTTP(w2, req2)

		// Assert
		assert.Equal(t, http.StatusOK, w2.Code)
	})

	t.Run("Resets after window expires", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.RateLimiter(1, 500*time.Millisecond))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)

		// Act - wait for window to expire
		time.Sleep(600 * time.Millisecond)

		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w3, req3)

		// Assert
		assert.Equal(t, http.StatusOK, w3.Code)
	})
}
