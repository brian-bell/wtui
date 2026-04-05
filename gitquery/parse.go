package gitquery

import (
	"strconv"
	"strings"
)

// ParseReflog parses the output of git reflog --format=%h%x00%gd%x00%ar%x00%gs.
func ParseReflog(text string) []ReflogEntry {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	var entries []ReflogEntry
	for _, line := range strings.Split(text, "\n") {
		parts := strings.SplitN(line, "\x00", 4)
		if len(parts) != 4 {
			continue
		}
		entries = append(entries, ReflogEntry{
			Hash:     parts[0],
			Selector: parts[1],
			Date:     parts[2],
			Subject:  parts[3],
		})
	}
	return entries
}

// ParseStashList parses the output of git stash list --format=%gd%x00%ai%x00%s.
func ParseStashList(text string) []Stash {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	var stashes []Stash
	for _, line := range strings.Split(text, "\n") {
		parts := strings.SplitN(line, "\x00", 3)
		if len(parts) != 3 {
			continue
		}
		idxStr := strings.TrimPrefix(parts[0], "stash@{")
		idxStr = strings.TrimSuffix(idxStr, "}")
		idx, _ := strconv.Atoi(idxStr)
		stashes = append(stashes, Stash{
			Index:   idx,
			Date:    parts[1],
			Message: parts[2],
		})
	}
	return stashes
}

// ParseAheadBehind parses the output of git rev-list --count --left-right.
func ParseAheadBehind(text string) (int, int) {
	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) != 2 {
		return 0, 0
	}
	ahead, _ := strconv.Atoi(parts[0])
	behind, _ := strconv.Atoi(parts[1])
	return ahead, behind
}

// ParseNumstat parses the output of git diff --numstat.
// Returns total lines added and deleted. Binary files (shown as - -) contribute 0.
func ParseNumstat(text string) (int, int) {
	var added, deleted int
	for _, line := range splitLines(text) {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		a, _ := strconv.Atoi(fields[0])
		d, _ := strconv.Atoi(fields[1])
		added += a
		deleted += d
	}
	return added, deleted
}

// ParseCommitLog parses the output of git log --format=%h%x00%an%x00%ar%x00%s.
func ParseCommitLog(text string) []Commit {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	var commits []Commit
	for _, line := range strings.Split(text, "\n") {
		parts := strings.SplitN(line, "\x00", 4)
		if len(parts) != 4 {
			continue
		}
		commits = append(commits, Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    parts[2],
			Subject: parts[3],
		})
	}
	return commits
}
