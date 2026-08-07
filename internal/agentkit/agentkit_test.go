package agentkit

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findRepoRoot walks up from the current working directory to find the
// directory containing go.mod, which is the repository root.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

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

	// Parse front matter and body once for all sub-tests.
	if !strings.HasPrefix(text, "---\n") {
		t.Fatal("coordinator.md: missing opening frontmatter delimiter")
	}
	endIdx := strings.Index(text[4:], "\n---")
	if endIdx < 0 {
		t.Fatal("coordinator.md: missing closing frontmatter delimiter")
	}
	frontmatter := text[4 : 4+endIdx]
	body := text[4+endIdx+4:] // skip past "\n---\n"

	// Pre-compute shared indices used by multiple sub-tests.
	constraintsIdx := strings.Index(body, "## Critical Constraints")
	protocolIdx := strings.Index(body, "## Protocol")
	toolsIdx := strings.Index(body, "## Available Tools")
	bodyLower := strings.ToLower(body)

	constraintsSection := ""
	if constraintsIdx >= 0 && protocolIdx >= 0 {
		constraintsSection = body[constraintsIdx:protocolIdx]
	}

	t.Run("YAML_frontmatter", func(t *testing.T) {
		fmChecks := []string{"name: coordinator", "mode: subagent", "description:"}
		for _, want := range fmChecks {
			if !strings.Contains(frontmatter, want) {
				t.Errorf("frontmatter missing %q", want)
			}
		}
	})

	t.Run("identity_first_opening", func(t *testing.T) {
		// Find the first non-heading paragraph after the heading.
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
	})

	t.Run("section_ordering", func(t *testing.T) {
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
	})

	t.Run("behavioral_rule_markers", func(t *testing.T) {
		// All 7 behavioral rules (6 original + 1 codified) must be present (FR-006).
		// Use a slice for deterministic iteration order.
		type ruleCheck struct {
			marker string
			desc   string
		}
		checks := []ruleCheck{
			{"comms_init", "comms init rule"},
			{"reserve", "file reservation rule"},
			{"edit code", "code editing prohibition rule"},
			{"forge_review", "review completions rule"},
			{"hivemind_store", "store learnings rule"},
			{"comms_inbox", "check inbox rule"},
			{"forge_broadcast", "broadcast context rule"},
		}
		for _, rc := range checks {
			if !strings.Contains(bodyLower, strings.ToLower(rc.marker)) {
				t.Errorf("missing behavioral rule marker %q (%s)", rc.marker, rc.desc)
			}
		}
	})

	t.Run("uppercase_RFC2119_keywords", func(t *testing.T) {
		if constraintsSection == "" {
			t.Skip("no constraints section found")
		}
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
	})

	t.Run("review_before_complete_ordering", func(t *testing.T) {
		// FR-003: forge_review MUST be called BEFORE forge_complete.
		if constraintsSection == "" {
			t.Skip("no constraints section found")
		}
		// Verify each required element is present individually.
		if !strings.Contains(body, "forge_review") {
			t.Error("missing forge_review reference in body")
		}
		if !strings.Contains(body, "forge_complete") {
			t.Error("missing forge_complete reference in body")
		}
		// Verify the constraint line containing forge_review also mentions
		// forge_complete with explicit ordering language (BEFORE), confirming
		// the ordering is stated in a single rule rather than inferred from
		// unrelated occurrences of "BEFORE".
		foundOrderingRule := false
		for _, line := range strings.Split(constraintsSection, "\n") {
			if strings.Contains(line, "forge_review") &&
				strings.Contains(line, "forge_complete") &&
				strings.Contains(line, "BEFORE") {
				foundOrderingRule = true
				break
			}
		}
		if !foundOrderingRule {
			t.Error("no single constraint line connects forge_review, BEFORE, and forge_complete (FR-003)")
		}
		// Verify forge_review appears before forge_complete in constraints section.
		reviewIdx := strings.Index(constraintsSection, "forge_review")
		completeIdx := strings.Index(constraintsSection, "forge_complete")
		if reviewIdx >= 0 && completeIdx >= 0 && reviewIdx >= completeIdx {
			t.Error("forge_review must appear before forge_complete in Critical Constraints")
		}
	})

	t.Run("compression_resilience", func(t *testing.T) {
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
	})
}

