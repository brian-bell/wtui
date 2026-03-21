# wt

A terminal UI for managing git worktrees across repositories.

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)

## Install

```bash
git clone https://github.com/brian-bell/wt.git
cd wt
make build
```

The binary is built to `bin/wt`.

## Usage

```bash
# Run with default root (~/dev)
./bin/wt

# Run with a custom root
WORKTREE_ROOT=~/projects ./bin/wt
```

### Keys

| Key | Action |
|-----|--------|
| `tab` | Cycle to next repo (wrapping) |
| `↑`/`k` | Move selection up in right pane |
| `↓`/`j` | Move selection down in right pane |
| `←`/`h` | Switch to previous mode |
| `→`/`l` | Switch to next mode |
| `1`/`2` | Jump to branches / stashes mode |
| `enter` | View dirty branch diff in branches mode or stash diff in stashes mode |
| `d` | Delete branch/worktree (branches mode) or drop stash (stashes mode) |
| `esc`/`q` | Close overlay or quit |

### Branches view (mode 1)

The right pane shows all local branches alphabetically with stacking indicators:

- `✔` green: even with upstream, clean working tree
- `●` yellow: ahead/behind upstream — shows `+N/-N` counts
- `●` red: dirty worktree — shows `N files +X/-Y` (lines added/deleted)
- `●` purple: no upstream or upstream gone

Worktree branches are annotated with `[<path>]`. If the same branch is checked out in more than one worktree, the UI shows `[<path1>, <path2>, ...]`. Detached worktrees appear as `(detached)` rows with their path annotation. Branches ahead of upstream show up to 5 unpushed commit messages, with overflow count. When a branch is dirty and is a worktree, `enter` opens a full-screen diff overlay for that worktree.

### Stashes view (mode 2)

Browse stashes for the selected repo. Use `↑`/`↓` to select a stash, `enter` to view its diff in a full-screen overlay, `d` to drop the selected stash (with confirmation).

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `WORKTREE_ROOT` | `~/dev` | Root directory to scan for git repos (up to 2 levels deep) |

## Development

```bash
make build   # Build binary to bin/wt
make test    # Run all tests
make run     # Build and run
make tidy    # go mod tidy
make clean   # Remove bin/
```

## Requirements

- Go 1.26+
- Git 2.15+ (worktree support)
