package model_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brian-bell/wt/gitquery"
	"github.com/brian-bell/wt/model"
	"github.com/brian-bell/wt/scanner"
)

func testRepos() []scanner.Repo {
	return []scanner.Repo{
		{Path: "/dev/alpha", DisplayName: "alpha"},
		{Path: "/dev/bravo", DisplayName: "bravo"},
		{Path: "/dev/charlie", DisplayName: "charlie"},
	}
}

// update sends a message and returns the concrete Model.
func update(m model.Model, msg tea.Msg) (model.Model, tea.Cmd) {
	tm, cmd := m.Update(msg)
	return tm.(model.Model), cmd
}

func TestModel_InitialSelection(t *testing.T) {
	m := model.New(testRepos())
	if m.Selected() != 0 {
		t.Errorf("expected initial selected 0, got %d", m.Selected())
	}
}

func TestModel_DownArrowDoesNotMoveRepoSelection(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.Selected() != 0 {
		t.Errorf("expected selected unchanged at 0, got %d", m.Selected())
	}
}

func TestModel_UpArrowDoesNotMoveRepoSelection(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // move to 1
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.Selected() != 1 {
		t.Errorf("expected selected unchanged at 1, got %d", m.Selected())
	}
}

func TestModel_DownMovesStashCursorInMode2(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.StashSelected() != 1 {
		t.Errorf("expected StashSelected 1, got %d", m.StashSelected())
	}
}

func TestModel_UpMovesStashCursorInMode2(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.StashSelected() != 0 {
		t.Errorf("expected StashSelected 0, got %d", m.StashSelected())
	}
}

func TestModel_StashCursorClampsAtBounds(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	// Clamp at top
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.StashSelected() != 0 {
		t.Errorf("expected StashSelected clamped at 0, got %d", m.StashSelected())
	}
	// Clamp at bottom (3 stashes, max index 2)
	for i := 0; i < 10; i++ {
		m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.StashSelected() != 2 {
		t.Errorf("expected StashSelected clamped at 2, got %d", m.StashSelected())
	}
}

func TestModel_QuitReturnsQuitCmd(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestModel_CtrlCReturnsQuitCmd(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestModel_EscReturnsQuitCmd(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestModel_WindowSizeUpdates(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.Width() != 120 || m.Height() != 40 {
		t.Errorf("expected 120x40, got %dx%d", m.Width(), m.Height())
	}
}

func TestModel_EmptyReposNoPanic(t *testing.T) {
	m := model.New(nil)
	_ = m.View()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
}

func TestModel_ModeSwitchOnKeyPress(t *testing.T) {
	m := model.New(testRepos())

	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if m.Mode() != 2 {
		t.Errorf("expected mode 2, got %d", m.Mode())
	}

	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if m.Mode() != 1 {
		t.Errorf("expected mode 1, got %d", m.Mode())
	}
}

func TestModel_Key3IsNoOp(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if m.Mode() != 1 {
		t.Errorf("expected mode unchanged at 1, got %d", m.Mode())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if m.Mode() != 2 {
		t.Errorf("expected mode unchanged at 2, got %d", m.Mode())
	}
}

func TestModel_TabCyclesRepos(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.Selected() != 1 {
		t.Errorf("expected selected 1, got %d", m.Selected())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.Selected() != 2 {
		t.Errorf("expected selected 2, got %d", m.Selected())
	}
}

func TestModel_TabWrapsAtEnd(t *testing.T) {
	m := model.New(testRepos())
	// 3 repos: cycle past the last one
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.Selected() != 0 {
		t.Errorf("expected selected 0 (wrapped), got %d", m.Selected())
	}
}

func TestModel_TabDoesNotChangeMode(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.Mode() != 2 {
		t.Errorf("expected mode unchanged at 2, got %d", m.Mode())
	}
}

func TestModel_TabFiresFetchForCurrentMode(t *testing.T) {
	m := model.New(testRepos())
	// In mode 1 (branches), tab should fetch worktrees for new repo
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Fatal("expected fetch cmd on tab, got nil")
	}
}

func TestModel_ModeSwitchPreservesSelection(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // select bravo
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if m.Selected() != 1 {
		t.Errorf("expected selection preserved at 1, got %d", m.Selected())
	}
}

func TestModel_InitFiresFetchBranches(t *testing.T) {
	m := model.New(testRepos())
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected fetchBranches cmd from Init, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.BranchResultMsg); !ok {
		t.Errorf("expected BranchResultMsg from Init, got %T", msg)
	}
}

func TestModel_BranchResultUpdatesState(t *testing.T) {
	m := model.New(testRepos())
	branches := []gitquery.Branch{
		{Name: "main", HasUpstream: true},
		{Name: "feature", HasUpstream: true, Ahead: 1},
	}
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})
	if len(m.Branches()) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(m.Branches()))
	}
	if m.Branches()[0].Name != "main" {
		t.Errorf("expected first branch 'main', got %q", m.Branches()[0].Name)
	}
}

