# Copilot Instructions — wspulse/core

## Project Overview

wspulse/core provides the **shared types** used across the wspulse WebSocket ecosystem: `Frame`, `Codec`, `JSONCodec`, and sentinel errors. Module path: `github.com/wspulse/core`. Package name: `wspulse`. This module has **zero external dependencies** (stdlib only).

## Architecture

- **`frame.go`** — `Frame` struct (the minimal transport unit), WebSocket message type constants (`TextMessage`, `BinaryMessage`), and shared sentinel errors (`ErrConnectionClosed`, `ErrSendBufferFull`).
- **`codec.go`** — `Codec` interface (`Encode`, `Decode`, `FrameType`), default `JSONCodec` implementation, and the unexported `wireFrame`/`jsonCodec` types.

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

- **Go style**: `gofmt`/`goimports`, snake_case filenames, GoDoc on all public symbols, `if err != nil` error handling (never `panic`), secrets from env vars only.
- **Naming — readability is the highest priority**:
  - Use full words for all identifiers. Code is AI-generated; there is no excuse for cryptic names.
  - **Allowed abbreviations** (universally recognized only): ID, URL, HTTP, API, JSON, Msg, Err, Ctx, Buf, Cfg, Fn, Opt, Req, Resp, Src, Dst, Addr, Auth, Init, Exec, Cmd, Env, Pkg, Fmt, Doc, Spec, Sync, Async, Max, Min, Len, Cap, Idx, Tmp, Ref, Val, Str, Int, Bool, Impl, Repo.
  - **Banned** — half-word truncations that harm readability: `sess`, `conn`, `svc`, `mgr`, `recv`, `svr`, `tbl`, `hdlr`, `dlg`, `desc`, `proc`, `coll`.
  - When in doubt, spell out the full word.
- **Markdown**: no emojis in documentation files.
- **Git**:
  - Follow the commit message rules in [commit-message-instructions.md](commit-message-instructions.md).
  - All commit messages in English.
  - Each commit must represent exactly one logical change.
  - Before every commit, run `make check` (runs fmt → lint → test in order).
- **Tests**: co-located with source (`_test.go`). Cover happy path and at least one error path.
  - **Test-first for bug fixes**: when a bug is discovered, write a failing test that reproduces it before touching production code. The PR must include this test.
  - **Benchmarks**: changes to encoding or frame allocation must include a benchmark. Verify with `make bench`.
- **API compatibility**:
  - Exported symbols are a public contract. Changing or removing any exported identifier is a breaking change requiring a major version bump.
  - Adding a method to an exported interface breaks all external implementations — treat it as a breaking change.
  - Mark deprecated symbols with `// Deprecated: use Xxx instead.` before removal.
- **Error format**: wrap errors as `fmt.Errorf("wspulse: <context>: %w", err)`; define sentinel errors as `errors.New("wspulse: <description>")`.

## Critical Rules

1. **Read before write** — always read the target file and relevant docs fully before editing.
2. **Minimal changes** — one concern per edit; no drive-by refactors.
3. **No hardcoded secrets** — all configuration via environment variables.
4. **Zero external dependencies** — core must only depend on Go stdlib. Any change introducing an external dependency must be explicitly justified and approved.
5. **No breaking changes without version bump** — never rename, remove, or change the signature of an exported symbol without bumping the major version. When unsure, add alongside the old symbol and deprecate.
6. **Accuracy** — if you have questions or need clarification, ask the user. Do not make assumptions without confirming.
7. **Language consistency** — when the user writes in Traditional Chinese, respond in Traditional Chinese; otherwise respond in English.

## Session Protocol

> Files under `doc/local/` are git-ignored and must **never** be committed.
> This applies to all plan files and the AI learning log.

- **Plan mode**: when implementing a new feature or multi-file fix, save a plan
  to `doc/local/plan/<feature-name>.md` before starting. Keep it updated with
  completed steps and any plan changes throughout the session.
- **AI learning log**: at the end of a session where mistakes were made or
  reusable techniques were discovered, append a short entry to the session log
  under `doc/local/` (exact subfolder TBD). Entry format:
  `Date` / `Issue or Learning` / `Root Cause` / `Prevention Rule`.
  Append only — never overwrite existing entries.
