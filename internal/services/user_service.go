package services

import (
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

type UserService interface {
	GetProfile(userID uint) (*models.User, error)
	UpdateProfile(userID uint, input *dto.UpdateProfileInput) error

	ForgotPassword(input *dto.ForgotPasswordInput) (*models.User, error)
	ResetPassword(input *dto.ResetPasswordInput) (*models.User, error)
	ChangePassword(userId uint, input *dto.ChangePasswordInput) (*models.User, error)
}

type userServiceImpl struct {
	repo          repositories.UserRepository
	bcryptService BcryptService
}

func NewUserService(repo repositories.UserRepository, bcryptService BcryptService) UserService {
	return &userServiceImpl{
		repo:          repo,
		bcryptService: bcryptService,
	}
}

func (service *userServiceImpl) ForgotPassword(input *dto.ForgotPasswordInput) (*models.User, error) {
	// 1. Check email exists
	user, err := service.repo.FindByField("email", input.Email)
	if err != nil {
		return nil, apperror.NewNotFoundError("Email not found")
	}

	// 2. Generate token and expiry
	token := utils.GenerateRandomString(32)
	expiredAt := time.Now().Add(1 * time.Hour).Unix()

	// 3. Update user with token and expiry
	user.Token = &token
	user.ExpiredAt = &expiredAt

	err = service.repo.Update(user)
	if err != nil {
		return nil, apperror.NewDBUpdateError(err.Error())
	}
	return user, nil
}

func (service *userServiceImpl) ResetPassword(input *dto.ResetPasswordInput) (*models.User, error) {
	// 1. Find user by token
	user, err := service.repo.FindByField("token", input.Token)
	if err != nil {
		return nil, apperror.NewNotFoundError("Invalid token")
	}

	// 2. Check if token is expired
	if user.ExpiredAt == nil || time.Now().Unix() > *user.ExpiredAt {
		return nil, apperror.NewTokenExpiredError("Token has expired")
	}

	// 3. Update newPassword
	newPassword, err := service.bcryptService.HashPassword(input.NewPassword)
	if err != nil {
		return nil, apperror.NewPasswordHashFailedError("Failed to hash password")
	}

	user.Password = newPassword
	user.Token = nil
	user.ExpiredAt = nil

	// 4. Save updated user
	err = service.repo.Update(user)
	if err != nil {
		return nil, apperror.NewDBUpdateError(err.Error())
	}
	return user, nil
}

func (service *userServiceImpl) ChangePassword(userId uint, input *dto.ChangePasswordInput) (*models.User, error) {
	// 1. Get user by ID
	user, err := service.repo.GetByID(userId)
	if err != nil {
		return nil, apperror.NewNotFoundError("User not found")
	}

	// 2. Check if old password is correct
	if isValid := service.bcryptService.CheckPasswordHash(input.OldPassword, user.Password); !isValid {
		return nil, apperror.NewInvalidPasswordError("Old password is incorrect")
	}

	// 3. Hash new password
	newPassword, err := service.bcryptService.HashPassword(input.NewPassword)
	if err != nil {
		return nil, apperror.NewPasswordHashFailedError("Failed to hash new password")
	}

	// 4. Check if new password and confirm password match
	if input.NewPassword != input.ConfirmPassword {
		return nil, apperror.NewPasswordMismatchError("New password and confirm password do not match")
	}

	// 5. Check old password and new password are not the same
	if input.OldPassword == input.NewPassword {
		return nil, apperror.NewPasswordUnchangedError("New password must be different from old password")
	}

	// 6. Update user password
	user.Password = newPassword
	err = service.repo.Update(user)
	if err != nil {
		return nil, apperror.NewDBUpdateError(err.Error())
	}
	return user, nil
}

func (service *userServiceImpl) GetProfile(userID uint) (*models.User, error) {
	user, err := service.repo.GetByID(userID)
	if err != nil {
		return nil, apperror.NewNotFoundError("User not found")
	}
	return user, nil
}

func (service *userServiceImpl) UpdateProfile(userID uint, input *dto.UpdateProfileInput) error {
	// 1. Get existing user
	user, err := service.repo.GetByID(userID)
	if err != nil {
		return apperror.NewNotFoundError("User not found")
	}

	// 2. Update fields if provided
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Address != nil {
		user.Address = input.Address
	}
	if input.Gender != nil {
		user.Gender = *input.Gender
	}

	if input.Birthday != nil {
		birthdayDate, err := utils.ParseDateStringYYYYMMDD(*input.Birthday)
		if err != nil {
			return err
		}
		user.Birthday = birthdayDate
	}

	// 3. Save updated user
	err = service.repo.Update(user)
	if err != nil {
		return apperror.NewDBUpdateError(err.Error())
	}
	return nil
}
