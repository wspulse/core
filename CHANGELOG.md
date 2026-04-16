# Changelog

## [Unreleased]

### Removed

- **BREAKING**: `Transport` interface removed from `core`. Each consuming module (`wspulse/hub`, `wspulse/client-go`) now defines its own internal transport interface with exactly the methods it needs. `MessageType`, `StatusCode`, and their constants remain in `core` and are unaffected.

---

## [0.4.0] - 2026-04-11

### Added

- `MessageType` named type (`int`) for WebSocket frame types
- `StatusCode` named type (`int`) for WebSocket close status codes; full RFC 6455 §7.4.1 set: `StatusNormalClosure` (1000), `StatusGoingAway` (1001), `StatusProtocolError` (1002), `StatusUnsupportedData` (1003), `StatusNoStatusReceived` (1005), `StatusAbnormalClosure` (1006), `StatusInvalidFramePayloadData` (1007), `StatusPolicyViolation` (1008), `StatusMessageTooBig` (1009), `StatusMandatoryExtension` (1010), `StatusInternalError` (1011), `StatusTLSHandshake` (1015). Local-only codes (1005, 1006, 1015) are documented as MUST NOT be sent in a close frame.

### Changed

- **BREAKING**: `Transport` interface redesigned with a context-based API. Removed `SetReadDeadline`, `SetWriteDeadline`, `SetPongHandler`, `ReadMessage`, `WriteMessage`. Added `Read(ctx)`, `Write(ctx, typ, data)`, `Ping(ctx)`, `Close(code, reason)`, `CloseNow`. `SetReadLimit` retained unchanged. Consuming modules must wrap `*coder/websocket.Conn` in a thin adapter.
- **BREAKING**: `TextMessage` and `BinaryMessage` changed from untyped `int` to the new `MessageType` type. Values are unchanged (1 and 2).
- **BREAKING**: `Codec.FrameType()` now returns `MessageType` instead of `int`. Update any custom `Codec` implementations accordingly.

---

## [0.3.1] - 2026-04-04

### Changed

- `Transport` GoDoc now documents the comparability requirement — implementations must be comparable (`==` / `!=`) because the server uses interface equality to detect stale transports

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
- `TextMessage` (1) and `BinaryMessage` (2) untyped int constants for WebSocket frame types
- `ErrConnectionClosed` and `ErrSendBufferFull` sentinel errors

[Unreleased]: https://github.com/wspulse/core/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/wspulse/core/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/wspulse/core/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/wspulse/core/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/wspulse/core/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/wspulse/core/releases/tag/v0.1.0
