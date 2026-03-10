package router

import (
	"math"

	wspulse "github.com/wspulse/core"
)

// abortIndex is the sentinel value assigned to Context.index by Abort.
// Any index >= abortIndex means the chain has been aborted.
// The maximum number of handlers in a single chain is abortIndex - 1 (62).
// This mirrors Gin's design: math.MaxInt8 >> 1 == 63.
const abortIndex int8 = math.MaxInt8 >> 1

// HandlerFunc is the function signature for all router handlers and middleware.
// The *Context argument carries the inbound connection, frame, flow control
// methods, and a per-dispatch key/value store.
type HandlerFunc func(*Context)

// HandlersChain is an ordered slice of HandlerFunc values representing a
// middleware + handler pipeline.
type HandlersChain []HandlerFunc

// Context carries the state for a single frame dispatch. It is obtained from
// a sync.Pool and reset between dispatches; callers must not hold a reference
// to a Context after the handler returns.
//
// All methods are safe to call only from the goroutine that invoked Dispatch.
// The wspulse readPump guarantees serial delivery per connection, so no
// synchronization is needed inside a handler chain.
type Context struct {
	// Connection is the logical WebSocket session that sent the frame.
	Connection Connection

	// Frame is the decoded inbound frame being dispatched.
	Frame wspulse.Frame

	// handlers is the merged middleware + handler chain for this dispatch.
	handlers HandlersChain

	// index tracks the current position in the handler chain.
	index int8

	// keys is the per-dispatch metadata store, lazily initialized.
	keys map[string]any
}

// Next executes the remaining handlers in the chain. It should be called
// inside middleware to pass control to the next handler.
func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		ctx.handlers[ctx.index](ctx)
		ctx.index++
	}
}

// Abort prevents any remaining handlers in the chain from being called.
// The current handler continues executing normally after Abort; only
// subsequent handlers are skipped.
func (ctx *Context) Abort() {
	ctx.index = abortIndex
}

// IsAborted reports whether Abort has been called on this context.
func (ctx *Context) IsAborted() bool {
	return ctx.index >= abortIndex
}

// Set stores a key/value pair in the per-dispatch metadata store.
// Keys must be non-empty strings. The store is lazily initialized on first use.
func (ctx *Context) Set(key string, value any) {
	if ctx.keys == nil {
		ctx.keys = make(map[string]any)
	}
	ctx.keys[key] = value
}

// Get returns the value stored under key and whether it exists.
func (ctx *Context) Get(key string) (any, bool) {
	value, exists := ctx.keys[key]
	return value, exists
}

// MustGet returns the value stored under key. It panics if the key does not exist.
func (ctx *Context) MustGet(key string) any {
	value, exists := ctx.Get(key)
	if !exists {
		panic("router: key not found in context: " + key)
	}
	return value
}

// GetString returns the string value stored under key, or "" if the key does
// not exist or its value is not a string.
func (ctx *Context) GetString(key string) string {
	value, exists := ctx.keys[key]
	if !exists {
		return ""
	}
	str, _ := value.(string)
	return str
}

// reset clears all fields so the Context can be safely returned to the pool.
func (ctx *Context) reset() {
	ctx.Connection = nil
	ctx.Frame = wspulse.Frame{}
	ctx.handlers = nil
	ctx.index = -1
	ctx.keys = nil
}
