package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/handlers"
)

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRouter := gin.Default()
	mockRouter.GET("/health", handlers.HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	mockRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestVersionInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRouter := gin.Default()
	mockRouter.GET("/version", handlers.VersionInfo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/version", nil)
	mockRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", response["version"])
	assert.NotEmpty(t, response["build_time"])
	assert.NotEmpty(t, response["uptime"])
}

func serveReadiness(ping func(context.Context) error) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/readyz", handlers.ReadinessCheck(ping))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	return w
}

func TestReadinessCheck_ReadyWhenTheDatabaseAnswers(t *testing.T) {
	w := serveReadiness(func(context.Context) error { return nil })

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ready"}`, w.Body.String())
}

func TestReadinessCheck_UnavailableWhenTheDatabaseFails(t *testing.T) {
	w := serveReadiness(func(context.Context) error { return errors.New("connection refused") })

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.NotContains(t, w.Body.String(), "connection refused")
}

func TestReadinessCheck_PingsWithADeadline(t *testing.T) {
	var hasDeadline bool
	serveReadiness(func(ctx context.Context) error {
		_, hasDeadline = ctx.Deadline()
		return nil
	})

	assert.True(t, hasDeadline)
}
