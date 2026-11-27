package workflowtest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// ci.yml's scope job calls jflowers/ship's pr-scope.yml, pinned by full
// SHA; ship owns and tests the classifier. Its code output is 'false'
// only for Dewey learnings PRs and gh-hud toolkit rollouts.
var shipScopeRE = regexp.MustCompile(`(?m)^\s+uses:\s*jflowers/ship/\.github/workflows/pr-scope\.yml@[0-9a-f]{40}(\s+#.*)?$`)

// TestCIGatingWorkflows_StructureEnforcesChoreSkip pins the job graph:
// the expensive jobs run only on code == 'true', gitleaks only on
// chores, and CI OK needs all of them, so it alone can be the required
// check.
func TestCIGatingWorkflows_StructureEnforcesChoreSkip(t *testing.T) {
	workflow := readRepositoryFile(t, ".github", "workflows", "ci.yml")

	jobs := childKeys(t, yamlSection(t, workflow, "jobs", 0), 2)
	sort.Strings(jobs)
	if want := "build-and-test,call-reusable-ci,ci-ok,gitleaks,scope"; strings.Join(jobs, ",") != want {
		t.Errorf("ci.yml jobs = %v, want exactly [%s]: a new job needs a gating decision and a place in CI OK", jobs, want)
	}

	scopeJob := yamlSection(t, workflow, "scope", 2)
	if !shipScopeRE.MatchString(scopeJob) {
		t.Error("scope job must call jflowers/ship's pr-scope.yml pinned to a full commit SHA")
	}
	assertPermissions(t, scopeJob, map[string]string{"contents": "read"})
	for _, key := range []string{"if", "needs"} {
		if regexp.MustCompile(`(?m)^ {4}` + key + `:`).MatchString(scopeJob) {
			t.Errorf("scope job must not declare %q: it decides for every run", key)
		}
	}

	for _, gated := range []struct{ id, name, code string }{
		{"build-and-test", "Build and Test", "true"},
		{"call-reusable-ci", "Standardized CI", "true"},
		{"gitleaks", "Gitleaks", "false"},
	} {
		job := yamlSection(t, workflow, gated.id, 2)
		// Job names are check names that release preflight and branch
		// protection read; renaming one is a separate decision.
		assertMatches(t, job, `(?m)^ {4}name:\s*`+regexp.QuoteMeta(gated.name)+`\s*$`, gated.id+" name")
		assertNeeds(t, job, gated.id, []string{"scope"})
		assertMatches(t, job, `(?m)^ {4}if:\s*needs\.scope\.outputs\.code\s*==\s*'`+gated.code+`'\s*$`, gated.id+" gating condition")
	}
	assertMatches(t, yamlSection(t, workflow, "call-reusable-ci", 2), `(?m)^\s+uses:\s*complytime/org-infra/\.github/workflows/reusable_ci\.yml@`, "org-infra reusable CI call")
	gitleaks := yamlSection(t, workflow, "gitleaks", 2)
	assertContains(t, gitleaks, `sha256sum -c -`, "gitleaks download checksum")
	assertContains(t, gitleaks, `--log-opts='HEAD^1..HEAD^2'`, "gitleaks scans the PR's commits")

	aggregateJob := yamlSection(t, workflow, "ci-ok", 2)
	assertMatches(t, aggregateJob, `(?m)^\s+name:\s*CI OK\s*$`, "aggregate job name")
	// The only exemption from always() is a cancelled push run; a
	// skipped aggregate satisfies a required check, so a cancelled
	// pull_request run must never skip it.
	assertMatches(t, aggregateJob, `(?m)^\s+if:\s*always\(\) && !\(github\.event_name == 'push' && cancelled\(\)\)\s*$`, "aggregate always-run condition, exempting only cancelled push runs")
	assertNeeds(t, aggregateJob, "ci-ok", []string{"scope", "build-and-test", "call-reusable-ci", "gitleaks"})
	// The verdict itself is executed by
	// TestCIGatingWorkflows_AggregateVerdictBehavior; here, pin what it
	// reads and that it is the job's only step, so no step-level if: can
	// bypass it.
	for env, expr := range ciOKEnv {
		assertMatches(t, aggregateJob, `(?m)^\s+`+env+`:\s*`+regexp.QuoteMeta(expr)+`\s*$`, "aggregate reads "+expr)
	}
	if steps := regexp.MustCompile(`(?m)^\s+- name:`).FindAllString(aggregateJob, -1); len(steps) != 1 {
		t.Errorf("ci-ok has %d steps, want exactly 1 (the verdict)", len(steps))
	}
	if regexp.MustCompile(`(?m)^\s+- name:.*\n(\s+#.*\n)*\s+if:`).MatchString(aggregateJob) {
		t.Error("ci-ok verdict step must not have a step-level if:, which could skip it")
	}
	if strings.Contains(yamlLiteralBlock(t, aggregateJob, "run"), "${{") {
		t.Error("ci-ok run script must not interpolate ${{ }} expressions; pass them through env")
	}

	onBlock := yamlSection(t, workflow, "on", 0)
	for _, filter := range []string{"paths", "paths-ignore"} {
		if regexp.MustCompile(`(?m)^\s+` + filter + `:\s`).MatchString(onBlock) {
			t.Errorf("on: block must not filter triggers with %q", filter)
		}
	}
}

