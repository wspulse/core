package router

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

// Recovery returns a HandlerFunc that catches any panic raised by downstream
// handlers in the chain. On recovery, it logs the panic value and the full
// stack trace at ERROR level via log/slog and then returns normally, keeping
// the connection alive.
//
// Place Recovery as the first middleware so it wraps the entire chain:
//
//	r := router.New()
//	r.Use(router.Recovery())
//	r.On("chat.message", handleChat)
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			// In Go 1.21+, panic(nil) causes recover() to return *runtime.PanicNilError
			// (not nil), so v != nil correctly catches all panics including panic(nil).
			if v := recover(); v != nil {
				var connectionID string
				if c.Connection != nil {
					connectionID = c.Connection.ID()
				}
				slog.Error("router: recovered from panic in handler",
					"panic", fmt.Sprintf("%v", v),
					"stack", string(debug.Stack()),
					"event", c.Message.Event,
					"connectionID", connectionID,
				)
			}
		}()
		c.Next()
	}
}
