package forge

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile creates parent directories and writes a file under dir.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestResolveSkillRef_PrimaryRoot(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".opencode/skills/deploy/SKILL.md", "# deploy")

	got, err := ResolveSkillRef(root, "deploy")
	if err != nil {
		t.Fatalf("ResolveSkillRef: %v", err)
	}
	want := filepath.Join(".opencode", "skills", "deploy", "SKILL.md")
	if got != want {
		t.Errorf("resolved = %q, want %q", got, want)
	}
}

func TestResolveSkillRef_FallbackRoot(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".opencode/skill/deploy/SKILL.md", "# deploy")

	got, err := ResolveSkillRef(root, "deploy")
	if err != nil {
		t.Fatalf("ResolveSkillRef: %v", err)
	}
	want := filepath.Join(".opencode", "skill", "deploy", "SKILL.md")
	if got != want {
		t.Errorf("resolved = %q, want %q", got, want)
	}
}

func TestResolveSkillRef_FirstRootWins(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".opencode/skills/deploy/SKILL.md", "# primary")
	writeFile(t, root, ".opencode/skill/deploy/SKILL.md", "# fallback")

	got, err := ResolveSkillRef(root, "deploy")
	if err != nil {
		t.Fatalf("ResolveSkillRef: %v", err)
	}
	want := filepath.Join(".opencode", "skills", "deploy", "SKILL.md")
	if got != want {
		t.Errorf("resolved = %q, want %q (primary root must win)", got, want)
	}
}

func TestResolveSkillRef_RepoRelativePath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "scripts/migrate.sh", "#!/bin/sh")

	got, err := ResolveSkillRef(root, "scripts/migrate.sh")
	if err != nil {
		t.Fatalf("ResolveSkillRef: %v", err)
	}
	if got != filepath.Join("scripts", "migrate.sh") {
		t.Errorf("resolved = %q, want scripts/migrate.sh", got)
	}
}

func TestResolveSkillRef_AbsolutePathRejected(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "scripts", "migrate.sh")
	// The absolute target exists: rejection must come from the
	// absolute-path guard, not accidental non-existence.
	writeFile(t, filepath.Dir(filepath.Dir(abs)), "scripts/migrate.sh", "#!/bin/sh")

	_, err := ResolveSkillRef(t.TempDir(), abs)
	if !errors.Is(err, ErrSkillNotFound) {
		t.Fatalf("absolute path should fail with ErrSkillNotFound, got: %v", err)
	}
	if msg := err.Error(); !strings.Contains(msg, "absolute") {
		t.Errorf("error should state absolute paths are rejected, got: %s", msg)
	}
}

func TestResolveSkillRef_TraversalRejected(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir project root: %v", err)
	}
	// The traversal target exists outside the root: rejection must
	// come from the traversal guard, not accidental non-existence.
	writeFile(t, parent, "secret.sh", "#!/bin/sh")

	for _, ref := range []string{"../secret.sh", "a/../../secret.sh", ".."} {
		_, err := ResolveSkillRef(root, ref)
		if !errors.Is(err, ErrSkillNotFound) {
			t.Errorf("traversal ref %q should fail with ErrSkillNotFound, got: %v", ref, err)
		}
		if msg := err.Error(); !strings.Contains(msg, "traversal") {
			t.Errorf("error for %q should mention traversal, got: %s", ref, msg)
		}
	}
}

func TestResolveSkillRef_NormalizedPathResolves(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "scripts/migrate.sh", "#!/bin/sh")

	got, err := ResolveSkillRef(root, "scripts/../scripts/migrate.sh")
	if err != nil {
		t.Fatalf("ResolveSkillRef: %v", err)
	}
	if got != filepath.Join("scripts", "migrate.sh") {
		t.Errorf("resolved = %q, want cleaned scripts/migrate.sh", got)
	}
}

