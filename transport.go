package wspulse

import "context"

// MessageType indicates the WebSocket message type used in Transport read and
// write operations. Values follow RFC 6455 §11.8 and match those used by
// github.com/coder/websocket and gorilla/websocket — numeric values are
// identical, so only a type cast is needed at module boundaries (no runtime
// calculation).
type MessageType int

const (
	// TextMessage denotes a UTF-8 encoded text frame (opcode 1).
	TextMessage MessageType = 1

	// BinaryMessage denotes a binary frame (opcode 2).
	BinaryMessage MessageType = 2
)

// StatusCode is a WebSocket close status code as defined by RFC 6455 §7.4.
// Values match those used by github.com/coder/websocket — numeric values are
// identical, so only a type cast is needed at module boundaries (no runtime
// calculation).
type StatusCode int

// Standard WebSocket close status codes from RFC 6455 §7.4.1.
// These are valid on the wire — safe to pass to Transport.Close.
const (
	// StatusNormalClosure indicates a normal, intentional close (1000).
	StatusNormalClosure StatusCode = 1000

	// StatusGoingAway indicates the endpoint is going away, e.g. server shutdown
	// or browser tab close (1001).
	StatusGoingAway StatusCode = 1001

	// StatusProtocolError indicates a protocol error was detected (1002).
	StatusProtocolError StatusCode = 1002

	// StatusUnsupportedData indicates the received data type cannot be handled,
	// e.g. a binary frame sent to an endpoint that only accepts text (1003).
	StatusUnsupportedData StatusCode = 1003

	// StatusInvalidFramePayloadData indicates the received message contains data
	// inconsistent with the message type, e.g. non-UTF-8 bytes in a text frame
	// (1007).
	StatusInvalidFramePayloadData StatusCode = 1007

	// StatusPolicyViolation indicates a message that violates an endpoint policy
	// was received (1008).
	StatusPolicyViolation StatusCode = 1008

	// StatusMessageTooBig indicates the received message is too large to process
	// (1009).
	StatusMessageTooBig StatusCode = 1009

	// StatusMandatoryExtension indicates the client expected the server to
	// negotiate a required extension but the server did not (1010).
	StatusMandatoryExtension StatusCode = 1010

	// StatusInternalError indicates the server encountered an unexpected condition
	// that prevented it from fulfilling the request (1011).
	StatusInternalError StatusCode = 1011
)

// Local-only WebSocket close status codes from RFC 6455 §7.4.1.
// These MUST NOT be sent in a close frame — do not pass them to Transport.Close.
// Use them only to classify locally observed error conditions (logging, metrics).
const (
	// StatusNoStatusReceived indicates a close frame was received but contained
	// no status code (1005).
	StatusNoStatusReceived StatusCode = 1005

	// StatusAbnormalClosure indicates the connection was closed without any close
	// frame, e.g. an abrupt TCP drop (1006).
	StatusAbnormalClosure StatusCode = 1006

	// StatusTLSHandshake indicates a TLS handshake failure (1015).
	StatusTLSHandshake StatusCode = 1015
)

// Transport abstracts the WebSocket connection for testability.
// The API is context-based: deadlines are expressed via context cancellation
// rather than explicit SetReadDeadline / SetWriteDeadline calls.
//
// *github.com/coder/websocket.Conn does not satisfy this interface directly;
// each consuming module (hub, client-go) wraps it in a thin adapter that
// converts between Transport's domain types and coder's native types.
//
// Implementations must be comparable (== / !=). The hub uses interface
// equality to detect stale transport-died notifications. Pointer receiver
// types satisfy this requirement naturally.
type Transport interface {
	// Read reads the next message from the connection.
	// Blocks until a message arrives, ctx is cancelled, or the connection closes.
	Read(ctx context.Context) (MessageType, []byte, error)

	// Write sends a message to the connection.
	// ctx may carry a deadline for the write operation.
	Write(ctx context.Context, typ MessageType, data []byte) error

	// Ping sends a ping to the peer and waits for a pong.
	// ctx may carry a deadline; if the pong does not arrive before the deadline,
	// Ping returns an error and the connection should be considered dead.
	// Must be called concurrently with Read.
	Ping(ctx context.Context) error

	// SetReadLimit sets the maximum size in bytes for a single message read
	// from the connection. Messages exceeding this limit are rejected.
	SetReadLimit(n int64)

	// Close performs the WebSocket close handshake with the given status code
	// and reason, then closes the underlying connection.
	// Implementations must enforce a bounded internal timeout — they must not
	// block indefinitely. The reference implementation (github.com/coder/websocket)
	// applies a 5 s write timeout for the close frame and waits up to 5 s for
	// the peer's response.
	// Callers must not pass local-only codes (StatusNoStatusReceived,
	// StatusAbnormalClosure, StatusTLSHandshake) — doing so is a protocol
	// violation. Use only the standard codes declared in the first const block.
	Close(code StatusCode, reason string) error

	// CloseNow closes the underlying connection immediately without
	// attempting a close handshake. Used in defer paths and error teardown.
	CloseNow() error
}
