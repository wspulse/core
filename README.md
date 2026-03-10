# wspulse/core

Shared types for the [wspulse](https://github.com/wspulse) WebSocket ecosystem.

This module provides `Frame`, `Codec`, `JSONCodec`, and sentinel errors used by both [wspulse/server](https://github.com/wspulse/server) and [wspulse/client-go](https://github.com/wspulse/client-go). It has **zero external dependencies** (Go stdlib only).

**Status:** v0 — API is being stabilized. Module path: `github.com/wspulse/core`.

---

## Install

```bash
go get github.com/wspulse/core
```

---

## Quick Start

```go
import wspulse "github.com/wspulse/core"

// Create a frame
frame := wspulse.Frame{
    ID:      "msg-001",
    Type:    "msg",
    Payload: []byte(`{"text":"hello"}`),
}

// Encode with the default JSON codec
data, err := wspulse.JSONCodec.Encode(frame)
if err != nil {
    log.Fatal(err)
}

// Decode
decoded, err := wspulse.JSONCodec.Decode(data)
if err != nil {
    log.Fatal(err)
}

fmt.Println(decoded.Type) // "msg"

// Check the codec's WebSocket frame type
wspulse.JSONCodec.FrameType() // wspulse.TextMessage (1)
```

### Sentinel errors

```go
if errors.Is(err, wspulse.ErrConnectionClosed) {
    // connection was already closed
}
if errors.Is(err, wspulse.ErrSendBufferFull) {
    // outbound buffer full, frame was dropped
}
```

---

## Public API

| Symbol                | Description                                                     |
| --------------------- | --------------------------------------------------------------- |
| `Frame`               | Transport unit: `ID`, `Type`, `Payload []byte`                  |
| `Codec`               | Interface: `Encode(Frame)`, `Decode([]byte)`, `FrameType() int` |
| `JSONCodec`           | Default codec — JSON text frames                                |
| `TextMessage`         | WebSocket text frame type constant (`1`)                        |
| `BinaryMessage`       | WebSocket binary frame type constant (`2`)                      |
| `ErrConnectionClosed` | Sentinel: connection is closed                                  |
| `ErrSendBufferFull`   | Sentinel: send buffer full, frame dropped                       |

---

## Packages

### `github.com/wspulse/core` (root)

Core shared types used across the wspulse ecosystem.

### `github.com/wspulse/core/router`

Gin-style event router for dispatching incoming `wspulse.Frame` values to registered handlers. Features global middleware, per-event handler chains, a configurable fallback for unmatched frames, and a built-in `Recovery()` middleware.

```go
import (
    wspulse "github.com/wspulse/core"
    "github.com/wspulse/core/router"
)

rtr := router.New()

// Global middleware (runs before every handler)
rtr.Use(router.Recovery())
rtr.Use(func(ctx *router.Context) {
    // authenticate / rate-limit / etc.
    ctx.Set("userID", authenticate(ctx.Connection))
    ctx.Next()
})

// Per-event handlers
rtr.On("chat", func(ctx *router.Context) {
    userID := ctx.GetString("userID")
    _ = ctx.Connection.Send(wspulse.Frame{Type: "ack", Payload: []byte(userID)})
})

// Dispatch (typically called from your readPump goroutine)
rtr.Dispatch(connection, frame)
```

Key properties:

- `Context.Next()` / `Abort()` / `IsAborted()` flow control (same as Gin)
- `Context.Set` / `Get` / `MustGet` / `GetString` typed key–value metadata
- `sync.Pool`-backed Context recycling — **0 allocations per dispatch**
- Lazy chain building: calling `Use` or `On` after the first `Dispatch` works correctly
- Panics at startup on empty event name or duplicate registration
- Max chain length: 62 handlers (middleware + route handlers combined)

---

## Related

- [wspulse/server](https://github.com/wspulse/server) — WebSocket server library
- [wspulse/client-go](https://github.com/wspulse/client-go) — Go WebSocket client with auto-reconnect

---

## License

[MIT](LICENSE)
