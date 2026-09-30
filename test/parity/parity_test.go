//go:build parity

// Parity tests verify that the Go rewrite produces MCP tool responses with
// the same JSON shape as the TypeScript cyborg-swarm original.
//
// Run with: go test -tags parity ./test/parity/ -count=1 -v
//
// These tests are excluded from normal `go test ./...` via the build tag.
// They use an in-memory SQLite database and call tool Execute functions
// directly (no subprocess needed).
package parity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/memory"
	commstools "github.com/unbound-force/replicator/internal/tools/comms"
	forgetools "github.com/unbound-force/replicator/internal/tools/forge"
	memtools "github.com/unbound-force/replicator/internal/tools/memory"
	orgtools "github.com/unbound-force/replicator/internal/tools/org"
	"github.com/unbound-force/replicator/internal/tools/registry"
)

// fixtureEntry represents a single tool's fixture data.
type fixtureEntry struct {
	Request            json.RawMessage `json:"request"`
	TypeScriptResponse json.RawMessage `json:"typescript_response"`
}

// fixtureFile maps tool names to their fixture entries.
type fixtureFile map[string]fixtureEntry

// loadFixtures reads all JSON fixture files from the fixtures directory.
func loadFixtures(t *testing.T) map[string]fixtureEntry {
	t.Helper()

	fixturesDir := filepath.Join("fixtures")
	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("read fixtures dir: %v", err)
	}

	all := make(map[string]fixtureEntry)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(fixturesDir, entry.Name()))
		if err != nil {
			t.Fatalf("read fixture %s: %v", entry.Name(), err)
		}

		var ff fixtureFile
		if err := json.Unmarshal(data, &ff); err != nil {
			t.Fatalf("parse fixture %s: %v", entry.Name(), err)
		}

		for name, entry := range ff {
			all[name] = entry
		}
	}

	return all
}

// setupRegistry creates an in-memory store and registers all tools.
func setupRegistry(t *testing.T) (*registry.Registry, *db.Store) {
	t.Helper()

	store, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("open memory db: %v", err)
	}

	reg := registry.New()

	// Register all tool families.
	orgtools.Register(reg, store)
	commstools.Register(reg, store)
	forgetools.Register(reg, store)

	// Memory tools use a Dewey proxy client. For parity tests, we create
	// a client pointing to a non-existent server -- the deprecated tools
	// don't need a real connection, and the proxy tools will return
	// DEWEY_UNAVAILABLE which is still a valid response shape.
	memClient := memory.NewClient("http://localhost:0")
	memtools.Register(reg, memClient)

	return reg, store
}

// createTempGitRepo creates a temporary directory initialized as a git repo.
func createTempGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git setup %v: %v\n%s", args, err, out)
		}
	}

	return dir
}

// wrapMCPResponse wraps a tool's string output in the MCP content envelope
// that the TypeScript version uses: {"content": [{"type": "text", "text": "..."}]}
func wrapMCPResponse(text string) json.RawMessage {
	envelope := map[string]any{
		"content": []map[string]string{
			{"type": "text", "text": text},
		},
	}
	data, _ := json.Marshal(envelope)
	return data
}

// extractTextFromMCP extracts the text field from an MCP content envelope.
func extractTextFromMCP(raw json.RawMessage) (json.RawMessage, error) {
	var envelope struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshal MCP envelope: %w", err)
	}
	if len(envelope.Content) == 0 {
		return nil, fmt.Errorf("empty content array")
	}
	return json.RawMessage(envelope.Content[0].Text), nil
}

