# Go Rewrite Design Spec
**Date:** 2026-03-31
**Status:** Approved
**Project:** claude-threads TypeScript → Go conversion

---

## Overview

Convert the claude-threads multi-platform Claude Code bot from TypeScript/Bun to Go. The Go binary lives in `go/` alongside the existing TypeScript source. The result is a **drop-in replacement**: identical CLI flags, identical `config.yaml` format, identical `sessions.json` schema, identical runtime behavior.

## Scope

Full feature parity with the TypeScript codebase:
- Multi-platform support: Mattermost (WebSocket + REST) and Slack (Socket Mode + Web API)
- Claude CLI subprocess management with MCP permission server
- Session lifecycle state machine with persistence and resume
- Interactive terminal UI (Bubble Tea, must look polished)
- All user commands: `!stop`, `!escape`, `!invite`, `!kick`, `!cd`, `!permissions`, `!kill`, `!help`
- Reaction-based controls: plan approval, question answering, permission prompts
- Git worktree integration
- Session collaboration (invite/kick, message approval)
- Auto-update checker and installer
- Thread context prompts, sticky channel messages
- Task list display with live updates
- Code diffs, file previews, syntax highlighting

## Non-Goals

- No new features beyond TypeScript parity
- No changes to CLI flags, config schema, or persistence format
- No public library API (all packages under `internal/`)

---

## Package Structure

```
go/
├── cmd/claude-threads/         # main.go — CLI entry point
├── internal/
│   ├── config/                 # YAML config loading + migration (go-yaml v3)
│   ├── platform/               # PlatformClient interface + normalized types
│   │   ├── mattermost/         # Mattermost WebSocket + REST implementation
│   │   └── slack/              # Slack Socket Mode + Web API implementation
│   ├── session/                # Session state machine, registry, lifecycle, timers
│   ├── operations/             # Event transformer, MessageManager, executors
│   │   └── executors/          # content, tasklist, question, prompt, subagent, approval, bugreport
│   ├── claude/                 # Claude CLI subprocess mgmt, version check, quick query
│   ├── mcp/                    # MCP permission server (official go-sdk)
│   ├── persistence/            # sessions.json read/write, thread logger
│   ├── git/                    # Worktree management (git subprocess)
│   ├── commands/               # Command parser + handler
│   ├── autoupdate/             # Version check + GitHub releases installer
│   ├── ui/                     # Bubble Tea TUI (App model, components, layouts)
│   │   ├── components/         # SessionList, SessionDetail, Header, Footer, etc.
│   │   ├── layouts/            # SplitPanel, Panel, RootLayout
│   │   └── styles/             # lipgloss style definitions
│   └── utils/                  # emoji, logger, format, colors, keep-alive, websocket
├── tests/
│   ├── integration/            # Integration tests (Docker: real Mattermost + mock Slack)
│   │   ├── suites/             # ~120 test scenarios mirroring TypeScript integration tests
│   │   └── fixtures/           # Mock servers, Docker helpers, test harness
│   └── unit/                   # Cross-package unit tests
├── go.mod
└── go.sum
```

Go module name: `github.com/anneschuth/claude-threads`

---

## Core Interfaces

### Platform Client

```go
// internal/platform/client.go
type Client interface {
    PostMessage(ctx context.Context, channelID, threadID, text string) (string, error)
    UpdateMessage(ctx context.Context, postID, text string) error
    DeleteMessage(ctx context.Context, postID string) error
    AddReaction(ctx context.Context, postID, emoji string) error
    RemoveReaction(ctx context.Context, postID, emoji string) error
    GetUser(ctx context.Context, userID string) (*User, error)
    GetFileContent(ctx context.Context, fileID string) ([]byte, error)
    Subscribe(ctx context.Context) (<-chan Event, error)
    PlatformID() string
    Close() error
}

type Formatter interface {
    FormatCode(code, language string) string
    FormatDiff(diff string) string
    FormatBold(text string) string
    FormatItalic(text string) string
    MaxMessageLength() int
}
```

### Session

```go
// internal/session/types.go
type Session struct {
    ID              string          // "platformID:threadID"
    PlatformID      string
    ThreadID        string
    StartedBy       string
    WorkingDir      string
    AllowedUsers    map[string]bool
    SkipPermissions bool
    Claude          *claude.CLI
    MessageManager  *operations.MessageManager
    ClaudeSessionID string          // for resume
    IsResumed       bool
    CreatedAt       time.Time
    LastActivityAt  time.Time
}
```

### Operations

```go
// internal/operations/types.go
type OperationType string

const (
    OpAppendContent  OperationType = "append_content"
    OpFlush          OperationType = "flush"
    OpTaskList       OperationType = "task_list"
    OpQuestion       OperationType = "question"
    OpApproval       OperationType = "approval"
    OpSystemMessage  OperationType = "system_message"
    OpSubagent       OperationType = "subagent"
    OpStatusUpdate   OperationType = "status_update"
    OpLifecycle      OperationType = "lifecycle"
)

type Operation interface {
    Type() OperationType
}
```

---

## Event Flow & Concurrency

```
Platform WebSocket/Socket Mode
    │
    ▼ goroutine per platform
platform.Event
    │
    ▼
SessionManager.HandleEvent()
    │ (mutex-protected registry lookup)
    ▼
session.Session
    │
    ▼
claude.CLI subprocess (stdin/stdout pipes)
    │ stdout goroutine reads stream-json
    ▼
operations.Transformer → []Operation
    │
    ▼
operations.MessageManager
    │
    ├── ContentExecutor      (owns streaming buffer state)
    ├── TaskListExecutor     (owns task list state)
    ├── QuestionExecutor     (owns pending question state)
    ├── PromptExecutor       (owns context/worktree/update prompts)
    ├── SubagentExecutor     (owns active subagent tracking)
    ├── MessageApprovalExecutor (owns unauthorized message queue)
    └── BugReportExecutor    (owns bug report flow)
    │
    ▼
platform.Client (post/update/react)
```

