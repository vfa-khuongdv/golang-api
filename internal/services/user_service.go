package services

import (
	"context"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type UserService interface {
	GetProfile(ctx context.Context, userID uint) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uint, input *dto.UpdateProfileInput) error

	ForgotPassword(ctx context.Context, input *dto.ForgotPasswordInput) error
	ResetPassword(ctx context.Context, input *dto.ResetPasswordInput) (*models.User, error)
	ChangePassword(ctx context.Context, userId uint, input *dto.ChangePasswordInput) (*models.User, error)
}

type userServiceImpl struct {
	repo          repositories.UserRepository
	bcryptService BcryptService
	mailerService MailerService
}

func NewUserService(repo repositories.UserRepository, bcryptService BcryptService, mailerService MailerService) UserService {
	return &userServiceImpl{
		repo:          repo,
		bcryptService: bcryptService,
		mailerService: mailerService,
	}
}

func (service *userServiceImpl) ForgotPassword(ctx context.Context, input *dto.ForgotPasswordInput) error {
	user, err := service.repo.FindByField(ctx, "email", input.Email)
	if err != nil {
		appErr, isAppErr := apperror.ToAppError(err)
		if isAppErr && appErr.Code == apperror.ErrUnauthorized {
			return nil
		}
		logger.Errorf("Forgot password failed for email %s: %v", input.Email, err)
		return apperror.NewDBQueryError("Failed to process forgot password request")
	}

	token := utils.GenerateRandomString(32)
	expiredAt := time.Now().Add(1 * time.Hour).Unix()

	user.Token = &token
	user.ExpiredAt = &expiredAt

	err = service.repo.Update(ctx, user)
	if err != nil {
		logger.Errorf("Failed to update user with reset token: %v", err)
		return apperror.NewDBUpdateError("Failed to save reset token")
	}

	if err := service.mailerService.SendMailForgotPassword(user); err != nil {
		return err
	}

	return nil
}

func (service *userServiceImpl) ResetPassword(ctx context.Context, input *dto.ResetPasswordInput) (*models.User, error) {
	user, err := service.repo.FindByField(ctx, "token", input.Token)
	if err != nil {
		return nil, apperror.NewNotFoundError("Invalid token")
	}

	if user.ExpiredAt == nil || time.Now().Unix() > *user.ExpiredAt {
		return nil, apperror.NewTokenExpiredError("Token has expired")
	}

	newPassword, err := service.bcryptService.HashPassword(input.NewPassword)
	if err != nil {
		return nil, apperror.NewPasswordHashFailedError("Failed to hash password")
	}

	user.Password = newPassword
	user.Token = nil
	user.ExpiredAt = nil

	err = service.repo.Update(ctx, user)
	if err != nil {
		logger.Errorf("Failed to update user password: %v", err)
		return nil, apperror.NewDBUpdateError("Failed to update password")
	}
	return user, nil
}

func (service *userServiceImpl) ChangePassword(ctx context.Context, userId uint, input *dto.ChangePasswordInput) (*models.User, error) {
	user, err := service.repo.GetByID(ctx, userId)
	if err != nil {
		return nil, apperror.NewNotFoundError("User not found")
	}

	if isValid := service.bcryptService.CheckPasswordHash(input.OldPassword, user.Password); !isValid {
		return nil, apperror.NewInvalidPasswordError("Old password is incorrect")
	}

	newPassword, err := service.bcryptService.HashPassword(input.NewPassword)
	if err != nil {
		return nil, apperror.NewPasswordHashFailedError("Failed to hash new password")
	}

	if input.NewPassword != input.ConfirmPassword {
		return nil, apperror.NewPasswordMismatchError("New password and confirm password do not match")
	}

	if input.OldPassword == input.NewPassword {
		return nil, apperror.NewPasswordUnchangedError("New password must be different from old password")
	}

	user.Password = newPassword
	err = service.repo.Update(ctx, user)
	if err != nil {
		logger.Errorf("Failed to update user password: %v", err)
		return nil, apperror.NewDBUpdateError("Failed to update password")
	}
	return user, nil
}

func (service *userServiceImpl) GetProfile(ctx context.Context, userID uint) (*models.User, error) {
	user, err := service.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.NewNotFoundError("User not found")
	}
	return user, nil
}

func (service *userServiceImpl) UpdateProfile(ctx context.Context, userID uint, input *dto.UpdateProfileInput) error {
	user, err := service.repo.GetByID(ctx, userID)
	if err != nil {
		return apperror.NewNotFoundError("User not found")
	}

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

	err = service.repo.Update(ctx, user)
	if err != nil {
		logger.Errorf("Failed to update user profile: %v", err)
		return apperror.NewDBUpdateError("Failed to update profile")
	}
	return nil
}