func TestResolveSkillRef_UnreadablePathFailsClosed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".opencode/skills/deploy/SKILL.md", "# deploy")

	skillsDir := filepath.Join(root, ".opencode", "skills")
	if err := os.Chmod(skillsDir, 0o000); err != nil {
		t.Fatalf("chmod skills dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(skillsDir, 0o755) })

	_, err := ResolveSkillRef(root, "deploy")
	if err == nil {
		t.Fatal("unreadable skill directory should fail closed")
	}
	if errors.Is(err, ErrSkillNotFound) {
		t.Fatalf("permission failure must not be reported as ErrSkillNotFound, got: %v", err)
	}
}

func TestResolveSkillRef_UnknownSkillFailsClosed(t *testing.T) {
	root := t.TempDir()

	_, err := ResolveSkillRef(root, "deploy")
	if !errors.Is(err, ErrSkillNotFound) {
		t.Fatalf("unknown skill should fail with ErrSkillNotFound, got: %v", err)
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "SKILL_NOT_FOUND") {
		t.Errorf("error should lead with SKILL_NOT_FOUND, got: %s", msg)
	}
	for _, location := range []string{".opencode/skills/deploy/SKILL.md", ".opencode/skill/deploy/SKILL.md"} {
		if !strings.Contains(msg, location) {
			t.Errorf("error should name searched location %s, got: %s", location, msg)
		}
	}
}

func TestResolveSkillRef_UnknownScriptPathFailsClosed(t *testing.T) {
	root := t.TempDir()

	_, err := ResolveSkillRef(root, "scripts/missing.sh")
	if !errors.Is(err, ErrSkillNotFound) {
		t.Fatalf("unknown script should fail with ErrSkillNotFound, got: %v", err)
	}
	if msg := err.Error(); !strings.Contains(msg, "scripts/missing.sh") {
		t.Errorf("error should name the missing path, got: %s", msg)
	}
}

func TestSpawnSubtask_ProjectPathDefaultsToCwd(t *testing.T) {
	// The default project root is "."; a skill dir in the working
	// directory resolves. t.Chdir resolves the default from within
	// a temp dir and restores the working directory afterwards.
	// (No test in this package uses t.Parallel.)
	root := t.TempDir()
	writeFile(t, root, ".opencode/skills/deploy/SKILL.md", "# deploy")

	t.Chdir(root)

	result, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
		Skill:              "deploy",
	})
	if err != nil {
		t.Fatalf("SpawnSubtask with default project_path: %v", err)
	}
	if result["skill"] != filepath.Join(".opencode", "skills", "deploy", "SKILL.md") {
		t.Errorf("skill = %v, want resolved default-root path", result["skill"])
	}
}

func TestSkillDirective_GoldenSkillName(t *testing.T) {
	got := SkillDirective("deploy")
	want := "Required skill: deploy\n\n" +
		"This task is classified as deterministic work. Load the skill (invoke the `skill` tool with name \"deploy\") " +
		"or run the script at deploy, and execute its full procedure. Do not substitute an improvised procedure."
	if got != want {
		t.Errorf("SkillDirective mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestSkillDirective_GoldenScriptPath(t *testing.T) {
	got := SkillDirective("scripts/migrate.sh")
	want := "Required skill: scripts/migrate.sh\n\n" +
		"This task is classified as deterministic work. Load the skill (invoke the `skill` tool with name \"scripts/migrate.sh\") " +
		"or run the script at scripts/migrate.sh, and execute its full procedure. Do not substitute an improvised procedure."
	if got != want {
		t.Errorf("SkillDirective mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestSpawnSubtask_EnvelopeRecordsDelegationFacts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".opencode/skills/deploy/SKILL.md", "# deploy")

	result, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
		Skill:              "deploy",
		ProjectPath:        root,
	})
	if err != nil {
		t.Fatalf("SpawnSubtask: %v", err)
	}
	if result["task_classification"] != ClassificationDeterministic {
		t.Errorf("task_classification = %v, want deterministic", result["task_classification"])
	}
	wantPath := filepath.Join(".opencode", "skills", "deploy", "SKILL.md")
	if result["skill"] != wantPath {
		t.Errorf("skill = %v, want %v", result["skill"], wantPath)
	}
	if result["skill_directive"] != SkillDirective("deploy") {
		t.Errorf("skill_directive = %v, want rendered directive", result["skill_directive"])
	}
	if _, ok := result["skill_absent_reason"]; ok {
		t.Error("envelope must not carry skill_absent_reason when absent")
	}
}

func TestSpawnSubtask_EnvelopeRecordsReasonVerbatim(t *testing.T) {
	result, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
		SkillAbsentReason:  "no skill exists; procedure is one-off",
	})
	if err != nil {
		t.Fatalf("SpawnSubtask: %v", err)
	}
	if result["skill_absent_reason"] != "no skill exists; procedure is one-off" {
		t.Errorf("skill_absent_reason = %v, want verbatim reason", result["skill_absent_reason"])
	}
	if _, ok := result["skill"]; ok {
		t.Error("envelope must not carry skill when only a reason is supplied")
	}
	if _, ok := result["skill_directive"]; ok {
		t.Error("envelope must not carry skill_directive when no reference is supplied")
	}
}

