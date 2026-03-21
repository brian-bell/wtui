package model

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brian-bell/wt/actions"
	"github.com/brian-bell/wt/gitquery"
	"github.com/brian-bell/wt/scanner"
	"github.com/brian-bell/wt/ui"
)

// Mode represents the active right-pane view.
type Mode int

const (
	ModeBranches Mode = iota + 1
	ModeStashes
)

// OverlayState represents what overlay (if any) is displayed.
type OverlayState int

const (
	OverlayNone OverlayState = iota
	OverlayStashDiff
	OverlayBranchDiff
	OverlayConfirm
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

// BranchDiffResultMsg is sent when a branch diff is fetched asynchronously.
type BranchDiffResultMsg struct {
	RepoPath   string
	BranchName string
	Diff       string
}

// WorktreeRemovedMsg is sent when a worktree removal completes.
type WorktreeRemovedMsg struct {
	RepoPath string
}

// BranchDeletedMsg is sent when a branch deletion completes.
type BranchDeletedMsg struct {
	RepoPath string
}

// StashDroppedMsg is sent when a stash drop completes.
type StashDroppedMsg struct {
	RepoPath string
}

// DeleteFailedMsg is sent when a delete operation fails and a force retry is available.
type DeleteFailedMsg struct {
	RepoPath    string
	Target      string       // worktree path or branch name
	ForceAction func() error // the --force variant to call
	IsWorktree  bool
}

// Model is the bubbletea application model.
type Model struct {
	repos          []scanner.Repo
	selected       int
	width          int
	height         int
	mode           Mode
	rows           []gitquery.BranchRow
	stashes        []gitquery.Stash
	branchSelected int
	stashSelected  int
	overlay        OverlayState
	overlayDiff    string
	overlayScroll  int
	confirmPrompt  string
	confirmAction  func() tea.Cmd
	confirmForce   bool
	branchScroll   int
}

// New creates a Model from discovered repos.
func New(repos []scanner.Repo) Model {
	return Model{repos: repos, mode: ModeBranches}
}

func (m Model) Selected() int              { return m.selected }
func (m Model) Width() int                 { return m.width }
func (m Model) Height() int                { return m.height }
func (m Model) Mode() Mode                 { return m.mode }
func (m Model) Rows() []gitquery.BranchRow { return m.rows }
func (m Model) Stashes() []gitquery.Stash  { return m.stashes }
func (m Model) BranchSelected() int        { return m.branchSelected }
func (m Model) StashSelected() int         { return m.stashSelected }
func (m Model) Overlay() OverlayState      { return m.overlay }
func (m Model) OverlayDiff() string        { return m.overlayDiff }
func (m Model) OverlayScroll() int         { return m.overlayScroll }
func (m Model) ConfirmPrompt() string      { return m.confirmPrompt }
func (m Model) ConfirmForce() bool         { return m.confirmForce }
func (m Model) BranchScroll() int          { return m.branchScroll }

func (m Model) Init() tea.Cmd {
	return m.fetchBranches()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case BranchResultMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			m.rows = gitquery.FlattenBranches(msg.Branches)
			if len(m.rows) == 0 || m.branchSelected >= len(m.rows) {
				m.branchSelected = 0
			}
		}
	case StashResultMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			m.stashes = msg.Stashes
		}
	case StashDiffResultMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			m.overlayDiff = msg.Diff
		}
	case BranchDiffResultMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			if row, ok := m.selectedRow(); ok && row.Branch.Name == msg.BranchName {
				m.overlayDiff = msg.Diff
			}
		}
	case StashDroppedMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			if m.stashSelected >= len(m.stashes)-1 && m.stashSelected > 0 {
				m.stashSelected--
			}
			return m, m.fetchStashes()
		}
	case WorktreeRemovedMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			return m, m.fetchBranches()
		}
	case BranchDeletedMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			return m, m.fetchBranches()
		}
	case DeleteFailedMsg:
		if m.selected < len(m.repos) && msg.RepoPath == m.repos[m.selected].Path {
			m.confirmPrompt = fmt.Sprintf("Force delete %s? (y/n)", msg.Target)
			m.confirmForce = true
			m.overlay = OverlayConfirm
			m.confirmAction = func() tea.Cmd {
				return func() tea.Msg {
					_ = msg.ForceAction()
					if msg.IsWorktree {
						return WorktreeRemovedMsg{RepoPath: msg.RepoPath}
					}
					return BranchDeletedMsg{RepoPath: msg.RepoPath}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Confirmation dialog — only allow confirm/cancel
	if m.overlay == OverlayConfirm {
		switch key {
		case "y", "enter":
			action := m.confirmAction
			m.overlay = OverlayNone
			m.confirmPrompt = ""
			m.confirmAction = nil
			m.confirmForce = false
			return m, action()
		case "n", "q", "esc":
			m.overlay = OverlayNone
			m.confirmPrompt = ""
			m.confirmAction = nil
			m.confirmForce = false
		}
		return m, nil
	}

	// Diff overlay — only allow scroll/close
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
		if m.mode == ModeBranches {
			if len(m.rows) > 0 && m.branchSelected > 0 {
				m.branchSelected--
				m = m.ensureBranchVisible()
				return m, nil
			}
		}
		if m.mode == ModeStashes && m.stashSelected > 0 {
			m.stashSelected--
		}
	case "down", "j":
		if m.mode == ModeBranches {
			if len(m.rows) > 0 && m.branchSelected < len(m.rows)-1 {
				m.branchSelected++
				m = m.ensureBranchVisible()
				return m, nil
			}
		}
		if m.mode == ModeStashes && len(m.stashes) > 0 && m.stashSelected < len(m.stashes)-1 {
			m.stashSelected++
		}
	case "left", "h":
		if m.mode > ModeBranches {
			m.mode--
			m.branchSelected = 0
			m.stashSelected = 0
			return m, m.fetchForMode()
		}
	case "right", "l":
		if m.mode < ModeStashes {
			m.mode++
			m.branchSelected = 0
			m.stashSelected = 0
			return m, m.fetchForMode()
		}
	case "1":
		if m.mode != ModeBranches {
			m.mode = ModeBranches
			m.branchSelected = 0
			return m, m.fetchBranches()
		}
	case "2":
		if m.mode != ModeStashes {
			m.mode = ModeStashes
			m.branchSelected = 0
			m.stashSelected = 0
			return m, m.fetchStashes()
		}
	case "tab":
		if len(m.repos) > 0 {
			m.selected = (m.selected + 1) % len(m.repos)
			m.branchSelected = 0
			m.stashSelected = 0
			return m, m.fetchForMode()
		}
	case "enter":
		if m.mode == ModeBranches && m.isSelectedBranchDirtyWorktree() {
			m.overlay = OverlayBranchDiff
			return m, m.fetchBranchDiff()
		}
		if m.mode == ModeStashes && len(m.stashes) > 0 {
			m.overlay = OverlayStashDiff
			return m, m.fetchStashDiff()
		}
	case "d":
		if m.mode == ModeStashes && len(m.stashes) > 0 && len(m.repos) > 0 {
			idx := m.stashes[m.stashSelected].Index
			repoPath := m.repos[m.selected].Path
			m.confirmPrompt = fmt.Sprintf("Drop stash@{%d}? (y/n)", idx)
			m.confirmAction = func() tea.Cmd {
				return func() tea.Msg {
					_ = actions.DropStash(repoPath, idx)
					return StashDroppedMsg{RepoPath: repoPath}
				}
			}
			m.overlay = OverlayConfirm
		}
		if m.mode == ModeBranches && len(m.repos) > 0 {
			row, ok := m.selectedRow()
			if !ok {
				break
			}
			repoPath := m.repos[m.selected].Path
			if row.WorktreePath != "" {
				worktreePath := row.WorktreePath
				m.confirmPrompt = fmt.Sprintf("Remove worktree %s? (y/n)", worktreePath)
				m.confirmAction = func() tea.Cmd {
					return func() tea.Msg {
						if err := actions.RemoveWorktree(repoPath, worktreePath); err != nil {
							return DeleteFailedMsg{
								RepoPath:    repoPath,
								Target:      worktreePath,
								ForceAction: func() error { return actions.ForceRemoveWorktree(repoPath, worktreePath) },
								IsWorktree:  true,
							}
						}
						return WorktreeRemovedMsg{RepoPath: repoPath}
					}
				}
			} else {
				branchName := row.Branch.Name
				m.confirmPrompt = fmt.Sprintf("Delete branch %s? (y/n)", branchName)
				m.confirmAction = func() tea.Cmd {
					return func() tea.Msg {
						if err := actions.DeleteBranch(repoPath, branchName); err != nil {
							return DeleteFailedMsg{
								RepoPath:    repoPath,
								Target:      branchName,
								ForceAction: func() error { return actions.ForceDeleteBranch(repoPath, branchName) },
							}
						}
						return BranchDeletedMsg{RepoPath: repoPath}
					}
				}
			}
			m.overlay = OverlayConfirm
		}
	case "r":
		return m, m.fetchForMode()
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	return ui.Render(ui.RenderParams{
		Repos:          m.repos,
		Selected:       m.selected,
		Width:          m.width,
		Height:         m.height,
		Mode:           int(m.mode),
		Branches:       m.rows,
		Stashes:        m.stashes,
		BranchSelected: m.branchSelected,
		StashSelected:  m.stashSelected,
		Overlay:        int(m.overlay),
		OverlayDiff:    m.overlayDiff,
		OverlayScroll:  m.overlayScroll,
		ConfirmPrompt:  m.confirmPrompt,
		ConfirmForce:   m.confirmForce,
		BranchScroll:   m.branchScroll,
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

func (m Model) fetchBranchDiff() tea.Cmd {
	if len(m.repos) == 0 {
		return nil
	}
	row, ok := m.selectedRow()
	if !ok || !row.Branch.Dirty || !row.Branch.IsWorktree {
		return nil
	}

	repoPath := m.repos[m.selected].Path
	worktreePath := row.WorktreePath
	if worktreePath == "" {
		worktreePath = repoPath
	}
	branchName := row.Branch.Name

	return func() tea.Msg {
		diff, _ := gitquery.BranchDiff(worktreePath)
		return BranchDiffResultMsg{
			RepoPath:   repoPath,
			BranchName: branchName,
			Diff:       diff,
		}
	}
}

func (m Model) selectedRow() (gitquery.BranchRow, bool) {
	if m.branchSelected < 0 || m.branchSelected >= len(m.rows) {
		return gitquery.BranchRow{}, false
	}
	return m.rows[m.branchSelected], true
}

func (m Model) ensureBranchVisible() Model {
	contentHeight := m.height - 1
	if contentHeight <= 0 {
		contentHeight = 20 // fallback when height not yet set
	}
	// Calculate the content line index for the selected row
	line := 0
	for i, row := range m.rows {
		if i == m.branchSelected {
			break
		}
		line++
		if !row.IsExpansion {
			n := len(row.Branch.Unpushed)
			if n > 5 {
				line += 6 // 5 + overflow line
			} else {
				line += n
			}
		}
	}
	if m.branchScroll > line {
		m.branchScroll = line
	}
	if line >= m.branchScroll+contentHeight {
		m.branchScroll = line - contentHeight + 1
	}
	return m
}

func (m Model) isSelectedBranchDirtyWorktree() bool {
	row, ok := m.selectedRow()
	return ok && row.Branch.Dirty && row.Branch.IsWorktree
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
