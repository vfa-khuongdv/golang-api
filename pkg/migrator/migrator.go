package migrator

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

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

var (
	openSQLConnection = func(driverName, dsn string) (*sql.DB, error) {
		return sql.Open(driverName, dsn)
	}
	buildMySQLDriver = func(db *sql.DB) (database.Driver, error) {
		return mysql.WithInstance(db, &mysql.Config{
			MigrationsTable: "schema_migrations",
		})
	}
	createMigrateInstance = func(sourceURL string, driver database.Driver) (MigrateIface, error) {
		return migrate.NewWithDatabaseInstance(sourceURL, "mysql", driver)
	}
)

func NewMigrator(migrationsPath, dsn string) (*Migrator, error) {
	if dsn == "" {
		return nil, fmt.Errorf("MySQL DSN must not be empty")
	}

	db, err := openSQLConnection("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	driver, err := buildMySQLDriver(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL driver: %w", err)
	}

	m, err := createMigrateInstance(fmt.Sprintf("file://%s", migrationsPath), driver)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrator: %w", err)
	}

	return &Migrator{m: m}, nil
}

func (m *Migrator) Close() {
	if m.m != nil {
		_, _ = m.m.Close()
	}
}

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

func (m *Migrator) Up() error {
	if err := m.m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("up migration failed: %w", err)
	}
	return nil
}

func (m *Migrator) Down() error {
	if err := m.m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("down migration failed: %w", err)
	}
	return nil
}

func (m *Migrator) Steps(steps int) error {
	if err := m.m.Steps(steps); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("step migration failed: %w", err)
	}
	return nil
}

func (m *Migrator) Version() (uint, bool, error) {
	return m.m.Version()
}
