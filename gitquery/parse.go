package gitquery

import "strings"

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
