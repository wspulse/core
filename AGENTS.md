# AGENTS.md — wspulse/core

This file is the entry point for all AI coding agents (GitHub Copilot, Codex,
Cursor, Claude, etc.). Full working rules are in
`.github/copilot-instructions.md` — read it completely before
making any changes.

---

## Quick Reference

**Module**: `github.com/wspulse/core` | **Package**: `wspulse`

**Key files**:

- `frame.go` — `Frame` struct, message type constants, sentinel errors
- `codec.go` — `Codec` interface, `JSONCodec` implementation

**Pre-commit gate**: `make check` (fmt → lint → test)

---

## Non-negotiable Rules

1. **Read before write** — read the target file before any edit.
2. **Zero external dependencies** — core must only depend on Go stdlib.
3. **No breaking changes without version bump.**
4. **No hardcoded secrets.**
5. **Minimal changes** — one concern per edit; no drive-by refactors.

---

## Session Protocol

> `doc/local/` is git-ignored. Never commit files under it.

- **Start of session**: read `doc/local/ai-learning.md` (if present) and check
  `doc/local/plan/` for any in-progress plan.
- **Feature work**: save plan to `doc/local/plan/<feature-name>.md` first.
- **End of session**: append mistakes/learnings to `doc/local/ai-learning.md`.
  Format: `Date` / `Issue or Learning` / `Root Cause` / `Prevention Rule`.
