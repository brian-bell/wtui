package model_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

func TestModel_DownArrowMovesSelection(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.Selected() != 1 {
		t.Errorf("expected selected 1, got %d", m.Selected())
	}
}

func TestModel_UpArrowMovesSelection(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.Selected() != 0 {
		t.Errorf("expected selected 0, got %d", m.Selected())
	}
}

func TestModel_DownClampsAtBottom(t *testing.T) {
	m := model.New(testRepos())
	for i := 0; i < 10; i++ {
		m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.Selected() != 2 {
		t.Errorf("expected selected 2 (last), got %d", m.Selected())
	}
}

func TestModel_UpClampsAtTop(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.Selected() != 0 {
		t.Errorf("expected selected 0, got %d", m.Selected())
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

func TestModel_ViewContainsExpectedContent(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()

	for _, name := range []string{"alpha", "bravo", "charlie"} {
		if !strings.Contains(view, name) {
			t.Errorf("view should contain repo name %q", name)
		}
	}
	if !strings.Contains(view, "nothing here yet") {
		t.Error("view should contain placeholder text")
	}
	if !strings.Contains(view, "q: quit") {
		t.Error("view should contain quit keybinding")
	}
}
