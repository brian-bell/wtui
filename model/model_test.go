package model_test

import (
	"fmt"
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

func TestModel_UpDownNavigatesAllBranches(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "clean-1"},
		{Name: "dirty-1", IsWorktree: true, Dirty: true, WorktreePaths: []string{"/dev/alpha"}},
		{Name: "clean-2"},
		{Name: "dirty-2", IsWorktree: true, Dirty: true, WorktreePaths: []string{"/dev/alpha"}},
	}
	m := model.New(testRepos())
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})

	if m.BranchSelected() != 0 {
		t.Errorf("expected cursor at 0, got %d", m.BranchSelected())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.BranchSelected() != 1 {
		t.Errorf("expected cursor at 1, got %d", m.BranchSelected())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.BranchSelected() != 2 {
		t.Errorf("expected cursor at 2, got %d", m.BranchSelected())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.BranchSelected() != 3 {
		t.Errorf("expected cursor at 3, got %d", m.BranchSelected())
	}
	// Clamp at last
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.BranchSelected() != 3 {
		t.Errorf("expected cursor clamped at 3, got %d", m.BranchSelected())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.BranchSelected() != 2 {
		t.Errorf("expected cursor at 2 on up, got %d", m.BranchSelected())
	}
}

func TestModel_EnterStillRequiresDirtyWorktree(t *testing.T) {
	branches := []gitquery.Branch{
		{Name: "clean-1"},
		{Name: "dirty-1", IsWorktree: true, Dirty: true, WorktreePaths: []string{"/dev/alpha"}},
		{Name: "clean-2"},
	}
	m := model.New(testRepos())
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})

	// Cursor at clean-1 (index 0): enter is no-op
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter on clean branch should be no-op")
	}

	// Navigate to dirty-1 (index 1): enter opens diff
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	_, cmd = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on dirty+worktree branch should open diff")
	}
	msg := cmd()
	diffMsg, ok := msg.(model.BranchDiffResultMsg)
	if !ok {
		t.Fatalf("expected BranchDiffResultMsg, got %T", msg)
	}
	if diffMsg.BranchName != "dirty-1" {
		t.Errorf("expected dirty-1, got %q", diffMsg.BranchName)
	}

	// Navigate to clean-2 (index 2): enter is no-op
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	_, cmd = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter on clean-2 should be no-op")
	}
}

func TestModel_EnterOpensBranchDiffOverlayForDirtyWorktree(t *testing.T) {
	m := model.New(testRepos())
	branches := []gitquery.Branch{
		{
			Name:          "feat",
			IsWorktree:    true,
			Dirty:         true,
			WorktreePaths: []string{"/dev/alpha"},
		},
		{Name: "main"},
	}
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})

	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.Overlay() != model.OverlayBranchDiff {
		t.Errorf("expected OverlayBranchDiff, got %d", m.Overlay())
	}
	if cmd == nil {
		t.Fatal("expected fetchBranchDiff cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.BranchDiffResultMsg); !ok {
		t.Errorf("expected BranchDiffResultMsg, got %T", msg)
	}
}

func TestModel_EnterDoesNothingForCleanBranch(t *testing.T) {
	m := model.New(testRepos())
	branches := []gitquery.Branch{
		{
			Name:       "feat",
			IsWorktree: true,
			Dirty:      false,
		},
	}
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})

	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected OverlayNone, got %d", m.Overlay())
	}
	if cmd != nil {
		t.Fatalf("expected no command for clean branch, got %T", cmd)
	}
}

func TestModel_StaleBranchDiffResultDiscarded(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // select bravo

	m, _ = update(m, model.BranchDiffResultMsg{
		RepoPath:   "/dev/alpha",
		BranchName: "feat",
		Diff:       "diff --git a/f.txt b/f.txt",
	})

	if m.OverlayDiff() != "" {
		t.Errorf("expected stale branch diff discarded, got %q", m.OverlayDiff())
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
	if len(m.Rows()) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(m.Rows()))
	}
	if m.Rows()[0].Branch.Name != "main" {
		t.Errorf("expected first branch 'main', got %q", m.Rows()[0].Branch.Name)
	}
}

