package repositories

import (
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	Update(token *models.RefreshToken) error
	FindByToken(token string) (*models.RefreshToken, error)
	First(token string) (*models.RefreshToken, error)
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
	return repo.db.Create(token).Error
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
		return nil, err
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
		return nil, err
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
	return repo.db.Save(token).Error
}
