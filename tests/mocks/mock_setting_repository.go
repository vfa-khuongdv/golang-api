package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockSettingRepository struct {
	mock.Mock
}

func (m *MockSettingRepository) GetValues(ctx context.Context, keys ...string) (map[string]string, error) {
	args := m.Called(ctx, keys)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}
