package services

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

type ISettingService interface {
	GetSetting() ([]models.Setting, error)
	GetSettingByKey(key string) (*models.Setting, error)
	Update(setting *models.Setting) error
	Create(setting *models.Setting) error
}

type SettingService struct {
	repo repositories.ISettingRepository
}

func NewSettingService(repo repositories.ISettingRepository) *SettingService {
	return &SettingService{
		repo: repo,
	}
}

// GetSetting retrieves all settings from the repository
// Returns:
//   - *[]models.Setting: pointer to a slice of Setting models containing all settings
//   - error: any error encountered during the retrieval operation
func (service *SettingService) GetSetting() ([]models.Setting, error) {
	data, err := service.repo.GetAll()
	if err != nil {
		return nil, apperror.NewInternalError(err.Error())
	}
	return data, nil
}

// GetSettingByKey retrieves a specific setting from the repository by its key
// Parameters:
//   - key: string representing the unique identifier of the setting
//
// Returns:
//   - *models.Setting: pointer to the Setting model if found
//   - error: any error encountered during the retrieval operation
func (service *SettingService) GetSettingByKey(key string) (*models.Setting, error) {
	data, err := service.repo.GetByKey(key)
	if err != nil {
		return nil, apperror.NewNotFoundError(err.Error())
	}
	return data, nil
}

// Update updates a single setting in the repository
// Parameters:
//   - setting: pointer to the Setting model to be updated
//
// Returns:
//   - *models.Setting: pointer to the updated Setting model
//   - error: any error encountered during the update operation
func (service *SettingService) Update(setting *models.Setting) error {
	err := service.repo.Update(setting)
	if err != nil {
		return apperror.NewDBUpdateError(err.Error())
	}
	return nil
}

// Create creates a new setting in the repository
// Parameters:
//   - setting: pointer to the Setting model to be created
//
// Returns:
//   - *models.Setting: pointer to the created Setting model
//   - *appError.AppError: any error encountered during the creation operation
func (service *SettingService) Create(setting *models.Setting) error {
	err := service.repo.Create(setting)
	if err != nil {
		return apperror.NewDBInsertError(err.Error())
	}
	return nil
}
