# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # Build binary to bin/wt
make test           # Run all tests
make run            # Build and run the TUI
go test ./scanner   # Run tests for a single package
gofmt -l .          # Check formatting (CI enforces zero output)
```

## Architecture

Go TUI for managing git worktrees across repositories. Uses **Bubble Tea** (MVC pattern) with **lipgloss** for styling.

**Data flow:** `main` reads `WORKTREE_ROOT` env var → `scanner.Scan()` discovers repos → `model.New(repos)` creates Bubble Tea model → `Init()` fires async `fetchBranches` → `ui.Render()` draws two-pane layout.

- **`cmd/wt/main.go`** — Entry point. Wires scanner output into the Bubble Tea program (alt-screen mode).
- **`scanner/`** — Discovers git repos under `WORKTREE_ROOT` (default `~/dev`), up to 2 levels deep. Excludes `*-worktrees` dirs. Detects both `.git` dirs and `.git` files (worktree markers). Returns repos sorted case-insensitively.
- **`gitquery/`** — Queries git data. `ListBranches(repoPath)` uses `git for-each-ref` to discover all local branches, then per branch queries upstream status, ahead/behind counts, unpushed commits, and worktree dirty status (files changed, lines added/deleted). `BranchDiff(worktreePath)` returns `git diff HEAD` for a worktree. `ListStashes(repoPath)` runs `git stash list --format=%gd%x00%ai%x00%s`. `StashDiff(repoPath, index)` runs `git stash show -p stash@{N}`. Only list-level failures are hard errors; per-item failures silently default to zero values.
- **`model/`** — Bubble Tea Model. Holds repo list, selection index, terminal dimensions, active mode (1=branches, 2=stashes), branch/stash data, stash cursor, overlay state, and overlay diff content. `tab` cycles repos (wrapping); `left/right` (`h/l`) switch modes; `up/down` (`j/k`) navigate stash selection in mode 2. Number keys `1/2` jump directly to modes. Mode switches and repo changes fire async fetch commands (`fetchBranches` for mode 1, `fetchStashes` for mode 2). Result messages (`BranchResultMsg`, `StashResultMsg`, `StashDiffResultMsg`) update state with stale-result protection. Overlay intercepts all keys when open (esc/q close, up/down scroll).
- **`ui/`** — Stateless rendering. Two-pane layout: left pane (30 chars, repo list with scrolling viewport) + divider + right pane (mode-aware: mode 1 shows branch details, mode 2 shows stash list with selection highlight). Full-screen diff overlay replaces two-pane when active. Context-aware status bar shows 2 modes with mode-specific keybindings. Branch status indicators: green `✔` (clean/even), yellow `●` (ahead/behind with `+N/-N` counts), red `●` (dirty with `N files +X/-Y`), purple `●` (no upstream). Indicators stack. Worktree branches show `[wt: <path>]` annotation. Up to 5 unpushed commits shown with overflow.

## CI

CI runs on push to `main` and PRs targeting `main`. Checks: `gofmt`, `make test`, `make build`.

## Testing

Tests use real temp directories with actual `.git` dirs/files — no mocks. Scanner tests create nested repo structures; model tests simulate key messages via Bubble Tea's `Update()`. Gitquery tests create real git repos with remotes, commits, and worktrees to verify dirty/clean, ahead/behind, and unpushed detection.
