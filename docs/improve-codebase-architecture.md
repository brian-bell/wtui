# Refactoring Candidates

Architectural deepening opportunities identified via codebase exploration. Ordered by impact.

## 1. Deepen `actions/` — from shallow wrappers to workflow-owning module

- **Cluster:** `actions/` (66 lines) + workflow logic scattered in `model/model.go` (confirm dialogs, retry-on-failure, multi-step sequences)
- **Why they're coupled:** Model orchestrates multi-step action workflows: "remove worktree → optionally delete branch → handle failure → offer force retry". The `actions/` package just wraps `exec.Command` — the real action logic lives in `model/`. Actions is one of the shallowest modules possible: 9 functions, each 2-5 lines, interface ≈ implementation.
- **Dependency category:** Same-process, side-effectful (executes git/system commands)
- **Test impact:** Model's action tests (`model_action_test.go`, 1,353 lines) would partially shift to action-level boundary tests. Currently model tests construct `tea.Msg` results manually to simulate git operations.

## 2. Decouple `ui/` from `gitquery` types — introduce render-oriented types at the boundary

- **Cluster:** `ui/ui.go` (637 lines) imports `gitquery` directly to access 5 struct types (`BranchRow`, `Stash`, `Worktree`, `Commit`, `ReflogEntry`). `model/` also imports both.
- **Why they're coupled:** `ui.RenderParams` contains `[]gitquery.Worktree`, `[]gitquery.Commit`, etc. Any field added/renamed/removed in gitquery ripples into ui rendering functions. The ui package reaches deep into gitquery struct internals (e.g., `branch.Unpushed`, `worktree.FilesChanged`).
- **Dependency category:** Shared data types creating transitive compile-time coupling
- **Test impact:** UI tests currently construct gitquery types directly. With render-oriented types, ui tests would be self-contained. gitquery tests wouldn't change.

## 3. Extract per-mode list behavior in `model/` — reduce the 5x cursor/scroll/fetch pattern

- **Cluster:** `model/model.go` repeats the same pattern 5 times (one per mode): cursor field + scroll field + `ensure*Visible()` + `fetch*()` + result handler + enter handler + delete handler. ~600 lines of structurally similar code.
- **Why they're coupled:** Each mode (Worktrees, Branches, Stashes, History, Reflog) implements the same "scrollable list with selection" concept but with mode-specific data types and actions. Adding a new mode requires touching 8+ locations.
- **Dependency category:** Internal structural duplication within a single package
- **Test impact:** Mode-specific tests in `model_test.go` (navigation) and `model_action_test.go` (actions) could partially consolidate. Currently 198 tests with significant structural repetition across modes.

## 4. Make `gitquery/` testable without real git repos — separate parsing from execution

- **Cluster:** `gitquery/gitquery.go` (527 lines) mixes `exec.Command` calls with string parsing logic. 6 pure parsing helpers exist but are interleaved with 10+ side-effectful functions.
- **Why they're coupled:** Functions like `ListBranches()` execute git commands, parse output, then call more git commands based on parsed results — all in one flow. The pure parsing logic (splitting worktree blocks, parsing branch lines) is tested only through integration tests that create real git repos.
- **Dependency category:** I/O execution mixed with pure data transformation
- **Test impact:** Current gitquery tests (1,147 lines) create real temporary git repos for every test. Separating parsing would allow fast unit tests for the parsing logic and focused integration tests for the git interaction.
