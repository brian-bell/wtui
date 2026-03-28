package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/brian-bell/wtui/gitquery"
	"github.com/brian-bell/wtui/scanner"
)

const LeftPaneWidth = 30

// BranchContentOverhead is the number of rows consumed by chrome around the
// branch list: status bar (1) + top/bottom borders (2) + mode header with
// separator (2). Both the model (ensureBranchVisible) and the renderer use
// this constant so they stay in sync.
const BranchContentOverhead = 5

// stashEntryOverhead is the number of visible chars consumed by the prefix,
// date, and separator in a stash line: "   " (3) + date (10) + "  " (2).
const stashEntryOverhead = 15

// StashEntryHeight returns the number of visual lines a stash entry will
// occupy given its message and the pane width. Returns 1 or 2.
func StashEntryHeight(msg string, paneWidth int) int {
	msgAvail := paneWidth - stashEntryOverhead
	if msgAvail < 1 || len([]rune(msg)) <= msgAvail {
		return 1
	}
	return 2
}

var (
	repoStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	selectedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Reverse(true)
	placeholderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	statusStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	branchStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	cleanStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	commitStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	activeModeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	inactiveModeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	stashDateStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	stashMsgStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	stashSelStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true).Reverse(true)
	branchSelStyle    = lipgloss.NewStyle().Bold(true).Reverse(true)
	noUpstreamStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	aheadBehindStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	dirtyRedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	diffAddStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	diffDelStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	diffHdrStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)

// RenderParams holds everything the renderer needs.
type RenderParams struct {
	Repos          []scanner.Repo
	Selected       int
	Width          int
	Height         int
	Mode           int
	Branches       []gitquery.BranchRow
	Stashes        []gitquery.Stash
	BranchSelected int
	StashSelected  int
	Overlay        int
	OverlayDiff    string
	OverlayScroll  int
	ConfirmPrompt  string
	ConfirmForce   bool
	BranchScroll   int
	StashScroll    int
	ActivePane     int
	Destructive    bool
}

// Render produces the full terminal view string.
func Render(p RenderParams) string {
	if p.Width == 0 {
		p.Width = 80
	}
	if p.Height == 0 {
		p.Height = 24
	}

	// Overlay takes over the entire screen
	if p.Overlay != 0 {
		return renderOverlay(p)
	}

	statusBar := RenderStatusBar(p.Width, p.Mode, p.Overlay, p.ActivePane, p.Destructive)

	// Border colors based on active pane
	activeBorderColor := lipgloss.Color("12")
	inactiveBorderColor := lipgloss.Color("238")
	destructiveBorderColor := lipgloss.Color("9")

	leftBorderColor := inactiveBorderColor
	rightBorderColor := inactiveBorderColor
	if p.Destructive {
		rightBorderColor = destructiveBorderColor
	} else if p.ActivePane == 1 {
		rightBorderColor = activeBorderColor
	}
	if p.ActivePane == 0 {
		leftBorderColor = activeBorderColor
	}

	leftContentWidth := LeftPaneWidth - 2 // left + right border
	innerHeight := p.Height - 3           // status bar + top/bottom borders

	leftLines := renderRepoList(p.Repos, p.Selected, leftContentWidth, innerHeight)
	leftContent := strings.Join(leftLines, "\n")
	leftPane := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(leftBorderColor).
		Width(leftContentWidth).
		Height(innerHeight).
		Render(leftContent)

	rightContentWidth := p.Width - LeftPaneWidth - 2 // left + right border
	if rightContentWidth < 0 {
		rightContentWidth = 0
	}

	modeHeader := renderModeHeader(p.Mode, rightContentWidth)
	rightContentHeight := p.Height - BranchContentOverhead

	var repoPath string
	if p.Selected < len(p.Repos) {
		repoPath = p.Repos[p.Selected].Path
	}

	// Hide cursor in right pane when left pane is active
	branchSel := p.BranchSelected
	stashSel := p.StashSelected
	if p.ActivePane == 0 {
		branchSel = -1
		stashSel = -1
	}

	var rightLines []string
	switch {
	case p.Mode == 1 && len(p.Branches) > 0:
		rightLines = renderBranchPaneSelected(p.Branches, branchSel, p.BranchScroll, rightContentWidth, rightContentHeight, repoPath)
	case p.Mode == 2 && len(p.Stashes) > 0:
		rightLines = renderStashPane(p.Stashes, stashSel, p.StashScroll, rightContentWidth, rightContentHeight)
	default:
		rightLines = renderPlaceholderPane(rightContentWidth, rightContentHeight)
	}

	rightContent := modeHeader + "\n" + strings.Join(rightLines, "\n")
	rightPane := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(rightBorderColor).
		Width(rightContentWidth).
		Height(innerHeight).
		Render(rightContent)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	return content + "\n" + statusBar
}

