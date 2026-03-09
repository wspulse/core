# Copilot Instructions — wspulse/core

## Project Overview

wspulse/core provides the **shared types** used across the wspulse WebSocket ecosystem: `Frame`, `Codec`, `JSONCodec`, and sentinel errors. Module path: `github.com/wspulse/core`. Package name: `wspulse`. This module has **zero external dependencies** (stdlib only).

## Architecture

- **`frame.go`** — `Frame` struct (the minimal transport unit), WebSocket message type constants (`TextMessage`, `BinaryMessage`), and shared sentinel errors (`ErrConnectionClosed`, `ErrSendBufferFull`).
- **`codec.go`** — `Codec` interface (`Encode`, `Decode`, `FrameType`), default `JSONCodec` implementation, and the unexported `wireFrame`/`jsonCodec` types.

## Development Workflow

```bash
# Run all tests with race detector
go test -race -count=3 ./...

# Vet
go vet ./...

# Lint (requires golangci-lint)
golangci-lint run ./...

# Format
goimports -w .
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
  - Before every commit, run in order:
    1. `goimports -w .` — fix imports and formatting
    2. `golangci-lint run ./...` — must pass with zero warnings
    3. `go test -race ./...` — must pass
- **Tests**: co-located with source (`_test.go`). Cover happy path and at least one error path.

## Critical Rules

1. **Read before write** — always read the target file and relevant docs fully before editing.
2. **Minimal changes** — one concern per edit; no drive-by refactors.
3. **No hardcoded secrets** — all configuration via environment variables.
4. **Zero external dependencies** — core must only depend on Go stdlib. Any change introducing an external dependency must be explicitly justified and approved.
5. **Accuracy** — if you have questions or need clarification, ask the user. Do not make assumptions without confirming.
6. **Language consistency** — when the user writes in Traditional Chinese, respond in Traditional Chinese; otherwise respond in English.
