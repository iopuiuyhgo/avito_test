package postgres

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"strings"
)

func getMigrate(postgresConnect string, path string) (*migrate.Migrate, error) {
	postgresConnect = strings.Replace(postgresConnect, "postgres://", "pgx://", 1)

	m, err := migrate.New(
		"file:/"+path,
		postgresConnect,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func UpMigrations(postgresConnect string, path string) error {
	m, err := getMigrate(postgresConnect, path)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	err, err2 := m.Close()
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func DownMigrations(postgresConnect string, path string) error {
	m, err := getMigrate(postgresConnect, path)
	if err != nil {
		return err
	}

	err = m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
