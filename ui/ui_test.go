package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/brian-bell/wtui/gitquery"
	"github.com/brian-bell/wtui/scanner"
)

func TestStatusBar_Mode1ContainsIndicatorLegend(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 1, true, false)
	for _, legend := range []string{"✔ clean", "● ahead/behind", "● dirty", "● no upstream"} {
		if !strings.Contains(bar, legend) {
			t.Errorf("mode 1 status bar should contain legend %q", legend)
		}
	}
}

func TestStatusBar_IndicatorLegendSpacing(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 1, true, false)
	for _, pair := range [][2]string{
		{"clean", "●"},
	} {
		a := strings.Index(bar, pair[0])
		b := strings.Index(bar[a+len(pair[0]):], pair[1])
		if a == -1 || b == -1 {
			t.Errorf("expected both %q and %q in bar", pair[0], pair[1])
			continue
		}
		gap := bar[a+len(pair[0]) : a+len(pair[0])+b]
		if gap != "  " {
			t.Errorf("expected 2 spaces between legend items, got %q", gap)
		}
	}
}

func TestStatusBar_Mode2OmitsIndicatorLegend(t *testing.T) {
	bar := RenderStatusBar(120, 2, 0, 1, true, false)
	if strings.Contains(bar, "clean") {
		t.Error("mode 2 status bar should not contain indicator legend")
	}
}

func TestStatusBar_PipeSeparatesLegendAndHints(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 1, true, false)
	upstreamIdx := strings.Index(bar, "no upstream")
	tabIdx := strings.Index(bar, "tab: pane")
	if upstreamIdx == -1 || tabIdx == -1 {
		t.Fatalf("expected both 'no upstream' and 'tab: pane' in bar: %q", bar)
	}
	between := bar[upstreamIdx+len("no upstream") : tabIdx]
	if !strings.Contains(between, "|") {
		t.Errorf("expected pipe separator between legend and hints, got %q", between)
	}
}

func TestStatusBar_TabAndQuitBeforeOtherHints(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 1, true, false)
	tabIdx := strings.Index(bar, "tab: pane")
	tIdx := strings.Index(bar, "t: terminal")
	if tabIdx == -1 || tIdx == -1 {
		t.Fatalf("expected both hints in bar: %q", bar)
	}
	if tabIdx > tIdx {
		t.Error("tab: pane should appear before t: terminal")
	}
	qIdx := strings.Index(bar, "q/esc: quit")
	if qIdx > tIdx {
		t.Error("q/esc: quit should appear before t: terminal")
	}
}

func TestStatusBar_ActionHintsHiddenWhenLeftPaneActive(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 0, true, false) // activePane=0 (left), destructive=true
	for _, hint := range []string{"t: terminal", "c: code", "d: delete"} {
		if strings.Contains(bar, hint) {
			t.Errorf("hint %q should be hidden when left pane is active", hint)
		}
	}
	// tab and q/esc should still appear
	for _, hint := range []string{"tab: pane", "q/esc: quit"} {
		if !strings.Contains(bar, hint) {
			t.Errorf("hint %q should appear even when left pane is active", hint)
		}
	}
}

func TestStatusBar_ActionHintsShownWhenRightPaneActive(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 1, true, false) // activePane=1 (right)
	for _, hint := range []string{"t: terminal", "c: code", "d: delete"} {
		if !strings.Contains(bar, hint) {
			t.Errorf("hint %q should be shown when right pane is active", hint)
		}
	}
}

func TestStatusBar_KeyHintSpacingIs2(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 1, true, false)
	for _, pair := range [][2]string{
		{"tab: pane", "q/esc: quit"},
		{"t: terminal", "c: code"},
		{"c: code", "d: delete"},
	} {
		a := strings.Index(bar, pair[0])
		b := strings.Index(bar, pair[1])
		if a == -1 || b == -1 {
			t.Errorf("expected both %q and %q in bar", pair[0], pair[1])
			continue
		}
		gap := bar[a+len(pair[0]) : b]
		if gap != "  " {
			t.Errorf("expected 2 spaces between %q and %q, got %q", pair[0], pair[1], gap)
		}
	}
}

func TestModeHeader_ShowsActiveMode(t *testing.T) {
	header := renderModeHeader(1, 40)
	if !strings.Contains(header, "[1] branches") {
		t.Error("mode header should show active mode 1 bracketed")
	}
	if strings.Contains(header, "[2]") {
		t.Error("inactive mode 2 should not be bracketed")
	}
	header = renderModeHeader(2, 40)
	if !strings.Contains(header, "[2] stashes") {
		t.Error("mode header should show active mode 2 bracketed")
	}
}

func TestModeHeader_HasSeparatorLine(t *testing.T) {
	header := renderModeHeader(1, 40)
	lines := strings.Split(header, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected mode header to have at least 2 lines, got %d", len(lines))
	}
	// Second line should be a separator (dashes or similar)
	separator := lines[1]
	if !strings.Contains(separator, "─") {
		t.Errorf("expected separator line with ─ chars, got %q", separator)
	}
}

