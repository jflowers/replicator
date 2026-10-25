package forge

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/unbound-force/replicator/internal/db"
)

// Sentinel errors for the spawn delegation contract. The codes lead
// the wrapped error messages so failures stay greppable in tool
// output and logs; match with errors.Is.
var (
	// ErrSkillReferenceRequired fails a deterministic-classified
	// spawn that supplies neither a skill/script reference nor a
	// recorded skill_absent_reason.
	ErrSkillReferenceRequired = errors.New("SKILL_REFERENCE_REQUIRED")

	// ErrSkillNotFound fails a spawn whose skill/script reference
	// cannot be resolved to an existing file under the project.
	ErrSkillNotFound = errors.New("SKILL_NOT_FOUND")
)

// Task classification values for the spawn delegation contract.
const (
	ClassificationDeterministic = "deterministic"
	ClassificationJudgment      = "judgment"
)

// Length bounds for delegation contract fields.
const (
	MaxSkillRefLen    = 256
	MaxSkillReasonLen = 512
)

// SpawnParams carries the inputs for spawning a subtask,
// including the delegation contract parameters
// (task_classification, skill, skill_absent_reason) and the
// optional ProjectPath used to resolve skill references
// (empty means the working directory).
type SpawnParams struct {
	BeadID             string   `json:"bead_id"`
	EpicID             string   `json:"epic_id"`
	Title              string   `json:"subtask_title"`
	Files              []string `json:"files"`
	Description        string   `json:"subtask_description"`
	SharedContext      string   `json:"shared_context"`
	TaskClassification string   `json:"task_classification"`
	Skill              string   `json:"skill"`
	SkillAbsentReason  string   `json:"skill_absent_reason"`
	ProjectPath        string   `json:"project_path"`
}

