// Package migrations embeds the SQL migrations and applies them with
// golang-migrate. The API applies them on start (constitution, Template
// Exception 3); integration tests apply the same files to the test database.
package migrations

import (
	"embed"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	// Registers the "pgx5" database driver.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// FS holds every NNNNNN_description.{up,down}.sql file in this directory.
//
//go:embed *.sql
var FS embed.FS

// Up applies every pending migration to the PostgreSQL database at
// databaseURL (postgres:// or postgresql:// URL). No pending migration is
// not an error.
func Up(databaseURL string) error {
	driverURL, err := toDriverURL(databaseURL)
	if err != nil {
		return err
	}

	source, err := iofs.New(FS, ".")
	if err != nil {
		return fmt.Errorf("migrations: open embedded files: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, driverURL)
	if err != nil {
		return fmt.Errorf("migrations: connect: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations: apply: %w", err)
	}
	return nil
}

// PostgresURL builds a postgres:// URL with the user and password escaped.
func PostgresURL(user, password, host string, port int, dbName string) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     fmt.Sprintf("%s:%d", host, port),
		Path:     "/" + dbName,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

// toDriverURL switches a postgres:// URL to golang-migrate's pgx v5 scheme.
func toDriverURL(databaseURL string) (string, error) {
	for _, prefix := range []string{"postgres://", "postgresql://"} {
		if strings.HasPrefix(databaseURL, prefix) {
			return "pgx5://" + strings.TrimPrefix(databaseURL, prefix), nil
		}
	}
	return "", errors.New("migrations: database URL must start with postgres:// or postgresql://")
}