func TestWorkerPrompt_HardenedStructure(t *testing.T) {
	// Verify the hardened worker.md has inline constraints, no separate
	// Constraints section, mandatory language, reservation failure recovery,
	// and stays under 35 lines.
	data, err := content.ReadFile("content/agents/worker.md")
	if err != nil {
		t.Fatalf("read worker.md: %v", err)
	}
	text := string(data)
	lines := strings.Split(text, "\n")

	// Line count: must be <= 35 (design decision D4).
	if len(lines) > 35 {
		t.Errorf("worker.md has %d lines, want <= 35", len(lines))
	}

	// No separate "## Constraints" section.
	if strings.Contains(text, "## Constraints") {
		t.Error("worker.md still contains a '## Constraints' heading; should be removed")
	}

	// stepWindow returns the text of the line at idx plus the next windowSize
	// lines, joined together. This allows assertions that tolerate content
	// being split across sub-bullets without requiring single-line co-location.
	const windowSize = 3
	stepWindow := func(idx int) string {
		end := idx + windowSize + 1
		if end > len(lines) {
			end = len(lines)
		}
		return strings.Join(lines[idx:end], "\n")
	}

	// Find the comms_reserve step and verify it contains MUST or NEVER
	// language about file editing.
	reserveIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "comms_reserve") {
			reserveIdx = i
			break
		}
	}
	if reserveIdx == -1 {
		t.Fatal("worker.md: no line containing 'comms_reserve' found")
	}
	reserveStep := lines[reserveIdx]
	if !strings.Contains(reserveStep, "MUST") && !strings.Contains(reserveStep, "NEVER") {
		t.Errorf("comms_reserve step lacks MUST/NEVER constraint: %q", reserveStep)
	}
	if !strings.Contains(reserveStep, "NEVER") {
		t.Errorf("comms_reserve step missing NEVER constraint about file editing: %q", reserveStep)
	}

	// Verify forge_progress step contains "MUST".
	var progressStep string
	for _, line := range lines {
		if strings.Contains(line, "forge_progress") {
			progressStep = line
			break
		}
	}
	if progressStep == "" {
		t.Fatal("worker.md: no line containing 'forge_progress' found")
	}
	if !strings.Contains(progressStep, "MUST") {
		t.Errorf("forge_progress step lacks MUST language: %q", progressStep)
	}

	// Verify hivemind_store step contains "MUST".
	var storeStep string
	for _, line := range lines {
		if strings.Contains(line, "hivemind_store") {
			storeStep = line
			break
		}
	}
	if storeStep == "" {
		t.Fatal("worker.md: no line containing 'hivemind_store' found")
	}
	if !strings.Contains(storeStep, "MUST") {
		t.Errorf("hivemind_store step lacks MUST language: %q", storeStep)
	}

	// Verify reservation failure recovery instruction is co-located with the
	// comms_reserve step. Search within a window of lines (current + next 3)
	// so the assertion tolerates sub-bullet reformatting.
	reserveWindow := stepWindow(reserveIdx)
	if !strings.Contains(reserveWindow, "comms_send") {
		t.Errorf("comms_reserve step window lacks comms_send recovery instruction:\n%s", reserveWindow)
	}
	if !strings.Contains(reserveWindow, "STOP") {
		t.Errorf("comms_reserve step window lacks STOP instruction for reservation failure:\n%s", reserveWindow)
	}
}


