package model_test

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brian-bell/wt/gitquery"
	"github.com/brian-bell/wt/model"
	"github.com/brian-bell/wt/scanner"
	"github.com/brian-bell/wt/ui"
)

// --- Shared helpers ---

func testRepos() []scanner.Repo {
	return []scanner.Repo{
		{Path: "/dev/alpha", DisplayName: "alpha"},
		{Path: "/dev/bravo", DisplayName: "bravo"},
		{Path: "/dev/charlie", DisplayName: "charlie"},
	}
}

func testStashes() []gitquery.Stash {
	return []gitquery.Stash{
		{Index: 0, Date: "2026-03-18 10:00:00 -0700", Message: "WIP: feature A"},
		{Index: 1, Date: "2026-03-17 09:00:00 -0700", Message: "backup: old approach"},
		{Index: 2, Date: "2026-03-16 08:00:00 -0700", Message: "experiment"},
	}
}

// update sends a message and returns the concrete Model.
func update(m model.Model, msg tea.Msg) (model.Model, tea.Cmd) {
	tm, cmd := m.Update(msg)
	return tm.(model.Model), cmd
}

// inRightPane switches focus to the right pane.
func inRightPane(m model.Model) model.Model {
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	return m
}

// selectBravo navigates to repo index 1 (bravo) in the left pane.
func selectBravo(m model.Model) model.Model {
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	return m
}

// --- Init & basics ---

func TestModel_InitialActivePaneIsLeft(t *testing.T) {
	m := model.New(testRepos())
	if m.ActivePane() != 0 {
		t.Errorf("expected left pane (0) active initially, got %d", m.ActivePane())
	}
}

func TestModel_InitialSelection(t *testing.T) {
	m := model.New(testRepos())
	if m.Selected() != 0 {
		t.Errorf("expected initial selected 0, got %d", m.Selected())
	}
}

func TestModel_DefaultModeIsBranches(t *testing.T) {
	m := model.New(testRepos())
	if m.Mode() != 1 {
		t.Errorf("expected default mode 1, got %d", m.Mode())
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

func TestModel_QuitKeys(t *testing.T) {
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyCtrlC},
		{Type: tea.KeyEscape},
	} {
		m := model.New(testRepos())
		_, cmd := update(m, key)
		if cmd == nil {
			t.Fatalf("key %v: expected quit command, got nil", key)
		}
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); !ok {
			t.Errorf("key %v: expected tea.QuitMsg, got %T", key, msg)
		}
	}
}

// --- Pane switching ---

func TestModel_TabFromRightPaneSwitchesToLeft(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // left → right
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // right → left
	if m.ActivePane() != 0 {
		t.Errorf("expected left pane (0) after second tab, got %d", m.ActivePane())
	}
}

func TestModel_TabTogglesPaneFocus(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.ActivePane() != 1 {
		t.Errorf("expected right pane after tab, got %d", m.ActivePane())
	}
	// TAB does not change repo selection
	if m.Selected() != 0 {
		t.Errorf("expected selected unchanged at 0, got %d", m.Selected())
	}
}

func TestModel_TabDoesNotChangeMode(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})                       // left → right
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}) // mode 2
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})                       // right → left
	if m.Mode() != 2 {
		t.Errorf("expected mode unchanged at 2, got %d", m.Mode())
	}
}

// --- Left pane navigation ---

func TestModel_LeftPaneDownNavigatesRepos(t *testing.T) {
	m := model.New(testRepos())
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.Selected() != 1 {
		t.Errorf("expected selected 1 after down in left pane, got %d", m.Selected())
	}
	if cmd == nil {
		t.Error("expected fetch cmd after repo navigation, got nil")
	}
}

func TestModel_LeftPaneUpNavigatesRepos(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown}) // selected=1
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})   // selected=0
	if m.Selected() != 0 {
		t.Errorf("expected selected 0 after up, got %d", m.Selected())
	}
}

func TestModel_LeftPaneDownWrapsToFirst(t *testing.T) {
	m := model.New(testRepos()) // 3 repos
	// Move to last repo
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown}) // 1
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown}) // 2
	// One more should wrap to 0
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.Selected() != 0 {
		t.Errorf("expected selected to wrap to 0, got %d", m.Selected())
	}
}

