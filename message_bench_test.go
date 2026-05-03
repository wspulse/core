package wspulse_test

import (
	"testing"

	wspulse "github.com/wspulse/core"
)

// BenchmarkMessageAlloc measures the cost of constructing a Message struct
// with a payload reference. The bench is allocation-watchdog: a regression
// adding even one allocation is amplified across millions of operations
// downstream (every Send and every Dispatch in the workspace constructs or
// passes a Message).
//
// `sink` defeats Go's dead-store elimination so the construction is not
// optimised away.
func BenchmarkMessageAlloc(b *testing.B) {
	payload := jsonPayload(64)
	var sink wspulse.Message
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = wspulse.Message{Event: "bench", Payload: payload}
	}
	_ = sink
}
