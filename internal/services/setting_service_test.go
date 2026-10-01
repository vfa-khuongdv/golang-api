package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/constants"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

var settingKeys = []string{
	constants.SettingMailHost,
	constants.SettingMailPort,
	constants.SettingMailUsername,
	constants.SettingMailPassword,
	constants.SettingMailFrom,
	constants.SettingFrontendURL,
}

func ptr[T any](v T) *T { return &v }

func TestSettingService_GetSettings(t *testing.T) {
	key := strings.Repeat("k", 40)

	t.Run("Maps Settings And Never Returns The Password", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		repo.On("GetValues", mock.Anything, settingKeys).Return(map[string]string{
			constants.SettingMailHost:     "smtp.example.com",
			constants.SettingMailPort:     "587",
			constants.SettingMailUsername: "mailer",
			constants.SettingMailPassword: "enc:v1:secret",
			constants.SettingMailFrom:     "noreply@example.com",
			constants.SettingFrontendURL:  "https://app.example.com",
		}, nil)
		svc := services.NewSettingService(repo, key)

		got, err := svc.GetSettings(context.Background())

		require.NoError(t, err)
		assert.Equal(t, &dto.SettingsResponse{
			MailHost:        "smtp.example.com",
			MailPort:        587,
			MailUsername:    "mailer",
			MailPasswordSet: true,
			MailFrom:        "noreply@example.com",
			FrontendURL:     "https://app.example.com",
		}, got)
		repo.AssertExpectations(t)
	})

	t.Run("Empty Password Means Not Set", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		repo.On("GetValues", mock.Anything, settingKeys).Return(map[string]string{
			constants.SettingMailPassword: "",
		}, nil)
		svc := services.NewSettingService(repo, key)

		got, err := svc.GetSettings(context.Background())

		require.NoError(t, err)
		assert.False(t, got.MailPasswordSet)
		assert.Equal(t, 0, got.MailPort)
	})

	t.Run("Repository Error", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		repo.On("GetValues", mock.Anything, settingKeys).Return(nil, errors.New("db down"))
		svc := services.NewSettingService(repo, key)

		got, err := svc.GetSettings(context.Background())

		assert.Nil(t, got)
		assert.EqualError(t, err, "db down")
	})
}

func TestSettingService_UpdateSettings(t *testing.T) {
	key := strings.Repeat("k", 40)

	t.Run("Writes Only Provided Fields", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		repo.On("SetValues", mock.Anything, map[string]string{
			constants.SettingMailHost: "smtp.example.com",
			constants.SettingMailPort: "465",
		}).Return(nil)
		svc := services.NewSettingService(repo, key)

		err := svc.UpdateSettings(context.Background(), &dto.UpdateSettingsInput{
			MailHost: ptr("smtp.example.com"),
			MailPort: ptr(465),
		})

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Encrypts The Mail Password", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		var stored string
		repo.On("SetValues", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			stored = args.Get(1).(map[string]string)[constants.SettingMailPassword]
		}).Return(nil)
		svc := services.NewSettingService(repo, key)

		err := svc.UpdateSettings(context.Background(), &dto.UpdateSettingsInput{MailPassword: ptr("s3cret")})

		require.NoError(t, err)
		assert.NotEqual(t, "s3cret", stored)
		plain, err := utils.DecryptSecret(key, stored)
		require.NoError(t, err)
		assert.Equal(t, "s3cret", plain)
	})

	t.Run("Empty Password Clears It Without Encrypting", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		repo.On("SetValues", mock.Anything, map[string]string{constants.SettingMailPassword: ""}).Return(nil)
		svc := services.NewSettingService(repo, key)

		err := svc.UpdateSettings(context.Background(), &dto.UpdateSettingsInput{MailPassword: ptr("")})

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Trims Trailing Slash From Frontend URL", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		repo.On("SetValues", mock.Anything, map[string]string{
			constants.SettingFrontendURL: "https://app.example.com",
		}).Return(nil)
		svc := services.NewSettingService(repo, key)

		err := svc.UpdateSettings(context.Background(), &dto.UpdateSettingsInput{FrontendURL: ptr("https://app.example.com/")})

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Nothing To Update Skips The Repository", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		svc := services.NewSettingService(repo, key)

		err := svc.UpdateSettings(context.Background(), &dto.UpdateSettingsInput{})

		require.NoError(t, err)
		repo.AssertNotCalled(t, "SetValues", mock.Anything, mock.Anything)
	})

	t.Run("Encryption Failure Is An Internal Error", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		svc := services.NewSettingService(repo, "too-short")

		err := svc.UpdateSettings(context.Background(), &dto.UpdateSettingsInput{MailPassword: ptr("s3cret")})

		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 500, appErr.HttpStatusCode)
		repo.AssertNotCalled(t, "SetValues", mock.Anything, mock.Anything)
	})

	t.Run("Repository Error", func(t *testing.T) {
		repo := new(mocks.MockSettingRepository)
		repo.On("SetValues", mock.Anything, mock.Anything).Return(errors.New("db down"))
		svc := services.NewSettingService(repo, key)

		err := svc.UpdateSettings(context.Background(), &dto.UpdateSettingsInput{MailHost: ptr("h")})

		assert.EqualError(t, err, "db down")
	})
}
