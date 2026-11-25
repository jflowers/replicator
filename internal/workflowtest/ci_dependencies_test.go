// Package workflowtest verifies repository workflows against their approved policies.
package workflowtest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

const (
	orgInfraSHA = "8f347ba1aa2c7e9eb3b70e69c1709ddb456cd64a"
	shipCaller  = "jflowers/ship/.github/workflows/deps-automerge.yml"
)

// ci_dependencies.yml only reviews: one read-only reusable reviewer for
// every pull request, nothing that writes, approves or merges.
func TestCIDependenciesWorkflow_ReviewsEveryPullRequestReadOnly(t *testing.T) {
	workflow := readDependencyWorkflow(t)

	assertMatches(t, workflow, `(?m)^# .+\n# --\n# .+`, "purpose header")
	assertMatches(t, workflow, `(?ms)^on:\s*\n\s+push:\s*\n\s+branches:\s*\[main\]`, "main push trigger")
	assertMatches(t, workflow, `(?ms)^on:.*?\n\s+pull_request:\s*\n\s+branches:\s*\[main\]`, "unprivileged main pull-request trigger")
	if strings.Contains(workflow, "pull_request_target:") {
		t.Fatal("workflow must not use the privileged pull_request_target trigger")
	}
	assertMatches(t, workflow, `(?ms)^permissions:\s*\n\s+contents:\s+read\s*\n\s+issues:\s+none\s*\n\s+pull-requests:\s+none`, "read-only workflow permissions")
	assertMatches(t, workflow, `(?ms)^concurrency:\s*\n\s+group:\s+.*github\.workflow.*github\.event_name.*github\.event\.pull_request\.number.*github\.ref.*\n\s+cancel-in-progress:\s+true`, "event- and PR- or ref-scoped concurrency")

	jobs := childKeys(t, yamlSection(t, workflow, "jobs", 0), 2)
	sort.Strings(jobs)
	if strings.Join(jobs, ",") != "call_deps_reviewer" {
		t.Fatalf("workflow jobs = %v, want exactly [call_deps_reviewer]", jobs)
	}
	depsJob := yamlSection(t, workflow, "call_deps_reviewer", 2)
	assertContains(t, depsJob, "reusable_deps_reviewer.yml@"+orgInfraSHA, "general reusable reviewer")
	assertPermissions(t, depsJob, map[string]string{"contents": "read", "issues": "none", "pull-requests": "none"})
	assertFullSHAPins(t, workflow, 1)
	if strings.Contains(workflow, "secrets:") {
		t.Error("workflow must not pass secrets to the reusable reviewer")
	}
	assertNoMutationCapabilities(t, workflow)
}

// dependabot-automerge.yml hands Dependabot's pull requests to ship's
// reusable workflow, pinned by full SHA, with only the two write
// permissions it needs and no secrets.
func TestDependabotAutomergeWorkflow_DelegatesToPinnedShip(t *testing.T) {
	workflow := readRepositoryFile(t, ".github", "workflows", "dependabot-automerge.yml")

	assertMatches(t, workflow, `(?m)^on:\s+pull_request\s*$`, "unprivileged pull-request trigger")
	if strings.Contains(workflow, "pull_request_target") {
		t.Fatal("workflow must not use the privileged pull_request_target trigger")
	}
	assertMatches(t, workflow, `(?m)^permissions:\s+\{\}\s*$`, "no default permissions")
	jobs := childKeys(t, yamlSection(t, workflow, "jobs", 0), 2)
	if strings.Join(jobs, ",") != "automerge" {
		t.Fatalf("workflow jobs = %v, want exactly [automerge]", jobs)
	}
	job := yamlSection(t, workflow, "automerge", 2)
	assertMatches(t, job, `(?m)^\s+uses:\s+`+regexp.QuoteMeta(shipCaller)+`@[0-9a-f]{40} # v\d+\.\d+\.\d+$`, "ship caller pinned to a full SHA with its version")
	assertPermissions(t, job, map[string]string{"contents": "write", "pull-requests": "write"})
	if strings.Contains(workflow, "secrets:") {
		t.Error("workflow must not pass secrets to ship")
	}
}

