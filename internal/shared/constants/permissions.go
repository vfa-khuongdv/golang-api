package constants

// Permission names checked by RequirePermission. Keep in sync with the rows
// seeded in migration 000005_create_rbac_tables.
const (
	PermissionRolesRead       = "roles:read"
	PermissionRolesCreate     = "roles:create"
	PermissionRolesUpdate     = "roles:update"
	PermissionRolesDelete     = "roles:delete"
	PermissionPermissionsRead = "permissions:read"
	PermissionUsersAssign     = "users:assign-roles"
	PermissionSettingsRead    = "settings:read"
	PermissionSettingsUpdate  = "settings:update"
)

// AllPermissions lists every permission the application checks.
func AllPermissions() []string {
	return []string{
		PermissionRolesRead,
		PermissionRolesCreate,
		PermissionRolesUpdate,
		PermissionRolesDelete,
		PermissionPermissionsRead,
		PermissionUsersAssign,
		PermissionSettingsRead,
		PermissionSettingsUpdate,
	}
}
