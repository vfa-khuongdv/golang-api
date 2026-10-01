package middlewares_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		userID     any
		mockSetup  func(*mocks.MockRoleService)
		wantStatus int
		wantNext   bool
	}{
		{
			name:   "user has the permission",
			userID: uint(1),
			mockSetup: func(m *mocks.MockRoleService) {
				m.On("HasPermission", mock.Anything, uint(1), "settings:read").Return(true, nil)
			},
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
		{
			name:   "user lacks the permission",
			userID: uint(1),
			mockSetup: func(m *mocks.MockRoleService) {
				m.On("HasPermission", mock.Anything, uint(1), "settings:read").Return(false, nil)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "no authenticated user",
			userID:     nil,
			mockSetup:  func(m *mocks.MockRoleService) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "permission lookup fails",
			userID: uint(1),
			mockSetup: func(m *mocks.MockRoleService) {
				m.On("HasPermission", mock.Anything, uint(1), "settings:read").Return(false, errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "lookup returns an AppError",
			userID: uint(1),
			mockSetup: func(m *mocks.MockRoleService) {
				m.On("HasPermission", mock.Anything, uint(1), "settings:read").Return(false, apperror.NewInternalServerError("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mocks.MockRoleService)
			tt.mockSetup(svc)
			nextCalled := false

			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != nil {
					c.Set("UserID", tt.userID)
				}
			})
			router.GET("/", middlewares.RequirePermission(svc, "settings:read"), func(c *gin.Context) {
				nextCalled = true
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantNext, nextCalled)
			svc.AssertExpectations(t)
		})
	}
}
