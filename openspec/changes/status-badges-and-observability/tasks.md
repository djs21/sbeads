## 1. Types, Schemas & Configuration Models (Bead: `sbeads-55g`)

- [x] 1.1 Update `internal/types/types.go` adding `Agent string` and `Assignee string` to `types.Issue` struct.
- [x] 1.2 Implement SQLite schema migration in `backend/storage.go` (`CreateTables`) executing `ALTER TABLE issues ADD COLUMN agent TEXT DEFAULT '';` and `ALTER TABLE issues ADD COLUMN assignee TEXT DEFAULT '';` safely.
- [x] 1.3 Update `backend/storage.go` queries (`CreateIssue`, `GetIssue`, `ListIssues`, `GetReadyIssues`) to scan and persist `agent` and `assignee`.
- [x] 1.4 Write unit tests in `backend/storage_test.go` verifying schema migration idempotency, agent persistence, and backward compatibility.

## 2. Parsing, Input Validation & State Machine (Bead: `sbeads-55g` & `sbeads-8q7`)

- [x] 2.1 Update `ClaimIssue(id, assignee, agent string) error` in `backend/storage.go` to store `agent` and `assignee` in database.
- [x] 2.2 Implement `UnclaimIssue(id string) error` in `backend/storage.go` resetting `is_claimed = 0`, `is_blocked = 0`, `agent = ''`, `assignee = ''`.
- [x] 2.3 Implement domain methods `Claim` and `Unclaim` in `internal/domain/issue.go` with transition validation (cannot unclaim an issue that is not claimed).
- [x] 2.4 Add domain-level unit tests in `internal/domain/issue_test.go` for claim and unclaim lifecycle transitions.

## 3. Status Engine & Badge Formatting Slice (Bead: `sbeads-30k` & `sbeads-a5c`)

- [x] 3.1 Create `internal/domain/status.go` (~90 LOC) defining status enum (`StatusOpen`, `StatusClaimed`, `StatusBlocked`, `StatusClosed`).
- [x] 3.2 Implement `ComputeStatus(issue *types.Issue, hasUnclosedDeps bool) Status` in `internal/domain/status.go`.
- [x] 3.3 Implement `FormatOneLiner(issue *types.Issue, status Status) string` generating `<id> [STATUS] [EPIC] (P<priority>/<type>) <title> <[@agent]>`.
- [x] 3.4 Write unit tests in `internal/domain/status_test.go` verifying badge evaluation, epic marker, priority tags, and agent appending.

## 4. Handlers, CLI Slices & Controllers (Bead: `sbeads-lrp` & `sbeads-8q7`)

- [x] 4.1 Implement `cmd/bd/unclaim.go` (~50 LOC) with `NewUnclaimCmd()` supporting `--issue-id` and positional ID.
- [x] 4.2 Implement `cmd/bd/list.go` (~120 LOC) with `NewListCmd()` supporting `--all`, `--closed`, `--blocked`, `--tree`, and `--json`.
- [x] 4.3 Implement tree rendering logic in `cmd/bd/list.go` to display child tasks hierarchically indented under their parent/epic.
- [x] 4.4 Register `NewListCmd()` and `NewUnclaimCmd()` in `cmd/bd/root.go`.

## 5. View Enhancements & Lifecycle Wiring (Bead: `sbeads-g6l`)

- [x] 5.1 Update `cmd/bd/ready.go` to format ready issues using `FormatOneLiner`.
- [x] 5.2 Update `cmd/bd/show.go` to display computed status, agent info, blockers (`Blocked by:`), and dependents (`Blocks:`).
- [x] 5.3 Implement helper in `backend/storage.go` or `internal/domain/dependency.go` to fetch detailed dependency edges (`GetDependencyEdges(issueID)`).

## 6. Unit & Slice Test Suites (Strict < 400 LOC Per Test File)

- [x] 6.1 Implement `cmd/bd/list_test.go` (~100 LOC) verifying `sbeads list`, `--all`, `--closed`, `--blocked`, `--tree`, and `--json`.
- [x] 6.2 Implement `cmd/bd/unclaim_test.go` (~60 LOC) verifying unclaim behavior and error handling.
- [x] 6.3 Update `cmd/bd/ready_test.go` and `cmd/bd/show_test.go` asserting enhanced output formats and dependency sections.

## 7. Documentation & Contract Updates

- [x] 7.1 Update `README.md` with `sbeads list`, `sbeads unclaim`, status badges reference, and tree view examples.
- [x] 7.2 Update `skills/sbeads/SKILL.md` with new agent coordination patterns (`claim` $\rightarrow$ `unclaim` $\rightarrow$ `close`) and list inspection.

## 8. End-to-End & Swarm Verification Scenarios

- [x] 8.1 Verify E2E multi-agent scenario: agent claims task, status changes to `[C] [@agent]`, unclaims task, status reverts to `[O]`.
- [x] 8.2 Verify E2E tree view: create epic with child tasks, verify `sbeads list --tree` renders tree hierarchy.
- [x] 8.3 Verify E2E ready output: verify priority tags `(P1/feat)` and `[O]` badges.

## 9. Final Quality Gates & Closeout

- [x] 9.1 Run full race-detector test suite: `go test -race ./...`.
- [x] 9.2 Compile release binary via `./build.sh --dev` without warnings.
- [x] 9.3 Perform line-of-code ceiling audit verifying no file in `cmd/bd/` or `internal/` exceeds 400 LOC.
- [x] 9.4 Validate OpenSpec change with `openspec validate status-badges-and-observability`.
