package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/unbound-force/replicator/internal/config"
	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/ui"
)

// runSetup creates the config directory, initializes the database, and
// verifies git is available. Uses styled output for pass/fail indicators.
func runSetup() error {
	styles := ui.NewStyles(os.Stdout)

	// 1. Create config directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "uf", "replicator")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	fmt.Printf("%s Config directory: %s\n", styles.Pass.Render("✓"), configDir)

	// 2. Initialize database.
	cfg := config.Load()
	store, err := db.Open(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	if err := store.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	fmt.Printf("%s Database: %s\n", styles.Pass.Render("✓"), cfg.DatabasePath)

	// 3. Verify git.
	cmd := exec.Command("git", "--version")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s Git: not found (%v)\n", styles.Fail.Render("✗"), err)
		fmt.Println("  Install git: https://git-scm.com/downloads")
	} else {
		fmt.Printf("%s Git: %s\n", styles.Pass.Render("✓"), strings.TrimSpace(string(out)))
	}

	fmt.Println()
	fmt.Println("Setup complete. Run 'replicator doctor' to verify all checks pass.")
	return nil
}
