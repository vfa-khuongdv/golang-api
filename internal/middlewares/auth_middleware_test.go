package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		authHeader         string
		setupMock          func(*mocks.MockJWTService)
		expectedStatusCode int
		expectedUserID     interface{}
		expectNext         bool
	}{
		{
			name:               "Missing Authorization header",
			authHeader:         "",
			setupMock:          func(m *mocks.MockJWTService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:               "Authorization header without Bearer prefix",
			authHeader:         "InvalidToken",
			setupMock:          func(m *mocks.MockJWTService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:               "Authorization header with wrong prefix",
			authHeader:         "Basic some-token",
			setupMock:          func(m *mocks.MockJWTService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:       "Valid token with successful validation",
			authHeader: "Bearer valid-token",
			setupMock: func(m *mocks.MockJWTService) {
				claims := &services.CustomClaims{
					ID: 123,
				}
				m.On("ValidateToken", "valid-token").Return(claims, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedUserID:     uint(123),
			expectNext:         true,
		},
		{
			name:       "Invalid token that fails validation",
			authHeader: "Bearer invalid-token",
			setupMock: func(m *mocks.MockJWTService) {
				m.On("ValidateToken", "invalid-token").Return((*services.CustomClaims)(nil), errors.New("invalid token"))
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
		{
			name:       "Empty token after Bearer prefix",
			authHeader: "Bearer ",
			setupMock: func(m *mocks.MockJWTService) {
				m.On("ValidateToken", "").Return((*services.CustomClaims)(nil), errors.New("empty token"))
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUserID:     nil,
			expectNext:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock JWT service
			mockJWTService := new(mocks.MockJWTService)
			tt.setupMock(mockJWTService)

			// Create a test router
			router := gin.New()

			// Variable to track if next was called
			nextCalled := false
			var capturedUserID interface{}

			// Add the auth middleware with a test endpoint using mock
			router.Use(func(c *gin.Context) {
				authHeader := c.GetHeader("Authorization")
				if authHeader == "" || !hasValidBearerPrefix(authHeader) {
					respondWithUnauthorizedError(c, "Authorization header required")
					return
				}

				tokenString := extractTokenFromHeader(authHeader)
				claims, err := mockJWTService.ValidateToken(tokenString)
				if err != nil {
					respondWithUnauthorizedError(c, "Unauthorized")
					return
				}

				c.Set("UserID", claims.ID)
				c.Next()
			})

			router.GET("/test", func(c *gin.Context) {
				nextCalled = true
				capturedUserID, _ = c.Get("UserID")
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create a test request
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert the response
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectNext, nextCalled)

			// If we expect the middleware to set UserID, check it
			if tt.expectedUserID != nil {
				assert.Equal(t, tt.expectedUserID, capturedUserID)
			}

			// Verify all expectations were met
			mockJWTService.AssertExpectations(t)
		})
	}
}

// TestAuthMiddleware_DirectCall tests the actual AuthMiddleware function directly
func TestAuthMiddleware_DirectCall(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		authHeader         string
		expectedStatusCode int
	}{
		{
			name:               "No authorization header",
			authHeader:         "",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "Invalid authorization header format",
			authHeader:         "InvalidFormat",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "Bearer with no token",
			authHeader:         "Bearer",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "Bearer with space but no token",
			authHeader:         "Bearer ",
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test router with the actual middleware
			router := gin.New()
			router.Use(AuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create a test request
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert the response
			assert.Equal(t, tt.expectedStatusCode, w.Code)
		})
	}
}

// TestAuthMiddleware_WithRealJWT tests with real JWT tokens to verify the complete flow
func TestAuthMiddleware_WithRealJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a real JWT service to generate tokens
	jwtService := services.NewJWTService()

	// Generate valid tokens with different scopes
	accessTokenResult, err := jwtService.GenerateAccessToken(123)
	assert.NoError(t, err)
	assert.NotNil(t, accessTokenResult)

	mfaTokenResult, err := jwtService.GenerateMfaToken(456)
	assert.NoError(t, err)
	assert.NotNil(t, mfaTokenResult)

	// Test with valid access token
	t.Run("Valid JWT access token", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthMiddleware())

		var capturedUserID interface{}
		router.GET("/test", func(c *gin.Context) {
			capturedUserID, _ = c.Get("UserID")
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+accessTokenResult.Token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uint(123), capturedUserID)
	})

	// Test with MFA token (should fail because scope is wrong)
	t.Run("Invalid JWT token - wrong scope (MFA token on access endpoint)", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+mfaTokenResult.Token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test with invalid token
	t.Run("Invalid JWT token", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestMfaMiddleware tests the MFA verification middleware
func TestMfaMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a real JWT service to generate tokens
	jwtService := services.NewJWTService()

	// Generate tokens with different scopes
	accessTokenResult, err := jwtService.GenerateAccessToken(123)
	assert.NoError(t, err)

	mfaTokenResult, err := jwtService.GenerateMfaToken(789)
	assert.NoError(t, err)

	// Test with valid MFA token
	t.Run("Valid MFA token", func(t *testing.T) {
		router := gin.New()
		router.Use(MfaMiddleware())

		var capturedUserID interface{}
		router.POST("/verify", func(c *gin.Context) {
			capturedUserID, _ = c.Get("UserID")
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("POST", "/verify", nil)
		req.Header.Set("Authorization", "Bearer "+mfaTokenResult.Token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uint(789), capturedUserID)
	})

	// Test with access token (should fail because scope is wrong)
	t.Run("Invalid token - wrong scope (access token on MFA endpoint)", func(t *testing.T) {
		router := gin.New()
		router.Use(MfaMiddleware())
		router.POST("/verify", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("POST", "/verify", nil)
		req.Header.Set("Authorization", "Bearer "+accessTokenResult.Token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test missing authorization header
	t.Run("Missing authorization header", func(t *testing.T) {
		router := gin.New()
		router.Use(MfaMiddleware())
		router.POST("/verify", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("POST", "/verify", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// Helper function to check if authorization header has valid Bearer prefix
func hasValidBearerPrefix(authHeader string) bool {
	return len(authHeader) >= 7 && authHeader[:7] == "Bearer "
}

// Helper function to extract token from authorization header
func extractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 {
		return authHeader[7:]
	}
	return ""
}

// Helper function to respond with unauthorized error
func respondWithUnauthorizedError(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    "UNAUTHORIZED",
		"message": message,
	})
}
