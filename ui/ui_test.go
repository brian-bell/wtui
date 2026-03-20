package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/brian-bell/wt/gitquery"
	"github.com/brian-bell/wt/scanner"
)

func lipglossWidth(s string) int {
	return lipgloss.Width(s)
}

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
	for _, legend := range []string{"✔ clean", "● dirty", "● no upstream"} {
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

func TestWorktreePane_NoUpstreamShowsRedDot(t *testing.T) {
	wts := []gitquery.Worktree{
		{Path: "/dev/alpha", Branch: "main", HasUpstream: false},
	}
	lines := renderWorktreePane(wts, 50, 10)
	joined := strings.Join(lines, "\n")
	// Should contain the red dot indicator (● styled red)
	if !strings.Contains(joined, "●") {
		t.Error("no-upstream worktree should show red dot indicator")
	}
}

func TestWorktreePane_WithUpstreamNoRedDot(t *testing.T) {
	wts := []gitquery.Worktree{
		{Path: "/dev/alpha", Branch: "main", HasUpstream: true, Dirty: false},
	}
	lines := renderWorktreePane(wts, 50, 10)
	joined := strings.Join(lines, "\n")
	// Clean + has upstream → only ✔, no ●
	if strings.Contains(joined, "●") {
		t.Error("clean worktree with upstream should not show any dot indicator")
	}
}

func TestWorktreePane_SkipsBareWorktrees(t *testing.T) {
	wts := []gitquery.Worktree{
		{Path: "/bare", Branch: "", IsBare: true},
		{Path: "/dev/alpha", Branch: "main", Dirty: false},
	}
	lines := renderWorktreePane(wts, 50, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "main") {
		t.Error("should contain non-bare worktree branch")
	}
	// Bare worktree has no branch name to display, but ensure no extra entries
	// Count non-empty lines: should only have the "main" line
	var nonEmpty int
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	if nonEmpty != 1 {
		t.Errorf("expected 1 non-empty line (main only), got %d", nonEmpty)
	}
}

func TestWorktreePane_CapsUnpushedAt5(t *testing.T) {
	msgs := make([]string, 8)
	for i := range msgs {
		msgs[i] = "abc1234 commit message"
	}
	wts := []gitquery.Worktree{
		{Path: "/dev/alpha", Branch: "feat", Unpushed: msgs},
	}
	lines := renderWorktreePane(wts, 50, 20)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "and 3 more") {
		t.Error("should show 'and 3 more' for 8 commits with cap of 5")
	}
	// Count commit lines (indented with 4 spaces, contain commit style)
	var commitLines int
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.Contains(trimmed, "commit message") || strings.Contains(trimmed, "and 3 more") {
			commitLines++
		}
	}
	// 5 shown + 1 "and 3 more" = 6
	if commitLines != 6 {
		t.Errorf("expected 6 commit-related lines (5 + overflow), got %d", commitLines)
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

func TestWorktreePane_TruncatesLongBranchToWidth(t *testing.T) {
	longBranch := strings.Repeat("x", 80)
	wts := []gitquery.Worktree{
		{Path: "/dev/alpha", Branch: longBranch},
	}
	width := 40
	lines := renderWorktreePane(wts, width, 5)
	for _, l := range lines {
		// lipgloss.Width handles ANSI escape codes
		if lipglossWidth(l) > width {
			t.Errorf("line exceeds pane width %d: visual width %d", width, lipglossWidth(l))
		}
	}
}

func TestWorktreePane_TruncatesLongCommitToWidth(t *testing.T) {
	longMsg := "abc1234 " + strings.Repeat("w", 80)
	wts := []gitquery.Worktree{
		{Path: "/dev/alpha", Branch: "main", Unpushed: []string{longMsg}},
	}
	width := 40
	lines := renderWorktreePane(wts, width, 5)
	for _, l := range lines {
		if lipglossWidth(l) > width {
			t.Errorf("line exceeds pane width %d: visual width %d", width, lipglossWidth(l))
		}
	}
}

func TestWorktreePane_NoTrailingBlankWhenBareIsLast(t *testing.T) {
	wts := []gitquery.Worktree{
		{Path: "/dev/alpha", Branch: "main", Dirty: false},
		{Path: "/bare", Branch: "", IsBare: true},
	}
	lines := renderWorktreePane(wts, 50, 5)
	// Only "main" line should be non-empty; no trailing blank separator
	var nonEmpty int
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	if nonEmpty != 1 {
		t.Errorf("expected 1 non-empty line, got %d; trailing blank from bare entry?", nonEmpty)
	}
	// The line immediately after "main" should be empty padding, not a separator
	// caused by the bare entry's index check
	if strings.TrimSpace(lines[0]) == "" {
		t.Error("first line should be the main branch, not empty")
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
