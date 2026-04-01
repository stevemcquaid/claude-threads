# Go Conversion Implementation Plan
**Created:** 2026-03-31
**Spec:** [docs/superpowers/specs/2026-03-31-golang-conversion-design.md](../superpowers/specs/2026-03-31-golang-conversion-design.md)
**Status:** In Progress

> **Living document** — update this file whenever the plan changes, then run `/updoc`.

---

## Summary

Convert claude-threads from TypeScript/Bun to Go. The Go binary lives in `go/` and is a drop-in replacement: same CLI, same config.yaml, same sessions.json, same behavior. Implementation follows bottom-up TDD: translate tests first (RED), then implement (GREEN).

---

## Prerequisites

- [ ] Go 1.23+ installed
- [ ] `go/` directory initialized with `go.mod`
- [ ] All dependencies vendored or in module cache
- [ ] Docker available for integration tests

---

## Phase 1: Foundation (`utils/`, `config/`)

**Goal:** No-dependency packages that everything else builds on.

### Tasks
- [ ] Initialize `go/` module (`go mod init github.com/anneschuth/claude-threads`)
- [ ] Translate `src/utils/emoji.ts` → `internal/utils/emoji.go` (tests first)
- [ ] Translate `src/utils/format.ts` → `internal/utils/format.go` (tests first)
- [ ] Translate `src/utils/logger.ts` → `internal/utils/logger.go` (tests first)
- [ ] Translate `src/utils/colors.ts` → `internal/utils/colors.go` (tests first)
- [ ] Translate `src/utils/session-log.ts` → `internal/utils/sessionlog.go` (tests first)
- [ ] Translate `src/utils/keep-alive.ts` → `internal/utils/keepalive.go` (tests first)
- [ ] Translate `src/utils/pr-detector.ts` → `internal/utils/prdetector.go` (tests first)
- [ ] Translate `src/utils/websocket.ts` → `internal/utils/websocket.go` (tests first)
- [ ] Translate `src/utils/error-handler/` → `internal/utils/errorhandler.go` (tests first)
- [ ] Translate `src/utils/battery.ts` → `internal/utils/battery.go` (tests first)
- [ ] Translate `src/utils/uptime.ts` → `internal/utils/uptime.go` (tests first)
- [ ] Translate `src/config/types.ts` + `src/config/migration.ts` → `internal/config/config.go` (tests first)
- [ ] Port PII redaction (custom implementation, no direct Go equivalent)
- [ ] All utils tests GREEN

---

## Phase 2: Platform Types & Interfaces ✅ COMPLETE

**Goal:** Define the platform abstraction layer — interfaces and normalized types only, no implementations.

### Tasks
- [x] Translate `src/platform/types.ts` → `internal/platform/types.go`
- [x] Translate `src/platform/client.ts` → `internal/platform/client.go` (interface)
- [x] Translate `src/platform/formatter.ts` → `internal/platform/formatter.go` (interface)
- [x] Translate `src/platform/permission-api.ts` → `internal/platform/permissionapi.go` (interface)
- [x] Translate `src/platform/utils.ts` → `internal/platform/utils.go` (tests first)
- [x] Write mock implementations of PlatformClient for use in tests

---

## Phase 3: Mattermost Client

**Goal:** Full Mattermost WebSocket + REST implementation.

### Tasks
- [x] Translate `src/platform/mattermost/types.ts` → `internal/platform/mattermost/types.go`
- [x] Translate `src/platform/mattermost/formatter.ts` → `internal/platform/mattermost/formatter.go` (tests first)
- [x] Translate `src/mattermost/api.ts` → `internal/platform/mattermost/api.go` (tests first — implemented in client.go)
- [x] Translate `src/platform/mattermost/client.ts` → `internal/platform/mattermost/client.go` (tests first)
- [x] WebSocket reconnection logic
- [x] File upload/download
- [x] All Mattermost tests GREEN

---

## Phase 4: Slack Client

**Goal:** Full Slack Socket Mode + Web API implementation.

### Tasks
- [ ] Translate `src/platform/slack/types.ts` → `internal/platform/slack/types.go`
- [ ] Translate `src/platform/slack/formatter.ts` → `internal/platform/slack/formatter.go` (tests first)
- [ ] Translate `src/platform/slack/permission-api.ts` → `internal/platform/slack/permissionapi.go` (tests first)
- [ ] Translate `src/platform/slack/client.ts` → `internal/platform/slack/client.go` (tests first)
- [ ] Socket Mode connection + reconnection
- [x] File upload/download
- [ ] All Slack tests GREEN

---

## Phase 5: Persistence

**Goal:** Session store + thread logger with identical on-disk format.

### Tasks
- [ ] Translate `src/persistence/session-store.ts` → `internal/persistence/sessionstore.go` (tests first)
  - `PersistedSession` struct must match TypeScript interface exactly
  - Port all migration logic from TypeScript
  - Defensive defaults for all fields
- [ ] Translate `src/persistence/thread-logger.ts` → `internal/persistence/threadlogger.go` (tests first)
- [ ] File permissions: `0600` on all written files
- [ ] All persistence tests GREEN