func TestParity(t *testing.T) {
	reg, store := setupRegistry(t)
	defer store.Close()

	fixtures := loadFixtures(t)
	if len(fixtures) == 0 {
		t.Fatal("no fixtures loaded")
	}

	// Create a temp git repo for worktree tools.
	gitDir := createTempGitRepo(t)

	// Tools that need pre-existing data or special handling.
	// These are tested in dedicated subtests below.
	skipTools := map[string]string{
		"org_cells_with_data": "requires pre-existing data, tested separately",
		"org_close":           "requires pre-existing cell",
		"org_update":          "requires pre-existing cell",
		"org_start":           "requires pre-existing cell",
	}

	// Tools that need argument rewriting (e.g., project_path).
	pathRewriteTools := map[string]bool{
		"forge_init":          true,
		"forge_worktree_list": true,
	}

	var results []ToolResult

	// Sort fixture names for deterministic execution order.
	// This ensures org_ready runs before org_create (which would
	// create cells and change the org_ready response shape).
	names := make([]string, 0, len(fixtures))
	for name := range fixtures {
		names = append(names, name)
	}
	sort.Strings(names)

	// Run empty-state tools first: org_cells, org_query, org_ready
	// must run before any org_create calls populate the database.
	emptyStateTools := []string{"org_cells", "org_query", "org_ready"}
	for _, name := range emptyStateTools {
		fixture, ok := fixtures[name]
		if !ok {
			continue
		}
		t.Run(name+"_empty", func(t *testing.T) {
			result := runFixtureTool(t, reg, name, fixture, gitDir, pathRewriteTools)
			results = append(results, result)
		})
	}

	// Run remaining tools (excluding skip and already-run empty-state tools).
	alreadyRun := map[string]bool{
		"org_cells": true,
		"org_query": true,
		"org_ready": true,
	}
	for _, name := range names {
		if _, skip := skipTools[name]; skip {
			continue
		}
		if alreadyRun[name] {
			continue
		}

		fixture := fixtures[name]
		t.Run(name, func(t *testing.T) {
			result := runFixtureTool(t, reg, name, fixture, gitDir, pathRewriteTools)
			results = append(results, result)
		})
	}

	// Test tools that need pre-existing data.
	t.Run("org_close_with_data", func(t *testing.T) {
		result := testToolWithSetup(t, reg, store, "org_close", fixtures)
		results = append(results, result)
	})

	t.Run("org_update_with_data", func(t *testing.T) {
		result := testToolWithSetup(t, reg, store, "org_update", fixtures)
		results = append(results, result)
	})

	t.Run("org_start_with_data", func(t *testing.T) {
		result := testToolWithSetup(t, reg, store, "org_start", fixtures)
		results = append(results, result)
	})

	t.Run("org_cells_with_data", func(t *testing.T) {
		result := testCellsWithData(t, reg, store, fixtures)
		results = append(results, result)
	})

	// Generate the parity report.
	var buf bytes.Buffer
	if err := GenerateReport(results, &buf); err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	t.Logf("\n%s", buf.String())
}

// runFixtureTool executes a single fixture tool test and returns the result.
func runFixtureTool(t *testing.T, reg *registry.Registry, name string, fixture fixtureEntry, gitDir string, pathRewriteTools map[string]bool) ToolResult {
	t.Helper()

	// Extract tool name and arguments from the fixture request.
	var req struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(fixture.Request, &req); err != nil {
		t.Fatalf("parse request: %v", err)
	}

	// Rewrite project_path for tools that need a real git repo.
	args := req.Arguments
	if pathRewriteTools[name] {
		var argMap map[string]any
		json.Unmarshal(args, &argMap)
		if _, ok := argMap["project_path"]; ok {
			argMap["project_path"] = gitDir
		}
		args, _ = json.Marshal(argMap)
	}

	tool := reg.Get(req.Name)
	if tool == nil {
		t.Fatalf("tool %q not registered", req.Name)
	}

	// Execute the Go tool.
	goResult, err := tool.Execute(args)
	if err != nil {
		t.Logf("tool %s returned error (may be expected): %v", req.Name, err)
		return ToolResult{
			Name:  name,
			Match: false,
			Differences: []Difference{{
				Path:         "$",
				ExpectedType: "object",
				ActualType:   fmt.Sprintf("error: %v", err),
			}},
		}
	}

	// Wrap Go result in MCP envelope for comparison.
	goMCP := wrapMCPResponse(goResult)

	// Compare MCP envelope shapes first.
	match, diffs := ShapeMatch(fixture.TypeScriptResponse, goMCP)

	if match {
		// Envelope matches. Now compare the inner text content shapes.
		expectedText, err := extractTextFromMCP(fixture.TypeScriptResponse)
		if err != nil {
			t.Fatalf("extract expected text: %v", err)
		}
		actualText := json.RawMessage(goResult)

		// Only compare inner shapes if both are valid JSON.
		// Some tools return plain strings (prompt generators).
		if isJSON(expectedText) && isJSON(actualText) {
			match, diffs = ShapeMatch(expectedText, actualText)
		}
	}

	if !match {
		for _, d := range diffs {
			t.Errorf("shape mismatch at %s: expected %s, got %s",
				d.Path, d.ExpectedType, d.ActualType)
		}
	}

	return ToolResult{
		Name:        name,
		Match:       match,
		Differences: diffs,
	}
}

