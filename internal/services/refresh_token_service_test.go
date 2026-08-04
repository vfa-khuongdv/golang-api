package services_test

import (
	"context"
	"testing"
	"time"

	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
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
		var persisted *models.RefreshToken
		s.repo.On("Create", mock.Anything, mock.MatchedBy(func(token *models.RefreshToken) bool {
			persisted = token
			return token.UserID == user.ID && token.IpAddress == ipAddress
		})).Return(nil)

		result, err := s.refreshTokenService.Create(context.Background(), user, ipAddress)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Token, 60)
		assert.Greater(t, result.ExpiresAt, int64(0))

		// The returned token must match the one persisted in the DB
		if s.NotNil(persisted) {
			assert.Equal(t, persisted.RefreshToken, result.Token)
			assert.Equal(t, persisted.ExpiredAt, result.ExpiresAt)
		}

		s.repo.AssertExpectations(t)
	})

	s.T().Run("Error", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
		svc := services.NewRefreshTokenService(mockRepo)

		_, err := svc.Create(context.Background(), user, ipAddress)
		assert.Error(t, err)
		var appErr *apperror.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, apperror.ErrDBInsert, appErr.Code)
		}
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

		var stored models.RefreshToken
		assert.NoError(t, db.Where("refresh_token = ?", result.Token.Token).First(&stored).Error)
		assert.Equal(t, uint(1), stored.UserID)
		assert.Equal(t, "127.0.0.2", stored.IpAddress)
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

	s.T().Run("BeginTxError", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		assert.NoError(t, err)
		err = db.AutoMigrate(&models.RefreshToken{})
		assert.NoError(t, err)

		repo := repositories.NewRefreshTokenRepository(db)
		svc := services.NewRefreshTokenService(repo)

		orig := &models.RefreshToken{
			RefreshToken: "existing_token",
			IpAddress:    "",
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

	s.T().Run("ExpiredToken", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		assert.NoError(t, err)
		err = db.AutoMigrate(&models.RefreshToken{})
		assert.NoError(t, err)

		repo := repositories.NewRefreshTokenRepository(db)
		svc := services.NewRefreshTokenService(repo)

		orig := &models.RefreshToken{
			RefreshToken: "expired_token",
			IpAddress:    "",
			ExpiredAt:    time.Now().Add(-time.Hour).Unix(),
			UserID:       1,
		}
		assert.NoError(t, repo.Create(context.Background(), orig))

		result, err := svc.Update(context.Background(), "expired_token", "127.0.0.1")

		assert.Error(t, err)
		assert.Nil(t, result)
		var appErr *apperror.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, apperror.ErrNotFound, appErr.Code)
		}
	})

	s.T().Run("RollbackAfterFindError", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		realDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		tx := realDB.Session(&gorm.Session{})
		mockRepo.On("BeginTx", mock.Anything).Return(tx, nil)
		mockRepo.On("FindByTokenWithTx", mock.Anything, mock.Anything, "some_token").Return((*models.RefreshToken)(nil), errors.New("find error"))
		svc := services.NewRefreshTokenService(mockRepo)

		result, err := svc.Update(context.Background(), "some_token", "127.0.0.1")
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	s.T().Run("UpdateWithTxError", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		realDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		tx := realDB.Session(&gorm.Session{})
		mockRepo.On("BeginTx", mock.Anything).Return(tx, nil)
		mockRepo.On("FindByTokenWithTx", mock.Anything, mock.Anything, "existing_token").Return(&models.RefreshToken{
			RefreshToken: "old_token",
			IpAddress:    "",
			ExpiredAt:    time.Now().Add(time.Hour).Unix(),
			UserID:       1,
		}, nil)
		mockRepo.On("UpdateWithTx", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("update error"))
		svc := services.NewRefreshTokenService(mockRepo)

		result, err := svc.Update(context.Background(), "existing_token", "127.0.0.1")
		assert.Error(t, err)
		assert.Nil(t, result)
		var appErr *apperror.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, apperror.ErrDBUpdate, appErr.Code)
		}
		mockRepo.AssertExpectations(t)
	})

	s.T().Run("CommitError", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		realDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		tx := realDB.Session(&gorm.Session{})
		mockRepo.On("BeginTx", mock.Anything).Return(tx, nil)
		mockRepo.On("FindByTokenWithTx", mock.Anything, mock.Anything, "existing_token").Return(&models.RefreshToken{
			RefreshToken: "old_token",
			IpAddress:    "",
			ExpiredAt:    time.Now().Add(time.Hour).Unix(),
			UserID:       1,
		}, nil)
		mockRepo.On("UpdateWithTx", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := services.NewRefreshTokenService(mockRepo)

		result, err := svc.Update(context.Background(), "existing_token", "127.0.0.1")
		assert.Error(t, err)
		assert.Nil(t, result)
		var appErr *apperror.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, apperror.ErrDBUpdate, appErr.Code)
		}
		mockRepo.AssertExpectations(t)
	})
}

func (s *RefreshTokenServiceTestSuite) TestDeleteByUserID() {
	userID := uint(1)

	s.T().Run("Success", func(t *testing.T) {
		s.repo.On("DeleteByUserID", mock.Anything, userID).Return(nil)

		err := s.refreshTokenService.DeleteByUserID(context.Background(), userID)

		assert.NoError(t, err)
		s.repo.AssertExpectations(t)
	})

	s.T().Run("Error", func(t *testing.T) {
		s.SetupTest()
		s.repo.On("DeleteByUserID", mock.Anything, userID).Return(errors.New("delete error"))

		err := s.refreshTokenService.DeleteByUserID(context.Background(), userID)

		assert.Error(t, err)
		var appErr *apperror.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, apperror.ErrDBDelete, appErr.Code)
		}
		s.repo.AssertExpectations(t)
	})
}

func TestRefreshTokenServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenServiceTestSuite))
}