func TestSpawnSubtask_EnvelopePreservesExistingFields(t *testing.T) {
	result, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "Do thing",
		Files: []string{"a.go"}, Description: "desc", SharedContext: "ctx",
		TaskClassification: ClassificationJudgment,
	})
	if err != nil {
		t.Fatalf("SpawnSubtask: %v", err)
	}
	for _, key := range []string{"status", "bead_id", "epic_id", "title", "files", "description", "shared_context", "timestamp", "task_classification"} {
		if _, ok := result[key]; !ok {
			t.Errorf("envelope lost pre-existing key %q", key)
		}
	}
	if result["task_classification"] != ClassificationJudgment {
		t.Errorf("task_classification = %v, want judgment", result["task_classification"])
	}
}

func TestSubtaskPrompt_RendersSkillDirective(t *testing.T) {
	prompt := SubtaskPrompt(
		"worker-1", "cell-abc", "epic-xyz", "Run migration",
		[]string{"migrate.go"}, "ctx", "scripts/migrate.sh",
	)
	if !strings.Contains(prompt, "## Required Skill") {
		t.Error("prompt should contain a Required Skill section")
	}
	if !strings.Contains(prompt, SkillDirective("scripts/migrate.sh")) {
		t.Error("prompt should embed the fixed skill directive")
	}
}

func TestSubtaskPrompt_GoldenNoSkill(t *testing.T) {
	got := SubtaskPrompt(
		"worker-1", "cell-abc", "epic-xyz", "Implement auth",
		[]string{"auth.go", "auth_test.go"}, "shared ctx", "",
	)
	// Byte-exact golden of the no-reference prompt shape; existing
	// prompt consumers must see no diff when no skill is supplied.
	want := "# Subtask: Implement auth\n\n" +
		"**Agent:** worker-1\n" +
		"**Cell ID:** cell-abc\n" +
		"**Epic ID:** epic-xyz\n\n" +
		"## Files to Modify\n\n" +
		"- `auth.go`\n" +
		"- `auth_test.go`\n" +
		"\n" +
		"## Shared Context\n\n" +
		"shared ctx\n\n" +
		"## Instructions\n\n" +
		"1. Reserve files before editing using `comms_reserve`\n" +
		"2. Implement the changes described above\n" +
		"3. Write tests for new code\n" +
		"4. Report progress with `forge_progress`\n" +
		"5. Complete with `forge_complete` when done\n"
	if got != want {
		t.Errorf("no-skill prompt mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestSubtaskPrompt_GoldenWithSkill(t *testing.T) {
	got := SubtaskPrompt(
		"worker-1", "cell-abc", "epic-xyz", "Run migration",
		[]string{"migrate.go"}, "ctx", "scripts/migrate.sh",
	)
	want := "# Subtask: Run migration\n\n" +
		"**Agent:** worker-1\n" +
		"**Cell ID:** cell-abc\n" +
		"**Epic ID:** epic-xyz\n\n" +
		"## Files to Modify\n\n" +
		"- `migrate.go`\n" +
		"\n" +
		"## Shared Context\n\n" +
		"ctx\n\n" +
		"## Required Skill\n\n" +
		"Required skill: scripts/migrate.sh\n\n" +
		"This task is classified as deterministic work. Load the skill (invoke the `skill` tool with name \"scripts/migrate.sh\") " +
		"or run the script at scripts/migrate.sh, and execute its full procedure. Do not substitute an improvised procedure.\n\n" +
		"## Instructions\n\n" +
		"1. Reserve files before editing using `comms_reserve`\n" +
		"2. Implement the changes described above\n" +
		"3. Write tests for new code\n" +
		"4. Report progress with `forge_progress`\n" +
		"5. Complete with `forge_complete` when done\n"
	if got != want {
		t.Errorf("with-skill prompt mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestSubtaskPrompt(t *testing.T) {
	prompt := SubtaskPrompt(
		"worker-1", "cell-abc", "epic-xyz",
		"Implement auth",
		[]string{"auth.go", "auth_test.go"},
		"Use JWT tokens", "",
	)

	checks := []string{
		"Implement auth",
		"worker-1",
		"cell-abc",
		"epic-xyz",
		"auth.go",
		"auth_test.go",
		"JWT tokens",
		"comms_reserve",
		"forge_complete",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("prompt missing %q", check)
		}
	}
}

func TestSubtaskPrompt_NoFiles(t *testing.T) {
	prompt := SubtaskPrompt("agent", "id", "epic", "Title", nil, "", "")
	if strings.Contains(prompt, "Files to Modify") {
		t.Error("should not show files section when empty")
	}
}

func TestSpawnSubtask(t *testing.T) {
	result, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "Do thing",
		Files: []string{"a.go"}, Description: "desc", SharedContext: "ctx",
		TaskClassification: ClassificationJudgment,
	})
	if err != nil {
		t.Fatalf("SpawnSubtask: %v", err)
	}
	if result["status"] != "ready_to_spawn" {
		t.Errorf("status = %v, want %q", result["status"], "ready_to_spawn")
	}
	if result["bead_id"] != "cell-1" {
		t.Errorf("bead_id = %v, want %q", result["bead_id"], "cell-1")
	}
}

func TestSpawnSubtask_MissingBeadID(t *testing.T) {
	_, err := SpawnSubtask(SpawnParams{EpicID: "epic-1", Title: "title"})
	if err == nil {
		t.Error("expected error for missing bead_id")
	}
}

func TestSpawnSubtask_MissingEpicID(t *testing.T) {
	_, err := SpawnSubtask(SpawnParams{BeadID: "cell-1", Title: "title"})
	if err == nil {
		t.Error("expected error for missing epic_id")
	}
}

func TestSpawnSubtask_MissingTitle(t *testing.T) {
	_, err := SpawnSubtask(SpawnParams{BeadID: "cell-1", EpicID: "epic-1"})
	if err == nil {
		t.Error("expected error for missing title")
	}
}

func TestSpawnSubtask_ClassificationRequired(t *testing.T) {
	_, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
	})
	if err == nil {
		t.Fatal("expected error for missing task_classification")
	}
	msg := err.Error()
	if !strings.Contains(msg, "task_classification") {
		t.Errorf("error should name task_classification, got: %s", msg)
	}
	if !strings.Contains(msg, "accepted values: deterministic, judgment") {
		t.Errorf("error should list accepted values, got: %s", msg)
	}
}

