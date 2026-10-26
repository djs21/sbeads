## 1. Storage & Query Engine (Ready Filter & Dependencies)

- [x] 1.1 Implement `GetReadyIssues()` in `backend/storage.go` executing single SQL query filtering `is_closed = 0 AND is_claimed = 0 AND is_blocked = 0 AND NOT EXISTS (SELECT 1 FROM dependencies d JOIN issues p ON d.parent_id = p.id WHERE d.child_id = issues.id AND p.is_closed = 0)`.
- [x] 1.2 Implement short ID and prefix lookup in `backend/storage.go` (`GetIssue` matching exact ID or prefix match e.g. `sbeads-xxx` or `xxx`).
- [x] 1.3 Add index `idx_dependencies_parent_id` on `dependencies(parent_id)` and `idx_issues_status` on `issues(is_closed, is_claimed, is_blocked)`.
- [x] 1.4 Write comprehensive TDD unit tests in `backend/storage_test.go` covering closed filtering, claimed filtering, blocked filtering, dependency tree resolution, and prefix lookup.

## 2. Domain Service Logic & State Verification

- [x] 2.1 Update `internal/domain/issue.go` `ReadyIssues()` to invoke storage `GetReadyIssues()`.
- [x] 2.2 Ensure `ClaimIssue` and `CloseIssue` properly update timestamps and validate state transitions.
- [x] 2.3 Add domain-level unit tests in `internal/domain/issue_test.go` verifying ready issue resolution against test databases.

## 3. Types & Short ID Generator (Base36 + Collision Retry)

- [x] 3.1 Update `internal/types/types.go` with 3-character Base36 hash ID generator (`[0-9a-z]`, 46,656 space) formatted as `<prefix>-<hash3>` (defaulting to `sbeads-` prefix or directory name).
- [x] 3.2 Add retry mechanism for storage `CreateIssue` to gracefully retry with new random hash on unique constraint collision.
- [x] 3.3 Add unit tests in `internal/types/types_test.go` verifying ID format, randomness distribution, and prefix customization.

## 4. Cobra Vertical Slices (< 400 LOC ceiling per file)

- [x] 4.1 Implement `cmd/bd/context.go` (~40 LOC) encapsulating `Storage`, `ServiceDomain`, and database discovery.
- [x] 4.2 Initialize Cobra root command slice in `cmd/bd/root.go` (~70 LOC) with persistent flags, version flag, and `PersistentPreRunE` hook.
- [x] 4.3 Implement `create` command slice in `cmd/bd/create.go` (~60 LOC) supporting POSIX flags (`-t`/`--title`, `-p`/`--priority`, `-d`/`--description`, `--type`, `--blocked`).
- [x] 4.4 Implement `ready` command slice in `cmd/bd/ready.go` (~40 LOC) formatting ready issues.
- [x] 4.5 Implement `claim` command slice in `cmd/bd/claim.go` (~50 LOC) supporting POSIX flags (`-a`/`--assignee`, `-m`/`--agent`) and positional ID.
- [x] 4.6 Implement `close` command slice in `cmd/bd/close.go` (~45 LOC) supporting issue closing by ID or short hash.
- [x] 4.7 Implement `show` command slice in `cmd/bd/show.go` (~50 LOC) rendering issue details, status, and dependency edges.
- [x] 4.8 Implement `dep` command slice in `cmd/bd/dep.go` (~90 LOC) supporting `dep add` and `dep list` subcommands.
- [x] 4.9 Implement `version` command slice in `cmd/bd/version.go` (~25 LOC).

## 5. CLI Lifecycle & Application Wiring

- [x] 5.1 Refactor `cmd/bd/main.go` entrypoint to be a minimal runner (~25 LOC) executing `Execute()`.
- [x] 5.2 Ensure `PersistentPreRunE` handles DB initialization exclusively for stateful commands, bypassing `help` and `version`.
- [x] 5.3 Configure Cobra `SilenceUsage: true` and `SilenceErrors: true` for clean error reporting without usage dump on execution errors.

## 6. Unit & Slice Test Suites (Strict < 400 LOC Per Test File)

- [x] 6.1 Implement `cmd/bd/root_test.go` (~80 LOC) asserting root help, `-h`/`--help`, and unknown command handling.
- [x] 6.2 Implement `cmd/bd/create_test.go` (~90 LOC) asserting flag parsing, positional title fallback, and validation.
- [x] 6.3 Implement `cmd/bd/ready_test.go` (~80 LOC) asserting ready queue output, closed issue exclusion, and dependency blocking.
- [x] 6.4 Implement `cmd/bd/claim_close_test.go` (~100 LOC) asserting claim/close state transitions, agent assignment, and error conditions.
- [x] 6.5 Implement `cmd/bd/dep_test.go` (~90 LOC) asserting `dep add` and `dep list` subcommands and error cases.

## 7. Documentation & Contract Updates

- [x] 7.1 Update `README.md` with modernized Cobra CLI syntax, POSIX flag options, and 3-char ID examples.
- [x] 7.2 Update project skills documentation (`skills/sbeads/SKILL.md`) to reflect Cobra subcommands and flag parameters.

## 8. E2E Verification & Shell Command Scenarios

- [x] 8.1 Verify `sbeads create` creates issues with `sbeads-xxx` 3-character Base36 hash IDs.
- [x] 8.2 Verify `sbeads dep add <child-id> <parent-id>` blocks child issue in `sbeads ready` output.
- [x] 8.3 Verify `sbeads close <parent-id>` unblocks child issue in `sbeads ready` output.
- [x] 8.4 Verify `sbeads help <cmd>` and `<cmd> -h` produce consistent, accurate help text.

## 9. Quality Gates & Closeout

- [x] 9.1 Run full race-detector test suite (`go test -race ./...`) and ensure all tests pass.
- [x] 9.2 Compile release binary via `./build.sh --dev` without warnings.
- [x] 9.3 Perform line-of-code ceiling audit verifying no file in `cmd/bd/` exceeds 400 LOC.
- [x] 9.4 Validate OpenSpec change with `openspec validate cli-modernization-and-core-fixes`.
