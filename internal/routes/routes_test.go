package routes_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/routes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRouter(t *testing.T) *gin.Engine {
	t.Helper()
	t.Setenv("JWT_KEY", "this-is-a-very-long-secret-key-for-routes-testing-32-chars")
	t.Setenv("SETTINGS_ENCRYPTION_KEY", strings.Repeat("r", 40))
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	return routes.SetupRouter(db)
}

// remainingFor sends a rate-limited request as if it came through a proxy at
// remoteAddr carrying the given X-Forwarded-For, and returns the remaining quota
// of the bucket the rate limiter put it in.
func remainingFor(router *gin.Engine, remoteAddr, xff string) string {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", nil)
	req.RemoteAddr = remoteAddr
	req.Header.Set("X-Forwarded-For", xff)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Header().Get("X-RateLimit-Remaining")
}

func TestSetupRouter_TrustedProxies(t *testing.T) {
	t.Run("default trusts the proxy so each client gets its own rate limit bucket", func(t *testing.T) {
		require.NoError(t, os.Unsetenv("TRUSTED_PROXIES"))
		router := newRouter(t)

		// Behind an ALB the peer is the ALB and the client IP is in X-Forwarded-For.
		assert.Equal(t, "9", remainingFor(router, "10.0.0.5:1234", "203.0.113.1"))
		assert.Equal(t, "9", remainingFor(router, "10.0.0.5:1234", "203.0.113.2"))
		assert.Equal(t, "8", remainingFor(router, "10.0.0.5:1234", "203.0.113.1"))
	})

	t.Run("explicitly empty trusts nothing so all clients behind a proxy share one bucket", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXIES", "")
		router := newRouter(t)

		assert.Equal(t, "9", remainingFor(router, "10.0.0.5:1234", "203.0.113.1"))
		assert.Equal(t, "8", remainingFor(router, "10.0.0.5:1234", "203.0.113.2"))
	})

	t.Run("a specific CIDR ignores X-Forwarded-For from peers outside it", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8")
		router := newRouter(t)

		assert.Equal(t, "9", remainingFor(router, "203.0.113.9:1234", "198.51.100.1"))
		assert.Equal(t, "8", remainingFor(router, "203.0.113.9:1234", "198.51.100.2"))
	})
}

func TestSetupRouter_NullInJSONArrayGetsAResponse(t *testing.T) {
	router := newRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(`{"ids":[null]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	assert.NotPanics(t, func() { router.ServeHTTP(w, req) })
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// captureLogs sends the global logger's output to a buffer for the test.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	logrus.SetOutput(buf)
	t.Cleanup(func() { logrus.SetOutput(os.Stderr) })
	return buf
}

func TestSetupRouter_WarnsAboutTheDefaultTrustedProxiesInProd(t *testing.T) {
	t.Run("Warns In Prod When TRUSTED_PROXIES Is Not Set", func(t *testing.T) {
		t.Setenv("STAGE", "prod")
		require.NoError(t, os.Unsetenv("TRUSTED_PROXIES"))
		logs := captureLogs(t)

		newRouter(t)

		assert.Contains(t, logs.String(), "TRUSTED_PROXIES")
	})

	t.Run("No Warning When TRUSTED_PROXIES Is Set", func(t *testing.T) {
		t.Setenv("STAGE", "prod")
		t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8")
		logs := captureLogs(t)

		newRouter(t)

		assert.NotContains(t, logs.String(), "TRUSTED_PROXIES")
	})

	t.Run("No Warning Outside Prod", func(t *testing.T) {
		t.Setenv("STAGE", "dev")
		require.NoError(t, os.Unsetenv("TRUSTED_PROXIES"))
		logs := captureLogs(t)

		newRouter(t)

		assert.NotContains(t, logs.String(), "TRUSTED_PROXIES")
	})
}