// renderModeHeader produces the mode selector line shown at the top of the right pane.
func renderModeHeader(mode, width int) string {
	modes := []struct {
		key  int
		name string
	}{
		{1, "branches"},
		{2, "stashes"},
	}

	var parts []string
	for _, m := range modes {
		if mode == m.key {
			parts = append(parts, activeModeStyle.Render(fmt.Sprintf("[%d] %s", m.key, m.name)))
		} else {
			parts = append(parts, inactiveModeStyle.Render(fmt.Sprintf(" %d %s", m.key, m.name)))
		}
	}
	line := " " + strings.Join(parts, " ")
	separator := strings.Repeat("─", width)
	return line + "\n" + separator
}

// RenderStatusBar produces the bottom status bar (hints only, no mode tabs).
func RenderStatusBar(width, mode, overlay, activePane int, destructive bool) string {
	var hints string
	if overlay == 3 {
		hints = "  y: confirm  n/esc: cancel"
	} else if overlay != 0 {
		hints = "  ↑/↓ scroll  esc: close"
	} else if mode == 2 {
		hints = "  tab: pane  q/esc: quit  ↑/↓ select  enter: diff"
		if destructive {
			hints += "  " + dirtyRedStyle.Render("d: drop")
		} else {
			hints += "  D: destructive mode"
		}
	} else {
		keys := "  |  tab: pane  q/esc: quit"
		if activePane == 1 {
			keys += "  t: terminal  c: code"
			if destructive {
				keys += "  " + dirtyRedStyle.Render("d: delete")
			}
		}
		if !destructive {
			keys += "  D: destructive mode"
		}
		hints = " " + cleanStyle.Render("✔") + " clean  " + aheadBehindStyle.Render("●") + " ahead/behind  " + dirtyRedStyle.Render("●") + " dirty  " + noUpstreamStyle.Render("●") + " no upstream" + keys
	}

	return statusStyle.Width(width).Render(hints)
}

func renderRepoList(repos []scanner.Repo, selected, width, height int) []string {
	lines := make([]string, height)

	start := 0
	if selected >= height {
		start = selected - height + 1
	}

	for i := 0; i < height; i++ {
		idx := start + i
		if idx < len(repos) {
			name := repos[idx].DisplayName
			if idx == selected {
				line := fmt.Sprintf(" > %s", name)
				lines[i] = selectedStyle.Width(width).Render(line)
			} else {
				line := fmt.Sprintf("   %s", name)
				lines[i] = repoStyle.Width(width).Render(line)
			}
		} else {
			lines[i] = strings.Repeat(" ", width)
		}
	}

	return lines
}

func renderBranchPane(rows []gitquery.BranchRow, width, height int) []string {
	return renderBranchPaneSelected(rows, 0, 0, width, height, "")
}

