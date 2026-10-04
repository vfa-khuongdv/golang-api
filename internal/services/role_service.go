package services

import (
	"context"
	"errors"
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
	// The methods below take the ID of the user making the change (actorID) and
	// refuse a change that involves a permission the actor does not hold, so
	// nobody can grant or take away more than they have.
	CreateRole(ctx context.Context, actorID uint, input *dto.RoleInput) (*models.Role, error)
	UpdateRole(ctx context.Context, actorID uint, id uint, input *dto.RoleInput) (*models.Role, error)
	DeleteRole(ctx context.Context, actorID uint, id uint) error
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	SetUserRoles(ctx context.Context, actorID uint, userID uint, roleIDs []uint) ([]models.Role, error)
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

func (s *roleServiceImpl) CreateRole(ctx context.Context, actorID uint, input *dto.RoleInput) (*models.Role, error) {
	if err := s.ensureNameFree(ctx, input.Name, 0); err != nil {
		return nil, err
	}
	perms, err := s.resolvePermissions(ctx, input.PermissionIDs)
	if err != nil {
		return nil, err
	}
	if err := s.ensureActorHolds(ctx, actorID, permissionIDs(perms)); err != nil {
		return nil, err
	}

	role := &models.Role{Name: input.Name, Description: input.Description, Permissions: perms}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleServiceImpl) UpdateRole(ctx context.Context, actorID uint, id uint, input *dto.RoleInput) (*models.Role, error) {
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
	// Both the permissions the role loses and the ones it gains.
	if err := s.ensureActorHolds(ctx, actorID, append(permissionIDs(role.Permissions), permissionIDs(perms)...)); err != nil {
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

func (s *roleServiceImpl) DeleteRole(ctx context.Context, actorID uint, id uint) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if role.Name == models.RoleAdmin {
		return apperror.NewForbiddenError("The admin role cannot be deleted")
	}
	if err := s.ensureActorHolds(ctx, actorID, permissionIDs(role.Permissions)); err != nil {
		return err
	}
	return s.roleRepo.Delete(ctx, id)
}

func (s *roleServiceImpl) SetUserRoles(ctx context.Context, actorID uint, userID uint, roleIDs []uint) ([]models.Role, error) {
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

	current, err := s.roleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check the roles the user gains or loses; the ones kept are unchanged.
	if changed := changedRoleIDs(current, ids); len(changed) > 0 {
		perms, err := s.roleRepo.FindPermissionIDsByRoleIDs(ctx, changed)
		if err != nil {
			return nil, err
		}
		if err := s.ensureActorHolds(ctx, actorID, perms); err != nil {
			return nil, err
		}
	}

	// The last admin must keep the role, or nobody could manage roles through
	// the API; the repository checks it in the same transaction as the write.
	err = s.roleRepo.SetUserRoles(ctx, userID, ids, removedAdminRoleID(current, ids))
	if errors.Is(err, repositories.ErrLastRoleHolder) {
		return nil, apperror.NewForbiddenError("Cannot remove the admin role from the last admin")
	}
	if err != nil {
		return nil, err
	}
	return s.roleRepo.FindByUserID(ctx, userID)
}

// removedAdminRoleID returns the ID of the admin role when current holds it and
// newRoleIDs drops it, and 0 otherwise.
func removedAdminRoleID(current []models.Role, newRoleIDs []uint) uint {
	for _, role := range current {
		if role.Name == models.RoleAdmin && !slices.Contains(newRoleIDs, role.ID) {
			return role.ID
		}
	}
	return 0
}

// ensureActorHolds rejects the change unless the actor holds every permission
// in permissionIDs. Without it, a user allowed to assign roles could assign
// themselves the admin role.
func (s *roleServiceImpl) ensureActorHolds(ctx context.Context, actorID uint, permissionIDs []uint) error {
	if len(permissionIDs) == 0 {
		return nil
	}
	held, err := s.roleRepo.FindPermissionIDsByUserID(ctx, actorID)
	if err != nil {
		return err
	}
	for _, id := range permissionIDs {
		if !slices.Contains(held, id) {
			return apperror.NewForbiddenError("You cannot grant or remove permissions you do not have")
		}
	}
	return nil
}

// changedRoleIDs returns the roles in exactly one of current and newRoleIDs.
func changedRoleIDs(current []models.Role, newRoleIDs []uint) []uint {
	changed := []uint{}
	currentIDs := make([]uint, 0, len(current))
	for _, role := range current {
		currentIDs = append(currentIDs, role.ID)
		if !slices.Contains(newRoleIDs, role.ID) {
			changed = append(changed, role.ID)
		}
	}
	for _, id := range newRoleIDs {
		if !slices.Contains(currentIDs, id) {
			changed = append(changed, id)
		}
	}
	return changed
}

func permissionIDs(perms []models.Permission) []uint {
	ids := make([]uint, len(perms))
	for i, p := range perms {
		ids[i] = p.ID
	}
	return ids
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
