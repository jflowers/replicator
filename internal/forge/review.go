package forge

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/unbound-force/replicator/internal/db"
)

// Review generates a review prompt for a completed subtask.
// This is a prompt generator -- it does NOT call an LLM.
func Review(projectKey, epicID, taskID string, filesTouched []string) string {
	var sb strings.Builder

	sb.WriteString("## Code Review\n\n")
	_, _ = fmt.Fprintf(&sb, "**Project:** %s\n", projectKey)
	_, _ = fmt.Fprintf(&sb, "**Epic:** %s\n", epicID)
	_, _ = fmt.Fprintf(&sb, "**Task:** %s\n\n", taskID)

	if len(filesTouched) > 0 {
		sb.WriteString("### Files Modified\n\n")
		for _, f := range filesTouched {
			_, _ = fmt.Fprintf(&sb, "- `%s`\n", f)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Review Checklist\n\n")
	sb.WriteString("1. Does the code follow project conventions?\n")
	sb.WriteString("2. Are all error paths handled?\n")
	sb.WriteString("3. Are tests comprehensive and passing?\n")
	sb.WriteString("4. Are there any security concerns?\n")
	sb.WriteString("5. Is the code well-documented?\n\n")

	sb.WriteString("### Verdict\n\n")
	sb.WriteString("Respond with: `approved` or `needs_changes` with specific issues.\n")

	return sb.String()
}

// ReviewFeedback records review feedback and tracks attempts (max 3).
// After 3 rejections, the task is marked as failed.
func ReviewFeedback(store *db.Store, projectKey, taskID, workerID, status, issues, summary string) (map[string]any, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	if workerID == "" {
		return nil, fmt.Errorf("worker_id is required")
	}
	if status != "approved" && status != "needs_changes" {
		return nil, fmt.Errorf("status must be 'approved' or 'needs_changes'")
	}

	// Count previous review attempts for this task.
	var attemptCount int
	if err := store.DB.QueryRow(
		"SELECT COUNT(*) FROM events WHERE type = 'review_feedback' AND json_extract(payload, '$.task_id') = ?",
		taskID,
	).Scan(&attemptCount); err != nil {
		return nil, fmt.Errorf("count review attempts: %w", err)
	}

	attempt := attemptCount + 1

	payload := map[string]any{
		"project_key": projectKey,
		"task_id":     taskID,
		"worker_id":   workerID,
		"status":      status,
		"issues":      issues,
		"summary":     summary,
		"attempt":     attempt,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	_, err = store.DB.Exec(
		"INSERT INTO events (type, payload, project_key) VALUES (?, ?, ?)",
		"review_feedback", string(payloadJSON), projectKey,
	)
	if err != nil {
		return nil, fmt.Errorf("record review feedback: %w", err)
	}

	result := map[string]any{
		"status":  status,
		"attempt": attempt,
		"task_id": taskID,
	}

	// After 3 rejections, mark as failed.
	if status == "needs_changes" && attempt >= 3 {
		result["failed"] = true
		result["message"] = "task failed after 3 review rejections"
	}

	return result, nil
}

// AdversarialReview generates a VDD-style adversarial review prompt.
// This is a prompt generator -- it does NOT call an LLM.
func AdversarialReview(diff, testOutput string) string {
	var sb strings.Builder

	sb.WriteString("## Adversarial Code Review (Sarcasmotron Mode)\n\n")
	sb.WriteString("You are a hyper-critical code reviewer with zero tolerance for slop.\n")
	sb.WriteString("Review the following diff with fresh eyes and hostile intent.\n\n")

	sb.WriteString("### Diff\n\n")
	sb.WriteString("```diff\n")
	sb.WriteString(diff)
	sb.WriteString("\n```\n\n")

	if testOutput != "" {
		sb.WriteString("### Test Output\n\n")
		sb.WriteString("```\n")
		sb.WriteString(testOutput)
		sb.WriteString("\n```\n\n")
	}

	sb.WriteString("### Review Criteria\n\n")
	sb.WriteString("- Logic errors and edge cases\n")
	sb.WriteString("- Security vulnerabilities\n")
	sb.WriteString("- Performance issues\n")
	sb.WriteString("- Missing error handling\n")
	sb.WriteString("- Test coverage gaps\n")
	sb.WriteString("- Code style violations\n\n")

	sb.WriteString("### Verdict\n\n")
	sb.WriteString("Respond with one of:\n")
	sb.WriteString("- `APPROVED`: Code is solid\n")
	sb.WriteString("- `NEEDS_CHANGES`: Real issues found (list them)\n")
	sb.WriteString("- `HALLUCINATING`: You invented issues (code is excellent)\n")

	return sb.String()
}

// EvaluationPrompt generates a self-evaluation prompt for a completed subtask.
// This is a prompt generator -- it does NOT call an LLM.
func EvaluationPrompt(beadID, title string, filesTouched []string) string {
	var sb strings.Builder

	sb.WriteString("## Self-Evaluation\n\n")
	_, _ = fmt.Fprintf(&sb, "**Task:** %s\n", title)
	_, _ = fmt.Fprintf(&sb, "**Cell ID:** %s\n\n", beadID)

	if len(filesTouched) > 0 {
		sb.WriteString("### Files Modified\n\n")
		for _, f := range filesTouched {
			_, _ = fmt.Fprintf(&sb, "- `%s`\n", f)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Evaluation Criteria\n\n")
	sb.WriteString("Rate each criterion 1-5:\n\n")
	sb.WriteString("1. **Correctness**: Does the code do what was asked?\n")
	sb.WriteString("2. **Completeness**: Are all acceptance criteria met?\n")
	sb.WriteString("3. **Code Quality**: Is the code clean and well-structured?\n")
	sb.WriteString("4. **Test Coverage**: Are tests comprehensive?\n")
	sb.WriteString("5. **Documentation**: Is the code well-documented?\n\n")

	sb.WriteString("### Output Format\n\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{
  "scores": {"correctness": 5, "completeness": 5, "quality": 5, "tests": 5, "docs": 5},
  "summary": "Brief summary of what was done",
  "issues": ["Any remaining issues"]
}`)
	sb.WriteString("\n```\n")

	return sb.String()
}

// Broadcast sends a context update to all agents working on the same epic.
func Broadcast(store *db.Store, projectPath, agentName, epicID, message, importance string, filesAffected []string) error {
	if epicID == "" {
		return fmt.Errorf("epic_id is required")
	}
	if message == "" {
		return fmt.Errorf("message is required")
	}

	if importance == "" {
		importance = "info"
	}

	payload := map[string]any{
		"agent_name":     agentName,
		"epic_id":        epicID,
		"message":        message,
		"importance":     importance,
		"files_affected": filesAffected,
		"project_path":   projectPath,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	_, err = store.DB.Exec(
		"INSERT INTO events (type, payload, project_key) VALUES (?, ?, ?)",
		"forge_broadcast", string(payloadJSON), projectPath,
	)
	if err != nil {
		return fmt.Errorf("record broadcast event: %w", err)
	}

	return nil
}
