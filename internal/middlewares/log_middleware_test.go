package middlewares_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
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

type errReadCloser struct{}

func (e errReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("read body failed")
}

func (e errReadCloser) Close() error {
	return nil
}

func setupLogCapture() (*syncBuffer, func()) {
	var buf syncBuffer
	prevOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	return &buf, func() { logrus.SetOutput(prevOutput) }
}

func lastCompleteJSON(b []byte) []byte {
	for _, line := range bytes.Split(b, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var m map[string]interface{}
		if json.Unmarshal(line, &m) == nil {
			return line
		}
	}
	return nil
}

func waitForLog(t *testing.T, buf *syncBuffer, timeout time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last []byte
	for time.Now().Before(deadline) {
		if cur := lastCompleteJSON(buf.Bytes()); cur != nil {
			if !bytes.Equal(cur, last) {
				last = cur
				time.Sleep(30 * time.Millisecond)
				continue
			}
			return cur
		}
		time.Sleep(10 * time.Millisecond)
	}
	return last
}

func TestLogMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("logs POST request body with sensitive data censored", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())

		router.POST("/test", func(c *gin.Context) {
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

		reqBody := map[string]interface{}{
			"username": "user1",
			"password": "secret_password",
			"email":    "user@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "info", logEntry["level"])
		assert.Equal(t, "POST", logEntry["method"])
		assert.Equal(t, "/test", logEntry["url"])
		assert.Equal(t, "200", logEntry["status_code"])

		reqMap, ok := logEntry["request"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "user1", reqMap["username"])
		assert.NotEqual(t, "secret_password", reqMap["password"])
		assert.Contains(t, reqMap["password"], "*")
		assert.NotEqual(t, "user@example.com", reqMap["email"])
		assert.Contains(t, reqMap["email"], "*")
		assert.Equal(t, middlewares.NotLoggedResponse, logEntry["response"])
	})

	t.Run("GET request succeeds and response not logged", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "GET", logEntry["method"])
		assert.Equal(t, "/ping", logEntry["url"])
		assert.Equal(t, "200", logEntry["status_code"])
		assert.Equal(t, middlewares.NotLoggedResponse, logEntry["response"])
	})

	t.Run("large request body truncated to 64KB", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.POST("/large", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		largeBody := strings.Repeat("a", (1<<16)+100)
		req := httptest.NewRequest(http.MethodPost, "/large", strings.NewReader(largeBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.NotEmpty(t, logEntry["request"])
		reqStr, ok := logEntry["request"].(string)
		assert.True(t, ok)
		assert.True(t, len(reqStr) <= (1<<16))
	})

	t.Run("large response body not logged for success status", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.GET("/large-response", func(c *gin.Context) {
			largeData := strings.Repeat("x", (1<<16)+1000)
			c.String(http.StatusOK, largeData)
		})

		req := httptest.NewRequest(http.MethodGet, "/large-response", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, middlewares.NotLoggedResponse, logEntry["response"])
	})

	t.Run("sensitive headers are censored in log output", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer secret-token-123")
		req.Header.Set("Cookie", "session_id=abc123")
		req.Header.Set("X-API-Key", "api-key-xyz")
		req.Header.Set("X-Custom-Header", "safe-value")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)

		headers, ok := logEntry["header"].(map[string]interface{})
		assert.True(t, ok)

		authHeader, ok := headers["Authorization"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Bearer secr*****", authHeader[0])

		cookieHeader, ok := headers["Cookie"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, "session_id=abc1*****", cookieHeader[0])

		apiKeyHeader, ok := headers["X-Api-Key"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, "api-*****", apiKeyHeader[0])

		customHeader, ok := headers["X-Custom-Header"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, "safe-value", customHeader[0])
	})

	t.Run("malformed request JSON logged as string", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.POST("/malformed", func(c *gin.Context) {
			c.Header("Content-Type", "application/json")
			c.String(http.StatusOK, "{invalid json:")
		})

		req := httptest.NewRequest(http.MethodPost, "/malformed", strings.NewReader("{invalid json:}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "POST", logEntry["method"])
		assert.Equal(t, "/malformed", logEntry["url"])
		assert.NotEmpty(t, logEntry["request"])
		assert.Equal(t, middlewares.NotLoggedResponse, logEntry["response"])
	})

	t.Run("non-JSON content type logged as string", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.POST("/text", func(c *gin.Context) {
			c.String(http.StatusOK, "plain text response")
		})

		req := httptest.NewRequest(http.MethodPost, "/text", strings.NewReader("plain text request"))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "plain text request", logEntry["request"])
		assert.Equal(t, middlewares.NotLoggedResponse, logEntry["response"])
	})

	t.Run("request body read error does not crash", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.POST("/read-error", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest(http.MethodPost, "/read-error", nil)
		req.Body = errReadCloser{}
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, waitForLog(t, buf, time.Second))
	})

	t.Run("concurrent requests handled safely", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.POST("/concurrent", func(c *gin.Context) {
			var req map[string]interface{}
			_ = c.ShouldBindJSON(&req)
			c.JSON(http.StatusOK, gin.H{"id": req["id"]})
		})

		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				reqBody := map[string]interface{}{"id": id, "password": "secret"}
				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/concurrent", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer token-"+fmt.Sprintf("%d", id))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				done <- true
			}(i)
		}
		for i := 0; i < 10; i++ {
			<-done
		}

		assert.NotEmpty(t, waitForLog(t, buf, 2*time.Second))
	})

	t.Run("PUT and PATCH request bodies are logged", func(t *testing.T) {
		for _, method := range []string{"PUT", "PATCH"} {
			t.Run(method, func(t *testing.T) {
				buf, restore := setupLogCapture()
				defer restore()

				router := gin.New()
				router.Use(middlewares.LogMiddleware())
				router.Handle(method, "/resource", func(c *gin.Context) {
					var req map[string]interface{}
					_ = c.ShouldBindJSON(&req)
					c.JSON(http.StatusOK, gin.H{"updated": true})
				})

				reqBody := map[string]interface{}{"password": "secret123"}
				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(method, "/resource", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)
				time.Sleep(50 * time.Millisecond)

				var logEntry map[string]interface{}
				err := json.Unmarshal(buf.Bytes(), &logEntry)
				assert.NoError(t, err)
				assert.Equal(t, method, logEntry["method"])
				reqMap, ok := logEntry["request"].(map[string]interface{})
				assert.True(t, ok)
				assert.Contains(t, reqMap["password"], "*")
			})
		}
	})

	t.Run("empty POST body does not cause errors", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.POST("/empty", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest(http.MethodPost, "/empty", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "POST", logEntry["method"])
		assert.Equal(t, "/empty", logEntry["url"])
		assert.NotNil(t, logEntry["request"])
	})

	t.Run("error response body censored and logged", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.POST("/error", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "internal server error",
				"secret": "leaked_data_123",
				"token":  "sensitive_token_xyz",
			})
		})

		req := httptest.NewRequest(http.MethodPost, "/error", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "500", logEntry["status_code"])
		assert.Equal(t, "error", logEntry["level"])

		respMap, ok := logEntry["response"].(map[string]interface{})
		assert.True(t, ok, "response should be logged for error status codes")
		assert.Equal(t, "internal server error", respMap["error"])
		assert.NotEqual(t, "leaked_data_123", respMap["secret"])
		assert.Contains(t, respMap["secret"], "*")
		assert.NotEqual(t, "sensitive_token_xyz", respMap["token"])
		assert.Contains(t, respMap["token"], "*")
	})

	t.Run("non-JSON error response logged as string", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.GET("/error-text", func(c *gin.Context) {
			c.String(http.StatusInternalServerError, "internal server error text")
		})

		req := httptest.NewRequest(http.MethodGet, "/error-text", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "500", logEntry["status_code"])
		assert.Equal(t, "error", logEntry["level"])

		respStr, ok := logEntry["response"].(string)
		assert.True(t, ok, "response should be logged as string for non-JSON error")
		assert.Equal(t, "internal server error text", respStr)
	})

	t.Run("malformed JSON error response logged as string", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.GET("/error-malformed", func(c *gin.Context) {
			c.Header("Content-Type", "application/json")
			c.String(http.StatusInternalServerError, "{invalid json response")
		})

		req := httptest.NewRequest(http.MethodGet, "/error-malformed", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "500", logEntry["status_code"])
		assert.Equal(t, "error", logEntry["level"])

		respStr, ok := logEntry["response"].(string)
		assert.True(t, ok, "response should be logged as string when JSON unmarshal fails")
		assert.Contains(t, respStr, "invalid json response")
	})

	t.Run("large error response body truncated", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.GET("/error-large", func(c *gin.Context) {
			largeData := strings.Repeat("x", (1<<16)+1000)
			c.String(http.StatusInternalServerError, largeData)
		})

		req := httptest.NewRequest(http.MethodGet, "/error-large", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "500", logEntry["status_code"])
		assert.Equal(t, "error", logEntry["level"])

		respStr, ok := logEntry["response"].(string)
		assert.True(t, ok, "response should be logged as string for large non-JSON error")
		assert.True(t, len(respStr) <= (1<<16), "response should be truncated to MAX_BODY_SIZE")
	})

	t.Run("client error response censored and logged", func(t *testing.T) {
		buf, restore := setupLogCapture()
		defer restore()

		router := gin.New()
		router.Use(middlewares.LogMiddleware())
		router.POST("/bad-request", func(c *gin.Context) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid input",
				"email":   "user@example.com",
				"session": "session_token_abc",
			})
		})

		req := httptest.NewRequest(http.MethodPost, "/bad-request", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		var logEntry map[string]interface{}
		err := json.Unmarshal(waitForLog(t, buf, time.Second), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "400", logEntry["status_code"])
		assert.Equal(t, "warning", logEntry["level"])

		respMap, ok := logEntry["response"].(map[string]interface{})
		assert.True(t, ok, "response should be logged for 4xx error status codes")
		assert.Equal(t, "invalid input", respMap["error"])
		assert.NotEqual(t, "user@example.com", respMap["email"])
		assert.Contains(t, respMap["email"], "*")
		assert.NotEqual(t, "session_token_abc", respMap["session"])
		assert.Contains(t, respMap["session"], "*")
	})
}
