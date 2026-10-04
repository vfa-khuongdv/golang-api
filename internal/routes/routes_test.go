package routes_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/routes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newRouter builds the router with a valid test config, after applying the
// given changes to it. By default every peer is trusted, as when
// TRUSTED_PROXIES is unset.
func newRouter(t *testing.T, changes ...func(*configs.Config)) *gin.Engine {
	t.Helper()
	cfg := &configs.Config{
		Server: configs.ServerConfig{
			GinMode:               gin.TestMode,
			Stage:                 "dev",
			TrustedProxies:        []string{"0.0.0.0/0"},
			TrustedProxiesDefault: true,
		},
		JWT:      configs.JWTConfig{Secret: "this-is-a-very-long-secret-key-for-routes-testing-32-chars"},
		Settings: configs.SettingsConfig{EncryptionKey: strings.Repeat("r", 40)},
		CORS:     configs.CORSConfig{AllowedOrigins: []string{"http://localhost:5173"}},
	}
	for _, change := range changes {
		change(cfg)
	}
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	return routes.SetupRouter(db, cfg)
}

// trustProxies sets the trusted proxies as an explicit TRUSTED_PROXIES would.
func trustProxies(proxies ...string) func(*configs.Config) {
	return func(cfg *configs.Config) {
		cfg.Server.TrustedProxies = proxies
		cfg.Server.TrustedProxiesDefault = false
	}
}

func inProd(cfg *configs.Config) { cfg.Server.Stage = "prod" }

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
		router := newRouter(t)

		// Behind an ALB the peer is the ALB and the client IP is in X-Forwarded-For.
		assert.Equal(t, "9", remainingFor(router, "10.0.0.5:1234", "203.0.113.1"))
		assert.Equal(t, "9", remainingFor(router, "10.0.0.5:1234", "203.0.113.2"))
		assert.Equal(t, "8", remainingFor(router, "10.0.0.5:1234", "203.0.113.1"))
	})

	t.Run("explicitly empty trusts nothing so all clients behind a proxy share one bucket", func(t *testing.T) {
		router := newRouter(t, trustProxies())

		assert.Equal(t, "9", remainingFor(router, "10.0.0.5:1234", "203.0.113.1"))
		assert.Equal(t, "8", remainingFor(router, "10.0.0.5:1234", "203.0.113.2"))
	})

	t.Run("a specific CIDR ignores X-Forwarded-For from peers outside it", func(t *testing.T) {
		router := newRouter(t, trustProxies("10.0.0.0/8"))

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
		logs := captureLogs(t)

		newRouter(t, inProd)

		assert.Contains(t, logs.String(), "TRUSTED_PROXIES")
	})

	t.Run("No Warning When TRUSTED_PROXIES Is Set", func(t *testing.T) {
		logs := captureLogs(t)

		newRouter(t, inProd, trustProxies("10.0.0.0/8"))

		assert.NotContains(t, logs.String(), "TRUSTED_PROXIES")
	})

	t.Run("No Warning Outside Prod", func(t *testing.T) {
		logs := captureLogs(t)

		newRouter(t)

		assert.NotContains(t, logs.String(), "TRUSTED_PROXIES")
	})
}

func TestSetupRouter_RejectsOversizedBodies(t *testing.T) {
	router := newRouter(t)
	body := `{"email":"` + strings.Repeat("a", 2<<20) + `@example.com","password":"x"}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestSetupRouter_SwaggerRoutesGoThroughTheMiddleware(t *testing.T) {
	router := newRouter(t)

	for _, path := range []string{"/swagger", "/api-docs", "/docs/swagger.json"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

		assert.NotEmpty(t, w.Header().Get("X-Request-ID"), path)
	}
}

// Every API route must be documented, so the docs cannot fall behind again.
func TestSwaggerDocumentsEveryAPIRoute(t *testing.T) {
	raw, err := os.ReadFile("../../docs/swagger.json")
	require.NoError(t, err)
	var doc struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(raw, &doc))

	for _, route := range newRouter(t, inProd).Routes() {
		path := regexp.MustCompile(`:(\w+)`).ReplaceAllString(route.Path, "{$1}")
		_, ok := doc.Paths[path][strings.ToLower(route.Method)]
		assert.True(t, ok, "%s %s is not in docs/swagger.json", route.Method, path)
	}
}

func TestSetupRouter_ReadyzPingsTheDatabase(t *testing.T) {
	w := httptest.NewRecorder()
	newRouter(t).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ready"}`, w.Body.String())
}
