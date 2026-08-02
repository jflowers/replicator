package agentkit

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffold_FreshDirectory(t *testing.T) {
	dir := t.TempDir()
	results, err := Scaffold(dir, false)
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	// Expect 15 files: 5 commands + 7 skills + 3 agents.
	if len(results) != 15 {
		t.Errorf("Scaffold returned %d results, want 15", len(results))
		for _, r := range results {
			t.Logf("  %s: %s", r.Action, r.Path)
		}
	}

	// All should be "created".
	for _, r := range results {
		if r.Action != "created" {
			t.Errorf("result %s: action = %q, want %q", r.Path, r.Action, "created")
		}
	}

	// Spot-check a few files exist on disk.
	checks := []string{
		".opencode/commands/forge.md",
		".opencode/commands/org.md",
		".opencode/commands/inbox.md",
		".opencode/commands/forge-status.md",
		".opencode/commands/handoff.md",
		".opencode/skills/always-on-guidance/SKILL.md",
		".opencode/skills/forge-coordination/SKILL.md",
		".opencode/skills/replicator-cli/SKILL.md",
		".opencode/skills/testing-patterns/SKILL.md",
		".opencode/skills/system-design/SKILL.md",
		".opencode/skills/learning-systems/SKILL.md",
		".opencode/skills/forge-global/SKILL.md",
		".opencode/agents/coordinator.md",
		".opencode/agents/worker.md",
		".opencode/agents/background-worker.md",
	}
	for _, rel := range checks {
		full := filepath.Join(dir, rel)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("expected file %s to exist: %v", rel, err)
		}
	}
}

func TestScaffold_SkipsExisting(t *testing.T) {
	dir := t.TempDir()

	// Pre-create a file that Scaffold would write.
	forgePath := filepath.Join(dir, ".opencode", "commands", "forge.md")
	os.MkdirAll(filepath.Dir(forgePath), 0o755)
	original := []byte("# custom content\n")
	os.WriteFile(forgePath, original, 0o644)

	results, err := Scaffold(dir, false)
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	// Find the forge.md result — should be "skipped".
	var found bool
	for _, r := range results {
		if r.Path == filepath.Join("commands", "forge.md") {
			found = true
			if r.Action != "skipped" {
				t.Errorf("forge.md action = %q, want %q", r.Action, "skipped")
			}
		}
	}
	if !found {
		t.Error("forge.md not found in results")
	}

	// Verify file content was NOT overwritten.
	data, _ := os.ReadFile(forgePath)
	if string(data) != string(original) {
		t.Errorf("forge.md was overwritten: got %q", string(data))
	}
}

func TestScaffold_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()

	// Pre-create a file that Scaffold would write.
	forgePath := filepath.Join(dir, ".opencode", "commands", "forge.md")
	os.MkdirAll(filepath.Dir(forgePath), 0o755)
	original := []byte("# custom content\n")
	os.WriteFile(forgePath, original, 0o644)

	results, err := Scaffold(dir, true)
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	// Find the forge.md result — should be "overwritten".
	for _, r := range results {
		if r.Path == filepath.Join("commands", "forge.md") {
			if r.Action != "overwritten" {
				t.Errorf("forge.md action = %q, want %q", r.Action, "overwritten")
			}
		}
	}

	// Verify file content WAS overwritten with embedded content.
	data, _ := os.ReadFile(forgePath)
	if string(data) == string(original) {
		t.Error("forge.md was NOT overwritten despite force=true")
	}
}

func TestScaffold_FileCount(t *testing.T) {
	dir := t.TempDir()
	results, err := Scaffold(dir, false)
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	// Count by category.
	var commands, skills, agents int
	for _, r := range results {
		switch {
		case strings.HasPrefix(r.Path, "commands/"):
			commands++
		case strings.HasPrefix(r.Path, "skills/"):
			skills++
		case strings.HasPrefix(r.Path, "agents/"):
			agents++
		}
	}

	if commands != 5 {
		t.Errorf("commands = %d, want 5", commands)
	}
	if skills != 7 {
		t.Errorf("skills = %d, want 7", skills)
	}
	if agents != 3 {
		t.Errorf("agents = %d, want 3", agents)
	}
}