// SubtaskPrompt generates a prompt for a spawned subtask agent.
// This is a prompt generator -- it does NOT call an LLM. When
// skillRef is non-empty the fixed skill directive is rendered so
// the worker sees the same instructions the spawn envelope
// recorded; otherwise the output is unchanged from the
// no-reference shape.
func SubtaskPrompt(agentName, beadID, epicID, title string, files []string, sharedContext, skillRef string) string {
	var sb strings.Builder

	_, _ = fmt.Fprintf(&sb, "# Subtask: %s\n\n", title)
	_, _ = fmt.Fprintf(&sb, "**Agent:** %s\n", agentName)
	_, _ = fmt.Fprintf(&sb, "**Cell ID:** %s\n", beadID)
	_, _ = fmt.Fprintf(&sb, "**Epic ID:** %s\n\n", epicID)

	if len(files) > 0 {
		sb.WriteString("## Files to Modify\n\n")
		for _, f := range files {
			_, _ = fmt.Fprintf(&sb, "- `%s`\n", f)
		}
		sb.WriteString("\n")
	}

	if sharedContext != "" {
		sb.WriteString("## Shared Context\n\n")
		sb.WriteString(sharedContext)
		sb.WriteString("\n\n")
	}

	if skillRef != "" {
		sb.WriteString("## Required Skill\n\n")
		sb.WriteString(SkillDirective(skillRef))
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Instructions\n\n")
	sb.WriteString("1. Reserve files before editing using `comms_reserve`\n")
	sb.WriteString("2. Implement the changes described above\n")
	sb.WriteString("3. Write tests for new code\n")
	sb.WriteString("4. Report progress with `forge_progress`\n")
	sb.WriteString("5. Complete with `forge_complete` when done\n")

	return sb.String()
}

// SpawnSubtask prepares metadata for spawning a subtask agent.
// Returns the metadata needed to launch the agent.
//
// It enforces the spawn delegation contract: task_classification
// is required and enum-checked, skill and skill_absent_reason are
// mutually exclusive and valid only for deterministic
// classification, and a deterministic-classified subtask must
// supply one or the other (ErrSkillReferenceRequired otherwise).
// A non-empty skill is resolved against the project root
// (ProjectPath, defaulting to the working directory): an
// unresolvable reference fails with ErrSkillNotFound, while other
// filesystem errors (for example permission denied) are returned
// unwrapped so operational failures stay distinguishable from
// absence. Empty-string Skill and SkillAbsentReason values are
// treated as absent at this layer; the MCP tool boundary rejects
// explicit empty strings.
func SpawnSubtask(params SpawnParams) (map[string]any, error) {
	if params.BeadID == "" {
		return nil, fmt.Errorf("bead_id is required")
	}
	if params.EpicID == "" {
		return nil, fmt.Errorf("epic_id is required")
	}
	if params.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if err := validateSpawnContract(&params); err != nil {
		return nil, err
	}

	result := map[string]any{
		"status":              "ready_to_spawn",
		"bead_id":             params.BeadID,
		"epic_id":             params.EpicID,
		"title":               params.Title,
		"files":               params.Files,
		"description":         params.Description,
		"shared_context":      params.SharedContext,
		"task_classification": params.TaskClassification,
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
	}

	if params.Skill != "" {
		resolved, err := ResolveSkillRef(params.ProjectPath, params.Skill)
		if err != nil {
			return nil, err
		}
		result["skill"] = resolved
		result["skill_directive"] = SkillDirective(params.Skill)
	}
	if params.SkillAbsentReason != "" {
		result["skill_absent_reason"] = params.SkillAbsentReason
	}

	return result, nil
}

// validateSpawnContract enforces the structural rules of the spawn
// delegation contract: classification is required and enum-checked,
// skill and skill_absent_reason are mutually exclusive, both are
// valid only for deterministic classification, and length bounds
// hold. Existence resolution happens after these checks pass.
func validateSpawnContract(p *SpawnParams) error {
	switch p.TaskClassification {
	case ClassificationDeterministic, ClassificationJudgment:
	default:
		if p.TaskClassification == "" {
			return fmt.Errorf("task_classification is required (accepted values: deterministic, judgment)")
		}
		return fmt.Errorf("task_classification must be deterministic or judgment (got %q)", p.TaskClassification)
	}

	if p.Skill != "" && p.SkillAbsentReason != "" {
		return fmt.Errorf("skill and skill_absent_reason are mutually exclusive; supply one or the other, not both")
	}

	if p.TaskClassification == ClassificationJudgment {
		if p.Skill != "" {
			return fmt.Errorf("skill is valid only for deterministic classification (got %q); omit it for judgment work", p.TaskClassification)
		}
		if p.SkillAbsentReason != "" {
			return fmt.Errorf("skill_absent_reason is valid only for deterministic classification (got %q); omit it for judgment work", p.TaskClassification)
		}
	}

	if p.Skill != "" && len(p.Skill) > MaxSkillRefLen {
		return fmt.Errorf("skill must be at most %d characters (got %d)", MaxSkillRefLen, len(p.Skill))
	}
	if p.SkillAbsentReason != "" && len(p.SkillAbsentReason) > MaxSkillReasonLen {
		return fmt.Errorf("skill_absent_reason must be at most %d characters (got %d)", MaxSkillReasonLen, len(p.SkillAbsentReason))
	}

	if p.TaskClassification == ClassificationDeterministic && p.Skill == "" && p.SkillAbsentReason == "" {
		return fmt.Errorf("%w: deterministic-classified subtask requires a verified skill/script reference or a recorded skill_absent_reason; supply skill or skill_absent_reason", ErrSkillReferenceRequired)
	}

	return nil
}

// skillRoots lists, in order, the repo-relative directories under
// which a bare skill name resolves to a SKILL.md file. The first
// root containing the skill wins.
var skillRoots = []string{
	filepath.Join(".opencode", "skills"),
	filepath.Join(".opencode", "skill"),
}

// ResolveSkillRef resolves a skill/script reference against the
// project root. A value without a path separator resolves as a
// skill name against the project skill directories; a value
// containing a separator resolves as a repo-relative file path
// (absolute paths and traversal outside the root are rejected).
// The returned path is repo-relative.
func ResolveSkillRef(projectPath, ref string) (string, error) {
	if projectPath == "" {
		projectPath = "."
	}
	if ref == "." || ref == ".." {
		return "", fmt.Errorf("%w: path traversal is not allowed (reference %q must stay inside the project root)", ErrSkillNotFound, ref)
	}
	if strings.ContainsRune(ref, '/') {
		return resolveRepoRelative(projectPath, ref)
	}
	var locations []string
	for _, root := range skillRoots {
		rel := filepath.Join(root, ref, "SKILL.md")
		exists, err := fileExists(filepath.Join(projectPath, rel))
		if err != nil {
			return "", fmt.Errorf("resolving skill %q under %s: %w", ref, projectPath, err)
		}
		if exists {
			return rel, nil
		}
		locations = append(locations, rel)
	}
	return "", fmt.Errorf("%w: skill %q not found (searched: %s under %s)", ErrSkillNotFound, ref, strings.Join(locations, ", "), projectPath)
}

// resolveRepoRelative resolves a slash-containing reference as a
// repo-relative path, rejecting absolute paths and any path that
// escapes the project root.
func resolveRepoRelative(projectPath, ref string) (string, error) {
	if filepath.IsAbs(ref) {
		return "", fmt.Errorf("%w: absolute paths are not allowed (reference %q must be repo-relative)", ErrSkillNotFound, ref)
	}
	cleaned := filepath.Clean(ref)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: path traversal is not allowed (reference %q must stay inside the project root)", ErrSkillNotFound, ref)
	}
	exists, err := fileExists(filepath.Join(projectPath, cleaned))
	if err != nil {
		return "", fmt.Errorf("resolving %q under %s: %w", cleaned, projectPath, err)
	}
	if !exists {
		return "", fmt.Errorf("%w: %s does not exist under %s", ErrSkillNotFound, cleaned, projectPath)
	}
	return cleaned, nil
}

