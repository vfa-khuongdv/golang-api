package handlers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/handlers"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

func newSettingsContext(method, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(method, "/api/v1/settings", bytes.NewBufferString(body))
	return c, w
}

func TestGetSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		svc := new(mocks.MockSettingService)
		svc.On("GetSettings", mock.Anything).Return(&dto.SettingsResponse{
			MailHost: "smtp.example.com", MailPort: 587, MailPasswordSet: true,
			MailFrom: "noreply@example.com", FrontendURL: "https://app.example.com",
		}, nil)
		c, w := newSettingsContext("GET", "")

		handlers.NewSettingHandler(svc).GetSettings(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"mail_host":"smtp.example.com","mail_port":587,"mail_username":"","mail_password_set":true,"mail_from":"noreply@example.com","frontend_url":"https://app.example.com"}`, w.Body.String())
	})

	t.Run("Service Error", func(t *testing.T) {
		svc := new(mocks.MockSettingService)
		svc.On("GetSettings", mock.Anything).Return(nil, apperror.NewInternalServerError("boom"))
		c, w := newSettingsContext("GET", "")

		handlers.NewSettingHandler(svc).GetSettings(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUpdateSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitValidator()

	t.Run("Success", func(t *testing.T) {
		svc := new(mocks.MockSettingService)
		host, port := "smtp.example.com", 587
		svc.On("UpdateSettings", mock.Anything, &dto.UpdateSettingsInput{MailHost: &host, MailPort: &port}).Return(nil)
		c, w := newSettingsContext("PUT", `{"mail_host":"smtp.example.com","mail_port":587}`)

		handlers.NewSettingHandler(svc).UpdateSettings(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Update settings successfully"}`, w.Body.String())
		svc.AssertExpectations(t)
	})

	invalid := map[string]string{
		"empty host":       `{"mail_host":""}`,
		"blank host":       `{"mail_host":"   "}`,
		"port zero":        `{"mail_port":0}`,
		"port too large":   `{"mail_port":70000}`,
		"invalid from":     `{"mail_from":"not-an-email"}`,
		"empty from":       `{"mail_from":""}`,
		"frontend not url": `{"frontend_url":"not-a-url"}`,
		"frontend js url":  `{"frontend_url":"javascript:alert(1)"}`,
		"frontend empty":   `{"frontend_url":""}`,
		"malformed json":   `{`,
		"port wrong type":  `{"mail_port":"abc"}`,
		"empty body":       ``,
	}
	for name, body := range invalid {
		t.Run("Validation Error - "+name, func(t *testing.T) {
			svc := new(mocks.MockSettingService)
			c, w := newSettingsContext("PUT", body)

			handlers.NewSettingHandler(svc).UpdateSettings(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			svc.AssertNotCalled(t, "UpdateSettings", mock.Anything, mock.Anything)
		})
	}

	t.Run("Service Error", func(t *testing.T) {
		svc := new(mocks.MockSettingService)
		svc.On("UpdateSettings", mock.Anything, mock.Anything).Return(errors.New("db down"))
		c, w := newSettingsContext("PUT", `{"mail_host":"smtp.example.com"}`)

		handlers.NewSettingHandler(svc).UpdateSettings(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
