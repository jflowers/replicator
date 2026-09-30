package org

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSync(t *testing.T) {
	store := testStore(t)
	dir := t.TempDir()

	// Initialize a git repo in the temp directory.
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		// Need an initial commit for git commit to work.
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}

	// Create some cells.
	if _, err := CreateCell(store, CreateCellInput{Title: "Task A"}); err != nil {
		t.Fatalf("CreateCell: %v", err)
	}
	if _, err := CreateCell(store, CreateCellInput{Title: "Task B", Type: "bug"}); err != nil {
		t.Fatalf("CreateCell: %v", err)
	}

	// Run sync.
	if err := Sync(store, dir); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	// Verify cells.json was created.
	cellsPath := filepath.Join(dir, ".uf", "replicator", "cells.json")
	data, err := os.ReadFile(cellsPath)
	if err != nil {
		t.Fatalf("read cells.json: %v", err)
	}

	var cells []Cell
	if err := json.Unmarshal(data, &cells); err != nil {
		t.Fatalf("unmarshal cells.json: %v", err)
	}
	if len(cells) != 2 {
		t.Errorf("expected 2 cells in JSON, got %d", len(cells))
	}

	// Verify git commit was made.
	logCmd := exec.Command("git", "log", "--oneline", "-1")
	logCmd.Dir = dir
	out, err := logCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log: %v\n%s", err, out)
	}
	if !contains(string(out), "hive sync") {
		t.Errorf("git log = %q, want commit message containing 'hive sync'", string(out))
	}
}

func TestSync_CreatesHiveDir(t *testing.T) {
	store := testStore(t)
	dir := t.TempDir()

	// Initialize git repo.
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}

	// Sync with no cells -- should still create .uf/replicator/cells.json.
	if err := Sync(store, dir); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	hiveDir := filepath.Join(dir, ".uf", "replicator")
	if _, err := os.Stat(hiveDir); os.IsNotExist(err) {
		t.Error(".uf/replicator directory was not created")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
