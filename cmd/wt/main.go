package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brian-bell/wt/model"
	"github.com/brian-bell/wt/scanner"
)

func main() {
	root := os.Getenv("WORKTREE_ROOT")

	repos, err := scanner.Scan(scanner.ScanOptions{Root: root})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error scanning repos: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model.New(repos), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
