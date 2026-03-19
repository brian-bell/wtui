# wt — Git Worktree Manager

A lightweight shell script for managing git worktrees across multiple repositories.

## Installation

Copy `wt` to a directory on your `PATH` and make it executable:

```bash
cp wt ~/bin/wt
chmod +x ~/bin/wt
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `WORKTREE_ROOT` | `~/dev` | Root directory containing your repositories |

Worktrees for a repository named `my-repo` are stored in `$WORKTREE_ROOT/my-repo-worktrees/`.

## Usage

```
wt <command> <args>
```

### Create a worktree

```bash
wt create <repo> <branch>
```

Creates a worktree for `<branch>` in the `<repo>-worktrees/` directory. If the branch
doesn't exist locally or on the remote, it is created automatically from HEAD.

Branch names containing `/` are sanitized to `-` for the worktree directory name
(e.g., `feature/login` is stored in the directory `feature-login`).

```bash
# Existing branch
wt create my-app main

# New branch (created from HEAD)
wt create my-app feature/new-dashboard
# -> ~/dev/my-app-worktrees/feature-new-dashboard
```

### List worktrees

```bash
wt list <repo>
```

Lists all worktrees registered to the repository.

```bash
wt list my-app
```

### Remove a worktree

```bash
wt remove <repo> <branch>
```

Removes a worktree by its branch name. Uses the same `/` to `-` sanitization when
resolving the directory.

```bash
wt remove my-app feature/new-dashboard
```

### Prune stale references

```bash
wt prune <repo>
```

Cleans up worktree metadata for directories that no longer exist on disk.

```bash
wt prune my-app
```

## Directory Layout

Given `WORKTREE_ROOT=~/dev` and a repository called `my-app`:

```
~/dev/
  my-app/                        # main repository clone
  my-app-worktrees/
    feature-new-dashboard/       # worktree for feature/new-dashboard
    bugfix-auth-token/           # worktree for bugfix/auth-token
```

## Requirements

- Git 2.15+ (worktree support)
- Bash 4+
