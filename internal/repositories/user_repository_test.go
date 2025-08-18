package repositories_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo *repositories.UserRepository
}

func (s *UserRepositoryTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)
	s.Require().NotNil(db)

	// Auto-migrate the models
	err = db.AutoMigrate(
		&models.User{},
	)
	s.Require().NoError(err)
	s.db = db
	s.repo = repositories.NewUserRepository(db)
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	db, err := s.db.DB()
	if err == nil {
		_ = db.Close()
	}
}

func (s *UserRepositoryTestSuite) TestGetAll() {
	mockUsers := []*models.User{
		{Name: "User1", Email: "email1@example.com", Password: "password1", Gender: 1},
		{Name: "User2", Email: "email2@example.com", Password: "password2", Gender: 1},
	}
	for _, user := range mockUsers {
		_, err := s.repo.Create(user)
		s.NoError(err, "Expected no error when creating mock users")
	}
	users, err := s.repo.GetAll()
	s.NoError(err, "Expected no error when getting all users")
	s.Len(users, 2, "Expected 2 users to be returned")
}

func (s *UserRepositoryTestSuite) TestGetAllError() {
	// Close the underlying DB connection to simulate error on DB access
	sqlDB, err := s.db.DB()
	s.Require().NoError(err)
	err = sqlDB.Close()
	s.Require().NoError(err)

	// Test GetAll method after closing the DB
	users, err := s.repo.GetAll()
	s.Error(err, "Expected error when getting all users after closing DB")
	s.Nil(users, "Expected users to be nil after error")
}

func (s *UserRepositoryTestSuite) TestGetByID() {
	mockUsers := []*models.User{
		{ID: 1, Name: "User1", Email: "email1@example.com", Password: "password1", Gender: 1},
		{ID: 2, Name: "User2", Email: "email2@example.com", Password: "password2", Gender: 1},
	}
	for _, user := range mockUsers {
		_, err := s.repo.Create(user)
		s.NoError(err, "Expected no error when creating mock user")
	}
	// Test getting user by ID
	user, err := s.repo.GetByID(1)
	s.NoError(err, "Expected no error when getting user by ID")
	s.NotNil(user, "Expected user to be not nil")
	s.Equal("User1", user.Name, "Expected user name to be 'User1'")
}

func (s *UserRepositoryTestSuite) TestGetByIDError() {
	// Test getting user by non-existing ID
	user, err := s.repo.GetByID(999)
	s.Error(err, "Expected error when getting user by non-existing ID")
	s.Nil(user, "Expected user to be nil when not found")
}

func (s *UserRepositoryTestSuite) TestCreate() {
	mockUser := &models.User{
		Name:     "New User",
		Email:    "email@example.com",
		Password: "password",
		Gender:   1,
	}
	createdUser, err := s.repo.Create(mockUser)
	s.NoError(err, "Expected no error when creating user")
	s.NotNil(createdUser, "Expected created user to be not nil")
	s.Equal("New User", createdUser.Name, "Expected user name to be 'New User'")
}

func (s *UserRepositoryTestSuite) TestCreate_Error_Duplicate() {
	user1 := &models.User{
		Email: "test@example.com",
		Name:  "testuser",
		// other required fields
	}

	user2 := &models.User{
		Email: "test@example.com", // duplicate email to cause constraint violation
		Name:  "anotheruser",
	}

	// Create first user successfully
	createdUser, err := s.repo.Create(user1)
	s.Require().NoError(err)
	s.NotNil(createdUser)

	// Try to create second user with duplicate email
	createdUser2, err := s.repo.Create(user2)
	s.Error(err, "Should return error on duplicate user creation")
	s.Nil(createdUser2, "Created user should be nil on error")
}

func (s *UserRepositoryTestSuite) TestDeleteError() {
	// Close the underlying DB connection to simulate error on DB access
	sqlDB, err := s.db.DB()
	s.Require().NoError(err)
	err = sqlDB.Close()
	s.Require().NoError(err)

	// Now s.db is still there but the connection is closed, this should cause errors

	// No need to create new repo, because it uses s.db which now has closed connection
	err = s.repo.Delete(999)
	s.Error(err, "Expected error when deleting user with non-existing ID")
}

