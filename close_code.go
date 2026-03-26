package wspulse

// WebSocket close codes in the application-defined range (4000-4999).
// These codes are used by the wspulse protocol to communicate session-level
// events that have no standard WebSocket close code equivalent.
const (
	// CloseSessionExpired is sent by the server when a client requests session
	// resumption but the session no longer exists (grace window expired or server
	// restarted). This close code is used only in the rare race case where the
	// HTTP upgrade completes before the hub discovers that the session has been
	// destroyed. The common case is rejected at the HTTP layer with 410 Gone.
	//
	// On receiving this close code, the client should invoke its onSessionExpired
	// callback and stop automatic reconnection.
	//
	// The value 4100 mirrors HTTP 410 Gone for easy association.
	CloseSessionExpired = 4100
)