func TestForgeMD_StructuralHardening(t *testing.T) {
	// Read forge.md from embedded content.
	data, err := content.ReadFile("content/commands/forge.md")
	if err != nil {
		t.Fatalf("read forge.md: %v", err)
	}
	text := string(data)
	lines := strings.Split(text, "\n")

	// Helper: find the line index of a heading (e.g., "## Critical Invariants").
	findHeading := func(heading string) int {
		for i, line := range lines {
			if strings.TrimSpace(line) == heading {
				return i
			}
		}
		return -1
	}

	// Helper: extract the section between a heading and the next same-level heading.
	sectionContent := func(heading string) string {
		start := findHeading(heading)
		if start < 0 {
			return ""
		}
		level := 0
		for _, ch := range heading {
			if ch == '#' {
				level++
			} else {
				break
			}
		}
		var sb strings.Builder
		for i := start + 1; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, strings.Repeat("#", level)+" ") && !strings.HasPrefix(trimmed, strings.Repeat("#", level+1)) {
				break
			}
			sb.WriteString(lines[i])
			sb.WriteString("\n")
		}
		return sb.String()
	}

	// Scenario 1: Critical Invariants section appears before Workflow section.
	t.Run("InvariantsBeforeWorkflow", func(t *testing.T) {
		invIdx := findHeading("## Critical Invariants")
		wfIdx := findHeading("## Workflow")
		if invIdx < 0 {
			t.Fatal("Critical Invariants section not found")
		}
		if wfIdx < 0 {
			t.Fatal("Workflow section not found")
		}
		if invIdx >= wfIdx {
			t.Errorf("Critical Invariants (line %d) must appear before Workflow (line %d)", invIdx, wfIdx)
		}
	})

	// Scenario 2: Review-before-complete invariant is present in Critical Invariants.
	t.Run("ReviewBeforeCompleteInvariant", func(t *testing.T) {
		section := sectionContent("## Critical Invariants")
		if section == "" {
			t.Fatal("Critical Invariants section not found")
		}
		lower := strings.ToLower(section)
		if !strings.Contains(lower, "review") || !strings.Contains(lower, "before") {
			t.Error("Critical Invariants must contain review-before-complete constraint")
		}
		if !strings.Contains(section, "MUST") {
			t.Error("Critical Invariants must use RFC 2119 MUST language for review constraint")
		}
	})

	// Scenario 3: Review gate mandatory constraint is present in Critical Invariants.
	t.Run("ReviewGateMandatory", func(t *testing.T) {
		section := sectionContent("## Critical Invariants")
		if section == "" {
			t.Fatal("Critical Invariants section not found")
		}
		// The review gate must be stated as a positive constraint (not naming bypass parameters).
		if !strings.Contains(section, "MUST NOT") {
			t.Error("Critical Invariants must use MUST NOT for review gate constraint")
		}
		if !strings.Contains(section, "NEVER") {
			t.Error("Critical Invariants must use NEVER for review gate constraint")
		}
		lower := strings.ToLower(section)
		if !strings.Contains(lower, "review gate") {
			t.Error("Critical Invariants must reference 'review gate' as mandatory")
		}
	})

	// Scenario 4: Review-before-complete constraint has redundant placement
	// (present in ALL THREE sections: Critical Invariants, Workflow, and Rules).
	t.Run("RedundantReviewConstraint", func(t *testing.T) {
		invariants := sectionContent("## Critical Invariants")
		workflow := sectionContent("## Workflow")
		rules := sectionContent("## Rules")

		inInvariants := strings.Contains(strings.ToLower(invariants), "review") &&
			strings.Contains(strings.ToLower(invariants), "before")
		inWorkflow := strings.Contains(strings.ToLower(workflow), "review") &&
			strings.Contains(strings.ToLower(workflow), "before")
		inRules := strings.Contains(strings.ToLower(rules), "review") &&
			strings.Contains(strings.ToLower(rules), "before") &&
			strings.Contains(strings.ToLower(rules), "complete")

		if !inInvariants {
			t.Error("review-before-complete not found in Critical Invariants")
		}
		if !inWorkflow {
			t.Error("review-before-complete not found in Workflow section")
		}
		if !inRules {
			t.Error("review-before-complete not found in Rules section")
		}
	})

	// Scenario 5: Step 7 text includes explicit ordering constraint.
	t.Run("Step7OrderingConstraint", func(t *testing.T) {
		workflow := sectionContent("## Workflow")
		if workflow == "" {
			t.Fatal("Workflow section not found")
		}
		// Find step 7 line — require MUST AND an ordering signal.
		var foundStep7 bool
		for _, line := range strings.Split(workflow, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "7.") {
				foundStep7 = true
				lower := strings.ToLower(line)
				hasMust := strings.Contains(line, "MUST")
				hasOrdering := strings.Contains(lower, "first") || strings.Contains(lower, "before step 8")
				if !hasMust {
					t.Error("Step 7 must use RFC 2119 MUST language")
				}
				if !hasOrdering {
					t.Error("Step 7 must contain ordering signal (FIRST or 'before step 8')")
				}
				break
			}
		}
		if !foundStep7 {
			t.Error("Step 7 not found in Workflow section")
		}
	})

	// Scenario 6: Review rule is first item in Rules section.
	t.Run("ReviewRuleFirstInRules", func(t *testing.T) {
		rules := sectionContent("## Rules")
		if rules == "" {
			t.Fatal("Rules section not found")
		}
		// Find first bullet in Rules section.
		for _, line := range strings.Split(rules, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") {
				lower := strings.ToLower(trimmed)
				if !strings.Contains(lower, "review") {
					t.Errorf("first Rules bullet must be about review, got: %s", trimmed)
				}
				break
			}
		}
	})

	// Scenario 7: No standalone Strategy Selection, Error Recovery, or Completion sections.
	t.Run("NoStandaloneSections", func(t *testing.T) {
		prohibited := []string{
			"## Strategy Selection",
			"## Error Recovery",
			"## Completion",
		}
		for _, heading := range prohibited {
			if findHeading(heading) >= 0 {
				t.Errorf("found prohibited standalone section: %s", heading)
			}
		}
	})

	// Scenario 8: Strategy selection content is inlined within step 3 (Decompose).
	t.Run("StrategyInlinedInStep3", func(t *testing.T) {
		workflow := sectionContent("## Workflow")
		if workflow == "" {
			t.Fatal("Workflow section not found")
		}
		// Find step 3 and its sub-items (lines between "3." and the next step "4.").
		wfLines := strings.Split(workflow, "\n")
		var step3Content strings.Builder
		inStep3 := false
		for _, line := range wfLines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "3.") {
				inStep3 = true
			} else if inStep3 && len(trimmed) > 0 && trimmed[0] >= '1' && trimmed[0] <= '9' && len(trimmed) > 1 && trimmed[1] == '.' {
				break
			}
			if inStep3 {
				step3Content.WriteString(line)
				step3Content.WriteString("\n")
			}
		}
		s3 := step3Content.String()
		if !strings.Contains(s3, "forge_get_strategy_insights") {
			t.Error("Step 3 must contain forge_get_strategy_insights (strategy selection inlined)")
		}
		if !strings.Contains(s3, "forge_decompose") {
			t.Error("Step 3 must contain forge_decompose")
		}
	})

	// Scenario 9: Error recovery content is inlined within step 6 (Monitor).
	t.Run("ErrorRecoveryInlinedInStep6", func(t *testing.T) {
		workflow := sectionContent("## Workflow")
		if workflow == "" {
			t.Fatal("Workflow section not found")
		}
		wfLines := strings.Split(workflow, "\n")
		var step6Content strings.Builder
		inStep6 := false
		for _, line := range wfLines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "6.") {
				inStep6 = true
			} else if inStep6 && len(trimmed) > 0 && trimmed[0] >= '1' && trimmed[0] <= '9' && len(trimmed) > 1 && trimmed[1] == '.' {
				break
			}
			if inStep6 {
				step6Content.WriteString(line)
				step6Content.WriteString("\n")
			}
		}
		s6 := step6Content.String()
		lower := strings.ToLower(s6)
		if !strings.Contains(lower, "blocked") {
			t.Error("Step 6 must contain blocked-worker recovery guidance")
		}
		if !strings.Contains(lower, "unblock") && !strings.Contains(lower, "reassign") {
			t.Error("Step 6 must contain recovery action (unblock or reassign)")
		}
	})

	// Scenario 10: Completion sub-steps are inlined within step 8.
	t.Run("CompletionInlinedInStep8", func(t *testing.T) {
		workflow := sectionContent("## Workflow")
		if workflow == "" {
			t.Fatal("Workflow section not found")
		}
		wfLines := strings.Split(workflow, "\n")
		var step8Content strings.Builder
		inStep8 := false
		for _, line := range wfLines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "8.") {
				inStep8 = true
			} else if inStep8 && len(trimmed) > 0 && trimmed[0] >= '1' && trimmed[0] <= '9' && len(trimmed) > 1 && trimmed[1] == '.' {
				break
			}
			if inStep8 {
				step8Content.WriteString(line)
				step8Content.WriteString("\n")
			}
		}
		s8 := step8Content.String()
		required := []string{
			"forge_complete",
			"forge_record_outcome",
			"hivemind_store",
			"org_sync",
		}
		for _, tool := range required {
			if !strings.Contains(s8, tool) {
				t.Errorf("Step 8 must contain %s (completion sub-step)", tool)
			}
		}
	})

	// Scenario 11: All MCP tool references from the forge workflow are present.
	t.Run("AllToolReferencesPresent", func(t *testing.T) {
		allTools := []string{
			"comms_init",
			"hivemind_find",
			"forge_decompose",
			"forge_get_strategy_insights",
			"org_create_epic",
			"forge_spawn_subtask",
			"comms_inbox",
			"forge_status",
			"org_cells",
			"comms_read_message",
			"comms_ack",
			"forge_review",
			"forge_complete",
			"forge_record_outcome",
			"hivemind_store",
			"org_sync",
			"comms_reserve",
		}
		for _, tool := range allTools {
			if !strings.Contains(text, tool) {
				t.Errorf("forge.md missing MCP tool reference: %s", tool)
			}
		}
	})
}

