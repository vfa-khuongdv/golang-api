package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
)

type MockSettingService struct {
	mock.Mock
}

func (m *MockSettingService) GetSettings(ctx context.Context) (*dto.SettingsResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SettingsResponse), args.Error(1)
}

func (m *MockSettingService) UpdateSettings(ctx context.Context, input *dto.UpdateSettingsInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
