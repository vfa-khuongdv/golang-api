package dto

// RoleInput is the body to create or update a role. PermissionIDs replaces the
// role's permissions; omit or leave empty for a role without permissions.
type RoleInput struct {
	Name          string `json:"name" binding:"required,max=100,not_blank"`
	Description   string `json:"description" binding:"omitempty,max=255"`
	PermissionIDs []uint `json:"permission_ids" binding:"omitempty,dive,gt=0"`
}

// SetUserRolesInput replaces all roles of a user. An empty list removes every role.
type SetUserRolesInput struct {
	RoleIDs []uint `json:"role_ids" binding:"required,dive,gt=0"`
}
