// Package agentkit provides embedded agent kit content for scaffolding
// new project directories. The kit includes command definitions, skill
// files, and agent role descriptions that are written to .opencode/
// during `replicator init`.
package agentkit

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed content/*
var content embed.FS

// ScaffoldResult describes the outcome of writing a single agent kit file.
type ScaffoldResult struct {
	Path   string `json:"path"`
	Action string `json:"action"` // "created", "skipped", "overwritten"
}

// dcpConfigContent is the canonical DCP config that enables protectTags.
// This matches the replicator repo's own .opencode/dcp.jsonc.
const dcpConfigContent = `{
  "$schema": "https://raw.githubusercontent.com/Opencode-DCP/opencode-dynamic-context-pruning/master/dcp.schema.json",
  // Enable <protect> tag preservation during DCP compression.
  // Slash command files in .opencode/commands/ use <protect> tags
  // to mark execution-critical sections (guardrails, checklists,
  // mandatory gates) that must survive context pruning.
  "compress": {
    "protectTags": true
  }
}
`

// ScaffoldDCP creates or updates the DCP configuration file in
// targetDir/.opencode/. It checks for dcp.jsonc first, then dcp.json.
// If neither exists, it creates dcp.jsonc. If one exists with protectTags
// already configured, it is skipped. If one exists without protectTags,
// the file is replaced with the canonical DCP config content.
func ScaffoldDCP(targetDir string) (ScaffoldResult, error) {
	openCodeDir := filepath.Join(targetDir, ".opencode")
	jsoncPath := filepath.Join(openCodeDir, "dcp.jsonc")
	jsonPath := filepath.Join(openCodeDir, "dcp.json")

	// Check .jsonc first, then .json (D3: prefer .jsonc).
	var existingPath, fileName string
	if _, err := os.Stat(jsoncPath); err == nil {
		existingPath = jsoncPath
		fileName = "dcp.jsonc"
	} else if _, err := os.Stat(jsonPath); err == nil {
		existingPath = jsonPath
		fileName = "dcp.json"
	}

	if existingPath != "" {
		// File exists — check for protectTags (D2: string scan).
		data, err := os.ReadFile(existingPath)
		if err != nil {
			return ScaffoldResult{}, fmt.Errorf("read %s: %w", fileName, err)
		}

		if strings.Contains(string(data), "protectTags") {
			return ScaffoldResult{Path: fileName, Action: "skipped"}, nil
		}

		// File exists but lacks protectTags — replace with canonical content (D10).
		if err := os.WriteFile(existingPath, []byte(dcpConfigContent), 0o644); err != nil {
			return ScaffoldResult{}, fmt.Errorf("write %s: %w", fileName, err)
		}
		return ScaffoldResult{Path: fileName, Action: "updated"}, nil
	}

	// Neither file exists — create .opencode/dcp.jsonc.
	if err := os.MkdirAll(openCodeDir, 0o755); err != nil {
		return ScaffoldResult{}, fmt.Errorf("create .opencode directory: %w", err)
	}
	if err := os.WriteFile(jsoncPath, []byte(dcpConfigContent), 0o644); err != nil {
		return ScaffoldResult{}, fmt.Errorf("write dcp.jsonc: %w", err)
	}
	return ScaffoldResult{Path: "dcp.jsonc", Action: "created"}, nil
}

// Scaffold writes the embedded agent kit files to targetDir/.opencode/.
// If force is false, existing files are skipped. If force is true,
// existing files are overwritten.
func Scaffold(targetDir string, force bool) ([]ScaffoldResult, error) {
	var results []ScaffoldResult

	err := fs.WalkDir(content, "content", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Strip "content/" prefix to get relative path under .opencode/.
		relPath, _ := filepath.Rel("content", path)
		destPath := filepath.Join(targetDir, ".opencode", relPath)

		// Check if file exists.
		if _, statErr := os.Stat(destPath); statErr == nil {
			if !force {
				results = append(results, ScaffoldResult{Path: relPath, Action: "skipped"})
				return nil
			}
			results = append(results, ScaffoldResult{Path: relPath, Action: "overwritten"})
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("stat %s: %w", relPath, statErr)
		} else {
			results = append(results, ScaffoldResult{Path: relPath, Action: "created"})
		}

		// Create parent directories.
		if mkErr := os.MkdirAll(filepath.Dir(destPath), 0o755); mkErr != nil {
			return fmt.Errorf("create directory for %s: %w", relPath, mkErr)
		}

		// Read from embedded FS and write to disk.
		data, readErr := content.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read embedded %s: %w", path, readErr)
		}

		if writeErr := os.WriteFile(destPath, data, 0o644); writeErr != nil {
			return fmt.Errorf("write %s: %w", relPath, writeErr)
		}
		return nil
	})

	return results, err
}
