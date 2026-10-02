package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.UserRole{}))
	return db
}

func TestRoleRepository_HasPermission(t *testing.T) {
	ctx := context.Background()

	t.Run("True When A Role Of The User Has The Permission", func(t *testing.T) {
		db := setupRoleTestDB(t)
		role := models.Role{Name: "editor", Permissions: []models.Permission{{Name: "settings:read"}}}
		require.NoError(t, db.Create(&role).Error)
		require.NoError(t, db.Create(&models.UserRole{UserID: 1, RoleID: role.ID}).Error)
		repo := repositories.NewRoleRepository(db)

		ok, err := repo.HasPermission(ctx, 1, "settings:read")

		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("False When The Permission Belongs To A Role The User Does Not Have", func(t *testing.T) {
		db := setupRoleTestDB(t)
		role := models.Role{Name: "editor", Permissions: []models.Permission{{Name: "settings:read"}}}
		require.NoError(t, db.Create(&role).Error)
		require.NoError(t, db.Create(&models.UserRole{UserID: 2, RoleID: role.ID}).Error)
		repo := repositories.NewRoleRepository(db)

		ok, err := repo.HasPermission(ctx, 1, "settings:read")

		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("False For A Different Permission", func(t *testing.T) {
		db := setupRoleTestDB(t)
		role := models.Role{Name: "editor", Permissions: []models.Permission{{Name: "settings:read"}}}
		require.NoError(t, db.Create(&role).Error)
		require.NoError(t, db.Create(&models.UserRole{UserID: 1, RoleID: role.ID}).Error)
		repo := repositories.NewRoleRepository(db)

		ok, err := repo.HasPermission(ctx, 1, "settings:update")

		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("DB Error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{}) // no tables
		require.NoError(t, err)
		repo := repositories.NewRoleRepository(db)

		_, err = repo.HasPermission(ctx, 1, "settings:read")

		assert.Error(t, err)
	})
}

func TestRoleRepository_CRUD(t *testing.T) {
	ctx := context.Background()

	t.Run("Create Then FindByID Loads Permissions", func(t *testing.T) {
		db := setupRoleTestDB(t)
		perm := models.Permission{Name: "roles:read"}
		require.NoError(t, db.Create(&perm).Error)
		repo := repositories.NewRoleRepository(db)

		role := &models.Role{Name: "auditor", Description: "read only", Permissions: []models.Permission{perm}}
		require.NoError(t, repo.Create(ctx, role))
		got, err := repo.FindByID(ctx, role.ID)

		require.NoError(t, err)
		assert.Equal(t, "auditor", got.Name)
		require.Len(t, got.Permissions, 1)
		assert.Equal(t, "roles:read", got.Permissions[0].Name)
	})

	t.Run("FindByName", func(t *testing.T) {
		db := setupRoleTestDB(t)
		require.NoError(t, db.Create(&models.Role{Name: "dup"}).Error)
		repo := repositories.NewRoleRepository(db)

		got, err := repo.FindByName(ctx, "dup")
		require.NoError(t, err)
		assert.Equal(t, "dup", got.Name)

		_, err = repo.FindByName(ctx, "missing")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("FindByID Not Found", func(t *testing.T) {
		repo := repositories.NewRoleRepository(setupRoleTestDB(t))

		_, err := repo.FindByID(ctx, 99)

		assert.ErrorContains(t, err, "not found")
	})

	t.Run("FindAll Returns Roles With Permissions", func(t *testing.T) {
		db := setupRoleTestDB(t)
		require.NoError(t, db.Create(&models.Role{Name: "a", Permissions: []models.Permission{{Name: "p1"}}}).Error)
		require.NoError(t, db.Create(&models.Role{Name: "b"}).Error)
		repo := repositories.NewRoleRepository(db)

		roles, err := repo.FindAll(ctx)

		require.NoError(t, err)
		require.Len(t, roles, 2)
		assert.Len(t, roles[0].Permissions, 1)
	})

	t.Run("Update Replaces Fields And Permissions", func(t *testing.T) {
		db := setupRoleTestDB(t)
		p1, p2 := models.Permission{Name: "p1"}, models.Permission{Name: "p2"}
		require.NoError(t, db.Create(&[]models.Permission{p1, p2}).Error)
		require.NoError(t, db.Find(&p1, "name = ?", "p1").Error)
		require.NoError(t, db.Find(&p2, "name = ?", "p2").Error)
		role := models.Role{Name: "old", Permissions: []models.Permission{p1}}
		require.NoError(t, db.Create(&role).Error)
		repo := repositories.NewRoleRepository(db)

		role.Name = "new"
		role.Permissions = []models.Permission{p2}
		require.NoError(t, repo.Update(ctx, &role))
		got, err := repo.FindByID(ctx, role.ID)

		require.NoError(t, err)
		assert.Equal(t, "new", got.Name)
		require.Len(t, got.Permissions, 1)
		assert.Equal(t, "p2", got.Permissions[0].Name)
	})

	t.Run("Delete Removes Role And Its Links", func(t *testing.T) {
		db := setupRoleTestDB(t)
		role := models.Role{Name: "gone", Permissions: []models.Permission{{Name: "p1"}}}
		require.NoError(t, db.Create(&role).Error)
		require.NoError(t, db.Create(&models.UserRole{UserID: 1, RoleID: role.ID}).Error)
		repo := repositories.NewRoleRepository(db)

		require.NoError(t, repo.Delete(ctx, role.ID))

		var roles, links, userRoles int64
		db.Model(&models.Role{}).Count(&roles)
		db.Table("role_permissions").Count(&links)
		db.Model(&models.UserRole{}).Count(&userRoles)
		assert.Zero(t, roles)
		assert.Zero(t, links)
		assert.Zero(t, userRoles)
	})
}

func TestRoleRepository_Lookups(t *testing.T) {
	ctx := context.Background()

	t.Run("FindPermissions Lists All", func(t *testing.T) {
		db := setupRoleTestDB(t)
		require.NoError(t, db.Create(&[]models.Permission{{Name: "a"}, {Name: "b"}}).Error)
		repo := repositories.NewRoleRepository(db)

		perms, err := repo.FindPermissions(ctx)

		require.NoError(t, err)
		assert.Len(t, perms, 2)
	})

	t.Run("FindPermissionsByIDs Skips Unknown IDs", func(t *testing.T) {
		db := setupRoleTestDB(t)
		p := models.Permission{Name: "a"}
		require.NoError(t, db.Create(&p).Error)
		repo := repositories.NewRoleRepository(db)

		perms, err := repo.FindPermissionsByIDs(ctx, []uint{p.ID, 999})

		require.NoError(t, err)
		require.Len(t, perms, 1)
	})

	t.Run("CountRolesByIDs", func(t *testing.T) {
		db := setupRoleTestDB(t)
		r := models.Role{Name: "r"}
		require.NoError(t, db.Create(&r).Error)
		repo := repositories.NewRoleRepository(db)

		n, err := repo.CountRolesByIDs(ctx, []uint{r.ID, 999})

		require.NoError(t, err)
		assert.EqualValues(t, 1, n)
	})
}

func TestRoleRepository_SetUserRoles(t *testing.T) {
	ctx := context.Background()

	t.Run("Replaces Existing Roles", func(t *testing.T) {
		db := setupRoleTestDB(t)
		r1, r2 := models.Role{Name: "r1"}, models.Role{Name: "r2"}
		require.NoError(t, db.Create(&r1).Error)
		require.NoError(t, db.Create(&r2).Error)
		require.NoError(t, db.Create(&models.UserRole{UserID: 7, RoleID: r1.ID}).Error)
		repo := repositories.NewRoleRepository(db)

		require.NoError(t, repo.SetUserRoles(ctx, 7, []uint{r2.ID}))

		var rows []models.UserRole
		require.NoError(t, db.Where("user_id = ?", 7).Find(&rows).Error)
		require.Len(t, rows, 1)
		assert.Equal(t, r2.ID, rows[0].RoleID)
	})

	t.Run("Empty List Clears Roles", func(t *testing.T) {
		db := setupRoleTestDB(t)
		require.NoError(t, db.Create(&models.UserRole{UserID: 7, RoleID: 1}).Error)
		repo := repositories.NewRoleRepository(db)

		require.NoError(t, repo.SetUserRoles(ctx, 7, nil))

		var n int64
		db.Model(&models.UserRole{}).Where("user_id = ?", 7).Count(&n)
		assert.Zero(t, n)
	})

	t.Run("FindUserRoles Returns Role Names", func(t *testing.T) {
		db := setupRoleTestDB(t)
		r := models.Role{Name: "r1"}
		require.NoError(t, db.Create(&r).Error)
		require.NoError(t, db.Create(&models.UserRole{UserID: 7, RoleID: r.ID}).Error)
		repo := repositories.NewRoleRepository(db)

		roles, err := repo.FindByUserID(ctx, 7)

		require.NoError(t, err)
		require.Len(t, roles, 1)
		assert.Equal(t, "r1", roles[0].Name)
	})
}

func TestRoleRepository_CountUsersWithRole(t *testing.T) {
	ctx := context.Background()

	t.Run("Counts Other Active Users Only", func(t *testing.T) {
		db := setupRoleTestDB(t)
		role := models.Role{Name: "admin"}
		require.NoError(t, db.Create(&role).Error)
		active1 := models.User{Email: "a1@example.com", Name: "a1", Password: "x"}
		active2 := models.User{Email: "a2@example.com", Name: "a2", Password: "x"}
		deleted := models.User{Email: "d@example.com", Name: "d", Password: "x"}
		require.NoError(t, db.Create(&[]*models.User{&active1, &active2, &deleted}).Error)
		require.NoError(t, db.Delete(&deleted).Error)
		for _, u := range []models.User{active1, active2, deleted} {
			require.NoError(t, db.Create(&models.UserRole{UserID: u.ID, RoleID: role.ID}).Error)
		}
		repo := repositories.NewRoleRepository(db)

		n, err := repo.CountUsersWithRole(ctx, role.ID, active1.ID)

		require.NoError(t, err)
		assert.EqualValues(t, 1, n, "excludes the given user and soft-deleted users")
	})

	t.Run("DB Error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{}) // no tables
		require.NoError(t, err)

		_, err = repositories.NewRoleRepository(db).CountUsersWithRole(ctx, 1, 2)

		assert.Error(t, err)
	})
}

func TestRoleRepository_FindPermissionIDsByUserID(t *testing.T) {
	ctx := context.Background()

	t.Run("Returns The Distinct Permissions Of All Roles Of The User", func(t *testing.T) {
		db := setupRoleTestDB(t)
		read, update, other := models.Permission{Name: "a:read"}, models.Permission{Name: "a:update"}, models.Permission{Name: "b:read"}
		require.NoError(t, db.Create(&[]*models.Permission{&read, &update, &other}).Error)
		r1 := models.Role{Name: "r1", Permissions: []models.Permission{read}}
		r2 := models.Role{Name: "r2", Permissions: []models.Permission{read, update}}
		r3 := models.Role{Name: "r3", Permissions: []models.Permission{other}}
		require.NoError(t, db.Create(&[]*models.Role{&r1, &r2, &r3}).Error)
		require.NoError(t, db.Create(&[]models.UserRole{{UserID: 7, RoleID: r1.ID}, {UserID: 7, RoleID: r2.ID}, {UserID: 8, RoleID: r3.ID}}).Error)
		repo := repositories.NewRoleRepository(db)

		ids, err := repo.FindPermissionIDsByUserID(ctx, 7)

		require.NoError(t, err)
		assert.ElementsMatch(t, []uint{read.ID, update.ID}, ids)
	})

	t.Run("DB Error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{}) // no tables
		require.NoError(t, err)

		_, err = repositories.NewRoleRepository(db).FindPermissionIDsByUserID(ctx, 1)

		assert.Error(t, err)
	})
}

func TestRoleRepository_FindPermissionIDsByRoleIDs(t *testing.T) {
	ctx := context.Background()

	t.Run("Returns The Distinct Permissions Of The Roles", func(t *testing.T) {
		db := setupRoleTestDB(t)
		read, update, other := models.Permission{Name: "a:read"}, models.Permission{Name: "a:update"}, models.Permission{Name: "b:read"}
		require.NoError(t, db.Create(&[]*models.Permission{&read, &update, &other}).Error)
		r1 := models.Role{Name: "r1", Permissions: []models.Permission{read}}
		r2 := models.Role{Name: "r2", Permissions: []models.Permission{read, update}}
		r3 := models.Role{Name: "r3", Permissions: []models.Permission{other}}
		require.NoError(t, db.Create(&[]*models.Role{&r1, &r2, &r3}).Error)
		repo := repositories.NewRoleRepository(db)

		ids, err := repo.FindPermissionIDsByRoleIDs(ctx, []uint{r1.ID, r2.ID})

		require.NoError(t, err)
		assert.ElementsMatch(t, []uint{read.ID, update.ID}, ids)
	})

	t.Run("No Roles Returns No Permissions", func(t *testing.T) {
		repo := repositories.NewRoleRepository(setupRoleTestDB(t))

		ids, err := repo.FindPermissionIDsByRoleIDs(ctx, nil)

		require.NoError(t, err)
		assert.Empty(t, ids)
	})

	t.Run("DB Error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{}) // no tables
		require.NoError(t, err)

		_, err = repositories.NewRoleRepository(db).FindPermissionIDsByRoleIDs(ctx, []uint{1})

		assert.Error(t, err)
	})
}
