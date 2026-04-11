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

// WebSocket close status codes from RFC 6455 §7.4.
//
// Codes are grouped by whether they may appear in a WebSocket close frame sent
// to the peer:
//
//   - Standard codes (1000–1003, 1007–1011): valid on the wire; pass to Close.
//   - Local-only codes (1005, 1006, 1015): RFC 6455 §7.4.1 explicitly reserves
//     these for local use. They MUST NOT be sent in a close frame. They are
//     exposed here solely for classifying locally observed error conditions
//     (e.g. logging, metrics, internal routing). Passing them to Close will
//     result in a protocol violation.
const (
	// StatusNormalClosure indicates a normal, intentional close (1000).
	StatusNormalClosure StatusCode = 1000

	// StatusGoingAway indicates the endpoint is going away, e.g. server shutdown
	// or browser tab close (1001).
	StatusGoingAway StatusCode = 1001

	// StatusProtocolError indicates a protocol error was detected (1002).
	StatusProtocolError StatusCode = 1002

	// StatusUnsupportedData indicates the received data type cannot be handled,
	// e.g. a text frame containing non-UTF-8 bytes (1003).
	StatusUnsupportedData StatusCode = 1003

	// StatusNoStatusReceived indicates a connection closed without a close frame
	// containing a status code (1005).
	//
	// LOCAL-ONLY — RFC 6455 §7.4.1: MUST NOT be sent in a close frame.
	// Use only for local error classification (logging, metrics).
	StatusNoStatusReceived StatusCode = 1005

	// StatusAbnormalClosure indicates a connection closed without a close frame,
	// e.g. an abrupt TCP drop (1006).
	//
	// LOCAL-ONLY — RFC 6455 §7.4.1: MUST NOT be sent in a close frame.
	// Use only for local error classification (logging, metrics).
	StatusAbnormalClosure StatusCode = 1006

	// StatusInvalidFramePayloadData indicates the received message contains data
	// inconsistent with the message type, e.g. non-UTF-8 bytes in a text message
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

	// StatusTLSHandshake indicates a TLS handshake failure (1015).
	//
	// LOCAL-ONLY — RFC 6455 §7.4.1: MUST NOT be sent in a close frame.
	// Use only for local error classification (logging, metrics).
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
	// Callers must not pass StatusAbnormalClosure — it is reserved for local
	// error classification and is not a valid on-wire close code.
	Close(code StatusCode, reason string) error

	// CloseNow closes the underlying connection immediately without
	// attempting a close handshake. Used in defer paths and error teardown.
	CloseNow() error
}
