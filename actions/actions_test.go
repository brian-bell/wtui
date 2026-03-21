package actions_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brian-bell/wt/actions"
)

func mustRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
}

func TestRemoveWorktree(t *testing.T) {
	// Set up a bare repo with a commit so worktrees work
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "repo")
	worktreePath := filepath.Join(dir, "wt")

	mustRun(t, dir, "git", "init", repoPath)
	mustRun(t, repoPath, "git", "config", "user.email", "test@test.com")
	mustRun(t, repoPath, "git", "config", "user.name", "Test")
	mustRun(t, repoPath, "git", "commit", "--allow-empty", "-m", "init")
	mustRun(t, repoPath, "git", "worktree", "add", worktreePath, "-b", "feat")

	// Worktree dir should exist before removal
	if _, err := os.Stat(worktreePath); err != nil {
		t.Fatalf("worktree dir should exist before removal: %v", err)
	}

	err := actions.RemoveWorktree(repoPath, worktreePath)
	if err != nil {
		t.Fatalf("RemoveWorktree returned error: %v", err)
	}

	// Worktree dir should be gone
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("expected worktree dir to be removed")
	}

	// git worktree list should no longer show the worktree
	out, _ := exec.Command("git", "-C", repoPath, "worktree", "list").Output()
	if strings.Contains(string(out), worktreePath) {
		t.Errorf("worktree still listed after removal:\n%s", out)
	}
}

func TestRemoveWorktree_Error(t *testing.T) {
	err := actions.RemoveWorktree("/nonexistent", "/also/nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent paths, got nil")
	}
}

func setupRepo(t *testing.T) (repoPath string) {
	t.Helper()
	dir := t.TempDir()
	repoPath = filepath.Join(dir, "repo")
	mustRun(t, dir, "git", "init", repoPath)
	mustRun(t, repoPath, "git", "config", "user.email", "test@test.com")
	mustRun(t, repoPath, "git", "config", "user.name", "Test")
	mustRun(t, repoPath, "git", "commit", "--allow-empty", "-m", "init")
	return repoPath
}

func TestForceRemoveWorktree(t *testing.T) {
	repoPath := setupRepo(t)
	worktreePath := filepath.Join(filepath.Dir(repoPath), "wt-dirty")

	mustRun(t, repoPath, "git", "worktree", "add", worktreePath, "-b", "dirty-feat")

	// Write a dirty file so normal remove fails
	if err := os.WriteFile(filepath.Join(worktreePath, "dirty.txt"), []byte("dirty"), 0644); err != nil {
		t.Fatal(err)
	}

	// Normal remove should fail
	if err := actions.RemoveWorktree(repoPath, worktreePath); err == nil {
		t.Fatal("expected normal remove to fail on dirty worktree")
	}

	// Force remove should succeed
	if err := actions.ForceRemoveWorktree(repoPath, worktreePath); err != nil {
		t.Fatalf("ForceRemoveWorktree returned error: %v", err)
	}

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("expected worktree dir to be removed after force")
	}
}

func TestDeleteBranch(t *testing.T) {
	repoPath := setupRepo(t)
	// Create and merge a branch so -d works
	mustRun(t, repoPath, "git", "checkout", "-b", "merged-feat")
	mustRun(t, repoPath, "git", "checkout", "-")

	if err := actions.DeleteBranch(repoPath, "merged-feat"); err != nil {
		t.Fatalf("DeleteBranch returned error: %v", err)
	}

	out, _ := exec.Command("git", "-C", repoPath, "branch").Output()
	if strings.Contains(string(out), "merged-feat") {
		t.Error("branch should be gone after DeleteBranch")
	}
}

func TestDeleteBranch_UnmergedFails(t *testing.T) {
	repoPath := setupRepo(t)
	mustRun(t, repoPath, "git", "checkout", "-b", "unmerged-feat")
	mustRun(t, repoPath, "git", "commit", "--allow-empty", "-m", "unmerged commit")
	mustRun(t, repoPath, "git", "checkout", "-")

	if err := actions.DeleteBranch(repoPath, "unmerged-feat"); err == nil {
		t.Error("expected DeleteBranch to fail for unmerged branch")
	}
}

func TestForceDeleteBranch(t *testing.T) {
	repoPath := setupRepo(t)
	mustRun(t, repoPath, "git", "checkout", "-b", "unmerged-feat")
	mustRun(t, repoPath, "git", "commit", "--allow-empty", "-m", "unmerged commit")
	mustRun(t, repoPath, "git", "checkout", "-")

	if err := actions.ForceDeleteBranch(repoPath, "unmerged-feat"); err != nil {
		t.Fatalf("ForceDeleteBranch returned error: %v", err)
	}

	out, _ := exec.Command("git", "-C", repoPath, "branch").Output()
	if strings.Contains(string(out), "unmerged-feat") {
		t.Error("branch should be gone after ForceDeleteBranch")
	}
}

func TestDropStash(t *testing.T) {
	repoPath := setupRepo(t)

	// Create a file and stash it
	if err := os.WriteFile(filepath.Join(repoPath, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	mustRun(t, repoPath, "git", "add", ".")
	mustRun(t, repoPath, "git", "stash")

	// Confirm stash exists
	out, _ := exec.Command("git", "-C", repoPath, "stash", "list").Output()
	if !strings.Contains(string(out), "stash@{0}") {
		t.Fatal("expected stash to exist before drop")
	}

	if err := actions.DropStash(repoPath, 0); err != nil {
		t.Fatalf("DropStash returned error: %v", err)
	}

	// Stash list should be empty
	out, _ = exec.Command("git", "-C", repoPath, "stash", "list").Output()
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected stash list empty after drop, got: %s", out)
	}
}

func TestDropStash_Error(t *testing.T) {
	err := actions.DropStash("/nonexistent", 0)
	if err == nil {
		t.Error("expected error for nonexistent repo, got nil")
	}
}
