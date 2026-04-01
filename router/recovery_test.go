package router_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wspulse "github.com/wspulse/core"
	"github.com/wspulse/core/router"
)

func TestRecovery_CatchesPanicAndContinues(t *testing.T) {
	rtr := router.New()
	rtr.Use(router.Recovery())
	rtr.On("ping", func(_ *router.Context) { panic("boom") })

	// Dispatch must not panic to the caller; it must be swallowed by Recovery.
	require.NotPanics(t, func() {
		rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "ping"})
	})
}

func TestRecovery_NonPanicPathUnaffected(t *testing.T) {
	called := false

	rtr := router.New()
	rtr.Use(router.Recovery())
	rtr.On("ping", func(_ *router.Context) { called = true })

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "ping"})

	assert.True(t, called, "handler should be called when no panic occurs")
}

func TestRecovery_ChainContinuesAfterRecovery(t *testing.T) {
	// Register a second middleware AFTER Recovery to verify Next flows correctly
	// when there is no panic.
	secondCalled := false

	rtr := router.New()
	rtr.Use(router.Recovery())
	rtr.Use(func(ctx *router.Context) { secondCalled = true; ctx.Next() })
	rtr.On("ping", func(_ *router.Context) {})

	rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "ping"})

	assert.True(t, secondCalled,
		"subsequent middleware should be called when Recovery is installed and no panic occurs")
}

func TestRecovery_PanicWithNilValue(t *testing.T) {
	rtr := router.New()
	rtr.Use(router.Recovery())
	rtr.On("ping", func(_ *router.Context) { panic(nil) }) //nolint:gocritic

	require.NotPanics(t, func() {
		rtr.Dispatch(newMockConnection("c1", "r1"), wspulse.Frame{Event: "ping"})
	})
}
