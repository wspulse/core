# Contributing to wspulse/core

Thank you for your interest in contributing. This document describes the process and conventions expected for all contributions.

## Before You Start

- Open an issue to discuss significant changes before starting work.
- For bug fixes, write a failing test that reproduces the issue before modifying production code. The PR must include this test.
- For new features, confirm scope and API design in an issue first.

## Development Setup

```bash
git clone https://github.com/wspulse/core
cd core
go mod tidy
```

Requires: Go 1.26+, [golangci-lint](https://golangci-lint.run/), [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports).

## Pre-Commit Checklist

Run `make check` before every commit. It runs in order:

1. `make fmt` — formats all source files
2. `make lint` — runs `go vet` and `golangci-lint`; must pass with zero warnings
3. `make test` — runs tests with `-race`; must pass

If any step fails, do not commit.

## Commit Messages

Follow the format in [`.github/instructions/commit-message-instructions.md`](.github/instructions/commit-message-instructions.md):

```
<type>: <subject>

1.<reason> → <change>
```

All commit messages must be in English.

## Naming Conventions

- Use full words for all identifiers. Cryptic abbreviations are not acceptable.
- Banned: `sess`, `conn`, `svc`, `mgr`, `recv`, `svr`, `tbl`, `hdlr`, `dlg`, `desc`, `proc`, `coll`.
- Allowed short forms: ID, URL, HTTP, API, JSON, Msg, Err, Ctx, Buf, Cfg, and others listed in the copilot instructions.

## API Compatibility

wspulse/core follows semantic versioning. Any change that removes, renames, or alters the signature of an exported symbol is a **breaking change** and requires a major version bump.

- Before removing a symbol, mark it `// Deprecated: use Xxx instead.` in a minor release.
- Adding a method to an exported interface is also a breaking change.
- When in doubt, add a new symbol alongside the old one.

## Zero Dependencies

core must have **zero external dependencies**. It may only use the Go standard library. Any PR that introduces an external dependency will be rejected.

## Pull Request Guidelines

- One PR per logical change.
- Do not reformat code unrelated to your change — it creates noise in the diff.
- All CI checks must pass before review.
- Describe what changed and why, not just what the diff shows.
