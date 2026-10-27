package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/unbound-force/replicator/internal/agentkit"
	"github.com/unbound-force/replicator/internal/ui"
)

type initOptions struct {
	targetDir    string
	skipExisting bool
	force        bool
}

func initCmd() *cobra.Command {
	var pathFlag string
	var skipExistingFlag bool
	var forceFlag bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a project directory for project operations",
		Long: `Creates a .uf/replicator/ directory with an empty cells.json and scaffolds
the agent kit into .opencode/ (commands, skills, and agent definitions).

Idempotent — safe to run multiple times. Agent kit files are smart-synced
against the embedded kit content, which is canonical: missing files are
created, byte-identical files are skipped, and files that drifted from the
embedded content are updated to match it. Use --skip-existing in
repositories with intentionally customized kit files — existing files are
then preserved and only missing files are created.

The .uf/replicator/cells.json state file and the DCP configuration
(.opencode/dcp.jsonc) keep their existing behavior: cells.json is created
only when missing, and the DCP configuration is created or updated only
when it lacks protect tag support.

Smart-sync repairs drift only relative to the embedded content in the
installed binary — rebuild or upgrade the binary to pick up newer kit
content.

This is the per-repo initialization command. It does not require the
global database (replicator setup) or any external services.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(initOptions{
				targetDir:    pathFlag,
				skipExisting: skipExistingFlag,
				force:        forceFlag,
			})
		},
	}
	cmd.Flags().StringVar(&pathFlag, "path", ".", "Target directory for initialization")
	cmd.Flags().BoolVar(&skipExistingFlag, "skip-existing", false, "Preserve existing agent kit files instead of updating drifted ones")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Overwrite existing agent kit files")
	if err := cmd.Flags().MarkDeprecated("force", "smart-sync is now the default; this flag is accepted but ignored"); err != nil {
		panic(fmt.Errorf("mark --force deprecated: %w", err))
	}
	return cmd
}

// runInit creates the .uf/replicator/ directory, seeds cells.json,
// scaffolds the agent kit into .opencode/, and creates or updates the
// DCP configuration (.opencode/dcp.jsonc) for protect tag support.
// Uses styled output: green for created, dim for skipped, yellow for
// updated or overwritten files.
func runInit(opts initOptions) error {
	styles := ui.NewStyles(os.Stdout)
	replicatorDir := filepath.Join(opts.targetDir, ".uf", "replicator")

	// Create .uf/replicator/ directory (idempotent).
	cellsPath := filepath.Join(replicatorDir, "cells.json")
	if _, err := os.Stat(cellsPath); err != nil {
		// cells.json doesn't exist — create directory and file.
		if err := os.MkdirAll(replicatorDir, 0o755); err != nil {
			return fmt.Errorf("create .uf/replicator directory: %w", err)
		}
		// #nosec G306 -- cells.json is repository state intended for collaborator access.
		if err := os.WriteFile(cellsPath, []byte("[]\n"), 0o644); err != nil {
			return fmt.Errorf("write cells.json: %w", err)
		}
		fmt.Println(styles.Pass.Render("created .uf/replicator/cells.json"))
	} else {
		fmt.Println(styles.Dim.Render("skipped .uf/replicator/cells.json (exists)"))
	}

	// Scaffold agent kit into .opencode/.
	mode := agentkit.ScaffoldModeSync
	if opts.skipExisting {
		mode = agentkit.ScaffoldModeSkipExisting
	}

	results, err := agentkit.Scaffold(opts.targetDir, mode)
	if err != nil {
		return fmt.Errorf("scaffold agent kit: %w", err)
	}

	for _, r := range results {
		switch r.Action {
		case "created":
			fmt.Println(styles.Pass.Render(fmt.Sprintf("created .opencode/%s", r.Path)))
		case "skipped":
			fmt.Println(styles.Dim.Render(fmt.Sprintf("skipped .opencode/%s (exists)", r.Path)))
		case "updated":
			fmt.Println(styles.Warn.Render(fmt.Sprintf("updated .opencode/%s", r.Path)))
		}
	}

	// Scaffold DCP config for protect tag support.
	dcpResult, err := agentkit.ScaffoldDCP(opts.targetDir)
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
