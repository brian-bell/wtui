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
- **`actions/`** — Executes destructive git operations. `RemoveWorktree`/`ForceRemoveWorktree` run `git worktree remove` (with/without `--force`). `DeleteBranch`/`ForceDeleteBranch` run `git branch -d`/`-D`. `DropStash` runs `git stash drop stash@{N}`.
- **`scanner/`** — Discovers git repos under `WORKTREE_ROOT` (default `~/dev`), up to 2 levels deep. Excludes `*-worktrees` dirs. Detects both `.git` dirs and `.git` files (worktree markers). Returns repos sorted case-insensitively.
- **`gitquery/`** — Queries git data. `ListBranches(repoPath)` uses `git for-each-ref` to discover all local branches, then per branch queries upstream status, ahead/behind counts, unpushed commits, and worktree dirty status (files changed, lines added/deleted). Branches can carry multiple worktree paths. `BranchDiff(worktreePath)` returns `git diff HEAD` for a worktree. `ListStashes(repoPath)` runs `git stash list --format=%gd%x00%ai%x00%s`. `StashDiff(repoPath, index)` runs `git stash show -p stash@{N}`. Only list-level failures are hard errors; per-item failures silently default to zero values.
- **`model/`** — Bubble Tea Model. Branch cursor (`branchSelected`) is an absolute index into `m.branches` (all branches selectable, not just dirty+worktree). `enter` for diff still gated to dirty+worktree branches. `d` on any branch: worktree branches → `RemoveWorktree` confirm; non-worktree branches → `DeleteBranch` confirm. On failure, a `DeleteFailedMsg` triggers a second red-text force confirm dialog (`confirmForce=true`) that uses `--force`/`-D`. `d` in stash mode → `DropStash` confirm; on success `StashDroppedMsg` triggers `fetchStashes()` and clamps `stashSelected` if needed. `branchScroll` tracks viewport offset via `ensureBranchVisible()`. `r` refreshes current mode data. `tab` cycles repos; `left/right` (`h/l`) switch modes; `up/down` (`j/k`) navigate; `1/2` jump to modes. Result messages (`BranchResultMsg`, `StashResultMsg`, `StashDiffResultMsg`, `WorktreeRemovedMsg`, `BranchDeletedMsg`, `DeleteFailedMsg`, `StashDroppedMsg`) update state with stale-result protection. Confirm overlay intercepts all keys: `y`/`enter` executes `confirmAction`, `n`/`q`/`esc` cancels; clears `confirmForce` on close.
- **`ui/`** — Stateless rendering. Two-pane layout: left pane (30 chars, repo list with scrolling viewport) + divider + right pane (mode-aware: mode 1 shows branch details with viewport scroll via `BranchScroll`, mode 2 shows stash list with selection highlight). Branch pane highlighting uses absolute branch index (all branches selectable). Full-screen diff overlay replaces two-pane when active. Confirmation dialog overlay (`Overlay==3`) shows a centered prompt; `ConfirmForce=true` renders it in red for force-delete dialogs. `RenderParams` includes `ConfirmPrompt string`, `ConfirmForce bool`, and `BranchScroll int`. `RenderStatusBar(width, mode, overlay int)` renders mode-specific key hints: mode 1 shows `d: delete`, mode 2 shows `d: drop`. Branch status indicators: green `✔` (clean/even), yellow `●` (ahead/behind with `+N/-N` counts), red `●` (dirty with `N files +X/-Y`), purple `●` (no upstream). Indicators stack. Worktree branches show `[<path>]` annotation. Up to 5 unpushed commits shown with overflow.

## CI

CI runs on push to `main` and PRs targeting `main`. Checks: `gofmt`, `make test`, `make build`.

## Testing

Tests use real temp directories with actual `.git` dirs/files — no mocks. Scanner tests create nested repo structures; model tests simulate key messages via Bubble Tea's `Update()`. Gitquery tests create real git repos with remotes, commits, and worktrees to verify dirty/clean, ahead/behind, and unpushed detection.