// fileExists reports whether path exists and is not a directory.
// A stat error other than not-exist is returned so callers can
// distinguish operational failures (for example permission
// denied) from absence.
func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

// SkillDirective renders the fixed worker directive for a skill
// or script reference. The wording is a design constant (mirrors
// the parent delegation contract in unbound-force) and is pinned
// by golden-string tests; do not vary it per call site.
func SkillDirective(ref string) string {
	return fmt.Sprintf(
		"Required skill: %s\n\n"+
			"This task is classified as deterministic work. Load the skill (invoke the `skill` tool with name %q) "+
			"or run the script at %s, and execute its full procedure. Do not substitute an improvised procedure.",
		ref, ref, ref,
	)
}

// CompleteSubtask records the completion of a subtask and updates the cell status.
func CompleteSubtask(store *db.Store, beadID, taskResult string, filesTouched []string) (map[string]any, error) {
	if beadID == "" {
		return nil, fmt.Errorf("bead_id is required")
	}

	payload := map[string]any{
		"bead_id":       beadID,
		"task_result":   taskResult,
		"files_touched": filesTouched,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	_, err = store.DB.Exec(
		"INSERT INTO events (type, payload) VALUES (?, ?)",
		"subtask_complete", string(payloadJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("record subtask_complete event: %w", err)
	}

	// Update cell status to closed.
	// The event is already recorded, so a cell update failure is non-fatal
	// but noted in the result for observability.
	now := time.Now().UTC().Format(time.RFC3339)
	_, cellErr := store.DB.Exec(
		"UPDATE beads SET status = 'closed', closed_at = ?, close_reason = 'completed', updated_at = ? WHERE id = ?",
		now, now, beadID,
	)

	result := map[string]any{
		"status":        "completed",
		"bead_id":       beadID,
		"files_touched": filesTouched,
	}
	if cellErr != nil {
		result["cell_update_error"] = cellErr.Error()
	}
	return result, nil
}
