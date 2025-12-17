package services

import (
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

type IUserService interface {
	GetUser(id uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUsers(page int, limit int) (*dto.Pagination[*models.User], error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	DeleteUser(id uint) error
	GetUserByToken(token string) (*models.User, error)
	GetProfile(id uint) (*models.User, error)
	UpdateProfile(user *models.User) error
}

type UserService struct {
	repo repositories.IUserRepository
}

func NewUserService(repo repositories.IUserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (service *UserService) GetUsers(page int, limit int) (*dto.Pagination[*models.User], error) {
	users, err := service.repo.GetUsers(page, limit)
	if err != nil {
		return nil, apperror.NewDBQueryError(err.Error())
	}
	return users, nil
}

// GetUser retrieves a user by their ID from the database.
// Parameters:
//   - id: The unique identifier of the user to retrieve
//
// Returns:
//   - *models.User: A pointer to the user record if found
//   - error: nil if successful, otherwise returns the error that occurred
//
// Example:
//
//	user, err := service.GetUser(1) // Gets user with ID 1
func (service *UserService) GetUser(id uint) (*models.User, error) {
	data, err := service.repo.GetByID(id)
	if err != nil {
		return nil, apperror.NewNotFoundError(err.Error())
	}
	return data, nil
}

// GetUserByEmail retrieves a user by their email address from the database.
// Parameters:
//   - email: The email address of the user to retrieve
//
// Returns:
//   - *models.User: A pointer to the user record if found
//   - *appError.AppError: nil if successful, otherwise returns the error that occurred
//
// Example:
//
//	user, err := service.GetUserByEmail("john@example.com")
func (service *UserService) GetUserByEmail(email string) (*models.User, error) {
	data, err := service.repo.FindByField("email", email)
	if err != nil {
		return nil, apperror.NewNotFoundError(err.Error())
	}
	return data, nil
}

// CreateUser creates a new user in the database and assigns roles to them.
// Parameters:
//   - user: Pointer to models.User containing the user information to create
//
// Returns:
//   - *error: nil if successful, otherwise returns the error that occurred
func (service *UserService) CreateUser(user *models.User) error {
	tx := service.repo.GetDB().Begin()
	if tx.Error != nil {
		return apperror.NewDBInsertError(tx.Error.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	_, err := service.repo.CreateWithTx(tx, user)
	if err != nil {
		tx.Rollback()
		return apperror.NewDBInsertError(err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return apperror.NewDBInsertError(err.Error())
	}

	return nil
}

// UpdateUser updates an existing user's information in the database.
// Parameters:
//   - user: Pointer to models.User containing the updated user information
//
// Returns:
//   - error: nil if successful, otherwise returns the error that occurred
//
// Example:
//
//	user := &models.User{
//	    ID: 1,
//	    Name: "Updated Name",
//	    Email: "updated@example.com",
//	}
//	err := service.UpdateUser(user)
func (service *UserService) UpdateUser(user *models.User) error {
	err := service.repo.Update(user)
	if err != nil {
		return apperror.NewDBUpdateError(err.Error())
	}
	return nil
}

// DeleteUser removes a user from the database by their ID.
// Parameters:
//   - id: The unique identifier of the user to delete
//
// Returns:
//   - *appError.AppError: nil if successful, otherwise returns the error that occurred
//
// Example:
//
//	err := service.DeleteUser(1) // Deletes user with ID 1
func (service *UserService) DeleteUser(id uint) error {
	err := service.repo.Delete(id)
	if err != nil {
		return apperror.NewDBDeleteError(err.Error())
	}
	return nil
}

// GetUserByToken retrieves a user by their authentication token from the database.
// Parameters:
//   - token: The authentication token string associated with the user
//
// Returns:
//   - *models.User: A pointer to the user record if found
//   - error: nil if successful, otherwise returns the error that occurred
//
// Example:
//
//	user, err := service.GetUserByToken("abc123token")
func (service *UserService) GetUserByToken(token string) (*models.User, error) {
	data, err := service.repo.FindByField("token", token)
	if err != nil {
		return nil, apperror.NewNotFoundError(err.Error())
	}
	return data, nil
}

// GetProfile retrieves a user's profile information by their ID from the database.
// Parameters:
//   - id: The unique identifier of the user whose profile to retrieve
//
// Returns:
//   - *models.User: A pointer to the user profile record if found
//   - error: nil if successful, otherwise returns the error that occurred
//
// Example:
//
//	profile, err := service.GetProfile(1) // Gets profile for user with ID 1
func (service *UserService) GetProfile(id uint) (*models.User, error) {
	data, err := service.repo.GetByID(id)
	if err != nil {
		return nil, apperror.NewNotFoundError(err.Error())
	}
	return data, nil
}

// UpdateProfile updates a user's profile information in the database.
// Parameters:
//   - user: Pointer to models.User containing the updated profile information
//
// Returns:
//   - error: nil if successful, otherwise returns the error that occurred
//
// Example:
//
//	err := service.UpdateProfile(user)
func (service *UserService) UpdateProfile(user *models.User) error {
	err := service.repo.Update(user)
	if err != nil {
		return apperror.NewDBUpdateError(err.Error())
	}
	return nil
}
