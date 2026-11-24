package workflowtest

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type gatingWorkflowFixture struct {
	file          string
	gatedJobID    string
	aggregateName string
}

func TestCIGatingWorkflows_StructureEnforcesLearningsSkip(t *testing.T) {
	fixtures := []gatingWorkflowFixture{
		{file: "ci.yml", gatedJobID: "build-and-test", aggregateName: "CI OK"},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.file, func(t *testing.T) {
			workflow := readRepositoryFile(t, ".github", "workflows", fixture.file)

			changesJob := yamlSection(t, workflow, "changes", 2)
			outputs := yamlSection(t, changesJob, "outputs", 4)
			if keys := childKeys(t, outputs, 6); strings.Join(keys, ",") != "code" {
				t.Errorf("changes job outputs = %v, want exactly [code]", keys)
			}

			assertDetectionLogic(t, changesJob)

			gatedJob := yamlSection(t, workflow, fixture.gatedJobID, 2)
			assertNeeds(t, gatedJob, fixture.gatedJobID, []string{"changes"})
			assertMatches(t, gatedJob, `(?m)^\s+if:\s*needs\.changes\.outputs\.code\s*==\s*['"]true['"]\s*$`, "learnings gating condition")

			aggregateJob := yamlSection(t, workflow, "ci-ok", 2)
			assertMatches(t, aggregateJob, `(?m)^\s+name:\s*`+regexp.QuoteMeta(fixture.aggregateName)+`\s*$`, "aggregate job name")
			// The only exemption from always() is a cancelled push run; a
			// skipped aggregate satisfies a required check, so a cancelled
			// pull_request run must never skip it.
			assertMatches(t, aggregateJob, `(?m)^\s+if:\s*always\(\) && !\(github\.event_name == 'push' && cancelled\(\)\)\s*$`, "aggregate always-run condition, exempting only cancelled push runs")
			assertNeeds(t, aggregateJob, "ci-ok", []string{"changes", fixture.gatedJobID})
			assertMatches(t, aggregateJob, `contains\(needs\.\*\.result,\s*['"]failure['"]\)`, "aggregate failure propagation")
			assertMatches(t, aggregateJob, `contains\(needs\.\*\.result,\s*['"]cancelled['"]\)`, "aggregate cancellation propagation")
			assertMatches(t, aggregateJob, `(?m)^\s+exit 1\s*$`, "aggregate failure exit")

			onBlock := yamlSection(t, workflow, "on", 0)
			for _, filter := range []string{"paths", "paths-ignore"} {
				if regexp.MustCompile(`(?m)^\s+` + filter + `:\s`).MatchString(onBlock) {
					t.Errorf("on: block must not filter triggers with %q", filter)
				}
			}
		})
	}
}

// assertDetectionLogic pins the parts of the changes job that keep the
// skip safe (CLS-FR-001): push short-circuit, a rename-proof diff of the
// PR merge commit against its first parent (so a stale
// pull_request.base.sha cannot count main's own changes), the exact
// learnings prefix, and failing closed.
func assertDetectionLogic(t *testing.T, changesJob string) {
	t.Helper()
	assertMatches(t, changesJob, `(?m)^\s+- if: github\.event_name == 'pull_request'\n\s+uses: actions/checkout@`, "checkout only on pull_request")
	assertMatches(t, changesJob, `(?m)^\s+fetch-depth:\s*2\s*$`, "merge commit and both parents fetched")
	assertMatches(t, changesJob, `if \[ "\$GITHUB_EVENT_NAME" = "push" \]; then\s+echo "[^"]*"\s+echo "code=true" >> "\$GITHUB_OUTPUT"\s+exit 0`, "push short-circuit to code=true")
	assertMatches(t, changesJob, `git rev-parse --verify --quiet 'HEAD\^2\^\{commit\}'\)"; then\s+echo "[^"]*"\s+exit 1`, "fail closed when HEAD is not a merge commit")
	assertMatches(t, changesJob, `if \[ "\$PR_PARENT" != "\$HEAD_SHA" \]; then\s+echo "[^"]*"\s+exit 1`, "fail closed when the merge commit is not for the PR head")
	assertMatches(t, changesJob, `git diff --name-only --no-renames 'HEAD\^1' HEAD\)"; then\s+echo "[^"]*"\s+exit 1`, "rename-proof first-parent diff that fails closed")
	assertMatches(t, changesJob, `(?m)^\s+""\|\.uf/dewey/learnings/\*\) ;;$`, "exact learnings prefix")
	if strings.Contains(changesJob, "pull_request.base.sha") {
		t.Error("changes job must not diff against pull_request.base.sha: it can be older than the main merged into the PR")
	}
}

// TestCIGatingWorkflows_StandardizedCIIsNeverGated guards CLS-FR-002:
// the reusable call bundles secret scanning (betterleaks, KICS) and PR
// title validation, which must run on every pull request -- including
// learnings-only ones, since learnings files can carry leaked secrets.
func TestCIGatingWorkflows_StandardizedCIIsNeverGated(t *testing.T) {
	workflow := readRepositoryFile(t, ".github", "workflows", "ci_checks.yml")

	jobs := yamlSection(t, workflow, "jobs", 0)
	if keys := childKeys(t, jobs, 2); strings.Join(keys, ",") != "call-reusable-ci" {
		t.Errorf("ci_checks.yml jobs = %v, want exactly [call-reusable-ci]", keys)
	}

	reusableCall := yamlSection(t, workflow, "call-reusable-ci", 2)
	for _, key := range []string{"if", "needs"} {
		if regexp.MustCompile(`(?m)^ {4}` + key + `:`).MatchString(reusableCall) {
			t.Errorf("call-reusable-ci must not declare %q: secret scanning and PR title validation run on every PR", key)
		}
	}
	assertMatches(t, reusableCall, `(?m)^\s+uses:\s*complytime/org-infra/\.github/workflows/reusable_ci\.yml@`, "org-infra reusable CI call")

	onBlock := yamlSection(t, workflow, "on", 0)
	for _, filter := range []string{"paths", "paths-ignore"} {
		if regexp.MustCompile(`(?m)^\s+` + filter + `:\s`).MatchString(onBlock) {
			t.Errorf("on: block must not filter triggers with %q", filter)
		}
	}
}

func TestCIGatingSettings_RequiredContextsAreAggregates(t *testing.T) {
	settings := readRepositoryFile(t, ".github", "settings.yml")
	checks := yamlSection(t, settings, "required_status_checks", 6)
	contexts := yamlSection(t, checks, "contexts", 8)
	got := yamlListItems(t, contexts, 10)
	sort.Strings(got)
	want := []string{"CI OK", "Standardized CI / Run linters"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("required_status_checks.contexts = %v, want exactly %v", got, want)
	}
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

func yamlListItems(t *testing.T, section string, indentation int) []string {
	t.Helper()
	pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^ {%d}- (.+)$`, indentation))
	matches := pattern.FindAllStringSubmatch(section, -1)
	if len(matches) == 0 {
		t.Fatalf("YAML section has no list items at indentation %d", indentation)
	}
	items := make([]string, 0, len(matches))
	for _, match := range matches {
		items = append(items, strings.Trim(strings.TrimSpace(match[1]), `"'`))
	}
	return items
}