// ciOKEnv maps the verdict's env vars to the needs expressions they read.
var ciOKEnv = map[string]string{
	"SCOPE":    "${{ needs.scope.result }}",
	"CODE":     "${{ needs.scope.outputs.code }}",
	"BUILD":    "${{ needs.build-and-test.result }}",
	"LINT":     "${{ needs.call-reusable-ci.result }}",
	"GITLEAKS": "${{ needs.gitleaks.result }}",
}

// main's required checks live in unbound-force/.github's safe-settings
// (repos/replicator.yml), not in .github/settings.yml (CLS-FR-003). The
// Settings app must remove the classic branch protection, not set it
// again: its strict rule left every open PR behind after each auto-merge.
func TestCIGatingSettings_ClassicProtectionRetired(t *testing.T) {
	settings := readRepositoryFile(t, ".github", "settings.yml")
	if err := classicProtectionRetired(settings); err != nil {
		t.Error(err)
	}
}

// The check must follow the YAML, not its formatting: comments, spacing
// and quoting don't matter, and restoring any protection must fail.
func TestCIGatingSettings_ClassicProtectionCheck(t *testing.T) {
	cases := []struct {
		name     string
		settings string
		wantErr  string
	}{
		{"null", "branches:\n  - name: main\n    protection: null\n", ""},
		{"tilde, quoted name, comments, wider indent", "# required_status_checks: lives in safe-settings\nbranches:\n    -   name: \"main\"  # the default branch\n\n        protection:   ~   # retired\n", ""},
		{"other branches keep theirs", "branches:\n  - name: release\n    protection:\n      enforce_admins: true\n  - name: main\n    protection: null\n", ""},
		{"protection restored", "branches:\n  - name: main\n    protection:\n      enforce_admins: false\n", "main's protection"},
		{"required checks restored", "branches:\n  - name: main\n    protection:\n      required_status_checks:\n        strict: true\n", "required_status_checks"},
		{"main missing", "branches:\n  - name: release\n    protection: null\n", "no main entry"},
		{"no branches", "repository:\n  has_wiki: true\n", "no main entry"},
		{"protection key missing", "branches:\n  - name: main\n", "main's protection"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := classicProtectionRetired(c.settings)
			switch {
			case c.wantErr == "" && err != nil:
				t.Errorf("unexpected error: %v", err)
			case c.wantErr != "" && err == nil:
				t.Errorf("want an error containing %q, got none", c.wantErr)
			case c.wantErr != "" && !strings.Contains(err.Error(), c.wantErr):
				t.Errorf("error %q does not contain %q", err, c.wantErr)
			}
		})
	}
}

// classicProtectionRetired reports whether a Settings app file leaves main
// without classic branch protection: no required_status_checks key
// anywhere, and main's entry under branches has protection null (or ~).
func classicProtectionRetired(settings string) error {
	type line struct {
		indent int
		text   string
	}
	var lines []line
	for _, raw := range strings.Split(settings, "\n") {
		text := raw
		if i := strings.Index(text, " #"); i >= 0 {
			text = text[:i]
		}
		if strings.HasPrefix(strings.TrimSpace(text), "#") || strings.TrimSpace(text) == "" {
			continue
		}
		lines = append(lines, line{leadingSpaces(text), strings.TrimRight(text, " ")})
	}
	keyRE := regexp.MustCompile(`^\s*(?:-\s+)?required_status_checks\s*:`)
	for _, l := range lines {
		if keyRE.MatchString(l.text) {
			return fmt.Errorf("settings.yml sets required_status_checks; main's checks belong in safe-settings")
		}
	}
	nameRE := regexp.MustCompile(`^\s*-\s+name:\s*["']?main["']?\s*$`)
	protRE := regexp.MustCompile(`^\s*protection:\s*(.*)$`)
	inBranches := false
	for i, l := range lines {
		if l.indent == 0 {
			inBranches = strings.HasPrefix(l.text, "branches:")
			continue
		}
		if !inBranches || !nameRE.MatchString(l.text) {
			continue
		}
		// main's item: the lines after it, indented deeper than its dash.
		for _, m := range lines[i+1:] {
			if m.indent <= l.indent {
				break
			}
			if v := protRE.FindStringSubmatch(m.text); v != nil {
				if value := strings.TrimSpace(v[1]); value == "null" || value == "~" {
					return nil
				}
				return fmt.Errorf("main's protection is set; it must be null")
			}
		}
		return fmt.Errorf("main's protection is missing; it must be null")
	}
	return fmt.Errorf("settings.yml has no main entry under branches with protection null")
}

