package org

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/unbound-force/replicator/internal/db"
)

// Sync serializes all cells to JSON and commits them to git.
//
// Writes cells to <projectPath>/.uf/replicator/cells.json, then runs
// `git add .uf/replicator/ && git commit -m "hive sync"` in the project directory.
func Sync(store *db.Store, projectPath string) error {
	cells, err := QueryCells(store, CellQuery{Limit: 10000})
	if err != nil {
		return fmt.Errorf("query cells for sync: %w", err)
	}

	hiveDir := filepath.Join(projectPath, ".uf", "replicator")
	if err := os.MkdirAll(hiveDir, 0o755); err != nil {
		return fmt.Errorf("create .uf/replicator dir: %w", err)
	}

	data, err := json.MarshalIndent(cells, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cells: %w", err)
	}

	cellsPath := filepath.Join(hiveDir, "cells.json")
	// #nosec G306 -- cells.json is repository state intended for collaborator access.
	if err := os.WriteFile(cellsPath, data, 0o644); err != nil {
		return fmt.Errorf("write cells.json: %w", err)
	}

	// Stage and commit the .uf/replicator directory.
	addCmd := exec.Command("git", "add", ".uf/replicator/")
	addCmd.Dir = projectPath
	if out, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w\n%s", err, out)
	}

	commitCmd := exec.Command("git", "commit", "-m", "hive sync", "--allow-empty")
	commitCmd.Dir = projectPath
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %w\n%s", err, out)
	}

	return nil
}
