package gitquery_test

import (
	"testing"

	"github.com/brian-bell/wtui/gitquery"
)

func TestParseCommitLog_ParsesMultipleCommits(t *testing.T) {
	input := "abc1234\x00Alice\x002 hours ago\x00Add feature\nabc5678\x00Bob\x003 days ago\x00Fix bug\n"

	commits := gitquery.ParseCommitLog(input)

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	if commits[0].Hash != "abc1234" {
		t.Errorf("expected Hash %q, got %q", "abc1234", commits[0].Hash)
	}
	if commits[0].Author != "Alice" {
		t.Errorf("expected Author %q, got %q", "Alice", commits[0].Author)
	}
	if commits[0].Date != "2 hours ago" {
		t.Errorf("expected Date %q, got %q", "2 hours ago", commits[0].Date)
	}
	if commits[0].Subject != "Add feature" {
		t.Errorf("expected Subject %q, got %q", "Add feature", commits[0].Subject)
	}
	if commits[1].Hash != "abc5678" {
		t.Errorf("expected Hash %q, got %q", "abc5678", commits[1].Hash)
	}
}

func TestParseCommitLog_EmptyInput(t *testing.T) {
	if commits := gitquery.ParseCommitLog(""); commits != nil {
		t.Errorf("expected nil, got %v", commits)
	}
	if commits := gitquery.ParseCommitLog("  \n"); commits != nil {
		t.Errorf("expected nil for whitespace, got %v", commits)
	}
}

func TestParseCommitLog_MalformedLineSkipped(t *testing.T) {
	input := "abc1234\x00Alice\x002 hours ago\x00Add feature\ngarbage line\nabc5678\x00Bob\x003 days ago\x00Fix bug\n"

	commits := gitquery.ParseCommitLog(input)

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits (malformed skipped), got %d", len(commits))
	}
}

func TestParseReflog_ParsesMultipleEntries(t *testing.T) {
	input := "abc1234\x00HEAD@{0}\x002 hours ago\x00commit: Add feature\nabc5678\x00HEAD@{1}\x003 days ago\x00checkout: moving from main to feat\n"

	entries := gitquery.ParseReflog(input)

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Hash != "abc1234" {
		t.Errorf("expected Hash %q, got %q", "abc1234", entries[0].Hash)
	}
	if entries[0].Selector != "HEAD@{0}" {
		t.Errorf("expected Selector %q, got %q", "HEAD@{0}", entries[0].Selector)
	}
	if entries[0].Date != "2 hours ago" {
		t.Errorf("expected Date %q, got %q", "2 hours ago", entries[0].Date)
	}
	if entries[0].Subject != "commit: Add feature" {
		t.Errorf("expected Subject %q, got %q", "commit: Add feature", entries[0].Subject)
	}
}

func TestParseReflog_EmptyInput(t *testing.T) {
	if entries := gitquery.ParseReflog(""); entries != nil {
		t.Errorf("expected nil, got %v", entries)
	}
}

func TestParseReflog_MalformedLineSkipped(t *testing.T) {
	input := "abc1234\x00HEAD@{0}\x002 hours ago\x00commit: Add feature\nbadline\n"

	entries := gitquery.ParseReflog(input)

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (malformed skipped), got %d", len(entries))
	}
}

func TestParseStashList_ParsesMultipleStashes(t *testing.T) {
	input := "stash@{0}\x002024-01-15 10:30:00 -0500\x00WIP on main: abc1234 some work\nstash@{1}\x002024-01-14 09:00:00 -0500\x00On main: save progress\n"

	stashes := gitquery.ParseStashList(input)

	if len(stashes) != 2 {
		t.Fatalf("expected 2 stashes, got %d", len(stashes))
	}
	if stashes[0].Index != 0 {
		t.Errorf("expected Index 0, got %d", stashes[0].Index)
	}
	if stashes[0].Message != "WIP on main: abc1234 some work" {
		t.Errorf("expected Message %q, got %q", "WIP on main: abc1234 some work", stashes[0].Message)
	}
	if stashes[1].Index != 1 {
		t.Errorf("expected Index 1, got %d", stashes[1].Index)
	}
	if stashes[1].Date != "2024-01-14 09:00:00 -0500" {
		t.Errorf("expected Date %q, got %q", "2024-01-14 09:00:00 -0500", stashes[1].Date)
	}
}

func TestParseStashList_EmptyInput(t *testing.T) {
	if stashes := gitquery.ParseStashList(""); stashes != nil {
		t.Errorf("expected nil, got %v", stashes)
	}
}

func TestParseStashList_MalformedLineSkipped(t *testing.T) {
	input := "stash@{0}\x002024-01-15 10:30:00 -0500\x00WIP\nbroken\n"

	stashes := gitquery.ParseStashList(input)

	if len(stashes) != 1 {
		t.Fatalf("expected 1 stash (malformed skipped), got %d", len(stashes))
	}
}

func TestParseNumstat_ParsesAddedDeleted(t *testing.T) {
	input := "3\t1\tfile.go\n10\t5\tother.go\n"

	added, deleted := gitquery.ParseNumstat(input)

	if added != 13 {
		t.Errorf("expected added 13, got %d", added)
	}
	if deleted != 6 {
		t.Errorf("expected deleted 6, got %d", deleted)
	}
}

func TestParseNumstat_EmptyInput(t *testing.T) {
	added, deleted := gitquery.ParseNumstat("")
	if added != 0 || deleted != 0 {
		t.Errorf("expected (0, 0), got (%d, %d)", added, deleted)
	}
}

func TestParseNumstat_BinaryFilesIgnored(t *testing.T) {
	input := "3\t1\ttext.go\n-\t-\tbinary.png\n"

	added, deleted := gitquery.ParseNumstat(input)

	if added != 3 {
		t.Errorf("expected added 3, got %d", added)
	}
	if deleted != 1 {
		t.Errorf("expected deleted 1, got %d", deleted)
	}
}

func TestParseNumstat_MalformedLineSkipped(t *testing.T) {
	input := "3\t1\tfile.go\nbadline\n2\t0\tother.go\n"

	added, deleted := gitquery.ParseNumstat(input)

	if added != 5 {
		t.Errorf("expected added 5, got %d", added)
	}
	if deleted != 1 {
		t.Errorf("expected deleted 1, got %d", deleted)
	}
}
