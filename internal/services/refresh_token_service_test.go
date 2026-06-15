package services_test

import (
	"context"
	"testing"

	originErrors "errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type RefreshTokenServiceTestSuite struct {
	suite.Suite
	repo                *mocks.MockRefreshTokenRepository
	refreshTokenService services.RefreshTokenService
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
		s.repo.On("Create", mock.Anything, mock.MatchedBy(func(token *models.RefreshToken) bool {
			return token.UserID == user.ID && token.IpAddress == ipAddress
		})).Return(nil)

		result, err := s.refreshTokenService.Create(context.Background(), user, ipAddress)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Token, 60)
		assert.Greater(t, result.ExpiresAt, int64(0))

		s.repo.AssertExpectations(t)
	})

	s.T().Run("Error", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(originErrors.New("database error"))
		svc := services.NewRefreshTokenService(mockRepo)

		_, err := svc.Create(context.Background(), user, ipAddress)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func (s *RefreshTokenServiceTestSuite) TestUpdate() {
	s.T().Run("Success", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		assert.NoError(t, err)
		err = db.AutoMigrate(&models.RefreshToken{})
		assert.NoError(t, err)

		repo := repositories.NewRefreshTokenRepository(db)
		svc := services.NewRefreshTokenService(repo)

		orig := &models.RefreshToken{
			RefreshToken: "existing_token",
			IpAddress:    "",
			UsedCount:    0,
			ExpiredAt:    9999999999,
			UserID:       1,
		}
		assert.NoError(t, repo.Create(context.Background(), orig))

		result, err := svc.Update(context.Background(), "existing_token", "127.0.0.2")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.UserId)
		assert.Len(t, result.Token.Token, 60)
		assert.Greater(t, result.Token.ExpiresAt, int64(0))
	})

	s.T().Run("TokenNotFound", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		assert.NoError(t, err)
		err = db.AutoMigrate(&models.RefreshToken{})
		assert.NoError(t, err)

		repo := repositories.NewRefreshTokenRepository(db)
		svc := services.NewRefreshTokenService(repo)

		result, err := svc.Update(context.Background(), "missing_token", "127.0.0.1")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	s.T().Run("UpdateError", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		assert.NoError(t, err)
		db = db.Debug()
		err = db.AutoMigrate(&models.RefreshToken{})
		assert.NoError(t, err)

		repo := repositories.NewRefreshTokenRepository(db)
		svc := services.NewRefreshTokenService(repo)

		orig := &models.RefreshToken{
			RefreshToken: "existing_token",
			IpAddress:    "",
			UsedCount:    0,
			ExpiredAt:    9999999999,
			UserID:       1,
		}
		assert.NoError(t, repo.Create(context.Background(), orig))

		// Force a tx error by closing the underlying connection
		sqlDB, errr := db.DB()
		assert.NoError(t, errr)
		_ = sqlDB.Close()

		result, err := svc.Update(context.Background(), "existing_token", "127.0.0.1")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRefreshTokenServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenServiceTestSuite))
}
