package services

import (
	"context"
	"slices"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

type RoleService interface {
	HasPermission(ctx context.Context, userID uint, permission string) (bool, error)
	ListRoles(ctx context.Context) ([]models.Role, error)
	GetRole(ctx context.Context, id uint) (*models.Role, error)
	CreateRole(ctx context.Context, input *dto.RoleInput) (*models.Role, error)
	UpdateRole(ctx context.Context, id uint, input *dto.RoleInput) (*models.Role, error)
	DeleteRole(ctx context.Context, id uint) error
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	SetUserRoles(ctx context.Context, userID uint, roleIDs []uint) ([]models.Role, error)
}

type roleServiceImpl struct {
	roleRepo repositories.RoleRepository
	userRepo repositories.UserRepository
}

func NewRoleService(roleRepo repositories.RoleRepository, userRepo repositories.UserRepository) RoleService {
	return &roleServiceImpl{roleRepo: roleRepo, userRepo: userRepo}
}

func (s *roleServiceImpl) HasPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	return s.roleRepo.HasPermission(ctx, userID, permission)
}

func (s *roleServiceImpl) ListRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.FindAll(ctx)
}

func (s *roleServiceImpl) GetRole(ctx context.Context, id uint) (*models.Role, error) {
	return s.roleRepo.FindByID(ctx, id)
}

func (s *roleServiceImpl) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.roleRepo.FindPermissions(ctx)
}

func (s *roleServiceImpl) CreateRole(ctx context.Context, input *dto.RoleInput) (*models.Role, error) {
	if err := s.ensureNameFree(ctx, input.Name, 0); err != nil {
		return nil, err
	}
	perms, err := s.resolvePermissions(ctx, input.PermissionIDs)
	if err != nil {
		return nil, err
	}

	role := &models.Role{Name: input.Name, Description: input.Description, Permissions: perms}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleServiceImpl) UpdateRole(ctx context.Context, id uint, input *dto.RoleInput) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role.Name == models.RoleAdmin {
		return nil, apperror.NewForbiddenError("The admin role cannot be modified")
	}
	if err := s.ensureNameFree(ctx, input.Name, role.ID); err != nil {
		return nil, err
	}
	perms, err := s.resolvePermissions(ctx, input.PermissionIDs)
	if err != nil {
		return nil, err
	}

	role.Name = input.Name
	role.Description = input.Description
	role.Permissions = perms
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleServiceImpl) DeleteRole(ctx context.Context, id uint) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if role.Name == models.RoleAdmin {
		return apperror.NewForbiddenError("The admin role cannot be deleted")
	}
	return s.roleRepo.Delete(ctx, id)
}

func (s *roleServiceImpl) SetUserRoles(ctx context.Context, userID uint, roleIDs []uint) ([]models.Role, error) {
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	ids := uniqueIDs(roleIDs)
	if len(ids) > 0 {
		count, err := s.roleRepo.CountRolesByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		if int(count) != len(ids) {
			return nil, apperror.NewBadRequestError("One or more roles do not exist")
		}
	}

	if err := s.ensureAdminRetained(ctx, userID, ids); err != nil {
		return nil, err
	}

	if err := s.roleRepo.SetUserRoles(ctx, userID, ids); err != nil {
		return nil, err
	}
	return s.roleRepo.FindByUserID(ctx, userID)
}

// ensureAdminRetained rejects a change that would leave no active user with the
// admin role, since nobody could then manage roles through the API.
func (s *roleServiceImpl) ensureAdminRetained(ctx context.Context, userID uint, newRoleIDs []uint) error {
	current, err := s.roleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, role := range current {
		if role.Name != models.RoleAdmin {
			continue
		}
		if slices.Contains(newRoleIDs, role.ID) {
			return nil
		}
		others, err := s.roleRepo.CountUsersWithRole(ctx, role.ID, userID)
		if err != nil {
			return err
		}
		if others == 0 {
			return apperror.NewForbiddenError("Cannot remove the admin role from the last admin")
		}
		return nil
	}
	return nil
}

// ensureNameFree returns a conflict error when another role (not exceptID)
// already uses the name.
func (s *roleServiceImpl) ensureNameFree(ctx context.Context, name string, exceptID uint) error {
	existing, err := s.roleRepo.FindByName(ctx, name)
	if err != nil {
		if appErr, ok := apperror.ToAppError(err); ok && appErr.Code == apperror.ErrNotFound {
			return nil
		}
		return err
	}
	if existing.ID != exceptID {
		return apperror.NewConflictError("Role name already exists")
	}
	return nil
}

func (s *roleServiceImpl) resolvePermissions(ctx context.Context, permissionIDs []uint) ([]models.Permission, error) {
	ids := uniqueIDs(permissionIDs)
	perms, err := s.roleRepo.FindPermissionsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(perms) != len(ids) {
		return nil, apperror.NewBadRequestError("One or more permissions do not exist")
	}
	if perms == nil {
		perms = []models.Permission{} // serialize as [] instead of null
	}
	return perms, nil
}

func uniqueIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	unique := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
