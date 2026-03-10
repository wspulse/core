package router_test

import (
	"strings"
	"testing"

	wspulse "github.com/wspulse/core"
	"github.com/wspulse/core/router"
)

// ---- routing ----------------------------------------------------------------

func TestRouter_On_DispatchRoutesToRegisteredHandler(t *testing.T) {
	called := false

	rtr := router.New()
	rtr.On("ping", func(_ *router.Context) { called = true })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "ping"})

	if !called {
		t.Error("registered handler was not called")
	}
}

func TestRouter_On_UnmatchedFrameCallsFallback(t *testing.T) {
	defaultFallbackCalled := false
	customFallback := func(ctx *router.Context) {
		if ctx.Frame.Type == "unknown" {
			defaultFallbackCalled = true
		}
	}

	rtr := router.New(router.WithFallback(customFallback))
	rtr.On("ping", func(_ *router.Context) {})

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "unknown"})

	if !defaultFallbackCalled {
		t.Error("fallback was not called for unmatched frame type")
	}
}

func TestRouter_On_DispatchRoutesCorrectEvent(t *testing.T) {
	var received string

	rtr := router.New()
	rtr.On("ping", func(_ *router.Context) { received = "ping" })
	rtr.On("chat", func(_ *router.Context) { received = "chat" })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "chat"})

	if received != "chat" {
		t.Errorf("expected %q, got %q", "chat", received)
	}
}

// ---- middleware --------------------------------------------------------------

func TestRouter_Use_MiddlewareRunsBeforeHandler(t *testing.T) {
	order := make([]string, 0, 2)

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { order = append(order, "middleware"); ctx.Next() })
	rtr.On("ping", func(_ *router.Context) { order = append(order, "handler") })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "ping"})

	if len(order) != 2 || order[0] != "middleware" || order[1] != "handler" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestRouter_Use_AbortInMiddlewareBlocksHandler(t *testing.T) {
	handlerCalled := false

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Abort() })
	rtr.On("ping", func(_ *router.Context) { handlerCalled = true })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "ping"})

	if handlerCalled {
		t.Error("handler should not be called after middleware Abort")
	}
}

// ---- WithFallback ------------------------------------------------------------

func TestRouter_WithFallback_CalledForUnregisteredEvent(t *testing.T) {
	var gotType string

	rtr := router.New(router.WithFallback(func(ctx *router.Context) {
		gotType = ctx.Frame.Type
	}))

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "mystery"})

	if gotType != "mystery" {
		t.Errorf("expected %q, got %q", "mystery", gotType)
	}
}

func TestRouter_WithFallback_RegisteredEventDoesNotUseFallback(t *testing.T) {
	fallbackCalled := false

	rtr := router.New(router.WithFallback(func(_ *router.Context) {
		fallbackCalled = true
	}))
	rtr.On("ping", func(_ *router.Context) {})

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "ping"})

	if fallbackCalled {
		t.Error("fallback should not be called for a registered event")
	}
}

func TestRouter_WithFallback_PanicsOnNilHandler(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic when WithFallback receives nil")
		}
	}()
	router.New(router.WithFallback(nil))
}

func TestRouter_DefaultFallback_CalledForUnmatchedFrameWithNoCustomFallback(t *testing.T) {
	// With no WithFallback option the defaultFallback (slog.Warn) is used.
	// We just verify Dispatch does not panic and the code path is exercised.
	rtr := router.New()
	rtr.On("ping", func(_ *router.Context) {})

	defer func() {
		if recover() != nil {
			t.Error("defaultFallback must not panic")
		}
	}()
	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "unknown"})
}

// ---- panic on misuse --------------------------------------------------------

func TestRouter_On_PanicsOnEmptyEventName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for empty event name")
		}
	}()
	rtr := router.New()
	rtr.On("", func(_ *router.Context) {})
}

func TestRouter_On_PanicsOnDuplicateRegistration(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for duplicate event registration")
		}
	}()
	rtr := router.New()
	rtr.On("ping", func(_ *router.Context) {})
	rtr.On("ping", func(_ *router.Context) {})
}

func TestRouter_CombineHandlers_PanicsWhenChainTooLong(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("expected panic for oversized handler chain")
			return
		}
		msg, ok := err.(string)
		if !ok {
			t.Errorf("expected string panic, got %T: %v", err, err)
			return
		}
		if !strings.Contains(msg, "exceeds maximum") {
			t.Errorf("panic message %q does not mention exceeds maximum", msg)
		}
	}()

	rtr := router.New()
	// add 63 middleware handlers so that combined with the route handler the
	// total (64) exceeds maxChainLength (63), triggering a panic in buildChains.
	for i := 0; i < 63; i++ {
		rtr.Use(func(ctx *router.Context) { ctx.Next() })
	}
	rtr.On("ping", func(_ *router.Context) {})
	// buildChains is lazy — the panic fires on the first Dispatch call.
	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "ping"})
}

// ---- benchmark ---------------------------------------------------------------

func BenchmarkDispatch(b *testing.B) {
	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Next() })
	rtr.On("ping", func(_ *router.Context) {})

	conn := newMockConnection("bench-conn", "bench-room")
	frm := wspulse.Frame{Type: "ping"}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		rtr.Dispatch(conn, frm)
	}
}
