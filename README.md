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
| `↑`/`k` | Move selection up |
| `↓`/`j` | Move selection down |
| `1` | Worktree view (branch, dirty status, ahead/behind, unpushed commits) |
| `2` | Stashes view (placeholder) |
| `3` | Branches view (placeholder) |
| `q` | Quit |

### Worktree view

The right pane shows each worktree's:

- Branch name with dirty (`●`) or clean (`✔`) indicator
- Ahead/behind counts relative to upstream (`+2/-1`)
- Unpushed commit messages (up to 5, with overflow count)

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
