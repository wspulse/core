package router_test

import (
	"testing"

	wspulse "github.com/wspulse/core"
	"github.com/wspulse/core/router"
)

func TestRecovery_CatchesPanicAndContinues(t *testing.T) {
	afterRecoveryCalled := false

	rtr := router.New()
	rtr.Use(router.Recovery())
	rtr.On("ping", func(_ *router.Context) { panic("boom") })

	// Dispatch must not panic to the caller; it must be swallowed by Recovery.
	defer func() {
		if recover() != nil {
			t.Error("panic escaped Recovery middleware")
		}
	}()

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "ping"})
	afterRecoveryCalled = true

	if !afterRecoveryCalled {
		t.Error("execution should continue after Dispatch when Recovery is installed")
	}
}

func TestRecovery_NonPanicPathUnaffected(t *testing.T) {
	called := false

	rtr := router.New()
	rtr.Use(router.Recovery())
	rtr.On("ping", func(_ *router.Context) { called = true })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "ping"})

	if !called {
		t.Error("handler should be called when no panic occurs")
	}
}

func TestRecovery_ChainContinuesAfterRecovery(t *testing.T) {
	// Register a second middleware AFTER Recovery to verify Next flows correctly
	// when there is no panic.
	secondCalled := false

	rtr := router.New()
	rtr.Use(router.Recovery())
	rtr.Use(func(ctx *router.Context) { secondCalled = true; ctx.Next() })
	rtr.On("ping", func(_ *router.Context) {})

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "ping"})

	if !secondCalled {
		t.Error("subsequent middleware should be called when Recovery is installed and no panic occurs")
	}
}

func TestRecovery_PanicWithNilValue(t *testing.T) {
	defer func() {
		if recover() != nil {
			t.Error("nil panic should not escape Recovery middleware")
		}
	}()

	rtr := router.New()
	rtr.Use(router.Recovery())
	rtr.On("ping", func(_ *router.Context) { panic(nil) }) //nolint:gocritic

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Type: "ping"})
}
