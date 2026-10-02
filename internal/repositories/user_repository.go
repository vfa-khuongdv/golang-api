package repositories

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUsers(ctx context.Context, page int, limit int) (*dto.Pagination[*models.User], error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	// UpdateColumns writes only the given columns of user (zero values
	// included), so a stale copy cannot overwrite columns changed elsewhere.
	UpdateColumns(ctx context.Context, user *models.User, columns ...string) error
	// RecordFailedLogin atomically increments failed_attempts and sets
	// locked_until to lockUntil once the counter reaches maxAttempts.
	RecordFailedLogin(ctx context.Context, userID uint, maxAttempts int, lockUntil int64) error
	// ResetFailedLogins clears failed_attempts and locked_until.
	ResetFailedLogins(ctx context.Context, userID uint) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByResetToken(ctx context.Context, token string) (*models.User, error)
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (repo *userRepositoryImpl) GetUsers(ctx context.Context, page, limit int) (*dto.Pagination[*models.User], error) {
	var totalRows int64
	offset := (page - 1) * limit
	db := repo.db.WithContext(ctx)

	if err := db.Model(&models.User{}).Count(&totalRows).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to count users: %v", err)
		return nil, apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to count users", err)
	}

	var users []*models.User
	if err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&users).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to fetch users: %v", err)
		return nil, apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to fetch users", err)
	}

	pagination := &dto.Pagination[*models.User]{
		Page:       page,
		Limit:      limit,
		TotalItems: int(totalRows),
		TotalPages: utils.CalculateTotalPages(totalRows, limit),
		Data:       users,
	}
	return pagination, nil
}

func (repo *userRepositoryImpl) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := repo.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(http.StatusNotFound, apperror.ErrNotFound, "User not found")
		}
		logger.WithContext(ctx).Errorf("DB error: failed to fetch user by id %d: %v", id, err)
		return nil, apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to fetch user", err)
	}
	return &user, nil
}

func (repo *userRepositoryImpl) UpdateColumns(ctx context.Context, user *models.User, columns ...string) error {
	if err := repo.db.WithContext(ctx).Model(user).Select(append(slices.Clip(columns), "updated_at")).Updates(user).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to update user id %d: %v", user.ID, err)
		return apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to update user", err)
	}
	return nil
}

func (repo *userRepositoryImpl) RecordFailedLogin(ctx context.Context, userID uint, maxAttempts int, lockUntil int64) error {
	// A single UPDATE so concurrent failures are all counted and no other column
	// is overwritten. locked_until is assigned first: MySQL evaluates SET
	// assignments left to right, so the CASE must see the old failed_attempts.
	err := repo.db.WithContext(ctx).Exec(
		"UPDATE users SET locked_until = CASE WHEN failed_attempts + 1 >= ? THEN ? ELSE locked_until END, "+
			"failed_attempts = failed_attempts + 1, updated_at = ? WHERE id = ?",
		maxAttempts, lockUntil, time.Now(), userID,
	).Error
	if err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to record failed login for user id %d: %v", userID, err)
		return apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to update user", err)
	}
	return nil
}

func (repo *userRepositoryImpl) ResetFailedLogins(ctx context.Context, userID uint) error {
	err := repo.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]any{"failed_attempts": 0, "locked_until": nil}).Error
	if err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to reset failed logins for user id %d: %v", userID, err)
		return apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to update user", err)
	}
	return nil
}

func (repo *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return repo.first(ctx, "email", repo.db.WithContext(ctx).Where("email = ?", email))
}

func (repo *userRepositoryImpl) FindByResetToken(ctx context.Context, token string) (*models.User, error) {
	return repo.first(ctx, "reset_token", repo.db.WithContext(ctx).Where("reset_token = ?", token))
}

// first fetches the first user matched by query; field is only used for logging.
func (repo *userRepositoryImpl) first(ctx context.Context, field string, query *gorm.DB) (*models.User, error) {
	var user models.User
	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(http.StatusNotFound, apperror.ErrNotFound, "User not found")
		}
		logger.WithContext(ctx).Errorf("DB error: failed to fetch user by %s: %v", field, err)
		return nil, apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to fetch user", err)
	}
	return &user, nil
}
