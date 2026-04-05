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
