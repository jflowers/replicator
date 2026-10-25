package forge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/replicator/internal/forge"
	"github.com/unbound-force/replicator/internal/tools/registry"
)

// schemaProperty returns the parsed properties map of a tool's input schema.
func schemaProperty(t *testing.T, tool *registry.Tool, name string) map[string]any {
	t.Helper()
	var schema struct {
		Required   []string       `json:"required"`
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("unmarshal schema for %s: %v", tool.Name, err)
	}
	for _, req := range schema.Required {
		if req == name {
			prop, ok := schema.Properties[name].(map[string]any)
			if !ok {
				t.Fatalf("%s: property %q marked required but not defined", tool.Name, name)
			}
			return prop
		}
	}
	t.Fatalf("%s: %q not in required list %v", tool.Name, name, schema.Required)
	return nil
}

func TestForgeSpawnSubtask_SchemaDocumentsContract(t *testing.T) {
	tool := forgeSpawnSubtask()

	prop := schemaProperty(t, tool, "task_classification")
	enum, ok := prop["enum"].([]any)
	if !ok || len(enum) != 2 || enum[0] != "deterministic" || enum[1] != "judgment" {
		t.Errorf("task_classification enum = %v, want [deterministic judgment]", prop["enum"])
	}

	var schema struct {
		Required   []string                  `json:"required"`
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	required := make(map[string]bool)
	for _, name := range schema.Required {
		required[name] = true
	}
	for _, name := range []string{"bead_id", "epic_id", "subtask_title", "files", "task_classification"} {
		if !required[name] {
			t.Errorf("schema should require %q (required: %v)", name, schema.Required)
		}
	}
	for _, name := range []string{"skill", "skill_absent_reason", "project_path"} {
		if _, ok := schema.Properties[name]; !ok {
			t.Errorf("schema should document %q", name)
		}
	}
	for _, name := range []string{"bead_id", "epic_id", "subtask_title", "files"} {
		if _, ok := schema.Properties[name]; !ok {
			t.Errorf("schema lost pre-existing property %q", name)
		}
	}
	if skill := schema.Properties["skill"]; skill["minLength"] != float64(1) || skill["maxLength"] != float64(forge.MaxSkillRefLen) {
		t.Errorf("skill bounds = %v/%v, want 1/%d", skill["minLength"], skill["maxLength"], forge.MaxSkillRefLen)
	}
	if reason := schema.Properties["skill_absent_reason"]; reason["minLength"] != float64(1) || reason["maxLength"] != float64(forge.MaxSkillReasonLen) {
		t.Errorf("skill_absent_reason bounds = %v/%v, want 1/%d", reason["minLength"], reason["maxLength"], forge.MaxSkillReasonLen)
	}
	if def := schema.Properties["project_path"]["default"]; def != "." {
		t.Errorf("project_path default = %v, want \".\"", def)
	}
}

func TestForgeSpawnSubtask_MissingClassificationRejected(t *testing.T) {
	tool := forgeSpawnSubtask()
	args := json.RawMessage(`{
		"bead_id": "cell-1", "epic_id": "epic-1", "subtask_title": "t",
		"files": []
	}`)
	_, err := tool.Execute(args)
	if err == nil {
		t.Fatal("omitted task_classification should be rejected")
	}
	msg := err.Error()
	if !strings.Contains(msg, "task_classification") {
		t.Errorf("error should name task_classification, got: %s", msg)
	}
	if !strings.Contains(msg, "accepted values: deterministic, judgment") {
		t.Errorf("error should list accepted values, got: %s", msg)
	}
}

func TestForgeSpawnSubtask_EmptyStringSkillRejected(t *testing.T) {
	tool := forgeSpawnSubtask()
	args := json.RawMessage(`{
		"bead_id": "cell-1", "epic_id": "epic-1", "subtask_title": "t",
		"files": [], "task_classification": "deterministic", "skill": ""
	}`)
	_, err := tool.Execute(args)
	if err == nil {
		t.Fatal("explicit empty skill should be rejected")
	}
	if msg := err.Error(); !strings.Contains(msg, "omit") {
		t.Errorf("error should tell caller to omit the field, got: %s", msg)
	}
}

func TestForgeSpawnSubtask_EmptyStringReasonRejected(t *testing.T) {
	tool := forgeSpawnSubtask()
	args := json.RawMessage(`{
		"bead_id": "cell-1", "epic_id": "epic-1", "subtask_title": "t",
		"files": [], "task_classification": "deterministic", "skill_absent_reason": ""
	}`)
	_, err := tool.Execute(args)
	if err == nil {
		t.Fatal("explicit empty skill_absent_reason should be rejected")
	}
	if msg := err.Error(); !strings.Contains(msg, "omit") {
		t.Errorf("error should tell caller to omit the field, got: %s", msg)
	}
}

func TestForgeSpawnSubtask_ExecuteReturnsEnvelope(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, ".opencode", "skills", "deploy")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# deploy"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	tool := forgeSpawnSubtask()
	args := json.RawMessage(`{
		"bead_id": "cell-1", "epic_id": "epic-1", "subtask_title": "Run deploy",
		"files": ["deploy.go"], "task_classification": "deterministic",
		"skill": "deploy", "project_path": "` + root + `"
	}`)
	out, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if envelope["task_classification"] != "deterministic" {
		t.Errorf("task_classification = %v", envelope["task_classification"])
	}
	if envelope["skill"] != filepath.Join(".opencode", "skills", "deploy", "SKILL.md") {
		t.Errorf("skill = %v", envelope["skill"])
	}
	if _, ok := envelope["skill_directive"]; !ok {
		t.Error("envelope should carry skill_directive")
	}
}

func TestForgeSpawnSubtask_UnknownSkillFailsClosed(t *testing.T) {
	// An empty temp root guarantees the skill is absent by
	// construction rather than by machine-specific assumption.
	root := t.TempDir()
	tool := forgeSpawnSubtask()
	args := json.RawMessage(`{
		"bead_id": "cell-1", "epic_id": "epic-1", "subtask_title": "t",
		"files": [], "task_classification": "deterministic",
		"skill": "does-not-exist", "project_path": "` + root + `"
	}`)
	_, err := tool.Execute(args)
	if err == nil {
		t.Fatal("unknown skill should fail closed")
	}
	if !strings.HasPrefix(err.Error(), "SKILL_NOT_FOUND") {
		t.Errorf("error should lead with SKILL_NOT_FOUND, got: %s", err.Error())
	}
}

func TestForgeSubtaskPrompt_SkillParamRendersDirective(t *testing.T) {
	tool := forgeSubtaskPrompt()
	args := json.RawMessage(`{
		"agent_name": "worker-1", "bead_id": "cell-1", "epic_id": "epic-1",
		"subtask_title": "Run migration", "files": ["migrate.go"],
		"skill": "scripts/migrate.sh"
	}`)
	out, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "Required Skill") {
		t.Error("prompt should contain the Required Skill section")
	}
	if !strings.Contains(out, "Do not substitute an improvised procedure.") {
		t.Error("prompt should embed the fixed directive wording")
	}
}

func TestForgeSubtaskPrompt_NoSkillOmitsDirective(t *testing.T) {
	tool := forgeSubtaskPrompt()
	args := json.RawMessage(`{
		"agent_name": "worker-1", "bead_id": "cell-1", "epic_id": "epic-1",
		"subtask_title": "Implement auth", "files": ["auth.go"]
	}`)
	out, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(out, "Required Skill") {
		t.Error("prompt without skill must not contain the directive section")
	}
}
