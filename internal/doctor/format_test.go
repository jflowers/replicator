package doctor

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestFormatText_Header(t *testing.T) {
	var buf bytes.Buffer
	results := []CheckResult{
		{Name: "git", Status: "pass", Message: "git version 2.40.0", Duration: 5 * time.Millisecond},
	}

	if err := FormatText(results, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Replicator Doctor") {
		t.Errorf("output missing header 'Replicator Doctor':\n%s", output)
	}
}

func TestFormatText_Indicators(t *testing.T) {
	var buf bytes.Buffer
	results := []CheckResult{
		{Name: "git", Status: "pass", Message: "ok", Duration: time.Millisecond},
		{Name: "dewey", Status: "warn", Message: "not reachable", Duration: time.Millisecond},
		{Name: "config", Status: "fail", Message: "missing", Duration: time.Millisecond},
	}

	if err := FormatText(results, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()

	// In non-TTY mode (bytes.Buffer), indicators should be plain text.
	if !strings.Contains(output, "[PASS]") {
		t.Errorf("output missing [PASS] indicator:\n%s", output)
	}
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("output missing [WARN] indicator:\n%s", output)
	}
	if !strings.Contains(output, "[FAIL]") {
		t.Errorf("output missing [FAIL] indicator:\n%s", output)
	}
}

func TestFormatText_SummaryBox(t *testing.T) {
	var buf bytes.Buffer
	results := []CheckResult{
		{Name: "a", Status: "pass", Message: "ok", Duration: time.Millisecond},
		{Name: "b", Status: "pass", Message: "ok", Duration: time.Millisecond},
		{Name: "c", Status: "warn", Message: "meh", Duration: time.Millisecond},
	}

	if err := FormatText(results, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()

	// Summary should contain counts.
	if !strings.Contains(output, "2 passed") {
		t.Errorf("summary missing '2 passed':\n%s", output)
	}
	if !strings.Contains(output, "1 warnings") {
		t.Errorf("summary missing '1 warnings':\n%s", output)
	}
	if !strings.Contains(output, "0 failed") {
		t.Errorf("summary missing '0 failed':\n%s", output)
	}
}

func TestFormatText_AllPassMessage(t *testing.T) {
	var buf bytes.Buffer
	results := []CheckResult{
		{Name: "a", Status: "pass", Message: "ok", Duration: time.Millisecond},
	}

	if err := FormatText(results, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Everything looks good") {
		t.Errorf("output missing success message:\n%s", output)
	}
}

func TestFormatText_FailureMessage(t *testing.T) {
	var buf bytes.Buffer
	results := []CheckResult{
		{Name: "a", Status: "fail", Message: "broken", Duration: time.Millisecond},
	}

	if err := FormatText(results, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "replicator setup") {
		t.Errorf("output missing fix suggestion:\n%s", output)
	}
}

func TestFormatText_NoANSI(t *testing.T) {
	var buf bytes.Buffer
	results := []CheckResult{
		{Name: "git", Status: "pass", Message: "ok", Duration: time.Millisecond},
		{Name: "db", Status: "fail", Message: "err", Duration: time.Millisecond},
	}

	if err := FormatText(results, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "\x1b[") {
		t.Errorf("output contains ANSI escape sequences in non-TTY mode:\n%s", output)
	}
}

func TestFormatText_Duration(t *testing.T) {
	var buf bytes.Buffer
	results := []CheckResult{
		{Name: "git", Status: "pass", Message: "ok", Duration: 42 * time.Millisecond},
	}

	if err := FormatText(results, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "42ms") {
		t.Errorf("output missing duration '42ms':\n%s", output)
	}
}

func TestFormatText_ReturnsWriterError(t *testing.T) {
	errWrite := errors.New("write sentinel")
	err := FormatText([]CheckResult{{Name: "git", Status: "pass", Message: "ok"}}, failingWriter{err: errWrite})
	if !errors.Is(err, errWrite) {
		t.Fatalf("error %v does not preserve %v", err, errWrite)
	}
}
