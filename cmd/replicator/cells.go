package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/unbound-force/replicator/internal/config"
	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/org"
)

// jsonOutput controls whether cells are printed as JSON or a styled table.
var jsonOutput bool

// listCells queries and prints hive cells.
// With --json, outputs indented JSON. Otherwise, renders a styled table.
func listCells(cfg *config.Config) (err error) {
	store, err := db.Open(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close database: %w", closeErr))
		}
	}()

	cells, err := org.QueryCells(store, org.CellQuery{})
	if err != nil {
		return fmt.Errorf("query cells: %w", err)
	}

	if jsonOutput {
		out, err := json.MarshalIndent(cells, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal cells: %w", err)
		}
		fmt.Println(string(out))
		return nil
	}

	return org.FormatCells(cells, os.Stdout)
}
