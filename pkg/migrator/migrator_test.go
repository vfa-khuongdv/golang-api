package migrator

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/stretchr/testify/assert"
)

type fakeMigrate struct {
	upCalled    bool
	downCalled  bool
	stepsCalled int
	version     uint
	dirty       bool
	returnErr   error
	versionErr  error
	closed      bool
}

func (f *fakeMigrate) Up() error { f.upCalled = true; return f.returnErr }
func (f *fakeMigrate) Down() error { f.downCalled = true; return f.returnErr }
func (f *fakeMigrate) Steps(n int) error { f.stepsCalled = n; return f.returnErr }
func (f *fakeMigrate) Version() (uint, bool, error) {
	if f.versionErr != nil {
		return 0, false, f.versionErr
	}
	return f.version, f.dirty, nil
}
func (f *fakeMigrate) Close() (error, error) { f.closed = true; return nil, nil }

func TestNewMigrator_Hooks(t *testing.T) {
	originalOpen := openSQLConnection
	originalBuild := buildMySQLDriver
	originalCreate := createMigrateInstance
	t.Cleanup(func() {
		openSQLConnection = originalOpen
		buildMySQLDriver = originalBuild
		createMigrateInstance = originalCreate
	})

	t.Run("EmptyDSN", func(t *testing.T) {
		m, err := NewMigrator("./migrations", "")
		assert.Nil(t, m)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must not be empty")
	})

	t.Run("DriverBuildError", func(t *testing.T) {
		openSQLConnection = func(_, _ string) (*sql.DB, error) { return &sql.DB{}, nil }
		buildMySQLDriver = func(_ *sql.DB) (database.Driver, error) { return nil, errors.New("build failed") }

		m, err := NewMigrator("./migrations", "root:pass@tcp(localhost:3306)/test")
		assert.Nil(t, m)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create MySQL driver")
	})

	t.Run("CreateInstanceError", func(t *testing.T) {
		openSQLConnection = func(_, _ string) (*sql.DB, error) { return &sql.DB{}, nil }
		buildMySQLDriver = func(_ *sql.DB) (database.Driver, error) { return nil, nil }
		createMigrateInstance = func(_ string, _ database.Driver) (MigrateIface, error) {
			return nil, errors.New("create failed")
		}

		m, err := NewMigrator("./migrations", "root:pass@tcp(localhost:3306)/test")
		assert.Nil(t, m)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to initialize migrator")
	})

	t.Run("Success", func(t *testing.T) {
		openSQLConnection = func(_, _ string) (*sql.DB, error) { return &sql.DB{}, nil }
		buildMySQLDriver = func(_ *sql.DB) (database.Driver, error) { return nil, nil }
		createMigrateInstance = func(_ string, _ database.Driver) (MigrateIface, error) { return &fakeMigrate{}, nil }

		m, err := NewMigrator("./migrations", "root:pass@tcp(localhost:3306)/test")
		assert.NoError(t, err)
		assert.NotNil(t, m)
	})
}

func TestOpenSQLConnection_DefaultFunc(t *testing.T) {
	db, err := openSQLConnection("mysql", "invalid-dsn")
	if db != nil {
		_ = db.Close()
	}
	assert.Error(t, err)
}

func TestNewMigrator_OpenConnectionError(t *testing.T) {
	originalOpen := openSQLConnection
	originalBuild := buildMySQLDriver
	originalCreate := createMigrateInstance
	t.Cleanup(func() {
		openSQLConnection = originalOpen
		buildMySQLDriver = originalBuild
		createMigrateInstance = originalCreate
	})

	openSQLConnection = func(_, _ string) (*sql.DB, error) {
		return nil, errors.New("open failed")
	}

	m, err := NewMigrator("./migrations", "root:pass@tcp(localhost:3306)/test")
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestBuildMySQLDriver_DefaultFunc(t *testing.T) {
	assert.Panics(t, func() {
		_, _ = buildMySQLDriver(&sql.DB{})
	})
}

func TestCreateMigrateInstance_DefaultFunc(t *testing.T) {
	m, err := createMigrateInstance("file://definitely-not-exist", nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestUp(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		f := &fakeMigrate{}
		m := &Migrator{m: f}
		assert.NoError(t, m.Up())
		assert.True(t, f.upCalled)
	})
	t.Run("ErrNoChange", func(t *testing.T) {
		f := &fakeMigrate{returnErr: migrate.ErrNoChange}
		m := &Migrator{m: f}
		assert.NoError(t, m.Up())
	})
	t.Run("Error", func(t *testing.T) {
		f := &fakeMigrate{returnErr: errors.New("boom")}
		m := &Migrator{m: f}
		err := m.Up()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "up migration failed")
	})
}

func TestDown(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		f := &fakeMigrate{}
		m := &Migrator{m: f}
		assert.NoError(t, m.Down())
		assert.True(t, f.downCalled)
	})
	t.Run("ErrNoChange", func(t *testing.T) {
		f := &fakeMigrate{returnErr: migrate.ErrNoChange}
		m := &Migrator{m: f}
		assert.NoError(t, m.Down())
	})
	t.Run("Error", func(t *testing.T) {
		f := &fakeMigrate{returnErr: errors.New("down failed")}
		m := &Migrator{m: f}
		err := m.Down()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "down migration failed")
	})
}

func TestSteps(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		f := &fakeMigrate{}
		m := &Migrator{m: f}
		assert.NoError(t, m.Steps(2))
		assert.Equal(t, 2, f.stepsCalled)
	})
	t.Run("ErrNoChange", func(t *testing.T) {
		f := &fakeMigrate{returnErr: migrate.ErrNoChange}
		m := &Migrator{m: f}
		assert.NoError(t, m.Steps(1))
	})
	t.Run("Error", func(t *testing.T) {
		f := &fakeMigrate{returnErr: errors.New("steps failed")}
		m := &Migrator{m: f}
		err := m.Steps(3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step migration failed")
	})
}

func TestVersion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		f := &fakeMigrate{version: 5, dirty: true}
		m := &Migrator{m: f}
		v, dirty, err := m.Version()
		assert.NoError(t, err)
		assert.Equal(t, uint(5), v)
		assert.True(t, dirty)
	})
	t.Run("Error", func(t *testing.T) {
		f := &fakeMigrate{versionErr: errors.New("version failed")}
		m := &Migrator{m: f}
		_, _, err := m.Version()
		assert.Error(t, err)
	})
}

func TestClose(t *testing.T) {
	f := &fakeMigrate{}
	m := &Migrator{m: f}
	m.Close()
	assert.True(t, f.closed)

	m2 := &Migrator{m: nil}
	m2.Close()
}

func TestNewMySQLDSN(t *testing.T) {
	cfg := MySQLConfig{
		User:     "root",
		Password: "pass",
		Host:     "127.0.0.1",
		Port:     "3306",
		DBName:   "testdb",
	}
	dsn := NewMySQLDSN(cfg)
	assert.Contains(t, dsn, "root:pass@tcp(127.0.0.1:3306)/testdb")
}