// testToolWithSetup creates a cell, then tests a tool that requires one.
func testToolWithSetup(t *testing.T, reg *registry.Registry, store *db.Store, toolName string, fixtures map[string]fixtureEntry) ToolResult {
	t.Helper()

	fixture, ok := fixtures[toolName]
	if !ok {
		return ToolResult{
			Name:  toolName,
			Match: false,
			Differences: []Difference{{
				Path:         "$",
				ExpectedType: "fixture",
				ActualType:   "missing",
			}},
		}
	}

	// Create a cell to operate on.
	createTool := reg.Get("org_create")
	createResult, err := createTool.Execute(json.RawMessage(`{"title": "test cell for ` + toolName + `"}`))
	if err != nil {
		t.Fatalf("create cell for %s: %v", toolName, err)
	}

	// Extract the cell ID.
	var cell struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(createResult), &cell); err != nil {
		t.Fatalf("parse created cell: %v", err)
	}

	// Build arguments with the real cell ID.
	var origArgs map[string]any
	var req struct {
		Arguments json.RawMessage `json:"arguments"`
	}
	json.Unmarshal(fixture.Request, &req)
	json.Unmarshal(req.Arguments, &origArgs)
	origArgs["id"] = cell.ID
	argsJSON, _ := json.Marshal(origArgs)

	// Execute the tool.
	tool := reg.Get(toolName)
	goResult, err := tool.Execute(argsJSON)
	if err != nil {
		return ToolResult{
			Name:  toolName,
			Match: false,
			Differences: []Difference{{
				Path:         "$",
				ExpectedType: "object",
				ActualType:   fmt.Sprintf("error: %v", err),
			}},
		}
	}

	// Compare inner response shapes.
	expectedText, _ := extractTextFromMCP(fixture.TypeScriptResponse)
	actualText := json.RawMessage(goResult)

	match, diffs := ShapeMatch(expectedText, actualText)
	if !match {
		for _, d := range diffs {
			t.Errorf("%s shape mismatch at %s: expected %s, got %s",
				toolName, d.Path, d.ExpectedType, d.ActualType)
		}
	}

	return ToolResult{
		Name:        toolName,
		Match:       match,
		Differences: diffs,
	}
}

// testCellsWithData creates a cell and then queries for it.
func testCellsWithData(t *testing.T, reg *registry.Registry, store *db.Store, fixtures map[string]fixtureEntry) ToolResult {
	t.Helper()

	fixture, ok := fixtures["org_cells_with_data"]
	if !ok {
		return ToolResult{
			Name:  "org_cells_with_data",
			Match: false,
			Differences: []Difference{{
				Path:         "$",
				ExpectedType: "fixture",
				ActualType:   "missing",
			}},
		}
	}

	// Create a cell so the query returns data.
	createTool := reg.Get("org_create")
	_, err := createTool.Execute(json.RawMessage(`{"title": "test cell for query"}`))
	if err != nil {
		t.Fatalf("create cell for org_cells_with_data: %v", err)
	}

	// Query cells.
	tool := reg.Get("org_cells")
	goResult, err := tool.Execute(json.RawMessage(`{"status": "open"}`))
	if err != nil {
		return ToolResult{
			Name:  "org_cells_with_data",
			Match: false,
			Differences: []Difference{{
				Path:         "$",
				ExpectedType: "array",
				ActualType:   fmt.Sprintf("error: %v", err),
			}},
		}
	}

	// Compare array element shapes.
	expectedText, _ := extractTextFromMCP(fixture.TypeScriptResponse)
	actualText := json.RawMessage(goResult)

	match, diffs := ShapeMatch(expectedText, actualText)
	if !match {
		for _, d := range diffs {
			t.Errorf("org_cells_with_data shape mismatch at %s: expected %s, got %s",
				d.Path, d.ExpectedType, d.ActualType)
		}
	}

	return ToolResult{
		Name:        "org_cells_with_data",
		Match:       match,
		Differences: diffs,
	}
}

// isJSON returns true if the raw message is valid JSON (not a plain string
// that isn't JSON-encoded).
func isJSON(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var v any
	return json.Unmarshal(raw, &v) == nil
}
