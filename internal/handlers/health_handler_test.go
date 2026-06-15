package handlers_test

import (
	"encoding/json"
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
