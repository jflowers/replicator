package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/unbound-force/replicator/internal/agentkit"
)

// embeddedKitContent returns the canonical embedded content for an agent
// kit file by scaffolding the kit into a scratch directory. relPath is
// the kit-relative path with slash separators (e.g. "commands/forge.md").
func embeddedKitContent(t *testing.T, relPath string) []byte {
	t.Helper()
	dir := t.TempDir()
	if _, err := agentkit.Scaffold(dir, agentkit.ScaffoldModeSync); err != nil {
		t.Fatalf("scaffold reference kit: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".opencode", filepath.FromSlash(relPath)))
	if err != nil {
		t.Fatalf("read reference %s: %v", relPath, err)
	}
	return data
}

// newInitTestRoot returns a cobra root command wired with the init
// subcommand, mirroring main.go's wiring, so init flag parsing can be
// exercised end to end in tests.
func newInitTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "replicator"}
	root.AddCommand(initCmd())
	return root
}

func TestRunInit_FreshDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := runInit(initOptions{targetDir: dir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	// Verify .uf/replicator/ directory created.
	replicatorDir := filepath.Join(dir, ".uf", "replicator")
	info, err := os.Stat(replicatorDir)
	if err != nil {
		t.Fatalf(".uf/replicator/ not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".uf/replicator/ is not a directory")
	}

	// Verify cells.json created with empty array.
	cellsPath := filepath.Join(replicatorDir, "cells.json")
	data, err := os.ReadFile(cellsPath)
	if err != nil {
		t.Fatalf("cells.json not created: %v", err)
	}
	if string(data) != "[]\n" {
		t.Errorf("cells.json content = %q, want %q", string(data), "[]\n")
	}

	// Verify agent kit files created (16 total: 1 cells.json + 15 agent kit).
	agentKitFiles := []string{
		".opencode/commands/forge.md",
		".opencode/commands/org.md",
		".opencode/commands/inbox.md",
		".opencode/commands/forge-status.md",
		".opencode/commands/handoff.md",
		".opencode/skills/always-on-guidance/SKILL.md",
		".opencode/skills/forge-coordination/SKILL.md",
		".opencode/skills/replicator-cli/SKILL.md",
		".opencode/skills/testing-patterns/SKILL.md",
		".opencode/skills/system-design/SKILL.md",
		".opencode/skills/learning-systems/SKILL.md",
		".opencode/skills/forge-global/SKILL.md",
		".opencode/agents/coordinator.md",
		".opencode/agents/worker.md",
		".opencode/agents/background-worker.md",
	}
	for _, rel := range agentKitFiles {
		full := filepath.Join(dir, rel)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("expected agent kit file %s to exist: %v", rel, err)
		}
	}

	// Verify DCP config was created with protectTags.
	dcpPath := filepath.Join(dir, ".opencode", "dcp.jsonc")
	dcpData, err := os.ReadFile(dcpPath)
	if err != nil {
		t.Fatalf("dcp.jsonc not created: %v", err)
	}
	if !strings.Contains(string(dcpData), "protectTags") {
		t.Error("dcp.jsonc missing protectTags")
	}
	if !strings.Contains(string(dcpData), "$schema") {
		t.Error("dcp.jsonc missing $schema")
	}
}

func TestRunInit_UpdatesDriftedFiles(t *testing.T) {
	dir := t.TempDir()

	embedded := embeddedKitContent(t, "commands/forge.md")

	forgePath := filepath.Join(dir, ".opencode", "commands", "forge.md")
	if err := os.MkdirAll(filepath.Dir(forgePath), 0o755); err != nil {
		t.Fatalf("setup MkdirAll: %v", err)
	}
	if err := os.WriteFile(forgePath, []byte("# stale local content\n"), 0o600); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	if err := runInit(initOptions{targetDir: dir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	data, err := os.ReadFile(forgePath)
	if err != nil {
		t.Fatalf("read forge.md after init: %v", err)
	}
	if string(data) != string(embedded) {
		t.Errorf("drifted forge.md did not converge to embedded content: got %d bytes, want %d bytes", len(data), len(embedded))
	}
}

func TestRunInit_AgentKitSkipsExisting(t *testing.T) {
	dir := t.TempDir()

	embedded := embeddedKitContent(t, "commands/forge.md")

	forgePath := filepath.Join(dir, ".opencode", "commands", "forge.md")
	if err := os.MkdirAll(filepath.Dir(forgePath), 0o755); err != nil {
		t.Fatalf("setup MkdirAll: %v", err)
	}
	if err := os.WriteFile(forgePath, embedded, 0o600); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	pinned := time.Unix(1000000000, 0)
	if err := os.Chtimes(forgePath, pinned, pinned); err != nil {
		t.Fatalf("setup Chtimes: %v", err)
	}
	before, err := os.Stat(forgePath)
	if err != nil {
		t.Fatalf("setup Stat: %v", err)
	}

	if err := runInit(initOptions{targetDir: dir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	after, err := os.Stat(forgePath)
	if err != nil {
		t.Fatalf("stat forge.md after init: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("byte-identical forge.md was rewritten by default: mtime changed from %v to %v", before.ModTime(), after.ModTime())
	}

	skipDir := t.TempDir()
	driftedPath := filepath.Join(skipDir, ".opencode", "commands", "forge.md")
	if err := os.MkdirAll(filepath.Dir(driftedPath), 0o755); err != nil {
		t.Fatalf("setup MkdirAll: %v", err)
	}
	custom := []byte("# my custom forge\n")
	if err := os.WriteFile(driftedPath, custom, 0o600); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	if err := runInit(initOptions{targetDir: skipDir, skipExisting: true}); err != nil {
		t.Fatalf("runInit with skip-existing: %v", err)
	}

	data, err := os.ReadFile(driftedPath)
	if err != nil {
		t.Fatalf("read forge.md after init: %v", err)
	}
	if string(data) != string(custom) {
		t.Errorf("drifted forge.md must be preserved under --skip-existing: got %q, want %q", string(data), string(custom))
	}
}

func TestRunInit_SkipExistingFlag(t *testing.T) {
	dir := t.TempDir()

	orgPath := filepath.Join(dir, ".opencode", "commands", "org.md")
	if err := os.MkdirAll(filepath.Dir(orgPath), 0o755); err != nil {
		t.Fatalf("setup MkdirAll: %v", err)
	}
	custom := []byte("# my custom org\n")
	if err := os.WriteFile(orgPath, custom, 0o600); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	if err := runInit(initOptions{targetDir: dir, skipExisting: true}); err != nil {
		t.Fatalf("runInit with skip-existing: %v", err)
	}

	data, err := os.ReadFile(orgPath)
	if err != nil {
		t.Fatalf("read org.md after init: %v", err)
	}
	if string(data) != string(custom) {
		t.Errorf("custom org.md was not preserved under --skip-existing: got %q, want %q", string(data), string(custom))
	}

	inboxPath := filepath.Join(dir, ".opencode", "commands", "inbox.md")
	if _, err := os.Stat(inboxPath); err != nil {
		t.Errorf("expected agent kit file inbox.md to be created under --skip-existing: %v", err)
	}
}

func TestRunInit_ForceIsNoOp(t *testing.T) {
	dir := t.TempDir()

	embeddedForge := embeddedKitContent(t, "commands/forge.md")
	embeddedOrg := embeddedKitContent(t, "commands/org.md")

	forgePath := filepath.Join(dir, ".opencode", "commands", "forge.md")
	if err := os.MkdirAll(filepath.Dir(forgePath), 0o755); err != nil {
		t.Fatalf("setup MkdirAll: %v", err)
	}
	if err := os.WriteFile(forgePath, embeddedForge, 0o600); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	pinned := time.Unix(1000000000, 0)
	if err := os.Chtimes(forgePath, pinned, pinned); err != nil {
		t.Fatalf("setup Chtimes: %v", err)
	}
	before, err := os.Stat(forgePath)
	if err != nil {
		t.Fatalf("setup Stat: %v", err)
	}

	orgPath := filepath.Join(dir, ".opencode", "commands", "org.md")
	if err := os.WriteFile(orgPath, []byte("# my custom org\n"), 0o600); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	if err := runInit(initOptions{targetDir: dir, force: true}); err != nil {
		t.Fatalf("runInit with force: %v", err)
	}

	after, err := os.Stat(forgePath)
	if err != nil {
		t.Fatalf("stat forge.md after init: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("byte-identical forge.md was rewritten under --force: mtime changed from %v to %v", before.ModTime(), after.ModTime())
	}

	orgData, err := os.ReadFile(orgPath)
	if err != nil {
		t.Fatalf("read org.md after init: %v", err)
	}
	if string(orgData) != string(embeddedOrg) {
		t.Errorf("drifted org.md was not updated under --force (behavior must equal default smart-sync): got %d bytes, want %d bytes", len(orgData), len(embeddedOrg))
	}
}

func TestInitCmd_ForceDeprecationNotice(t *testing.T) {
	dir := t.TempDir()

	var buf bytes.Buffer
	root := newInitTestRoot()
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--force", "--path", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute init --force: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Flag --force has been deprecated") {
		t.Errorf("expected --force deprecation notice in command output, got: %q", out)
	}
	if !strings.Contains(out, "smart-sync") {
		t.Errorf("deprecation notice should name smart-sync as the replacement default, got: %q", out)
	}
}

func TestRunInit_SkipExistingWinsOverForce(t *testing.T) {
	dir := t.TempDir()

	orgPath := filepath.Join(dir, ".opencode", "commands", "org.md")
	if err := os.MkdirAll(filepath.Dir(orgPath), 0o755); err != nil {
		t.Fatalf("setup MkdirAll: %v", err)
	}
	custom := []byte("# my custom org\n")
	if err := os.WriteFile(orgPath, custom, 0o600); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	var buf bytes.Buffer
	root := newInitTestRoot()
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--force", "--skip-existing", "--path", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute init --force --skip-existing: %v", err)
	}

	data, err := os.ReadFile(orgPath)
	if err != nil {
		t.Fatalf("read org.md after init: %v", err)
	}
	if string(data) != string(custom) {
		t.Errorf("--skip-existing must take precedence over --force: org.md was modified: got %q, want %q", string(data), string(custom))
	}

	if !strings.Contains(buf.String(), "Flag --force has been deprecated") {
		t.Errorf("expected --force deprecation notice even when combined with --skip-existing, got: %q", buf.String())
	}
}

func TestRunInit_AlreadyInitialized(t *testing.T) {
	dir := t.TempDir()

	// First init.
	if err := runInit(initOptions{targetDir: dir}); err != nil {
		t.Fatalf("first runInit: %v", err)
	}

	// Write something to cells.json to verify it's not overwritten.
	cellsPath := filepath.Join(dir, ".uf", "replicator", "cells.json")
	if err := os.WriteFile(cellsPath, []byte(`[{"id":"test"}]`), 0o600); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	// Second init — should be idempotent.
	if err := runInit(initOptions{targetDir: dir}); err != nil {
		t.Fatalf("second runInit: %v", err)
	}

	// Verify cells.json was NOT overwritten.
	data, err := os.ReadFile(cellsPath)
	if err != nil {
		t.Fatalf("read cells.json: %v", err)
	}
	if string(data) != `[{"id":"test"}]` {
		t.Errorf("cells.json was overwritten: got %q", string(data))
	}

	// Verify agent kit files still exist (scaffolded on first run, skipped on second).
	forgePath := filepath.Join(dir, ".opencode", "commands", "forge.md")
	if _, err := os.Stat(forgePath); err != nil {
		t.Errorf("agent kit files should exist after second init: %v", err)
	}

	// Verify DCP config still exists and content was preserved.
	dcpPath := filepath.Join(dir, ".opencode", "dcp.jsonc")
	dcpBefore, err := os.ReadFile(dcpPath)
	if err != nil {
		t.Fatalf("read dcp.jsonc before second init: %v", err)
	}
	// Re-run init and verify DCP content is unchanged.
	if err := runInit(initOptions{targetDir: dir}); err != nil {
		t.Fatalf("third runInit: %v", err)
	}
	dcpAfter, err := os.ReadFile(dcpPath)
	if err != nil {
		t.Fatalf("read dcp.jsonc after third init: %v", err)
	}
	if string(dcpAfter) != string(dcpBefore) {
		t.Error("dcp.jsonc was modified on re-init")
	}
}

func TestRunInit_CustomPath(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "myproject")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("setup MkdirAll: %v", err)
	}

	if err := runInit(initOptions{targetDir: target}); err != nil {
		t.Fatalf("runInit with custom path: %v", err)
	}

	cellsPath := filepath.Join(target, ".uf", "replicator", "cells.json")
	if _, err := os.Stat(cellsPath); err != nil {
		t.Fatalf("cells.json not created at custom path: %v", err)
	}

	forgePath := filepath.Join(target, ".opencode", "commands", "forge.md")
	if _, err := os.Stat(forgePath); err != nil {
		t.Fatalf("agent kit not created at custom path: %v", err)
	}
}

func TestRunInit_InvalidPath(t *testing.T) {
	err := runInit(initOptions{targetDir: "/nonexistent/path/that/cannot/exist"})
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}
