package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

func waitForLog(t *testing.T, buf *syncBuffer, timeout time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if b := buf.Bytes(); len(b) > 0 {
			return b
		}
		time.Sleep(10 * time.Millisecond)
	}
	return buf.Bytes()
}

func TestCensorQueryParams(t *testing.T) {
	t.Run("sensitive key masked", func(t *testing.T) {
		// Arrange
		input := map[string][]string{"token": {"abc123"}, "q": {"search"}}

		// Act
		result := censorQueryParams(input)

		// Assert
		assert.Equal(t, "abc1*****", result["token"][0])
		assert.Equal(t, "search", result["q"][0])
	})

	t.Run("non-sensitive key unchanged", func(t *testing.T) {
		// Arrange
		input := map[string][]string{"page": {"2"}, "limit": {"10"}}

		// Act
		result := censorQueryParams(input)

		// Assert
		assert.Equal(t, "2", result["page"][0])
		assert.Equal(t, "10", result["limit"][0])
	})

	t.Run("empty params", func(t *testing.T) {
		// Act
		result := censorQueryParams(map[string][]string{})

		// Assert
		assert.Empty(t, result)
	})
}

func TestContainsIgnoreCase(t *testing.T) {
	t.Run("match found", func(t *testing.T) {
		// Assert
		assert.True(t, containsIgnoreCase([]string{"Password", "Token"}, "password"))
		assert.True(t, containsIgnoreCase([]string{"Password", "Token"}, "TOKEN"))
		assert.True(t, containsIgnoreCase([]string{"api-key"}, "API-KEY"))
	})

	t.Run("no match", func(t *testing.T) {
		// Assert
		assert.False(t, containsIgnoreCase([]string{"Password", "Token"}, "username"))
		assert.False(t, containsIgnoreCase([]string{}, "anything"))
	})
}

func TestFilterSensitiveHeaders(t *testing.T) {
	t.Run("authorization with bearer", func(t *testing.T) {
		// Arrange
		input := map[string][]string{"Authorization": {"Bearer eyJhbGci"}}

		// Act
		result := filterSensitiveHeaders(input)

		// Assert
		assert.Equal(t, "Bearer eyJh*****", result["Authorization"][0])
	})

	t.Run("authorization without scheme", func(t *testing.T) {
		// Arrange
		input := map[string][]string{"Authorization": {"token-abc"}}

		// Act
		result := filterSensitiveHeaders(input)

		// Assert
		assert.Contains(t, result["Authorization"][0], "*****")
	})

	t.Run("cookie masked", func(t *testing.T) {
		// Arrange
		input := map[string][]string{"Cookie": {"session_id=abc123"}}

		// Act
		result := filterSensitiveHeaders(input)

		// Assert
		assert.Equal(t, "session_id=abc1*****", result["Cookie"][0])
	})

	t.Run("set-cookie masked", func(t *testing.T) {
		// Arrange
		input := map[string][]string{"Set-Cookie": {"session_id=abc123; Path=/; HttpOnly"}}

		// Act
		result := filterSensitiveHeaders(input)

		// Assert
		assert.Equal(t, "session_id=abc1*****; Path=/; HttpOnly", result["Set-Cookie"][0])
	})

	t.Run("non-sensitive header unchanged", func(t *testing.T) {
		// Arrange
		input := map[string][]string{"Content-Type": {"application/json"}}

		// Act
		result := filterSensitiveHeaders(input)

		// Assert
		assert.Equal(t, "application/json", result["Content-Type"][0])
	})
}

func TestLogMiddleware_LogsRequest(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	var buf syncBuffer
	prevOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer logrus.SetOutput(prevOutput)

	router := gin.New()
	router.Use(LogMiddleware())
	router.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	resp := httptest.NewRecorder()

	// Act
	router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.NotEmpty(t, waitForLog(t, &buf, time.Second))
}