func renderBranchPaneSelected(rows []gitquery.BranchRow, selected, scroll, width, height int, repoPath string) []string {
	var content []string

	for i, row := range rows {
		b := row.Branch
		branch := branchStyle.Render(b.Name)

		var indicators string
		if b.Ahead > 0 || b.Behind > 0 {
			indicators += aheadBehindStyle.Render(" ●")
			indicators += fmt.Sprintf(" +%d/-%d", b.Ahead, b.Behind)
		}
		if b.Dirty {
			indicators += dirtyRedStyle.Render(" ●")
			indicators += fmt.Sprintf(" %d files ", b.FilesChanged)
			indicators += diffAddStyle.Render(fmt.Sprintf("+%d", b.LinesAdded))
			indicators += "/" + diffDelStyle.Render(fmt.Sprintf("-%d", b.LinesDeleted))
		}
		if !b.HasUpstream || b.UpstreamGone {
			indicators += noUpstreamStyle.Render(" ●")
		}

		if indicators == "" {
			indicators = cleanStyle.Render(" ✔")
		}

		var locationLabel string
		if row.WorktreePath != "" {
			if repoPath != "" && row.WorktreePath == repoPath {
				locationLabel = " " + commitStyle.Render("[root]")
			} else {
				locationLabel = " " + commitStyle.Render(fmt.Sprintf("[%s]", row.WorktreePath))
			}
		}

		line := "   " + branch + indicators + locationLabel
		if i == selected {
			line = branchSelStyle.Render(" > " + strings.TrimPrefix(line, "   "))
		}
		content = append(content, line)

		// Unpushed commits (max 5) — skipped for expansion rows
		if !row.IsExpansion {
			maxShow := 5
			for j, msg := range b.Unpushed {
				if j >= maxShow {
					remaining := len(b.Unpushed) - maxShow
					content = append(content, "    "+commitStyle.Render(fmt.Sprintf("... and %d more", remaining)))
					break
				}
				content = append(content, "    "+commitStyle.Render(msg))
			}
		}
	}

	// Truncate lines to pane width
	for i, line := range content {
		content[i] = truncateToWidth(line, width)
	}

	// Apply scroll offset
	if scroll > len(content) {
		scroll = len(content)
	}
	visible := content[scroll:]

	lines := make([]string, height)
	copy(lines, visible)
	return lines
}

func renderStashPane(stashes []gitquery.Stash, selected, scroll, width, height int) []string {
	var content []string

	for i, s := range stashes {
		content = append(content, stashEntryLines(s, i, selected, width)...)
	}

	// Apply scroll offset
	if scroll >= len(content) {
		scroll = max(0, len(content)-1)
	}
	visible := content[scroll:]

	lines := make([]string, height)
	copy(lines, visible)
	return lines
}

// stashEntryLines renders a single stash entry as 1-2 visual lines.
// Long messages wrap to a second line, capped at 2 lines total.
func stashEntryLines(s gitquery.Stash, index, selected, width int) []string {
	date := s.Date
	if len(date) > 10 {
		date = date[:10]
	}

	msgRunes := []rune(s.Message)
	msgAvail := width - stashEntryOverhead
	if msgAvail < 1 {
		msgAvail = 1
	}

	if index == selected {
		return stashEntrySelected(date, msgRunes, msgAvail, width)
	}
	return stashEntryNormal(date, msgRunes, msgAvail, width)
}

func stashEntrySelected(date string, msgRunes []rune, msgAvail, width int) []string {
	if len(msgRunes) <= msgAvail {
		raw := fmt.Sprintf(" > %s  %s", date, string(msgRunes))
		pad := width - lipgloss.Width(raw)
		if pad > 0 {
			raw += strings.Repeat(" ", pad)
		}
		return []string{stashSelStyle.Render(raw)}
	}

	// Two lines
	line1Msg := string(msgRunes[:msgAvail])
	line2Runes := msgRunes[msgAvail:]
	indent := strings.Repeat(" ", stashEntryOverhead)
	line2Avail := width - stashEntryOverhead
	if len(line2Runes) > line2Avail {
		line2Runes = line2Runes[:line2Avail]
	}

	raw1 := fmt.Sprintf(" > %s  %s", date, line1Msg)
	raw2 := indent + string(line2Runes)
	pad1 := width - lipgloss.Width(raw1)
	if pad1 > 0 {
		raw1 += strings.Repeat(" ", pad1)
	}
	pad2 := width - lipgloss.Width(raw2)
	if pad2 > 0 {
		raw2 += strings.Repeat(" ", pad2)
	}
	return []string{stashSelStyle.Render(raw1), stashSelStyle.Render(raw2)}
}

