package gitquery

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Worktree represents a single git worktree with its status.
type Worktree struct {
	Path     string
	Branch   string
	IsBare   bool
	Dirty    bool
	Ahead    int
	Behind   int
	Unpushed []string
}

// ListWorktrees returns worktree information for the given repo path.
func ListWorktrees(repoPath string) ([]Worktree, error) {
	out, err := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}

	var worktrees []Worktree
	for _, block := range splitWorktreeBlocks(string(out)) {
		wt := parseWorktreeBlock(block)
		if wt.Path == "" {
			continue
		}
		if !wt.IsBare {
			fillStatus(&wt)
		}
		worktrees = append(worktrees, wt)
	}
	return worktrees, nil
}

func splitWorktreeBlocks(output string) []string {
	var blocks []string
	var current []string
	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		if line == "" {
			if len(current) > 0 {
				blocks = append(blocks, strings.Join(current, "\n"))
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return blocks
}

func parseWorktreeBlock(block string) Worktree {
	var wt Worktree
	for _, line := range strings.Split(block, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			wt.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch refs/heads/"):
			wt.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "bare":
			wt.IsBare = true
		case line == "detached":
			wt.Branch = "(detached)"
		}
	}
	return wt
}

func fillStatus(wt *Worktree) {
	// Dirty check
	out, err := exec.Command("git", "-C", wt.Path, "status", "--porcelain").Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		wt.Dirty = true
	}

	// Ahead/behind
	out, err = exec.Command("git", "-C", wt.Path, "rev-list", "--count", "--left-right", "@{upstream}...HEAD").Output()
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(string(out)))
		if len(parts) == 2 {
			wt.Behind, _ = strconv.Atoi(parts[0])
			wt.Ahead, _ = strconv.Atoi(parts[1])
		}
	}

	// Unpushed commit messages
	out, err = exec.Command("git", "-C", wt.Path, "log", "--oneline", "@{upstream}..HEAD").Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if len(line) > 60 {
				line = line[:57] + "..."
			}
			wt.Unpushed = append(wt.Unpushed, line)
		}
	}
}
