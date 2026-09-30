package forge

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/unbound-force/replicator/internal/db"
)

// SubtaskPrompt generates a prompt for a spawned subtask agent.
// This is a prompt generator -- it does NOT call an LLM.
func SubtaskPrompt(agentName, beadID, epicID, title string, files []string, sharedContext string) string {
	var sb strings.Builder

	_, _ = fmt.Fprintf(&sb, "# Subtask: %s\n\n", title)
	_, _ = fmt.Fprintf(&sb, "**Agent:** %s\n", agentName)
	_, _ = fmt.Fprintf(&sb, "**Cell ID:** %s\n", beadID)
	_, _ = fmt.Fprintf(&sb, "**Epic ID:** %s\n\n", epicID)

	if len(files) > 0 {
		sb.WriteString("## Files to Modify\n\n")
		for _, f := range files {
			_, _ = fmt.Fprintf(&sb, "- `%s`\n", f)
		}
		sb.WriteString("\n")
	}

	if sharedContext != "" {
		sb.WriteString("## Shared Context\n\n")
		sb.WriteString(sharedContext)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Instructions\n\n")
	sb.WriteString("1. Reserve files before editing using `comms_reserve`\n")
	sb.WriteString("2. Implement the changes described above\n")
	sb.WriteString("3. Write tests for new code\n")
	sb.WriteString("4. Report progress with `forge_progress`\n")
	sb.WriteString("5. Complete with `forge_complete` when done\n")

	return sb.String()
}

// SpawnSubtask prepares metadata for spawning a subtask agent.
// Returns the metadata needed to launch the agent.
func SpawnSubtask(beadID, epicID, title string, files []string, description, sharedContext string) (map[string]any, error) {
	if beadID == "" {
		return nil, fmt.Errorf("bead_id is required")
	}
	if epicID == "" {
		return nil, fmt.Errorf("epic_id is required")
	}
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	result := map[string]any{
		"status":         "ready_to_spawn",
		"bead_id":        beadID,
		"epic_id":        epicID,
		"title":          title,
		"files":          files,
		"description":    description,
		"shared_context": sharedContext,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}

	return result, nil
}

// CompleteSubtask records the completion of a subtask and updates the cell status.
func CompleteSubtask(store *db.Store, beadID, taskResult string, filesTouched []string) (map[string]any, error) {
	if beadID == "" {
		return nil, fmt.Errorf("bead_id is required")
	}

	payload := map[string]any{
		"bead_id":       beadID,
		"task_result":   taskResult,
		"files_touched": filesTouched,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	_, err = store.DB.Exec(
		"INSERT INTO events (type, payload) VALUES (?, ?)",
		"subtask_complete", string(payloadJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("record subtask_complete event: %w", err)
	}

	// Update cell status to closed.
	// The event is already recorded, so a cell update failure is non-fatal
	// but noted in the result for observability.
	now := time.Now().UTC().Format(time.RFC3339)
	_, cellErr := store.DB.Exec(
		"UPDATE beads SET status = 'closed', closed_at = ?, close_reason = 'completed', updated_at = ? WHERE id = ?",
		now, now, beadID,
	)

	result := map[string]any{
		"status":        "completed",
		"bead_id":       beadID,
		"files_touched": filesTouched,
	}
	if cellErr != nil {
		result["cell_update_error"] = cellErr.Error()
	}
	return result, nil
}
