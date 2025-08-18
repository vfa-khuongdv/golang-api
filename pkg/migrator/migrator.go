package migrator

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // MySQL database/sql driver
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrateIface makes Migrator testable without a real DB.
type MigrateIface interface {
	Up() error
	Down() error
	Steps(int) error
	Version() (uint, bool, error)
	Close() (error, error)
}

type Migrator struct {
	m MigrateIface
}

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// NewMigrator creates a new database migrator instance.
// It takes a migrations path and a MySQL DSN string as input.
func NewMigrator(migrationsPath, dsn string) (*Migrator, error) {
	if dsn == "" {
		return nil, fmt.Errorf("MySQL DSN must not be empty")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrator: %w", err)
	}

	return &Migrator{m: m}, nil
}

// Close closes the migrator instance and releases associated resources.
func (m *Migrator) Close() {
	if m.m != nil {
		_, _ = m.m.Close()
	}
}

// NewMySQLDSN creates a MySQL DSN string from individual connection parameters.
func NewMySQLDSN(config MySQLConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)
}

// Up applies all available up migrations.
func (m *Migrator) Up() error {
	if err := m.m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("up migration failed: %w", err)
	}
	return nil
}

// Down rolls back all migrations.
func (m *Migrator) Down() error {
	if err := m.m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("down migration failed: %w", err)
	}
	return nil
}

// Steps migrates up or down by a given number of steps.
func (m *Migrator) Steps(steps int) error {
	if err := m.m.Steps(steps); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("step migration failed: %w", err)
	}
	return nil
}

// Version returns the current migration version and dirty state.
func (m *Migrator) Version() (uint, bool, error) {
	return m.m.Version()
}
