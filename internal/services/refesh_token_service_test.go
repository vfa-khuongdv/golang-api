package services_test

import (
	"testing"

	originErrors "errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

type RefreshTokenServiceTestSuite struct {
	suite.Suite
	repo                *mocks.MockRefreshTokenRepository
	refreshTokenService *services.RefreshTokenService
}

func (s *RefreshTokenServiceTestSuite) SetupTest() {
	s.repo = new(mocks.MockRefreshTokenRepository)
	s.refreshTokenService = services.NewRefreshTokenService(s.repo)
}

func (s *RefreshTokenServiceTestSuite) TestCreate() {
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
	}
	ipAddress := "127.0.0.1"

	s.T().Run("Success", func(t *testing.T) {
		s.repo.On("Create", mock.MatchedBy(func(token *models.RefreshToken) bool {
			return token.UserID == user.ID && token.IpAddress == ipAddress
		})).Return(nil)

		result, err := s.refreshTokenService.Create(user, ipAddress)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Token, 60)
		assert.Greater(t, result.ExpiresAt, int64(0))

		s.repo.AssertExpectations(t)
	})

	s.T().Run("Error", func(t *testing.T) {
		s.repo = new(mocks.MockRefreshTokenRepository) // reset
		s.refreshTokenService = services.NewRefreshTokenService(s.repo)

		s.repo.On("Create", mock.Anything).Return(originErrors.New("database error"))
		_, err := s.refreshTokenService.Create(user, ipAddress)
		assert.Error(t, err)
		s.repo.AssertExpectations(t)
	})
}

func (s *RefreshTokenServiceTestSuite) TestUpdate() {
	originalToken := &models.RefreshToken{
		RefreshToken: "existing_token",
		IpAddress:    "",
		UsedCount:    0,
		ExpiredAt:    0,
		UserID:       1,
	}

	s.T().Run("Success", func(t *testing.T) {
		s.repo.On("FindByToken", "existing_token").Return(originalToken, nil).Once()
		s.repo.On("Update", mock.AnythingOfType("*models.RefreshToken")).Return(nil).Once()

		result, err := s.refreshTokenService.Update("existing_token", "127.0.0.2")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, originalToken.UserID, result.UserId)
		assert.Len(t, result.Token.Token, 60)
		assert.Greater(t, result.Token.ExpiresAt, int64(0))

		s.repo.AssertExpectations(t)
	})

	s.T().Run("TokenNotFound", func(t *testing.T) {
		s.repo.On("FindByToken", "missing_token").Return((*models.RefreshToken)(nil), assert.AnError).Once()

		result, err := s.refreshTokenService.Update("missing_token", "127.0.0.1")

		assert.Error(t, err)
		assert.Nil(t, result)

		s.repo.AssertExpectations(t)
	})

	s.T().Run("Error", func(t *testing.T) {
		s.repo.On("FindByToken", "existing_token").Return(originalToken, nil).Once()
		s.repo.On("Update", mock.AnythingOfType("*models.RefreshToken")).Return(originErrors.New("Update item error")).Once()

		result, err := s.refreshTokenService.Update("existing_token", "127.0.0.1")

		assert.Error(t, err)
		assert.Nil(t, result)

		s.repo.AssertExpectations(t)
	})
}

func TestRefreshTokenServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenServiceTestSuite))
}
