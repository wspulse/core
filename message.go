package wspulse

import "errors"

// Message is the minimal transport unit for all WebSocket communication.
// Payload bytes are opaque to the wspulse layer; their format depends on the Codec in use.
// When using JSONCodec (the default), Payload must be valid JSON bytes (e.g. output of
// json.Marshal). When using a binary codec, Payload is the codec-encoded bytes.
type Message struct {
	// Event identifies the message purpose. wspulse does not interpret this value.
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
	// The message is dropped; handle backpressure at the application layer.
	ErrSendBufferFull = errors.New("wspulse: send buffer full, message dropped")
)
