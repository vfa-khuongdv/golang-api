package repositories

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Get retrieves a role by its ID from the database
// Parameters:
//   - id: The unique identifier of the role to retrieve
//
// Returns:
//   - *models.Role: Pointer to the retrieved role if found
//   - error: nil if successful, error message if failed
func (repo *RoleRepository) Get(id int64) (*models.Role, error) {
	var role models.Role
	if err := repo.db.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// Create inserts a new role record into the database
// Parameters:
//   - role: Pointer to the role model to be created
//
// Returns:
//   - error: nil if successful, error message if failed
func (repo *RoleRepository) Create(role *models.Role) error {
	if err := repo.db.Create(role).Error; err != nil {
		return err
	}
	return nil
}

// Update modifies an existing role record in the database
// Parameters:
//   - role: Pointer to the role model containing updated data
//
// Returns:
//   - error: nil if successful, error message if failed
func (repo *RoleRepository) Update(role *models.Role) error {
	if err := repo.db.Save(role).Error; err != nil {
		return err
	}
	return nil
}

// Delete removes a role record from the database
// Parameters:
//   - role: Pointer to the role model to be deleted
//
// Returns:
//   - error: nil if successful, error message if failed
func (repo *RoleRepository) Delete(role *models.Role) error {
	if err := repo.db.Delete(role).Error; err != nil {
		return err
	}
	return nil
}