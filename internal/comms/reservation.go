package comms

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/unbound-force/replicator/internal/db"
)

// Reservation represents a file path reservation.
type Reservation struct {
	ID         int    `json:"id"`
	AgentName  string `json:"agent_name"`
	Path       string `json:"path"`
	Exclusive  bool   `json:"exclusive"`
	Reason     string `json:"reason,omitempty"`
	TTLSeconds int    `json:"ttl_seconds"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at"`
}

// Reserve creates file reservations for the given paths.
// Checks for exclusive conflicts (excluding expired reservations).
func Reserve(store *db.Store, agentName string, paths []string, exclusive bool, reason string, ttlSeconds int) ([]Reservation, error) {
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(ttlSeconds) * time.Second).Format(time.RFC3339)
	nowStr := now.Format(time.RFC3339)

	tx, err := store.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	var reservations []Reservation
	err = runReservationTransaction(func() error {
		// Check for exclusive conflicts on each path.
		for _, path := range paths {
			var count int
			queryErr := tx.QueryRow(`
			SELECT COUNT(*) FROM reservations
			WHERE path = ?
			  AND exclusive = 1
			  AND agent_name != ?
			  AND expires_at > ?`,
				path, agentName, nowStr,
			).Scan(&count)
			if queryErr != nil {
				return fmt.Errorf("check conflict for %q: %w", path, queryErr)
			}
			if count > 0 {
				return fmt.Errorf("path %q is exclusively reserved by another agent", path)
			}
		}

		exclusiveInt := 0
		if exclusive {
			exclusiveInt = 1
		}

		reservations = make([]Reservation, 0, len(paths))
		for _, path := range paths {
			result, err := tx.Exec(`
			INSERT INTO reservations (agent_name, path, exclusive, reason, ttl_seconds, created_at, expires_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
				agentName, path, exclusiveInt, reason, ttlSeconds, nowStr, expiresAt,
			)
			if err != nil {
				return fmt.Errorf("insert reservation for %q: %w", path, err)
			}

			id, _ := result.LastInsertId()
			reservations = append(reservations, Reservation{
				ID:         int(id),
				AgentName:  agentName,
				Path:       path,
				Exclusive:  exclusive,
				Reason:     reason,
				TTLSeconds: ttlSeconds,
				CreatedAt:  nowStr,
				ExpiresAt:  expiresAt,
			})
		}
		return nil
	}, func() error {
		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("commit reservations: %w", commitErr)
		}
		return nil
	}, tx.Rollback)
	if err != nil {
		return nil, err
	}

	return reservations, nil
}

func runReservationTransaction(work, commit, rollback func() error) (err error) {
	committed := false
	defer func() {
		rollbackErr := rollback()
		if rollbackErr == nil || (committed && errors.Is(rollbackErr, sql.ErrTxDone)) {
			return
		}
		err = errors.Join(err, fmt.Errorf("rollback reservations: %w", rollbackErr))
	}()
	if err = work(); err != nil {
		return err
	}
	if err = commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

// Release removes reservations by path or reservation ID.
func Release(store *db.Store, paths []string, reservationIDs []int) error {
	if len(paths) == 0 && len(reservationIDs) == 0 {
		return fmt.Errorf("must specify paths or reservation IDs to release")
	}

	for _, path := range paths {
		_, err := store.DB.Exec("DELETE FROM reservations WHERE path = ?", path)
		if err != nil {
			return fmt.Errorf("release path %q: %w", path, err)
		}
	}

	for _, id := range reservationIDs {
		_, err := store.DB.Exec("DELETE FROM reservations WHERE id = ?", id)
		if err != nil {
			return fmt.Errorf("release reservation %d: %w", id, err)
		}
	}

	return nil
}

// ReleaseAll removes all reservations in a project.
// If projectPath is empty, removes all reservations regardless of project.
func ReleaseAll(store *db.Store, projectPath string) error {
	var err error
	if projectPath == "" {
		_, err = store.DB.Exec("DELETE FROM reservations")
	} else {
		_, err = store.DB.Exec("DELETE FROM reservations WHERE project_path = ? OR project_path IS NULL OR project_path = ''", projectPath)
	}
	if err != nil {
		return fmt.Errorf("release all: %w", err)
	}
	return nil
}

// ReleaseAgent removes all reservations for a specific agent.
func ReleaseAgent(store *db.Store, agentName string) error {
	_, err := store.DB.Exec("DELETE FROM reservations WHERE agent_name = ?", agentName)
	if err != nil {
		return fmt.Errorf("release agent %q: %w", agentName, err)
	}
	return nil
}
