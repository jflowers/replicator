// Package db manages the SQLite database connection and schema migrations.
//
// Uses modernc.org/sqlite (pure Go, no CGo) for zero-dependency builds.
// The schema is compatible with cyborg-swarm's libSQL database.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	_ "modernc.org/sqlite"
)

// Store wraps a SQLite database connection.
type Store struct {
	DB *sql.DB
}

// Open opens (or creates) a SQLite database at the given path
// and runs schema migrations.
func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON", path)
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Set connection pool for SQLite (single writer, multiple readers).
	sqlDB.SetMaxOpenConns(1)

	if err := sqlDB.Ping(); err != nil {
		return nil, closeOnOpenError(fmt.Errorf("ping sqlite: %w", err), sqlDB)
	}

	store := &Store{DB: sqlDB}
	if err := store.migrate(); err != nil {
		return nil, closeOnOpenError(fmt.Errorf("migrate: %w", err), sqlDB)
	}

	return store, nil
}

// OpenMemory opens an in-memory SQLite database for testing.
func OpenMemory() (*Store, error) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	store := &Store{DB: sqlDB}
	if err := store.migrate(); err != nil {
		return nil, closeOnOpenError(fmt.Errorf("migrate memory database: %w", err), sqlDB)
	}
	return store, nil
}

func closeOnOpenError(primary error, closer io.Closer) error {
	if primary == nil {
		return nil
	}
	if closeErr := closer.Close(); closeErr != nil {
		return errors.Join(primary, fmt.Errorf("close database after open failure: %w", closeErr))
	}
	return primary
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.DB.Close()
}

// migrate runs all schema migrations.
func (s *Store) migrate() error {
	for _, m := range migrations {
		if _, err := s.DB.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, m)
		}
	}
	return nil
}

// migrations are executed in order on every Open.
// Each statement uses IF NOT EXISTS for idempotency.
var migrations = []string{
	migrationEvents,
	migrationAgents,
	migrationCells,
	migrationCellEvents,
	migrationSessions,
	migrationMessages,
	migrationReservations,
}
