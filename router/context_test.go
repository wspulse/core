package router_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wspulse "github.com/wspulse/core"
	"github.com/wspulse/core/router"
)

func TestContext_Next_ExecutesAllHandlersInOrder(t *testing.T) {
	order := make([]int, 0, 3)

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { order = append(order, 1); ctx.Next() })
	rtr.Use(func(ctx *router.Context) { order = append(order, 2); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { order = append(order, 3) })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestContext_Abort_StopsChain(t *testing.T) {
	secondCalled := false

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Abort() })
	rtr.On("ping", func(ctx *router.Context) { secondCalled = true })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.False(t, secondCalled, "handler should not have been called after Abort")
}

func TestContext_IsAborted_FalseBeforeAbort(t *testing.T) {
	var wasAborted bool

	rtr := router.New()
	rtr.On("ping", func(ctx *router.Context) { wasAborted = ctx.IsAborted() })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.False(t, wasAborted, "IsAborted should be false before Abort is called")
}

func TestContext_IsAborted_TrueAfterAbort(t *testing.T) {
	var abortedInMiddleware bool

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) {
		ctx.Abort()
		abortedInMiddleware = ctx.IsAborted()
	})
	rtr.On("ping", func(_ *router.Context) {})

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.True(t, abortedInMiddleware, "IsAborted should be true after Abort is called")
}

func TestContext_Set_Get_RoundTrip(t *testing.T) {
	var got any
	var exists bool

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Set("userID", "alice"); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { got, exists = ctx.Get("userID") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	require.True(t, exists)
	assert.Equal(t, "alice", got)
}

func TestContext_Get_MissingKey(t *testing.T) {
	var exists bool

	rtr := router.New()
	rtr.On("ping", func(ctx *router.Context) { _, exists = ctx.Get("missing") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.False(t, exists, "Get should return false for a missing key")
}

func TestContext_MustGet_ReturnsValue(t *testing.T) {
	var got any

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Set("role", "admin"); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { got = ctx.MustGet("role") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.Equal(t, "admin", got)
}

func TestContext_MustGet_PanicsOnMissingKey(t *testing.T) {
	panicked := false

	rtr := router.New()
	rtr.On("ping", func(ctx *router.Context) {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		ctx.MustGet("nonexistent")
	})

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.True(t, panicked, "MustGet should panic for a missing key")
}

func TestContext_GetString_ReturnsString(t *testing.T) {
	var got string

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Set("name", "bob"); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { got = ctx.GetString("name") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.Equal(t, "bob", got)
}

func TestContext_GetString_MissingKeyReturnsEmpty(t *testing.T) {
	var got string

	rtr := router.New()
	rtr.On("ping", func(ctx *router.Context) { got = ctx.GetString("missing") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.Empty(t, got)
}

func TestContext_GetString_WrongTypeReturnsEmpty(t *testing.T) {
	var got string

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Set("count", 42); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { got = ctx.GetString("count") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.Empty(t, got, "expected empty string for non-string value")
}

func TestContext_Set_OverwritesExistingKey(t *testing.T) {
	var got any

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) {
		ctx.Set("key", "first")
		ctx.Set("key", "second")
		ctx.Next()
	})
	rtr.On("ping", func(ctx *router.Context) { got, _ = ctx.Get("key") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Message{Event: "ping"})

	assert.Equal(t, "second", got)
}

func TestContext_ConnectionAndMessageAccessible(t *testing.T) {
	var gotID, gotRoomID, gotMessageEvent string

	rtr := router.New()
	rtr.On("chat", func(ctx *router.Context) {
		gotID = ctx.Connection.ID()
		gotRoomID = ctx.Connection.RoomID()
		gotMessageEvent = ctx.Message.Event
	})

	rtr.Dispatch(
		newMockConnection("conn-1", "room-A"),
		wspulse.Message{Event: "chat"},
	)

	assert.Equal(t, "conn-1", gotID)
	assert.Equal(t, "room-A", gotRoomID)
	assert.Equal(t, "chat", gotMessageEvent)
}