// Dependabot waits a day for third-party releases and groups ship's pins
// on their own, so a ship release isn't held up by another update.
func TestDependabotConfig_CooldownAndShipGroup(t *testing.T) {
	config := readRepositoryFile(t, ".github", "dependabot.yml")

	if got := len(regexp.MustCompile(`(?m)^\s+cooldown:\s*\n\s+default-days:\s+1\s*$`).FindAllString(config, -1)); got != 2 {
		t.Errorf("cooldown with default-days: 1 appears %d times, want once per ecosystem (2)", got)
	}
	assertMatches(t, config, `(?m)^\s+ship:\s*\n\s+patterns:\s+\["jflowers/ship\*"\]\s*$`, "ship group")
}

func TestFullsendZizmorException_IsNarrowAndDocumented(t *testing.T) {
	config := readRepositoryFile(t, ".github", "zizmor.yml")
	wantConfig := `# Fullsend's generated pull_request_target shim cannot be hardened locally without breaking the supported integration.
# Issue #127 explicitly authorizes this finding-specific exception pending an upstream-compatible hardening path.
rules:
  dangerous-triggers:
    ignore:
      - fullsend.yaml
`
	if config != wantConfig {
		t.Fatalf("zizmor configuration = %q, want only the documented Fullsend dangerous-triggers exception", config)
	}

	documentation := readRepositoryFile(t, "docs", "ci-workflow-security.md")
	for _, requirement := range []string{
		"dangerous-triggers",
		"Issue #127",
		".github/workflows/fullsend.yaml",
		"Fullsend-managed generated shim",
		"pull_request_target",
		"permissions: {}",
		"FULLSEND_GCP_WIF_PROVIDER",
		"FULLSEND_GCP_PROJECT_ID",
		"OTEL_EXPORTER_OTLP_TRACES_HEADERS",
		"OTEL_EXPORTER_OTLP_HEADERS",
		"d5f36921ac754705619f38c637ef692873809fbc",
		"no checkout action",
		"upstream-supported Fullsend configuration or generator update",
	} {
		assertContains(t, documentation, requirement, "Fullsend exception documentation")
	}
}

func readDependencyWorkflow(t *testing.T) string {
	return readRepositoryFile(t, ".github", "workflows", "ci_dependencies.yml")
}

func readRepositoryFile(t *testing.T, pathElements ...string) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller: could not determine test source path")
	}
	path := filepath.Join(append([]string{filepath.Dir(currentFile), "..", ".."}, pathElements...)...)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read repository file %q: %v", path, err)
	}
	return string(content)
}