func TestModel_LeftPaneUpWrapsToLast(t *testing.T) {
	m := model.New(testRepos()) // 3 repos
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.Selected() != 2 {
		t.Errorf("expected selected to wrap to 2, got %d", m.Selected())
	}
}

func TestModel_RepoSwitchClearsRightPaneData(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: []gitquery.Branch{
		{Name: "main", IsWorktree: true, WorktreePaths: []string{"/dev/alpha"}},
	}})
	if len(m.Rows()) != 1 {
		t.Fatal("expected 1 row before switching repos")
	}
	// Switch to next repo — old data should be cleared immediately
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if len(m.Rows()) != 0 {
		t.Errorf("expected rows cleared on repo switch, got %d", len(m.Rows()))
	}
	if len(m.Stashes()) != 0 {
		t.Errorf("expected stashes cleared on repo switch, got %d", len(m.Stashes()))
	}
}

func TestModel_LeftPaneDownResetsRightPaneCursors(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab}) // switch to right pane
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: []gitquery.Branch{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	}})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown}) // move branch cursor
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown}) // branchSelected=2
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})  // back to left pane
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown}) // navigate to bravo
	if m.BranchSelected() != 0 {
		t.Errorf("expected branchSelected reset to 0, got %d", m.BranchSelected())
	}
}

func TestModel_LeftPaneModeKeysAreNoOps(t *testing.T) {
	m := model.New(testRepos())
	// 2, right, l — should not change mode in left pane
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'2'}},
		{Type: tea.KeyRight},
		{Type: tea.KeyRunes, Runes: []rune{'l'}},
	} {
		m2, cmd := update(m, key)
		if m2.Mode() != 1 {
			t.Errorf("key %v changed mode in left pane: got %d", key, m2.Mode())
		}
		if cmd != nil {
			t.Errorf("key %v produced cmd in left pane: %T", key, cmd)
		}
	}
}

func TestModel_LeftPaneActionKeysAreNoOps(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: []gitquery.Branch{
		{Name: "feat", IsWorktree: true, Dirty: true, WorktreePaths: []string{"/dev/alpha"}},
	}})
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyEnter},
		{Type: tea.KeyRunes, Runes: []rune{'d'}},
		{Type: tea.KeyRunes, Runes: []rune{'t'}},
		{Type: tea.KeyRunes, Runes: []rune{'c'}},
	} {
		_, cmd := update(m, key)
		if cmd != nil {
			t.Errorf("key %v produced cmd in left pane: %T", key, cmd)
		}
	}
}

// --- Right pane navigation ---

func TestModel_RightPaneUpDownDoesNotMoveRepoSelection(t *testing.T) {
	m := model.New(testRepos())
	m = selectBravo(m) // selected=1
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.Selected() != 1 {
		t.Errorf("expected selected unchanged at 1 in right pane, got %d", m.Selected())
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
	m = inRightPane(m)

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
	// Wrap to first
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.BranchSelected() != 0 {
		t.Errorf("expected cursor to wrap to 0, got %d", m.BranchSelected())
	}
	// Wrap backward to last
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.BranchSelected() != 3 {
		t.Errorf("expected cursor to wrap to 3, got %d", m.BranchSelected())
	}
}

func TestModel_StashCursorWraps(t *testing.T) {
	m := model.New(testRepos())
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	// Wrap backward from 0 to last
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.StashSelected() != 2 {
		t.Errorf("expected StashSelected to wrap to 2, got %d", m.StashSelected())
	}
	// Wrap forward from last to 0
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.StashSelected() != 0 {
		t.Errorf("expected StashSelected to wrap to 0, got %d", m.StashSelected())
	}
}

func TestModel_BranchScrollFollowsCursor(t *testing.T) {
	// Create 10 branches, terminal height only shows 3
	branches := make([]gitquery.Branch, 10)
	for i := range branches {
		branches[i] = gitquery.Branch{Name: fmt.Sprintf("branch-%d", i)}
	}
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: ui.BranchContentOverhead + 3}) // 3 content lines
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})

	// Cursor starts at 0, scroll at 0
	if m.BranchScroll() != 0 {
		t.Errorf("expected scroll 0 at start, got %d", m.BranchScroll())
	}

	m = inRightPane(m)
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