func TestForgeCoordinationSkill_StructuralHardening(t *testing.T) {
	data, err := content.ReadFile("content/skills/forge-coordination/SKILL.md")
	if err != nil {
		t.Fatalf("read embedded forge-coordination SKILL.md: %v", err)
	}
	text := string(data)

	// (1) Role-scoped section headers exist.
	roleSections := []string{
		"## Coordinator-Only Operations",
		"## Worker-Only Operations",
	}
	for _, sec := range roleSections {
		if !strings.Contains(text, sec) {
			t.Errorf("forge-coordination SKILL.md: missing role-scoped section %q", sec)
		}
	}

	// (2) MUST/MUST NOT rules appear before protocol steps in each role section.
	coordinatorOpsIdx := strings.Index(text, "## Coordinator-Only Operations")
	coordinatorProtocolIdx := strings.Index(text, "### Coordinator Protocol")
	workerOpsIdx := strings.Index(text, "## Worker-Only Operations")
	workerProtocolIdx := strings.Index(text, "### Worker Protocol")

	if coordinatorOpsIdx < 0 || coordinatorProtocolIdx < 0 {
		t.Fatal("forge-coordination SKILL.md: missing Coordinator sections")
	}
	if workerOpsIdx < 0 || workerProtocolIdx < 0 {
		t.Fatal("forge-coordination SKILL.md: missing Worker sections")
	}

	// Coordinator MUST NOT rules must be between the section header and the protocol.
	coordSection := text[coordinatorOpsIdx:coordinatorProtocolIdx]
	coordMustRules := []string{
		"MUST NOT reserve files",
		"MUST NOT edit files directly",
	}
	for _, rule := range coordMustRules {
		if !strings.Contains(coordSection, rule) {
			t.Errorf("forge-coordination SKILL.md: coordinator MUST rule %q not found before Coordinator Protocol", rule)
		}
	}

	// Worker MUST rules must be between the section header and the protocol.
	workerSection := text[workerOpsIdx:workerProtocolIdx]
	workerMustRules := []string{
		"MUST reserve files before editing",
		"MUST NOT call `comms_release_all()`",
	}
	for _, rule := range workerMustRules {
		if !strings.Contains(workerSection, rule) {
			t.Errorf("forge-coordination SKILL.md: worker MUST rule %q not found before Worker Protocol", rule)
		}
	}

	// (3) exclusive=true appears in the comms_reserve call within Worker Protocol.
	workerProtocolSection := text[workerProtocolIdx:]
	if !strings.Contains(workerProtocolSection, "exclusive=true") {
		t.Error("forge-coordination SKILL.md: Worker Protocol missing exclusive=true in comms_reserve call")
	}

	// (4) Removed sections from original document are absent.
	removedSections := []string{
		"## File Reservation Rules",
		"## Conflict Resolution",
	}
	for _, sec := range removedSections {
		if strings.Contains(text, sec) {
			t.Errorf("forge-coordination SKILL.md: removed section %q should not be present", sec)
		}
	}

	// (5) Both copies are byte-identical (embedded agentkit vs .opencode at repo root).
	// Find repo root by walking up from the working directory to locate go.mod.
	repoRoot := findRepoRoot(t)
	opencodePath := filepath.Join(repoRoot, ".opencode", "skills", "forge-coordination", "SKILL.md")
	opencodeCopy, err := os.ReadFile(opencodePath)
	if err != nil {
		t.Fatalf("read .opencode forge-coordination SKILL.md: %v", err)
	}
	if string(data) != string(opencodeCopy) {
		t.Error("forge-coordination SKILL.md: embedded copy and .opencode copy are not byte-identical")
	}
}