func TestHandoffMD_StructuralHardening(t *testing.T) {
	data, err := content.ReadFile("content/commands/handoff.md")
	if err != nil {
		t.Fatalf("read embedded handoff.md: %v", err)
	}
	text := string(data)

	// (1) Ordering constraint text appears before the first numbered step.
	const orderingConstraint = "Steps MUST execute in this exact order"
	constraintIdx := strings.Index(text, orderingConstraint)
	if constraintIdx < 0 {
		t.Error("handoff.md: missing ordering constraint text")
	}
	// Find first numbered step (line starting with "0." or "1.").
	firstStepIdx := strings.Index(text, "\n0.")
	if firstStepIdx < 0 {
		firstStepIdx = strings.Index(text, "\n1.")
	}
	if firstStepIdx < 0 {
		t.Error("handoff.md: no numbered workflow steps found")
	} else if constraintIdx >= firstStepIdx {
		t.Error("handoff.md: ordering constraint must appear before the first numbered step")
	}

	// (2) Handoff note categories appear within the org_session_end step section.
	sessionEndIdx := strings.Index(text, "org_session_end")
	if sessionEndIdx < 0 {
		t.Fatal("handoff.md: missing org_session_end reference")
	}
	afterSessionEnd := text[sessionEndIdx:]
	categories := []string{"Completed", "In Progress", "Blocked", "Next Steps", "Gotchas"}
	for _, cat := range categories {
		if !strings.Contains(afterSessionEnd, cat) {
			t.Errorf("handoff.md: handoff note category %q not found after org_session_end", cat)
		}
	}

	// (3) Separate "## Handoff Note Template" section is absent.
	if strings.Contains(text, "## Handoff Note Template") {
		t.Error("handoff.md: separate '## Handoff Note Template' section should be removed")
	}

	// (4) Forge precondition check text is present.
	if !strings.Contains(text, "Forge precondition check") {
		t.Error("handoff.md: missing forge precondition check step")
	}

	// (5) Step dependency rationale is present for each step (2-5).
	for _, dep := range []string{
		"Depends on step 1",
		"Depends on step 2",
		"Depends on step 3",
		"Depends on step 4",
	} {
		if !strings.Contains(text, dep) {
			t.Errorf("handoff.md: missing dependency rationale %q", dep)
		}
	}
}