func (s *UserRepositoryTestSuite) TestFindByField() {

	mockUsers := []*models.User{
		{Name: "Find User", Email: "email@example.com", Password: "password", Token: utils.StringToPtr("token1"), Gender: 1},
		{Name: "Another User", Email: "another@example.com", Password: "password", Token: utils.StringToPtr("token2"), Gender: 1},
	}

	for _, user := range mockUsers {
		_, err := s.repo.Create(user)
		s.NoError(err, "Expected no error when creating mock users")
	}

	// find by field email
	foundUser, err := s.repo.FindByField("email", "email@example.com")
	s.NoError(err, "Expected no error when finding user by email")
	s.NotNil(foundUser, "Expected found user to be not nil")
	s.Equal("Find User", foundUser.Name, "Expected user name to be 'Find User'")
	// find by field name
	foundUserByName, err := s.repo.FindByField("name", "Another User")
	s.NoError(err, "Expected no error when finding user by name")
	s.NotNil(foundUserByName, "Expected found user by name to be not nil")
	// find by field token
	foundUserByToken, err := s.repo.FindByField("token", "token2")
	s.NoError(err, "Expected no error when finding user by token")
	s.NotNil(foundUserByToken, "Expected found user by token to be not nil")
	s.Equal("Another User", foundUserByToken.Name, "Expected user name to be 'Another User'")

	// Test finding user by non-existing field
	nonExistentUser, err := s.repo.FindByField("email", "notfound@example.com")
	s.Error(err, "Expected error when finding user by non-existing email")
	s.Nil(nonExistentUser, "Expected non-existent user to be nil")

	// Test finding user by non-existing field
	item, err := s.repo.FindByField("sql;", "Non Existent User")
	s.Error(err, "Expected error when finding user by invalid field")
	s.Nil(item, "Expected non-existent user to be nil")

}

func (s *UserRepositoryTestSuite) TestFindByFieldError() {
	// Test finding user by non-existing field
	nonExistentUser, err := s.repo.FindByField("email", "notfound@example.com")
	s.Error(err, "Expected error when finding user by non-existing email")
	s.Nil(nonExistentUser, "Expected non-existent user to be nil")
}

func (s *UserRepositoryTestSuite) TestGetProfile() {
	type MockUser struct {
		users *models.User
	}
	mockUsers := []MockUser{
		{
			users: &models.User{
				Name:     "Profile User",
				Email:    "profile@example.com",
				Password: "password",
				Gender:   1,
			},
		},
	}

	// Create mock users and roles
	for _, mock := range mockUsers {
		createdUser, err := s.repo.Create(mock.users)
		s.NoError(err, "Expected no error when creating mock user")
		s.NotNil(createdUser, "Expected created user to be not nil")
	}
	profile, err := s.repo.GetProfile(mockUsers[0].users.ID)

	s.NoError(err, "Expected no error when getting user profile")
	s.NotNil(profile, "Expected user profile to be not nil")
	s.Equal("Profile User", profile.Name, "Expected user name to be 'Profile User'")

}

func (s *UserRepositoryTestSuite) TestGetProfileError() {
	// Test getting profile for non-existing user
	profile, err := s.repo.GetProfile(999)
	s.Error(err, "Expected error when getting profile for non-existing user")
	s.Nil(profile, "Expected profile to be nil when user does not exist")
}

func (s *UserRepositoryTestSuite) TestUpdateProfile() {
	mockUser := &models.User{
		ID:       1,
		Name:     "Profile User",
		Email:    "email@example.com",
		Password: "password",
		Gender:   1,
	}
	// 1. Create a mock user
	createdUser, err := s.repo.Create(mockUser)
	s.NoError(err, "Expected no error when creating mock user")
	s.NotNil(createdUser, "Expected created user to be not nil")

	// 2. Update the user profile
	createdUser.Name = "Updated Profile User"
	err = s.repo.UpdateProfile(createdUser)
	s.NoError(err, "Expected no error when updating user profile")

	// 3. Retrieve the updated user profile
	updatedUser, err := s.repo.GetByID(createdUser.ID)
	s.NoError(err, "Expected no error when getting updated user by ID")
	s.NotNil(updatedUser, "Expected updated user to be not nil")
	s.Equal("Updated Profile User", updatedUser.Name, "Expected user name to be 'Updated Profile User'")

}

func (s *UserRepositoryTestSuite) TestCreateWithTx_Error_DuplicateEmail() {
	// Assume Email is unique
	user1 := &models.User{
		Email:    "duplicate@example.com",
		Name:     "user1",
		Password: "pass",
	}

	user2 := &models.User{
		Email:    "duplicate@example.com", // same email
		Name:     "user2",
		Password: "pass",
	}

	err := s.db.Create(user1).Error
	s.Require().NoError(err)

	tx := s.db.Begin()
	s.Require().NoError(tx.Error)

	createdUser, err := s.repo.CreateWithTx(tx, user2)
	s.Error(err, "Should return error due to duplicate email")
	s.Nil(createdUser, "Created user should be nil on duplicate constraint")

	tx.Rollback()
}

func (s *UserRepositoryTestSuite) TestGetDB() {
	db := s.repo.GetDB()
	s.NotNil(db, "Expected database connection to be not nil")
}

