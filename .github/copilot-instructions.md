# Copilot Instructions — wspulse/core

## Project Overview

wspulse/core provides the **shared types** used across the wspulse WebSocket ecosystem: `Frame`, `Codec`, `JSONCodec`, and sentinel errors. Module path: `github.com/wspulse/core`. Package name: `wspulse`. This module has **zero external dependencies** (stdlib only).

## Architecture

- **`frame.go`** — `Frame` struct (the minimal transport unit), WebSocket message type constants (`TextMessage`, `BinaryMessage`), and shared sentinel errors (`ErrConnectionClosed`, `ErrSendBufferFull`).
- **`codec.go`** — `Codec` interface (`Encode`, `Decode`, `FrameType`), default `JSONCodec` implementation, and the unexported `wireFrame`/`jsonCodec` types.
- **`router/`** — Gin-style event router (`Router`, `Context`, `Connection`, `HandlerFunc`, `Recovery`). Dispatches inbound `Frame` values by `Frame.Event` through a global middleware + per-event handler chain. Lazy chain building, `sync.Pool`-backed `Context` recycling, atomic fast path for steady-state dispatches.

## Development Workflow

```bash
make fmt        # format (gofmt + goimports)
make lint       # vet + golangci-lint
make test       # race detector, count=3
make check      # fmt + lint + test (pre-commit gate)
make bench      # benchmarks with memory stats
make test-cover # coverage report → coverage.html
make tidy       # tidy module dependencies
```

## Conventions

- **Go style**: `gofmt`/`goimports`, snake_case filenames, GoDoc on all public symbols, `if err != nil` error handling (no runtime panics; setup-time programmer-error panics are allowed — see **Panic policy** critical rule), secrets from env vars only.
- **Naming**:
  - **Interface names** must use full words — no abbreviations. Write `Connection`, not `Conn`; `Configuration`, not `Cfg`; `Manager`, not `Mgr`.
  - **Variable and parameter names** follow standard Go style: single-letter or short receivers (`r` for `*Router`, `c` for `*Context`), idiomatic short names for local scope (`conn`, `fn`, `err`, `ok`, `n`, `i`, `v`), and descriptive names for package-level identifiers.
- **Markdown**: no emojis in documentation files.
- **Git**:
  - Follow the commit message rules in [commit-message-instructions.md](instructions/commit-message-instructions.md).
  - All commit messages in English.
  - Each commit must represent exactly one logical change.
  - Before every commit, run `make check` (runs fmt → lint → test in order).
  - **Branch strategy**: never push directly to `develop` or `main`.
    - `feat/<name>` or `feature/<name>` — new feature
    - `refactor/<name>` — restructure without behaviour change
    - `bugfix/<name>` — bug fix
    - `fix/<name>` — quick fix (e.g. config, docs, CI)
    - `chore/<name>` — maintenance, CI/CD, dependencies, docs
    - CI triggers on all branch prefixes above and on PRs targeting `main`/`develop`. Tags do **not** trigger CI (the tag is created after CI already passed). Open a PR into `develop`; `develop` requires status checks to pass.
- **Tests**: co-located with source (`_test.go`). Cover happy path and at least one error path. Required for new public functions.
  - **Test-first for bug fixes**: **mandatory** — see Critical Rule 6 for the required step-by-step procedure. Do not touch production code without a prior failing test.
  - **Benchmarks**: changes to encoding or frame allocation must include a benchmark. Verify with `make bench`.
- **API compatibility**:
  - Exported symbols are a public contract. Changing or removing any exported identifier is a breaking change requiring a major version bump.
  - Adding a method to an exported interface breaks all external implementations — treat it as a breaking change.
  - Mark deprecated symbols with `// Deprecated: use Xxx instead.` before removal.
- **Error format**: wrap errors as `fmt.Errorf("wspulse: <context>: %w", err)`; define sentinel errors as `errors.New("wspulse: <description>")`.
- **File encoding**: all files must be UTF-8 without BOM. Do not use any other encoding.

## Critical Rules

1. **Read before write** — always read the target file and relevant docs fully before editing.
2. **Minimal changes** — one concern per edit; no drive-by refactors.
3. **No hardcoded secrets** — all configuration via environment variables.
4. **Zero external dependencies** — core must only depend on Go stdlib. Any change introducing an external dependency must be explicitly justified and approved.
5. **No breaking changes without version bump** — never rename, remove, or change the signature of an exported symbol without bumping the major version. When unsure, add alongside the old symbol and deprecate.
6. **STOP — test first, fix second** — when a bug is discovered or reported, do NOT touch production code until a failing test exists. Follow this exact sequence without skipping or reordering:
    1. Write a failing test that reproduces the bug.
    2. Run the test and confirm it **fails** (proving the test actually catches the bug).
    3. Fix the production code.
    4. Run the test again and confirm it **passes**.
    5. Run `make check` to verify nothing else broke.
    6. If you are about to edit production code and no failing test exists yet — stop and go back to step 1.
7. **STOP — before every commit, verify this checklist:**
    1. Run `make check` (fmt → lint → test) and confirm it passes. Skip if the commit contains only non-code changes (e.g. documentation, comments, Markdown).
    2. Run GitHub Copilot code review (`github.copilot.chat.review.changes`) on the working-tree diff and resolve every comment before proceeding.
    3. Commit message follows [commit-message-instructions.md](instructions/commit-message-instructions.md): correct type, subject ≤ 50 chars, numbered body items stating reason → change.
    4. This commit contains exactly one logical change — no unrelated modifications.
    5. If any item fails — fix it before committing.
8. **Accuracy** — if you have questions or need clarification, ask the user. Do not make assumptions without confirming.
9. **Language consistency** — when the user writes in Traditional Chinese, respond in Traditional Chinese; otherwise respond in English.
10. **Panic policy — fail early, never at steady-state runtime** — Enforce errors at the earliest possible phase:
    1. Prefer compile-time enforcement via the type system.
    2. **Setup-time programmer errors** (nil handler, empty event name, duplicate registration, invalid option): `panic`. These indicate a caller logic bug; crashing at startup is correct — the process should never start accepting traffic with a misconfigured router or server.
    3. **Steady-state runtime** (`Dispatch`, `Send`, `Close`, reconnect loops, and any code that runs after startup completes): return `error`, never `panic`.

## Session Protocol

> Files under `doc/local/` are git-ignored and must **never** be committed.
> This applies to both plan files and `doc/local/ai-learning.md`.

- **At the start of every session**: check whether `doc/local/plan/` contains
  an in-progress plan for the current task, and read `doc/local/ai-learning.md`
  (if it exists) to recall past mistakes and techniques before writing any code.
- **Plan mode**: when implementing a new feature or multi-file fix, save a plan
  to `doc/local/plan/<feature-name>.md` before starting. Keep it updated with
  completed steps and any plan changes throughout the session.
- **AI learning log**: at the end of a session where mistakes were made or
  reusable techniques were discovered, append a short entry to
  `doc/local/ai-learning.md`. Entry format:
  `Date` / `Issue or Learning` / `Root Cause` / `Prevention Rule`.
  Append only — never overwrite existing entries.
