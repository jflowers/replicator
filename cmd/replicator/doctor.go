package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/unbound-force/replicator/internal/config"
	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/doctor"
)

// runDoctor executes health checks and prints styled results.
func runDoctor(cfg *config.Config) (err error) {
	store, err := db.Open(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close database: %w", closeErr))
		}
	}()

	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	results, err := doctor.Run(store, cfg, projectDir)
	if err != nil {
		return fmt.Errorf("run checks: %w", err)
	}

	return doctor.FormatText(results, os.Stdout)
}
