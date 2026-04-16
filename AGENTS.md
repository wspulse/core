# AGENTS.md — wspulse/core

This file is the entry point for all AI coding agents (GitHub Copilot, Codex,
Cursor, Claude, etc.). Full working rules are in
`.github/copilot-instructions.md` — read it completely before
making any changes.

---

## Quick Reference

**Module**: `github.com/wspulse/core` | **Package**: `wspulse`

**Key files**:

- `frame.go` — `Frame` struct, sentinel errors
- `codec.go` — `Codec` interface, `JSONCodec` implementation
- `protocol.go` — `MessageType` and `StatusCode` wire protocol types (RFC 6455)
- `router/router.go` — `Router`, `Use`, `On`, `Dispatch`
- `router/context.go` — `Context`, `HandlerFunc`, `HandlersChain`, `abortIndex`
- `router/conn.go` — `Connection` interface
- `router/recovery.go` — `Recovery()` middleware

**Pre-commit gate**: `make check` (fmt → lint → test)

---

## Non-negotiable Rules

1. **Read before write** — read the target file before any edit.
2. **Zero production dependencies** — core's production code must only depend on Go stdlib. Test dependencies are permitted when justified.
3. **No breaking changes without version bump.**
4. **No hardcoded secrets.**
5. **Minimal changes** — one concern per edit; no drive-by refactors.

---

## Session Protocol

> `doc/local/` is git-ignored. Never commit files under it.

- **Start of session**: read `doc/local/ai-learning.md` in full (create with header if missing) and check `doc/local/plan/` for any in-progress plan.
- **Feature work**: save plan to `doc/local/plan/<feature-name>.md` before starting.
- **End of session**: append at least one entry to `doc/local/ai-learning.md` — **mandatory even if no mistakes were made**. An empty file proves the session protocol was ignored.
  Format: `Date` / `Issue or Learning` / `Root Cause` / `Prevention Rule`.
