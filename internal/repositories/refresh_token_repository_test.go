package repositories_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type RefreshTokenRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo *repositories.RefreshTokenRepository
}

func (s *RefreshTokenRepositoryTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)
	s.Require().NotNil(db)

	// Auto-migrate the models
	err = db.AutoMigrate(&models.RefreshToken{})
	s.Require().NoError(err)
	s.db = db
	s.repo = repositories.NewRefreshTokenRepository(db)
}

func (s *RefreshTokenRepositoryTestSuite) TearDownTest() {
	db, err := s.db.DB()
	if err == nil {
		_ = db.Close()
	}
}

func (s *RefreshTokenRepositoryTestSuite) TestCreate() {
	item := &models.RefreshToken{
		RefreshToken: "test_refresh_token",
		IpAddress:    "127.0.0.1",
		UsedCount:    0,
		ExpiredAt:    1710000000, // Example timestamp
		UserID:       1,
	}

	err := s.repo.Create(item)
	s.NoError(err, "Expected no error when creating a refresh token")
	s.NotEqual(uint(0), item.ID, "Expected refresh token ID to be set after creation")
}

func (s *RefreshTokenRepositoryTestSuite) TestCreate_Error_DuplicateToken() {
	// Assume Token is unique
	token1 := &models.RefreshToken{
		RefreshToken: "duplicate_token",
		UserID:       1,
	}

	token2 := &models.RefreshToken{
		RefreshToken: "duplicate_token", // same token as above
		UserID:       2,
	}

	// First insert should succeed
	err := s.repo.Create(token1)
	s.Require().NoError(err)

	// Second insert should fail due to unique constraint
	err = s.repo.Create(token2)
	s.Error(err, "Expected error due to duplicate token")
}
func (s *RefreshTokenRepositoryTestSuite) TestFirst() {

	items := []*models.RefreshToken{
		{
			RefreshToken: "test_refresh_token_1",
			IpAddress:    "",
			UsedCount:    0,
			ExpiredAt:    1710000000, // Example timestamp
			UserID:       1,
		},
		{
			RefreshToken: "test_refresh_token_2",
			IpAddress:    "127.0.0.1",
			UsedCount:    0,
			ExpiredAt:    1710000000, // Example timestamp
			UserID:       2,
		},
	}
	for _, item := range items {
		err := s.repo.Create(item)
		s.NoError(err, "Expected no error when creating a refresh token")
	}

	// Test finding the first token
	foundItem, err := s.repo.First(items[0].RefreshToken)
	s.NoError(err, "Expected no error when finding a refresh token")
	s.NotNil(foundItem, "Expected to find the refresh token")
	s.Equal(items[0].RefreshToken, foundItem.RefreshToken, "Expected found refresh token to match the created one")
	s.Equal(items[0].IpAddress, foundItem.IpAddress, "Expected found refresh token IP address to match the created one")
	s.Equal(items[0].UserID, foundItem.UserID, "Expected found refresh token UserID to match the created one")
	s.Equal(items[0].UsedCount, foundItem.UsedCount, "Expected found refresh token UsedCount to match the created one")
}

func (s *RefreshTokenRepositoryTestSuite) TestFirst_NotFound() {
	// Test finding a non-existent token
	foundItem, err := s.repo.First("non_existent_token")
	s.Error(err, "Expected error when finding a non-existent refresh token")
	s.Nil(foundItem, "Expected to not find a non-existent refresh token")
}

func (s *RefreshTokenRepositoryTestSuite) TestFindByTokenNotExpired() {
	now := time.Now().Unix() + int64(time.Minute)
	item := &models.RefreshToken{
		RefreshToken: "test_refresh_token",
		IpAddress:    "127.0.0.1",
		UsedCount:    0,
		ExpiredAt:    now,
		UserID:       1,
	}

	err := s.repo.Create(item)
	s.NoError(err, "Expected no error when creating a refresh token")

	s.NotEqual(uint(0), item.ID, "Expected refresh token ID to be set after creation")

	// Test finding the token
	foundItem, err := s.repo.FindByToken(item.RefreshToken)
	s.NoError(err, "Expected no error when finding a refresh token")
	s.NotNil(foundItem, "Expected to find the refresh token")
	s.Equal(item.RefreshToken, foundItem.RefreshToken, "Expected found refresh token to match the created one")
	s.Equal(item.IpAddress, foundItem.IpAddress, "Expected found refresh token IP address to match the created one")
	s.Equal(item.UserID, foundItem.UserID, "Expected found refresh token UserID to match the created one")
	s.Equal(item.UsedCount, foundItem.UsedCount, "Expected found refresh token UsedCount to match the created one")
	s.Equal(item.ExpiredAt, foundItem.ExpiredAt, "Expected found refresh token ExpiredAt to match the created one")

}

func (s *RefreshTokenRepositoryTestSuite) TestFindByTokenExpired() {
	now := time.Now().Unix() - int64(time.Minute)
	token := &models.RefreshToken{
		RefreshToken: "test_refresh_token_expired",
		IpAddress:    "127.0.0.1",
		UsedCount:    0,
		ExpiredAt:    now,
		UserID:       1,
	}

	err := s.repo.Create(token)
	s.NoError(err, "Expected no error when creating a refresh token")
	s.NotEqual(uint(0), token.ID, "Expected refresh token ID to be set after creation")

	// Test finding the expired token
	foundItem, err := s.repo.FindByToken(token.RefreshToken)
	s.Error(err, "Expected no error when finding a refresh token")
	s.Nil(foundItem, "Expected to not find the expired refresh token")
}

func (s *RefreshTokenRepositoryTestSuite) TestUpdate() {
	item := &models.RefreshToken{
		RefreshToken: "test_original_refresh_token",
		IpAddress:    "",
		UsedCount:    0,
		ExpiredAt:    1710000000, // Example timestamp
		UserID:       1,
	}
	// Create the token first
	err := s.repo.Create(item)
	s.NoError(err, "Expected no error when creating a refresh token")

	// Update the token
	item.IpAddress = "127.0.0.1"
	item.RefreshToken = "test_updated_refresh_token"
	item.UsedCount = 1
	item.ExpiredAt = time.Now().Unix() + int64(time.Hour)

	err = s.repo.Update(item)
	s.NoError(err, "Expected no error when updating a refresh token")

	// Verify the update
	foundItem, err := s.repo.FindByToken(item.RefreshToken)

	s.NoError(err, "Expected no error when finding the updated refresh token")
	s.NotNil(foundItem, "Expected to find the updated refresh token")
	s.Equal(item.RefreshToken, foundItem.RefreshToken, "Expected found refresh token to match the updated one")
	s.Equal(item.IpAddress, foundItem.IpAddress, "Expected found refresh token IP address to match the updated one")
	s.Equal(item.UserID, foundItem.UserID, "Expected found refresh token UserID to match the updated one")
	s.Equal(item.UsedCount, foundItem.UsedCount, "Expected found refresh token UsedCount to match the updated one")
	s.Equal(item.ExpiredAt, foundItem.ExpiredAt, "Expected found refresh token ExpiredAt to match the updated one")

}

func TestRefreshTokenRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenRepositoryTestSuite))
}