func TestSpawnSubtask_ClassificationEnum(t *testing.T) {
	for _, classification := range []string{"", "Deterministic", "urgent", "determinstic"} {
		_, err := SpawnSubtask(SpawnParams{
			BeadID: "cell-1", EpicID: "epic-1", Title: "title",
			TaskClassification: classification,
		})
		if err == nil {
			t.Errorf("classification %q should be rejected", classification)
			continue
		}
		msg := err.Error()
		if !strings.Contains(msg, "deterministic") || !strings.Contains(msg, "judgment") {
			t.Errorf("error for %q should list accepted values, got: %s", classification, msg)
		}
	}
}

func TestSpawnSubtask_JudgmentWithoutReferencePasses(t *testing.T) {
	result, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationJudgment,
	})
	if err != nil {
		t.Fatalf("judgment without reference should pass: %v", err)
	}
	if result["status"] != "ready_to_spawn" {
		t.Errorf("status = %v, want ready_to_spawn", result["status"])
	}
}

func TestSpawnSubtask_MutuallyExclusiveFields(t *testing.T) {
	_, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
		Skill:              "deploy", SkillAbsentReason: "no skill exists",
	})
	if err == nil {
		t.Fatal("skill + skill_absent_reason should be rejected")
	}
	if msg := err.Error(); !strings.Contains(msg, "mutually exclusive") {
		t.Errorf("error should state mutual exclusivity, got: %s", msg)
	}
}

func TestSpawnSubtask_SkillWithJudgmentFails(t *testing.T) {
	_, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationJudgment,
		Skill:              "deploy",
	})
	if err == nil {
		t.Fatal("skill with judgment classification should be rejected")
	}
	if msg := err.Error(); !strings.Contains(msg, "valid only for deterministic") {
		t.Errorf("error should state skill is valid only for deterministic classification, got: %s", msg)
	}
}

