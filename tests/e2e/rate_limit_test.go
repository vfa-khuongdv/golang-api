package e2e

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginRateLimit(t *testing.T) {
	router, _ := setupTestRouter()

	payload := `{"email":"test@example.com","password":"password123"}`
	limit := 10

	for i := 0; i < limit; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusTooManyRequests, w.Code,
			"Request %d should not be rate limited yet", i+1)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
}

// Regression: with no trusted proxies configured, the X-Forwarded-For header
// must be ignored when deriving the rate-limit key. Spoofing a fresh XFF value
// per request must NOT bypass the per-IP limit.
func TestLoginRateLimitIgnoresSpoofedXForwardedFor(t *testing.T) {
	router, _ := setupTestRouter()

	payload := `{"email":"test@example.com","password":"password123"}`
	limit := 10

	for i := 0; i < limit; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		// Each request claims a different client IP. This must have no effect.
		req.Header.Set("X-Forwarded-For", "10.0.0."+strconv.Itoa(i+1))
		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusTooManyRequests, w.Code,
			"Request %d should not be rate limited yet", i+1)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "10.0.0.99")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
}

func TestForgotPasswordRateLimit(t *testing.T) {
	router, _ := setupTestRouter()

	payload := `{"email":"test@example.com"}`
	limit := 10

	for i := 0; i < limit; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/forgot-password", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusTooManyRequests, w.Code,
			"Request %d should not be rate limited yet", i+1)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/forgot-password", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
}
