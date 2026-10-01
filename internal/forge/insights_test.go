package forge

import (
	"encoding/json"
	"testing"
)

func TestGetStrategyInsights_Empty(t *testing.T) {
	store := testStore(t)

	result, err := GetStrategyInsights(store, "build auth")
	if err != nil {
		t.Fatalf("GetStrategyInsights: %v", err)
	}
	if result["task"] != "build auth" {
		t.Errorf("task = %v, want %q", result["task"], "build auth")
	}
	if result["recommendation"] != "feature-based" {
		t.Errorf("recommendation = %v, want %q", result["recommendation"], "feature-based")
	}
}

func TestGetStrategyInsights_WithData(t *testing.T) {
	store := testStore(t)

	// Insert some outcomes.
	outcomes := []struct {
		strategy string
		success  bool
	}{
		{"file-based", true},
		{"file-based", true},
		{"file-based", false},
		{"feature-based", true},
		{"feature-based", false},
	}

	for _, o := range outcomes {
		payload, _ := json.Marshal(map[string]any{
			"bead_id":  "cell-1",
			"strategy": o.strategy,
			"success":  o.success,
		})
		if _, err := store.DB.Exec("INSERT INTO events (type, payload) VALUES (?, ?)", "forge_outcome", string(payload)); err != nil {
			t.Fatalf("insert outcome: %v", err)
		}
	}

	result, err := GetStrategyInsights(store, "task")
	if err != nil {
		t.Fatalf("GetStrategyInsights: %v", err)
	}

	strategies := result["strategies"].(map[string]any)
	fileBased := strategies["file-based"].(map[string]any)
	if fileBased["total"] != 3 {
		t.Errorf("file-based total = %v, want 3", fileBased["total"])
	}
	if fileBased["success"] != 2 {
		t.Errorf("file-based success = %v, want 2", fileBased["success"])
	}

	// file-based has 66.7% success, feature-based has 50% -- file-based should be recommended.
	if result["recommendation"] != "file-based" {
		t.Errorf("recommendation = %v, want %q", result["recommendation"], "file-based")
	}
}

func TestGetFileInsights_Empty(t *testing.T) {
	store := testStore(t)

	result, err := GetFileInsights(store, []string{})
	if err != nil {
		t.Fatalf("GetFileInsights: %v", err)
	}
	insights := result["insights"].(map[string]any)
	if len(insights) != 0 {
		t.Errorf("expected empty insights, got %d", len(insights))
	}
}

func TestGetFileInsights_WithData(t *testing.T) {
	store := testStore(t)

	// Insert outcomes touching specific files.
	payloads := []map[string]any{
		{"bead_id": "c1", "success": false, "files_touched": []string{"auth.go"}, "error_count": 2},
		{"bead_id": "c2", "success": true, "files_touched": []string{"auth.go", "db.go"}, "error_count": 0},
		{"bead_id": "c3", "success": false, "files_touched": []string{"db.go"}, "error_count": 1},
	}
	for _, p := range payloads {
		data, _ := json.Marshal(p)
		if _, err := store.DB.Exec("INSERT INTO events (type, payload) VALUES (?, ?)", "forge_outcome", string(data)); err != nil {
			t.Fatalf("insert outcome: %v", err)
		}
	}

	result, err := GetFileInsights(store, []string{"auth.go", "db.go"})
	if err != nil {
		t.Fatalf("GetFileInsights: %v", err)
	}

	// The insights map contains *fileStats but is typed as map[string]any.
	// We need to check the values through the map[string]any interface.
	insightsRaw := result["insights"]
	// Marshal and re-parse to get consistent types.
	data, _ := json.Marshal(insightsRaw)
	var insights map[string]map[string]int
	if err := json.Unmarshal(data, &insights); err != nil {
		t.Fatalf("unmarshal insights: %v", err)
	}

	authStats := insights["auth.go"]
	if authStats == nil {
		t.Fatal("expected insights for auth.go")
	}
	if authStats["total"] != 2 {
		t.Errorf("auth.go total = %d, want 2", authStats["total"])
	}
	if authStats["failures"] != 1 {
		t.Errorf("auth.go failures = %d, want 1", authStats["failures"])
	}
}

func TestGetPatternInsights_Empty(t *testing.T) {
	store := testStore(t)

	result, err := GetPatternInsights(store)
	if err != nil {
		t.Fatalf("GetPatternInsights: %v", err)
	}
	if result["total_outcomes"] != 0 {
		t.Errorf("total_outcomes = %v, want 0", result["total_outcomes"])
	}
}

func TestGetPatternInsights_WithData(t *testing.T) {
	store := testStore(t)

	// Insert failure outcomes with criteria.
	payloads := []map[string]any{
		{"bead_id": "c1", "success": false, "criteria": []string{"type_error", "test_failure"}, "error_count": 1, "retry_count": 0},
		{"bead_id": "c2", "success": false, "criteria": []string{"type_error"}, "error_count": 2, "retry_count": 1},
		{"bead_id": "c3", "success": true, "criteria": []string{"clean"}, "error_count": 0, "retry_count": 0},
	}
	for _, p := range payloads {
		data, _ := json.Marshal(p)
		if _, err := store.DB.Exec("INSERT INTO events (type, payload) VALUES (?, ?)", "forge_outcome", string(data)); err != nil {
			t.Fatalf("insert outcome: %v", err)
		}
	}

	result, err := GetPatternInsights(store)
	if err != nil {
		t.Fatalf("GetPatternInsights: %v", err)
	}
	if result["total_outcomes"] != 3 {
		t.Errorf("total_outcomes = %v, want 3", result["total_outcomes"])
	}
	if result["total_failures"] != 2 {
		t.Errorf("total_failures = %v, want 2", result["total_failures"])
	}
}
