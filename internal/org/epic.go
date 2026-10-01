package org

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/unbound-force/replicator/internal/db"
)

// SubtaskInput defines a subtask to create under an epic.
type SubtaskInput struct {
	Title    string   `json:"title"`
	Priority int      `json:"priority,omitempty"`
	Files    []string `json:"files,omitempty"`
}

// CreateEpicInput defines the input for creating an epic with subtasks.
type CreateEpicInput struct {
	EpicTitle       string         `json:"epic_title"`
	EpicDescription string         `json:"epic_description,omitempty"`
	Subtasks        []SubtaskInput `json:"subtasks"`
}

// CreateEpic atomically creates an epic cell and its subtasks in a single transaction.
func CreateEpic(store *db.Store, input CreateEpicInput) (*Cell, []Cell, error) {
	tx, err := store.DB.Begin()
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	var epic *Cell
	var subtasks []Cell
	err = runEpicTransaction(func() error {
		epicID, err := generateID()
		if err != nil {
			return fmt.Errorf("generate epic ID: %w", err)
		}

		now := time.Now().UTC().Format(time.RFC3339)
		labels, _ := json.Marshal([]string{})

		_, err = tx.Exec(`
		INSERT INTO beads (id, title, description, type, status, priority, labels, created_at, updated_at)
		VALUES (?, ?, ?, 'epic', 'open', 1, ?, ?, ?)`,
			epicID, input.EpicTitle, input.EpicDescription, string(labels), now, now,
		)
		if err != nil {
			return fmt.Errorf("insert epic: %w", err)
		}

		epic = &Cell{
			ID:          epicID,
			Title:       input.EpicTitle,
			Description: input.EpicDescription,
			Type:        "epic",
			Status:      "open",
			Priority:    1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		subtasks = make([]Cell, 0, len(input.Subtasks))
		for _, st := range input.Subtasks {
			subID, err := generateID()
			if err != nil {
				return fmt.Errorf("generate subtask ID: %w", err)
			}

			priority := st.Priority
			if priority == 0 {
				priority = 1
			}

			// Store files list in the metadata JSON field.
			metadata := "{}"
			if len(st.Files) > 0 {
				filesJSON, _ := json.Marshal(st.Files)
				metadata = fmt.Sprintf(`{"files":%s}`, string(filesJSON))
			}

			_, err = tx.Exec(`
			INSERT INTO beads (id, title, type, status, priority, parent_id, labels, metadata, created_at, updated_at)
			VALUES (?, ?, 'task', 'open', ?, ?, ?, ?, ?, ?)`,
				subID, st.Title, priority, epicID, string(labels), metadata, now, now,
			)
			if err != nil {
				return fmt.Errorf("insert subtask %q: %w", st.Title, err)
			}

			parentID := epicID
			subtasks = append(subtasks, Cell{
				ID:        subID,
				Title:     st.Title,
				Type:      "task",
				Status:    "open",
				Priority:  priority,
				ParentID:  &parentID,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}
		return nil
	}, func() error {
		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("commit epic: %w", commitErr)
		}
		return nil
	}, tx.Rollback)
	if err != nil {
		return nil, nil, err
	}

	return epic, subtasks, nil
}

func runEpicTransaction(work, commit, rollback func() error) (err error) {
	committed := false
	defer func() {
		rollbackErr := rollback()
		if rollbackErr == nil || (committed && errors.Is(rollbackErr, sql.ErrTxDone)) {
			return
		}
		err = errors.Join(err, fmt.Errorf("rollback epic: %w", rollbackErr))
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