func (s *UserRepositoryTestSuite) TestUpdate() {
	// Create a mock user
	mockUser := &models.User{
		ID:       1,
		Name:     "Update User",
		Email:    "update@example.com",
		Password: "password",
		Gender:   1,
	}

	// Create the user in the database
	createdUser, err := s.repo.Create(mockUser)
	s.NoError(err, "Expected no error when creating mock user")
	s.NotNil(createdUser, "Expected created user to be not nil")
	// Update the mock user
	mockUser.ID = createdUser.ID // Ensure we update the correct user
	mockUser.Name = "Update User"
	mockUser.Email = "update@example.com"
	mockUser.Password = "newpassword"

	// Update the user in the database
	err = s.repo.Update(mockUser)
	s.NoError(err, "Expected no error when updating user")
	// Retrieve the updated user
	updatedUser, err := s.repo.GetByID(mockUser.ID)
	s.NoError(err, "Expected no error when getting updated user by ID")
	s.NotNil(updatedUser, "Expected updated user to be not nil")
	s.Equal("Update User", updatedUser.Name, "Expected user name to be 'Update User'")
	s.Equal("update@example.com", updatedUser.Email, "Expected user email to be 'update@example.com'")
	s.Equal("newpassword", updatedUser.Password, "Expected user password to be 'newpassword'")
	s.Equal(int16(1), updatedUser.Gender, "Expected user gender to be 1")
}

func (s *UserRepositoryTestSuite) TestPaginateUser() {
	// Create test users
	mockUsers := []*models.User{
		{Name: "User1", Email: "user1@example.com", Password: "password1", Gender: 1},
		{Name: "User2", Email: "user2@example.com", Password: "password2", Gender: 2},
		{Name: "User3", Email: "user3@example.com", Password: "password3", Gender: 1},
		{Name: "User4", Email: "user4@example.com", Password: "password4", Gender: 2},
		{Name: "User5", Email: "user5@example.com", Password: "password5", Gender: 1},
	}

	for _, user := range mockUsers {
		_, err := s.repo.Create(user)
		s.NoError(err, "Expected no error when creating mock users")
	}

	// Test first page with limit 2
	pagination, err := s.repo.PaginateUser(1, 2)
	s.NoError(err, "Expected no error when paginating users")
	s.NotNil(pagination, "Expected pagination to be not nil")
	s.Equal(1, pagination.Page, "Expected page to be 1")
	s.Equal(2, pagination.Limit, "Expected limit to be 2")
	s.Equal(5, pagination.TotalItems, "Expected total items to be 5")
	s.Equal(3, pagination.TotalPages, "Expected total pages to be 3")
	s.Len(pagination.Data, 2, "Expected 2 users in first page")

	// Test second page with limit 2
	pagination, err = s.repo.PaginateUser(2, 2)
	s.NoError(err, "Expected no error when paginating users")
	s.NotNil(pagination, "Expected pagination to be not nil")
	s.Equal(2, pagination.Page, "Expected page to be 2")
	s.Equal(2, pagination.Limit, "Expected limit to be 2")
	s.Equal(5, pagination.TotalItems, "Expected total items to be 5")
	s.Equal(3, pagination.TotalPages, "Expected total pages to be 3")
	s.Len(pagination.Data, 2, "Expected 2 users in second page")

	// Test third page with limit 2
	pagination, err = s.repo.PaginateUser(3, 2)
	s.NoError(err, "Expected no error when paginating users")
	s.NotNil(pagination, "Expected pagination to be not nil")
	s.Equal(3, pagination.Page, "Expected page to be 3")
	s.Equal(2, pagination.Limit, "Expected limit to be 2")
	s.Equal(5, pagination.TotalItems, "Expected total items to be 5")
	s.Equal(3, pagination.TotalPages, "Expected total pages to be 3")
	s.Len(pagination.Data, 1, "Expected 1 user in third page")

	// Test page out of range
	pagination, err = s.repo.PaginateUser(5, 2)
	s.NoError(err, "Expected no error when paginating users with page out of range")
	s.NotNil(pagination, "Expected pagination to be not nil")
	s.Equal(5, pagination.Page, "Expected page to be 5")
	s.Equal(2, pagination.Limit, "Expected limit to be 2")
	s.Equal(5, pagination.TotalItems, "Expected total items to be 5")
	s.Equal(3, pagination.TotalPages, "Expected total pages to be 3")
	s.Len(pagination.Data, 0, "Expected 0 users when page is out of range")

	// Test with different limit
	pagination, err = s.repo.PaginateUser(1, 10)
	s.NoError(err, "Expected no error when paginating users with limit 10")
	s.NotNil(pagination, "Expected pagination to be not nil")
	s.Equal(1, pagination.Page, "Expected page to be 1")
	s.Equal(10, pagination.Limit, "Expected limit to be 10")
	s.Equal(5, pagination.TotalItems, "Expected total items to be 5")
	s.Equal(1, pagination.TotalPages, "Expected total pages to be 1")
	s.Len(pagination.Data, 5, "Expected 5 users when limit is 10")
}

func (s *UserRepositoryTestSuite) TestPaginateUserError() {
	// Close the underlying DB connection to simulate error on DB access
	sqlDB, err := s.db.DB()
	s.Require().NoError(err)
	err = sqlDB.Close()
	s.Require().NoError(err)

	// Test PaginateUser method after closing the DB
	pagination, err := s.repo.PaginateUser(1, 10)
	s.Error(err, "Expected error when paginating users after closing DB")
	s.Nil(pagination, "Expected pagination to be nil after error")
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
