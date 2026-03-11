package router

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	wspulse "github.com/wspulse/core"
)

// maxChainLength is the maximum number of handlers that may appear in a single
// merged chain (global middleware + route handlers). Reaching or exceeding this
// limit causes a panic at setup time (in Use or On). Value mirrors abortIndex (63).
const maxChainLength = int(abortIndex)

// Option is a functional option for configuring a Router.
type Option func(*Router)

// WithFallback sets a custom handler invoked when no route matches the
// incoming frame's Event. The fallback participates in the normal handler
// chain, so global middleware registered via Use runs before it.
//
// The default fallback logs the unmatched frame event at WARN level using
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
	merged atomic.Bool

	// buildMu guards the lazy buildChains call in Dispatch. The outer atomic
	// read provides a fast path; the mutex ensures only one goroutine runs
	// buildChains when merged is false.
	buildMu sync.Mutex
}

// New returns a new Router with the provided options applied.
// The default fallback logs unmatched frame events at WARN level.
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
	for i, h := range handlers {
		if h == nil {
			panic(fmt.Sprintf("router: Use: handler at index %d must not be nil", i))
		}
	}
	// Validate lengths before mutating state so a panic leaves the router unchanged.
	newTotal := len(r.handlers) + len(handlers)
	if newTotal+1 >= maxChainLength {
		panic(fmt.Sprintf(
			"router: Use: handler chain length %d exceeds maximum %d (too many global middleware)",
			newTotal+1, maxChainLength-1,
		))
	}
	for event, routeHandlers := range r.rawRoutes {
		if newTotal+len(routeHandlers) >= maxChainLength {
			panic(fmt.Sprintf(
				"router: Use: handler chain length %d for event %q would exceed maximum %d",
				newTotal+len(routeHandlers), event, maxChainLength-1,
			))
		}
	}
	r.handlers = append(r.handlers, handlers...)
	r.merged.Store(false)
}

// On registers one or more handlers for the given Frame.Event value ("event" in
// JSON). Panics if event is empty, if no handlers are provided, if any handler
// is nil, or if event is already registered.
// On must not be called concurrently with Dispatch.
func (r *Router) On(event string, handlers ...HandlerFunc) {
	if event == "" {
		panic("router: On: event must not be empty")
	}
	if len(handlers) == 0 {
		panic(fmt.Sprintf("router: On: no handlers provided for event %q", event))
	}
	for i, h := range handlers {
		if h == nil {
			panic(fmt.Sprintf("router: On: handler at index %d for event %q must not be nil", i, event))
		}
	}
	if _, exists := r.rawRoutes[event]; exists {
		panic(fmt.Sprintf("router: On: duplicate registration for event %q", event))
	}
	if len(r.handlers)+len(handlers) >= maxChainLength {
		panic(fmt.Sprintf(
			"router: On: handler chain length %d for event %q exceeds maximum %d",
			len(r.handlers)+len(handlers), event, maxChainLength-1,
		))
	}
	r.rawRoutes[event] = handlers
	r.merged.Store(false)
}

// Dispatch looks up the handler chain for frame.Event and executes it.
// If no handler is registered for frame.Event, the fallback is called instead.
// Global middleware runs before the matched handler or fallback in all cases.
//
// Dispatch is safe to call concurrently from multiple goroutines after all
// routes have been registered. However, calling Use or On while Dispatch is
// running is not safe.
func (r *Router) Dispatch(conn Connection, frame wspulse.Frame) {
	if !r.merged.Load() {
		r.buildMu.Lock()
		if !r.merged.Load() {
			r.buildChains()
		}
		r.buildMu.Unlock()
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

	r.merged.Store(true)
}

// combineHandlers merges the global middleware with the provided handlers into
// a single chain. Chain length is validated eagerly in Use and On;
// by the time buildChains calls this function the combined length is guaranteed
// to be within bounds.
func (r *Router) combineHandlers(handlers HandlersChain) HandlersChain {
	total := len(r.handlers) + len(handlers)
	merged := make(HandlersChain, total)
	copy(merged, r.handlers)
	copy(merged[len(r.handlers):], handlers)
	return merged
}

// defaultFallback is the built-in fallback handler. It logs the unmatched
// event at WARN level using log/slog (Go 1.21+ stdlib).
func defaultFallback(c *Context) {
	var connectionID string
	if c.Connection != nil {
		connectionID = c.Connection.ID()
	}
	slog.Warn("router: unmatched event",
		"event", c.Frame.Event,
		"connectionID", connectionID,
	)
}
