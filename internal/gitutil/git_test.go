package gitutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a temporary git repo with an initial commit.
// Returns the repo path. Skips if testing.Short() is set.
func initRepo(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("requires git")
	}

	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if _, err := Run(dir, args...); err != nil {
			t.Fatalf("git %s: %v", strings.Join(args, " "), err)
		}
	}

	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Create an initial commit so HEAD exists.
	f := filepath.Join(dir, "README.md")
	if err := os.WriteFile(f, []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "initial commit")

	return dir
}

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("requires git")
	}
	dir := t.TempDir()
	if _, err := Run(dir, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	out, err := Run(dir, "status")
	if err != nil {
		t.Fatalf("git status: %v", err)
	}
	if !strings.Contains(out, "nothing to commit") && !strings.Contains(out, "No commits yet") {
		t.Errorf("unexpected status output: %s", out)
	}
}

func TestRun_Error(t *testing.T) {
	if testing.Short() {
		t.Skip("requires git")
	}
	dir := t.TempDir()
	_, err := Run(dir, "log")
	if err == nil {
		t.Error("expected error for git log in non-repo")
	}
}

func TestCurrentCommit(t *testing.T) {
	repo := initRepo(t)

	sha, err := CurrentCommit(repo)
	if err != nil {
		t.Fatalf("CurrentCommit: %v", err)
	}
	if len(sha) != 40 {
		t.Errorf("expected 40-char SHA, got %d chars: %s", len(sha), sha)
	}
}

func TestWorktreeAdd_And_List(t *testing.T) {
	repo := initRepo(t)

	commit, _ := CurrentCommit(repo)
	wtPath := filepath.Join(t.TempDir(), "wt-test")

	err := WorktreeAdd(repo, wtPath, "wt-branch", commit)
	if err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	// Verify the worktree directory exists.
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Fatal("worktree directory was not created")
	}

	// List worktrees.
	worktrees, err := WorktreeList(repo)
	if err != nil {
		t.Fatalf("WorktreeList: %v", err)
	}
	if len(worktrees) < 2 {
		t.Fatalf("expected at least 2 worktrees (main + new), got %d", len(worktrees))
	}

	// Find our worktree -- resolve symlinks for macOS /var -> /private/var.
	resolvedWtPath, _ := filepath.EvalSymlinks(wtPath)
	found := false
	for _, wt := range worktrees {
		if wt.Path == wtPath || wt.Path == resolvedWtPath {
			found = true
			if !strings.HasSuffix(wt.Branch, "wt-branch") {
				t.Errorf("branch = %q, want suffix %q", wt.Branch, "wt-branch")
			}
		}
	}
	if !found {
		t.Errorf("worktree %q (resolved: %q) not found in list: %+v", wtPath, resolvedWtPath, worktrees)
	}
}

func TestWorktreeRemove(t *testing.T) {
	repo := initRepo(t)

	commit, _ := CurrentCommit(repo)
	wtPath := filepath.Join(t.TempDir(), "wt-remove")

	if err := WorktreeAdd(repo, wtPath, "remove-branch", commit); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	err := WorktreeRemove(repo, wtPath)
	if err != nil {
		t.Fatalf("WorktreeRemove: %v", err)
	}

	// Verify directory is gone.
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("worktree directory should be removed")
	}
}

func TestCherryPick(t *testing.T) {
	repo := initRepo(t)

	startCommit, _ := CurrentCommit(repo)

	// Create a worktree and make a commit in it.
	wtPath := filepath.Join(t.TempDir(), "wt-cherry")
	if err := WorktreeAdd(repo, wtPath, "cherry-branch", startCommit); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	// Make a commit in the worktree.
	f := filepath.Join(wtPath, "new-file.txt")
	if err := os.WriteFile(f, []byte("cherry content\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := Run(wtPath, "add", "."); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err := Run(wtPath, "commit", "-m", "cherry commit"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// Cherry-pick from worktree branch back to main.
	err := CherryPick(repo, "cherry-branch", startCommit)
	if err != nil {
		t.Fatalf("CherryPick: %v", err)
	}

	// Verify the file exists in the main repo.
	mainFile := filepath.Join(repo, "new-file.txt")
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		t.Error("cherry-picked file should exist in main repo")
	}
}

func TestCherryPick_NoCommits(t *testing.T) {
	repo := initRepo(t)

	startCommit, _ := CurrentCommit(repo)

	// Create a worktree with no new commits.
	wtPath := filepath.Join(t.TempDir(), "wt-empty")
	if err := WorktreeAdd(repo, wtPath, "empty-branch", startCommit); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	// Cherry-pick should succeed with no-op.
	err := CherryPick(repo, "empty-branch", startCommit)
	if err != nil {
		t.Fatalf("CherryPick with no commits: %v", err)
	}
}

func TestWorktreeList_Empty(t *testing.T) {
	repo := initRepo(t)

	worktrees, err := WorktreeList(repo)
	if err != nil {
		t.Fatalf("WorktreeList: %v", err)
	}
	// Should have at least the main worktree.
	if len(worktrees) < 1 {
		t.Errorf("expected at least 1 worktree, got %d", len(worktrees))
	}
}
