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
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "main", HasUpstream: true, Ahead: 0, Behind: 0, Dirty: false}},
	}
	lines := renderBranchPane(rows, 50, 10)
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
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "feature/auth", HasUpstream: true, Ahead: 3, Behind: 1, Dirty: false}},
	}
	lines := renderBranchPane(rows, 60, 10)
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
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "feature/wip", HasUpstream: true, Dirty: true, IsWorktree: true,
			FilesChanged: 3, LinesAdded: 10, LinesDeleted: 5}},
	}
	lines := renderBranchPane(rows, 60, 10)
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
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "local-only", HasUpstream: false}},
	}
	lines := renderBranchPane(rows, 50, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "●") {
		t.Error("no-upstream branch should show ● indicator")
	}
	if strings.Contains(joined, "✔") {
		t.Error("no-upstream branch should not show ✔")
	}
}

func TestBranchPane_UpstreamGoneShowsPurpleDot(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "stale", HasUpstream: true, UpstreamGone: true}},
	}
	lines := renderBranchPane(rows, 50, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "●") {
		t.Error("upstream-gone branch should show ● indicator")
	}
	if strings.Contains(joined, "✔") {
		t.Error("upstream-gone branch should not show ✔")
	}
}

func TestBranchPane_StacksAheadAndDirtyIndicators(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "feat", HasUpstream: true, Ahead: 2, Behind: 0, Dirty: true, IsWorktree: true,
			FilesChanged: 1, LinesAdded: 5, LinesDeleted: 2}},
	}
	lines := renderBranchPane(rows, 80, 10)
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
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "feat", HasUpstream: true, IsWorktree: true}, WorktreePath: "/dev/proj-feat"},
	}
	lines := renderBranchPane(rows, 60, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "[/dev/proj-feat]") {
		t.Error("worktree branch should show [<path>] annotation")
	}
}

func TestBranchPane_DuplicateWorktreeAnnotation(t *testing.T) {
	b := gitquery.Branch{Name: "feat", HasUpstream: true, IsWorktree: true}
	rows := []gitquery.BranchRow{
		{Branch: b, WorktreePath: "/dev/proj-feat"},
		{Branch: b, WorktreePath: "/tmp/proj-feat-copy", IsExpansion: true},
	}
	lines := renderBranchPane(rows, 80, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "/dev/proj-feat") || !strings.Contains(joined, "/tmp/proj-feat-copy") {
		t.Error("duplicate worktree branch should show both paths")
	}
	if strings.Contains(joined, "duplicate") || strings.Contains(joined, "wt:") {
		t.Error("duplicate worktree branch should not show labels")
	}
}

func TestBranchPane_DetachedWorktreeRow(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "(detached)", IsWorktree: true}, WorktreePath: "/tmp/wt-detached"},
	}
	lines := renderBranchPane(rows, 80, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "(detached)") {
		t.Error("detached worktree should render as a detached row")
	}
	if !strings.Contains(joined, "[/tmp/wt-detached]") {
		t.Error("detached worktree should show its path annotation")
	}
}

func TestBranchPane_NonWorktreeNoAnnotation(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "feat", HasUpstream: true, IsWorktree: false}},
	}
	lines := renderBranchPane(rows, 60, 10)
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "[wt:") || strings.Contains(joined, "[duplicate:") {
		t.Error("non-worktree branch should not show worktree annotation")
	}
}

func TestRender_HighlightsSelectedBranch(t *testing.T) {
	// BranchSelected: 0 highlights first branch (clean), not the dirty one
	view := Render(RenderParams{
		Repos:    []scanner.Repo{{Path: "/a", DisplayName: "alpha"}},
		Selected: 0,
		Width:    80,
		Height:   10,
		Mode:     1,
		Branches: []gitquery.BranchRow{
			{Branch: gitquery.Branch{Name: "clean"}},
			{Branch: gitquery.Branch{Name: "dirty", IsWorktree: true, Dirty: true}, WorktreePath: "/a"},
		},
		BranchSelected: 0,
	})
	if !strings.Contains(view, "> clean") {
		t.Error("first branch should be highlighted when BranchSelected=0")
	}
	if strings.Contains(view, "> dirty") {
		t.Error("dirty branch should not be highlighted when BranchSelected=0")
	}
}

