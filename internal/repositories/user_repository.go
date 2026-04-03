package repositories

import (
	"context"
	"errors"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetAll(ctx context.Context) ([]*models.User, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
	CreateWithTx(ctx context.Context, tx *gorm.DB, user *models.User) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, userId uint) error
	FindByField(ctx context.Context, field string, value string) (*models.User, error)
	GetUsers(ctx context.Context, page int, limit int) (*dto.Pagination[*models.User], error)
	BeginTx(ctx context.Context) (*gorm.DB, error)
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
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to count users", err)
	}

	var users []*models.User
	if err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&users).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to fetch users: %v", err)
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

func (repo *userRepositoryImpl) GetAll(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := repo.db.WithContext(ctx).Find(&users).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to fetch users: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch users", err)
	}
	return users, nil
}

func (repo *userRepositoryImpl) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := repo.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.ErrNotFound, 1001, "User not found")
		}
		logger.WithContext(ctx).Errorf("DB error: failed to fetch user by id %d: %v", id, err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch user", err)
	}
	return &user, nil
}

func (repo *userRepositoryImpl) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if err := repo.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to create user: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to create user", err)
	}
	return user, nil
}

func (repo *userRepositoryImpl) CreateWithTx(ctx context.Context, tx *gorm.DB, user *models.User) (*models.User, error) {
	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to create user with tx: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to create user", err)
	}
	return user, nil
}

func (repo *userRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	if err := repo.db.WithContext(ctx).Save(user).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to update user id %d: %v", user.ID, err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to update user", err)
	}
	return nil
}

func (repo *userRepositoryImpl) Delete(ctx context.Context, userId uint) error {
	var user models.User
	if err := repo.db.WithContext(ctx).Delete(&user, userId).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to delete user id %d: %v", userId, err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to delete user", err)
	}
	return nil
}

func (repo *userRepositoryImpl) FindByField(ctx context.Context, field string, value string) (*models.User, error) {
	allowedFields := map[string]bool{
		"name":  true,
		"email": true,
		"token": true,
	}

	if !allowedFields[field] {
		return nil, apperror.New(apperror.ErrBadRequest, 1002, "Invalid field")
	}

	var user models.User
	if err := repo.db.WithContext(ctx).Where(field+" = ?", value).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.ErrUnauthorized, 1003, "User not found")
		}
		logger.WithContext(ctx).Errorf("DB error: failed to fetch user by field %s: %v", field, err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch user", err)
	}
	return &user, nil
}

func (repo *userRepositoryImpl) BeginTx(ctx context.Context) (*gorm.DB, error) {
	tx := repo.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to begin transaction: %v", tx.Error)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to begin transaction", tx.Error)
	}
	return tx, nil
}
