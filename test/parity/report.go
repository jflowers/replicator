package parity

import (
	"fmt"
	"io"
)

// ToolResult captures the parity test outcome for a single MCP tool.
type ToolResult struct {
	Name        string
	Match       bool
	Differences []Difference
}

// GenerateReport writes a human-readable parity report to w and returns any
// writer error without changing the report format.
//
// Output format:
//
//	Parity Report
//	=============
//	org_cells       ✓
//	org_create      ✓
//	hivemind_find   ✗  $.content[0].text: expected object, got string
//	---
//	22/23 tools match (95.7%)
func GenerateReport(results []ToolResult, w io.Writer) error {
	if _, err := fmt.Fprintln(w, "Parity Report"); err != nil {
		return fmt.Errorf("write report header: %w", err)
	}
	if _, err := fmt.Fprintln(w, "============="); err != nil {
		return fmt.Errorf("write report separator: %w", err)
	}

	matched := 0
	for _, r := range results {
		if r.Match {
			if _, err := fmt.Fprintf(w, "%-30s ✓\n", r.Name); err != nil {
				return fmt.Errorf("write report row: %w", err)
			}
			matched++
		} else {
			// Show first difference inline, rest on subsequent lines.
			for i, d := range r.Differences {
				if i == 0 {
					if _, err := fmt.Fprintf(w, "%-30s ✗  %s: expected %s, got %s\n",
						r.Name, d.Path, d.ExpectedType, d.ActualType); err != nil {
						return fmt.Errorf("write report difference: %w", err)
					}
				} else {
					if _, err := fmt.Fprintf(w, "%-30s    %s: expected %s, got %s\n",
						"", d.Path, d.ExpectedType, d.ActualType); err != nil {
						return fmt.Errorf("write report difference: %w", err)
					}
				}
			}
		}
	}

	if _, err := fmt.Fprintln(w, "---"); err != nil {
		return fmt.Errorf("write report footer: %w", err)
	}

	total := len(results)
	if total == 0 {
		if _, err := fmt.Fprintln(w, "0/0 tools match (0.0%)"); err != nil {
			return fmt.Errorf("write report summary: %w", err)
		}
		return nil
	}

	pct := float64(matched) / float64(total) * 100
	if _, err := fmt.Fprintf(w, "%d/%d tools match (%.1f%%)\n", matched, total, pct); err != nil {
		return fmt.Errorf("write report summary: %w", err)
	}
	return nil
}
