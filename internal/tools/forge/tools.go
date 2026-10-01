// Package forge registers MCP tools for forge orchestration operations.
package forge

import (
	"encoding/json"

	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/forge"
	"github.com/unbound-force/replicator/internal/tools/registry"
)

// Register adds all forge orchestration tools to the registry.
func Register(reg *registry.Registry, store *db.Store) {
	// Init & decomposition.
	reg.Register(forgeInit(store))
	reg.Register(forgeSelectStrategy())
	reg.Register(forgePlanPrompt())
	reg.Register(forgeDecompose())
	reg.Register(forgeValidateDecomposition())

	// Subtask spawning.
	reg.Register(forgeSubtaskPrompt())
	reg.Register(forgeSpawnSubtask())
	reg.Register(forgeCompleteSubtask(store))

	// Progress & status.
	reg.Register(forgeProgress(store))
	reg.Register(forgeComplete(store))
	reg.Register(forgeStatus(store))
	reg.Register(forgeRecordOutcome(store))

	// Worktree management.
	reg.Register(forgeWorktreeCreate())
	reg.Register(forgeWorktreeMerge())
	reg.Register(forgeWorktreeCleanup())
	reg.Register(forgeWorktreeList())

	// Review & broadcast.
	reg.Register(forgeReview())
	reg.Register(forgeReviewFeedback(store))
	reg.Register(forgeAdversarialReview())
	reg.Register(forgeEvaluationPrompt())
	reg.Register(forgeBroadcast(store))

	// Insights.
	reg.Register(forgeGetStrategyInsights(store))
	reg.Register(forgeGetFileInsights(store))
	reg.Register(forgeGetPatternInsights(store))
}

// --- Init & Decomposition ---

func forgeInit(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_init",
		Description: "Initialize swarm session and check tool availability.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"project_path": {"type": "string"},
				"isolation":    {"type": "string", "enum": ["worktree", "reservation"]}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectPath string `json:"project_path"`
				Isolation   string `json:"isolation"`
			}
			if len(args) > 0 {
				if err := json.Unmarshal(args, &input); err != nil {
					return "", err
				}
			}
			result, err := forge.Init(store, input.ProjectPath, input.Isolation)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeSelectStrategy() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_select_strategy",
		Description: "Analyze task and recommend decomposition strategy.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["task"],
			"properties": {
				"task":             {"type": "string", "minLength": 1},
				"codebase_context": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				Task            string `json:"task"`
				CodebaseContext string `json:"codebase_context"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			strategy, err := forge.SelectStrategy(input.Task, input.CodebaseContext)
			if err != nil {
				return "", err
			}
			result := map[string]string{"strategy": strategy}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgePlanPrompt() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_plan_prompt",
		Description: "Generate strategy-specific decomposition prompt.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["task"],
			"properties": {
				"task":          {"type": "string", "minLength": 1},
				"strategy":     {"type": "string", "enum": ["file-based", "feature-based", "risk-based", "auto"]},
				"context":      {"type": "string"},
				"max_subtasks": {"type": "integer", "minimum": 2, "maximum": 10}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				Task        string `json:"task"`
				Strategy    string `json:"strategy"`
				Context     string `json:"context"`
				MaxSubtasks int    `json:"max_subtasks"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			prompt := forge.PlanPrompt(input.Task, input.Strategy, input.Context, input.MaxSubtasks)
			return prompt, nil
		},
	}
}