func TestRender_HighlightsSecondBranch(t *testing.T) {
	view := Render(RenderParams{
		Repos:    []scanner.Repo{{Path: "/a", DisplayName: "alpha"}},
		Selected: 0,
		Width:    80,
		Height:   10,
		Mode:     1,
		Branches: []gitquery.BranchRow{
			{Branch: gitquery.Branch{Name: "clean"}},
			{Branch: gitquery.Branch{Name: "dirty", IsWorktree: true, Dirty: true}, WorktreePath: "/a"},
		},
		BranchSelected: 1,
	})
	if !strings.Contains(view, "> dirty") {
		t.Error("dirty branch should be highlighted when BranchSelected=1")
	}
	if strings.Contains(view, "> clean") {
		t.Error("clean branch should not be highlighted when BranchSelected=1")
	}
}

func TestBranchPane_UnpushedCommitsShown(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "feat", HasUpstream: true, Ahead: 2,
			Unpushed: []string{"abc1234 Fix bug", "def5678 Add feature"}}},
	}
	lines := renderBranchPane(rows, 60, 10)
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
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "feat", HasUpstream: true, Ahead: 8, Unpushed: msgs}},
	}
	lines := renderBranchPane(rows, 60, 20)
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

func TestBranchPane_ScrollsToSelectedBranch(t *testing.T) {
	rows := make([]gitquery.BranchRow, 10)
	for i := range rows {
		rows[i] = gitquery.BranchRow{Branch: gitquery.Branch{Name: fmt.Sprintf("branch-%d", i)}}
	}
	// BranchScroll=8 with height=3 means we see branches 8 and 9
	lines := renderBranchPaneSelected(rows, 9, 8, 60, 3)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "branch-9") {
		t.Error("should show branch-9 when scrolled to see it")
	}
	if strings.Contains(joined, "branch-0") {
		t.Error("branch-0 should be scrolled out of view")
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

func TestRender_ConfirmDialogShowsPrompt(t *testing.T) {
	view := Render(RenderParams{
		Repos:         []scanner.Repo{{Path: "/dev/alpha", DisplayName: "alpha"}},
		Width:         80,
		Height:        24,
		Mode:          1,
		Overlay:       int(3), // OverlayConfirm
		ConfirmPrompt: "Remove worktree /dev/alpha/feat? (y/n)",
	})
	if !strings.Contains(view, "Remove worktree /dev/alpha/feat") {
		t.Error("confirm dialog should show prompt text")
	}
	if !strings.Contains(view, "y/n") {
		t.Error("confirm dialog should show y/n hint")
	}
}

func TestRender_ForceConfirmDialogShowsPrompt(t *testing.T) {
	view := Render(RenderParams{
		Repos:         []scanner.Repo{{Path: "/dev/alpha", DisplayName: "alpha"}},
		Width:         80,
		Height:        24,
		Mode:          1,
		Overlay:       int(3), // OverlayConfirm
		ConfirmPrompt: "Force delete /dev/alpha/feat? (y/n)",
		ConfirmForce:  true,
	})
	if !strings.Contains(view, "Force delete /dev/alpha/feat") {
		t.Error("force confirm dialog should show prompt text")
	}
}

func TestStatusBar_ShowsRefreshHint(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0)
	if !strings.Contains(bar, "r: refresh") {
		t.Errorf("status bar should contain 'r: refresh', got: %q", bar)
	}
	bar = RenderStatusBar(120, 2, 0)
	if !strings.Contains(bar, "r: refresh") {
		t.Errorf("mode 2 status bar should contain 'r: refresh', got: %q", bar)
	}
}

func TestStatusBar_ShowsDeleteHintInMode1(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0)
	if !strings.Contains(bar, "d: delete") {
		t.Errorf("mode 1 status bar should always contain 'd: delete', got: %q", bar)
	}
}

func TestBranchPane_MultiWorktreeExpandsRows(t *testing.T) {
	b := gitquery.Branch{Name: "feat", HasUpstream: true, Unpushed: []string{"abc1234 Fix thing"}}
	rows := []gitquery.BranchRow{
		{Branch: b, WorktreePath: "/dev/feat-A"},
		{Branch: b, WorktreePath: "/dev/feat-B", IsExpansion: true},
	}
	lines := renderBranchPane(rows, 80, 10)
	joined := strings.Join(lines, "\n")
	// Both paths should appear
	if !strings.Contains(joined, "/dev/feat-A") {
		t.Error("should show first worktree path /dev/feat-A")
	}
	if !strings.Contains(joined, "/dev/feat-B") {
		t.Error("should show second worktree path /dev/feat-B")
	}
	// Unpushed commit should appear once (on first row), not on expansion row
	if strings.Count(joined, "Fix thing") != 1 {
		t.Errorf("unpushed commit should appear exactly once, got %d", strings.Count(joined, "Fix thing"))
	}
}
