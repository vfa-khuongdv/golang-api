package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockBcryptService struct {
	hashResult string
	hashErr    error
	checkValid bool
}

func (m *mockBcryptService) HashPassword(_ string) (string, error) {
	if m.hashResult != "" {
		return m.hashResult, m.hashErr
	}
	return "hashed-password", m.hashErr
}

func (m *mockBcryptService) CheckPasswordHash(_, _ string) bool {
	return m.checkValid
}

func (m *mockBcryptService) HashPasswordWithCost(password string, _ int) (string, error) {
	return m.HashPassword(password)
}

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

func (s *UserServiceTestSuite) TestForgotPassword() {
	s.T().Run("Success", func(t *testing.T) {
		// Arrange
		email := "test@example.com"
		user := &models.User{Email: email}

		s.repo.On("FindByField", "email", email).Return(user, nil).Once()
		s.repo.On("Update", user).Return(nil).Once()

		// Act
		result, err := s.service.ForgotPassword(&dto.ForgotPasswordInput{Email: email})

		// Assert
		s.NoError(err)
		s.NotNil(result)
		s.NotNil(result.Token)
		s.NotNil(result.ExpiredAt)
	})

	s.T().Run("UserNotFound", func(t *testing.T) {
		// Arrange
		email := "unknown@example.com"
		s.repo.On("FindByField", "email", email).Return((*models.User)(nil), gorm.ErrRecordNotFound).Once()

		// Act
		result, err := s.service.ForgotPassword(&dto.ForgotPasswordInput{Email: email})

		// Assert
		s.NoError(err)
		s.Nil(result)
	})

	s.T().Run("RepositoryQueryError", func(t *testing.T) {
		email := "error@example.com"
		s.repo.On("FindByField", "email", email).Return((*models.User)(nil), errors.New("db query failed")).Once()

		result, err := s.service.ForgotPassword(&dto.ForgotPasswordInput{Email: email})

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("UpdateFailure", func(t *testing.T) {
		email := "update-fail@example.com"
		user := &models.User{Email: email}

		s.repo.On("FindByField", "email", email).Return(user, nil).Once()
		s.repo.On("Update", user).Return(errors.New("update failed")).Once()

		result, err := s.service.ForgotPassword(&dto.ForgotPasswordInput{Email: email})

		s.Nil(result)
		s.Error(err)
	})
}

func (s *UserServiceTestSuite) TestResetPassword() {
	s.T().Run("TokenNotFound", func(t *testing.T) {
		input := &dto.ResetPasswordInput{Token: "invalid-token", NewPassword: "new-password"}
		s.repo.On("FindByField", "token", input.Token).Return(&models.User{}, errors.New("not found")).Once()

		user, err := s.service.ResetPassword(input)

		s.Nil(user)
		s.Error(err)
	})

	s.T().Run("TokenExpiredWhenExpiredAtNil", func(t *testing.T) {
		input := &dto.ResetPasswordInput{Token: "token-1", NewPassword: "new-password"}
		user := &models.User{ID: 1, Token: &input.Token, ExpiredAt: nil}
		s.repo.On("FindByField", "token", input.Token).Return(user, nil).Once()

		result, err := s.service.ResetPassword(input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("TokenExpiredByTimestamp", func(t *testing.T) {
		input := &dto.ResetPasswordInput{Token: "token-2", NewPassword: "new-password"}
		expiredAt := time.Now().Add(-1 * time.Minute).Unix()
		user := &models.User{ID: 1, Token: &input.Token, ExpiredAt: &expiredAt}
		s.repo.On("FindByField", "token", input.Token).Return(user, nil).Once()

		result, err := s.service.ResetPassword(input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("HashPasswordFailure", func(t *testing.T) {
		input := &dto.ResetPasswordInput{Token: "token-3", NewPassword: "new-password"}
		notExpired := time.Now().Add(10 * time.Minute).Unix()
		user := &models.User{ID: 1, Token: &input.Token, ExpiredAt: &notExpired}

		mockBcrypt := &mockBcryptService{hashErr: errors.New("hash failed"), checkValid: true}
		localService := services.NewUserService(s.repo, mockBcrypt)

		s.repo.On("FindByField", "token", input.Token).Return(user, nil).Once()

		result, err := localService.ResetPassword(input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("UpdateFailure", func(t *testing.T) {
		input := &dto.ResetPasswordInput{Token: "token-4", NewPassword: "new-password"}
		notExpired := time.Now().Add(10 * time.Minute).Unix()
		user := &models.User{ID: 1, Token: &input.Token, ExpiredAt: &notExpired}

		s.repo.On("FindByField", "token", input.Token).Return(user, nil).Once()
		s.repo.On("Update", user).Return(errors.New("update failed")).Once()

		result, err := s.service.ResetPassword(input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("Success", func(t *testing.T) {
		input := &dto.ResetPasswordInput{Token: "token-5", NewPassword: "new-password"}
		notExpired := time.Now().Add(10 * time.Minute).Unix()
		user := &models.User{ID: 1, Token: &input.Token, ExpiredAt: &notExpired}

		s.repo.On("FindByField", "token", input.Token).Return(user, nil).Once()
		s.repo.On("Update", user).Return(nil).Once()

		result, err := s.service.ResetPassword(input)

		s.NoError(err)
		s.NotNil(result)
		s.NotEqual(input.NewPassword, result.Password)
		s.Nil(result.Token)
		s.Nil(result.ExpiredAt)
	})
}

func (s *UserServiceTestSuite) TestChangePassword() {
	s.T().Run("UserNotFound", func(t *testing.T) {
		input := &dto.ChangePasswordInput{
			OldPassword:     "old-password",
			NewPassword:     "new-password",
			ConfirmPassword: "new-password",
		}
		s.repo.On("GetByID", uint(100)).Return(&models.User{}, errors.New("not found")).Once()

		result, err := s.service.ChangePassword(100, input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("OldPasswordMismatch", func(t *testing.T) {
		input := &dto.ChangePasswordInput{
			OldPassword:     "wrong-old",
			NewPassword:     "new-password",
			ConfirmPassword: "new-password",
		}
		hash, err := s.bcrypt.HashPassword("correct-old")
		s.Require().NoError(err)
		user := &models.User{ID: 1, Password: hash}
		s.repo.On("GetByID", uint(1)).Return(user, nil).Once()

		result, err := s.service.ChangePassword(1, input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("NewAndConfirmPasswordMismatch", func(t *testing.T) {
		input := &dto.ChangePasswordInput{
			OldPassword:     "old-password",
			NewPassword:     "new-password",
			ConfirmPassword: "different-password",
		}
		hash, err := s.bcrypt.HashPassword(input.OldPassword)
		s.Require().NoError(err)
		user := &models.User{ID: 1, Password: hash}
		s.repo.On("GetByID", uint(2)).Return(user, nil).Once()

		result, err := s.service.ChangePassword(2, input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("NewPasswordUnchanged", func(t *testing.T) {
		input := &dto.ChangePasswordInput{
			OldPassword:     "same-password",
			NewPassword:     "same-password",
			ConfirmPassword: "same-password",
		}
		hash, err := s.bcrypt.HashPassword(input.OldPassword)
		s.Require().NoError(err)
		user := &models.User{ID: 1, Password: hash}
		s.repo.On("GetByID", uint(3)).Return(user, nil).Once()

		result, err := s.service.ChangePassword(3, input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("HashPasswordFailure", func(t *testing.T) {
		input := &dto.ChangePasswordInput{
			OldPassword:     "old-password",
			NewPassword:     "new-password",
			ConfirmPassword: "new-password",
		}
		mockBcrypt := &mockBcryptService{hashErr: errors.New("hash failed"), checkValid: true}
		localService := services.NewUserService(s.repo, mockBcrypt)
		user := &models.User{ID: 1, Password: "existing-hash"}
		s.repo.On("GetByID", uint(4)).Return(user, nil).Once()

		result, err := localService.ChangePassword(4, input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("UpdateFailure", func(t *testing.T) {
		input := &dto.ChangePasswordInput{
			OldPassword:     "old-password",
			NewPassword:     "new-password",
			ConfirmPassword: "new-password",
		}
		hash, err := s.bcrypt.HashPassword(input.OldPassword)
		s.Require().NoError(err)
		user := &models.User{ID: 1, Password: hash}
		s.repo.On("GetByID", uint(5)).Return(user, nil).Once()
		s.repo.On("Update", user).Return(errors.New("update failed")).Once()

		result, err := s.service.ChangePassword(5, input)

		s.Nil(result)
		s.Error(err)
	})

	s.T().Run("Success", func(t *testing.T) {
		input := &dto.ChangePasswordInput{
			OldPassword:     "old-password",
			NewPassword:     "new-password",
			ConfirmPassword: "new-password",
		}
		hash, err := s.bcrypt.HashPassword(input.OldPassword)
		s.Require().NoError(err)
		user := &models.User{ID: 1, Password: hash}
		s.repo.On("GetByID", uint(6)).Return(user, nil).Once()
		s.repo.On("Update", user).Return(nil).Once()

		result, err := s.service.ChangePassword(6, input)

		s.NoError(err)
		s.NotNil(result)
		s.True(s.bcrypt.CheckPasswordHash(input.NewPassword, result.Password))
	})
}

func (s *UserServiceTestSuite) TestUpdateProfileErrors() {
	s.T().Run("UserNotFound", func(t *testing.T) {
		input := &dto.UpdateProfileInput{Name: utils.StringToPtr("John")}
		s.repo.On("GetByID", uint(77)).Return((*models.User)(nil), errors.New("not found")).Once()

		err := s.service.UpdateProfile(77, input)
		s.Error(err)
	})

	s.T().Run("InvalidBirthdayFormat", func(t *testing.T) {
		user := &models.User{ID: 1, Email: "a@b.com", Password: "hash"}
		input := &dto.UpdateProfileInput{Birthday: utils.StringToPtr("invalid-date")}
		s.repo.On("GetByID", uint(1)).Return(user, nil).Once()

		err := s.service.UpdateProfile(1, input)
		s.Error(err)
	})
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
