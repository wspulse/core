# Changelog

## [Unreleased]

### Added

- `Transport` interface — abstracts WebSocket connection for testability with a context-based API
- `MessageType` named type (`int`) for WebSocket frame types; constants `MessageText` (1) and `MessageBinary` (2)
- `StatusCode` named type (`int`) for WebSocket close status codes; constants `StatusNormalClosure` (1000), `StatusGoingAway` (1001), `StatusAbnormalClosure` (1006)

### Removed

- **BREAKING**: Untyped int constants `TextMessage` and `BinaryMessage` replaced by `MessageText` and `MessageBinary` on the new `MessageType` type

---

## [0.3.1] - 2026-04-04

### Changed

- `Transport` GoDoc now documents the comparability requirement — implementations must be comparable (`==` / `!=`) because the server uses interface equality to detect stale transports

---

## [0.3.0] - 2026-04-02

### Removed

- **BREAKING**: `Frame.ID` field removed — transport layer does not use it. Applications needing message IDs should use Payload.

---

## [0.2.0] - 2026-03-11

### Added

- `router` subpackage: Gin-style event routing for wspulse frames
- `router.New(opts ...Option) *Router` with `Use`, `On(event, ...handlers)`, `Dispatch`
- `router.WithFallback(fn HandlerFunc)` — custom handler for unmatched events
- `router.Context` with `Next`/`Abort`/`IsAborted` flow control and `Set`/`Get`/`MustGet`/`GetString` metadata
- `router.Connection` interface — consumer-defined; `server.Connection` satisfies it via structural subtyping
- `router.Recovery()` built-in panic-recovery middleware; logs via `slog.Error`, keeps connection alive
- `sync.Pool`-backed Context recycling — 0 allocations per dispatch

### Changed

- `Frame.Type` renamed to `Frame.Event` (**breaking**)
- Wire format JSON key `"type"` changed to `"event"` (**breaking**)

---

## [0.1.0] - 2026-03-10

### Added

- `Frame` struct with `ID string`, `Type string`, and `Payload []byte` fields
- `Codec` interface: `Encode(Frame) ([]byte, error)`, `Decode([]byte) (Frame, error)`, `FrameType() int`
- `JSONCodec` default implementation — JSON text frames, zero external dependencies
- `TextMessage` (1) and `BinaryMessage` (2) untyped int constants for WebSocket frame types
- `ErrConnectionClosed` and `ErrSendBufferFull` sentinel errors

[Unreleased]: https://github.com/wspulse/core/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/wspulse/core/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/wspulse/core/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/wspulse/core/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/wspulse/core/releases/tag/v0.1.0
