package router_test

import (
	"testing"

	wspulse "github.com/wspulse/core"
	"github.com/wspulse/core/router"
)

func TestContext_Next_ExecutesAllHandlersInOrder(t *testing.T) {
	order := make([]int, 0, 3)

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { order = append(order, 1); ctx.Next() })
	rtr.Use(func(ctx *router.Context) { order = append(order, 2); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { order = append(order, 3) })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("unexpected execution order: %v", order)
	}
}

func TestContext_Abort_StopsChain(t *testing.T) {
	secondCalled := false

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Abort() })
	rtr.On("ping", func(ctx *router.Context) { secondCalled = true })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if secondCalled {
		t.Error("handler should not have been called after Abort")
	}
}

func TestContext_IsAborted_FalseBeforeAbort(t *testing.T) {
	var wasAborted bool

	rtr := router.New()
	rtr.On("ping", func(ctx *router.Context) { wasAborted = ctx.IsAborted() })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if wasAborted {
		t.Error("IsAborted should be false before Abort is called")
	}
}

func TestContext_IsAborted_TrueAfterAbort(t *testing.T) {
	var abortedInMiddleware bool

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) {
		ctx.Abort()
		abortedInMiddleware = ctx.IsAborted()
	})
	rtr.On("ping", func(_ *router.Context) {})

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if !abortedInMiddleware {
		t.Error("IsAborted should be true after Abort is called")
	}
}

func TestContext_Set_Get_RoundTrip(t *testing.T) {
	var got any
	var exists bool

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Set("userID", "alice"); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { got, exists = ctx.Get("userID") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if !exists || got != "alice" {
		t.Errorf("expected (alice, true), got (%v, %v)", got, exists)
	}
}

func TestContext_Get_MissingKey(t *testing.T) {
	var exists bool

	rtr := router.New()
	rtr.On("ping", func(ctx *router.Context) { _, exists = ctx.Get("missing") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if exists {
		t.Error("Get should return false for a missing key")
	}
}

func TestContext_MustGet_ReturnsValue(t *testing.T) {
	var got any

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Set("role", "admin"); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { got = ctx.MustGet("role") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if got != "admin" {
		t.Errorf("expected %q, got %v", "admin", got)
	}
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

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if !panicked {
		t.Error("MustGet should panic for a missing key")
	}
}

func TestContext_GetString_ReturnsString(t *testing.T) {
	var got string

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Set("name", "bob"); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { got = ctx.GetString("name") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if got != "bob" {
		t.Errorf("expected %q, got %q", "bob", got)
	}
}

func TestContext_GetString_MissingKeyReturnsEmpty(t *testing.T) {
	var got string

	rtr := router.New()
	rtr.On("ping", func(ctx *router.Context) { got = ctx.GetString("missing") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestContext_GetString_WrongTypeReturnsEmpty(t *testing.T) {
	var got string

	rtr := router.New()
	rtr.Use(func(ctx *router.Context) { ctx.Set("count", 42); ctx.Next() })
	rtr.On("ping", func(ctx *router.Context) { got = ctx.GetString("count") })

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if got != "" {
		t.Errorf("expected empty string for non-string value, got %q", got)
	}
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

	rtr.Dispatch(newMockConnection("c1", "room1"), wspulse.Frame{Type: "ping"})

	if got != "second" {
		t.Errorf("expected %q, got %v", "second", got)
	}
}

func TestContext_ConnectionAndFrameAccessible(t *testing.T) {
	var gotID, gotRoomID, gotFrameType string

	rtr := router.New()
	rtr.On("chat", func(ctx *router.Context) {
		gotID = ctx.Connection.ID()
		gotRoomID = ctx.Connection.RoomID()
		gotFrameType = ctx.Frame.Type
	})

	rtr.Dispatch(
		newMockConnection("conn-1", "room-A"),
		wspulse.Frame{Type: "chat"},
	)

	if gotID != "conn-1" {
		t.Errorf("expected ID %q, got %q", "conn-1", gotID)
	}
	if gotRoomID != "room-A" {
		t.Errorf("expected RoomID %q, got %q", "room-A", gotRoomID)
	}
	if gotFrameType != "chat" {
		t.Errorf("expected FrameType %q, got %q", "chat", gotFrameType)
	}
}