// --- Mode switching ---

func TestModel_ModeSwitchOnKeyPress(t *testing.T) {
	m := model.New(testRepos())
	m = inRightPane(m)
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
	m = inRightPane(m)
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

func TestModel_Pressing1WhileInMode1NoFetch(t *testing.T) {
	m := model.New(testRepos())
	// Already in mode 1; pressing 1 should not fire a redundant fetch
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if cmd != nil {
		t.Error("pressing 1 while already in mode 1 should not fire fetch")
	}
}

func TestModel_ModeSwitchPreservesSelection(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})                      // select bravo (left pane)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})                       // switch to right pane
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}) // mode 2
	if m.Selected() != 1 {
		t.Errorf("expected selection preserved at 1, got %d", m.Selected())
	}
}

func TestModel_RightSwitchesToMode2(t *testing.T) {
	m := model.New(testRepos())
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.Mode() != 2 {
		t.Errorf("expected mode 2, got %d", m.Mode())
	}
}

func TestModel_LeftSwitchesToMode1(t *testing.T) {
	m := model.New(testRepos())
	m = inRightPane(m)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.Mode() != 1 {
		t.Errorf("expected mode 1, got %d", m.Mode())
	}
}

func TestModel_HLSwitchModes(t *testing.T) {
	m := model.New(testRepos())
	m = inRightPane(m)
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
	m = inRightPane(m)
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
	m = inRightPane(m)
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

func TestModel_Mode1SwitchFiresFetchBranches(t *testing.T) {
	m := model.New(testRepos())
	m = inRightPane(m)
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

func TestModel_SwitchToMode2FiresFetchStashes(t *testing.T) {
	m := model.New(testRepos())
	m = inRightPane(m)
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if cmd == nil {
		t.Fatal("expected fetchStashes cmd on switch to mode 2, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.StashResultMsg); !ok {
		t.Errorf("expected StashResultMsg, got %T", msg)
	}
}

// --- Message handlers ---

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
	m = selectBravo(m) // selected=bravo
	branches := []gitquery.Branch{{Name: "main"}}
	m, _ = update(m, model.BranchResultMsg{RepoPath: "/dev/alpha", Branches: branches})
	if len(m.Rows()) != 0 {
		t.Errorf("expected stale result discarded, got %d rows", len(m.Rows()))
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
	m = selectBravo(m) // selected=bravo
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	if len(m.Stashes()) != 0 {
		t.Errorf("expected stale stash result discarded, got %d stashes", len(m.Stashes()))
	}
}

func TestModel_StaleBranchDiffResultDiscarded(t *testing.T) {
	m := model.New(testRepos())
	m = selectBravo(m) // selected=bravo

	m, _ = update(m, model.BranchDiffResultMsg{
		RepoPath:   "/dev/alpha",
		BranchName: "feat",
		Diff:       "diff --git a/f.txt b/f.txt",
	})

	if m.OverlayDiff() != "" {
		t.Errorf("expected stale branch diff discarded, got %q", m.OverlayDiff())
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
	m = selectBravo(m) // selected=bravo
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
	m = selectBravo(m) // selected=bravo
	_, cmd := update(m, model.WorktreeRemovedMsg{RepoPath: "/dev/alpha"})
	if cmd != nil {
		t.Error("expected stale WorktreeRemovedMsg to be ignored (no fetch cmd)")
	}
}

func TestModel_StashDroppedMsgTriggersStashFetch(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, model.StashDroppedMsg{RepoPath: "/dev/alpha"})
	if cmd == nil {
		t.Fatal("expected fetchStashes cmd after StashDroppedMsg, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.StashResultMsg); !ok {
		t.Errorf("expected StashResultMsg, got %T", msg)
	}
}

func TestModel_StaleStashDroppedMsgIgnored(t *testing.T) {
	m := model.New(testRepos())
	m = selectBravo(m) // selected=bravo
	_, cmd := update(m, model.StashDroppedMsg{RepoPath: "/dev/alpha"})
	if cmd != nil {
		t.Error("expected stale StashDroppedMsg to be ignored")
	}
}
