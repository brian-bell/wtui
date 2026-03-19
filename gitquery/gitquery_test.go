package gitquery_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brian-bell/wt/gitquery"
)

// realPath resolves symlinks (macOS /var → /private/var).
func realPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

// initRepo creates a git repo in dir with one commit on "main".
func initRepo(t *testing.T, dir string) {
	t.Helper()
	run(t, dir, "git", "init", "-b", "main")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func TestListWorktrees_SingleMainWorktree(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(wts))
	}
	if wts[0].Branch != "main" {
		t.Errorf("expected branch 'main', got %q", wts[0].Branch)
	}
	if wts[0].Path != dir {
		t.Errorf("expected path %q, got %q", dir, wts[0].Path)
	}
	if wts[0].IsBare {
		t.Error("expected IsBare=false")
	}
}

func TestListWorktrees_MultipleWorktrees(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	featurePath := filepath.Join(dir, "..", "feature-wt")
	run(t, dir, "git", "worktree", "add", "-b", "feature/auth", featurePath)

	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(wts))
	}

	branches := map[string]bool{}
	for _, wt := range wts {
		branches[wt.Branch] = true
	}
	if !branches["main"] {
		t.Error("expected branch 'main'")
	}
	if !branches["feature/auth"] {
		t.Error("expected branch 'feature/auth'")
	}
}

func TestListWorktrees_DirtyWorktree(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	// Create an uncommitted file → dirty
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}

	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(wts))
	}
	if !wts[0].Dirty {
		t.Error("expected Dirty=true for worktree with uncommitted file")
	}
}

func TestListWorktrees_CleanWorktree(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wts[0].Dirty {
		t.Error("expected Dirty=false for clean worktree")
	}
}

func TestListWorktrees_AheadBehind(t *testing.T) {
	// Create a bare "remote" and clone it
	tmp := realPath(t, t.TempDir())
	bare := filepath.Join(tmp, "remote.git")
	clone := filepath.Join(tmp, "clone")

	run(t, tmp, "git", "init", "--bare", "-b", "main", bare)
	run(t, tmp, "git", "clone", bare, clone)
	run(t, clone, "git", "config", "user.email", "test@test.com")
	run(t, clone, "git", "config", "user.name", "Test")

	// Create initial commit and push
	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", "first")
	run(t, clone, "git", "push")

	// Make 2 local commits (ahead by 2)
	for i := range 2 {
		if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte(string(rune('b'+i))), 0o644); err != nil {
			t.Fatal(err)
		}
		run(t, clone, "git", "add", ".")
		run(t, clone, "git", "commit", "-m", "local commit")
	}

	wts, err := gitquery.ListWorktrees(clone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(wts))
	}
	if wts[0].Ahead != 2 {
		t.Errorf("expected Ahead=2, got %d", wts[0].Ahead)
	}
	if wts[0].Behind != 0 {
		t.Errorf("expected Behind=0, got %d", wts[0].Behind)
	}
}

func TestListWorktrees_UnpushedCommits(t *testing.T) {
	tmp := realPath(t, t.TempDir())
	bare := filepath.Join(tmp, "remote.git")
	clone := filepath.Join(tmp, "clone")

	run(t, tmp, "git", "init", "--bare", "-b", "main", bare)
	run(t, tmp, "git", "clone", bare, clone)
	run(t, clone, "git", "config", "user.email", "test@test.com")
	run(t, clone, "git", "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", "first")
	run(t, clone, "git", "push")

	// Make a local commit with a known message
	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", "Fix login bug")

	wts, err := gitquery.ListWorktrees(clone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts[0].Unpushed) != 1 {
		t.Fatalf("expected 1 unpushed commit, got %d", len(wts[0].Unpushed))
	}
	if !strings.Contains(wts[0].Unpushed[0], "Fix login bug") {
		t.Errorf("expected unpushed to contain 'Fix login bug', got %q", wts[0].Unpushed[0])
	}
}

func TestListWorktrees_UnpushedTruncation(t *testing.T) {
	tmp := realPath(t, t.TempDir())
	bare := filepath.Join(tmp, "remote.git")
	clone := filepath.Join(tmp, "clone")

	run(t, tmp, "git", "init", "--bare", "-b", "main", bare)
	run(t, tmp, "git", "clone", bare, clone)
	run(t, clone, "git", "config", "user.email", "test@test.com")
	run(t, clone, "git", "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", "first")
	run(t, clone, "git", "push")

	// Commit with a very long message (>60 chars)
	longMsg := strings.Repeat("x", 80)
	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", longMsg)

	wts, err := gitquery.ListWorktrees(clone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts[0].Unpushed) != 1 {
		t.Fatalf("expected 1 unpushed commit, got %d", len(wts[0].Unpushed))
	}
	if len(wts[0].Unpushed[0]) > 60 {
		t.Errorf("expected unpushed message truncated to ≤60 chars, got %d: %q", len(wts[0].Unpushed[0]), wts[0].Unpushed[0])
	}
}

func TestListWorktrees_NoUpstream(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	// No remote set up → no upstream
	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wts[0].Ahead != 0 {
		t.Errorf("expected Ahead=0 with no upstream, got %d", wts[0].Ahead)
	}
	if wts[0].Behind != 0 {
		t.Errorf("expected Behind=0 with no upstream, got %d", wts[0].Behind)
	}
	if len(wts[0].Unpushed) != 0 {
		t.Errorf("expected no unpushed with no upstream, got %d", len(wts[0].Unpushed))
	}
}

func TestListWorktrees_InvalidPath(t *testing.T) {
	_, err := gitquery.ListWorktrees("/no/such/path")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}
