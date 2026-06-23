package middlewares_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestEmptyBodyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects empty body on POST", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.EmptyBodyMiddleware())
		router.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "OK"})
		})

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		expectedJSON := fmt.Sprintf(`{
			"code": %d,
			"message": "Request body cannot be empty"
		}`, apperror.ErrEmptyData)
		assert.JSONEq(t, expectedJSON, resp.Body.String())
	})

	t.Run("allows non-empty body", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.EmptyBodyMiddleware())
		router.POST("/test", func(c *gin.Context) {
			body, _ := c.GetRawData()
			c.JSON(http.StatusOK, gin.H{"received": string(body)})
		})

		body := bytes.NewBufferString(`{"key":"value"}`)
		req := httptest.NewRequest(http.MethodPost, "/test", body)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.JSONEq(t, `{"received": "{\"key\":\"value\"}"}`, resp.Body.String())
	})

	t.Run("ignores GET requests", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.EmptyBodyMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "OK"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.JSONEq(t, `{"message": "OK"}`, resp.Body.String())
	})

	t.Run("rejects whitespace-only body", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.EmptyBodyMiddleware())
		router.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "OK"})
		})

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("   \n  \t  "))
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Request body cannot be empty")
	})

	t.Run("applies to PUT and PATCH", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.EmptyBodyMiddleware())
		router.PUT("/test-put", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "OK"})
		})
		router.PATCH("/test-patch", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "OK"})
		})

		// Act & Assert - PUT
		reqPut := httptest.NewRequest(http.MethodPut, "/test-put", nil)
		respPut := httptest.NewRecorder()
		router.ServeHTTP(respPut, reqPut)
		assert.Equal(t, http.StatusBadRequest, respPut.Code)

		// Act & Assert - PATCH
		reqPatch := httptest.NewRequest(http.MethodPatch, "/test-patch", nil)
		respPatch := httptest.NewRecorder()
		router.ServeHTTP(respPatch, reqPatch)
		assert.Equal(t, http.StatusBadRequest, respPatch.Code)
	})

	t.Run("ignores DELETE requests", func(t *testing.T) {
		// Arrange
		router := gin.New()
		router.Use(middlewares.EmptyBodyMiddleware())
		router.DELETE("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "OK"})
		})

		req := httptest.NewRequest(http.MethodDelete, "/test", nil)
		resp := httptest.NewRecorder()

		// Act
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.JSONEq(t, `{"message": "OK"}`, resp.Body.String())
	})
}
