package services

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

type IPermissionService interface {
	GetAll() ([]models.Permission, error)
}

type PermissionService struct {
	repo repositories.IPermissionRepository
}

func NewPermissionService(repo repositories.IPermissionRepository) *PermissionService {
	return &PermissionService{
		repo: repo,
	}
}

// GetAll retrieves all permissions from the repository
// Returns:
//   - *[]models.Permission: Pointer to slice of Permission models containing all permissions
//   - error: Error if any occurred during the operation
func (repo *PermissionService) GetAll() ([]models.Permission, error) {
	permission, err := repo.repo.GetAll()
	if err != nil {
		return nil, apperror.NewInternalError(err.Error())
	}
	return permission, nil
}
