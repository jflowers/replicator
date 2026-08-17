// Package doctor runs health checks for the replicator environment.
//
// Checks verify that required dependencies (git, database, Dewey, config dir)
// are available and functional. Results include pass/fail/warn status and
// timing for each check.
package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/unbound-force/replicator/internal/config"
	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/mcpclient"
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
// The projectDir parameter is used for per-project checks (e.g., DCP config).
func Run(store *db.Store, cfg *config.Config, projectDir string) ([]CheckResult, error) {
	var results []CheckResult

	results = append(results, checkGit())
	results = append(results, checkDatabase(store))
	results = append(results, checkDewey(cfg.DeweyURL))
	results = append(results, checkConfigDir())
	results = append(results, checkDCPConfig(projectDir))

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

// deweyHealthProbe uses the shared MCP client to verify Dewey is alive.
// It sends an initialize handshake followed by a dewey_health tools/call.
func deweyHealthProbe(deweyURL string) error {
	client := mcpclient.New(deweyURL, mcpclient.Config{
		Name:    "replicator-doctor",
		Version: "1.0.0",
		Timeout: 5 * time.Second,
	})
	_, err := client.Call("dewey_health", map[string]any{})
	return err
}

// checkDCPConfig verifies DCP configuration when protect-tagged commands exist.
// It scans .opencode/commands/*.md for <protect> tags. If none are found,
// the check passes (DCP is not needed). If protect tags are found, it checks
// for .opencode/dcp.jsonc (or dcp.json) with protectTags enabled.
func checkDCPConfig(projectDir string) CheckResult {
	start := time.Now()

	cmdDir := filepath.Join(projectDir, ".opencode", "commands")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		// No commands directory — no protect tags to worry about.
		elapsed := time.Since(start)
		return CheckResult{
			Name:     "dcp_config",
			Status:   "pass",
			Message:  "no protect-tagged commands found",
			Duration: elapsed,
		}
	}

	// Scan command files for <protect> tags.
	hasProtectTags := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(cmdDir, e.Name()))
		if readErr != nil {
			elapsed := time.Since(start)
			return CheckResult{
				Name:     "dcp_config",
				Status:   "warn",
				Message:  fmt.Sprintf("cannot read command file %s: %v", e.Name(), readErr),
				Duration: elapsed,
			}
		}
		if strings.Contains(string(data), "<protect>") {
			hasProtectTags = true
			break
		}
	}

	if !hasProtectTags {
		elapsed := time.Since(start)
		return CheckResult{
			Name:     "dcp_config",
			Status:   "pass",
			Message:  "no protect-tagged commands found",
			Duration: elapsed,
		}
	}

	// Protect tags found — check for DCP config file.
	jsoncPath := filepath.Join(projectDir, ".opencode", "dcp.jsonc")
	jsonPath := filepath.Join(projectDir, ".opencode", "dcp.json")

	var configPath string
	if _, statErr := os.Stat(jsoncPath); statErr == nil {
		configPath = jsoncPath
	} else if _, statErr := os.Stat(jsonPath); statErr == nil {
		configPath = jsonPath
	}

	if configPath == "" {
		elapsed := time.Since(start)
		return CheckResult{
			Name:     "dcp_config",
			Status:   "warn",
			Message:  "protect-tagged commands found but no DCP config; run replicator init",
			Duration: elapsed,
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		elapsed := time.Since(start)
		return CheckResult{
			Name:     "dcp_config",
			Status:   "warn",
			Message:  fmt.Sprintf("cannot read DCP config: %v", err),
			Duration: elapsed,
		}
	}

	if !strings.Contains(string(data), "protectTags") {
		elapsed := time.Since(start)
		return CheckResult{
			Name:     "dcp_config",
			Status:   "warn",
			Message:  "DCP config exists but protectTags is not enabled",
			Duration: elapsed,
		}
	}

	elapsed := time.Since(start)
	return CheckResult{
		Name:     "dcp_config",
		Status:   "pass",
		Message:  "protectTags enabled in DCP config",
		Duration: elapsed,
	}
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
