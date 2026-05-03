package router_test

import (
	"fmt"
	"strings"
	"testing"

	wspulse "github.com/wspulse/core"
	"github.com/wspulse/core/router"
)

// benchConn is a noop Connection implementation used by the dispatch bench.
// The router does not call any of these methods during a normal Dispatch
// (handlers may, but the bench handlers do not).
type benchConn struct{}

func (benchConn) ID() string                   { return "bench-conn" }
func (benchConn) RoomID() string               { return "bench-room" }
func (benchConn) Send(_ wspulse.Message) error { return nil }
func (benchConn) Close() error                 { return nil }
func (benchConn) Done() <-chan struct{}        { return nil }

// jsonPayload returns a valid JSON string payload of total byte length size.
func jsonPayload(size int) []byte {
	if size < 2 {
		size = 2
	}
	return []byte(`"` + strings.Repeat("x", size-2) + `"`)
}

// BenchmarkRouterDispatch measures the cost of one Dispatch call.
//
// The two axes:
//   - `routes`: number of distinct events registered (router map size).
//     Tests lookup cost as the registry grows. Each registered event has
//     one terminal handler.
//   - `middleware`: number of global middleware funcs in the chain.
//     Tests per-handler chain execution cost. Total chain length per
//     Dispatch is middleware + 1.
//
// The bench dispatches the middle-indexed event each iteration to keep
// map-bucket access pattern stable across runs. A warmup Dispatch is
// issued before ResetTimer so the one-time lazy chain build and first
// sync.Pool allocation are not charged to the first iteration.
func BenchmarkRouterDispatch(b *testing.B) {
	noop := func(*router.Context) {}

	for _, routeCount := range []int{1, 10, 100} {
		for _, mwDepth := range []int{0, 3} {
			name := fmt.Sprintf("routes=%d/middleware=%d", routeCount, mwDepth)
			b.Run(name, func(b *testing.B) {
				r := router.New()
				for i := 0; i < mwDepth; i++ {
					r.Use(noop)
				}

				events := make([]string, routeCount)
				for i := 0; i < routeCount; i++ {
					ev := fmt.Sprintf("event%d", i)
					events[i] = ev
					r.On(ev, noop)
				}

				conn := benchConn{}
				msg := wspulse.Message{
					Event:   events[routeCount/2],
					Payload: jsonPayload(64),
				}

				// Warmup: triggers lazy chain build and primes sync.Pool so
				// neither cost is charged to the first timed iteration.
				r.Dispatch(conn, msg)

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r.Dispatch(conn, msg)
				}
			})
		}
	}
}