func TestModel_StaleBranchResultDiscarded(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // select bravo
	branches := []gitquery.Branch{{Name: "main"}}
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})
	if len(m.Rows()) != 0 {
		t.Errorf("expected stale result discarded, got %d rows", len(m.Rows()))
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

// --- Confirmation dialog + refresh tests ---

func worktreeBranch() gitquery.Branch {
	return gitquery.Branch{
		Name:          "feat",
		IsWorktree:    true,
		Dirty:         true,
		WorktreePaths: []string{"/dev/alpha/feat"},
	}
}

func modelWithWorktreeBranch() model.Model {
	m := model.New(testRepos())
	m, _ = update(m, model.BranchResultMsg{
		RepoPath: "/dev/alpha",
		Branches: []gitquery.Branch{worktreeBranch()},
	})
	return m
}

func TestModel_DKeyOpensConfirmOverlay(t *testing.T) {
	m := modelWithWorktreeBranch()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if m.Overlay() != model.OverlayConfirm {
		t.Errorf("expected OverlayConfirm, got %d", m.Overlay())
	}
	if !strings.Contains(m.ConfirmPrompt(), "/dev/alpha/feat") {
		t.Errorf("expected confirm prompt to contain worktree path, got %q", m.ConfirmPrompt())
	}
}

func TestModel_DKeyOnNonWorktreeBranchOpensDeleteConfirm(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, model.BranchResultMsg{
		RepoPath: "/dev/alpha",
		Branches: []gitquery.Branch{{Name: "main"}},
	})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if m.Overlay() != model.OverlayConfirm {
		t.Errorf("expected OverlayConfirm for non-worktree branch, got %d", m.Overlay())
	}
	if !strings.Contains(m.ConfirmPrompt(), "main") {
		t.Errorf("expected confirm prompt to contain branch name, got %q", m.ConfirmPrompt())
	}
	if !strings.Contains(m.ConfirmPrompt(), "Delete branch") {
		t.Errorf("expected 'Delete branch' in prompt, got %q", m.ConfirmPrompt())
	}
}

func TestModel_DKeyNoOpWithNoBranches(t *testing.T) {
	m := model.New(testRepos())
	// No branches loaded
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected OverlayNone when no branches, got %d", m.Overlay())
	}
}

func TestModel_ConfirmCancelEsc(t *testing.T) {
	m := modelWithWorktreeBranch()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEscape})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed on esc, got %d", m.Overlay())
	}
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Error("esc in confirm dialog should not quit")
		}
	}
}

func TestModel_ConfirmCancelQ(t *testing.T) {
	m := modelWithWorktreeBranch()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed on q, got %d", m.Overlay())
	}
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Error("q in confirm dialog should not quit")
		}
	}
}

func TestModel_ConfirmCancelN(t *testing.T) {
	m := modelWithWorktreeBranch()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed on n, got %d", m.Overlay())
	}
}

func TestModel_ConfirmYClosesOverlayAndReturnsCmd(t *testing.T) {
	m := modelWithWorktreeBranch()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed after confirm, got %d", m.Overlay())
	}
	if cmd == nil {
		t.Fatal("expected action cmd after confirm, got nil")
	}
}

func TestModel_ConfirmEnterExecutesAction(t *testing.T) {
	m := modelWithWorktreeBranch()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed after enter, got %d", m.Overlay())
	}
	if cmd == nil {
		t.Fatal("expected action cmd after enter confirm, got nil")
	}
}

func TestModel_WorktreeRemoveFailReturnsDeleteFailedMsg(t *testing.T) {
	// With a fake path, RemoveWorktree will fail → returns DeleteFailedMsg
	m := modelWithWorktreeBranch()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.DeleteFailedMsg); !ok {
		t.Fatalf("expected DeleteFailedMsg on fake-path failure, got %T", msg)
	}
}

func TestModel_BranchDeleteFailReturnsDeleteFailedMsg(t *testing.T) {
	// With a fake repo path, DeleteBranch will fail → returns DeleteFailedMsg
	m := model.New(testRepos())
	m, _ = update(m, model.BranchResultMsg{
		RepoPath: "/dev/alpha",
		Branches: []gitquery.Branch{{Name: "feat"}},
	})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.DeleteFailedMsg); !ok {
		t.Fatalf("expected DeleteFailedMsg on fake-path failure, got %T", msg)
	}
}

