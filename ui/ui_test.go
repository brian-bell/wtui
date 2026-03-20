package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/brian-bell/wt/gitquery"
	"github.com/brian-bell/wt/scanner"
)

func TestStatusBar_ActiveModeIsBracketed(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0)
	if !strings.Contains(bar, "[1] branches") {
		t.Error("active mode 1 should be bracketed")
	}
	if strings.Contains(bar, "[2]") {
		t.Error("inactive modes should not be bracketed")
	}

	bar = RenderStatusBar(120, 2, 0)
	if !strings.Contains(bar, "[2] stashes") {
		t.Error("active mode 2 should be bracketed")
	}
	if strings.Contains(bar, "[1]") {
		t.Error("inactive modes should not be bracketed")
	}
}

func TestStatusBar_Mode1ContainsIndicatorLegend(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0)
	for _, legend := range []string{"✔ clean", "● ahead/behind", "● dirty", "● no upstream"} {
		if !strings.Contains(bar, legend) {
			t.Errorf("mode 1 status bar should contain legend %q", legend)
		}
	}
}

func TestStatusBar_Mode2OmitsIndicatorLegend(t *testing.T) {
	bar := RenderStatusBar(120, 2, 0)
	if strings.Contains(bar, "clean") {
		t.Error("mode 2 status bar should not contain indicator legend")
	}
}

func TestStatusBar_ContainsHints(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0)
	for _, hint := range []string{"tab: repo", "←/→: mode", "q/esc: quit"} {
		if !strings.Contains(bar, hint) {
			t.Errorf("status bar should contain %q", hint)
		}
	}
}

func TestRepoList_ScrollsWhenSelectionExceedsHeight(t *testing.T) {
	repos := []scanner.Repo{
		{Path: "/a", DisplayName: "alpha"},
		{Path: "/b", DisplayName: "bravo"},
		{Path: "/c", DisplayName: "charlie"},
		{Path: "/d", DisplayName: "delta"},
		{Path: "/e", DisplayName: "echo"},
	}
	// Height of 3 means only 3 visible at a time
	lines := renderRepoList(repos, 4, 3) // selected=4 (echo), height=3
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "echo") {
		t.Error("selected item 'echo' should be visible")
	}
	if strings.Contains(joined, "alpha") {
		t.Error("'alpha' should be scrolled off the top")
	}
}

func TestBranchPane_CleanBranchShowsGreenCheck(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "main", HasUpstream: true, Ahead: 0, Behind: 0, Dirty: false},
	}
	lines := renderBranchPane(branches, 50, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "main") {
		t.Error("should contain branch name 'main'")
	}
	if !strings.Contains(joined, "✔") {
		t.Error("clean branch with upstream should show ✔")
	}
	if strings.Contains(joined, "●") {
		t.Error("clean branch with upstream should not show ●")
	}
}

func TestBranchPane_AheadBehindShowsYellowDotWithCounts(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "feature/auth", HasUpstream: true, Ahead: 3, Behind: 1, Dirty: false},
	}
	lines := renderBranchPane(branches, 60, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "●") {
		t.Error("ahead/behind branch should show ● indicator")
	}
	if !strings.Contains(joined, "+3/-1") {
		t.Error("should show ahead/behind counts as +3/-1")
	}
	if strings.Contains(joined, "✔") {
		t.Error("ahead/behind branch should not show ✔")
	}
}

func TestBranchPane_DirtyShowsRedDotWithFileStats(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "feature/wip", HasUpstream: true, Dirty: true, IsWorktree: true,
			FilesChanged: 3, LinesAdded: 10, LinesDeleted: 5},
	}
	lines := renderBranchPane(branches, 60, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "●") {
		t.Error("dirty branch should show ● indicator")
	}
	if !strings.Contains(joined, "3 files") {
		t.Error("dirty branch should show file count")
	}
	if !strings.Contains(joined, "+10") {
		t.Error("dirty branch should show lines added")
	}
	if !strings.Contains(joined, "-5") {
		t.Error("dirty branch should show lines deleted")
	}
}

func TestBranchPane_NoUpstreamShowsPurpleDot(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "local-only", HasUpstream: false},
	}
	lines := renderBranchPane(branches, 50, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "●") {
		t.Error("no-upstream branch should show ● indicator")
	}
	if strings.Contains(joined, "✔") {
		t.Error("no-upstream branch should not show ✔")
	}
}

func TestBranchPane_UpstreamGoneShowsPurpleDot(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "stale", HasUpstream: true, UpstreamGone: true},
	}
	lines := renderBranchPane(branches, 50, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "●") {
		t.Error("upstream-gone branch should show ● indicator")
	}
	if strings.Contains(joined, "✔") {
		t.Error("upstream-gone branch should not show ✔")
	}
}