func TestModel_StaleBranchResultDiscarded(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // select bravo
	branches := []gitquery.Branch{{Name: "main"}}
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})
	if len(m.Branches()) != 0 {
		t.Errorf("expected stale result discarded, got %d branches", len(m.Branches()))
	}
}

func TestModel_Mode1SwitchFiresFetchBranches(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if cmd == nil {
		t.Fatal("expected fetch cmd on switch to mode 1, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.BranchResultMsg); !ok {
		t.Errorf("expected BranchResultMsg, got %T", msg)
	}
}

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

func TestModel_Pressing1WhileInMode1NoFetch(t *testing.T) {
	m := model.New(testRepos())
	// Already in mode 1; pressing 1 should not fire a redundant fetch
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if cmd != nil {
		t.Error("pressing 1 while already in mode 1 should not fire fetch")
	}
}

func TestModel_DefaultModeIsBranches(t *testing.T) {
	m := model.New(testRepos())
	if m.Mode() != 1 {
		t.Errorf("expected default mode 1, got %d", m.Mode())
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
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	view := m.View()
	if !strings.Contains(view, "nothing here yet") {
		t.Error("mode 2 should show placeholder")
	}
}

func TestModel_ViewStatusBarShowsTwoModes(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})

	view := m.View()
	// Mode 1 active
	if !strings.Contains(view, "[1] branches") {
		t.Error("mode 1 active: status bar should contain '[1] branches'")
	}
	if !strings.Contains(view, "2 stashes") {
		t.Error("mode 1 active: status bar should show inactive '2 stashes'")
	}
	// Should NOT contain mode 3
	if strings.Contains(view, "[3]") || strings.Contains(view, "3 ") {
		t.Error("status bar should not contain mode 3")
	}

	// Switch to mode 2
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	view = m.View()
	if !strings.Contains(view, "[2] stashes") {
		t.Error("mode 2 active: status bar should contain '[2] stashes'")
	}
	if !strings.Contains(view, "1 branches") {
		t.Error("mode 2 active: status bar should show inactive '1 branches'")
	}
}

func TestModel_ViewStatusBarShowsKeyHints(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})

	view := m.View()
	if !strings.Contains(view, "tab") {
		t.Error("status bar should contain 'tab' hint")
	}
	if !strings.Contains(view, "←/→") {
		t.Error("status bar should contain '←/→' hint")
	}
}

// --- Stash model tests (Slices 7-20) ---

func testStashes() []gitquery.Stash {
	return []gitquery.Stash{
		{Index: 0, Date: "2026-03-18 10:00:00 -0700", Message: "WIP: feature A"},
		{Index: 1, Date: "2026-03-17 09:00:00 -0700", Message: "backup: old approach"},
		{Index: 2, Date: "2026-03-16 08:00:00 -0700", Message: "experiment"},
	}
}

func TestModel_SwitchToMode2FiresFetchStashes(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if cmd == nil {
		t.Fatal("expected fetchStashes cmd on switch to mode 2, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.StashResultMsg); !ok {
		t.Errorf("expected StashResultMsg, got %T", msg)
	}
}

func TestModel_StashResultUpdatesState(t *testing.T) {
	m := model.New(testRepos())
	stashes := testStashes()
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: stashes})
	if len(m.Stashes()) != 3 {
		t.Fatalf("expected 3 stashes, got %d", len(m.Stashes()))
	}
}

