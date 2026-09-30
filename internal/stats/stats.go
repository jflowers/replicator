// Package stats provides database statistics for the replicator CLI.
//
// Queries the events and cells tables to produce a human-readable summary
// of system activity and work item status. Uses lipgloss styling for
// section headers and visual hierarchy.
package stats

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/ui"
)

// eventCount holds a type and its count from the events table.
type eventCount struct {
	Type  string
	Count int
}

// Run queries the database for statistics and writes a formatted report.
func Run(store *db.Store, w io.Writer) error {
	styles := ui.NewStyles(w)
	var output strings.Builder
	writeLine := func(values ...any) { _, _ = fmt.Fprintln(&output, values...) }
	writeFormat := func(format string, values ...any) { _, _ = fmt.Fprintf(&output, format, values...) }

	// Events by type.
	eventCounts, err := queryEventCounts(store)
	if err != nil {
		return fmt.Errorf("query event counts: %w", err)
	}

	// Recent events (last 24h).
	recentCount, err := queryRecentEvents(store)
	if err != nil {
		return fmt.Errorf("query recent events: %w", err)
	}

	// Cells by status.
	cellCounts, err := queryCellCounts(store)
	if err != nil {
		return fmt.Errorf("query cell counts: %w", err)
	}

	// Total cells.
	totalCells := 0
	for _, c := range cellCounts {
		totalCells += c.Count
	}

	// Print report with styled headers.
	writeLine(styles.Title.Render("📊 Replicator Stats"))
	writeLine()

	writeLine(styles.Bold.Render("Events by Type:"))
	if len(eventCounts) == 0 {
		writeLine(styles.Dim.Render("  (no events)"))
	}
	for _, ec := range eventCounts {
		writeFormat("  %-30s %d\n", ec.Type, ec.Count)
	}
	writeLine()

	writeFormat("%s %d events\n", styles.Bold.Render("Recent Activity (24h):"), recentCount)
	writeLine()

	writeLine(styles.Bold.Render(fmt.Sprintf("Cells (%d total):", totalCells)))
	if len(cellCounts) == 0 {
		writeLine(styles.Dim.Render("  (no cells)"))
	}
	for _, cc := range cellCounts {
		writeFormat("  %-15s %d\n", cc.Type, cc.Count)
	}

	if _, err := io.WriteString(w, output.String()); err != nil {
		return fmt.Errorf("write stats report: %w", err)
	}
	return nil
}

func queryEventCounts(store *db.Store) (counts []eventCount, err error) {
	rows, err := store.DB.Query(`
		SELECT type, COUNT(*) as count
		FROM events
		GROUP BY type
		ORDER BY count DESC`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			counts = nil
			err = combineRowsCloseError(err, closeErr)
		}
	}()

	for rows.Next() {
		var ec eventCount
		if err := rows.Scan(&ec.Type, &ec.Count); err != nil {
			return nil, err
		}
		counts = append(counts, ec)
	}
	return counts, rows.Err()
}

func queryRecentEvents(store *db.Store) (int, error) {
	var count int
	err := store.DB.QueryRow(`
		SELECT COUNT(*)
		FROM events
		WHERE created_at >= datetime('now', '-24 hours')`).Scan(&count)
	return count, err
}

func queryCellCounts(store *db.Store) (counts []eventCount, err error) {
	rows, err := store.DB.Query(`
		SELECT status, COUNT(*) as count
		FROM beads
		GROUP BY status
		ORDER BY count DESC`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			counts = nil
			err = combineRowsCloseError(err, closeErr)
		}
	}()

	for rows.Next() {
		var ec eventCount
		if err := rows.Scan(&ec.Type, &ec.Count); err != nil {
			return nil, err
		}
		counts = append(counts, ec)
	}
	return counts, rows.Err()
}

func combineRowsCloseError(primary, closeErr error) error {
	if closeErr == nil {
		return primary
	}
	return errors.Join(primary, fmt.Errorf("close query rows: %w", closeErr))
}
