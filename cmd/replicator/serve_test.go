package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type errorCloser struct {
	err   error
	calls int
}

func (c *errorCloser) Close() error {
	c.calls++
	return c.err
}

func TestSetupLogger_CreatesLogFile(t *testing.T) {
	// Run setupLogger in a temp directory so it creates
	// .uf/replicator/replicator.log there.
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	logger, closer := setupLogger()
	if closer != nil {
		t.Cleanup(func() {
			if err := closer.Close(); err != nil {
				t.Errorf("close logger: %v", err)
			}
		})
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	logPath := filepath.Join(dir, ".uf", "replicator", "replicator.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file not created: %v", err)
	}
}

func TestSetupLogger_Truncates(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	// First call: write a marker to the log file.
	logDir := filepath.Join(dir, ".uf", "replicator")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	logPath := filepath.Join(logDir, "replicator.log")
	if err := os.WriteFile(logPath, []byte("MARKER_SHOULD_BE_GONE"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Second call: setupLogger uses os.Create which truncates.
	_, closer := setupLogger()
	if closer != nil {
		if err := closer.Close(); err != nil {
			t.Fatalf("close logger: %v", err)
		}
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) == "MARKER_SHOULD_BE_GONE" {
		t.Error("log file was not truncated; marker still present")
	}
}

func TestSetupLogger_ReadOnlyDir(t *testing.T) {
	// Verify that a read-only directory doesn't cause a panic.
	// setupLogger should fall back to stderr-only logging.
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	// Create the .uf/replicator dir as read-only so file creation fails.
	logDir := filepath.Join(dir, ".uf", "replicator")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Chmod(logDir, 0o444); err != nil {
		t.Fatalf("Chmod read-only: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(logDir, 0o755); err != nil {
			t.Errorf("restore directory permissions: %v", err)
		}
	})

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	// Should not panic — falls back to stderr-only.
	logger, closer := setupLogger()
	if closer != nil {
		t.Cleanup(func() {
			if err := closer.Close(); err != nil {
				t.Errorf("close logger: %v", err)
			}
		})
	}
	if logger == nil {
		t.Fatal("expected non-nil logger even when log file creation fails")
	}
}

func TestRunAndClose_PreservesServeAndCloseErrors(t *testing.T) {
	errServe := errors.New("serve sentinel")
	errClose := errors.New("close sentinel")

	for _, test := range []struct {
		name     string
		serveErr error
		closeErr error
		want     []error
	}{
		{name: "success"},
		{name: "serve", serveErr: errServe, want: []error{errServe}},
		{name: "close", closeErr: errClose, want: []error{errClose}},
		{name: "serve and close", serveErr: errServe, closeErr: errClose, want: []error{errServe, errClose}},
	} {
		t.Run(test.name, func(t *testing.T) {
			closer := &errorCloser{err: test.closeErr}
			err := runAndClose(func() error { return test.serveErr }, io.Closer(closer))
			if closer.calls != 1 {
				t.Fatalf("close calls = %d, want 1", closer.calls)
			}
			if len(test.want) == 0 && err != nil {
				t.Fatalf("runAndClose: %v", err)
			}
			for _, want := range test.want {
				if !errors.Is(err, want) {
					t.Errorf("error %v does not preserve %v", err, want)
				}
			}
			if test.serveErr != nil && !strings.Contains(err.Error(), "serve") {
				t.Errorf("error %q missing serve context", err)
			}
			if test.closeErr != nil && !strings.Contains(err.Error(), "close log") {
				t.Errorf("error %q missing close context", err)
			}
		})
	}
}
