package model

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/brian-bell/wt/gitquery"
	"github.com/brian-bell/wt/scanner"
	"github.com/brian-bell/wt/ui"
)

// Mode represents the active right-pane view.
type Mode int

const (
	ModeWorktrees Mode = 1
	ModeStashes   Mode = 2
	ModeBranches  Mode = 3
)

// WorktreeResultMsg is sent when worktree data is fetched asynchronously.
type WorktreeResultMsg struct {
	RepoPath  string
	Worktrees []gitquery.Worktree
}

// Model is the bubbletea application model.
type Model struct {
	repos     []scanner.Repo
	selected  int
	width     int
	height    int
	mode      Mode
	worktrees []gitquery.Worktree
}

// New creates a Model from discovered repos.
func New(repos []scanner.Repo) Model {
	return Model{repos: repos, mode: ModeWorktrees}
}

func (m Model) Selected() int                  { return m.selected }
func (m Model) Width() int                     { return m.width }
func (m Model) Height() int                    { return m.height }
func (m Model) Mode() Mode                     { return m.mode }
func (m Model) Worktrees() []gitquery.Worktree { return m.worktrees }

func (m Model) Init() tea.Cmd {
	return m.fetchWorktrees()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
				return m, m.fetchWorktrees()
			}
		case "down", "j":
			if m.selected < len(m.repos)-1 {
				m.selected++
				return m, m.fetchWorktrees()
			}
		case "1":
			m.mode = ModeWorktrees
			return m, m.fetchWorktrees()
		case "2":
			m.mode = ModeStashes
		case "3":
			m.mode = ModeBranches
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case WorktreeResultMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			m.worktrees = msg.Worktrees
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	return ui.Render(ui.RenderParams{
		Repos:     m.repos,
		Selected:  m.selected,
		Width:     m.width,
		Height:    m.height,
		Mode:      int(m.mode),
		Worktrees: m.worktrees,
	})
}

func (m Model) fetchWorktrees() tea.Cmd {
	if len(m.repos) == 0 {
		return nil
	}
	repoPath := m.repos[m.selected].Path
	return func() tea.Msg {
		wts, _ := gitquery.ListWorktrees(repoPath)
		return WorktreeResultMsg{RepoPath: repoPath, Worktrees: wts}
	}
}
