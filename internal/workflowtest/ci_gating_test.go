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
		{file: "ci_checks.yml", gatedJobID: "call-reusable-ci", aggregateName: "Standardized CI OK"},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.file, func(t *testing.T) {
			workflow := readRepositoryFile(t, ".github", "workflows", fixture.file)

			changesJob := yamlSection(t, workflow, "changes", 2)
			outputs := yamlSection(t, changesJob, "outputs", 4)
			if keys := childKeys(t, outputs, 6); strings.Join(keys, ",") != "code" {
				t.Errorf("changes job outputs = %v, want exactly [code]", keys)
			}

			gatedJob := yamlSection(t, workflow, fixture.gatedJobID, 2)
			assertNeeds(t, gatedJob, fixture.gatedJobID, []string{"changes"})
			assertMatches(t, gatedJob, `(?m)^\s+if:\s*needs\.changes\.outputs\.code\s*==\s*['"]true['"]\s*$`, "learnings gating condition")

			aggregateJob := yamlSection(t, workflow, "ci-ok", 2)
			assertMatches(t, aggregateJob, `(?m)^\s+name:\s*`+regexp.QuoteMeta(fixture.aggregateName)+`\s*$`, "aggregate job name")
			assertMatches(t, aggregateJob, `(?m)^\s+if:\s*always\(\)\s*$`, "aggregate always-run condition")
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

func TestCIGatingSettings_RequiredContextsAreAggregates(t *testing.T) {
	settings := readRepositoryFile(t, ".github", "settings.yml")
	checks := yamlSection(t, settings, "required_status_checks", 6)
	contexts := yamlSection(t, checks, "contexts", 8)
	got := yamlListItems(t, contexts, 10)
	sort.Strings(got)
	want := []string{"CI OK", "Standardized CI OK"}
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
