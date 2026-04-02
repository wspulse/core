package wspulse

import "time"

// Transport abstracts the WebSocket connection for testability.
// *gorilla/websocket.Conn satisfies this interface via duck typing —
// no wrapper or adapter is needed in production code.
type Transport interface {
	// ReadMessage reads a complete message from the connection.
	// The returned messageType is TextMessage (1) or BinaryMessage (2).
	ReadMessage() (messageType int, p []byte, err error)

	// WriteMessage writes a complete message to the connection.
	// messageType should be TextMessage (1) or BinaryMessage (2),
	// consistent with Codec.FrameType.
	WriteMessage(messageType int, data []byte) error

	// SetReadLimit sets the maximum size in bytes for a message read
	// from the connection.
	SetReadLimit(limit int64)

	// SetReadDeadline sets the read deadline on the underlying
	// network connection.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the write deadline on the underlying
	// network connection.
	SetWriteDeadline(t time.Time) error

	// SetPongHandler sets the handler for pong messages received
	// from the peer.
	SetPongHandler(h func(appData string) error)

	// Close closes the underlying network connection.
	Close() error
}