func TestAlwaysOnGuidance_StructuralHardening(t *testing.T) {
	data, err := content.ReadFile("content/skills/always-on-guidance/SKILL.md")
	if err != nil {
		t.Fatalf("read embedded always-on-guidance/SKILL.md: %v", err)
	}
	text := string(data)

	// (1) Critical Safety section exists and appears before Tool Usage Discipline.
	safetyIdx := strings.Index(text, "## Critical Safety")
	if safetyIdx < 0 {
		t.Error("always-on-guidance: missing '## Critical Safety' section")
	}
	toolUsageIdx := strings.Index(text, "## Tool Usage Discipline")
	if toolUsageIdx < 0 {
		t.Error("always-on-guidance: missing '## Tool Usage Discipline' section")
	}
	if safetyIdx >= 0 && toolUsageIdx >= 0 && safetyIdx >= toolUsageIdx {
		t.Error("always-on-guidance: '## Critical Safety' must appear before '## Tool Usage Discipline'")
	}

	// (2) Force push rule uses RFC 2119 uppercase keyword (DR-002).
	if !strings.Contains(text, "NEVER force push") {
		t.Error("always-on-guidance: force push rule must use RFC 2119 keyword 'NEVER'")
	}

	// (3) hivemind_find is the first item in Tool Usage Discipline section.
	if toolUsageIdx >= 0 {
		afterToolUsage := text[toolUsageIdx:]
		firstDashIdx := strings.Index(afterToolUsage, "\n- ")
		if firstDashIdx < 0 {
			t.Error("always-on-guidance: no list items in Tool Usage Discipline")
		} else {
			// Extract the first list item line.
			firstItemStart := firstDashIdx + 3 // skip "\n- "
			firstItemEnd := strings.Index(afterToolUsage[firstItemStart:], "\n")
			if firstItemEnd < 0 {
				firstItemEnd = len(afterToolUsage) - firstItemStart
			}
			firstItem := afterToolUsage[firstItemStart : firstItemStart+firstItemEnd]
			if !strings.Contains(firstItem, "hivemind_find") {
				t.Errorf("always-on-guidance: first Tool Usage item should mention hivemind_find, got %q", firstItem)
			}
		}
	}

	// (4) Code Quality split into sub-headers.
	for _, sub := range []string{"### Structure", "### Clarity"} {
		if !strings.Contains(text, sub) {
			t.Errorf("always-on-guidance: missing Code Quality sub-header %q", sub)
		}
	}

	// (5) Testing split into sub-headers.
	for _, sub := range []string{"### Test Infrastructure", "### Test Practice"} {
		if !strings.Contains(text, sub) {
			t.Errorf("always-on-guidance: missing Testing sub-header %q", sub)
		}
	}

	// (6) Error Handling split into sub-headers.
	for _, sub := range []string{"### Error Propagation", "### Error Coverage"} {
		if !strings.Contains(text, sub) {
			t.Errorf("always-on-guidance: missing Error Handling sub-header %q", sub)
		}
	}
}

