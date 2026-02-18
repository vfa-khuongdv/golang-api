package middlewares

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// syncBuffer is a thread-safe wrapper for bytes.Buffer
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

type errReadCloser struct{}

func (e errReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("read body failed")
}

func (e errReadCloser) Close() error {
	return nil
}

func (sb *syncBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *syncBuffer) Bytes() []byte {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Bytes()
}

func TestLogMiddleware(t *testing.T) {
	// Setup log capture with thread-safe buffer
	var buf syncBuffer
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
	// Setup log capture with thread-safe buffer
	var buf syncBuffer
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
	// Setup log capture with thread-safe buffer
	var buf syncBuffer
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
func TestLogMiddleware_LargeResponseBody(t *testing.T) {
	// Setup log capture with thread-safe buffer
	var buf syncBuffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.GET("/large-response", func(c *gin.Context) {
		// Return a response larger than 64KB
		largeData := strings.Repeat("x", (1<<16)+1000)
		c.String(http.StatusOK, largeData)
	})

	req, _ := http.NewRequest("GET", "/large-response", nil)
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

	// Verify response was truncated to maxBodySize (64KB)
	respStr, ok := logResponse.Response.(string)
	assert.True(t, ok)
	assert.LessOrEqual(t, len(respStr), 1<<16, "Response should be truncated to 64KB")
}

func TestLogMiddleware_SensitiveHeaders(t *testing.T) {
	// Setup log capture with thread-safe buffer
	var buf syncBuffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer secret-token-123")
	req.Header.Set("Cookie", "session_id=abc123")
	req.Header.Set("X-API-Key", "api-key-xyz")
	req.Header.Set("X-Custom-Header", "safe-value")

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

	// Verify sensitive headers are censored
	headers, ok := logResponse.Header.(map[string]interface{})
	assert.True(t, ok)

	authHeader, ok := headers["Authorization"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "*****", authHeader[0])

	cookieHeader, ok := headers["Cookie"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "*****", cookieHeader[0])

	apiKeyHeader, ok := headers["X-Api-Key"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "*****", apiKeyHeader[0])

	// Verify non-sensitive headers are not censored
	customHeader, ok := headers["X-Custom-Header"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "safe-value", customHeader[0])
}

func TestLogMiddleware_MalformedJSON(t *testing.T) {
	// Setup log capture with thread-safe buffer
	var buf syncBuffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.POST("/malformed", func(c *gin.Context) {
		// Return malformed JSON in response
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, "{invalid json:")
	})

	req, _ := http.NewRequest("POST", "/malformed", strings.NewReader("{invalid json:}"))
	req.Header.Set("Content-Type", "application/json")

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

	// Verify malformed JSON is logged as string, not causing crash
	assert.Equal(t, "POST", logResponse.Method)
	assert.Equal(t, "/malformed", logResponse.URL)
	assert.NotEmpty(t, logResponse.Request)
	assert.NotEmpty(t, logResponse.Response)
}

func TestLogMiddleware_NonJSONContentType(t *testing.T) {
	// Setup log capture with thread-safe buffer
	var buf syncBuffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.POST("/text", func(c *gin.Context) {
		c.String(http.StatusOK, "plain text response")
	})

	req, _ := http.NewRequest("POST", "/text", strings.NewReader("plain text request"))
	req.Header.Set("Content-Type", "text/plain")

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

	// Verify non-JSON content is logged as string
	assert.Equal(t, "plain text request", logResponse.Request)
	assert.Equal(t, "plain text response", logResponse.Response)
}

func TestLogMiddleware_RequestBodyReadError(t *testing.T) {
	var buf syncBuffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.POST("/read-error", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("POST", "/read-error", nil)
	req.Body = errReadCloser{}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, buf.Bytes())
}

func TestLogMiddleware_MarshalLogEntryError(t *testing.T) {
	var buf syncBuffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	originalMarshal := marshalLogEntry
	marshalLogEntry = func(_ any) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}
	defer func() {
		marshalLogEntry = originalMarshal
	}()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())
	r.GET("/marshal-error", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/marshal-error", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, buf.Bytes())
}

func TestLogMiddleware_Concurrent(t *testing.T) {
	// Setup log capture with thread-safe buffer
	var buf syncBuffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.POST("/concurrent", func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		c.JSON(http.StatusOK, gin.H{"id": req["id"]})
	})

	// Send multiple concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			reqBody := map[string]interface{}{"id": id, "password": "secret"}
			bodyBytes, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/concurrent", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer token-"+string(rune(id)))

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	time.Sleep(200 * time.Millisecond)

	// Verify no crashes and logs were produced
	// The exact number of log entries may vary due to concurrent writes
	// but we verify no panic occurred
	assert.NotEmpty(t, buf.Bytes())
}

func TestLogMiddleware_PUTandPATCH(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"PUT request", "PUT"},
		{"PATCH request", "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf syncBuffer
			logrus.SetOutput(&buf)
			logrus.SetFormatter(&logrus.JSONFormatter{})
			defer logrus.SetOutput(nil)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(LogMiddleware())

			r.Handle(tt.method, "/resource", func(c *gin.Context) {
				var req map[string]interface{}
				c.ShouldBindJSON(&req)
				c.JSON(http.StatusOK, gin.H{"updated": true})
			})

			reqBody := map[string]interface{}{"password": "secret123"}
			bodyBytes, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest(tt.method, "/resource", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

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

			assert.Equal(t, tt.method, logResponse.Method)

			// Verify password was censored
			reqMap, ok := logResponse.Request.(map[string]interface{})
			assert.True(t, ok)
			assert.Contains(t, reqMap["password"], "*")
		})
	}
}

func TestLogMiddleware_EmptyBody(t *testing.T) {
	// Setup log capture with thread-safe buffer
	var buf syncBuffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LogMiddleware())

	r.POST("/empty", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("POST", "/empty", nil)
	req.Header.Set("Content-Type", "application/json")

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

	assert.Equal(t, "POST", logResponse.Method)
	assert.Equal(t, "/empty", logResponse.URL)
	// Empty body should not cause errors
	assert.NotNil(t, logResponse.Request)
}
