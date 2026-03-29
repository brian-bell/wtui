package gitquery

import (
	"fmt"
	"os"
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

// Branch represents a local git branch with its status.
type Branch struct {
	Name          string
	HasUpstream   bool
	UpstreamGone  bool
	Ahead         int
	Behind        int
	Unpushed      []string
	IsWorktree    bool
	WorktreePaths []string
	WorktreeStale []bool // parallel to WorktreePaths; true when directory is missing
	Dirty         bool
	FilesChanged  int
	LinesAdded    int
	LinesDeleted  int
}

// BranchRow is one display row in the branch pane.
// A branch with N worktree paths expands into N rows.
type BranchRow struct {
	Branch       Branch
	WorktreePath string // specific path for this row; empty for non-worktree branches
	IsExpansion  bool   // true for 2nd+ rows of a multi-worktree branch
	Stale        bool   // true when the worktree directory no longer exists on disk
}

// FlattenBranches converts a branch list into display rows,
// expanding multi-worktree branches into one row per path.
func FlattenBranches(branches []Branch) []BranchRow {
	var rows []BranchRow
	for _, b := range branches {
		if len(b.WorktreePaths) == 0 {
			rows = append(rows, BranchRow{Branch: b})
			continue
		}
		for i, p := range b.WorktreePaths {
			stale := i < len(b.WorktreeStale) && b.WorktreeStale[i]
			rows = append(rows, BranchRow{Branch: b, WorktreePath: p, IsExpansion: i > 0, Stale: stale})
		}
	}
	return rows
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

const refFormat = "%(refname:short)\t%(upstream)\t%(upstream:track)"

// ListBranches returns all local branches sorted alphabetically by name.
func ListBranches(repoPath string) ([]Branch, error) {
	out, err := gitCmd(repoPath, "for-each-ref", "--format="+refFormat, "refs/heads/")
	if err != nil {
		return nil, err
	}

	wtMap, detachedPaths, err := branchWorktreeMap(repoPath)
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

		if wtPaths, ok := wtMap[b.Name]; ok {
			b.IsWorktree = true
			b.WorktreePaths = wtPaths
			b.WorktreeStale = checkStale(wtPaths)
			populateDirtyStatus(&b, wtPaths)
		}

		branches = append(branches, b)
	}

	for _, path := range detachedPaths {
		b := Branch{
			Name:          "(detached)",
			IsWorktree:    true,
			WorktreePaths: []string{path},
			WorktreeStale: checkStale([]string{path}),
		}
		populateDirtyStatus(&b, b.WorktreePaths)
		branches = append(branches, b)
	}

	sort.Slice(branches, func(i, j int) bool {
		if branches[i].Name != branches[j].Name {
			return branches[i].Name < branches[j].Name
		}
		return firstWorktreePath(branches[i].WorktreePaths) < firstWorktreePath(branches[j].WorktreePaths)
	})

	return branches, nil
}

// BranchDiff returns the diff output for a worktree.
func BranchDiff(worktreePath string) (string, error) {
	return gitCmd(worktreePath, "diff", "HEAD")
}

// worktreeInfo is internal data parsed from git worktree list output.
type worktreeInfo struct {
	path     string
	branch   string
	isBare   bool
	detached bool
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

func parseWorktreeBlock(block string) worktreeInfo {
	var wt worktreeInfo
	for _, line := range strings.Split(block, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			wt.path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch refs/heads/"):
			wt.branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "bare":
			wt.isBare = true
		case line == "detached":
			wt.detached = true
			wt.branch = "(detached)"
		}
	}
	return wt
}

// branchWorktreeMap returns a map of branch name -> worktree paths and detached worktree paths.
func branchWorktreeMap(repoPath string) (map[string][]string, []string, error) {
	out, err := gitCmd(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, nil, err
	}

	m := make(map[string][]string)
	var detachedPaths []string
	for _, block := range splitWorktreeBlocks(out) {
		wt := parseWorktreeBlock(block)
		if wt.isBare {
			continue
		}
		if wt.detached {
			detachedPaths = append(detachedPaths, wt.path)
			continue
		}
		if wt.branch != "" {
			m[wt.branch] = append(m[wt.branch], wt.path)
		}
	}
	return m, detachedPaths, nil
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

func populateDirtyStatus(b *Branch, paths []string) {
	for _, path := range paths {
		statusOut, err := gitCmd(path, "status", "--porcelain")
		if err != nil {
			continue
		}
		statusLines := splitLines(statusOut)
		if len(statusLines) == 0 {
			continue
		}
		b.Dirty = true
		b.FilesChanged += len(statusLines)

		diffOut, err := gitCmd(path, "diff", "HEAD", "--numstat")
		if err != nil {
			continue
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
}

func checkStale(paths []string) []bool {
	stale := make([]bool, len(paths))
	for i, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			stale[i] = true
		}
	}
	return stale
}

func firstWorktreePath(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
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
