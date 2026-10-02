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
	repo                repositories.UserRepository
	mailerService       MailerService
	refreshTokenService RefreshTokenService
}

// NewUserService creates a UserService. refreshTokenService revokes the user's
// sessions after a password reset or change.
func NewUserService(repo repositories.UserRepository, mailerService MailerService, refreshTokenService RefreshTokenService) UserService {
	return &userServiceImpl{
		repo:                repo,
		mailerService:       mailerService,
		refreshTokenService: refreshTokenService,
	}
}

func (service *userServiceImpl) ForgotPassword(ctx context.Context, input *dto.ForgotPasswordInput) error {
	user, err := service.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		appErr, isAppErr := apperror.ToAppError(err)
		if isAppErr && appErr.Code == apperror.ErrNotFound {
			logger.WithEvent(ctx, logger.EventPasswordResetRequest).Warnf("Forgot password attempt for non-existent email: %s", utils.MaskWithPrefix(input.Email, 4))
			return nil
		}
		logger.WithEvent(ctx, logger.EventPasswordResetRequest).Errorf("Forgot password failed for email %s: %v", utils.MaskWithPrefix(input.Email, 4), err)
		return apperror.NewDBQueryError("Failed to process forgot password request")
	}

	// Generate raw token for email, store only the hash in DB
	rawToken := utils.GenerateRandomString(32)
	hashedToken := utils.HashToken(rawToken)
	expiredAt := time.Now().Add(1 * time.Hour).Unix()

	user.ResetToken = &hashedToken
	user.ResetExpiredAt = &expiredAt

	// Only the columns changed here, so a stale copy of the user never
	// overwrites a concurrent change (e.g. a new password).
	err = service.repo.UpdateColumns(ctx, user, "reset_token", "reset_expired_at")
	if err != nil {
		logger.WithEvent(ctx, logger.EventPasswordResetRequest).Errorf("Failed to update user with reset token: %v", err)
		return apperror.NewDBUpdateError("Failed to save reset token")
	}

	// Send raw token to user via email (never store raw token in DB). The mail
	// goes out in the background and a failure is only logged: answering at
	// once, and the same way as for an unknown email, keeps the endpoint from
	// revealing which emails are registered, by status code or by SMTP delay.
	userWithRawToken := *user
	userWithRawToken.ResetToken = &rawToken
	mailCtx := context.WithoutCancel(ctx) // keeps the request ID, outlives the request
	go func() {
		if err := service.mailerService.SendMailForgotPassword(mailCtx, &userWithRawToken); err != nil {
			logger.WithEvent(mailCtx, logger.EventPasswordResetRequest).Errorf("Failed to send forgot password email to %s: %v", utils.MaskWithPrefix(user.Email, 4), err)
		}
	}()

	return nil
}

func (service *userServiceImpl) ResetPassword(ctx context.Context, input *dto.ResetPasswordInput) (*models.User, error) {
	// Hash the input token to compare with stored hash
	hashedToken := utils.HashToken(input.Token)
	user, err := service.repo.FindByResetToken(ctx, hashedToken)
	if err != nil {
		return nil, apperror.NewNotFoundError("Invalid token")
	}

	if user.ResetExpiredAt == nil || time.Now().Unix() > *user.ResetExpiredAt {
		return nil, apperror.NewTokenExpiredError("Token has expired")
	}

	newPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return nil, apperror.NewPasswordHashFailedError("Failed to hash password")
	}

	user.Password = newPassword
	user.ResetToken = nil
	user.ResetExpiredAt = nil

	// Reset failed attempts on successful password reset
	user.FailedAttempts = 0
	user.LockedUntil = nil

	err = service.repo.UpdateColumns(ctx, user, "password", "reset_token", "reset_expired_at", "failed_attempts", "locked_until")
	if err != nil {
		logger.WithEvent(ctx, logger.EventPasswordReset).Errorf("Failed to update user password: %v", err)
		return nil, apperror.NewDBUpdateError("Failed to update password")
	}

	// Sign out every session, so whoever knew the old password loses access.
	if err := service.refreshTokenService.DeleteByUserID(ctx, user.ID); err != nil {
		return nil, err
	}
	return user, nil
}

func (service *userServiceImpl) ChangePassword(ctx context.Context, userId uint, input *dto.ChangePasswordInput) (*models.User, error) {
	user, err := service.repo.GetByID(ctx, userId)
	if err != nil {
		return nil, apperror.NewNotFoundError("User not found")
	}

	if isValid := utils.CheckPasswordHash(input.OldPassword, user.Password); !isValid {
		return nil, apperror.NewInvalidPasswordError("Old password is incorrect")
	}

	if input.NewPassword != input.ConfirmPassword {
		return nil, apperror.NewPasswordMismatchError("New password and confirm password do not match")
	}

	if input.OldPassword == input.NewPassword {
		return nil, apperror.NewPasswordUnchangedError("New password must be different from old password")
	}

	newPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return nil, apperror.NewPasswordHashFailedError("Failed to hash new password")
	}

	user.Password = newPassword
	err = service.repo.UpdateColumns(ctx, user, "password")
	if err != nil {
		logger.WithEvent(ctx, logger.EventPasswordChangeFailed).Errorf("Failed to update user password: %v", err)
		return nil, apperror.NewDBUpdateError("Failed to update password")
	}

	// Sign out every session, so whoever knew the old password loses access.
	if err := service.refreshTokenService.DeleteByUserID(ctx, user.ID); err != nil {
		return nil, err
	}
	return user, nil
}

func (service *userServiceImpl) GetProfile(ctx context.Context, userID uint) (*models.User, error) {
	user, err := service.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.NewNotFoundError("User not found")
	}
	logger.WithEvent(ctx, logger.EventProfileGet).Infof("Retrieved profile for user ID %d", userID)
	return user, nil
}

func (service *userServiceImpl) UpdateProfile(ctx context.Context, userID uint, input *dto.UpdateProfileInput) error {
	user, err := service.repo.GetByID(ctx, userID)
	if err != nil {
		return apperror.NewNotFoundError("User not found")
	}

	// Only the fields sent are written, so the update cannot undo a concurrent
	// change to another column (e.g. a password reset).
	var columns []string
	if input.Name != nil {
		user.Name = *input.Name
		columns = append(columns, "name")
	}
	if input.Address != nil {
		user.Address = input.Address
		columns = append(columns, "address")
	}
	if input.Gender != nil {
		user.Gender = models.Gender(*input.Gender)
		columns = append(columns, "gender")
	}

	if input.Birthday != nil {
		birthdayDate, err := utils.ParseDateStringYYYYMMDD(*input.Birthday)
		if err != nil {
			return err
		}
		user.Birthday = birthdayDate
		columns = append(columns, "birthday")
	}

	if len(columns) == 0 {
		return nil
	}
	err = service.repo.UpdateColumns(ctx, user, columns...)
	if err != nil {
		logger.WithEvent(ctx, logger.EventProfileUpdateFailed).Errorf("Failed to update user profile: %v", err)
		return apperror.NewDBUpdateError("Failed to update profile")
	}
	return nil
}