---

## Phase 6: Claude CLI Management

**Goal:** Subprocess management for Claude CLI with stream-json I/O.

### Tasks
- [ ] Translate `src/claude/types.ts` → `internal/claude/types.go`
- [ ] Translate `src/claude/version-check.ts` → `internal/claude/versioncheck.go` (tests first)
- [ ] Translate `src/claude/cli.ts` → `internal/claude/cli.go` (tests first)
  - Spawn subprocess with `os/exec`
  - Stream stdin/stdout pipes
  - MCP config injection
  - Environment setup
- [ ] Translate `src/claude/quick-query.ts` → `internal/claude/quickquery.go` (tests first)
- [ ] All Claude tests GREEN

---

## Phase 7: MCP Permission Server

**Goal:** MCP permission server using official Go SDK.

### Tasks
- [ ] Evaluate `github.com/modelcontextprotocol/go-sdk` maturity; switch to `mark3labs/mcp-go` if needed (document decision in spec)
- [ ] Translate `src/mcp/permission-server.ts` → `internal/mcp/permissionserver.go` (tests first)
  - Stdio transport
  - `permission_prompt` tool registration
  - Platform WebSocket for reaction listening
  - allow/deny response
- [ ] Translate `src/platform/permission-api-factory.ts` → `internal/platform/permissionapifactory.go`
- [ ] All MCP tests GREEN

---

## Phase 8: Git Worktree Management

**Goal:** Git worktree operations via subprocess.

### Tasks
- [ ] Translate `src/git/` → `internal/git/` (tests first)
  - Worktree create, list, remove, cleanup
  - Orphaned worktree detection
- [ ] All git tests GREEN

---

## Phase 9: Commands

**Goal:** User command parser and handler (!stop, !invite, !kick, etc.).

### Tasks
- [ ] Translate `src/commands/` → `internal/commands/` (tests first)
  - Command registry and parser
  - All command handlers: `!stop`, `!escape`, `!kill`, `!invite`, `!kick`, `!cd`, `!permissions`, `!help`, `!bug`
  - Help generator
  - Prompt generator
- [ ] All command tests GREEN

---

## Phase 10: Operations Layer

**Goal:** The brain — event transformation, message batching, all executors.

### Tasks
- [ ] Translate `src/operations/types.ts` → `internal/operations/types.go`
- [ ] Translate `src/operations/transformer.ts` → `internal/operations/transformer.go` (tests first)
- [ ] Translate `src/operations/post-helpers/` → `internal/operations/posthelpers/` (tests first)
- [ ] Translate `src/operations/streaming/handler.ts` → `internal/operations/streaming/handler.go` (tests first)
- [ ] Translate `src/operations/content-breaker.ts` → `internal/operations/contentbreaker.go` (tests first)
- [ ] Translate each executor (tests first each):
  - `executors/content.ts` → `internal/operations/executors/content.go`
  - `executors/task-list.ts` → `internal/operations/executors/tasklist.go`
  - `executors/question-approval.ts` → `internal/operations/executors/questionapproval.go`
  - `executors/prompt.ts` → `internal/operations/executors/prompt.go`
  - `executors/subagent.ts` → `internal/operations/executors/subagent.go`
  - `executors/message-approval.ts` → `internal/operations/executors/messageapproval.go`
  - `executors/bug-report.ts` → `internal/operations/executors/bugreport.go`
  - `executors/system.ts` → `internal/operations/executors/system.go`
  - `executors/worktree-prompt.ts` → `internal/operations/executors/worktreeprompt.go`
- [ ] Translate `src/operations/message-manager.ts` → `internal/operations/messagemanager.go` (tests first)
- [ ] Translate remaining operations submodules (tests first each):
  - `bug-report/`, `context-prompt/`, `events/`, `monitor/`, `plugin/`
  - `session-context/`, `side-conversation/`, `sticky-message/`
  - `suggestions/`, `tool-formatters/`, `worktree/`
  - `post-tracker.ts`
- [ ] All operations tests GREEN (~893 test cases in message-manager alone)

---

## Phase 11: Session Management

**Goal:** Session state machine, registry, lifecycle, timers.

### Tasks
- [ ] Translate `src/session/types.ts` → `internal/session/types.go`
- [ ] Translate `src/session/registry.ts` → `internal/session/registry.go` (tests first)
- [ ] Translate `src/session/timer-manager.ts` → `internal/session/timermanager.go` (tests first)
- [ ] Translate `src/session/lifecycle.ts` → `internal/session/lifecycle.go` (tests first)
- [ ] Translate `src/session/manager.ts` → `internal/session/manager.go` (tests first)
- [ ] Translate `src/cleanup/` → `internal/cleanup/` (tests first)
- [ ] All session tests GREEN

---

## Phase 12: Auto-Update

**Goal:** Version check and self-update via GitHub releases.

### Tasks
- [ ] Translate `src/auto-update/` → `internal/autoupdate/` (tests first)
  - GitHub releases API check
  - Download + replace binary
  - Scheduler (hourly check)