func TestBranchPane_StacksAheadAndDirtyIndicators(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "feat", HasUpstream: true, Ahead: 2, Behind: 0, Dirty: true, IsWorktree: true,
			FilesChanged: 1, LinesAdded: 5, LinesDeleted: 2},
	}
	lines := renderBranchPane(branches, 80, 10)
	joined := strings.Join(lines, "\n")
	// Should have both +2/-0 (ahead) and 1 files (dirty)
	if !strings.Contains(joined, "+2/-0") {
		t.Error("stacked: should show ahead/behind counts")
	}
	if !strings.Contains(joined, "1 files") {
		t.Error("stacked: should show dirty file count")
	}
	// Should have two ● indicators
	if strings.Count(joined, "●") < 2 {
		t.Errorf("stacked: expected at least 2 dot indicators, got %d", strings.Count(joined, "●"))
	}
	if strings.Contains(joined, "✔") {
		t.Error("stacked: should not show ✔ when there are indicators")
	}
}

func TestBranchPane_WorktreeAnnotation(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "feat", HasUpstream: true, IsWorktree: true, WorktreePaths: []string{"/dev/proj-feat"}},
	}
	lines := renderBranchPane(branches, 60, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "[/dev/proj-feat]") {
		t.Error("worktree branch should show [<path>] annotation")
	}
}

func TestBranchPane_DuplicateWorktreeAnnotation(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "feat", HasUpstream: true, IsWorktree: true, WorktreePaths: []string{"/dev/proj-feat", "/tmp/proj-feat-copy"}},
	}
	lines := renderBranchPane(branches, 80, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "/dev/proj-feat") || !strings.Contains(joined, "/tmp/proj-feat-copy") {
		t.Error("duplicate worktree branch should show both paths")
	}
	if strings.Contains(joined, "duplicate") || strings.Contains(joined, "wt:") {
		t.Error("duplicate worktree branch should not show labels")
	}
}

func TestBranchPane_DetachedWorktreeRow(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "(detached)", IsWorktree: true, WorktreePaths: []string{"/tmp/wt-detached"}},
	}
	lines := renderBranchPane(branches, 80, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "(detached)") {
		t.Error("detached worktree should render as a detached row")
	}
	if !strings.Contains(joined, "[/tmp/wt-detached]") {
		t.Error("detached worktree should show its path annotation")
	}
}

func TestBranchPane_NonWorktreeNoAnnotation(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "feat", HasUpstream: true, IsWorktree: false},
	}
	lines := renderBranchPane(branches, 60, 10)
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "[wt:") || strings.Contains(joined, "[duplicate:") {
		t.Error("non-worktree branch should not show worktree annotation")
	}
}

func TestBranchPane_UnpushedCommitsShown(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "feat", HasUpstream: true, Ahead: 2,
			Unpushed: []string{"abc1234 Fix bug", "def5678 Add feature"}},
	}
	lines := renderBranchPane(branches, 60, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Fix bug") {
		t.Error("should show unpushed commit message")
	}
	if !strings.Contains(joined, "Add feature") {
		t.Error("should show second unpushed commit message")
	}
}

func TestBranchPane_UnpushedCapsAt5WithOverflow(t *testing.T) {
	msgs := make([]string, 8)
	for i := range msgs {
		msgs[i] = fmt.Sprintf("abc%d commit message %d", i, i)
	}
	branches := []gitquery.Branch{
		{Name: "feat", HasUpstream: true, Ahead: 8, Unpushed: msgs},
	}
	lines := renderBranchPane(branches, 60, 20)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "and 3 more") {
		t.Error("should show 'and 3 more' overflow for 8 commits with cap of 5")
	}
	// Count lines that contain commit content
	var commitLines int
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.Contains(trimmed, "commit message") || strings.Contains(trimmed, "and 3 more") {
			commitLines++
		}
	}
	if commitLines != 6 {
		t.Errorf("expected 6 commit-related lines (5 + overflow), got %d", commitLines)
	}
}

func TestRender_CombinesPanesWithDivider(t *testing.T) {
	view := Render(RenderParams{
		Repos:    []scanner.Repo{{Path: "/a", DisplayName: "alpha"}},
		Selected: 0,
		Width:    80,
		Height:   10,
		Mode:     1,
	})
	if !strings.Contains(view, "│") {
		t.Error("view should contain divider")
	}
	if !strings.Contains(view, "alpha") {
		t.Error("view should contain repo name")
	}
}
