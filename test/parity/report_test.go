//go:build parity

package parity

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type failAfterWriter struct {
	failCall int
	calls    int
	err      error
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == w.failCall {
		return 0, w.err
	}
	return len(p), nil
}

func TestGenerateReport_ReturnsWriterFailures(t *testing.T) {
	errWrite := errors.New("write sentinel")
	results := []ToolResult{
		{Name: "match", Match: true},
		{Name: "difference", Differences: []Difference{{Path: "$.x", ExpectedType: "string", ActualType: "number"}, {Path: "$.y", ExpectedType: "object", ActualType: "array"}}},
	}

	for _, test := range []struct {
		name    string
		call    int
		context string
	}{
		{name: "header", call: 1, context: "write report header"},
		{name: "separator", call: 2, context: "write report separator"},
		{name: "matching row", call: 3, context: "write report row"},
		{name: "first difference", call: 4, context: "write report difference"},
		{name: "subsequent difference", call: 5, context: "write report difference"},
		{name: "footer", call: 6, context: "write report footer"},
		{name: "summary", call: 7, context: "write report summary"},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := GenerateReport(results, &failAfterWriter{failCall: test.call, err: errWrite})
			if !errors.Is(err, errWrite) {
				t.Fatalf("error %v does not preserve %v", err, errWrite)
			}
			if !strings.Contains(err.Error(), test.context) {
				t.Errorf("error %q does not contain %q", err, test.context)
			}
		})
	}
}

func TestGenerateReport_ReturnsEmptySummaryWriterFailure(t *testing.T) {
	errWrite := errors.New("write sentinel")
	err := GenerateReport(nil, &failAfterWriter{failCall: 4, err: errWrite})
	if !errors.Is(err, errWrite) || !strings.Contains(err.Error(), "write report summary") {
		t.Fatalf("GenerateReport error = %v, want contextual summary error", err)
	}
}

func TestGenerateReport_PreservesOutputFormat(t *testing.T) {
	results := []ToolResult{{Name: "match", Match: true}}
	var output bytes.Buffer
	if err := GenerateReport(results, &output); err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	want := fmt.Sprintf("Parity Report\n=============\n%-30s ✓\n---\n1/1 tools match (100.0%%)\n", "match")
	if output.String() != want {
		t.Errorf("report = %q, want %q", output.String(), want)
	}
}