**Concurrency rules:**
- Each `Session` is owned by a single goroutine (no shared mutable state across sessions)
- `SessionManager` uses a `sync.RWMutex` for the session registry
- Each session spawns: (1) Claude stdout reader goroutine, (2) content flush ticker goroutine, (3) timeout timer
- Platform event fan-out: one WebSocket reader goroutine per platform → sends to per-session channels
- MCP permission server: separate goroutine per session, communicates via stdio with Claude CLI

---

## Terminal UI (Bubble Tea)

The TUI mirrors the TypeScript Ink UI in layout and information density.

**Components:**
- `SessionListPanel` — left panel: active sessions with status indicators, cost, uptime
- `SessionDetailPanel` — right panel: selected session's live output stream
- `HeaderBar` — bot name, version, platform connections, Claude CLI version
- `FooterBar` — keyboard shortcuts, battery status
- `StatusLine` — per-session status (idle/running/waiting for approval)
- `TaskListView` — inline task list within session detail
- `PermissionPrompt` — permission request overlay

**Styling:** `charmbracelet/lipgloss` for colors, borders, padding. Must look polished — consistent color scheme, clear visual hierarchy, smooth updates.

---

## Dependency Mapping

| TypeScript | Go |
|---|---|
| `commander` | `spf13/cobra` |
| `js-yaml` | `gopkg.in/yaml.v3` |
| `zod` | `go-playground/validator/v10` |
| `ws` WebSocket | `gorilla/websocket` |
| `ink` / React | `charmbracelet/bubbletea` |
| lipgloss equivalent | `charmbracelet/lipgloss` |
| `prompts` interactive | `charmbracelet/huh` |
| `semver` | `github.com/Masterminds/semver/v3` |
| `diff` library | `github.com/sergi/go-diff/diffmatchpatch` |
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` |
| `update-notifier` | Custom: `net/http` → GitHub releases API |
| `@redactpii/node` | Custom: regex-based PII scrubber |
| `express-rate-limit` | `golang.org/x/time/rate` |
| `hono` HTTP | `net/http` standard library |
| `bun` subprocess | `os/exec` standard library |
| `yauzl` / `yazl` ZIP | `archive/zip` standard library |

---

## Testing Strategy

**Approach:** TDD — translate tests first (RED), then implement (GREEN).

**Unit tests** (per package):
- `github.com/stretchr/testify/assert` + `require` for assertions
- `github.com/stretchr/testify/mock` for interface mocks (PlatformClient, ClaudeCLI)
- Mirror TypeScript test structure: same test groupings, same coverage targets
- Target: ~2,200 test cases translated

**Integration tests** (`tests/integration/`):
- Same Docker setup: real Mattermost instance + mock Slack server
- ~120 test scenarios covering: session lifecycle, commands, reactions, multi-user, persistence, worktrees, file attachments, platform-specific behavior
- Go test binary with `TestMain` for Docker setup/teardown

**Build order per package:**
1. Write `*_test.go` — run — RED
2. Write implementation — run — GREEN
3. Refactor — run — still GREEN

---

## Implementation Order (Bottom-Up)

Packages are implemented in dependency order:

| Phase | Packages | Notes |
|---|---|---|
| 1 | `utils/`, `config/` | No external deps, foundation for everything |
| 2 | `platform/` (types + interfaces) | Interface definitions only, no implementations |
| 3 | `platform/mattermost/` | WebSocket + REST client |
| 4 | `platform/slack/` | Socket Mode + Web API client |
| 5 | `persistence/` | sessions.json, thread logger |
| 6 | `claude/` | CLI subprocess management, version check |
| 7 | `mcp/` | Permission server (go-sdk) |
| 8 | `git/` | Worktree management |
| 9 | `commands/` | Command parser + handler |
| 10 | `operations/` | Transformer + MessageManager + all executors |
| 11 | `session/` | State machine, registry, lifecycle, timers |
| 12 | `autoupdate/` | Version check + installer |
| 13 | `ui/` | Bubble Tea TUI |
| 14 | `cmd/claude-threads/` | Main entry point, wiring |
| 15 | Integration tests | Docker-based end-to-end |

---

## Backward Compatibility Requirements

- CLI flags: identical to TypeScript version (same names, same defaults)
- `config.yaml`: same schema, same field names, same defaults
- `sessions.json`: `PersistedSession` struct fields must match TypeScript interface exactly
- All defensive defaults preserved (`?? fallbacks` → Go zero-value checks)
- Session migration logic must be ported from `session-store.ts`

---

## Spec Update Policy

This document is a living spec. Whenever the implementation plan changes:
1. Update this file with the change and reason
2. Run `/updoc` to propagate changes to related documentation
3. Commit the updated spec alongside the code change

---

## Open Questions

- Go module name: confirm `github.com/anneschuth/claude-threads` or different?
- MCP Go SDK maturity: the official SDK (`modelcontextprotocol/go-sdk`) is newer — evaluate against `mark3labs/mcp-go` during Phase 7 implementation; switch if the official SDK is missing needed features
- PII redaction: `@redactpii/node` has no direct Go equivalent — custom implementation scope TBD during Phase 1
