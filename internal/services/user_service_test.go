package services_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	repo    *mocks.MockUserRepository
	service services.UserService
}

func (s *UserServiceTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.Require().NoError(err)
	s.Require().NotNil(db)

	err = db.AutoMigrate(&models.User{})
	s.Require().NoError(err)
	s.db = db
	s.repo = new(mocks.MockUserRepository)
	s.service = services.NewUserService(s.repo)

}

func (s *UserServiceTestSuite) TearDownTest() {
	s.repo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGetUser() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange
		expectedUser := &models.User{ID: 1, Email: "example@gmail.com", Password: "password123"}
		s.repo.On("GetByID", uint(1)).Return(expectedUser, nil).Once()

		// Act
		user, err := s.service.GetUser(1)

		// Assert
		s.NoError(err)
		s.Equal(expectedUser, user)
	})
	s.T().Run("Error", func(t *testing.T) {
		// Arrange
		s.repo.On("GetByID", uint(999)).Return(&models.User{}, errors.New("user not found")).Once()

		// Act
		user, err := s.service.GetUser(999)

		// Assert
		s.Error(err)
		s.Nil(user)
	})
}

func (s *UserServiceTestSuite) TestGetUserByEmail() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange
		expectedUser := &models.User{ID: 1, Email: "example@gmail.com", Password: "password123"}
		s.repo.On("FindByField", "email", "example@gmail.com").Return(expectedUser, nil).Once()

		// Act
		user, err := s.service.GetUserByEmail("example@gmail.com")

		// Assert
		s.NoError(err)
		s.Equal(expectedUser, user)
	})
	s.T().Run("Error", func(t *testing.T) {
		// Arrange
		s.repo.On("FindByField", "email", "notfound@gmail.com").Return(&models.User{}, errors.New("user not found")).Once()

		// Act
		user, err := s.service.GetUserByEmail("notfound@gmail.com")

		// Assert
		s.Error(err)
		s.Nil(user)
	})
}

func (s *UserServiceTestSuite) TestDeleteUser() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange
		s.repo.On("Delete", uint(1)).Return(nil).Once()

		// Act
		err := s.service.DeleteUser(1)

		// Assert
		s.NoError(err)
	})

	s.T().Run("Error", func(t *testing.T) {
		// Arrange
		s.repo.On("Delete", uint(999)).Return(errors.New("user not found")).Once()

		// Act
		err := s.service.DeleteUser(999)

		// Assert
		s.Error(err)
	})
}

func (s *UserServiceTestSuite) TestGetUserByToken() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange
		expectedUser := &models.User{ID: 1, Email: "email@example.com", Password: "password123"}
		s.repo.On("FindByField", "token", "valid_token").Return(expectedUser, nil).Once()

		// Act
		user, err := s.service.GetUserByToken("valid_token")

		// Assert
		s.NoError(err)
		s.Equal(expectedUser, user)
	})
	s.T().Run("Error", func(t *testing.T) {
		// Arrange
		s.repo.On("FindByField", "token", "invalid_token").Return(&models.User{}, errors.New("user not found")).Once()

		// Act
		user, err := s.service.GetUserByToken("invalid_token")

		// Assert
		s.Error(err)
		s.Nil(user)
	})
}

func (s *UserServiceTestSuite) TestGetProfile() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange
		expectedUser := &models.User{ID: 1, Email: "email@example.com", Password: "password123"}
		s.repo.On("GetByID", uint(1)).Return(expectedUser, nil).Once()

		// Act
		user, err := s.service.GetProfile(1)

		// Assert
		s.NoError(err)
		s.Equal(expectedUser, user)
	})
	s.T().Run("Error", func(t *testing.T) {
		// Arrange
		s.repo.On("GetByID", uint(999)).Return(&models.User{}, errors.New("profile not found")).Once()

		// Act
		user, err := s.service.GetProfile(999)

		// Assert
		s.Error(err)
		s.Nil(user)
	})
}

func (s *UserServiceTestSuite) TestUpdateProfile() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange
		user := &models.User{ID: 1, Email: "", Password: "newpassword123"}
		s.repo.On("Update", user).Return(nil).Once()

		// Act
		err := s.service.UpdateProfile(user)

		// Assert
		s.NoError(err)
	})
	s.T().Run("Error", func(t *testing.T) {
		// Arrange
		user := &models.User{ID: 999, Email: "", Password: "newpassword123"}
		s.repo.On("Update", user).Return(errors.New("update failed")).Once()

		// Act
		err := s.service.UpdateProfile(user)

		// Assert
		s.Error(err)
	})
}

func (s *UserServiceTestSuite) TestUpdateUser() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange
		user := &models.User{ID: 1, Email: "updated@example.com", Password: "newpassword123"}
		s.repo.On("Update", user).Return(nil).Once()

		// Act
		err := s.service.UpdateUser(user)

		// Assert
		s.NoError(err)
	})

	s.T().Run("Error", func(t *testing.T) {
		// Arrange
		user := &models.User{ID: 999, Email: "updated@example.com", Password: "newpassword123"}
		s.repo.On("Update", user).Return(errors.New("update failed")).Once()

		// Act
		err := s.service.UpdateUser(user)

		// Assert
		s.Error(err)
	})
}

func (s *UserServiceTestSuite) TestCreateUser() {
	user := &models.User{
		Email:    "newuser@example.com",
		Password: "password123",
		Name:     "New User",
		Gender:   1,
	}

	s.T().Run("Transaction Begin Error", func(t *testing.T) {
		// Arrange: Create a new closed database to simulate the error
		closedDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		s.Require().NoError(err)
		sqlDB, err := closedDB.DB()
		s.Require().NoError(err)
		err = sqlDB.Close()
		s.Require().NoError(err)
		s.repo.On("GetDB").Return(closedDB).Once()

		// Act
		err = s.service.CreateUser(user)

		// Assert
		s.Error(err)
		s.Contains(err.Error(), "sql: database is closed")
	})

	s.T().Run("Create Error with Working Transaction", func(t *testing.T) {
		// Arrange: Reset mocks for next test
		s.repo.Mock = mock.Mock{}
		s.repo.On("GetDB").Return(s.db).Once()
		s.repo.On("CreateWithTx", mock.AnythingOfType("*gorm.DB"), user).Return((*models.User)(nil), errors.New("create failed")).Once()

		// Act
		err := s.service.CreateUser(user)

		// Assert
		s.Error(err)
		s.Contains(err.Error(), "create failed")
	})
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
