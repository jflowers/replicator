package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/memory"
	commstools "github.com/unbound-force/replicator/internal/tools/comms"
	forgetools "github.com/unbound-force/replicator/internal/tools/forge"
	memorytools "github.com/unbound-force/replicator/internal/tools/memory"
	"github.com/unbound-force/replicator/internal/tools/org"
	"github.com/unbound-force/replicator/internal/tools/registry"
)

type docsWriteCloser struct {
	writer   io.Writer
	writeErr error
	closeErr error
	closed   int
}

func (w *docsWriteCloser) Write(p []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return w.writer.Write(p)
}

func (w *docsWriteCloser) Close() error {
	w.closed++
	return w.closeErr
}

func buildFullRegistry(t *testing.T) *registry.Registry {
	t.Helper()
	store, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	reg := registry.New()
	org.Register(reg, store)
	commstools.Register(reg, store)
	forgetools.Register(reg, store)
	memClient := memory.NewClient("http://localhost:3333/mcp/")
	memorytools.Register(reg, memClient)
	return reg
}

func TestWriteDocs_ContainsAllTools(t *testing.T) {
	reg := buildFullRegistry(t)

	var buf bytes.Buffer
	if err := writeDocs(&buf, reg); err != nil {
		t.Fatalf("writeDocs: %v", err)
	}

	output := buf.String()

	// Verify all registered tools appear in the output.
	for _, tool := range reg.List() {
		if !strings.Contains(output, tool.Name) {
			t.Errorf("tool %q not found in docs output", tool.Name)
		}
	}
}

func TestWriteDocs_HasCategoryHeaders(t *testing.T) {
	reg := buildFullRegistry(t)

	var buf bytes.Buffer
	if err := writeDocs(&buf, reg); err != nil {
		t.Fatalf("writeDocs: %v", err)
	}
	output := buf.String()

	for _, header := range []string{"## Org", "## Comms", "## Forge", "## Memory"} {
		if !strings.Contains(output, header) {
			t.Errorf("missing category header: %q", header)
		}
	}
}

func TestWriteDocs_ToolCount(t *testing.T) {
	reg := buildFullRegistry(t)

	if reg.Count() < 50 {
		t.Errorf("expected at least 50 tools, got %d", reg.Count())
	}

	var buf bytes.Buffer
	if err := writeDocs(&buf, reg); err != nil {
		t.Fatalf("writeDocs: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, "tools registered") {
		t.Error("output missing tool count line")
	}
}

func TestWriteDocsAndClose_PreservesWriteAndCloseErrors(t *testing.T) {
	reg := buildFullRegistry(t)
	errWrite := errors.New("write sentinel")
	errClose := errors.New("close sentinel")

	for _, test := range []struct {
		name     string
		writeErr error
		closeErr error
		want     []error
	}{
		{name: "success"},
		{name: "write", writeErr: errWrite, want: []error{errWrite}},
		{name: "close", closeErr: errClose, want: []error{errClose}},
		{name: "write and close", writeErr: errWrite, closeErr: errClose, want: []error{errWrite, errClose}},
	} {
		t.Run(test.name, func(t *testing.T) {
			closer := &docsWriteCloser{writer: io.Discard, writeErr: test.writeErr, closeErr: test.closeErr}
			err := writeDocsAndClose(closer, reg)
			if closer.closed != 1 {
				t.Fatalf("close calls = %d, want 1", closer.closed)
			}
			if len(test.want) == 0 && err != nil {
				t.Fatalf("writeDocsAndClose: %v", err)
			}
			for _, want := range test.want {
				if !errors.Is(err, want) {
					t.Errorf("error %v does not preserve %v", err, want)
				}
			}
			if test.writeErr != nil && !strings.Contains(err.Error(), "write docs") {
				t.Errorf("error %q missing write context", err)
			}
			if test.closeErr != nil && !strings.Contains(err.Error(), "close docs") {
				t.Errorf("error %q missing close context", err)
			}
		})
	}
}