func TestRender_ModeHeaderInRightPane(t *testing.T) {
	view := Render(RenderParams{
		Repos:    []scanner.Repo{{Path: "/a", DisplayName: "alpha"}},
		Selected: 0,
		Width:    80,
		Height:   10,
		Mode:     1,
	})
	if !strings.Contains(view, "[1] branches") {
		t.Error("render should contain mode header '[1] branches' in right pane")
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
	// Height of 3 means only 3 visible at a time; scroll=2 shows repos 2-4
	lines := renderRepoList(repos, 4, 2, LeftPaneWidth-2, 3)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "echo") {
		t.Error("selected item 'echo' should be visible")
	}
	if strings.Contains(joined, "alpha") {
		t.Error("'alpha' should be scrolled off the top")
	}
}

func TestRepoList_TruncatesLongNames(t *testing.T) {
	width := LeftPaneWidth - 2
	repos := []scanner.Repo{
		{Path: "/a", DisplayName: "this-is-a-very-long-repository-name-that-exceeds-width"},
	}
	lines := renderRepoList(repos, 0, 0, width, 3)
	for i, line := range lines {
		if lipgloss.Width(line) > width {
			t.Errorf("line %d width %d exceeds pane width %d", i, lipgloss.Width(line), width)
		}
	}
}

func TestStashPane_LongMessageAlwaysShowsTwoLines(t *testing.T) {
	width := 50
	longMsg := "this is a very long stash message that should wrap to a second line always"
	stashes := []gitquery.Stash{
		{Index: 0, Date: "2026-03-18 10:00:00", Message: longMsg},
	}
	// Not selected (selected=-1): should still show 2 lines for the long message
	lines := renderStashPane(stashes, -1, 0, width, 10)
	// Count non-empty lines
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	if nonEmpty < 2 {
		t.Errorf("expected at least 2 non-empty lines for long stash message, got %d", nonEmpty)
	}
}

func TestStashPane_ShortMessageShowsOneLine(t *testing.T) {
	width := 50
	stashes := []gitquery.Stash{
		{Index: 0, Date: "2026-03-18 10:00:00", Message: "short"},
	}
	lines := renderStashPane(stashes, -1, 0, width, 10)
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	if nonEmpty != 1 {
		t.Errorf("expected 1 non-empty line for short stash message, got %d", nonEmpty)
	}
}

func TestStashPane_SelectedLongMessageHighlightsBothLines(t *testing.T) {
	width := 50
	// Message wraps to 2 lines but the remainder is short, so selected
	// (padded to full width) and unselected (unpadded) must differ.
	longMsg := "this is a long stash message that wraps ok"
	stashes := []gitquery.Stash{
		{Index: 0, Date: "2026-03-18 10:00:00", Message: longMsg},
	}
	// Render with stash selected vs not selected
	selLines := renderStashPane(stashes, 0, 0, width, 10)
	unselLines := renderStashPane(stashes, -1, 0, width, 10)

	// The continuation line (index 1) should differ between selected and
	// unselected renders — stashSelStyle.Width(width) pads the selected
	// continuation to full width, while the unselected one is unpadded.
	if selLines[1] == unselLines[1] {
		t.Error("continuation line should be styled differently when stash is selected")
	}
}

func TestStashPane_ScrollOffset(t *testing.T) {
	width := 50
	stashes := []gitquery.Stash{
		{Index: 0, Date: "2026-03-18", Message: "first"},
		{Index: 1, Date: "2026-03-17", Message: "second"},
		{Index: 2, Date: "2026-03-16", Message: "third"},
	}
	// scroll=1 should skip the first stash line
	lines := renderStashPane(stashes, 1, 1, width, 3)
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "first") {
		t.Error("'first' should be scrolled off the top")
	}
	if !strings.Contains(joined, "second") {
		t.Error("'second' should be visible")
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

func TestBranchPane_StaleWorktreeShowsStaleIndicator(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "stale-feat", HasUpstream: true, IsWorktree: true}, WorktreePath: "/dev/gone", Stale: true},
	}
	lines := renderBranchPane(rows, 60, 10)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "✗") {
		t.Error("stale worktree should show ✗ indicator")
	}
	if strings.Contains(joined, "✔") {
		t.Error("stale worktree should NOT show ✔ indicator")
	}
	if !strings.Contains(joined, "stale") {
		t.Error("stale worktree should show 'stale' label")
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
		ActivePane:     1,
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
		ActivePane:     1,
	})
	if !strings.Contains(view, "> dirty") {
		t.Error("dirty branch should be highlighted when BranchSelected=1")
	}
	if strings.Contains(view, "> clean") {
		t.Error("clean branch should not be highlighted when BranchSelected=1")
	}
}

