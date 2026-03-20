package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/brian-bell/wt/gitquery"
	"github.com/brian-bell/wt/scanner"
)

const LeftPaneWidth = 30

var (
	repoStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	selectedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Reverse(true)
	placeholderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	statusStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dividerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
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
	Branches       []gitquery.Branch
	Stashes        []gitquery.Stash
	BranchSelected int
	StashSelected  int
	Overlay        int
	OverlayDiff    string
	OverlayScroll  int
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

	statusBar := RenderStatusBar(p.Width, p.Mode, p.Overlay)
	contentHeight := p.Height - 1 // reserve 1 row for status bar

	// Build left pane
	leftLines := renderRepoList(p.Repos, p.Selected, contentHeight)

	// Build right pane
	rightWidth := p.Width - LeftPaneWidth - 1 // 1 for divider
	if rightWidth < 0 {
		rightWidth = 0
	}

	var rightLines []string
	switch {
	case p.Mode == 1 && len(p.Branches) > 0:
		rightLines = renderBranchPaneSelected(p.Branches, p.BranchSelected, rightWidth, contentHeight)
	case p.Mode == 2 && len(p.Stashes) > 0:
		rightLines = renderStashPane(p.Stashes, p.StashSelected, rightWidth, contentHeight)
	default:
		rightLines = renderPlaceholderPane(rightWidth, contentHeight)
	}

	// Build divider
	divider := make([]string, contentHeight)
	for i := range divider {
		divider[i] = dividerStyle.Render("│")
	}

	// Combine panes
	left := strings.Join(leftLines, "\n")
	mid := strings.Join(divider, "\n")
	right := strings.Join(rightLines, "\n")

	content := lipgloss.JoinHorizontal(lipgloss.Top, left, mid, right)

	return content + "\n" + statusBar
}

// RenderStatusBar produces the bottom status bar.
func RenderStatusBar(width, mode, overlay int) string {
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

	var hints string
	if overlay != 0 {
		hints = "  ↑/↓ scroll  esc: close"
	} else if mode == 2 {
		hints = "  ↑/↓ select  enter: diff  tab: repo  ←/→: mode  q/esc: quit"
	} else {
		hints = "  ↑/↓ enter  " + cleanStyle.Render("✔") + " clean  " + aheadBehindStyle.Render("●") + " ahead/behind  " + dirtyRedStyle.Render("●") + " dirty  " + noUpstreamStyle.Render("●") + " no upstream  tab: repo  ←/→: mode  q/esc: quit"
	}

	text := "  " + strings.Join(parts, " ") + hints
	return statusStyle.Width(width).Render(text)
}

func renderRepoList(repos []scanner.Repo, selected, height int) []string {
	lines := make([]string, height)

	// Calculate viewport for scrolling
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
				lines[i] = selectedStyle.Width(LeftPaneWidth).Render(line)
			} else {
				line := fmt.Sprintf("   %s", name)
				lines[i] = repoStyle.Width(LeftPaneWidth).Render(line)
			}
		} else {
			lines[i] = strings.Repeat(" ", LeftPaneWidth)
		}
	}

	return lines
}

func renderBranchPane(branches []gitquery.Branch, width, height int) []string {
	return renderBranchPaneSelected(branches, 0, width, height)
}

func renderBranchPaneSelected(branches []gitquery.Branch, selected, width, height int) []string {
	var content []string
	diffableIndex := 0

	for _, b := range branches {
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

		var annotation string
		if b.IsWorktree {
			annotation = formatWorktreeAnnotation(b.WorktreePaths)
		}

		line := "  " + branch + indicators + annotation
		if b.Dirty && b.IsWorktree && diffableIndex == selected {
			line = branchSelStyle.Render(" > " + strings.TrimPrefix(line, "  "))
		}
		content = append(content, line)
		if b.Dirty && b.IsWorktree {
			diffableIndex++
		}

		// Unpushed commits (max 5)
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

	// Truncate lines to pane width
	for i, line := range content {
		content[i] = truncateToWidth(line, width)
	}

	// Pad to fill height
	lines := make([]string, height)
	for i := 0; i < height; i++ {
		if i < len(content) {
			lines[i] = content[i]
		} else {
			lines[i] = ""
		}
	}
	return lines
}

func formatWorktreeAnnotation(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	shown := paths
	suffix := ""
	if len(paths) > 2 {
		shown = paths[:2]
		suffix = fmt.Sprintf(", +%d more", len(paths)-2)
	}

	return " " + commitStyle.Render(fmt.Sprintf("[%s%s]", strings.Join(shown, ", "), suffix))
}

func renderStashPane(stashes []gitquery.Stash, selected, width, height int) []string {
	var content []string

	for i, s := range stashes {
		// Truncate date to just the date portion (first 10 chars)
		date := s.Date
		if len(date) > 10 {
			date = date[:10]
		}

		dateStr := stashDateStyle.Render(date)
		msgStr := stashMsgStyle.Render(s.Message)
		line := fmt.Sprintf("  %s  %s", dateStr, msgStr)

		if i == selected {
			// Render selected line with highlight
			line = stashSelStyle.Width(width).Render(fmt.Sprintf(" > %s  %s", date, s.Message))
		}

		content = append(content, truncateToWidth(line, width))
	}

	// Pad to fill height
	lines := make([]string, height)
	for i := 0; i < height; i++ {
		if i < len(content) {
			lines[i] = content[i]
		} else {
			lines[i] = ""
		}
	}
	return lines
}

func renderOverlay(p RenderParams) string {
	statusBar := RenderStatusBar(p.Width, p.Mode, p.Overlay)
	contentHeight := p.Height - 1

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
		if i < len(visible) {
			line := visible[i]
			// Colorize diff lines
			switch {
			case strings.HasPrefix(line, "+"):
				lines[i] = diffAddStyle.Render(line)
			case strings.HasPrefix(line, "-"):
				lines[i] = diffDelStyle.Render(line)
			case strings.HasPrefix(line, "@@"):
				lines[i] = diffHdrStyle.Render(line)
			case strings.HasPrefix(line, "diff "):
				lines[i] = diffHdrStyle.Render(line)
			default:
				lines[i] = line
			}
			lines[i] = truncateToWidth(lines[i], p.Width)
		} else {
			lines[i] = ""
		}
	}

	return strings.Join(lines, "\n") + "\n" + statusBar
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
	for i := 0; i < height; i++ {
		if i == mid {
			pad := (width - lipgloss.Width(placeholder)) / 2
			if pad < 0 {
				pad = 0
			}
			lines[i] = strings.Repeat(" ", pad) + placeholder
		} else {
			lines[i] = ""
		}
	}

	return lines
}
