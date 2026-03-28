package model_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brian-bell/wtui/gitquery"
	"github.com/brian-bell/wtui/model"
)

func TestModel_ViewShowsBranchData(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	branches := []gitquery.Branch{
		{Name: "main", HasUpstream: true},
		{Name: "feature/auth", HasUpstream: true, Ahead: 2, Behind: 1,
			Unpushed: []string{"abc1234 Fix login bug", "def5678 Add profile page"}},
	}
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})

	view := m.View()
	if !strings.Contains(view, "main") {
		t.Error("view should contain branch 'main'")
	}
	if !strings.Contains(view, "feature/auth") {
		t.Error("view should contain branch 'feature/auth'")
	}
	if !strings.Contains(view, "Fix login bug") {
		t.Error("view should contain unpushed commit message")
	}
	if !strings.Contains(view, "+2/-1") {
		t.Error("view should contain ahead/behind counts")
	}
}

func TestModel_ViewContainsExpectedContent(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()

	for _, name := range []string{"alpha", "bravo", "charlie"} {
		if !strings.Contains(view, name) {
			t.Errorf("view should contain repo name %q", name)
		}
	}
	if !strings.Contains(view, "q/esc: quit") {
		t.Error("view should contain quit keybinding")
	}
}

func TestModel_ViewMode2ShowsPlaceholder(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	view := m.View()
	if !strings.Contains(view, "nothing here yet") {
		t.Error("mode 2 should show placeholder")
	}
}

func TestModel_ViewModeHeaderShowsThreeModes(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})

	view := m.View()
	// Mode 1 active — mode header in right pane
	if !strings.Contains(view, "[1] branches") {
		t.Error("mode 1 active: right pane header should contain '[1] branches'")
	}
	if !strings.Contains(view, "2 stashes") {
		t.Error("mode 1 active: right pane header should show inactive '2 stashes'")
	}
	if !strings.Contains(view, "3 history") {
		t.Error("mode 1 active: right pane header should show inactive '3 history'")
	}

	// Switch to mode 2
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	view = m.View()
	if !strings.Contains(view, "[2] stashes") {
		t.Error("mode 2 active: right pane header should contain '[2] stashes'")
	}
	if !strings.Contains(view, "1 branches") {
		t.Error("mode 2 active: right pane header should show inactive '1 branches'")
	}

	// Switch to mode 3
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	view = m.View()
	if !strings.Contains(view, "[3] history") {
		t.Error("mode 3 active: right pane header should contain '[3] history'")
	}
	if !strings.Contains(view, "1 branches") {
		t.Error("mode 3 active: right pane header should show inactive '1 branches'")
	}
	if !strings.Contains(view, "2 stashes") {
		t.Error("mode 3 active: right pane header should show inactive '2 stashes'")
	}
}

func TestModel_ViewStatusBarShowsKeyHints(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})

	view := m.View()
	if !strings.Contains(view, "tab: pane") {
		t.Error("status bar should contain 'tab: pane' hint")
	}
}

func TestModel_ViewMode2ShowsStashContent(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})

	view := m.View()
	if !strings.Contains(view, "WIP: feature A") {
		t.Error("view should contain stash message 'WIP: feature A'")
	}
	if !strings.Contains(view, "backup: old approach") {
		t.Error("view should contain stash message 'backup: old approach'")
	}
}

func TestModel_ViewOverlayShowsDiff(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = update(m, model.StashDiffResultMsg{RepoPath: "/dev/alpha", Index: 0, Diff: "diff --git a/f.txt\n--- a/f.txt\n+++ b/f.txt"})

	view := m.View()
	if !strings.Contains(view, "diff --git") {
		t.Error("overlay should show diff content")
	}
	if !strings.Contains(view, "esc") {
		t.Error("overlay should show esc hint")
	}
}

func TestModel_StatusBarMode2ShowsStashKeys(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight}) // mode 2

	view := m.View()
	if !strings.Contains(view, "enter") {
		t.Error("mode 2 status bar should mention 'enter'")
	}
	if !strings.Contains(view, "↑/↓") {
		t.Error("mode 2 status bar should mention '↑/↓'")
	}
}

func TestModel_StatusBarMode2ShowsDropHint(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}}) // enable destructive
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})                     // mode 2

	view := m.View()
	if !strings.Contains(view, "d: drop") {
		t.Error("mode 2 status bar should mention 'd: drop' in destructive mode")
	}
}

// --- Destructive mode view tests ---

func TestModel_ViewReadOnlyHidesDeleteHint(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m = inRightPane(m)

	view := m.View()
	if strings.Contains(view, "d: delete") {
		t.Error("read-only mode should NOT show 'd: delete'")
	}
}

func TestModel_ViewReadOnlyHidesDropHint(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight}) // mode 2

	view := m.View()
	if strings.Contains(view, "d: drop") {
		t.Error("read-only mode should NOT show 'd: drop'")
	}
}

func TestModel_ViewReadOnlyShowsDestructiveModeHint(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})

	view := m.View()
	if !strings.Contains(view, "D: destructive mode") {
		t.Error("read-only mode should show 'D: destructive mode' hint")
	}
}

func TestModel_ViewDestructiveModeShowsDeleteHint(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})

	view := m.View()
	if !strings.Contains(view, "d: delete") {
		t.Error("destructive mode should show 'd: delete'")
	}
}

func TestModel_ViewMode3ShowsPlaceholder(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})

	view := m.View()
	if !strings.Contains(view, "nothing here yet") {
		t.Error("mode 3 with no commits should show placeholder")
	}
}

func TestModel_ViewMode3ShowsCommitContent(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m, _ = update(m, model.CommitResultMsg{RepoPath: "/dev/alpha", Commits: testCommits()})

	view := m.View()
	if !strings.Contains(view, "Fix login bug") {
		t.Error("view should contain commit subject 'Fix login bug'")
	}
	if !strings.Contains(view, "alice") {
		t.Error("view should contain author 'alice'")
	}
}

func TestModel_StatusBarMode3ShowsHistoryKeys(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})

	view := m.View()
	if !strings.Contains(view, "enter: diff") {
		t.Error("mode 3 status bar should mention 'enter: diff'")
	}
	if !strings.Contains(view, "y: copy hash") {
		t.Error("mode 3 status bar should mention 'y: copy hash'")
	}
	if !strings.Contains(view, "t: terminal") {
		t.Error("mode 3 status bar should mention 't: terminal'")
	}
	if !strings.Contains(view, "c: code") {
		t.Error("mode 3 status bar should mention 'c: code'")
	}
}

func TestModel_ViewDestructiveModeHidesDestructiveHint(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})

	view := m.View()
	if strings.Contains(view, "D: destructive mode") {
		t.Error("destructive mode should NOT show 'D: destructive mode' hint")
	}
}
