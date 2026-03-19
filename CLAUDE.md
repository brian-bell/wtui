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

**Data flow:** `main` reads `WORKTREE_ROOT` env var → `scanner.Scan()` discovers repos → `model.New(repos)` creates Bubble Tea model → `Init()` fires async `fetchWorktrees` → `ui.Render()` draws two-pane layout.

- **`cmd/wt/main.go`** — Entry point. Wires scanner output into the Bubble Tea program (alt-screen mode).
- **`scanner/`** — Discovers git repos under `WORKTREE_ROOT` (default `~/dev`), up to 2 levels deep. Excludes `*-worktrees` dirs. Detects both `.git` dirs and `.git` files (worktree markers). Returns repos sorted case-insensitively.
- **`gitquery/`** — Queries git worktree status. `ListWorktrees(repoPath)` shells out to `git worktree list --porcelain`, then per non-bare worktree runs `git status --porcelain`, `git rev-list --left-right @{upstream}...HEAD`, and `git log --oneline @{upstream}..HEAD`. Only `worktree list` failure is a hard error; per-worktree failures silently default to zero values.
- **`model/`** — Bubble Tea Model. Holds repo list, selection index, terminal dimensions, active mode (1=worktrees, 2=stashes, 3=branches), and worktree data. Keys `1`/`2`/`3` switch modes. Nav keys and mode-1 switch fire async `fetchWorktrees` cmd; `WorktreeResultMsg` updates state with stale-result protection (repo path must match current selection).
- **`ui/`** — Stateless rendering. Two-pane layout: left pane (30 chars, repo list with scrolling viewport) + divider + right pane (mode-aware: mode 1 shows worktree details, modes 2/3 show placeholder). Status bar at bottom.

## CI

CI runs on push to `main` and PRs targeting `main`. Checks: `gofmt`, `make test`, `make build`.

## Testing

Tests use real temp directories with actual `.git` dirs/files — no mocks. Scanner tests create nested repo structures; model tests simulate key messages via Bubble Tea's `Update()`. Gitquery tests create real git repos with remotes, commits, and worktrees to verify dirty/clean, ahead/behind, and unpushed detection.