- [ ] All auto-update tests GREEN

---

## Phase 13: Terminal UI (Bubble Tea)

**Goal:** Polished terminal UI matching TypeScript Ink UI functionality.

### Tasks
- [ ] Define lipgloss styles (`internal/ui/styles/`)
  - Color scheme (match TypeScript UI palette)
  - Border styles, padding, margins
- [ ] Translate `src/ui/types.ts` → `internal/ui/types.go`
- [ ] Implement Bubble Tea model (`internal/ui/app.go`)
  - `Model`, `Init()`, `Update()`, `View()` 
- [ ] Implement components (tests where applicable):
  - `SessionListPanel` — session list with status, cost, uptime
  - `SessionDetailPanel` — live output stream
  - `HeaderBar` — version, platform status, Claude version
  - `FooterBar` — shortcuts, battery
  - `TaskListView` — inline task display
  - `PermissionPromptOverlay`
- [ ] Implement keyboard handling (match TypeScript keybindings)
- [ ] Headless mode (no TUI, for testing)
- [ ] Visual polish pass — consistent colors, clear hierarchy, smooth redraws
- [ ] Translate `src/statusline/` → `internal/statusline/` (status line writer)

---

## Phase 14: Main Entry Point

**Goal:** Wire everything together into a runnable binary.

### Tasks
- [ ] Translate `src/index.ts` → `cmd/claude-threads/main.go`
  - Cobra CLI setup (match all TypeScript flags exactly)
  - Config loading
  - Platform client initialization
  - Session manager startup
  - TUI startup
  - Graceful shutdown
- [ ] Translate `src/message-handler.ts` → `internal/messagehandler.go` (tests first)
- [ ] Translate `src/onboarding.ts` → `internal/onboarding/onboarding.go`
  - Interactive setup wizard using `charmbracelet/huh`
- [ ] Translate `src/changelog.ts` + `src/version.ts` → `internal/version/`
- [ ] Binary produces identical output to TypeScript version for all commands
- [ ] `go build ./cmd/claude-threads` succeeds

---

## Phase 15: Integration Tests

**Goal:** All ~120 integration test scenarios passing in Go.

### Tasks
- [ ] Port Docker test infrastructure (`tests/integration/fixtures/`)
  - Mock Claude CLI (same behavior as TypeScript mock)
  - Mattermost test setup helpers
  - Slack mock server
- [ ] Translate all integration test suites (`tests/integration/suites/`):
  - `connection.test.ts`, `messaging.test.ts`, `reactions.test.ts`
  - `session-lifecycle.test.ts`, `session-resume.test.ts`, `session-kill.test.ts`
  - `session-limits.test.ts`, `session-errors.test.ts`, `session-permissions.test.ts`
  - `session-commands.test.ts`, `session-context-prompt.test.ts`
  - `session-questions.test.ts`, `session-tasks.test.ts`
  - `session-worktree.test.ts`, `session-sticky.test.ts`
  - `session-plan-approval.test.ts`, `session-multi-user.test.ts`
  - `session-update-reaction.test.ts`
  - `file-attachments.test.ts`, `slack-file-attachments.test.ts`
  - `platform-example.test.ts`
- [ ] All integration tests GREEN
- [ ] CI workflow updated to run Go integration tests

---

## Phase 16: Final Polish

### Tasks
- [ ] `go vet ./...` — no warnings
- [ ] `golangci-lint run` — no lint errors
- [ ] Cross-platform build: Linux + macOS (+ Windows if feasible)
- [ ] Update root `README.md` to document Go binary
- [ ] Update `CLAUDE.md` with Go development commands
- [ ] Add `go/Makefile` with standard targets: `build`, `test`, `lint`, `integration`
- [ ] Verify drop-in compatibility: run Go binary against live Mattermost/Slack, confirm behavior matches TypeScript version

---

## Progress Tracking

| Phase | Status | Notes |
|-------|--------|-------|
| 1: Foundation | ✅ Complete | All utils + config packages, all tests GREEN, go vet clean |
| 2: Platform Types | ⬜ Not started | |
| 3: Mattermost | ⬜ Not started | |
| 4: Slack | ⬜ Not started | |
| 5: Persistence | ⬜ Not started | |
| 6: Claude CLI | ⬜ Not started | |
| 7: MCP Server | ⬜ Not started | |
| 8: Git Worktree | ⬜ Not started | |
| 9: Commands | ⬜ Not started | |
| 10: Operations | ⬜ Not started | Largest phase (~893 test cases) |
| 11: Session Mgmt | ⬜ Not started | |
| 12: Auto-Update | ⬜ Not started | |
| 13: TUI | ⬜ Not started | Bubble Tea |
| 14: Main Entry | ⬜ Not started | |
| 15: Integration | ⬜ Not started | |
| 16: Polish | ⬜ Not started | |

---

## Change Log

| Date | Change | Reason |
|------|--------|--------|
| 2026-03-31 | Initial plan created | Go rewrite project start |
| 2026-04-01 | Phase 1 complete | All utils + config tests passing, go vet clean |
