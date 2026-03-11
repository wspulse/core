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
// Context and its handlers are not concurrency-safe. All methods must be
// called from the goroutine that performs dispatch, and callers are expected
// to enforce serial handler execution per logical connection.
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
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort prevents any remaining handlers in the chain from being called.
// The current handler continues executing normally after Abort; only
// subsequent handlers are skipped.
func (c *Context) Abort() {
	c.index = abortIndex
}

// IsAborted reports whether Abort has been called on this context.
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// Set stores a key/value pair in the per-dispatch metadata store.
// Keys are arbitrary strings. The store is lazily initialized on first use.
func (c *Context) Set(key string, value any) {
	if c.keys == nil {
		c.keys = make(map[string]any)
	}
	c.keys[key] = value
}

// Get returns the value stored under key and whether it exists.
func (c *Context) Get(key string) (any, bool) {
	value, exists := c.keys[key]
	return value, exists
}

// MustGet returns the value stored under key. It panics if the key does not exist.
func (c *Context) MustGet(key string) any {
	value, exists := c.Get(key)
	if !exists {
		panic("router: key not found in context: " + key)
	}
	return value
}

// GetString returns the string value stored under key, or "" if the key does
// not exist or its value is not a string.
func (c *Context) GetString(key string) string {
	value, exists := c.keys[key]
	if !exists {
		return ""
	}
	str, _ := value.(string)
	return str
}

// reset clears all fields so the Context can be safely returned to the pool.
func (c *Context) reset() {
	c.Connection = nil
	c.Frame = wspulse.Frame{}
	c.handlers = nil
	c.index = -1
	clear(c.keys) // preserve map allocation across pool reuses; no-op when nil
}
