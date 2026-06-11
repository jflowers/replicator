// Package doctor runs health checks for the replicator environment.
//
// Checks verify that required dependencies (git, database, Dewey, config dir)
// are available and functional. Results include pass/fail/warn status and
// timing for each check.
package doctor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/unbound-force/replicator/internal/config"
	"github.com/unbound-force/replicator/internal/db"
)

// CheckResult holds the outcome of a single health check.
type CheckResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"` // "pass", "fail", "warn"
	Message  string        `json:"message"`
	Duration time.Duration `json:"duration"`
}

// Run executes all health checks and returns the results.
// Individual check failures do not stop subsequent checks.
func Run(store *db.Store, cfg *config.Config) ([]CheckResult, error) {
	var results []CheckResult

	results = append(results, checkGit())
	results = append(results, checkDatabase(store))
	results = append(results, checkDewey(cfg.DeweyURL))
	results = append(results, checkConfigDir())

	return results, nil
}

// checkGit verifies that git is installed and returns its version.
func checkGit() CheckResult {
	start := time.Now()

	cmd := exec.Command("git", "--version")
	out, err := cmd.Output()
	elapsed := time.Since(start)

	if err != nil {
		return CheckResult{
			Name:     "git",
			Status:   "fail",
			Message:  fmt.Sprintf("git not found: %v", err),
			Duration: elapsed,
		}
	}

	version := strings.TrimSpace(string(out))
	return CheckResult{
		Name:     "git",
		Status:   "pass",
		Message:  version,
		Duration: elapsed,
	}
}

// checkDatabase verifies the SQLite database is accessible.
func checkDatabase(store *db.Store) CheckResult {
	start := time.Now()

	err := store.DB.Ping()
	elapsed := time.Since(start)

	if err != nil {
		return CheckResult{
			Name:     "database",
			Status:   "fail",
			Message:  fmt.Sprintf("database ping failed: %v", err),
			Duration: elapsed,
		}
	}

	return CheckResult{
		Name:     "database",
		Status:   "pass",
		Message:  "SQLite database is accessible",
		Duration: elapsed,
	}
}

// checkDewey verifies the Dewey semantic search endpoint is reachable.
// Sends an MCP initialize request (JSON-RPC 2.0 POST) to verify connectivity.
// The MCP Streamable HTTP transport requires POST with Accept header including
// both application/json and text/event-stream. A failure is a warning, not an
// error, because Dewey is optional for core operations.
func checkDewey(deweyURL string) CheckResult {
	start := time.Now()

	err := deweyHealthProbe(deweyURL)
	elapsed := time.Since(start)

	if err != nil {
		return CheckResult{
			Name:     "dewey",
			Status:   "warn",
			Message:  fmt.Sprintf("Dewey not reachable at %s: %v", deweyURL, err),
			Duration: elapsed,
		}
	}

	return CheckResult{
		Name:     "dewey",
		Status:   "pass",
		Message:  fmt.Sprintf("Dewey is reachable at %s", deweyURL),
		Duration: elapsed,
	}
}

// deweyHealthProbe sends an MCP initialize request to verify Dewey is alive.
// This is a lightweight probe that does not establish a full session.
func deweyHealthProbe(deweyURL string) error {
	reqBody := map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialize",
		"id":      1,
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":   map[string]any{},
			"clientInfo": map[string]any{
				"name":    "replicator-doctor",
				"version": "1.0.0",
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, deweyURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	// Read the SSE response — look for a JSON-RPC result in the event stream.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// The response is SSE format: "event: message\ndata: {json}\n\n"
	// Extract the JSON data line.
	for _, line := range strings.Split(string(respBody), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var rpcResp struct {
			Result any `json:"result"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &rpcResp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		if rpcResp.Error != nil {
			return fmt.Errorf("dewey error: %s", rpcResp.Error.Message)
		}
		// Got a successful initialize response — Dewey is alive.
		return nil
	}

	return fmt.Errorf("no valid response from Dewey")
}

// checkConfigDir verifies the config directory exists.
func checkConfigDir() CheckResult {
	start := time.Now()

	home, err := os.UserHomeDir()
	if err != nil {
		elapsed := time.Since(start)
		return CheckResult{
			Name:     "config_dir",
			Status:   "fail",
			Message:  fmt.Sprintf("cannot determine home directory: %v", err),
			Duration: elapsed,
		}
	}

	configDir := home + "/.config/uf/replicator"
	info, err := os.Stat(configDir)
	elapsed := time.Since(start)

	if os.IsNotExist(err) {
		return CheckResult{
			Name:     "config_dir",
			Status:   "fail",
			Message:  fmt.Sprintf("config directory does not exist: %s", configDir),
			Duration: elapsed,
		}
	}
	if err != nil {
		return CheckResult{
			Name:     "config_dir",
			Status:   "fail",
			Message:  fmt.Sprintf("cannot access config directory: %v", err),
			Duration: elapsed,
		}
	}
	if !info.IsDir() {
		return CheckResult{
			Name:     "config_dir",
			Status:   "fail",
			Message:  fmt.Sprintf("%s exists but is not a directory", configDir),
			Duration: elapsed,
		}
	}

	return CheckResult{
		Name:     "config_dir",
		Status:   "pass",
		Message:  fmt.Sprintf("config directory exists: %s", configDir),
		Duration: elapsed,
	}
}