func assertNeeds(t *testing.T, job, jobID string, want []string) {
	t.Helper()
	needs := needsJobs(t, job, jobID)
	sort.Strings(needs)
	sort.Strings(want)
	if strings.Join(needs, ",") != strings.Join(want, ",") {
		t.Errorf("job %q needs = %v, want exactly %v", jobID, needs, want)
	}
}

func needsJobs(t *testing.T, job, jobID string) []string {
	t.Helper()
	match := regexp.MustCompile(`(?m)^\s+needs:\s*\[([^\]]*)\]\s*$`).FindStringSubmatch(job)
	if match == nil {
		t.Fatalf("job %q does not declare an inline needs list", jobID)
	}
	var needs []string
	for _, item := range strings.Split(match[1], ",") {
		if item = strings.TrimSpace(item); item != "" {
			needs = append(needs, item)
		}
	}
	if len(needs) == 0 {
		t.Fatalf("job %q declares an empty needs list", jobID)
	}
	return needs
}

// TestCIGatingWorkflows_AggregateVerdictBehavior runs CI OK's verdict
// script against needed-job results, so weakening it (say, accepting a
// skipped Build and Test on a code PR) fails this test.
func TestCIGatingWorkflows_AggregateVerdictBehavior(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found")
	}
	workflow := readRepositoryFile(t, ".github", "workflows", "ci.yml")
	script := yamlLiteralBlock(t, yamlSection(t, workflow, "ci-ok", 2), "run")
	scriptPath := filepath.Join(t.TempDir(), "verdict.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		// scope result, code output, build-and-test, call-reusable-ci, gitleaks
		scope, code, build, lint, leaks string
		pass                            bool
	}{
		{"success", "true", "success", "success", "skipped", true},
		{"success", "false", "skipped", "skipped", "success", true},
		{"failure", "", "skipped", "skipped", "skipped", false},
		{"cancelled", "", "skipped", "skipped", "skipped", false},
		{"success", "", "skipped", "skipped", "skipped", false},
		{"success", "maybe", "success", "success", "skipped", false},
		{"success", "true", "failure", "success", "skipped", false},
		{"success", "true", "success", "failure", "skipped", false},
		{"success", "true", "success", "cancelled", "skipped", false},
		{"success", "true", "skipped", "success", "skipped", false},
		{"success", "true", "success", "skipped", "skipped", false},
		{"success", "true", "success", "success", "success", false},
		{"success", "true", "success", "neutral", "skipped", false},
		{"success", "false", "skipped", "skipped", "failure", false},
		{"success", "false", "skipped", "skipped", "skipped", false},
		{"success", "false", "success", "skipped", "success", false},
		{"success", "false", "skipped", "success", "success", false},
		{"success", "true", "", "", "", false},
	}
	for _, tc := range cases {
		name := fmt.Sprintf("scope=%s code=%s build=%s lint=%s gitleaks=%s", tc.scope, tc.code, tc.build, tc.lint, tc.leaks)
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(bash, "--noprofile", "--norc", "-eo", "pipefail", scriptPath)
			cmd.Env = []string{
				"PATH=/usr/bin:/bin",
				"SCOPE=" + tc.scope, "CODE=" + tc.code, "BUILD=" + tc.build,
				"LINT=" + tc.lint, "GITLEAKS=" + tc.leaks,
			}
			out, runErr := cmd.CombinedOutput()
			if tc.pass && runErr != nil {
				t.Errorf("verdict failed, want pass: %v\n%s", runErr, out)
			}
			if !tc.pass && runErr == nil {
				t.Errorf("verdict passed, want failure\n%s", out)
			}
		})
	}
}
