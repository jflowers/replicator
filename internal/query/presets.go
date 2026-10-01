// Package query provides preset database queries for the replicator CLI.
//
// Each preset is a named SQL query that produces a styled table.
// Presets cover common observability needs: agent activity, cell status,
// forge completion rates, and recent events.
package query

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/ui"
)

type queryRows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close() error
}

type countRow interface {
	Scan(...any) error
}

// Preset names.
const (
	AgentActivity24h    = "agent_activity_24h"
	CellsByStatus       = "cells_by_status"
	ForgeCompletionRate = "forge_completion_rate"
	RecentEvents        = "recent_events"
)

// ListPresets returns all available preset names.
func ListPresets() []string {
	return []string{
		AgentActivity24h,
		CellsByStatus,
		ForgeCompletionRate,
		RecentEvents,
	}
}

// Run executes a preset query and writes the results to w.
func Run(store *db.Store, presetName string, w io.Writer) error {
	switch presetName {
	case AgentActivity24h:
		return runAgentActivity(store, w)
	case CellsByStatus:
		return runCellsByStatus(store, w)
	case ForgeCompletionRate:
		return runForgeCompletionRate(store, w)
	case RecentEvents:
		return runRecentEvents(store, w)
	default:
		return fmt.Errorf("unknown preset: %q (use --list to see available presets)", presetName)
	}
}

func runAgentActivity(store *db.Store, w io.Writer) error {
	rows, err := store.DB.Query(`
		SELECT COALESCE(agent_name, '(unknown)') as agent, COUNT(*) as events
		FROM events
		WHERE created_at >= datetime('now', '-24 hours')
		GROUP BY agent_name
		ORDER BY events DESC
		LIMIT 20`)
	if err != nil {
		return fmt.Errorf("query agent activity: %w", err)
	}
	styles := ui.NewStyles(w)
	tableRows, err := readTableRows(rows, func(rows queryRows) ([]string, error) {
		var agent string
		var events int
		if err := rows.Scan(&agent, &events); err != nil {
			return nil, err
		}
		return []string{agent, strconv.Itoa(events)}, nil
	})
	if err != nil {
		return err
	}

	if len(tableRows) == 0 {
		return writeLine(w, styles.Dim.Render("(no activity in last 24 hours)"))
	}

	t := ui.NewTable(styles, []string{"AGENT", "EVENTS (24h)"}, tableRows)
	return writeLine(w, t.String())
}

func runCellsByStatus(store *db.Store, w io.Writer) error {
	rows, err := store.DB.Query(`
		SELECT status, type, COUNT(*) as count
		FROM beads
		GROUP BY status, type
		ORDER BY status, type`)
	if err != nil {
		return fmt.Errorf("query cells by status: %w", err)
	}
	styles := ui.NewStyles(w)
	tableRows, err := readTableRows(rows, func(rows queryRows) ([]string, error) {
		var status, cellType string
		var n int
		if err := rows.Scan(&status, &cellType, &n); err != nil {
			return nil, err
		}
		return []string{status, cellType, strconv.Itoa(n)}, nil
	})
	if err != nil {
		return err
	}

	if len(tableRows) == 0 {
		return writeLine(w, styles.Dim.Render("(no cells)"))
	}

	t := ui.NewTable(styles, []string{"STATUS", "TYPE", "COUNT"}, tableRows)
	return writeLine(w, t.String())
}

func runForgeCompletionRate(store *db.Store, w io.Writer) error {
	styles := ui.NewStyles(w)

	// Count completed vs total forge events.
	total, completed, err := scanForgeCounts(
		store.DB.QueryRow(`SELECT COUNT(*) FROM events WHERE type LIKE 'forge_%'`),
		func() countRow {
			return store.DB.QueryRow(`SELECT COUNT(*) FROM events WHERE type = 'forge_complete'`)
		},
	)
	if err != nil {
		return err
	}

	if err := writeLine(w, styles.Bold.Render("Forge Completion Rate:")); err != nil {
		return err
	}
	if err := writeFormat(w, "  Total forge events:     %d\n", total); err != nil {
		return err
	}
	if err := writeFormat(w, "  Completed:              %d\n", completed); err != nil {
		return err
	}
	if total > 0 {
		rate := float64(completed) / float64(total) * 100
		return writeFormat(w, "  Completion rate:        %.1f%%\n", rate)
	} else {
		return writeLine(w, styles.Dim.Render("  Completion rate:        N/A (no forge events)"))
	}
}

func runRecentEvents(store *db.Store, w io.Writer) error {
	rows, err := store.DB.Query(`
		SELECT id, type, project_key, created_at
		FROM events
		ORDER BY created_at DESC
		LIMIT 20`)
	if err != nil {
		return fmt.Errorf("query recent events: %w", err)
	}
	styles := ui.NewStyles(w)
	tableRows, err := readTableRows(rows, func(rows queryRows) ([]string, error) {
		var id int
		var eventType, projectKey, createdAt string
		if err := rows.Scan(&id, &eventType, &projectKey, &createdAt); err != nil {
			return nil, err
		}
		return []string{strconv.Itoa(id), eventType, projectKey, createdAt}, nil
	})
	if err != nil {
		return err
	}

	if len(tableRows) == 0 {
		return writeLine(w, styles.Dim.Render("(no events)"))
	}

	t := ui.NewTable(styles, []string{"ID", "TYPE", "PROJECT", "CREATED"}, tableRows)
	return writeLine(w, t.String())
}

func writeLine(w io.Writer, values ...any) error {
	if _, err := fmt.Fprintln(w, values...); err != nil {
		return fmt.Errorf("write preset output: %w", err)
	}
	return nil
}

func writeFormat(w io.Writer, format string, values ...any) error {
	if _, err := fmt.Fprintf(w, format, values...); err != nil {
		return fmt.Errorf("write preset output: %w", err)
	}
	return nil
}

func readTableRows(rows queryRows, scan func(queryRows) ([]string, error)) (tableRows [][]string, err error) {
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			tableRows = nil
			err = errors.Join(err, fmt.Errorf("close rows: %w", closeErr))
		}
	}()

	for rows.Next() {
		row, scanErr := scan(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan rows: %w", scanErr)
		}
		tableRows = append(tableRows, row)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate rows: %w", rowsErr)
	}
	return tableRows, nil
}

func scanForgeCounts(totalRow countRow, completedRow func() countRow) (int, int, error) {
	var total int
	if err := totalRow.Scan(&total); err != nil {
		return 0, 0, fmt.Errorf("scan total forge events: %w", err)
	}
	var completed int
	if err := completedRow().Scan(&completed); err != nil {
		return 0, 0, fmt.Errorf("scan completed forge events: %w", err)
	}
	return total, completed, nil
}
