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

// --- Messages ---

type BranchResultMsg struct {
	RepoPath string
	Branches []gitquery.Branch
}

type StashResultMsg struct {
	RepoPath string
	Stashes  []gitquery.Stash
}

type StashDiffResultMsg struct {
	RepoPath string
	Index    int
	Diff     string
}

type BranchDiffResultMsg struct {
	RepoPath   string
	BranchName string
	Diff       string
}

type WorktreeRemovedMsg struct {
	RepoPath string
}

type BranchDeletedMsg struct {
	RepoPath string
}

type StashDroppedMsg struct {
	RepoPath string
}

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
	activePane     int // 0=left (repos), 1=right (content)
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
func (m Model) ActivePane() int            { return m.activePane }

func (m Model) Init() tea.Cmd {
	return m.fetchBranches()
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
		ActivePane:     m.activePane,
	})
}

// --- Update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case BranchResultMsg:
		return m.handleBranchResult(msg), nil
	case StashResultMsg:
		return m.handleStashResult(msg), nil
	case StashDiffResultMsg:
		return m.handleStashDiffResult(msg), nil
	case BranchDiffResultMsg:
		return m.handleBranchDiffResult(msg), nil
	case StashDroppedMsg:
		return m.handleStashDropped(msg)
	case WorktreeRemovedMsg:
		return m.handleWorktreeRemoved(msg)
	case BranchDeletedMsg:
		return m.handleBranchDeleted(msg)
	case DeleteFailedMsg:
		return m.handleDeleteFailed(msg), nil
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.overlay == OverlayConfirm {
		return m.handleConfirmKey(key)
	}
	if m.activePane == 0 {
		return m.handleLeftPaneKey(key)
	}
	if m.overlay != OverlayNone {
		return m.handleOverlayKey(key)
	}
	return m.handleRightPaneKey(key)
}

// --- Key handlers by context ---

func (m Model) handleConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "enter":
		action := m.confirmAction
		m = m.clearConfirm()
		return m, action()
	case "n", "q", "esc":
		m = m.clearConfirm()
	}
	return m, nil
}

func (m Model) handleLeftPaneKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "tab":
		m.activePane = 1
	case "up", "k":
		if len(m.repos) > 0 {
			if m.selected > 0 {
				m.selected--
			} else {
				m.selected = len(m.repos) - 1
			}
			m = m.resetRightPaneCursors()
			return m, m.fetchForMode()
		}
	case "down", "j":
		if len(m.repos) > 0 {
			if m.selected < len(m.repos)-1 {
				m.selected++
			} else {
				m.selected = 0
			}
			m = m.resetRightPaneCursors()
			return m, m.fetchForMode()
		}
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleOverlayKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "esc":
		m.overlay = OverlayNone
		m.overlayDiff = ""
		m.overlayScroll = 0
	case "up", "k":
		if m.overlayScroll > 0 {
			m.overlayScroll--
		}
	case "down", "j":
		m.overlayScroll++
	}
	return m, nil
}

func (m Model) handleRightPaneKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		return m.handleCursorUp()
	case "down", "j":
		return m.handleCursorDown()
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
		m.activePane = 0
	case "enter":
		return m.handleEnter()
	case "d":
		return m.handleDelete()
	case "t":
		return m.handleOpenTerminal()
	case "c":
		return m.handleOpenCode()
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

// --- Cursor navigation ---

func (m Model) handleCursorUp() (tea.Model, tea.Cmd) {
	if m.mode == ModeBranches && len(m.rows) > 0 {
		if m.branchSelected > 0 {
			m.branchSelected--
		} else {
			m.branchSelected = len(m.rows) - 1
		}
		m = m.ensureBranchVisible()
		return m, nil
	}
	if m.mode == ModeStashes && len(m.stashes) > 0 {
		if m.stashSelected > 0 {
			m.stashSelected--
		} else {
			m.stashSelected = len(m.stashes) - 1
		}
	}
	return m, nil
}

func (m Model) handleCursorDown() (tea.Model, tea.Cmd) {
	if m.mode == ModeBranches && len(m.rows) > 0 {
		if m.branchSelected < len(m.rows)-1 {
			m.branchSelected++
		} else {
			m.branchSelected = 0
		}
		m = m.ensureBranchVisible()
		return m, nil
	}
	if m.mode == ModeStashes && len(m.stashes) > 0 {
		if m.stashSelected < len(m.stashes)-1 {
			m.stashSelected++
		} else {
			m.stashSelected = 0
		}
	}
	return m, nil
}

// --- Action handlers ---

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.mode == ModeBranches && m.isSelectedBranchDirtyWorktree() {
		m.overlay = OverlayBranchDiff
		return m, m.fetchBranchDiff()
	}
	if m.mode == ModeStashes && len(m.stashes) > 0 {
		m.overlay = OverlayStashDiff
		return m, m.fetchStashDiff()
	}
	return m, nil
}

