package middlewares_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestBodySizeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	newRouter := func(readErr *error, called *bool) *gin.Engine {
		router := gin.New()
		router.Use(middlewares.BodySizeLimit(10))
		router.POST("/upload", func(c *gin.Context) {
			*called = true
			_, *readErr = io.ReadAll(c.Request.Body)
			c.Status(http.StatusOK)
		})
		return router
	}

	t.Run("A Body Within The Limit Passes", func(t *testing.T) {
		var readErr error
		var called bool
		w := httptest.NewRecorder()

		newRouter(&readErr, &called).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("0123456789")))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, readErr)
	})

	t.Run("A Declared Length Over The Limit Is Rejected Before The Handler", func(t *testing.T) {
		var readErr error
		var called bool
		w := httptest.NewRecorder()

		newRouter(&readErr, &called).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("0123456789A")))

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Contains(t, w.Body.String(), `"code":1007`)
		assert.False(t, called)
		assert.Equal(t, apperror.ErrPayloadTooLarge, 1007)
	})

	t.Run("A Request Without A Body Passes", func(t *testing.T) {
		var called bool
		req := httptest.NewRequest(http.MethodPost, "/upload", nil)
		req.Body = nil
		w := httptest.NewRecorder()

		assert.NotPanics(t, func() {
			router := gin.New()
			router.Use(middlewares.BodySizeLimit(10))
			router.POST("/upload", func(c *gin.Context) { called = true; c.Status(http.StatusOK) })
			router.ServeHTTP(w, req)
		})
		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, called)
	})

	t.Run("A Body Without A Declared Length Is Cut At The Limit", func(t *testing.T) {
		var readErr error
		var called bool
		req := httptest.NewRequest(http.MethodPost, "/upload", io.NopCloser(strings.NewReader("0123456789A")))
		req.ContentLength = -1 // chunked: the size is unknown up front

		newRouter(&readErr, &called).ServeHTTP(httptest.NewRecorder(), req)

		var maxErr *http.MaxBytesError
		assert.True(t, errors.As(readErr, &maxErr), "got %v", readErr)
	})
}
