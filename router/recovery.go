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
//	//	r := router.New()
//	//	r.Use(router.Recovery())
//	//	r.On("chat.message", handleChat)
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if v := recover(); v != nil {
				slog.Error("router: recovered from panic in handler",
					"panic", fmt.Sprintf("%v", v),
					"stack", string(debug.Stack()),
					"event", c.Frame.Event,
					"connectionID", c.Connection.ID(),
				)
			}
		}()
		c.Next()
	}
}