func (m Model) handleDelete() (tea.Model, tea.Cmd) {
	if m.mode == ModeStashes && len(m.stashes) > 0 && len(m.repos) > 0 {
		return m.confirmStashDrop()
	}
	if m.mode == ModeBranches && len(m.repos) > 0 {
		return m.confirmBranchDelete()
	}
	return m, nil
}

func (m Model) handleOpenTerminal() (tea.Model, tea.Cmd) {
	if m.mode == ModeBranches && len(m.repos) > 0 {
		if row, ok := m.selectedRow(); ok && row.WorktreePath != "" {
			path := row.WorktreePath
			return m, func() tea.Msg { _ = actions.OpenTerminal(path); return nil }
		}
	}
	return m, nil
}

func (m Model) handleOpenCode() (tea.Model, tea.Cmd) {
	if m.mode == ModeBranches && len(m.repos) > 0 {
		if row, ok := m.selectedRow(); ok && row.WorktreePath != "" {
			path := row.WorktreePath
			return m, func() tea.Msg { _ = actions.OpenVSCode(path); return nil }
		}
	}
	return m, nil
}

// --- Confirm dialogs ---

func (m Model) confirmStashDrop() (tea.Model, tea.Cmd) {
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
	return m, nil
}

func (m Model) confirmBranchDelete() (tea.Model, tea.Cmd) {
	row, ok := m.selectedRow()
	if !ok {
		return m, nil
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
	return m, nil
}

func (m Model) clearConfirm() Model {
	m.overlay = OverlayNone
	m.confirmPrompt = ""
	m.confirmAction = nil
	m.confirmForce = false
	return m
}

func (m Model) resetRightPaneCursors() Model {
	m.branchSelected = 0
	m.stashSelected = 0
	m.branchScroll = 0
	m.rows = nil
	m.stashes = nil
	return m
}

// --- Message handlers ---

func (m Model) isCurrentRepo(repoPath string) bool {
	return m.selected < len(m.repos) && m.repos[m.selected].Path == repoPath
}

func (m Model) handleBranchResult(msg BranchResultMsg) Model {
	if m.isCurrentRepo(msg.RepoPath) {
		m.rows = gitquery.FlattenBranches(msg.Branches)
		if len(m.rows) == 0 || m.branchSelected >= len(m.rows) {
			m.branchSelected = 0
		}
	}
	return m
}

func (m Model) handleStashResult(msg StashResultMsg) Model {
	if m.isCurrentRepo(msg.RepoPath) {
		m.stashes = msg.Stashes
	}
	return m
}

func (m Model) handleStashDiffResult(msg StashDiffResultMsg) Model {
	if m.isCurrentRepo(msg.RepoPath) {
		m.overlayDiff = msg.Diff
	}
	return m
}

func (m Model) handleBranchDiffResult(msg BranchDiffResultMsg) Model {
	if m.isCurrentRepo(msg.RepoPath) {
		if row, ok := m.selectedRow(); ok && row.Branch.Name == msg.BranchName {
			m.overlayDiff = msg.Diff
		}
	}
	return m
}

func (m Model) handleStashDropped(msg StashDroppedMsg) (tea.Model, tea.Cmd) {
	if m.isCurrentRepo(msg.RepoPath) {
		if m.stashSelected >= len(m.stashes)-1 && m.stashSelected > 0 {
			m.stashSelected--
		}
		return m, m.fetchStashes()
	}
	return m, nil
}

func (m Model) handleWorktreeRemoved(msg WorktreeRemovedMsg) (tea.Model, tea.Cmd) {
	if m.isCurrentRepo(msg.RepoPath) {
		return m, m.fetchBranches()
	}
	return m, nil
}

func (m Model) handleBranchDeleted(msg BranchDeletedMsg) (tea.Model, tea.Cmd) {
	if m.isCurrentRepo(msg.RepoPath) {
		return m, m.fetchBranches()
	}
	return m, nil
}

func (m Model) handleDeleteFailed(msg DeleteFailedMsg) Model {
	if m.isCurrentRepo(msg.RepoPath) {
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
	return m
}

// --- Fetch commands ---

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

// --- Helpers ---

func (m Model) selectedRow() (gitquery.BranchRow, bool) {
	if m.branchSelected < 0 || m.branchSelected >= len(m.rows) {
		return gitquery.BranchRow{}, false
	}
	return m.rows[m.branchSelected], true
}

func (m Model) isSelectedBranchDirtyWorktree() bool {
	row, ok := m.selectedRow()
	return ok && row.Branch.Dirty && row.Branch.IsWorktree
}

func (m Model) ensureBranchVisible() Model {
	contentHeight := m.height - ui.BranchContentOverhead
	if contentHeight <= 0 {
		contentHeight = 16
	}
	line := 0
	for i, row := range m.rows {
		if i == m.branchSelected {
			break
		}
		line++
		if !row.IsExpansion {
			n := len(row.Branch.Unpushed)
			if n > 5 {
				line += 6
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