func TestSpawnSubtask_ReasonWithJudgmentFails(t *testing.T) {
	_, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationJudgment,
		SkillAbsentReason:  "no skill exists",
	})
	if err == nil {
		t.Fatal("skill_absent_reason with judgment classification should be rejected")
	}
	if msg := err.Error(); !strings.Contains(msg, "valid only for deterministic") {
		t.Errorf("error should state skill_absent_reason is valid only for deterministic classification, got: %s", msg)
	}
}

func TestSpawnSubtask_DeterministicRequiresReferenceOrReason(t *testing.T) {
	_, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
	})
	if err == nil {
		t.Fatal("deterministic without skill or skill_absent_reason should fail closed")
	}
	if !errors.Is(err, ErrSkillReferenceRequired) {
		t.Fatalf("error should match ErrSkillReferenceRequired via errors.Is, got: %v", err)
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "SKILL_REFERENCE_REQUIRED") {
		t.Errorf("error message should lead with SKILL_REFERENCE_REQUIRED, got: %s", msg)
	}
	if !strings.Contains(msg, "skill_absent_reason") {
		t.Errorf("error message should carry remediation naming skill_absent_reason, got: %s", msg)
	}
}

func TestSpawnSubtask_DeterministicWithReasonPasses(t *testing.T) {
	result, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
		SkillAbsentReason:  "no skill exists for this task yet",
	})
	if err != nil {
		t.Fatalf("deterministic with skill_absent_reason should pass: %v", err)
	}
	if result["status"] != "ready_to_spawn" {
		t.Errorf("status = %v, want ready_to_spawn", result["status"])
	}
}

func TestSpawnSubtask_LengthBounds(t *testing.T) {
	_, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
		Skill:              strings.Repeat("a", 257),
	})
	if err == nil {
		t.Fatal("skill over 256 chars should be rejected")
	}
	if msg := err.Error(); !strings.Contains(msg, "at most 256") {
		t.Errorf("error should name the 256-char skill bound, got: %s", msg)
	}

	_, err = SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
		SkillAbsentReason:  strings.Repeat("a", 513),
	})
	if err == nil {
		t.Fatal("skill_absent_reason over 512 chars should be rejected")
	}
	if msg := err.Error(); !strings.Contains(msg, "at most 512") {
		t.Errorf("error should name the 512-char reason bound, got: %s", msg)
	}
}

func TestSpawnSubtask_LengthBoundaryValues(t *testing.T) {
	// A skill of exactly the maximum length passes structural
	// validation and fails only at resolution (the bound is inclusive).
	_, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
		Skill:              strings.Repeat("a", MaxSkillRefLen),
	})
	if !errors.Is(err, ErrSkillNotFound) {
		t.Fatalf("boundary-length skill should pass validation and fail at resolution, got: %v", err)
	}

	result, err := SpawnSubtask(SpawnParams{
		BeadID: "cell-1", EpicID: "epic-1", Title: "title",
		TaskClassification: ClassificationDeterministic,
		SkillAbsentReason:  strings.Repeat("a", MaxSkillReasonLen),
	})
	if err != nil {
		t.Fatalf("boundary-length reason should be accepted: %v", err)
	}
	if result["status"] != "ready_to_spawn" {
		t.Errorf("status = %v, want ready_to_spawn", result["status"])
	}
}

func TestCompleteSubtask(t *testing.T) {
	store := testStore(t)

	// Create a cell to complete.
	if _, err := store.DB.Exec(
		"INSERT INTO beads (id, title, type, status) VALUES (?, ?, ?, ?)",
		"cell-test", "Test task", "task", "in_progress",
	); err != nil {
		t.Fatalf("insert cell: %v", err)
	}

	result, err := CompleteSubtask(store, "cell-test", "All done", []string{"a.go"})
	if err != nil {
		t.Fatalf("CompleteSubtask: %v", err)
	}
	if result["status"] != "completed" {
		t.Errorf("status = %v, want %q", result["status"], "completed")
	}

	// Verify event was recorded.
	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM events WHERE type = 'subtask_complete'").Scan(&count); err != nil {
		t.Fatalf("scan completion count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 subtask_complete event, got %d", count)
	}
}

func TestCompleteSubtask_MissingBeadID(t *testing.T) {
	store := testStore(t)
	_, err := CompleteSubtask(store, "", "result", nil)
	if err == nil {
		t.Error("expected error for missing bead_id")
	}
}