func TestModel_StaleStashResultDiscarded(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // select bravo
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	if len(m.Stashes()) != 0 {
		t.Errorf("expected stale stash result discarded, got %d stashes", len(m.Stashes()))
	}
}

func TestModel_RightSwitchesToMode2(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.Mode() != 2 {
		t.Errorf("expected mode 2, got %d", m.Mode())
	}
}

func TestModel_LeftSwitchesToMode1(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.Mode() != 1 {
		t.Errorf("expected mode 1, got %d", m.Mode())
	}
}

func TestModel_HLSwitchModes(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if m.Mode() != 2 {
		t.Errorf("expected mode 2, got %d", m.Mode())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if m.Mode() != 1 {
		t.Errorf("expected mode 1, got %d", m.Mode())
	}
}

func TestModel_ModeClampsAtEdges(t *testing.T) {
	m := model.New(testRepos())
	// Already at mode 1, left should stay at 1
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.Mode() != 1 {
		t.Errorf("expected mode 1 (clamped), got %d", m.Mode())
	}
	// Go to mode 2, right should stay at 2
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.Mode() != 2 {
		t.Errorf("expected mode 2 (clamped), got %d", m.Mode())
	}
}

func TestModel_ModeSwitchViaArrowFiresFetch(t *testing.T) {
	m := model.New(testRepos())
	// Right to mode 2 should fetch stashes
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRight})
	if cmd == nil {
		t.Fatal("expected fetch cmd on mode switch to 2, got nil")
	}
	// Left back to mode 1 should fetch worktrees
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	_, cmd = update(m, tea.KeyMsg{Type: tea.KeyLeft})
	if cmd == nil {
		t.Fatal("expected fetch cmd on mode switch to 1, got nil")
	}
}

func TestModel_EnterOpensOverlay(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.Overlay() != model.OverlayStashDiff {
		t.Errorf("expected OverlayStashDiff, got %d", m.Overlay())
	}
	if cmd == nil {
		t.Fatal("expected fetchStashDiff cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.StashDiffResultMsg); !ok {
		t.Errorf("expected StashDiffResultMsg, got %T", msg)
	}
}

func TestModel_StashDiffResultStoresDiff(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, model.StashDiffResultMsg{RepoPath: "/dev/alpha", Index: 0, Diff: "diff --git a/f.txt"})
	if m.OverlayDiff() != "diff --git a/f.txt" {
		t.Errorf("expected diff stored, got %q", m.OverlayDiff())
	}
}

func TestModel_EscClosesOverlay(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	// Now close overlay with esc
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEscape})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed, got %d", m.Overlay())
	}
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Error("esc with overlay open should not quit")
		}
	}
}

func TestModel_QClosesOverlayNotQuit(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	// Close with q
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed, got %d", m.Overlay())
	}
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Error("q with overlay open should not quit")
		}
	}
}

func TestModel_OverlayScrollUpDown(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = update(m, model.StashDiffResultMsg{RepoPath: "/dev/alpha", Index: 0, Diff: "line1\nline2\nline3"})

	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.OverlayScroll() != 1 {
		t.Errorf("expected scroll 1, got %d", m.OverlayScroll())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.OverlayScroll() != 0 {
		t.Errorf("expected scroll 0, got %d", m.OverlayScroll())
	}
	// Up at 0 stays 0
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.OverlayScroll() != 0 {
		t.Errorf("expected scroll clamped at 0, got %d", m.OverlayScroll())
	}
}

func TestModel_ModeKeysIgnoredInOverlay(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	// Press "1" — should not change mode
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if m.Mode() != 2 {
		t.Errorf("expected mode unchanged at 2, got %d", m.Mode())
	}
	if m.Overlay() != model.OverlayStashDiff {
		t.Errorf("expected overlay still open, got %d", m.Overlay())
	}
}

func TestModel_ViewMode2ShowsStashContent(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
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
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight}) // mode 2

	view := m.View()
	if !strings.Contains(view, "enter") {
		t.Error("mode 2 status bar should mention 'enter'")
	}
	if !strings.Contains(view, "↑/↓") {
		t.Error("mode 2 status bar should mention '↑/↓'")
	}
}
