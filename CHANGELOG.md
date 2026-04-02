# Changelog

## [Unreleased]

---

## [0.3.0] - 2026-04-02

### Added

- `Transport` interface — abstracts WebSocket connection for testability. `*gorilla/websocket.Conn` satisfies it via duck typing.

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
- `TextMessage` (1) and `BinaryMessage` (2) WebSocket frame type constants
- `ErrConnectionClosed` and `ErrSendBufferFull` sentinel errors

[Unreleased]: https://github.com/wspulse/core/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/wspulse/core/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/wspulse/core/releases/tag/v0.1.0