func TestModel_DeleteFailedMsgOpensForceConfirm(t *testing.T) {
	m := model.New(testRepos())
	forceActionCalled := false
	m, _ = update(m, model.DeleteFailedMsg{
		RepoPath: "/dev/alpha",
		Target:   "/dev/alpha/feat",
		ForceAction: func() error {
			forceActionCalled = true
			return nil
		},
		IsWorktree: true,
	})
	if m.Overlay() != model.OverlayConfirm {
		t.Errorf("expected OverlayConfirm after DeleteFailedMsg, got %d", m.Overlay())
	}
	if !m.ConfirmForce() {
		t.Error("expected ConfirmForce=true after DeleteFailedMsg")
	}
	if !strings.Contains(m.ConfirmPrompt(), "Force delete") {
		t.Errorf("expected 'Force delete' in prompt, got %q", m.ConfirmPrompt())
	}
	if !strings.Contains(m.ConfirmPrompt(), "/dev/alpha/feat") {
		t.Errorf("expected target in prompt, got %q", m.ConfirmPrompt())
	}
	_ = forceActionCalled
}

func TestModel_ForceConfirmCancelClearsForce(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, model.DeleteFailedMsg{
		RepoPath:    "/dev/alpha",
		Target:      "feat",
		ForceAction: func() error { return nil },
		IsWorktree:  false,
	})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed after cancel, got %d", m.Overlay())
	}
	if m.ConfirmForce() {
		t.Error("expected ConfirmForce cleared after cancel")
	}
}

func TestModel_ForceConfirmYExecutesForceAction(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, model.DeleteFailedMsg{
		RepoPath:    "/dev/alpha",
		Target:      "feat",
		ForceAction: func() error { return nil },
		IsWorktree:  false,
	})
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed after force confirm, got %d", m.Overlay())
	}
	if m.ConfirmForce() {
		t.Error("expected ConfirmForce cleared after confirm")
	}
	if cmd == nil {
		t.Fatal("expected cmd from force action, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.BranchDeletedMsg); !ok {
		t.Fatalf("expected BranchDeletedMsg from force action (IsWorktree=false), got %T", msg)
	}
}

func TestModel_ForceConfirmWorktreeYReturnsWorktreeRemovedMsg(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, model.DeleteFailedMsg{
		RepoPath:    "/dev/alpha",
		Target:      "/dev/alpha/feat",
		ForceAction: func() error { return nil },
		IsWorktree:  true,
	})
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.WorktreeRemovedMsg); !ok {
		t.Fatalf("expected WorktreeRemovedMsg from force worktree action, got %T", msg)
	}
}

func TestModel_BranchDeletedMsgTriggersFetch(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, model.BranchDeletedMsg{RepoPath: "/dev/alpha"})
	if cmd == nil {
		t.Fatal("expected fetchBranches cmd after BranchDeletedMsg, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.BranchResultMsg); !ok {
		t.Errorf("expected BranchResultMsg, got %T", msg)
	}
}

func TestModel_StaleBranchDeletedMsgIgnored(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // select bravo
	_, cmd := update(m, model.BranchDeletedMsg{RepoPath: "/dev/alpha"})
	if cmd != nil {
		t.Error("expected stale BranchDeletedMsg to be ignored")
	}
}

func TestModel_WorktreeRemovedMsgTriggersBranchFetch(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, model.WorktreeRemovedMsg{RepoPath: "/dev/alpha"})
	if cmd == nil {
		t.Fatal("expected fetchBranches cmd after WorktreeRemovedMsg, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.BranchResultMsg); !ok {
		t.Errorf("expected BranchResultMsg, got %T", msg)
	}
}

func TestModel_StaleWorktreeRemovedMsgIgnored(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // select bravo
	_, cmd := update(m, model.WorktreeRemovedMsg{RepoPath: "/dev/alpha"})
	if cmd != nil {
		t.Error("expected stale WorktreeRemovedMsg to be ignored (no fetch cmd)")
	}
}

