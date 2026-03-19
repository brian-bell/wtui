package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/brian-bell/wt/scanner"
)

const LeftPaneWidth = 30

var (
	repoStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	selectedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Reverse(true)
	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	statusStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dividerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

// RenderParams holds everything the renderer needs.
type RenderParams struct {
	Repos    []scanner.Repo
	Selected int
	Width    int
	Height   int
}

// Render produces the full terminal view string.
func Render(p RenderParams) string {
	if p.Width == 0 {
		p.Width = 80
	}
	if p.Height == 0 {
		p.Height = 24
	}

	statusBar := RenderStatusBar(p.Width)
	contentHeight := p.Height - 1 // reserve 1 row for status bar

	// Build left pane
	leftLines := renderRepoList(p.Repos, p.Selected, contentHeight)

	// Build right pane
	rightWidth := p.Width - LeftPaneWidth - 1 // 1 for divider
	if rightWidth < 0 {
		rightWidth = 0
	}
	rightLines := renderRightPane(rightWidth, contentHeight)

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
func RenderStatusBar(width int) string {
	text := "  ↑/↓: navigate  q: quit"
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

func renderRightPane(width, height int) []string {
	lines := make([]string, height)

	placeholder := placeholderStyle.Render("nothing here yet")

	mid := height / 2
	for i := 0; i < height; i++ {
		if i == mid {
			// Center the placeholder horizontally
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
