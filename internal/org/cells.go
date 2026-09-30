// Package hive implements work item tracking (cells).
//
// Cells are the core work item abstraction -- equivalent to issues/tasks
// but using the hive/swarm metaphor. Each cell has a type (task, bug,
// feature, epic, chore), a status (open, in_progress, blocked, closed),
// and optional parent-child relationships for epics.
package org

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/unbound-force/replicator/internal/db"
)

// Cell represents a single work item in the hive.
//
// Fields like Description and ParentID are always serialized (no omitempty)
// to maintain shape parity with the TypeScript cyborg-swarm responses.
// Optional fields that are truly absent use pointer types with explicit null.
type Cell struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Priority    int      `json:"priority"`
	ParentID    *string  `json:"parent_id"`
	ProjectKey  string   `json:"project_key,omitempty"`
	AssignedTo  *string  `json:"assigned_to,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	ClosedAt    *string  `json:"closed_at,omitempty"`
	CloseReason *string  `json:"close_reason,omitempty"`
}

// CellQuery defines filters for querying cells.
type CellQuery struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	Type   string `json:"type,omitempty"`
	Ready  bool   `json:"ready,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// CreateCellInput defines the input for creating a new cell.
type CreateCellInput struct {
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Type        string  `json:"type,omitempty"`
	Priority    int     `json:"priority,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
}

// CreateCell creates a new cell in the hive.
func CreateCell(store *db.Store, input CreateCellInput) (*Cell, error) {
	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("generate ID: %w", err)
	}

	cellType := input.Type
	if cellType == "" {
		cellType = "task"
	}
	priority := input.Priority
	if priority == 0 {
		priority = 1
	}

	now := time.Now().UTC().Format(time.RFC3339)
	labels, _ := json.Marshal([]string{})

	_, err = store.DB.Exec(`
		INSERT INTO beads (id, title, description, type, status, priority, parent_id, labels, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'open', ?, ?, ?, ?, ?)`,
		id, input.Title, input.Description, cellType, priority, input.ParentID, string(labels), now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert cell: %w", err)
	}

	return &Cell{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		Type:        cellType,
		Status:      "open",
		Priority:    priority,
		ParentID:    input.ParentID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// QueryCells retrieves cells matching the given filters.
func QueryCells(store *db.Store, q CellQuery) (cells []Cell, err error) {
	query := "SELECT id, title, description, type, status, priority, parent_id, created_at, updated_at, closed_at, close_reason FROM beads WHERE 1=1"
	args := []any{}

	if q.ID != "" {
		query += " AND id LIKE ?"
		args = append(args, q.ID+"%")
	}
	if q.Status != "" {
		query += " AND status = ?"
		args = append(args, q.Status)
	}
	if q.Type != "" {
		query += " AND type = ?"
		args = append(args, q.Type)
	}

	// When Ready is true, filter to cells that are open, not epics, and whose
	// parent (if any) is not in an unfinished state. This mirrors ReadyCell
	// logic but returns multiple results.
	if q.Ready {
		query += ` AND status = 'open'
		  AND type != 'epic'
		  AND (parent_id IS NULL
		       OR NOT EXISTS (
		           SELECT 1 FROM beads p
		           WHERE p.id = beads.parent_id
		             AND p.status IN ('open', 'in_progress', 'blocked')
		       ))`
	}

	query += " ORDER BY priority DESC, created_at DESC"

	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	query += " LIMIT ?"
	args = append(args, limit)

	rows, err := store.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query cells: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			cells = nil
			err = errors.Join(err, fmt.Errorf("close cell rows: %w", closeErr))
		}
	}()

	for rows.Next() {
		var c Cell
		var desc, parentID, closedAt, closeReason *string
		err := rows.Scan(&c.ID, &c.Title, &desc, &c.Type, &c.Status, &c.Priority,
			&parentID, &c.CreatedAt, &c.UpdatedAt, &closedAt, &closeReason)
		if err != nil {
			return nil, fmt.Errorf("scan cell: %w", err)
		}
		if desc != nil {
			c.Description = *desc
		}
		c.ParentID = parentID
		c.ClosedAt = closedAt
		c.CloseReason = closeReason
		cells = append(cells, c)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate cells: %w", rowsErr)
	}

	if cells == nil {
		cells = []Cell{}
	}
	return cells, nil
}

// CloseCell marks a cell as closed with a reason.
func CloseCell(store *db.Store, id, reason string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := store.DB.Exec(
		"UPDATE beads SET status = 'closed', closed_at = ?, close_reason = ?, updated_at = ? WHERE id = ?",
		now, reason, now, id,
	)
	if err != nil {
		return fmt.Errorf("close cell: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cell %q not found", id)
	}
	return nil
}

// UpdateCell updates a cell's mutable fields.
func UpdateCell(store *db.Store, id string, status *string, description *string, priority *int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	sets := []string{"updated_at = ?"}
	args := []any{now}

	if status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *status)
	}
	if description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *description)
	}
	if priority != nil {
		sets = append(sets, "priority = ?")
		args = append(args, *priority)
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE beads SET %s WHERE id = ?", joinStrings(sets, ", "))

	result, err := store.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("update cell: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cell %q not found", id)
	}
	return nil
}

// StartCell sets a cell's status to in_progress.
func StartCell(store *db.Store, id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := store.DB.Exec(
		"UPDATE beads SET status = 'in_progress', updated_at = ? WHERE id = ?",
		now, id,
	)
	if err != nil {
		return fmt.Errorf("start cell: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cell %q not found", id)
	}
	return nil
}

// ReadyCell finds the first unblocked open cell -- one whose parent (if any)
// is not in an open, in_progress, or blocked state. Results are ordered by
// priority DESC so the highest-priority ready cell is returned.
func ReadyCell(store *db.Store) (*Cell, error) {
	// Exclude epics from ready results -- they are containers, not actionable work.
	row := store.DB.QueryRow(`
		SELECT b.id, b.title, b.description, b.type, b.status, b.priority,
		       b.parent_id, b.created_at, b.updated_at
		FROM beads b
		WHERE b.status = 'open'
		  AND b.type != 'epic'
		  AND (b.parent_id IS NULL
		       OR NOT EXISTS (
		           SELECT 1 FROM beads p
		           WHERE p.id = b.parent_id
		             AND p.status IN ('open', 'in_progress', 'blocked')
		       ))
		ORDER BY b.priority DESC, b.created_at ASC
		LIMIT 1
	`)

	var c Cell
	var desc, parentID *string
	err := row.Scan(&c.ID, &c.Title, &desc, &c.Type, &c.Status, &c.Priority,
		&parentID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ready cell: %w", err)
	}
	if desc != nil {
		c.Description = *desc
	}
	c.ParentID = parentID
	return &c, nil
}

func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "cell-" + hex.EncodeToString(b), nil
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