func TestForgeGlobal_StructuralHardening(t *testing.T) {
	data, err := content.ReadFile("content/skills/forge-global/SKILL.md")
	if err != nil {
		t.Fatalf("read embedded forge-global/SKILL.md: %v", err)
	}
	text := string(data)

	// (1) Decision table format with Signal/Forge/Skip columns.
	if !strings.Contains(text, "| Signal | Forge | Skip |") {
		t.Error("forge-global: missing decision table header (Signal/Forge/Skip)")
	}

	// (2) All 6 original criteria present in decision table.
	criteria := []string{
		"File count",
		"Task structure",
		"Work type",
		"3+ files",
		"single-file change",
		"parallelize",
	}
	for _, c := range criteria {
		if !strings.Contains(text, c) {
			t.Errorf("forge-global: decision table missing criterion %q", c)
		}
	}

	// (3) Temporal ordering markers in File Reservation Protocol.
	for _, marker := range []string{"FIRST,", "THEN,", "FINALLY,"} {
		if !strings.Contains(text, marker) {
			t.Errorf("forge-global: missing temporal marker %q in File Reservation Protocol", marker)
		}
	}

	// (4) TTL inlined in step 1 with specific value.
	if !strings.Contains(text, "ttl_seconds=300") {
		t.Error("forge-global: ttl_seconds=300 must be inlined in reservation step")
	}

	// (5) TTL explanation parenthetical present (5-minute auto-release).
	if !strings.Contains(text, "(5-minute auto-release)") {
		t.Error("forge-global: missing '(5-minute auto-release)' explanation for TTL")
	}

	// (6) No standalone TTL bullet (old format removed).
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Old format was a standalone step like "3. Set `ttl_seconds` to auto-release..."
		if strings.HasPrefix(trimmed, "3.") && strings.Contains(trimmed, "ttl_seconds") && strings.Contains(trimmed, "auto-release") {
			t.Error("forge-global: standalone TTL step should be removed (inlined into step 1)")
		}
	}
}

func TestSkillFiles_DriftDetection(t *testing.T) {
	// TC-007: Verify embedded skill files match the .opencode/ scaffolded copies.
	// This detects drift where one copy is updated but the other is not.
	//
	// Find the repo root by walking up from the test working directory
	// until we find go.mod.
	repoRoot := findRepoRoot(t)

	skills := []string{
		"always-on-guidance",
		"forge-global",
	}

	for _, skill := range skills {
		embeddedPath := filepath.Join("content", "skills", skill, "SKILL.md")
		embedded, err := content.ReadFile(embeddedPath)
		if err != nil {
			t.Fatalf("read embedded %s: %v", embeddedPath, err)
		}

		scaffoldedPath := filepath.Join(repoRoot, ".opencode", "skills", skill, "SKILL.md")
		scaffolded, err := os.ReadFile(scaffoldedPath)
		if err != nil {
			t.Fatalf("read scaffolded %s: %v", scaffoldedPath, err)
		}

		if string(embedded) != string(scaffolded) {
			t.Errorf("drift detected: embedded %s differs from scaffolded .opencode/skills/%s/SKILL.md", embeddedPath, skill)
		}
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
