package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Repo represents a discovered git repository.
type Repo struct {
	Path        string
	DisplayName string
}

// ScanOptions configures the scanner.
type ScanOptions struct {
	Root     string
	MaxDepth int
}

// Scan discovers git repositories under the configured root.
// Returns repos sorted alphabetically by DisplayName.
func Scan(opts ScanOptions) ([]Repo, error) {
	root := opts.Root
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		root = filepath.Join(home, "dev")
	}

	maxDepth := opts.MaxDepth
	if maxDepth == 0 {
		maxDepth = 2
	}

	var repos []Repo

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), "-worktrees") {
			continue
		}

		path := filepath.Join(root, entry.Name())

		if isRepo(path) {
			repos = append(repos, Repo{
				Path:        path,
				DisplayName: entry.Name(),
			})
			continue
		}

		if maxDepth >= 2 {
			subEntries, err := os.ReadDir(path)
			if err != nil {
				continue
			}
			for _, sub := range subEntries {
				if !sub.IsDir() {
					continue
				}
				if strings.HasSuffix(sub.Name(), "-worktrees") {
					continue
				}
				subPath := filepath.Join(path, sub.Name())
				if isRepo(subPath) {
					repos = append(repos, Repo{
						Path:        subPath,
						DisplayName: entry.Name() + "/" + sub.Name(),
					})
				}
			}
		}
	}

	sort.Slice(repos, func(i, j int) bool {
		return strings.ToLower(repos[i].DisplayName) < strings.ToLower(repos[j].DisplayName)
	})

	return repos, nil
}

func isRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}
