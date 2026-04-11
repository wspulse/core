package wspulse

import "context"

// MessageType indicates the WebSocket message type used in Transport read and
// write operations. Values follow RFC 6455 §11.8 and match those used by
// github.com/coder/websocket and gorilla/websocket — numeric values are
// identical, so only a type cast is needed at module boundaries (no runtime
// calculation).
type MessageType int

const (
	// MessageText denotes a UTF-8 encoded text frame (opcode 1).
	MessageText MessageType = 1

	// MessageBinary denotes a binary frame (opcode 2).
	MessageBinary MessageType = 2
)

// StatusCode is a WebSocket close status code as defined by RFC 6455 §7.4.
// Values match those used by github.com/coder/websocket — numeric values are
// identical, so only a type cast is needed at module boundaries (no runtime
// calculation).
type StatusCode int

// WebSocket close status codes from RFC 6455 §7.4.
const (
	// StatusNormalClosure indicates a normal, intentional close (1000).
	StatusNormalClosure StatusCode = 1000

	// StatusGoingAway indicates the endpoint is going away, e.g. server shutdown
	// or browser tab close (1001).
	StatusGoingAway StatusCode = 1001

	// StatusAbnormalClosure indicates a connection closed without a close frame,
	// e.g. an abrupt TCP drop (1006). RFC 6455 reserves this code for local error
	// classification only — implementations must not include it in close frames
	// sent to the peer.
	StatusAbnormalClosure StatusCode = 1006
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
	Close(code StatusCode, reason string) error

	// CloseNow closes the underlying connection immediately without
	// attempting a close handshake. Used in defer paths and error teardown.
	CloseNow() error
}