func stashEntryNormal(date string, msgRunes []rune, msgAvail, width int) []string {
	dateStr := stashDateStyle.Render(date)

	if len(msgRunes) <= msgAvail {
		msgStr := stashMsgStyle.Render(string(msgRunes))
		line := fmt.Sprintf("   %s  %s", dateStr, msgStr)
		return []string{truncateToWidth(line, width)}
	}

	// Two lines
	line1Msg := string(msgRunes[:msgAvail])
	line2Runes := msgRunes[msgAvail:]
	indent := strings.Repeat(" ", stashEntryOverhead)
	line2Avail := width - stashEntryOverhead
	if len(line2Runes) > line2Avail {
		line2Runes = line2Runes[:line2Avail]
	}

	line1 := fmt.Sprintf("   %s  %s", dateStr, stashMsgStyle.Render(line1Msg))
	line2 := indent + stashMsgStyle.Render(string(line2Runes))
	return []string{truncateToWidth(line1, width), truncateToWidth(line2, width)}
}

func renderOverlay(p RenderParams) string {
	statusBar := RenderStatusBar(p.Width, p.Mode, p.Overlay, p.ActivePane, p.Destructive)
	contentHeight := p.Height - 1

	// Confirmation dialog overlay
	if p.Overlay == 3 {
		lines := renderConfirmDialog(p.ConfirmPrompt, p.ConfirmForce, p.Width, contentHeight)
		return strings.Join(lines, "\n") + "\n" + statusBar
	}

	var diffLines []string
	if p.OverlayDiff != "" {
		diffLines = strings.Split(p.OverlayDiff, "\n")
	}

	// Apply scroll offset
	start := p.OverlayScroll
	if start > len(diffLines) {
		start = len(diffLines)
	}
	visible := diffLines[start:]

	lines := make([]string, contentHeight)
	for i := 0; i < contentHeight; i++ {
		if i >= len(visible) {
			break
		}
		line := visible[i]
		switch {
		case strings.HasPrefix(line, "+"):
			lines[i] = diffAddStyle.Render(line)
		case strings.HasPrefix(line, "-"):
			lines[i] = diffDelStyle.Render(line)
		case strings.HasPrefix(line, "@@"), strings.HasPrefix(line, "diff "):
			lines[i] = diffHdrStyle.Render(line)
		default:
			lines[i] = line
		}
		lines[i] = truncateToWidth(lines[i], p.Width)
	}

	return strings.Join(lines, "\n") + "\n" + statusBar
}

func renderConfirmDialog(prompt string, force bool, width, height int) []string {
	lines := make([]string, height)
	mid := height / 2
	if mid < len(lines) {
		pad := (width - lipgloss.Width(prompt)) / 2
		if pad < 0 {
			pad = 0
		}
		style := activeModeStyle
		if force {
			style = dirtyRedStyle.Bold(true)
		}
		lines[mid] = strings.Repeat(" ", pad) + style.Render(prompt)
	}
	return lines
}

// truncateToWidth trims a styled string to fit within maxWidth visible columns.
func truncateToWidth(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Strip ANSI, truncate runes, re-measure. Crude but correct for our use.
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}

func renderPlaceholderPane(width, height int) []string {
	lines := make([]string, height)
	placeholder := placeholderStyle.Render("nothing here yet")
	mid := height / 2
	pad := (width - lipgloss.Width(placeholder)) / 2
	if pad < 0 {
		pad = 0
	}
	lines[mid] = strings.Repeat(" ", pad) + placeholder
	return lines
}
