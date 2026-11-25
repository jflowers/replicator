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
	assertMatches(t, changesJob, `git diff --raw --no-renames 'HEAD\^1' HEAD\)"; then\s+echo "[^"]*"\s+exit 1`, "rename-proof first-parent raw diff that fails closed")
	assertMatches(t, changesJob, `(?m)^\s+\.uf/dewey/learnings/\?\*\.md\) ;;$`, "learnings are .md files under the exact prefix")
	assertMatches(t, changesJob, `(?m)^\s+000000\|100644\|100755\) ;;$`, "only regular files and deletions qualify")
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

// TestCIGatingWorkflows_DetectionScriptBehavior runs the changes job's
// script against real pull request merge commits, so the learnings rule
// is pinned by execution, not only by the text checks above.
func TestCIGatingWorkflows_DetectionScriptBehavior(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found")
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found")
	}
	workflow := readRepositoryFile(t, ".github", "workflows", "ci.yml")
	script := yamlLiteralBlock(t, yamlSection(t, workflow, "changes", 2), "run")
	scriptPath := filepath.Join(t.TempDir(), "detect.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}

	const learnings = ".uf/dewey/learnings/"
	scenarios := []struct {
		name string
		// pr changes the PR branch; main, when set, advances main after
		// the branch point so the merge base is stale.
		pr, main func(t *testing.T, r detectRepo)
		event    string // default pull_request
		wrongSHA bool   // HEAD_SHA does not match the merge's second parent
		noMerge  bool   // HEAD is a plain commit, not a merge
		want     string // "true", "false", or "fail" for a non-zero exit
	}{
		{name: "learning added", pr: func(t *testing.T, r detectRepo) { r.write(t, learnings+"new.md", "x\n") }, want: "false"},
		{name: "learning edited", pr: func(t *testing.T, r detectRepo) { r.write(t, learnings+"old.md", "edited\n") }, want: "false"},
		{name: "learning deleted", pr: func(t *testing.T, r detectRepo) { r.git(t, "rm", "-q", learnings+"old.md") }, want: "false"},
		{
			name: "learning added while main moved on",
			pr:   func(t *testing.T, r detectRepo) { r.write(t, learnings+"new.md", "x\n") },
			main: func(t *testing.T, r detectRepo) { r.write(t, "main.go", "package main\n\nfunc later() {}\n") },
			want: "false",
		},
		{name: "code file", pr: func(t *testing.T, r detectRepo) { r.write(t, "main.go", "package main\n\nfunc f() {}\n") }, want: "true"},
		{
			name: "learning plus code",
			pr: func(t *testing.T, r detectRepo) {
				r.write(t, learnings+"new.md", "x\n")
				r.write(t, "go.mod", "module example\n")
			},
			want: "true",
		},
		{name: "non-markdown file under learnings", pr: func(t *testing.T, r detectRepo) { r.write(t, learnings+"notes.txt", "x\n") }, want: "true"},
		{name: "look-alike directory", pr: func(t *testing.T, r detectRepo) { r.write(t, ".uf/dewey/learnings-old/a.md", "x\n") }, want: "true"},
		{name: "code file renamed into learnings", pr: func(t *testing.T, r detectRepo) { r.git(t, "mv", "main.go", learnings+"main.md") }, want: "true"},
		{
			name: "symlink under learnings",
			pr: func(t *testing.T, r detectRepo) {
				if err := os.Symlink("../../../main.go", filepath.Join(r.dir, learnings, "link.md")); err != nil {
					t.Fatal(err)
				}
			},
			want: "true",
		},
		{
			name: "submodule under learnings",
			pr: func(t *testing.T, r detectRepo) {
				r.git(t, "update-index", "--add", "--cacheinfo", "160000,"+r.head(t)+","+learnings+"sub.md")
			},
			want: "true",
		},
		{name: "empty diff", pr: func(t *testing.T, r detectRepo) {}, want: "true"},
		{name: "push event", pr: func(t *testing.T, r detectRepo) { r.write(t, learnings+"new.md", "x\n") }, event: "push", want: "true"},
		{name: "HEAD not a merge commit", pr: func(t *testing.T, r detectRepo) { r.write(t, learnings+"new.md", "x\n") }, noMerge: true, want: "fail"},
		{name: "merge not for the PR head", pr: func(t *testing.T, r detectRepo) { r.write(t, learnings+"new.md", "x\n") }, wrongSHA: true, want: "fail"},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			r := detectRepo{dir: t.TempDir(), git: nil}
			r.git = func(t *testing.T, args ...string) { t.Helper(); runGit(t, gitPath, r.dir, args...) }
			r.git(t, "init", "-q", "-b", "main")
			r.git(t, "config", "user.email", "test@example.com")
			r.git(t, "config", "user.name", "Test")
			r.git(t, "config", "commit.gpgsign", "false")
			r.write(t, "main.go", "package main\n")
			r.write(t, learnings+"old.md", "old\n")
			r.git(t, "add", "-A")
			r.git(t, "commit", "-q", "-m", "base")

			r.git(t, "checkout", "-q", "-b", "pr")
			sc.pr(t, r)
			r.git(t, "add", "-A")
			r.git(t, "commit", "-q", "--allow-empty", "-m", "pr")
			prHead := r.head(t)
			if !sc.noMerge {
				r.git(t, "checkout", "-q", "main")
				if sc.main != nil {
					sc.main(t, r)
					r.git(t, "add", "-A")
					r.git(t, "commit", "-q", "-m", "main moves on")
				}
				r.git(t, "merge", "-q", "--no-ff", "--no-edit", "pr")
			}
			headSHA := prHead
			if sc.wrongSHA {
				headSHA = strings.Repeat("0", 40)
			}
			event := sc.event
			if event == "" {
				event = "pull_request"
			}

			outputPath := filepath.Join(t.TempDir(), "output")
			cmd := exec.Command(bash, "--noprofile", "--norc", "-eo", "pipefail", scriptPath)
			cmd.Dir = r.dir
			cmd.Env = []string{
				"PATH=" + filepath.Dir(gitPath) + ":/usr/bin:/bin",
				"HOME=" + t.TempDir(),
				"GIT_CONFIG_NOSYSTEM=1",
				"GITHUB_EVENT_NAME=" + event,
				"GITHUB_OUTPUT=" + outputPath,
				"HEAD_SHA=" + headSHA,
			}
			out, runErr := cmd.CombinedOutput()
			if sc.want == "fail" {
				if runErr == nil {
					t.Fatalf("script exited 0, want a failure\n%s", out)
				}
				return
			}
			if runErr != nil {
				t.Fatalf("script failed: %v\n%s", runErr, out)
			}
			written, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("read GITHUB_OUTPUT: %v\n%s", err, out)
			}
			if got := strings.TrimSpace(string(written)); got != "code="+sc.want {
				t.Errorf("GITHUB_OUTPUT = %q, want %q\n%s", got, "code="+sc.want, out)
			}
		})
	}
}

type detectRepo struct {
	dir string
	git func(t *testing.T, args ...string)
}

func (r detectRepo) write(t *testing.T, path, content string) {
	t.Helper()
	full := filepath.Join(r.dir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func (r detectRepo) head(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "-C", r.dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func runGit(t *testing.T, gitPath, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(gitPath, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}
