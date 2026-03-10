package router

import (
	"fmt"
	"log/slog"
	"sync"

	wspulse "github.com/wspulse/core"
)

// maxChainLength is the maximum number of handlers that may appear in a single
// merged chain (global middleware + route handlers). Exceeding this limit
// causes combineHandlers to panic. Value mirrors abortIndex (63).
const maxChainLength = int(abortIndex)

// Option is a functional option for configuring a Router.
type Option func(*Router)

// WithFallback sets a custom handler invoked when no route matches the
// incoming frame's Event. The fallback participates in the normal handler
// chain, so global middleware registered via Use runs before it.
//
// The default fallback logs the unmatched frame type at WARN level using
// the standard library's log/slog package.
func WithFallback(fn HandlerFunc) Option {
	if fn == nil {
		panic("router: WithFallback: handler must not be nil")
	}
	return func(r *Router) {
		r.fallback = fn
	}
}

// Router dispatches inbound wspulse frames to registered handlers based on
// Frame.Event. It supports global middleware (Use), per-event handlers (On),
// and a configurable fallback for unmatched frames.
//
// Router is safe for concurrent reads after all routes have been registered.
// Do not call On or Use concurrently with Dispatch.
type Router struct {
	// handlers holds global middleware registered via Use.
	handlers HandlersChain

	// routes maps Frame.Event to the merged (middleware + handler) chain.
	// Chains are merged lazily at the first Dispatch call.
	routes map[string]HandlersChain

	// rawRoutes holds the per-event handlers as registered, before merging
	// with global middleware. Merging is deferred so that Use calls after On
	// calls are respected.
	rawRoutes map[string]HandlersChain

	// fallback is called when no route matches Frame.Event.
	fallback HandlerFunc

	// pool recycles Context objects to eliminate per-dispatch allocations.
	pool sync.Pool

	// merged signals that route chains have been built and are ready for dispatch.
	// Reset to false whenever Use or On is called.
	merged bool
}

// New returns a new Router with the provided options applied.
// The default fallback logs unmatched frame types at WARN level.
func New(opts ...Option) *Router {
	r := &Router{
		routes:    make(map[string]HandlersChain),
		rawRoutes: make(map[string]HandlersChain),
		fallback:  defaultFallback,
	}
	r.pool.New = func() any {
		return &Context{}
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Use appends one or more middleware handlers to the global middleware chain.
// Global middleware runs before every handler, including the fallback.
// Use must not be called concurrently with Dispatch.
func (r *Router) Use(handlers ...HandlerFunc) {
	r.handlers = append(r.handlers, handlers...)
	r.merged = false
}

// On registers one or more handlers for the given Frame.Event value ("event" in
// JSON). Panics if event is empty. Panics if event is already registered.
// On must not be called concurrently with Dispatch.
func (r *Router) On(event string, handlers ...HandlerFunc) {
	if event == "" {
		panic("router: On: event must not be empty")
	}
	if _, exists := r.rawRoutes[event]; exists {
		panic(fmt.Sprintf("router: On: duplicate registration for event %q", event))
	}
	r.rawRoutes[event] = handlers
	r.merged = false
}

// Dispatch looks up the handler chain for frame.Event and executes it.
// If no handler is registered for frame.Event, the fallback is called instead.
// Global middleware runs before the matched handler or fallback in all cases.
//
// Dispatch is safe to call concurrently from multiple goroutines after all
// routes have been registered. However, calling Use or On while Dispatch is
// running is not safe.
func (r *Router) Dispatch(conn Connection, frame wspulse.Frame) {
	if !r.merged {
		r.buildChains()
	}

	c := r.pool.Get().(*Context)
	c.reset()
	c.Connection = conn
	c.Frame = frame
	c.index = -1

	chain, exists := r.routes[frame.Event]
	if !exists {
		chain = r.routes[""]
	}
	c.handlers = chain
	c.Next()

	r.pool.Put(c)
}

// buildChains merges global middleware with each route's handlers and the
// fallback, caching the result in routes.
func (r *Router) buildChains() {
	r.routes = make(map[string]HandlersChain, len(r.rawRoutes)+1)

	for event, handlers := range r.rawRoutes {
		r.routes[event] = r.combineHandlers(handlers)
	}
	// The empty string key is the fallback chain.
	r.routes[""] = r.combineHandlers(HandlersChain{r.fallback})

	r.merged = true
}

// combineHandlers merges the global middleware with the provided handlers into
// a single chain. Panics if the resulting chain would exceed maxChainLength (62).
func (r *Router) combineHandlers(handlers HandlersChain) HandlersChain {
	total := len(r.handlers) + len(handlers)
	if total >= maxChainLength {
		panic(fmt.Sprintf(
			"router: handler chain length %d exceeds maximum %d",
			total, maxChainLength-1,
		))
	}
	merged := make(HandlersChain, total)
	copy(merged, r.handlers)
	copy(merged[len(r.handlers):], handlers)
	return merged
}

// defaultFallback is the built-in fallback handler. It logs the unmatched
// event at WARN level using log/slog (Go 1.21+ stdlib).
func defaultFallback(c *Context) {
	slog.Warn("router: unmatched event",
		"event", c.Frame.Event,
		"connectionID", c.Connection.ID(),
	)
}
