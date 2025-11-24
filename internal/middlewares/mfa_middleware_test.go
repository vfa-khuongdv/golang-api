package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMfaVerificationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "No UserID in context",
			setupContext: func(c *gin.Context) {
				// No UserID set
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid UserID"}`,
		},
		{
			name: "MFA Pending is true",
			setupContext: func(c *gin.Context) {
				c.Set("UserID", uint(1))
				c.Set("MFAPending", true)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"MFA verification required"}`,
		},
		{
			name: "MFA Pending is false",
			setupContext: func(c *gin.Context) {
				c.Set("UserID", uint(1))
				c.Set("MFAPending", false)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name: "MFA Pending not set",
			setupContext: func(c *gin.Context) {
				c.Set("UserID", uint(1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupContext(c)

			middleware := MfaVerificationMiddleware()
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

// MockMfaService is a mock implementation of the interface expected by MfaRequiredMiddleware
type MockMfaService struct {
	GetMfaStatusFunc func(userID uint) (bool, error)
}

func (m *MockMfaService) GetMfaStatus(userID uint) (bool, error) {
	return m.GetMfaStatusFunc(userID)
}

func TestMfaRequiredMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		mockBehavior   func(*MockMfaService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "No UserID in context",
			setupContext: func(c *gin.Context) {
				// No UserID set
			},
			mockBehavior: func(m *MockMfaService) {
				// Not called
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid UserID"}`,
		},
		{
			name: "Error checking MFA status",
			setupContext: func(c *gin.Context) {
				c.Set("UserID", uint(1))
			},
			mockBehavior: func(m *MockMfaService) {
				m.GetMfaStatusFunc = func(userID uint) (bool, error) {
					return false, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to check MFA status"}`,
		},
		{
			name: "MFA status check success",
			setupContext: func(c *gin.Context) {
				c.Set("UserID", uint(1))
			},
			mockBehavior: func(m *MockMfaService) {
				m.GetMfaStatusFunc = func(userID uint) (bool, error) {
					return true, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupContext(c)

			mockService := &MockMfaService{}
			tt.mockBehavior(mockService)

			middleware := MfaRequiredMiddleware(mockService)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}
