// Package router provides Gin-style event routing for wspulse frames.
// Inbound frames are dispatched by [Frame.Type] through a middleware chain
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
//	// Integrate with wspulse/server:
//	srv := server.NewServer(connectFunc,
//	    server.WithOnMessage(func(connection server.Connection, frame wspulse.Frame) {
//	        rtr.Dispatch(connection, frame)
//	    }),
//	)
package router

import wspulse "github.com/wspulse/core"

// Connection represents the logical WebSocket session from the router's
// perspective. It is a consumer-defined interface: any type that provides
// these five methods satisfies it.
//
// wspulse/server's server.Connection satisfies this interface via Go
// structural subtyping — no adapter is required.
type Connection interface {
	// ID returns the unique connection identifier.
	ID() string

	// RoomID returns the room this connection belongs to.
	RoomID() string

	// Send enqueues frame for delivery to the remote peer.
	Send(frame wspulse.Frame) error

	// Close initiates a graceful shutdown of the session.
	Close() error

	// Done returns a channel that is closed when the session terminates.
	Done() <-chan struct{}
}