func TestModel_RKeyRefreshesBranchesInMode1(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected fetch cmd on r, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.BranchResultMsg); !ok {
		t.Errorf("expected BranchResultMsg on r in mode 1, got %T", msg)
	}
}

func TestModel_RKeyRefreshesStashesInMode2(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected fetch cmd on r in mode 2, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.StashResultMsg); !ok {
		t.Errorf("expected StashResultMsg on r in mode 2, got %T", msg)
	}
}

func TestModel_ConfirmDialogBlocksModeSwitch(t *testing.T) {
	m := modelWithWorktreeBranch()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if m.Mode() != model.ModeBranches {
		t.Errorf("confirm dialog should block mode switch, mode changed to %d", m.Mode())
	}
}

func TestModel_BranchScrollFollowsCursor(t *testing.T) {
	// Create 10 branches, terminal height only shows 3
	branches := make([]gitquery.Branch, 10)
	for i := range branches {
		branches[i] = gitquery.Branch{Name: fmt.Sprintf("branch-%d", i)}
	}
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 4}) // 3 content lines + status bar
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})

	// Cursor starts at 0, scroll at 0
	if m.BranchScroll() != 0 {
		t.Errorf("expected scroll 0 at start, got %d", m.BranchScroll())
	}

	// Move cursor down past the viewport
	for i := 0; i < 9; i++ {
		m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.BranchSelected() != 9 {
		t.Errorf("expected cursor at 9, got %d", m.BranchSelected())
	}
	// Scroll should have advanced to show cursor
	if m.BranchScroll() == 0 {
		t.Error("expected scroll to advance when cursor moves past viewport")
	}
	// Cursor must be within [scroll, scroll+contentHeight)
	contentHeight := 3
	if m.BranchSelected() < m.BranchScroll() || m.BranchSelected() >= m.BranchScroll()+contentHeight {
		t.Errorf("cursor %d not in scroll viewport [%d, %d)", m.BranchSelected(), m.BranchScroll(), m.BranchScroll()+contentHeight)
	}

	// Move back up to 0
	for i := 0; i < 9; i++ {
		m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	}
	if m.BranchSelected() != 0 {
		t.Errorf("expected cursor back at 0, got %d", m.BranchSelected())
	}
	if m.BranchScroll() != 0 {
		t.Errorf("expected scroll back to 0, got %d", m.BranchScroll())
	}
}

func TestModel_MultiWorktreeExpandsToRows(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// Branch with 2 worktree paths — should expand to 2 navigable rows
	branches := []gitquery.Branch{
		{Name: "feat", IsWorktree: true, WorktreePaths: []string{"/dev/feat-A", "/dev/feat-B"}},
	}
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})
	if m.BranchSelected() != 0 {
		t.Errorf("expected cursor at 0 initially, got %d", m.BranchSelected())
	}
	// Down should move to row 1 (expansion row)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.BranchSelected() != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.BranchSelected())
	}
	// Another down should clamp at 1 (only 2 rows total)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.BranchSelected() != 1 {
		t.Errorf("expected cursor clamped at 1, got %d", m.BranchSelected())
	}
}

func TestModel_DKeyOnExpansionRowTargetsSpecificPath(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	branches := []gitquery.Branch{
		{Name: "feat", IsWorktree: true, WorktreePaths: []string{"/dev/feat-A", "/dev/feat-B"}},
	}
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})
	// Navigate to expansion row (index 1 = /dev/feat-B)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.BranchSelected() != 1 {
		t.Fatalf("expected cursor at row 1, got %d", m.BranchSelected())
	}
	// Press d — should prompt for /dev/feat-B, not /dev/feat-A
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	prompt := m.ConfirmPrompt()
	if !strings.Contains(prompt, "/dev/feat-B") {
		t.Errorf("d on expansion row should target /dev/feat-B, got prompt: %q", prompt)
	}
	if strings.Contains(prompt, "/dev/feat-A") {
		t.Errorf("d on expansion row should not mention /dev/feat-A, got prompt: %q", prompt)
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
