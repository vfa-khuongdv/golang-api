package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogMiddleware(t *testing.T) {
	// Setup log capture
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil) // Reset after test

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.POST("/test", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    req,
			"secret":  "hidden_response_value",
			"token":   "response_token_123",
		})
	})

	// Create a request with sensitive data
	reqBody := map[string]interface{}{
		"username": "user1",
		"password": "secret_password",
		"email":    "user@example.com",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Wait for goroutine to finish logging
	time.Sleep(100 * time.Millisecond)

	// Parse log output
	var logEntry struct {
		Level   string `json:"level"`
		Message string `json:"msg"`
	}

	// The logger.Info logs the JSON string as the message, so we need to parse the message itself
	// However, looking at log_middleware.go:
	// logger.Info(string(jsonData))
	// This means the "msg" field of the logrus JSON output will contain the JSON string of LogResponse.

	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "info", logEntry.Level)

	var logResponse LogResponse
	err = json.Unmarshal([]byte(logEntry.Message), &logResponse)
	assert.NoError(t, err)

	// Verify fields
	assert.Equal(t, "POST", logResponse.Method)
	assert.Equal(t, "/test", logResponse.URL)
	assert.Equal(t, "200", logResponse.StatusCode)

	// Verify Request Body Censoring
	reqMap, ok := logResponse.Request.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "user1", reqMap["username"])
	assert.NotEqual(t, "secret_password", reqMap["password"])
	assert.Contains(t, reqMap["password"], "*")
	assert.NotEqual(t, "user@example.com", reqMap["email"])
	assert.Contains(t, reqMap["email"], "*")

	// Verify Response Body Censoring
	respMap, ok := logResponse.Response.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "success", respMap["message"])
	assert.NotEqual(t, "response_token_123", respMap["token"])
	assert.Contains(t, respMap["token"], "*")
}

func TestLogMiddleware_GetRequest(t *testing.T) {
	// Setup log capture
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	var logEntry struct {
		Message string `json:"msg"`
	}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	var logResponse LogResponse
	err = json.Unmarshal([]byte(logEntry.Message), &logResponse)
	assert.NoError(t, err)

	assert.Equal(t, "GET", logResponse.Method)
	assert.Equal(t, "/ping", logResponse.URL)
	assert.Equal(t, "200", logResponse.StatusCode)
	assert.Equal(t, "pong", logResponse.Response)
}

func TestLogMiddleware_LargeBody(t *testing.T) {
	// Setup log capture
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.POST("/large", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Create a body larger than 64KB
	largeBody := strings.Repeat("a", (1<<16)+100)
	req, _ := http.NewRequest("POST", "/large", strings.NewReader(largeBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	var logEntry struct {
		Message string `json:"msg"`
	}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	var logResponse LogResponse
	err = json.Unmarshal([]byte(logEntry.Message), &logResponse)
	assert.NoError(t, err)

	// The middleware reads up to 64KB.
	// Since it's not JSON, it logs as string.
	// We verify that it captured something and didn't crash.
	assert.NotEmpty(t, logResponse.Request)
	reqStr, ok := logResponse.Request.(string)
	assert.True(t, ok)
	assert.True(t, len(reqStr) <= (1<<16))
}
