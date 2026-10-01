package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/unbound-force/replicator/internal/config"
	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/stats"
)

// runStats queries the database and prints statistics.
func runStats(cfg *config.Config) (err error) {
	store, err := db.Open(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close database: %w", closeErr))
		}
	}()

	return stats.Run(store, os.Stdout)
}
