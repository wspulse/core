package wspulse_test

import (
	"testing"

	wspulse "github.com/wspulse/core"
)

// messageAllocSink is package-level on purpose: writes to it are observable
// across the program, which prevents the compiler from eliminating the
// loop body in BenchmarkMessageAlloc as dead code. A function-local var
// (with `_ = sink` at the end) is not enough — the previous baseline of
// 0.3164 ns/op (sub-cycle on Apple M1 Max) confirmed the assignments were
// being optimised away.
var messageAllocSink wspulse.Message

// BenchmarkMessageAlloc measures the cost of constructing a Message struct
// with a payload reference. The bench is allocation-watchdog: a regression
// adding even one allocation is amplified across millions of operations
// downstream (every Send and every Dispatch in the workspace constructs or
// passes a Message).
func BenchmarkMessageAlloc(b *testing.B) {
	payload := jsonPayload(64)
	b.ReportAllocs()
	for b.Loop() {
		messageAllocSink = wspulse.Message{Event: "bench", Payload: payload}
	}
}
