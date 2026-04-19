// Package router provides Gin-style event routing for wspulse messages.
// Inbound messages are dispatched by [wspulse.Message.Event] through a middleware chain
// with flow control ([Context.Next]/[Context.Abort]) and metadata passing
// ([Context.Set]/[Context.Get]).
//
// Typical usage:
//
//	rtr := router.New()
//	rtr.Use(router.Recovery())
//	rtr.On("chat.message", handleChat)
//	rtr.On("chat.join", handleJoin)
//
//	// Dispatch inbound messages into the router from your
//	// connection manager (e.g. wspulse/hub's OnMessage callback):
//	onMessage := func(conn Connection, msg wspulse.Message) {
//	    rtr.Dispatch(conn, msg)
//	}
package router

import wspulse "github.com/wspulse/core"

// Connection represents the logical WebSocket session from the router's
// perspective. It is a consumer-defined interface: any type that provides
// these five methods satisfies it.
//
// wspulse/hub's Connection satisfies this interface via Go
// structural subtyping — no adapter is required.
type Connection interface {
	// ID returns the unique connection identifier.
	ID() string

	// RoomID returns the room this connection belongs to.
	RoomID() string

	// Send enqueues msg for delivery to the remote peer.
	Send(msg wspulse.Message) error

	// Close initiates a graceful shutdown of the session.
	Close() error

	// Done returns a channel that is closed when the session terminates.
	Done() <-chan struct{}
}
