package services_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	repo    *mocks.MockUserRepository
	service services.UserService
	bcrypt  services.BcryptService
}

func (s *UserServiceTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.Require().NoError(err)
	s.Require().NotNil(db)

	err = db.AutoMigrate(&models.User{})
	s.Require().NoError(err)
	s.db = db
	s.repo = new(mocks.MockUserRepository)
	s.bcrypt = services.NewBcryptService()
	s.service = services.NewUserService(s.repo, s.bcrypt)

}

func (s *UserServiceTestSuite) TearDownTest() {
	s.repo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGetProfile() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange

		userID := uint(1)
		expectedUser := &models.User{ID: 1, Email: "email@example.com", Password: "password123"}
		s.repo.On("GetByID", userID).Return(expectedUser, nil).Once()

		// Act
		user, err := s.service.GetProfile(userID)

		// Assert
		s.NoError(err)
		s.Equal(expectedUser, user)
	})

	s.T().Run("Error", func(t *testing.T) {
		// Arrange
		userID := uint(999)
		s.repo.On("GetByID", userID).Return(&models.User{}, errors.New("profile not found")).Once()

		// Act
		user, err := s.service.GetProfile(userID)

		// Assert
		s.Error(err)
		s.Nil(user)
	})
}

func (s *UserServiceTestSuite) TestUpdateProfile() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange
		user := &models.User{ID: 1, Email: "", Password: "newpassword123"}
		userID := uint(1)
		input := dto.UpdateProfileInput{
			Name:     utils.StringToPtr("John Doe"),
			Birthday: utils.StringToPtr("2020-01-01"),
			Address:  utils.StringToPtr("123 Main St"),
			Gender:   utils.IntToPtr(int16(1)),
		}

		s.repo.On("GetByID", userID).Return(user, nil).Once()
		s.repo.On("Update", user).Return(nil).Once()

		// Act
		err := s.service.UpdateProfile(userID, &input)

		// Assert
		s.NoError(err)
	})
	s.T().Run("Error", func(t *testing.T) {
		// Arrange
		userID := uint(999)
		user := &models.User{ID: userID, Email: "", Password: "newpassword123"}
		input := &dto.UpdateProfileInput{
			Name: utils.StringToPtr("John Doe"),
		}

		s.repo.On("GetByID", userID).Return(user, nil).Once()
		s.repo.On("Update", user).Return(errors.New("update failed")).Once()

		// Act
		err := s.service.UpdateProfile(userID, input)

		// Assert
		s.Error(err)
	})
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
