package e2e

import (
	"net/http"
	"net/http/httptest"
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

func TestForgotPasswordRateLimit(t *testing.T) {
	router, _ := setupTestRouter()

	payload := `{"email":"test@example.com"}`
	limit := 10

	for i := 0; i < limit; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/forgot-password", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/forgot-password", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
