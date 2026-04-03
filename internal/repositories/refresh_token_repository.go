package repositories

import (
	"errors"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	Update(token *models.RefreshToken) error
	FindByToken(token string) (*models.RefreshToken, error)
	First(token string) (*models.RefreshToken, error)
	UpdateWithTx(token *models.RefreshToken, tx *gorm.DB) error
}

type refreshTokenRepositoryImpl struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new instance of RefreshTokenRepository
// Parameters:
//   - db: pointer to the gorm.DB instance for database operations
//
// Returns:
//   - RefreshTokenRepository: pointer to the newly created RefreshTokenRepository
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepositoryImpl{db: db}
}

// Create creates a new refresh token in the database
// Parameters:
//   - token: pointer to the RefreshToken model to be saved
//
// Returns:
//   - error: nil if successful, error otherwise
func (repo *refreshTokenRepositoryImpl) Create(token *models.RefreshToken) error {
	if err := repo.db.Create(token).Error; err != nil {
		logger.Errorf("DB error: failed to create refresh token: %v", err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to create refresh token", err)
	}
	return nil
}

// First retrieves the first refresh token from the database by its token value
// Parameters:
//   - token: string representing the refresh token to search for
//
// Returns:
//   - *models.RefreshToken: pointer to the found RefreshToken model, nil if not found
//   - error: nil if successful, error otherwise
func (repo *refreshTokenRepositoryImpl) First(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	if err := repo.db.Where("refresh_token = ?", token).First(&refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.ErrNotFound, 1001, "Refresh token not found")
		}
		logger.Errorf("DB error: failed to fetch refresh token: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch refresh token", err)
	}
	return &refreshToken, nil
}

// FindByToken retrieves a refresh token from the database by its token value
// Parameters:
//   - token: string representing the refresh token to search for
//
// Returns:
//   - *models.RefreshToken: pointer to the found RefreshToken model, nil if not found
//   - error: nil if successful, error otherwise
func (repo *refreshTokenRepositoryImpl) FindByToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	if err := repo.db.Where("refresh_token = ? and expired_at > ?", token, time.Now().Unix()).First(&refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.ErrNotFound, 1001, "Refresh token not found or expired")
		}
		logger.Errorf("DB error: failed to fetch refresh token: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch refresh token", err)
	}
	return &refreshToken, nil
}

// Update updates an existing refresh token in the database
// Parameters:
//   - token: pointer to the RefreshToken model to be updated
//
// Returns:
//   - error: nil if successful, error otherwise
func (repo *refreshTokenRepositoryImpl) Update(token *models.RefreshToken) error {
	if err := repo.db.Save(token).Error; err != nil {
		logger.Errorf("DB error: failed to update refresh token: %v", err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to update refresh token", err)
	}
	return nil
}

func (repo *refreshTokenRepositoryImpl) UpdateWithTx(token *models.RefreshToken, tx *gorm.DB) error {
	if err := tx.Save(token).Error; err != nil {
		logger.Errorf("DB error: failed to update refresh token with tx: %v", err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to update refresh token", err)
	}
	return nil
}
