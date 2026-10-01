package repositories

import (
	"context"
	"errors"
	"net/http"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type RoleRepository interface {
	// HasPermission reports whether any role of the user grants the permission.
	HasPermission(ctx context.Context, userID uint, permission string) (bool, error)
	FindAll(ctx context.Context) ([]models.Role, error)
	FindByID(ctx context.Context, id uint) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	FindByUserID(ctx context.Context, userID uint) ([]models.Role, error)
	// Create inserts the role and links role.Permissions.
	Create(ctx context.Context, role *models.Role) error
	// Update saves the role fields and replaces its permissions with role.Permissions.
	Update(ctx context.Context, role *models.Role) error
	// Delete removes the role together with its permission and user links.
	Delete(ctx context.Context, id uint) error
	FindPermissions(ctx context.Context) ([]models.Permission, error)
	// FindPermissionsByIDs skips IDs that do not exist.
	FindPermissionsByIDs(ctx context.Context, ids []uint) ([]models.Permission, error)
	// CountRolesByIDs counts how many of the IDs exist.
	CountRolesByIDs(ctx context.Context, ids []uint) (int64, error)
	// SetUserRoles replaces all roles of the user atomically.
	SetUserRoles(ctx context.Context, userID uint, roleIDs []uint) error
}

type roleRepositoryImpl struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepositoryImpl{db: db}
}

func dbError(ctx context.Context, msg string, err error) error {
	logger.WithContext(ctx).Errorf("DB error: %s: %v", msg, err)
	return apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, msg, err)
}

func (repo *roleRepositoryImpl) HasPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	var count int64
	err := repo.db.WithContext(ctx).Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.name = ?", userID, permission).
		Count(&count).Error
	if err != nil {
		return false, dbError(ctx, "Failed to check permission", err)
	}
	return count > 0, nil
}

func (repo *roleRepositoryImpl) FindAll(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	if err := repo.db.WithContext(ctx).Preload("Permissions").Order("id").Find(&roles).Error; err != nil {
		return nil, dbError(ctx, "Failed to fetch roles", err)
	}
	return roles, nil
}

func (repo *roleRepositoryImpl) findOne(ctx context.Context, query string, arg any) (*models.Role, error) {
	var role models.Role
	err := repo.db.WithContext(ctx).Preload("Permissions").Where(query, arg).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(http.StatusNotFound, apperror.ErrNotFound, "Role not found")
		}
		return nil, dbError(ctx, "Failed to fetch role", err)
	}
	return &role, nil
}

func (repo *roleRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.Role, error) {
	return repo.findOne(ctx, "id = ?", id)
}

func (repo *roleRepositoryImpl) FindByName(ctx context.Context, name string) (*models.Role, error) {
	return repo.findOne(ctx, "name = ?", name)
}

func (repo *roleRepositoryImpl) FindByUserID(ctx context.Context, userID uint) ([]models.Role, error) {
	var roles []models.Role
	err := repo.db.WithContext(ctx).Preload("Permissions").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).Order("roles.id").Find(&roles).Error
	if err != nil {
		return nil, dbError(ctx, "Failed to fetch user roles", err)
	}
	return roles, nil
}

func (repo *roleRepositoryImpl) Create(ctx context.Context, role *models.Role) error {
	if err := repo.db.WithContext(ctx).Create(role).Error; err != nil {
		return dbError(ctx, "Failed to create role", err)
	}
	return nil
}

func (repo *roleRepositoryImpl) Update(ctx context.Context, role *models.Role) error {
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Permissions").Save(role).Error; err != nil {
			return err
		}
		return tx.Model(role).Association("Permissions").Replace(role.Permissions)
	})
	if err != nil {
		return dbError(ctx, "Failed to update role", err)
	}
	return nil
}

func (repo *roleRepositoryImpl) Delete(ctx context.Context, id uint) error {
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Role{}, id).Error
	})
	if err != nil {
		return dbError(ctx, "Failed to delete role", err)
	}
	return nil
}

func (repo *roleRepositoryImpl) FindPermissions(ctx context.Context) ([]models.Permission, error) {
	var perms []models.Permission
	if err := repo.db.WithContext(ctx).Order("id").Find(&perms).Error; err != nil {
		return nil, dbError(ctx, "Failed to fetch permissions", err)
	}
	return perms, nil
}

func (repo *roleRepositoryImpl) FindPermissionsByIDs(ctx context.Context, ids []uint) ([]models.Permission, error) {
	var perms []models.Permission
	if len(ids) == 0 {
		return perms, nil
	}
	if err := repo.db.WithContext(ctx).Where("id IN ?", ids).Order("id").Find(&perms).Error; err != nil {
		return nil, dbError(ctx, "Failed to fetch permissions", err)
	}
	return perms, nil
}

func (repo *roleRepositoryImpl) CountRolesByIDs(ctx context.Context, ids []uint) (int64, error) {
	var count int64
	if len(ids) == 0 {
		return 0, nil
	}
	if err := repo.db.WithContext(ctx).Model(&models.Role{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return 0, dbError(ctx, "Failed to count roles", err)
	}
	return count, nil
}

func (repo *roleRepositoryImpl) SetUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		rows := make([]models.UserRole, len(roleIDs))
		for i, id := range roleIDs {
			rows[i] = models.UserRole{UserID: userID, RoleID: id}
		}
		return tx.Create(&rows).Error
	})
	if err != nil {
		return dbError(ctx, "Failed to set user roles", err)
	}
	return nil
}
