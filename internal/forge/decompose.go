package forge

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SelectStrategy recommends a decomposition strategy based on task description keywords.
// Returns one of: "file-based", "feature-based", "risk-based".
func SelectStrategy(task, context string) (string, error) {
	if task == "" {
		return "", fmt.Errorf("task description is required")
	}

	lower := strings.ToLower(task + " " + context)

	// Risk-based: security, migration, breaking changes, performance.
	riskKeywords := []string{"security", "migration", "breaking", "performance", "critical", "vulnerability", "upgrade"}
	for _, kw := range riskKeywords {
		if strings.Contains(lower, kw) {
			return "risk-based", nil
		}
	}

	// File-based: refactor, rename, move, reorganize, restructure.
	fileKeywords := []string{"refactor", "rename", "move", "reorganize", "restructure", "file", "directory"}
	for _, kw := range fileKeywords {
		if strings.Contains(lower, kw) {
			return "file-based", nil
		}
	}

	// Default: feature-based for new features, enhancements, etc.
	return "feature-based", nil
}

// PlanPrompt generates a structured prompt for task planning with a given strategy.
// This is a prompt generator -- it does NOT call an LLM.
func PlanPrompt(task, strategy, context string, maxSubtasks int) string {
	if maxSubtasks <= 0 {
		maxSubtasks = 5
	}
	if strategy == "" {
		strategy = "feature-based"
	}

	var sb strings.Builder
	sb.WriteString("## Task Decomposition Plan\n\n")
	_, _ = fmt.Fprintf(&sb, "**Strategy:** %s\n", strategy) // strings.Builder writes cannot fail.
	_, _ = fmt.Fprintf(&sb, "**Max Subtasks:** %d\n\n", maxSubtasks)
	_, _ = fmt.Fprintf(&sb, "**Task:** %s\n\n", task)

	if context != "" {
		_, _ = fmt.Fprintf(&sb, "**Context:**\n%s\n\n", context)
	}

	sb.WriteString("### Instructions\n\n")

	switch strategy {
	case "file-based":
		sb.WriteString("Decompose this task by file boundaries. Each subtask should modify a distinct set of files.\n")
		sb.WriteString("Group related file changes together. Minimize cross-file dependencies between subtasks.\n")
	case "risk-based":
		sb.WriteString("Decompose this task by risk level. Order subtasks from lowest to highest risk.\n")
		sb.WriteString("Isolate risky changes so they can be reviewed independently.\n")
	default:
		sb.WriteString("Decompose this task by feature/functionality. Each subtask delivers a coherent piece of functionality.\n")
		sb.WriteString("Ensure subtasks can be tested independently.\n")
	}

	sb.WriteString("\n### Output Format\n\n")
	sb.WriteString("Return a JSON object with this structure:\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{
  "epic_title": "string",
  "subtasks": [
    {
      "title": "string",
      "files": ["path/to/file.go"],
      "priority": 1
    }
  ]
}`)
	sb.WriteString("\n```\n")

	return sb.String()
}

// Decompose generates a decomposition prompt for breaking a task into subtasks.
// This is a prompt generator -- it does NOT call an LLM.
func Decompose(task, context string, maxSubtasks int) string {
	if maxSubtasks <= 0 {
		maxSubtasks = 5
	}

	var sb strings.Builder
	sb.WriteString("## Task Decomposition\n\n")
	_, _ = fmt.Fprintf(&sb, "Break the following task into at most %d independent subtasks.\n\n", maxSubtasks)
	_, _ = fmt.Fprintf(&sb, "**Task:** %s\n\n", task)

	if context != "" {
		_, _ = fmt.Fprintf(&sb, "**Context:**\n%s\n\n", context)
	}

	sb.WriteString("### Requirements\n\n")
	sb.WriteString("1. Each subtask should be independently executable\n")
	sb.WriteString("2. List the files each subtask will modify\n")
	sb.WriteString("3. Identify dependencies between subtasks\n")
	sb.WriteString("4. Assign priority (0=highest, 3=lowest)\n\n")

	sb.WriteString("### Output Format\n\n")
	sb.WriteString("Return a JSON object with this structure:\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{
  "epic_title": "string",
  "subtasks": [
    {
      "title": "string",
      "files": ["path/to/file.go"],
      "priority": 1
    }
  ]
}`)
	sb.WriteString("\n```\n")

	return sb.String()
}

// ValidateDecomposition validates a JSON decomposition response against the
// expected structure. Returns the parsed structure or an error describing
// what's missing.
func ValidateDecomposition(response string) (map[string]any, error) {
	if response == "" {
		return nil, fmt.Errorf("empty response")
	}

	// Try to extract JSON from markdown code blocks.
	cleaned := extractJSON(response)

	var result map[string]any
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Validate required fields.
	if _, ok := result["epic_title"]; !ok {
		return nil, fmt.Errorf("missing required field: epic_title")
	}

	subtasks, ok := result["subtasks"]
	if !ok {
		return nil, fmt.Errorf("missing required field: subtasks")
	}

	subtaskList, ok := subtasks.([]any)
	if !ok {
		return nil, fmt.Errorf("subtasks must be an array")
	}

	if len(subtaskList) == 0 {
		return nil, fmt.Errorf("subtasks array is empty")
	}

	// Validate each subtask has a title.
	for i, st := range subtaskList {
		stMap, ok := st.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("subtask %d is not an object", i)
		}
		if _, ok := stMap["title"]; !ok {
			return nil, fmt.Errorf("subtask %d missing required field: title", i)
		}
	}

	return result, nil
}

// extractJSON attempts to extract JSON from a markdown code block.
func extractJSON(s string) string {
	// Look for ```json ... ``` blocks.
	if idx := strings.Index(s, "```json"); idx >= 0 {
		start := idx + len("```json")
		if end := strings.Index(s[start:], "```"); end >= 0 {
			return strings.TrimSpace(s[start : start+end])
		}
	}
	// Look for ``` ... ``` blocks.
	if idx := strings.Index(s, "```"); idx >= 0 {
		start := idx + len("```")
		if end := strings.Index(s[start:], "```"); end >= 0 {
			return strings.TrimSpace(s[start : start+end])
		}
	}
	return strings.TrimSpace(s)
}
