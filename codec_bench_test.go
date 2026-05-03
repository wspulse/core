package wspulse_test

import (
	"fmt"
	"strings"
	"testing"

	wspulse "github.com/wspulse/core"
)

// jsonPayload returns a valid JSON string payload of total byte length size.
// size includes the surrounding quotes; minimum is 2 (an empty JSON string).
func jsonPayload(size int) []byte {
	if size < 2 {
		size = 2
	}
	return []byte(`"` + strings.Repeat("x", size-2) + `"`)
}

// messageSizes is the standard payload size matrix for codec benchmarks.
// Values match the workspace bench-harness plan.
var messageSizes = []struct {
	label string
	size  int
}{
	{"64B", 64},
	{"1KiB", 1024},
	{"16KiB", 16 * 1024},
}

func BenchmarkJSONCodecEncode(b *testing.B) {
	for _, ms := range messageSizes {
		b.Run(fmt.Sprintf("messageSize=%s", ms.label), func(b *testing.B) {
			msg := wspulse.Message{Event: "bench", Payload: jsonPayload(ms.size)}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := wspulse.JSONCodec.Encode(msg)
				if err != nil {
					b.Fatalf("Encode: %v", err)
				}
			}
		})
	}
}

func BenchmarkJSONCodecDecode(b *testing.B) {
	for _, ms := range messageSizes {
		b.Run(fmt.Sprintf("messageSize=%s", ms.label), func(b *testing.B) {
			msg := wspulse.Message{Event: "bench", Payload: jsonPayload(ms.size)}
			encoded, err := wspulse.JSONCodec.Encode(msg)
			if err != nil {
				b.Fatalf("setup Encode: %v", err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := wspulse.JSONCodec.Decode(encoded); err != nil {
					b.Fatalf("Decode: %v", err)
				}
			}
		})
	}
}