func forgeDecompose() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_decompose",
		Description: "Generate decomposition prompt for breaking task into subtasks.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["task"],
			"properties": {
				"task":          {"type": "string", "minLength": 1},
				"context":      {"type": "string"},
				"max_subtasks": {"type": "integer", "minimum": 2, "maximum": 10}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				Task        string `json:"task"`
				Context     string `json:"context"`
				MaxSubtasks int    `json:"max_subtasks"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			prompt := forge.Decompose(input.Task, input.Context, input.MaxSubtasks)
			return prompt, nil
		},
	}
}

func forgeValidateDecomposition() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_validate_decomposition",
		Description: "Validate a decomposition response against CellTreeSchema.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["response"],
			"properties": {
				"response": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				Response string `json:"response"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.ValidateDecomposition(input.Response)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

// --- Subtask Spawning ---

func forgeSubtaskPrompt() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_subtask_prompt",
		Description: "Generate the prompt for a spawned subtask agent.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["agent_name", "bead_id", "epic_id", "subtask_title", "files"],
			"properties": {
				"agent_name":          {"type": "string"},
				"bead_id":             {"type": "string"},
				"epic_id":             {"type": "string"},
				"subtask_title":       {"type": "string"},
				"files":               {"type": "array", "items": {"type": "string"}},
				"shared_context":      {"type": "string"},
				"subtask_description": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				AgentName    string   `json:"agent_name"`
				BeadID       string   `json:"bead_id"`
				EpicID       string   `json:"epic_id"`
				SubtaskTitle string   `json:"subtask_title"`
				Files        []string `json:"files"`
				SharedCtx    string   `json:"shared_context"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			prompt := forge.SubtaskPrompt(input.AgentName, input.BeadID, input.EpicID, input.SubtaskTitle, input.Files, input.SharedCtx)
			return prompt, nil
		},
	}
}

func forgeSpawnSubtask() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_spawn_subtask",
		Description: "Prepare a subtask for spawning with Task tool.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["bead_id", "epic_id", "subtask_title", "files"],
			"properties": {
				"bead_id":              {"type": "string"},
				"epic_id":              {"type": "string"},
				"subtask_title":        {"type": "string"},
				"files":                {"type": "array", "items": {"type": "string"}},
				"subtask_description":  {"type": "string"},
				"shared_context":       {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				BeadID      string   `json:"bead_id"`
				EpicID      string   `json:"epic_id"`
				Title       string   `json:"subtask_title"`
				Files       []string `json:"files"`
				Description string   `json:"subtask_description"`
				SharedCtx   string   `json:"shared_context"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.SpawnSubtask(input.BeadID, input.EpicID, input.Title, input.Files, input.Description, input.SharedCtx)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeCompleteSubtask(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_complete_subtask",
		Description: "Handle subtask completion after Task agent returns.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["bead_id", "task_result"],
			"properties": {
				"bead_id":       {"type": "string"},
				"task_result":   {"type": "string"},
				"files_touched": {"type": "array", "items": {"type": "string"}}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				BeadID       string   `json:"bead_id"`
				TaskResult   string   `json:"task_result"`
				FilesTouched []string `json:"files_touched"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.CompleteSubtask(store, input.BeadID, input.TaskResult, input.FilesTouched)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

// --- Progress & Status ---

func forgeProgress(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_progress",
		Description: "Report progress on a subtask to coordinator.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_key", "agent_name", "bead_id", "status"],
			"properties": {
				"project_key":      {"type": "string"},
				"agent_name":       {"type": "string"},
				"bead_id":          {"type": "string"},
				"status":           {"type": "string", "enum": ["in_progress", "blocked", "completed", "failed"]},
				"progress_percent": {"type": "number", "minimum": 0, "maximum": 100},
				"message":          {"type": "string"},
				"files_touched":    {"type": "array", "items": {"type": "string"}}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectKey      string   `json:"project_key"`
				AgentName       string   `json:"agent_name"`
				BeadID          string   `json:"bead_id"`
				Status          string   `json:"status"`
				ProgressPercent int      `json:"progress_percent"`
				Message         string   `json:"message"`
				FilesTouched    []string `json:"files_touched"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			err := forge.Progress(store, input.ProjectKey, input.AgentName, input.BeadID, input.Status, input.ProgressPercent, input.Message, input.FilesTouched)
			if err != nil {
				return "", err
			}
			return `{"status": "recorded"}`, nil
		},
	}
}

func forgeComplete(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_complete",
		Description: "Mark subtask complete with Verification Gate.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_key", "agent_name", "bead_id", "summary"],
			"properties": {
				"project_key":       {"type": "string"},
				"agent_name":        {"type": "string"},
				"bead_id":           {"type": "string"},
				"summary":           {"type": "string"},
				"files_touched":     {"type": "array", "items": {"type": "string"}},
				"evaluation":        {"type": "string"},
				"skip_verification": {"type": "boolean"},
				"skip_review":       {"type": "boolean"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectKey       string   `json:"project_key"`
				AgentName        string   `json:"agent_name"`
				BeadID           string   `json:"bead_id"`
				Summary          string   `json:"summary"`
				FilesTouched     []string `json:"files_touched"`
				Evaluation       string   `json:"evaluation"`
				SkipVerification bool     `json:"skip_verification"`
				SkipReview       bool     `json:"skip_review"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.Complete(store, input.ProjectKey, input.AgentName, input.BeadID, input.Summary, input.FilesTouched, input.Evaluation, input.SkipVerification, input.SkipReview)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeStatus(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_status",
		Description: "Get status of a swarm by epic ID.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["epic_id", "project_key"],
			"properties": {
				"epic_id":     {"type": "string"},
				"project_key": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				EpicID     string `json:"epic_id"`
				ProjectKey string `json:"project_key"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.Status(store, input.EpicID, input.ProjectKey)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeRecordOutcome(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_record_outcome",
		Description: "Record subtask outcome for implicit feedback scoring.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["bead_id", "duration_ms", "success"],
			"properties": {
				"bead_id":       {"type": "string"},
				"duration_ms":   {"type": "integer", "minimum": 0},
				"success":       {"type": "boolean"},
				"strategy":      {"type": "string", "enum": ["file-based", "feature-based", "risk-based"]},
				"files_touched": {"type": "array", "items": {"type": "string"}},
				"error_count":   {"type": "integer", "minimum": 0},
				"retry_count":   {"type": "integer", "minimum": 0},
				"criteria":      {"type": "array", "items": {"type": "string"}}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				BeadID       string   `json:"bead_id"`
				DurationMs   int      `json:"duration_ms"`
				Success      bool     `json:"success"`
				Strategy     string   `json:"strategy"`
				FilesTouched []string `json:"files_touched"`
				ErrorCount   int      `json:"error_count"`
				RetryCount   int      `json:"retry_count"`
				Criteria     []string `json:"criteria"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			err := forge.RecordOutcome(store, input.BeadID, input.DurationMs, input.Success, input.Strategy, input.FilesTouched, input.ErrorCount, input.RetryCount, input.Criteria)
			if err != nil {
				return "", err
			}
			return `{"status": "recorded"}`, nil
		},
	}
}

// --- Worktree Management ---

func forgeWorktreeCreate() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_worktree_create",
		Description: "Create a git worktree for isolated task execution.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_path", "task_id", "start_commit"],
			"properties": {
				"project_path": {"type": "string"},
				"task_id":      {"type": "string"},
				"start_commit": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectPath string `json:"project_path"`
				TaskID      string `json:"task_id"`
				StartCommit string `json:"start_commit"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.WorktreeCreate(input.ProjectPath, input.TaskID, input.StartCommit)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeWorktreeMerge() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_worktree_merge",
		Description: "Cherry-pick commits from worktree back to main branch.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_path", "task_id"],
			"properties": {
				"project_path": {"type": "string"},
				"task_id":      {"type": "string"},
				"start_commit": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectPath string `json:"project_path"`
				TaskID      string `json:"task_id"`
				StartCommit string `json:"start_commit"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.WorktreeMerge(input.ProjectPath, input.TaskID, input.StartCommit)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeWorktreeCleanup() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_worktree_cleanup",
		Description: "Remove a worktree after completion or abort. Idempotent.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_path"],
			"properties": {
				"project_path": {"type": "string"},
				"task_id":      {"type": "string"},
				"cleanup_all":  {"type": "boolean"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectPath string `json:"project_path"`
				TaskID      string `json:"task_id"`
				CleanupAll  bool   `json:"cleanup_all"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.WorktreeCleanup(input.ProjectPath, input.TaskID, input.CleanupAll)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeWorktreeList() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_worktree_list",
		Description: "List all active worktrees for a project.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_path"],
			"properties": {
				"project_path": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectPath string `json:"project_path"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.WorktreeList(input.ProjectPath)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

// --- Review & Broadcast ---

func forgeReview() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_review",
		Description: "Generate a review prompt for a completed subtask.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_key", "epic_id", "task_id"],
			"properties": {
				"project_key":   {"type": "string"},
				"epic_id":       {"type": "string"},
				"task_id":       {"type": "string"},
				"files_touched": {"type": "array", "items": {"type": "string"}}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectKey   string   `json:"project_key"`
				EpicID       string   `json:"epic_id"`
				TaskID       string   `json:"task_id"`
				FilesTouched []string `json:"files_touched"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			prompt := forge.Review(input.ProjectKey, input.EpicID, input.TaskID, input.FilesTouched)
			return prompt, nil
		},
	}
}

func forgeReviewFeedback(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_review_feedback",
		Description: "Send review feedback to a worker. Tracks attempts (max 3).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_key", "task_id", "worker_id", "status"],
			"properties": {
				"project_key": {"type": "string"},
				"task_id":     {"type": "string"},
				"worker_id":   {"type": "string"},
				"status":      {"type": "string", "enum": ["approved", "needs_changes"]},
				"issues":      {"type": "string"},
				"summary":     {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectKey string `json:"project_key"`
				TaskID     string `json:"task_id"`
				WorkerID   string `json:"worker_id"`
				Status     string `json:"status"`
				Issues     string `json:"issues"`
				Summary    string `json:"summary"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.ReviewFeedback(store, input.ProjectKey, input.TaskID, input.WorkerID, input.Status, input.Issues, input.Summary)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeAdversarialReview() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_adversarial_review",
		Description: "VDD-style adversarial code review using hostile, fresh-context agent.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["diff"],
			"properties": {
				"diff":        {"type": "string"},
				"test_output": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				Diff       string `json:"diff"`
				TestOutput string `json:"test_output"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			prompt := forge.AdversarialReview(input.Diff, input.TestOutput)
			return prompt, nil
		},
	}
}

func forgeEvaluationPrompt() *registry.Tool {
	return &registry.Tool{
		Name:        "forge_evaluation_prompt",
		Description: "Generate self-evaluation prompt for a completed subtask.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["bead_id", "subtask_title", "files_touched"],
			"properties": {
				"bead_id":        {"type": "string"},
				"subtask_title":  {"type": "string"},
				"files_touched":  {"type": "array", "items": {"type": "string"}}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				BeadID       string   `json:"bead_id"`
				SubtaskTitle string   `json:"subtask_title"`
				FilesTouched []string `json:"files_touched"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			prompt := forge.EvaluationPrompt(input.BeadID, input.SubtaskTitle, input.FilesTouched)
			return prompt, nil
		},
	}
}

func forgeBroadcast(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_broadcast",
		Description: "Broadcast context update to all agents working on the same epic.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_path", "agent_name", "epic_id", "message"],
			"properties": {
				"project_path":   {"type": "string"},
				"agent_name":     {"type": "string"},
				"epic_id":        {"type": "string"},
				"message":        {"type": "string"},
				"importance":     {"type": "string", "enum": ["info", "warning", "blocker"]},
				"files_affected": {"type": "array", "items": {"type": "string"}}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ProjectPath   string   `json:"project_path"`
				AgentName     string   `json:"agent_name"`
				EpicID        string   `json:"epic_id"`
				Message       string   `json:"message"`
				Importance    string   `json:"importance"`
				FilesAffected []string `json:"files_affected"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			err := forge.Broadcast(store, input.ProjectPath, input.AgentName, input.EpicID, input.Message, input.Importance, input.FilesAffected)
			if err != nil {
				return "", err
			}
			return `{"status": "broadcast_sent"}`, nil
		},
	}
}

// --- Insights ---

func forgeGetStrategyInsights(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_get_strategy_insights",
		Description: "Get strategy success rates for decomposition planning.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["task"],
			"properties": {
				"task": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				Task string `json:"task"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.GetStrategyInsights(store, input.Task)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeGetFileInsights(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_get_file_insights",
		Description: "Get file-specific gotchas for worker context.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["files"],
			"properties": {
				"files": {"type": "array", "items": {"type": "string"}}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				Files []string `json:"files"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			result, err := forge.GetFileInsights(store, input.Files)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}

func forgeGetPatternInsights(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "forge_get_pattern_insights",
		Description: "Get common failure patterns across forges.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			result, err := forge.GetPatternInsights(store)
			if err != nil {
				return "", err
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		},
	}
}
