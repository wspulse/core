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

## Related

- [wspulse/server](https://github.com/wspulse/server) — WebSocket server library
- [wspulse/client-go](https://github.com/wspulse/client-go) — Go WebSocket client with auto-reconnect

---

## License

[MIT](LICENSE)
