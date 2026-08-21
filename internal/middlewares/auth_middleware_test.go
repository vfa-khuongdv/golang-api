package middlewares_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		authHeader         string
		mockSetup          func(*mocks.MockJWTService)
		expectedStatusCode int
		expectedUserID     interface{}
		expectNext         bool
	}{
		{
			name:               "missing Authorization header",
			authHeader:         "",
			mockSetup:          func(m *mocks.MockJWTService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:               "Authorization header without Bearer prefix",
			authHeader:         "InvalidToken",
			mockSetup:          func(m *mocks.MockJWTService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:               "Authorization header with wrong prefix",
			authHeader:         "Basic some-token",
			mockSetup:          func(m *mocks.MockJWTService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:               "lowercase bearer scheme is rejected",
			authHeader:         "bearer valid-token",
			mockSetup:          func(m *mocks.MockJWTService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:       "valid token with successful validation",
			authHeader: "Bearer valid-token",
			mockSetup: func(m *mocks.MockJWTService) {
				claims := &services.CustomClaims{ID: 123, Scope: services.TokenScopeAccess}
				m.On("ValidateTokenWithScope", "valid-token", services.TokenScopeAccess).Return(claims, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedUserID:     uint(123),
			expectNext:         true,
		},
		{
			name:       "invalid token that fails validation",
			authHeader: "Bearer invalid-token",
			mockSetup: func(m *mocks.MockJWTService) {
				m.On("ValidateTokenWithScope", "invalid-token", services.TokenScopeAccess).Return((*services.CustomClaims)(nil), errors.New("invalid token"))
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:       "empty token after Bearer prefix",
			authHeader: "Bearer ",
			mockSetup: func(m *mocks.MockJWTService) {
				m.On("ValidateTokenWithScope", "", services.TokenScopeAccess).Return((*services.CustomClaims)(nil), errors.New("empty token"))
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:       "token with wrong scope",
			authHeader: "Bearer wrong-scope-token",
			mockSetup: func(m *mocks.MockJWTService) {
				m.On("ValidateTokenWithScope", "wrong-scope-token", services.TokenScopeAccess).Return((*services.CustomClaims)(nil), errors.New("invalid scope"))
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:       "valid token with zero user id",
			authHeader: "Bearer zero-token",
			mockSetup: func(m *mocks.MockJWTService) {
				claims := &services.CustomClaims{ID: 0, Scope: services.TokenScopeAccess}
				m.On("ValidateTokenWithScope", "zero-token", services.TokenScopeAccess).Return(claims, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedUserID:     uint(0),
			expectNext:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockJWTService := new(mocks.MockJWTService)
			tt.mockSetup(mockJWTService)

			router := gin.New()
			nextCalled := false
			var capturedUserID interface{}

			router.Use(middlewares.AuthMiddleware(mockJWTService))
			router.GET("/test", func(c *gin.Context) {
				nextCalled = true
				capturedUserID, _ = c.Get("UserID")
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectNext, nextCalled)
			if tt.expectedUserID != nil {
				assert.Equal(t, tt.expectedUserID, capturedUserID)
			} else {
				assert.Nil(t, capturedUserID)
			}
			mockJWTService.AssertExpectations(t)
		})
	}
}

func TestAuthMiddleware_DirectCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_KEY", "this-is-a-very-long-secret-key-for-middleware-testing-32-chars")

	jwtService, err := services.NewJWTService()
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	tests := []struct {
		name               string
		authHeader         string
		expectedStatusCode int
	}{
		{name: "no authorization header", authHeader: "", expectedStatusCode: http.StatusUnauthorized},
		{name: "invalid authorization header format", authHeader: "InvalidFormat", expectedStatusCode: http.StatusUnauthorized},
		{name: "Bearer with no token", authHeader: "Bearer", expectedStatusCode: http.StatusUnauthorized},
		{name: "Bearer with space but no token", authHeader: "Bearer ", expectedStatusCode: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middlewares.AuthMiddleware(jwtService))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
		})
	}
}

func TestAuthMiddleware_WithRealJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_KEY", "this-is-a-very-long-secret-key-for-middleware-testing-32-chars")

	jwtService, err := services.NewJWTService()
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	accessTokenResult, err := jwtService.GenerateAccessToken(123)
	assert.NoError(t, err)
	assert.NotNil(t, accessTokenResult)

	t.Run("valid JWT access token", func(t *testing.T) {
		router := gin.New()
		router.Use(middlewares.AuthMiddleware(jwtService))

		var capturedUserID interface{}
		router.GET("/test", func(c *gin.Context) {
			capturedUserID, _ = c.Get("UserID")
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+accessTokenResult.Token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uint(123), capturedUserID)
	})
}

func init() {
	gin.SetMode(gin.TestMode)
}
