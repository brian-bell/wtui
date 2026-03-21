package actions

import "os/exec"

// RemoveWorktree runs `git worktree remove` for the given worktree path.
func RemoveWorktree(repoPath, worktreePath string) error {
	return exec.Command("git", "-C", repoPath, "worktree", "remove", worktreePath).Run()
}

// ForceRemoveWorktree runs `git worktree remove --force`.
func ForceRemoveWorktree(repoPath, worktreePath string) error {
	return exec.Command("git", "-C", repoPath, "worktree", "remove", "--force", worktreePath).Run()
}

// DeleteBranch runs `git branch -d`.
func DeleteBranch(repoPath, name string) error {
	return exec.Command("git", "-C", repoPath, "branch", "-d", name).Run()
}

// ForceDeleteBranch runs `git branch -D`.
func ForceDeleteBranch(repoPath, name string) error {
	return exec.Command("git", "-C", repoPath, "branch", "-D", name).Run()
}
