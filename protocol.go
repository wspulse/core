package wspulse

// MessageType indicates the WebSocket message type used in read and write
// operations. Values follow RFC 6455 §11.8 and match those used by
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
// These are valid on the wire — safe to send in a WebSocket close frame.
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

	// StatusMandatoryExtension indicates the client required a WebSocket
	// extension that the server did not negotiate (1010). Valid on the wire,
	// but not used by the wspulse ecosystem — included for completeness when
	// classifying close frames received from peers.
	StatusMandatoryExtension StatusCode = 1010

	// StatusInternalError indicates the server encountered an unexpected condition
	// that prevented it from fulfilling the request (1011).
	StatusInternalError StatusCode = 1011
)

// Local-only WebSocket close status codes from RFC 6455 §7.4.1.
// These MUST NOT be sent in a WebSocket close frame.
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
