package constants_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/constants"
)

// The default permissions are seeded by SQL, so the migration must stay in
// sync with the constants the routes check.
func TestAllPermissionsAreSeededByMigration(t *testing.T) {
	sql, err := os.ReadFile("../../database/migrations/000005_create_rbac_tables.up.sql")
	require.NoError(t, err)

	require.NotEmpty(t, constants.AllPermissions())
	for _, permission := range constants.AllPermissions() {
		assert.True(t, strings.Contains(string(sql), "'"+permission+"'"), "migration does not seed %q", permission)
	}
}
