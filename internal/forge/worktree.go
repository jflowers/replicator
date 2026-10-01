package forge

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/unbound-force/replicator/internal/gitutil"
)

// WorktreeCreate creates a git worktree for isolated task execution.
// The worktree is placed at <projectPath>/.worktrees/<taskID>.
func WorktreeCreate(projectPath, taskID, startCommit string) (map[string]any, error) {
	if projectPath == "" {
		return nil, fmt.Errorf("project_path is required")
	}
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	if startCommit == "" {
		return nil, fmt.Errorf("start_commit is required")
	}

	worktreePath := filepath.Join(projectPath, ".worktrees", taskID)
	branch := "swarm/" + taskID

	if err := gitutil.WorktreeAdd(projectPath, worktreePath, branch, startCommit); err != nil {
		return nil, fmt.Errorf("create worktree: %w", err)
	}

	return map[string]any{
		"status":        "created",
		"worktree_path": worktreePath,
		"branch":        branch,
		"start_commit":  startCommit,
		"task_id":       taskID,
	}, nil
}

// WorktreeMerge cherry-picks commits from a worktree branch back to the main branch.
func WorktreeMerge(projectPath, taskID, startCommit string) (map[string]any, error) {
	if projectPath == "" {
		return nil, fmt.Errorf("project_path is required")
	}
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}

	branch := "swarm/" + taskID

	if err := gitutil.CherryPick(projectPath, branch, startCommit); err != nil {
		return nil, fmt.Errorf("merge worktree: %w", err)
	}

	return map[string]any{
		"status":  "merged",
		"branch":  branch,
		"task_id": taskID,
	}, nil
}

// WorktreeCleanup removes a worktree. If cleanupAll is true, removes all
// worktrees in the .worktrees directory. Idempotent -- safe to call multiple times.
func WorktreeCleanup(projectPath string, taskID string, cleanupAll bool) (map[string]any, error) {
	return worktreeCleanupWithRemover(projectPath, taskID, cleanupAll, gitutil.WorktreeRemove)
}

type worktreeRemovalError struct {
	cause error
}

func (e *worktreeRemovalError) Error() string { return "remove worktree" }

func (e *worktreeRemovalError) Unwrap() error { return e.cause }

func worktreeCleanupWithRemover(projectPath string, taskID string, cleanupAll bool, remove func(string, string) error) (map[string]any, error) {
	if projectPath == "" {
		return nil, fmt.Errorf("project_path is required")
	}

	if cleanupAll {
		worktrees, err := gitutil.WorktreeList(projectPath)
		if err != nil {
			return nil, fmt.Errorf("list worktrees: %w", err)
		}

		removed := 0
		worktreeDir := filepath.Join(projectPath, ".worktrees")
		for _, wt := range worktrees {
			// Only remove worktrees under .worktrees/.
			if len(wt.Path) > len(worktreeDir) && wt.Path[:len(worktreeDir)] == worktreeDir {
				// Ignore errors for idempotency.
				if err := remove(projectPath, wt.Path); err == nil {
					removed++
				}
			}
		}

		return map[string]any{
			"status":  "cleaned",
			"removed": removed,
		}, nil
	}

	if taskID == "" {
		return nil, fmt.Errorf("task_id is required when cleanup_all is false")
	}

	worktreePath := filepath.Join(projectPath, ".worktrees", taskID)
	if _, err := os.Stat(worktreePath); errors.Is(err, os.ErrNotExist) {
		return map[string]any{"status": "cleaned", "task_id": taskID}, nil
	}
	if err := remove(projectPath, worktreePath); err != nil {
		// Preserve sentinel matching without exposing Git output or absolute paths at the MCP boundary.
		return nil, &worktreeRemovalError{cause: err}
	}

	return map[string]any{
		"status":  "cleaned",
		"task_id": taskID,
	}, nil
}

// WorktreeList lists all active worktrees for a project.
func WorktreeList(projectPath string) (map[string]any, error) {
	if projectPath == "" {
		return nil, fmt.Errorf("project_path is required")
	}

	worktrees, err := gitutil.WorktreeList(projectPath)
	if err != nil {
		return nil, fmt.Errorf("list worktrees: %w", err)
	}

	return map[string]any{
		"worktrees": worktrees,
		"count":     len(worktrees),
	}, nil
}
