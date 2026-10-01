package forge

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/unbound-force/replicator/internal/db"
)

// Progress records a progress event for a subtask.
func Progress(store *db.Store, projectKey, agentName, beadID, status string, progressPercent int, message string, filesTouched []string) error {
	if beadID == "" {
		return fmt.Errorf("bead_id is required")
	}

	payload := map[string]any{
		"agent_name":       agentName,
		"bead_id":          beadID,
		"status":           status,
		"progress_percent": progressPercent,
		"message":          message,
		"files_touched":    filesTouched,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	_, err = store.DB.Exec(
		"INSERT INTO events (type, payload, project_key) VALUES (?, ?, ?)",
		"forge_progress", string(payloadJSON), projectKey,
	)
	if err != nil {
		return fmt.Errorf("record progress event: %w", err)
	}

	return nil
}

// Complete marks a subtask as complete, recording the event and updating the cell.
func Complete(store *db.Store, projectKey, agentName, beadID, summary string, filesTouched []string, evaluation string, skipVerification, skipReview bool) (map[string]any, error) {
	if beadID == "" {
		return nil, fmt.Errorf("bead_id is required")
	}

	payload := map[string]any{
		"agent_name":        agentName,
		"bead_id":           beadID,
		"summary":           summary,
		"files_touched":     filesTouched,
		"evaluation":        evaluation,
		"skip_verification": skipVerification,
		"skip_review":       skipReview,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	_, err = store.DB.Exec(
		"INSERT INTO events (type, payload, project_key) VALUES (?, ?, ?)",
		"forge_complete", string(payloadJSON), projectKey,
	)
	if err != nil {
		return nil, fmt.Errorf("record complete event: %w", err)
	}

	// Update cell status.
	// The event is already recorded, so a cell update failure is non-fatal
	// but noted in the result for observability.
	now := time.Now().UTC().Format(time.RFC3339)
	_, cellErr := store.DB.Exec(
		"UPDATE beads SET status = 'closed', closed_at = ?, close_reason = ?, updated_at = ? WHERE id = ?",
		now, summary, now, beadID,
	)

	result := map[string]any{
		"status":        "completed",
		"bead_id":       beadID,
		"agent_name":    agentName,
		"files_touched": filesTouched,
	}
	if cellErr != nil {
		result["cell_update_error"] = cellErr.Error()
	}
	return result, nil
}

// Status aggregates subtask statuses for an epic from events.
func Status(store *db.Store, epicID, projectKey string) (result map[string]any, err error) {
	if epicID == "" {
		return nil, fmt.Errorf("epic_id is required")
	}

	// Query child cells of the epic.
	rows, err := store.DB.Query(
		"SELECT id, title, status FROM beads WHERE parent_id = ?", epicID,
	)
	if err != nil {
		return nil, fmt.Errorf("query subtasks: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			result = nil
			err = errors.Join(err, fmt.Errorf("close subtask rows: %w", closeErr))
		}
	}()

	var subtasks []map[string]string
	counts := map[string]int{
		"open":        0,
		"in_progress": 0,
		"blocked":     0,
		"closed":      0,
	}

	for rows.Next() {
		var id, title, status string
		if err := rows.Scan(&id, &title, &status); err != nil {
			return nil, fmt.Errorf("scan subtask: %w", err)
		}
		subtasks = append(subtasks, map[string]string{
			"id":     id,
			"title":  title,
			"status": status,
		})
		counts[status]++
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate subtasks: %w", rowsErr)
	}

	total := len(subtasks)
	if subtasks == nil {
		subtasks = []map[string]string{}
	}

	return map[string]any{
		"epic_id":     epicID,
		"project_key": projectKey,
		"total":       total,
		"counts":      counts,
		"subtasks":    subtasks,
	}, nil
}

// RecordOutcome persists a subtask outcome for implicit feedback scoring.
func RecordOutcome(store *db.Store, beadID string, durationMs int, success bool, strategy string, filesTouched []string, errorCount, retryCount int, criteria []string) error {
	if beadID == "" {
		return fmt.Errorf("bead_id is required")
	}

	payload := map[string]any{
		"bead_id":       beadID,
		"duration_ms":   durationMs,
		"success":       success,
		"strategy":      strategy,
		"files_touched": filesTouched,
		"error_count":   errorCount,
		"retry_count":   retryCount,
		"criteria":      criteria,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	_, err = store.DB.Exec(
		"INSERT INTO events (type, payload) VALUES (?, ?)",
		"forge_outcome", string(payloadJSON),
	)
	if err != nil {
		return fmt.Errorf("record outcome event: %w", err)
	}

	return nil
}
