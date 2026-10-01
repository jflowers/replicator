package doctor

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/unbound-force/replicator/internal/ui"
)

// FormatText renders health check results as styled terminal output.
//
// Uses lipgloss for styling with automatic NO_COLOR and pipe detection
// following the UF doctor formatting pattern. When output is not a TTY,
// falls back to plain-text indicators ([PASS], [WARN], [FAIL]).
func FormatText(results []CheckResult, w io.Writer) error {
	styles := ui.NewStyles(w)
	var output strings.Builder
	writeLine := func(values ...any) {
		_, _ = fmt.Fprintln(&output, values...) // strings.Builder writes cannot fail.
	}
	writeFormat := func(format string, values ...any) {
		_, _ = fmt.Fprintf(&output, format, values...) // strings.Builder writes cannot fail.
	}

	// Header with stethoscope emoji.
	writeLine(styles.Title.Render("🩺 Replicator Doctor"))
	writeLine()

	// Tally counters for the summary box.
	var passed, warned, failed int

	for _, r := range results {
		indicator := styles.Indicator(r.Status)
		name := fmt.Sprintf("%-14s", r.Name)
		duration := styles.Dim.Render(fmt.Sprintf("(%s)", r.Duration.Round(time.Millisecond)))

		writeFormat("  %s %s %s %s\n", indicator, name, r.Message, duration)

		switch r.Status {
		case "pass":
			passed++
		case "warn":
			warned++
		case "fail":
			failed++
		}
	}

	writeLine()

	// Boxed summary with emoji counters.
	summaryContent := fmt.Sprintf("  ✅ %d passed  ⚠️  %d warnings  ❌ %d failed",
		passed, warned, failed)
	writeLine(styles.Box.Render(summaryContent))

	// Contextual completion message.
	if failed == 0 && warned == 0 {
		writeLine(styles.Pass.Render("🎉 Everything looks good!"))
	} else if failed > 0 {
		writeLine(styles.Dim.Render("  Run 'replicator setup' to fix common issues."))
	} else {
		writeLine(styles.Dim.Render("  All critical checks passed."))
	}

	if _, err := io.WriteString(w, output.String()); err != nil {
		return fmt.Errorf("write doctor report: %w", err)
	}
	return nil
}
