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
	ModeBranches Mode = 1
	ModeStashes  Mode = 2
)

// OverlayState represents what overlay (if any) is displayed.
type OverlayState int

const (
	OverlayNone      OverlayState = 0
	OverlayStashDiff OverlayState = 1
)

// BranchResultMsg is sent when branch data is fetched asynchronously.
type BranchResultMsg struct {
	RepoPath string
	Branches []gitquery.Branch
}

// StashResultMsg is sent when stash data is fetched asynchronously.
type StashResultMsg struct {
	RepoPath string
	Stashes  []gitquery.Stash
}

// StashDiffResultMsg is sent when a stash diff is fetched asynchronously.
type StashDiffResultMsg struct {
	RepoPath string
	Index    int
	Diff     string
}

// Model is the bubbletea application model.
type Model struct {
	repos         []scanner.Repo
	selected      int
	width         int
	height        int
	mode          Mode
	branches      []gitquery.Branch
	stashes       []gitquery.Stash
	stashSelected int
	overlay       OverlayState
	overlayDiff   string
	overlayScroll int
}

// New creates a Model from discovered repos.
func New(repos []scanner.Repo) Model {
	return Model{repos: repos, mode: ModeBranches}
}

func (m Model) Selected() int               { return m.selected }
func (m Model) Width() int                  { return m.width }
func (m Model) Height() int                 { return m.height }
func (m Model) Mode() Mode                  { return m.mode }
func (m Model) Branches() []gitquery.Branch { return m.branches }
func (m Model) Stashes() []gitquery.Stash   { return m.stashes }
func (m Model) StashSelected() int          { return m.stashSelected }
func (m Model) Overlay() OverlayState       { return m.overlay }
func (m Model) OverlayDiff() string         { return m.overlayDiff }
func (m Model) OverlayScroll() int          { return m.overlayScroll }

func (m Model) Init() tea.Cmd {
	return m.fetchBranches()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case BranchResultMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			m.branches = msg.Branches
		}
	case StashResultMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			m.stashes = msg.Stashes
		}
	case StashDiffResultMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			m.overlayDiff = msg.Diff
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Overlay is open — only allow overlay controls
	if m.overlay != OverlayNone {
		switch key {
		case "q", "esc":
			m.overlay = OverlayNone
			m.overlayDiff = ""
			m.overlayScroll = 0
			return m, nil
		case "up", "k":
			if m.overlayScroll > 0 {
				m.overlayScroll--
			}
			return m, nil
		case "down", "j":
			m.overlayScroll++
			return m, nil
		}
		return m, nil
	}

	switch key {
	case "up", "k":
		if m.mode == ModeStashes && m.stashSelected > 0 {
			m.stashSelected--
		}
	case "down", "j":
		if m.mode == ModeStashes && len(m.stashes) > 0 && m.stashSelected < len(m.stashes)-1 {
			m.stashSelected++
		}
	case "left", "h":
		if m.mode > ModeBranches {
			m.mode--
			m.stashSelected = 0
			return m, m.fetchForMode()
		}
	case "right", "l":
		if m.mode < ModeStashes {
			m.mode++
			m.stashSelected = 0
			return m, m.fetchForMode()
		}
	case "1":
		if m.mode != ModeBranches {
			m.mode = ModeBranches
			return m, m.fetchBranches()
		}
	case "2":
		if m.mode != ModeStashes {
			m.mode = ModeStashes
			m.stashSelected = 0
			return m, m.fetchStashes()
		}
	case "tab":
		if len(m.repos) > 0 {
			m.selected = (m.selected + 1) % len(m.repos)
			m.stashSelected = 0
			return m, m.fetchForMode()
		}
	case "enter":
		if m.mode == ModeStashes && len(m.stashes) > 0 {
			m.overlay = OverlayStashDiff
			return m, m.fetchStashDiff()
		}
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	return ui.Render(ui.RenderParams{
		Repos:         m.repos,
		Selected:      m.selected,
		Width:         m.width,
		Height:        m.height,
		Mode:          int(m.mode),
		Branches:      m.branches,
		Stashes:       m.stashes,
		StashSelected: m.stashSelected,
		Overlay:       int(m.overlay),
		OverlayDiff:   m.overlayDiff,
		OverlayScroll: m.overlayScroll,
	})
}

func (m Model) fetchForMode() tea.Cmd {
	switch m.mode {
	case ModeBranches:
		return m.fetchBranches()
	case ModeStashes:
		return m.fetchStashes()
	}
	return nil
}

func (m Model) fetchBranches() tea.Cmd {
	if len(m.repos) == 0 {
		return nil
	}
	repoPath := m.repos[m.selected].Path
	return func() tea.Msg {
		branches, _ := gitquery.ListBranches(repoPath)
		return BranchResultMsg{RepoPath: repoPath, Branches: branches}
	}
}

func (m Model) fetchStashes() tea.Cmd {
	if len(m.repos) == 0 {
		return nil
	}
	repoPath := m.repos[m.selected].Path
	return func() tea.Msg {
		stashes, _ := gitquery.ListStashes(repoPath)
		return StashResultMsg{RepoPath: repoPath, Stashes: stashes}
	}
}

func (m Model) fetchStashDiff() tea.Cmd {
	if len(m.repos) == 0 || len(m.stashes) == 0 {
		return nil
	}
	repoPath := m.repos[m.selected].Path
	index := m.stashes[m.stashSelected].Index
	return func() tea.Msg {
		diff, _ := gitquery.StashDiff(repoPath, index)
		return StashDiffResultMsg{RepoPath: repoPath, Index: index, Diff: diff}
	}
}
