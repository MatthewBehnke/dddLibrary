// Package migrate applies the embedded schema migrations to a Postgres database
// using golang-migrate. It is a platform concern shared by the server (which
// migrates on startup) and integration tests.
package migrate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the "pgx5" scheme + pgx stdlib driver
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/matthewbehnke/dddlibrary/migrations"
)

// Run applies all outstanding migrations to the database at databaseURL. A
// standard "postgres://" URL is accepted and rewritten to the "pgx5://" scheme
// golang-migrate's pgx v5 driver expects. It is a no-op when the schema is
// already current.
func Run(databaseURL string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("open migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, toPgx5URL(databaseURL))
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// toPgx5URL rewrites a postgres/postgresql scheme to pgx5, which selects
// golang-migrate's pgx v5 database driver.
func toPgx5URL(u string) string {
	for _, prefix := range []string{"postgres://", "postgresql://"} {
		if strings.HasPrefix(u, prefix) {
			return "pgx5://" + strings.TrimPrefix(u, prefix)
		}
	}
	return u
}
