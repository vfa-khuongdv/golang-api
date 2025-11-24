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

func TestEmptyBodyMiddleware_RejectsEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middlewares.EmptyBodyMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	expectedJSON := fmt.Sprintf(`{
		"code": %d,
		"message": "Request body cannot be empty"
	}`, apperror.ErrEmptyData)

	assert.JSONEq(t, expectedJSON, resp.Body.String())
}

func TestEmptyBodyMiddleware_AllowsNonEmptyBody(t *testing.T) {
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

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `{"received": "{\"key\":\"value\"}"}`, resp.Body.String())
}

func TestEmptyBodyMiddleware_IgnoreGET(t *testing.T) {
	router := gin.New()
	router.Use(middlewares.EmptyBodyMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `{"message": "OK"}`, resp.Body.String())
}

func TestEmptyBodyMiddleware_RejectsWhitespaceBody(t *testing.T) {
	router := gin.New()
	router.Use(middlewares.EmptyBodyMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("   \n  \t  "))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "Request body cannot be empty")
}

func TestEmptyBodyMiddleware_AppliesToPUTandPATCH(t *testing.T) {
	router := gin.New()
	router.Use(middlewares.EmptyBodyMiddleware())
	router.PUT("/test-put", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})
	router.PATCH("/test-patch", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Test PUT with empty body
	reqPut := httptest.NewRequest(http.MethodPut, "/test-put", nil)
	respPut := httptest.NewRecorder()
	router.ServeHTTP(respPut, reqPut)
	assert.Equal(t, http.StatusBadRequest, respPut.Code)

	// Test PATCH with empty body
	reqPatch := httptest.NewRequest(http.MethodPatch, "/test-patch", nil)
	respPatch := httptest.NewRecorder()
	router.ServeHTTP(respPatch, reqPatch)
	assert.Equal(t, http.StatusBadRequest, respPatch.Code)
}

func TestEmptyBodyMiddleware_IgnoresDELETE(t *testing.T) {
	router := gin.New()
	router.Use(middlewares.EmptyBodyMiddleware())
	router.DELETE("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	req := httptest.NewRequest(http.MethodDelete, "/test", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `{"message": "OK"}`, resp.Body.String())
}
