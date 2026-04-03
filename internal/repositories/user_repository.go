package repositories

import (
	"errors"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetAll() ([]*models.User, error)
	GetByID(id uint) (*models.User, error)
	Create(user *models.User) (*models.User, error)
	CreateWithTx(tx *gorm.DB, user *models.User) (*models.User, error)
	Update(user *models.User) error
	Delete(userId uint) error
	FindByField(field string, value string) (*models.User, error)
	GetUsers(page int, limit int) (*dto.Pagination[*models.User], error)
	GetDB() *gorm.DB
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

// GetUsers retrieves a paginated list of users from the database
// Parameters:
//   - page: The page number to retrieve (default is 1)
//   - limit: The number of users per page (default is 10)
//
// Returns:
//   - *dto.Pagination: A pointer to the pagination object containing user data
//   - error: nil if successful, otherwise returns the error that occurred
//
// Example:
//   - users, err := repo.GetUsers(1, 50) // Gets the first page of users
func (repo *userRepositoryImpl) GetUsers(page, limit int) (*dto.Pagination[*models.User], error) {
	var totalRows int64
	offset := (page - 1) * limit

	if err := repo.db.Model(&models.User{}).Count(&totalRows).Error; err != nil {
		logger.Errorf("DB error: failed to count users: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to count users", err)
	}

	var users []*models.User
	if err := repo.db.Offset(offset).Limit(limit).Order("id DESC").Find(&users).Error; err != nil {
		logger.Errorf("DB error: failed to fetch users: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch users", err)
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

// GetAll retrieves all users from the database
// Parameters:
//   - None
//
// Returns:
//   - []*models.User: Slice containing all User models in the database
//   - error: Error if there was a database error, nil on success
func (repo *userRepositoryImpl) GetAll() ([]*models.User, error) {
	var users []*models.User
	if err := repo.db.Find(&users).Error; err != nil {
		logger.Errorf("DB error: failed to fetch users: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch users", err)
	}
	return users, nil
}

// GetByID retrieves a user from the database by their ID
// Parameters:
//   - id: The unique identifier of the user to retrieve
//
// Returns:
//   - *models.User: Pointer to the retrieved User model
//   - error: Error if the user is not found or if there was a database error
func (repo *userRepositoryImpl) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := repo.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.ErrNotFound, 1001, "User not found")
		}
		logger.Errorf("DB error: failed to fetch user by id %d: %v", id, err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch user", err)
	}
	return &user, nil
}

// Create creates a new user in the database
// Parameters:
//   - user: Pointer to the User model to be created
//
// Returns:
//   - *models.User: Pointer to the created User model with assigned ID
//   - error: Error if there was a problem creating the user, nil on success
func (repo *userRepositoryImpl) Create(user *models.User) (*models.User, error) {
	if err := repo.db.Create(user).Error; err != nil {
		logger.Errorf("DB error: failed to create user: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to create user", err)
	}
	return user, nil
}

// CreateWithTx creates a new user in the database within a transaction
// Parameters:
//   - tx: Pointer to the gorm.DB transaction
//   - user: Pointer to the User model to be created
//
// Returns:
//   - *models.User: Pointer to the created User model with assigned ID
//   - error: Error if there was a problem creating the user, nil on success
func (repo *userRepositoryImpl) CreateWithTx(tx *gorm.DB, user *models.User) (*models.User, error) {
	if err := tx.Create(user).Error; err != nil {
		logger.Errorf("DB error: failed to create user with tx: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to create user", err)
	}
	return user, nil
}

// Update updates an existing user in the database
// Parameters:
//   - user: Pointer to the User model to be updated
//
// Returns:
//   - error: Error if there was a problem updating the user, nil on success
func (repo *userRepositoryImpl) Update(user *models.User) error {
	if err := repo.db.Save(user).Error; err != nil {
		logger.Errorf("DB error: failed to update user id %d: %v", user.ID, err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to update user", err)
	}
	return nil
}

// Delete removes a user from the database
// Parameters:
//   - id: userId to be deleted
//
// Returns:
//   - error: Error if there was a problem deleting the user, nil on success
func (repo *userRepositoryImpl) Delete(userId uint) error {
	var user models.User
	if err := repo.db.Delete(&user, userId).Error; err != nil {
		logger.Errorf("DB error: failed to delete user id %d: %v", userId, err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to delete user", err)
	}
	return nil
}

// FindByField retrieves a user from the database by a specified field and value
// Parameters:
//   - field: The field to search by (e.g., "name", "email", "token")
//   - value: The value to match against the specified field
//
// Returns:
//   - *models.User: Pointer to the retrieved User model if found
//   - error: Error if user not found or if there was a database error
func (repo *userRepositoryImpl) FindByField(field string, value string) (*models.User, error) {
	switch field {
	case "name":
		field = "name"
	case "email":
		field = "email"
	case "token":
		field = "token"
	default:
		return nil, apperror.New(apperror.ErrBadRequest, 1002, "Invalid field")
	}

	var user models.User
	if err := repo.db.Where(field+" = ?", value).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.ErrUnauthorized, 1003, "User not found")
		}
		logger.Errorf("DB error: failed to fetch user by field %s: %v", field, err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch user", err)
	}
	return &user, nil
}

// GetDB returns the database connection
// Used for transaction handling and other direct database operations
//
// Returns:
//   - *gorm.DB: The database connection
func (repo *userRepositoryImpl) GetDB() *gorm.DB {
	return repo.db
}
