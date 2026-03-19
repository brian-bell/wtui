package model

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/brian-bell/wt/scanner"
	"github.com/brian-bell/wt/ui"
)

// Model is the bubbletea application model.
type Model struct {
	repos    []scanner.Repo
	selected int
	width    int
	height   int
}

// New creates a Model from discovered repos.
func New(repos []scanner.Repo) Model {
	return Model{repos: repos}
}

func (m Model) Selected() int { return m.selected }
func (m Model) Width() int    { return m.width }
func (m Model) Height() int   { return m.height }

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.repos)-1 {
				m.selected++
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	return ui.Render(ui.RenderParams{
		Repos:    m.repos,
		Selected: m.selected,
		Width:    m.width,
		Height:   m.height,
	})
}
