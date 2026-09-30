// Package org registers MCP tools for org cell operations.
package org

import (
	"encoding/json"

	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/org"
	"github.com/unbound-force/replicator/internal/tools/registry"
)

// Register adds all org tools to the registry.
func Register(reg *registry.Registry, store *db.Store) {
	reg.Register(orgCells(store))
	reg.Register(orgCreate(store))
	reg.Register(orgClose(store))
	reg.Register(orgUpdate(store))
	reg.Register(orgCreateEpic(store))
	reg.Register(orgQuery(store))
	reg.Register(orgStart(store))
	reg.Register(orgReady(store))
	reg.Register(orgSync(store))
	reg.Register(orgSessionStart(store))
	reg.Register(orgSessionEnd(store))
}

func orgCells(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_cells",
		Description: "Query cells from the org database with flexible filtering. Use to list work items, find by status/type, or look up a cell by partial ID.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id":     {"type": "string", "description": "Partial cell ID to match"},
				"status": {"type": "string", "enum": ["open", "in_progress", "blocked", "closed"]},
				"type":   {"type": "string", "enum": ["task", "bug", "feature", "epic", "chore"]},
				"ready":  {"type": "boolean", "description": "If true, return only unblocked cells"},
				"limit":  {"type": "number", "description": "Max results (default 50)"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var q org.CellQuery
			if len(args) > 0 {
				if err := json.Unmarshal(args, &q); err != nil {
					return "", err
				}
			}
			cells, err := org.QueryCells(store, q)
			if err != nil {
				return "", err
			}
			out, err := json.MarshalIndent(cells, "", "  ")
			return string(out), err
		},
	}
}

func orgCreate(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_create",
		Description: "Create a new cell (work item) in the org.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["title"],
			"properties": {
				"title":       {"type": "string"},
				"description": {"type": "string"},
				"type":        {"type": "string", "enum": ["task", "bug", "feature", "epic", "chore"]},
				"priority":    {"type": "number", "minimum": 0, "maximum": 3},
				"parent_id":   {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input org.CreateCellInput
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			cell, err := org.CreateCell(store, input)
			if err != nil {
				return "", err
			}
			out, err := json.MarshalIndent(cell, "", "  ")
			return string(out), err
		},
	}
}

func orgClose(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_close",
		Description: "Close a cell with a reason.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["id", "reason"],
			"properties": {
				"id":     {"type": "string"},
				"reason": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ID     string `json:"id"`
				Reason string `json:"reason"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			if err := org.CloseCell(store, input.ID, input.Reason); err != nil {
				return "", err
			}
			return `{"status": "closed"}`, nil
		},
	}
}

func orgUpdate(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_update",
		Description: "Update a cell's status, description, or priority.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["id"],
			"properties": {
				"id":          {"type": "string"},
				"status":      {"type": "string", "enum": ["open", "in_progress", "blocked", "closed"]},
				"description": {"type": "string"},
				"priority":    {"type": "number", "minimum": 0, "maximum": 3}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ID          string  `json:"id"`
				Status      *string `json:"status,omitempty"`
				Description *string `json:"description,omitempty"`
				Priority    *int    `json:"priority,omitempty"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			if err := org.UpdateCell(store, input.ID, input.Status, input.Description, input.Priority); err != nil {
				return "", err
			}
			return `{"status": "updated"}`, nil
		},
	}
}

func orgCreateEpic(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_create_epic",
		Description: "Create an epic with subtasks in one atomic operation.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["epic_title", "subtasks"],
			"properties": {
				"epic_title":       {"type": "string"},
				"epic_description": {"type": "string"},
				"subtasks": {
					"type": "array",
					"items": {
						"type": "object",
						"required": ["title"],
						"properties": {
							"title":    {"type": "string"},
							"priority": {"type": "number", "minimum": 0, "maximum": 3},
							"files":    {"type": "array", "items": {"type": "string"}}
						}
					}
				}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input org.CreateEpicInput
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			epic, subtasks, err := org.CreateEpic(store, input)
			if err != nil {
				return "", err
			}
			result := map[string]any{
				"epic":     epic,
				"subtasks": subtasks,
			}
			out, err := json.MarshalIndent(result, "", "  ")
			return string(out), err
		},
	}
}

func orgQuery(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_query",
		Description: "Query cells with filters (alias for org_cells).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id":     {"type": "string"},
				"status": {"type": "string", "enum": ["open", "in_progress", "blocked", "closed"]},
				"type":   {"type": "string", "enum": ["task", "bug", "feature", "epic", "chore"]},
				"ready":  {"type": "boolean"},
				"limit":  {"type": "number"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var q org.CellQuery
			if len(args) > 0 {
				if err := json.Unmarshal(args, &q); err != nil {
					return "", err
				}
			}
			cells, err := org.QueryCells(store, q)
			if err != nil {
				return "", err
			}
			out, err := json.MarshalIndent(cells, "", "  ")
			return string(out), err
		},
	}
}

func orgStart(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_start",
		Description: "Mark a cell as in-progress.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["id"],
			"properties": {
				"id": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			if err := org.StartCell(store, input.ID); err != nil {
				return "", err
			}
			return `{"status": "in_progress"}`, nil
		},
	}
}

func orgReady(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_ready",
		Description: "Get the next ready cell (unblocked, highest priority).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			cell, err := org.ReadyCell(store)
			if err != nil {
				return "", err
			}
			if cell == nil {
				return `{"message": "no ready cells"}`, nil
			}
			out, err := json.MarshalIndent(cell, "", "  ")
			return string(out), err
		},
	}
}

func orgSync(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_sync",
		Description: "Sync cells to git and push.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"auto_pull":    {"type": "boolean"},
				"project_path": {"type": "string", "description": "Project directory to sync (default: \".\")"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				AutoPull    bool   `json:"auto_pull,omitempty"`
				ProjectPath string `json:"project_path,omitempty"`
			}
			if len(args) > 0 {
				if err := json.Unmarshal(args, &input); err != nil {
					return "", err
				}
			}
			projectPath := input.ProjectPath
			if projectPath == "" {
				projectPath = "."
			}
			if err := org.Sync(store, projectPath); err != nil {
				return "", err
			}
			return `{"status": "synced"}`, nil
		},
	}
}

func orgSessionStart(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_session_start",
		Description: "Start a new work session. Returns previous session's handoff notes if available.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"active_cell_id": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				ActiveCellID string `json:"active_cell_id,omitempty"`
			}
			if len(args) > 0 {
				if err := json.Unmarshal(args, &input); err != nil {
					return "", err
				}
			}
			notes, err := org.SessionStart(store, input.ActiveCellID)
			if err != nil {
				return "", err
			}
			result := map[string]string{
				"status":        "started",
				"handoff_notes": notes,
			}
			out, err := json.MarshalIndent(result, "", "  ")
			return string(out), err
		},
	}
}

func orgSessionEnd(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "org_session_end",
		Description: "End current session with handoff notes for next session.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"handoff_notes": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				HandoffNotes string `json:"handoff_notes,omitempty"`
			}
			if len(args) > 0 {
				if err := json.Unmarshal(args, &input); err != nil {
					return "", err
				}
			}
			if err := org.SessionEnd(store, input.HandoffNotes); err != nil {
				return "", err
			}
			return `{"status": "ended"}`, nil
		},
	}
}
