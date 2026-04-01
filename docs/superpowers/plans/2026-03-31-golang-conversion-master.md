# Go Conversion — Master Plan Index

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert claude-threads from TypeScript/Bun to Go, producing a drop-in replacement binary in `go/`.

**Architecture:** Bottom-up TDD — each phase translates tests first (RED), then implements (GREEN). Phases are independent and build on each other in dependency order.

**Tech Stack:** Go 1.23+, Bubble Tea TUI, Testify, gorilla/websocket, official MCP Go SDK, cobra CLI, go-yaml v3

**Spec:** [docs/superpowers/specs/2026-03-31-golang-conversion-design.md](../../superpowers/specs/2026-03-31-golang-conversion-design.md)

> **Living document** — update whenever phases are added, reordered, or split. Run `/updoc` after changes.

---

## Phase Plans

Each phase has its own detailed plan file. Start the next phase only after the previous is complete (all tests GREEN).

| Phase | Plan File | Status | Description |
|-------|-----------|--------|-------------|
| 0 | [phase-00-module-init.md](./2026-03-31-golang-phase-00-module-init.md) | ✅ | Go module init, go.mod, directory scaffold |
| 1 | [phase-01-foundation.md](./2026-03-31-golang-phase-01-foundation.md) | ✅ | `utils/`, `config/` — no external deps |
| 2 | [phase-02-platform-types.md](./2026-03-31-golang-phase-02-platform-types.md) | ✅ | Platform interfaces + normalized types |
| 3 | [phase-03-mattermost.md](./2026-03-31-golang-phase-03-mattermost.md) | ✅ | Mattermost WebSocket + REST client |
| 4 | [phase-04-slack.md](./2026-03-31-golang-phase-04-slack.md) | ⬜ | Slack Socket Mode + Web API client |
| 5 | [phase-05-persistence.md](./2026-03-31-golang-phase-05-persistence.md) | ⬜ | sessions.json, thread logger |
| 6 | [phase-06-claude-cli.md](./2026-03-31-golang-phase-06-claude-cli.md) | ⬜ | Claude CLI subprocess management |
| 7 | [phase-07-mcp.md](./2026-03-31-golang-phase-07-mcp.md) | ⬜ | MCP permission server |
| 8 | [phase-08-git.md](./2026-03-31-golang-phase-08-git.md) | ⬜ | Git worktree management |
| 9 | [phase-09-commands.md](./2026-03-31-golang-phase-09-commands.md) | ⬜ | User command parser + handlers |
| 10 | [phase-10-operations.md](./2026-03-31-golang-phase-10-operations.md) | ⬜ | Operations layer (largest phase) |
| 11 | [phase-11-session.md](./2026-03-31-golang-phase-11-session.md) | ⬜ | Session state machine |
| 12 | [phase-12-autoupdate.md](./2026-03-31-golang-phase-12-autoupdate.md) | ⬜ | Auto-update checker + installer |
| 13 | [phase-13-ui.md](./2026-03-31-golang-phase-13-ui.md) | ⬜ | Bubble Tea TUI |
| 14 | [phase-14-main.md](./2026-03-31-golang-phase-14-main.md) | ⬜ | Main entry point, CLI wiring |
| 15 | [phase-15-integration.md](./2026-03-31-golang-phase-15-integration.md) | ⬜ | Integration tests (Docker) |
| 16 | [phase-16-polish.md](./2026-03-31-golang-phase-16-polish.md) | ⬜ | Lint, cross-platform build, docs |

---

## Change Log

| Date | Change | Reason |
|------|--------|--------|
| 2026-03-31 | Master plan created with 16 phases | Initial Go rewrite planning |
| 2026-04-01 | Phases 0 and 1 marked complete | All foundation tests GREEN, go vet clean |
| 2026-04-01 | Phase 2 marked complete | Platform types, interfaces, utils (39 tests GREEN), MockPlatformClient |
| 2026-04-01 | Phase 3 marked complete | MattermostClient, BasePlatformClient, MattermostFormatter, all tests GREEN |
