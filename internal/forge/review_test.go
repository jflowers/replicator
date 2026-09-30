package forge

import (
	"strings"
	"testing"
)

func TestReview(t *testing.T) {
	prompt := Review("proj-1", "epic-1", "task-1", []string{"auth.go", "auth_test.go"})

	checks := []string{"proj-1", "epic-1", "task-1", "auth.go", "auth_test.go", "approved", "needs_changes"}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("review prompt missing %q", check)
		}
	}
}

func TestReview_NoFiles(t *testing.T) {
	prompt := Review("proj", "epic", "task", nil)
	if strings.Contains(prompt, "Files Modified") {
		t.Error("should not show files section when empty")
	}
}

func TestReviewFeedback_Approved(t *testing.T) {
	store := testStore(t)

	result, err := ReviewFeedback(store, "proj", "task-1", "worker-1", "approved", "", "looks good")
	if err != nil {
		t.Fatalf("ReviewFeedback: %v", err)
	}
	if result["status"] != "approved" {
		t.Errorf("status = %v, want %q", result["status"], "approved")
	}
	if result["attempt"] != 1 {
		t.Errorf("attempt = %v, want 1", result["attempt"])
	}
}

func TestReviewFeedback_NeedsChanges(t *testing.T) {
	store := testStore(t)

	result, err := ReviewFeedback(store, "proj", "task-1", "worker-1", "needs_changes", "fix error handling", "")
	if err != nil {
		t.Fatalf("ReviewFeedback: %v", err)
	}
	if result["status"] != "needs_changes" {
		t.Errorf("status = %v, want %q", result["status"], "needs_changes")
	}
}

func TestReviewFeedback_FailsAfter3(t *testing.T) {
	store := testStore(t)

	// Submit 3 rejections.
	for i := 0; i < 3; i++ {
		result, err := ReviewFeedback(store, "proj", "task-fail", "worker-1", "needs_changes", "issues", "")
		if err != nil {
			t.Fatalf("ReviewFeedback attempt %d: %v", i+1, err)
		}
		if i == 2 {
			if result["failed"] != true {
				t.Error("expected failed=true after 3 rejections")
			}
		}
	}
}

func TestReviewFeedback_MissingTaskID(t *testing.T) {
	store := testStore(t)
	_, err := ReviewFeedback(store, "proj", "", "worker", "approved", "", "")
	if err == nil {
		t.Error("expected error for missing task_id")
	}
}

func TestReviewFeedback_MissingWorkerID(t *testing.T) {
	store := testStore(t)
	_, err := ReviewFeedback(store, "proj", "task", "", "approved", "", "")
	if err == nil {
		t.Error("expected error for missing worker_id")
	}
}

func TestReviewFeedback_InvalidStatus(t *testing.T) {
	store := testStore(t)
	_, err := ReviewFeedback(store, "proj", "task", "worker", "invalid", "", "")
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestAdversarialReview(t *testing.T) {
	prompt := AdversarialReview("+ new code\n- old code", "PASS: all tests")

	checks := []string{"Sarcasmotron", "new code", "old code", "PASS: all tests", "APPROVED", "NEEDS_CHANGES", "HALLUCINATING"}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("adversarial review missing %q", check)
		}
	}
}

func TestAdversarialReview_NoTestOutput(t *testing.T) {
	prompt := AdversarialReview("+ code", "")
	if strings.Contains(prompt, "Test Output") {
		t.Error("should not show test output section when empty")
	}
}

func TestEvaluationPrompt(t *testing.T) {
	prompt := EvaluationPrompt("cell-1", "Implement auth", []string{"auth.go"})

	checks := []string{"cell-1", "Implement auth", "auth.go", "Correctness", "Completeness"}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("evaluation prompt missing %q", check)
		}
	}
}

func TestEvaluationPrompt_NoFiles(t *testing.T) {
	prompt := EvaluationPrompt("cell-1", "Task", nil)
	if strings.Contains(prompt, "Files Modified") {
		t.Error("should not show files section when empty")
	}
}

func TestBroadcast(t *testing.T) {
	store := testStore(t)

	err := Broadcast(store, "/tmp/proj", "worker-1", "epic-1", "API changed", "warning", []string{"api.go"})
	if err != nil {
		t.Fatalf("Broadcast: %v", err)
	}

	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM events WHERE type = 'forge_broadcast'").Scan(&count); err != nil {
		t.Fatalf("scan broadcast count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 broadcast event, got %d", count)
	}
}

func TestBroadcast_DefaultImportance(t *testing.T) {
	store := testStore(t)

	err := Broadcast(store, "/tmp", "agent", "epic-1", "update", "", nil)
	if err != nil {
		t.Fatalf("Broadcast: %v", err)
	}

	// Verify default importance was set.
	var payload string
	if err := store.DB.QueryRow("SELECT payload FROM events WHERE type = 'forge_broadcast'").Scan(&payload); err != nil {
		t.Fatalf("scan broadcast payload: %v", err)
	}
	if !strings.Contains(payload, `"info"`) {
		t.Errorf("expected default importance 'info' in payload: %s", payload)
	}
}

func TestBroadcast_MissingEpicID(t *testing.T) {
	store := testStore(t)
	err := Broadcast(store, "/tmp", "agent", "", "msg", "", nil)
	if err == nil {
		t.Error("expected error for missing epic_id")
	}
}

func TestBroadcast_MissingMessage(t *testing.T) {
	store := testStore(t)
	err := Broadcast(store, "/tmp", "agent", "epic-1", "", "", nil)
	if err == nil {
		t.Error("expected error for missing message")
	}
}
