# wspulse/core

[![CI](https://github.com/wspulse/core/actions/workflows/ci.yml/badge.svg)](https://github.com/wspulse/core/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/wspulse/core.svg)](https://pkg.go.dev/github.com/wspulse/core)
[![Go](https://img.shields.io/badge/Go-1.26-blue.svg?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Shared types for the [wspulse](https://github.com/wspulse) WebSocket ecosystem.

This module provides `Message`, `Codec`, `JSONCodec`, wire-protocol types (`MessageType`, `StatusCode`), and sentinel errors used by both [wspulse/hub](https://github.com/wspulse/hub) and [wspulse/client-go](https://github.com/wspulse/client-go). It has **zero production dependencies** (Go stdlib only).

**Status:** v0 --- API is being stabilized. Module path: `github.com/wspulse/core`.

---

## Install

```bash
go get github.com/wspulse/core
```

---

## Quick Start

```go
import wspulse "github.com/wspulse/core"

// Create a message
msg := wspulse.Message{
    Event:   "msg",
    Payload: []byte(`{"text":"hello"}`),
}

// Encode with the default JSON codec
data, err := wspulse.JSONCodec.Encode(msg)
if err != nil {
    log.Fatal(err)
}

// Decode
decoded, err := wspulse.JSONCodec.Decode(data)
if err != nil {
    log.Fatal(err)
}

fmt.Println(decoded.Event) // "msg"

// Check the codec's WebSocket wire type
wspulse.JSONCodec.WireType() // returns wspulse.TextMessage
```

### Sentinel errors

```go
if errors.Is(err, wspulse.ErrConnectionClosed) {
    // connection was already closed
}
if errors.Is(err, wspulse.ErrSendBufferFull) {
    // outbound buffer full, message was dropped
}
```

---

## Packages

### `github.com/wspulse/core` (root)

Core shared types used across the wspulse ecosystem.

### `github.com/wspulse/core/router`

Gin-style event router for dispatching incoming `wspulse.Message` values to registered handlers. Features global middleware, per-event handler chains, a configurable fallback for unmatched messages, and a built-in `Recovery()` middleware.

#### Routing key --- the `"event"` JSON field

Every message is encoded on the wire as a JSON object. The `"event"` field is what the router uses to select the handler:

```json
{
  "event": "chat.message",
  "payload": { "text": "hello" }
}
```

`msg.Event` on the Go side maps directly to `"event"` in JSON. Register handlers with `r.On("chat.message", ...)` to match that value. The first parameter to `On` is named `event` to make this correspondence explicit.

#### Usage

```go
import (
    "encoding/json"

    wspulse "github.com/wspulse/core"
    "github.com/wspulse/core/router"
)

r := router.New()

// Global middleware --- runs before every handler
r.Use(router.Recovery())
r.Use(func(c *router.Context) {
    // authenticate, rate-limit, set metadata ...
    c.Set("userID", authenticate(c.Connection))
    c.Next()
})

// Per-event handlers --- matched against msg.Event ("event" in JSON)
r.On("chat.message", func(c *router.Context) {
    userID := c.GetString("userID")
    payload, _ := json.Marshal(map[string]any{"ok": true, "from": userID})
    _ = c.Connection.Send(wspulse.Message{
        Event:   "chat.ack",
        Payload: payload,
    })
})
r.On("ping", func(c *router.Context) {
    _ = c.Connection.Send(wspulse.Message{Event: "pong"})
})

// Dispatch --- call this from WithOnMessage in wspulse/hub
r.Dispatch(connection, msg)
```

Key properties:

- Routing key is `msg.Event`, which maps to the `"event"` field in the JSON wire format
- `Context.Next()` / `Abort()` / `IsAborted()` flow control (same as Gin)
- `Context.Set` / `Get` / `MustGet` / `GetString` typed key-value metadata
- `sync.Pool`-backed Context recycling --- **0 steady-state allocations per dispatch** (metadata map allocated once per pooled Context on first `Set`; preserved across pool reuses)
- Lazy chain building: `Use` or `On` can be called in any order before the first `Dispatch`
- Panics at startup on empty event name or duplicate registration
- Max chain length: 62 handlers (middleware + route handlers combined)

---

## Development

```bash
make fmt        # auto-format source files (gofmt + goimports)
make check      # validate format, lint, test with race detector (pre-commit gate)
make test       # go test -race -count=50 ./... (override: TEST_COUNT=N)
make test-cover # go test with coverage report -> coverage.html
make bench      # run benchmarks with memory allocation stats
make tidy       # go mod tidy (GOWORK=off)
make clean      # remove build artifacts and test cache
```

---

## Related

- [wspulse/hub](https://github.com/wspulse/hub) --- WebSocket server library
- [wspulse/client-go](https://github.com/wspulse/client-go) --- Go WebSocket client with auto-reconnect
