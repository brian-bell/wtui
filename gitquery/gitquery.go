package gitquery

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// Stash represents a single git stash entry.
type Stash struct {
	Index   int
	Date    string
	Message string
}

// ListStashes returns stash entries for the given repo path.
func ListStashes(repoPath string) ([]Stash, error) {
	text, err := gitCmd(repoPath, "stash", "list", "--format=%gd%x00%ai%x00%s")
	if err != nil {
		return nil, fmt.Errorf("listing stashes: %w", err)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	var stashes []Stash
	for _, line := range strings.Split(text, "\n") {
		parts := strings.SplitN(line, "\x00", 3)
		if len(parts) != 3 {
			continue
		}
		// parts[0] is like "stash@{0}"
		idxStr := strings.TrimPrefix(parts[0], "stash@{")
		idxStr = strings.TrimSuffix(idxStr, "}")
		idx, _ := strconv.Atoi(idxStr)
		stashes = append(stashes, Stash{
			Index:   idx,
			Date:    parts[1],
			Message: parts[2],
		})
	}
	return stashes, nil
}

// StashDiff returns the diff for a specific stash entry.
func StashDiff(repoPath string, index int) (string, error) {
	ref := fmt.Sprintf("stash@{%d}", index)
	out, err := gitCmd(repoPath, "stash", "show", "-p", ref)
	if err != nil {
		return "", fmt.Errorf("stash diff for %s: %w", ref, err)
	}
	return out, nil
}

// Worktree represents a single git worktree with its status.
type Worktree struct {
	Path        string
	Branch      string
	IsBare      bool
	Dirty       bool
	HasUpstream bool
	Ahead       int
	Behind      int
	Unpushed    []string
}

// ListWorktrees returns worktree information for the given repo path.
func ListWorktrees(repoPath string) ([]Worktree, error) {
	out, err := gitCmd(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}

	var worktrees []Worktree
	for _, block := range splitWorktreeBlocks(out) {
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
	out, err := gitCmd(wt.Path, "status", "--porcelain")
	if err == nil && len(strings.TrimSpace(out)) > 0 {
		wt.Dirty = true
	}

	// Ahead/behind
	out, err = gitCmd(wt.Path, "rev-list", "--count", "--left-right", "@{upstream}...HEAD")
	if err == nil {
		wt.HasUpstream = true
		parts := strings.Fields(strings.TrimSpace(out))
		if len(parts) == 2 {
			wt.Behind, _ = strconv.Atoi(parts[0])
			wt.Ahead, _ = strconv.Atoi(parts[1])
		}
	}

	// Unpushed commit messages
	out, err = gitCmd(wt.Path, "log", "--oneline", "@{upstream}..HEAD")
	if err == nil {
		wt.Unpushed = splitLines(out)
	}
}

// Branch represents a local git branch with its status.
type Branch struct {
	Name         string
	HasUpstream  bool
	UpstreamGone bool
	Ahead        int
	Behind       int
	Unpushed     []string
	IsWorktree   bool
	WorktreePath string
	Dirty        bool
	FilesChanged int
	LinesAdded   int
	LinesDeleted int
}

const refFormat = "%(refname:short)\t%(upstream)\t%(upstream:track)"

// ListBranches returns all local branches sorted alphabetically by name.
func ListBranches(repoPath string) ([]Branch, error) {
	out, err := gitCmd(repoPath, "for-each-ref", "--format="+refFormat, "refs/heads/")
	if err != nil {
		return nil, err
	}

	wtMap, err := branchWorktreeMap(repoPath)
	if err != nil {
		return nil, err
	}

	lines := splitLines(out)
	branches := make([]Branch, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		b, upstream := parseBranchLine(line)

		if b.HasUpstream && !b.UpstreamGone {
			ahead, behind, err := branchAheadBehind(repoPath, b.Name, upstream)
			if err == nil {
				b.Ahead = ahead
				b.Behind = behind
			}
			if b.Ahead > 0 {
				b.Unpushed = unpushedCommits(repoPath, b.Name, upstream)
			}
		}

		if wtPath, ok := wtMap[b.Name]; ok {
			b.IsWorktree = true
			b.WorktreePath = wtPath
			populateDirtyStatus(&b)
		}

		branches = append(branches, b)
	}

	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Name < branches[j].Name
	})

	return branches, nil
}

// BranchDiff returns the diff output for a worktree.
func BranchDiff(worktreePath string) (string, error) {
	return gitCmd(worktreePath, "diff", "HEAD")
}

// branchWorktreeMap returns a map of branch name -> worktree path.
func branchWorktreeMap(repoPath string) (map[string]string, error) {
	out, err := gitCmd(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, block := range splitWorktreeBlocks(out) {
		wt := parseWorktreeBlock(block)
		if wt.Branch != "" && !wt.IsBare {
			m[wt.Branch] = wt.Path
		}
	}
	return m, nil
}

func parseBranchLine(line string) (Branch, string) {
	parts := strings.SplitN(line, "\t", 3)
	b := Branch{Name: parts[0]}

	var upstream string
	if len(parts) > 1 && parts[1] != "" {
		b.HasUpstream = true
		upstream = parts[1]
		if len(parts) > 2 && strings.Contains(parts[2], "gone") {
			b.UpstreamGone = true
		}
	}

	return b, upstream
}

func branchAheadBehind(repoPath, branchName, upstream string) (int, int, error) {
	out, err := gitCmd(repoPath, "rev-list", "--count", "--left-right", branchName+"..."+upstream)
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) != 2 {
		return 0, 0, nil
	}

	ahead, _ := strconv.Atoi(parts[0])
	behind, _ := strconv.Atoi(parts[1])
	return ahead, behind, nil
}

func unpushedCommits(repoPath, branchName, upstream string) []string {
	out, err := gitCmd(repoPath, "log", "--oneline", upstream+".."+branchName)
	if err != nil {
		return nil
	}
	return splitLines(out)
}

func populateDirtyStatus(b *Branch) {
	statusOut, err := gitCmd(b.WorktreePath, "status", "--porcelain")
	if err != nil {
		return
	}
	statusLines := splitLines(statusOut)
	if len(statusLines) == 0 {
		return
	}
	b.Dirty = true
	b.FilesChanged = len(statusLines)

	diffOut, err := gitCmd(b.WorktreePath, "diff", "HEAD", "--numstat")
	if err != nil {
		return
	}
	for _, line := range splitLines(diffOut) {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		// Binary files show "-\t-\tfilename"; Atoi returns 0 for "-".
		added, _ := strconv.Atoi(fields[0])
		deleted, _ := strconv.Atoi(fields[1])
		b.LinesAdded += added
		b.LinesDeleted += deleted
	}
}

func gitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func splitLines(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
