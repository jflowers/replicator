package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/unbound-force/replicator/internal/agentkit"
	"github.com/unbound-force/replicator/internal/ui"
)

func initCmd() *cobra.Command {
	var pathFlag string
	var forceFlag bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a project directory for project operations",
		Long: `Creates a .uf/replicator/ directory with an empty cells.json and scaffolds
the agent kit into .opencode/ (commands, skills, and agent definitions).

Idempotent — safe to run multiple times. Existing agent kit files are
skipped unless --force is used.

This is the per-repo initialization command. It does not require the
global database (replicator setup) or any external services.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(pathFlag, forceFlag)
		},
	}
	cmd.Flags().StringVar(&pathFlag, "path", ".", "Target directory for initialization")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Overwrite existing agent kit files")
	return cmd
}

// runInit creates the .uf/replicator/ directory, seeds cells.json,
// scaffolds the agent kit into .opencode/, and creates or updates the
// DCP configuration (.opencode/dcp.jsonc) for protect tag support.
// Uses styled output: green for created, dim for skipped, yellow for
// overwritten.
func runInit(targetDir string, force bool) error {
	styles := ui.NewStyles(os.Stdout)
	replicatorDir := filepath.Join(targetDir, ".uf", "replicator")

	// Create .uf/replicator/ directory (idempotent).
	cellsPath := filepath.Join(replicatorDir, "cells.json")
	if _, err := os.Stat(cellsPath); err != nil {
		// cells.json doesn't exist — create directory and file.
		if err := os.MkdirAll(replicatorDir, 0o755); err != nil {
			return fmt.Errorf("create .uf/replicator directory: %w", err)
		}
		if err := os.WriteFile(cellsPath, []byte("[]\n"), 0o644); err != nil {
			return fmt.Errorf("write cells.json: %w", err)
		}
		fmt.Println(styles.Pass.Render("created .uf/replicator/cells.json"))
	} else {
		fmt.Println(styles.Dim.Render("skipped .uf/replicator/cells.json (exists)"))
	}

	// Scaffold agent kit into .opencode/.
	if force {
		fmt.Println(styles.Warn.Render("--force: existing agent kit files will be overwritten"))
	}

	results, err := agentkit.Scaffold(targetDir, force)
	if err != nil {
		return fmt.Errorf("scaffold agent kit: %w", err)
	}

	for _, r := range results {
		switch r.Action {
		case "created":
			fmt.Println(styles.Pass.Render(fmt.Sprintf("created .opencode/%s", r.Path)))
		case "skipped":
			fmt.Println(styles.Dim.Render(fmt.Sprintf("skipped .opencode/%s (exists)", r.Path)))
		case "overwritten":
			fmt.Println(styles.Warn.Render(fmt.Sprintf("overwritten .opencode/%s", r.Path)))
		}
	}

	// Scaffold DCP config for protect tag support.
	dcpResult, err := agentkit.ScaffoldDCP(targetDir)
	if err != nil {
		return fmt.Errorf("scaffold DCP config: %w", err)
	}
	switch dcpResult.Action {
	case "created":
		fmt.Println(styles.Pass.Render(fmt.Sprintf("created .opencode/%s", dcpResult.Path)))
	case "skipped":
		fmt.Println(styles.Dim.Render(fmt.Sprintf("skipped .opencode/%s (exists)", dcpResult.Path)))
	case "overwritten":
		fmt.Println(styles.Warn.Render(fmt.Sprintf("overwritten .opencode/%s (added protectTags)", dcpResult.Path)))
	}

	return nil
}