func TestRender_HidesCursorWhenLeftPaneActive(t *testing.T) {
	view := Render(RenderParams{
		Repos:    []scanner.Repo{{Path: "/a", DisplayName: "alpha"}},
		Selected: 0,
		Width:    80,
		Height:   10,
		Mode:     1,
		Branches: []gitquery.BranchRow{
			{Branch: gitquery.Branch{Name: "main"}},
		},
		BranchSelected: 0,
		ActivePane:     0,
	})
	if strings.Contains(view, "> main") {
		t.Error("branch cursor should be hidden when left pane is active")
	}
}

func TestBranchPane_CursorDoesNotShiftBranchName(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "first", HasUpstream: true}},
		{Branch: gitquery.Branch{Name: "second", HasUpstream: true}},
	}
	// Render with no selection (selected = -1)
	unselected := renderBranchPaneSelected(rows, -1, 0, 80, 10, "/dev/alpha")
	// Render with first selected
	selected := renderBranchPaneSelected(rows, 0, 0, 80, 10, "/dev/alpha")

	// Find position of "first" in both renders — should be at the same column
	unselIdx := strings.Index(unselected[0], "first")
	selIdx := strings.Index(selected[0], "first")
	if unselIdx == -1 || selIdx == -1 {
		t.Fatalf("branch name 'first' not found in output: unsel=%q sel=%q", unselected[0], selected[0])
	}
	if unselIdx != selIdx {
		t.Errorf("branch name shifts when selected: unselected col %d, selected col %d", unselIdx, selIdx)
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
	lines := renderBranchPaneSelected(rows, 9, 8, 60, 3, "")
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
		Overlay:       3,
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
		Overlay:       3,
		ConfirmPrompt: "Force delete /dev/alpha/feat? (y/n)",
		ConfirmForce:  true,
	})
	if !strings.Contains(view, "Force delete /dev/alpha/feat") {
		t.Error("force confirm dialog should show prompt text")
	}
}

func TestStatusBar_PruneHintShownWhenStale(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 1, true, true)
	if !strings.Contains(bar, "p: prune") {
		t.Errorf("expected 'p: prune' hint when stale worktree selected, got %q", bar)
	}
}

func TestStatusBar_PruneHintHiddenWhenNotStale(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 1, true, false)
	if strings.Contains(bar, "p: prune") {
		t.Error("'p: prune' should not appear when no stale worktree selected")
	}
}

func TestStatusBar_PruneHintHiddenWithoutDestructive(t *testing.T) {
	bar := RenderStatusBar(120, 1, 0, 1, false, true)
	if strings.Contains(bar, "p: prune") {
		t.Error("'p: prune' should not appear without destructive mode")
	}
}

func TestStatusBar_Mode2HintsSpacing(t *testing.T) {
	bar := RenderStatusBar(120, 2, 0, 1, true, false)
	for _, pair := range [][2]string{
		{"tab: pane", "q/esc: quit"},
		{"↑/↓ select", "enter: diff"},
		{"enter: diff", "d: drop"},
	} {
		a := strings.Index(bar, pair[0])
		b := strings.Index(bar, pair[1])
		if a == -1 || b == -1 {
			t.Errorf("expected both %q and %q in bar", pair[0], pair[1])
			continue
		}
		gap := bar[a+len(pair[0]) : b]
		if gap != "  " {
			t.Errorf("expected 2 spaces between %q and %q, got %q", pair[0], pair[1], gap)
		}
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

func TestBranchPane_MainWorktreeShowsRootLabelAfterIndicators(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "main", HasUpstream: true, IsWorktree: true}, WorktreePath: "/dev/alpha"},
	}
	lines := renderBranchPaneSelected(rows, 0, 0, 80, 10, "/dev/alpha")
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "[root]") {
		t.Errorf("main worktree branch should show [root] label, got: %q", joined)
	}
	// [root] should appear after the branch name, not before
	mainIdx := strings.Index(joined, "main")
	rootIdx := strings.Index(joined, "[root]")
	if mainIdx == -1 || rootIdx == -1 || rootIdx < mainIdx {
		t.Errorf("expected [root] after branch name 'main', got: %q", joined)
	}
}

func TestBranchPane_AdditionalWorktreeShowsPath(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "feat", HasUpstream: true, IsWorktree: true}, WorktreePath: "/dev/alpha-worktrees/feat"},
	}
	lines := renderBranchPaneSelected(rows, 0, 0, 80, 10, "/dev/alpha")
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "/dev/alpha-worktrees/feat") {
		t.Errorf("additional worktree branch should show path, got: %q", joined)
	}
	if strings.Contains(joined, "[root]") {
		t.Error("additional worktree branch should not show [root]")
	}
}

func TestBranchPane_NonWorktreeBranchShowsNoLabel(t *testing.T) {
	rows := []gitquery.BranchRow{
		{Branch: gitquery.Branch{Name: "stale", HasUpstream: true}, WorktreePath: ""},
	}
	lines := renderBranchPaneSelected(rows, 0, 0, 80, 10, "/dev/alpha")
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "[root]") {
		t.Error("non-worktree branch should not show [root]")
	}
}
