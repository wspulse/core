package router_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wspulse "github.com/wspulse/core"
	"github.com/wspulse/core/router"
)

// ---- routing ----------------------------------------------------------------

func TestRouter_On_DispatchRoutesToRegisteredHandler(t *testing.T) {
	called := false

	rtr := router.New()
	rtr.On("ping", func(_ *router.Context) { called = true })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "ping"})

	assert.True(t, called, "registered handler was not called")
}

func TestRouter_On_UnmatchedFrameCallsFallback(t *testing.T) {
	defaultFallbackCalled := false
	customFallback := func(ctx *router.Context) {
		if ctx.Frame.Event == "unknown" {
			defaultFallbackCalled = true
		}
	}

	rtr := router.New(router.WithFallback(customFallback))
	rtr.On("ping", func(_ *router.Context) {})

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "unknown"})

	assert.True(t, defaultFallbackCalled, "fallback was not called for unmatched frame type")
}

func TestRouter_On_DispatchRoutesCorrectEvent(t *testing.T) {
	var received string

	rtr := router.New()
	rtr.On("ping", func(_ *router.Context) { received = "ping" })
	rtr.On("chat", func(_ *router.Context) { received = "chat" })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "chat"})

	assert.Equal(t, "chat", received)
}

// ---- middleware --------------------------------------------------------------

func TestRouter_Use_MiddlewareRunsBeforeHandler(t *testing.T) {
	order := make([]string, 0, 2)

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { order = append(order, "middleware"); ctx.Next() })
	rtr.On("ping", func(_ *router.Context) { order = append(order, "handler") })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "ping"})

	assert.Equal(t, []string{"middleware", "handler"}, order)
}

func TestRouter_Use_AbortInMiddlewareBlocksHandler(t *testing.T) {
	handlerCalled := false

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Abort() })
	rtr.On("ping", func(_ *router.Context) { handlerCalled = true })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "ping"})

	assert.False(t, handlerCalled, "handler should not be called after middleware Abort")
}

// ---- WithFallback ------------------------------------------------------------

func TestRouter_WithFallback_CalledForUnregisteredEvent(t *testing.T) {
	var gotEvent string

	rtr := router.New(router.WithFallback(func(ctx *router.Context) {
		gotEvent = ctx.Frame.Event
	}))

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "mystery"})

	assert.Equal(t, "mystery", gotEvent)
}

func TestRouter_WithFallback_RegisteredEventDoesNotUseFallback(t *testing.T) {
	fallbackCalled := false

	rtr := router.New(router.WithFallback(func(_ *router.Context) {
		fallbackCalled = true
	}))
	rtr.On("ping", func(_ *router.Context) {})

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "ping"})

	assert.False(t, fallbackCalled, "fallback should not be called for a registered event")
}

func TestRouter_WithFallback_PanicsOnNilHandler(t *testing.T) {
	require.Panics(t, func() {
		router.New(router.WithFallback(nil))
	})
}

func TestRouter_DefaultFallback_CalledForUnmatchedFrameWithNoCustomFallback(t *testing.T) {
	// With no WithFallback option the defaultFallback (slog.Warn) is used.
	// We just verify Dispatch does not panic and the code path is exercised.
	rtr := router.New()
	rtr.On("ping", func(_ *router.Context) {})

	require.NotPanics(t, func() {
		rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "unknown"})
	})
}

// ---- panic on misuse --------------------------------------------------------

func TestRouter_On_PanicsOnEmptyEventName(t *testing.T) {
	require.Panics(t, func() {
		rtr := router.New()
		rtr.On("", func(_ *router.Context) {})
	})
}

func TestRouter_On_PanicsOnDuplicateRegistration(t *testing.T) {
	require.Panics(t, func() {
		rtr := router.New()
		rtr.On("ping", func(_ *router.Context) {})
		rtr.On("ping", func(_ *router.Context) {})
	})
}

func TestRouter_On_PanicsOnNoHandlers(t *testing.T) {
	require.Panics(t, func() {
		rtr := router.New()
		rtr.On("ping")
	})
}

func TestRouter_On_PanicsOnNilHandler(t *testing.T) {
	require.Panics(t, func() {
		rtr := router.New()
		rtr.On("ping", nil)
	})
}

func TestRouter_Use_PanicsOnNilHandler(t *testing.T) {
	require.Panics(t, func() {
		rtr := router.New()
		rtr.Use(nil)
	})
}

func TestRouter_CombineHandlers_PanicsWhenChainTooLong(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for oversized handler chain")
		msg, ok := r.(string)
		require.True(t, ok, "expected string panic, got %T: %v", r, r)
		assert.Contains(t, msg, "exceeds maximum")
	}()

	rtr := router.New()
	// Add enough middleware so that when combined with any single handler the
	// chain length reaches maxChainLength (63), triggering a panic in Use.
	for i := 0; i < 63; i++ {
		rtr.Use(func(ctx *router.Context) { ctx.Next() })
	}
	rtr.On("ping", func(_ *router.Context) {})
	// Panic fires in Use before On or Dispatch are reached.
	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "ping"})
}

// ---- benchmark ---------------------------------------------------------------

func BenchmarkDispatch(b *testing.B) {
	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Next() })
	rtr.On("ping", func(_ *router.Context) {})

	conn := newMockConnection("bench-conn", "bench-room")
	frm := wspulse.Frame{Event: "ping"}

	// Warm up: trigger lazy chain build and initial sync.Pool population so
	// the benchmark measures steady-state dispatch cost.
	rtr.Dispatch(conn, frm)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		rtr.Dispatch(conn, frm)
	}
}