func yamlSection(t *testing.T, content, key string, indentWidth int) string {
	t.Helper()
	lines := strings.Split(content, "\n")
	prefix := strings.Repeat(" ", indentWidth) + key + ":"
	start := -1
	for index, line := range lines {
		if line == prefix {
			start = index
			break
		}
	}
	if start < 0 {
		t.Fatalf("workflow does not contain YAML section %q at indentation %d", key, indentWidth)
	}
	end := len(lines)
	for index := start + 1; index < len(lines); index++ {
		line := lines[index]
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if leadingSpaces(line) <= indentWidth {
			end = index
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func childKeys(t *testing.T, section string, indentation int) []string {
	t.Helper()
	pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^ {%d}([A-Za-z0-9_-]+):(?:\s.*)?$`, indentation))
	matches := pattern.FindAllStringSubmatch(section, -1)
	keys := make([]string, 0, len(matches))
	for _, match := range matches {
		keys = append(keys, match[1])
	}
	if len(keys) == 0 {
		t.Fatalf("section has no child keys at indentation %d", indentation)
	}
	return keys
}

func yamlLiteralBlock(t *testing.T, section, key string) string {
	t.Helper()
	lines := strings.Split(section, "\n")
	pattern := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(key) + `:\s*[|>][-+]?\s*$`)
	start := -1
	keyIndent := 0
	for index, line := range lines {
		if pattern.MatchString(line) {
			start = index + 1
			keyIndent = leadingSpaces(line)
			break
		}
	}
	if start < 0 {
		t.Fatalf("YAML section does not contain literal block %q", key)
	}
	end := len(lines)
	blockIndent := -1
	for index := start; index < len(lines); index++ {
		line := lines[index]
		if strings.TrimSpace(line) == "" {
			continue
		}
		indentation := leadingSpaces(line)
		if indentation <= keyIndent {
			end = index
			break
		}
		if blockIndent < 0 || indentation < blockIndent {
			blockIndent = indentation
		}
	}
	if blockIndent < 0 {
		t.Fatalf("YAML literal block %q is empty", key)
	}
	block := make([]string, 0, end-start)
	for _, line := range lines[start:end] {
		if len(line) >= blockIndent {
			line = line[blockIndent:]
		}
		block = append(block, line)
	}
	return strings.Join(block, "\n")
}

func leadingSpaces(value string) int {
	return len(value) - len(strings.TrimLeft(value, " "))
}

func assertFullSHAPins(t *testing.T, workflow string, minimum int) {
	t.Helper()
	usesPattern := regexp.MustCompile(`(?m)^\s*(?:-\s+)?uses:\s+([^\s@]+)@([^\s#]+)(?:\s+#\s+(.+))?$`)
	matches := usesPattern.FindAllStringSubmatch(workflow, -1)
	if len(matches) < minimum {
		t.Fatalf("uses references = %d, want at least %d pinned references", len(matches), minimum)
	}
	shaPattern := regexp.MustCompile(`^[0-9a-f]{40}$`)
	versionPattern := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	for _, match := range matches {
		if !shaPattern.MatchString(match[2]) {
			t.Errorf("uses reference %s@%s is not pinned to a full commit SHA", match[1], match[2])
		}
		if !versionPattern.MatchString(strings.TrimSpace(match[3])) {
			t.Errorf("uses reference %s@%s has missing or inaccurate-looking version comment %q", match[1], match[2], match[3])
		}
	}
}

func assertPermissions(t *testing.T, job string, want map[string]string) {
	t.Helper()
	section := yamlSection(t, job, "permissions", 4)
	keys := childKeys(t, section, 6)
	if len(keys) != len(want) {
		t.Fatalf("permission keys = %v, want exactly %v", keys, want)
	}
	for _, key := range keys {
		value, ok := want[key]
		if !ok {
			t.Errorf("unexpected permission %q", key)
			continue
		}
		assertMatches(t, section, `(?m)^\s+`+regexp.QuoteMeta(key)+`:\s+`+regexp.QuoteMeta(value)+`\s*$`, key+" permission")
	}
}

func assertNoMutationCapabilities(t *testing.T, workflow string) {
	t.Helper()
	prohibited := []string{
		`(?i)uses:[^\n]*actions/checkout`,
		`(?i)pulls\.merge`,
		`(?i)enablePullRequestAutoMerge`,
		`(?i)disablePullRequestAutoMerge`,
		`(?i)gh\s+pr\s+merge`,
		`(?i)merge-method\s*:`,
		`(?i)uses:[^\n]*(auto-merge|automerge)`,
		`(?i)dismissReview|dismiss_review`,
		`(?i)(repos|branches)\.(update|delete).*(protection|required|statusCheck)`,
	}
	for _, expression := range prohibited {
		if regexp.MustCompile(expression).MatchString(workflow) {
			t.Errorf("workflow contains prohibited merge or repository-governance mutation matching %q", expression)
		}
	}
}

func assertContains(t *testing.T, content, expected, description string) {
	t.Helper()
	if !strings.Contains(content, expected) {
		t.Errorf("workflow is missing %s %q", description, expected)
	}
}

func assertMatches(t *testing.T, content, expression, description string) {
	t.Helper()
	if !regexp.MustCompile(expression).MatchString(content) {
		t.Errorf("workflow is missing %s matching %q", description, expression)
	}
}
