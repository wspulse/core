package wspulse

import "errors"

// WebSocket message type constants.
// These mirror the standard WebSocket frame types without importing gorilla/websocket,
// keeping this module free of external dependencies.
const (
	// TextMessage denotes a UTF-8 encoded text WebSocket frame (opcode 1).
	TextMessage = 1

	// BinaryMessage denotes a binary WebSocket frame (opcode 2).
	BinaryMessage = 2
)

// Frame is the minimal transport unit for all WebSocket communication.
// Payload bytes are opaque to the wspulse layer; their format depends on the Codec in use.
// When using JSONCodec (the default), Payload must be valid JSON bytes (e.g. output of
// json.Marshal). When using a binary codec, Payload is the codec-encoded bytes.
type Frame struct {
	// Event identifies the frame purpose. wspulse does not interpret this value.
	// Conventional values: "msg" (user data), "sys" (system event), "ack" (acknowledgement).
	Event string

	// Payload is the encoded message body. Its format is determined by the Codec.
	Payload []byte
}

// Sentinel errors shared across the wspulse ecosystem.
var (
	// ErrConnectionClosed is returned when sending to a connection that has already been closed.
	ErrConnectionClosed = errors.New("wspulse: connection is closed")

	// ErrSendBufferFull is returned when the outbound buffer is full.
	// The frame is dropped; handle backpressure at the application layer.
	ErrSendBufferFull = errors.New("wspulse: send buffer full, frame dropped")
)