func TestCoordinatorPrompt_StructuralResilience(t *testing.T) {
	// Verify structural properties of coordinator.md that ensure
	// compression resilience: identity-first opening, constraints
	// before protocol, uppercase keywords, and behavioral parity.
	data, err := content.ReadFile("content/agents/coordinator.md")
	if err != nil {
		t.Fatalf("read coordinator.md: %v", err)
	}

	text := string(data)

	// --- YAML front matter validation ---
	if !strings.HasPrefix(text, "---\n") {
		t.Fatal("coordinator.md: missing opening frontmatter delimiter")
	}
	endIdx := strings.Index(text[4:], "\n---")
	if endIdx < 0 {
		t.Fatal("coordinator.md: missing closing frontmatter delimiter")
	}
	frontmatter := text[4 : 4+endIdx]

	fmChecks := []string{"name: coordinator", "mode: subagent", "description:"}
	for _, want := range fmChecks {
		if !strings.Contains(frontmatter, want) {
			t.Errorf("frontmatter missing %q", want)
		}
	}

	// Body is everything after the closing "---\n".
	body := text[4+endIdx+4:] // skip past "\n---\n"

	// --- Identity-first opening contains prohibition ---
	// Find the first non-heading paragraph. Skip past the "# Heading"
	// line and any blank lines to find the identity statement.
	// Split into paragraphs (separated by blank lines) and find the
	// first one that is not a heading.
	paragraphs := strings.Split(body, "\n\n")
	firstParagraph := ""
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		firstParagraph = trimmed
		break
	}
	if firstParagraph == "" {
		t.Fatal("coordinator.md: no non-heading paragraph found in body")
	}
	if !strings.Contains(firstParagraph, "NEVER") {
		t.Error("first paragraph missing uppercase 'NEVER' keyword")
	}
	if !strings.Contains(strings.ToLower(firstParagraph), "reserve") {
		t.Error("first paragraph missing file reservation prohibition ('reserve')")
	}

	// --- Section ordering: Critical Constraints before Protocol before Available Tools ---
	constraintsIdx := strings.Index(body, "## Critical Constraints")
	protocolIdx := strings.Index(body, "## Protocol")
	toolsIdx := strings.Index(body, "## Available Tools")

	if constraintsIdx < 0 {
		t.Error("missing '## Critical Constraints' section")
	}
	if protocolIdx < 0 {
		t.Error("missing '## Protocol' section")
	}
	if toolsIdx < 0 {
		t.Error("missing '## Available Tools' section")
	}

	if constraintsIdx >= 0 && protocolIdx >= 0 && constraintsIdx >= protocolIdx {
		t.Error("'## Critical Constraints' must appear before '## Protocol'")
	}
	if protocolIdx >= 0 && toolsIdx >= 0 && protocolIdx >= toolsIdx {
		t.Error("'## Protocol' must appear before '## Available Tools'")
	}

	// --- Behavioral rule markers: all 6 original rules must be present ---
	ruleMarkers := map[string]string{
		"comms_init":      "comms init rule",
		"reserve":         "file reservation rule",
		"forge_review":    "review completions rule",
		"hivemind_store":  "store learnings rule",
		"comms_inbox":     "check inbox rule",
		"forge_broadcast": "broadcast context rule",
	}
	bodyLower := strings.ToLower(body)
	for marker, desc := range ruleMarkers {
		if !strings.Contains(bodyLower, strings.ToLower(marker)) {
			t.Errorf("missing behavioral rule marker %q (%s)", marker, desc)
		}
	}

	// --- Uppercase RFC 2119 keywords in constraints ---
	constraintsSection := ""
	if constraintsIdx >= 0 && protocolIdx >= 0 {
		constraintsSection = body[constraintsIdx:protocolIdx]
	}
	if constraintsSection != "" {
		if !strings.Contains(constraintsSection, "NEVER") {
			t.Error("Critical Constraints section missing uppercase 'NEVER' keyword")
		}
		if !strings.Contains(constraintsSection, "MUST") {
			t.Error("Critical Constraints section missing uppercase 'MUST' keyword")
		}
		// Verify every bullet line in constraints has an uppercase keyword.
		for _, line := range strings.Split(constraintsSection, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "-") {
				hasKeyword := strings.Contains(line, "NEVER") ||
					strings.Contains(line, "MUST") ||
					strings.Contains(line, "ALWAYS")
				if !hasKeyword {
					t.Errorf("constraint line missing uppercase RFC 2119 keyword: %q", line)
				}
			}
		}
	}

	// --- FR-003: Explicit review-before-complete ordering ---
	if !strings.Contains(body, "BEFORE") ||
		!strings.Contains(body, "forge_review") ||
		!strings.Contains(body, "forge_complete") {
		t.Error("missing explicit ordering: forge_review BEFORE forge_complete (FR-003)")
	}
	// Verify forge_review appears before forge_complete in constraints section.
	if constraintsSection != "" {
		reviewIdx := strings.Index(constraintsSection, "forge_review")
		completeIdx := strings.Index(constraintsSection, "forge_complete")
		if reviewIdx >= 0 && completeIdx >= 0 && reviewIdx >= completeIdx {
			t.Error("forge_review must appear before forge_complete in Critical Constraints")
		}
	}

	// --- Compression resilience: first 50% of body lines contain critical constraints ---
	bodyLines := strings.Split(body, "\n")
	halfLen := len(bodyLines) / 2
	firstHalf := strings.Join(bodyLines[:halfLen], "\n")
	firstHalfLower := strings.ToLower(firstHalf)

	if !strings.Contains(firstHalf, "NEVER") || !strings.Contains(firstHalfLower, "reserve") {
		t.Error("first 50%% of body lines must contain file reservation prohibition (NEVER + reserve)")
	}
	if !strings.Contains(firstHalfLower, "forge_review") || !strings.Contains(firstHalfLower, "forge_complete") {
		t.Error("first 50%% of body lines must contain review-before-complete ordering (forge_review + forge_complete)")
	}
	if !strings.Contains(firstHalfLower, "edit code") {
		t.Error("first 50%% of body lines must contain code editing prohibition ('edit code')")
	}
}

func TestSkillTemplates_HaveNameField(t *testing.T) {
	// Walk the embedded content filesystem and verify every SKILL.md
	// has a "name: <directory-name>" field in its YAML frontmatter.
	var checked int

	err := fs.WalkDir(content, "content/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}

		data, readErr := content.ReadFile(path)
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}

		text := string(data)

		// Extract directory name (e.g., "always-on-guidance" from
		// "content/skills/always-on-guidance/SKILL.md").
		dir := filepath.Dir(path)
		dirName := filepath.Base(dir)

		// Verify frontmatter delimiters exist.
		if !strings.HasPrefix(text, "---\n") {
			t.Errorf("%s: missing opening frontmatter delimiter", path)
			return nil
		}
		endIdx := strings.Index(text[4:], "\n---")
		if endIdx < 0 {
			t.Errorf("%s: missing closing frontmatter delimiter", path)
			return nil
		}
		frontmatter := text[4 : 4+endIdx]

		// Check for "name: <dirName>" line.
		wantLine := "name: " + dirName
		found := false
		for _, line := range strings.Split(frontmatter, "\n") {
			if strings.TrimSpace(line) == wantLine {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: frontmatter missing %q", path, wantLine)
		}

		checked++
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded skills: %v", err)
	}
	if checked != 7 {
		t.Errorf("checked %d skill templates, want 7", checked)
	}
}
