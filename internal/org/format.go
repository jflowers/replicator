package org

import (
	"fmt"
	"io"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/unbound-force/replicator/internal/ui"
)

// FormatCells renders a list of cells as a styled terminal table.
//
// Uses lipgloss/table with status-based coloring: green for open,
// yellow for in_progress, red for blocked, gray for closed.
// Falls back to plain text when output is not a TTY.
// Prints a dim "No cells found" message when the list is empty.
func FormatCells(cells []Cell, w io.Writer) error {
	styles := ui.NewStyles(w)

	if len(cells) == 0 {
		if _, err := fmt.Fprintln(w, styles.Dim.Render("No cells found")); err != nil {
			return fmt.Errorf("write cells output: %w", err)
		}
		return nil
	}

	headers := []string{"ID", "TITLE", "STATUS", "TYPE", "PRIORITY"}

	rows := make([][]string, len(cells))
	for i, c := range cells {
		// Truncate ID to 8 chars for readability.
		id := c.ID
		if len(id) > 8 {
			id = id[:8]
		}

		// Truncate long titles to keep table compact.
		title := c.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		rows[i] = []string{
			id,
			title,
			c.Status,
			c.Type,
			strconv.Itoa(c.Priority),
		}
	}

	t := ui.NewTable(styles, headers, rows)

	// Apply status-based coloring via StyleFunc.
	// Row -1 is the header row (lipgloss/table convention).
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return styles.Bold
		}
		// Color the STATUS column (index 2) based on value.
		if col == 2 && row >= 0 && row < len(cells) {
			switch cells[row].Status {
			case "open":
				return styles.Pass
			case "in_progress":
				return styles.Warn
			case "blocked":
				return styles.Fail
			case "closed":
				return styles.Dim
			}
		}
		return lipgloss.NewStyle()
	})

	// Constrain width for consistent terminal rendering.
	t.Width(80)

	if _, err := fmt.Fprintln(w, t.String()); err != nil {
		return fmt.Errorf("write cells output: %w", err)
	}
	return nil
}
