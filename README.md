# wspulse/core

[![CI](https://github.com/wspulse/core/actions/workflows/ci.yml/badge.svg)](https://github.com/wspulse/core/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/wspulse/core.svg)](https://pkg.go.dev/github.com/wspulse/core)
[![Go](https://img.shields.io/badge/Go-1.26-blue.svg?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

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
    Event:    "msg",
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

fmt.Println(decoded.Event) // "msg"

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
| `Frame`               | Transport unit: `ID`, `Event`, `Payload []byte`                 |
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

#### Routing key — the `"event"` JSON field

Every frame is encoded on the wire as a JSON object. The `"event"` field is what the router uses to select the handler:

```json
{
  "id": "msg-001",
  "event": "chat.message",
  "payload": { "text": "hello" }
}
```

`frame.Event` on the Go side maps directly to `"event"` in JSON. Register handlers with `r.On("chat.message", ...)` to match that value. The first parameter to `On` is named `event` to make this correspondence explicit.

#### Usage

```go
import (
    "encoding/json"

    wspulse "github.com/wspulse/core"
    "github.com/wspulse/core/router"
)

r := router.New()

// Global middleware — runs before every handler
r.Use(router.Recovery())
r.Use(func(c *router.Context) {
    // authenticate, rate-limit, set metadata …
    c.Set("userID", authenticate(c.Connection))
    c.Next()
})

// Per-event handlers — matched against frame.Event ("event" in JSON)
r.On("chat.message", func(c *router.Context) {
    userID := c.GetString("userID")
    payload, _ := json.Marshal(map[string]any{"ok": true, "from": userID})
    _ = c.Connection.Send(wspulse.Frame{
        Event:   "chat.ack",
        Payload: payload,
    })
})
r.On("ping", func(c *router.Context) {
    _ = c.Connection.Send(wspulse.Frame{Event: "pong"})
})

// Dispatch — call this from WithOnMessage in wspulse/server
r.Dispatch(connection, frame)
```

Key properties:

- Routing key is `frame.Event`, which maps to the `"event"` field in the JSON wire format
- `Context.Next()` / `Abort()` / `IsAborted()` flow control (same as Gin)
- `Context.Set` / `Get` / `MustGet` / `GetString` typed key-value metadata
- `sync.Pool`-backed Context recycling — **0 steady-state allocations per dispatch** (metadata map allocated once per pooled Context on first `Set`; preserved across pool reuses)
- Lazy chain building: `Use` or `On` can be called in any order before the first `Dispatch`
- Panics at startup on empty event name or duplicate registration
- Max chain length: 62 handlers (middleware + route handlers combined)

---

## Development

```bash
make fmt        # auto-format source files (gofmt + goimports)
make check      # validate format, lint, test with race detector (pre-commit gate)
make test       # go test -race -count=3 ./...
make test-cover # go test with coverage report → coverage.html
make bench      # run benchmarks with memory allocation stats
make tidy       # go mod tidy (GOWORK=off)
make clean      # remove build artifacts and test cache
```

---

## Related

- [wspulse/server](https://github.com/wspulse/server) — WebSocket server library
- [wspulse/client-go](https://github.com/wspulse/client-go) — Go WebSocket client with auto-reconnect
