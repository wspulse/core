package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBenchResults_CoreNames(t *testing.T) {
	t.Parallel()

	input := []byte(`
goos: darwin
goarch: arm64
cpu: Apple M1 Max
BenchmarkJSONCodecEncode/messageSize=64B-10            5000000      720 ns/op    256 B/op    3 allocs/op
BenchmarkJSONCodecDecode/messageSize=16KiB-10            50000     6500 ns/op   4500 B/op    7 allocs/op
BenchmarkMessageAlloc-10                            500000000     2.5 ns/op       0 B/op    0 allocs/op
BenchmarkRouterDispatch/handlers=100/middleware=3-10   100000     1530 ns/op    400 B/op    9 allocs/op
PASS
`)

	got, err := parseBenchResults(input)
	require.NoError(t, err)

	assert.Equal(t, "720", got.Results["BenchmarkJSONCodecEncode/messageSize=64B"].NSPerOp)
	assert.Equal(t, "6500", got.Results["BenchmarkJSONCodecDecode/messageSize=16KiB"].NSPerOp)
	assert.Equal(t, "2.5", got.Results["BenchmarkMessageAlloc"].NSPerOp)
	assert.Equal(t, "1530", got.Results["BenchmarkRouterDispatch/handlers=100/middleware=3"].NSPerOp)
}

func TestRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "doc"), 0o755))

	benchFile := filepath.Join(root, "bench.txt")
	require.NoError(t, os.WriteFile(benchFile, []byte(`
goos: darwin
goarch: arm64
cpu: Apple M1 Max
BenchmarkJSONCodecEncode/messageSize=64B-10            5000000     720 ns/op    256 B/op    3 allocs/op
BenchmarkJSONCodecEncode/messageSize=1KiB-10           1000000    1500 ns/op   1300 B/op    3 allocs/op
BenchmarkJSONCodecEncode/messageSize=16KiB-10           100000   12000 ns/op  17000 B/op    3 allocs/op
BenchmarkJSONCodecDecode/messageSize=64B-10            3000000     820 ns/op    320 B/op    5 allocs/op
BenchmarkJSONCodecDecode/messageSize=1KiB-10            800000    1700 ns/op   1400 B/op    5 allocs/op
BenchmarkJSONCodecDecode/messageSize=16KiB-10            50000    6500 ns/op   4500 B/op    7 allocs/op
BenchmarkMessageAlloc-10                            500000000     2.5 ns/op       0 B/op    0 allocs/op
BenchmarkRouterDispatch/handlers=1/middleware=0-10    2000000     560 ns/op    256 B/op    2 allocs/op
BenchmarkRouterDispatch/handlers=1/middleware=3-10    1500000     720 ns/op    288 B/op    5 allocs/op
BenchmarkRouterDispatch/handlers=10/middleware=0-10   1500000     650 ns/op    280 B/op    3 allocs/op
BenchmarkRouterDispatch/handlers=10/middleware=3-10   1000000     820 ns/op    320 B/op    6 allocs/op
BenchmarkRouterDispatch/handlers=100/middleware=0-10   500000    1230 ns/op    320 B/op    6 allocs/op
BenchmarkRouterDispatch/handlers=100/middleware=3-10   100000    1530 ns/op    400 B/op    9 allocs/op
`), 0o644))

	docPath := filepath.Join(root, "doc", "bench.md")
	require.NoError(t, os.WriteFile(docPath, []byte(strings.TrimSpace(`
# Benchmarks

<!-- benchsync:core:start -->
stale
<!-- benchsync:core:end -->
`)+"\n"), 0o644))

	require.NoError(t, run([]string{"-root", root, "-input", benchFile}))

	got, err := os.ReadFile(docPath)
	require.NoError(t, err)
	doc := string(got)

	assert.Contains(t, doc, "Measured on `darwin/arm64` (`Apple M1 Max`).")
	assert.Contains(t, doc, "| `JSONCodec Encode (64 B)` | 720 | 256 | 3 |")
	assert.Contains(t, doc, "| `JSONCodec Decode (16 KiB)` | 6,500 | 4,500 | 7 |")
	assert.Contains(t, doc, "| `Message struct alloc` | 2.5 | 0 | 0 |")
	assert.Contains(t, doc, "| `Router Dispatch (1 handler, 0 mw)` | 560 | 256 | 2 |")
	assert.Contains(t, doc, "| `Router Dispatch (100 handlers, 3 mw)` | 1,530 | 400 | 9 |")
}

func TestRun_MissingResult(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "doc"), 0o755))

	benchFile := filepath.Join(root, "bench.txt")
	require.NoError(t, os.WriteFile(benchFile, []byte(`
BenchmarkMessageAlloc-10  100 100 ns/op 0 B/op 0 allocs/op
`), 0o644))

	docPath := filepath.Join(root, "doc", "bench.md")
	require.NoError(t, os.WriteFile(docPath, []byte("<!-- benchsync:core:start -->\n<!-- benchsync:core:end -->\n"), 0o644))

	err := run([]string{"-root", root, "-input", benchFile})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing benchmark result")
}

func TestFormatMetric(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "107", formatMetric("107.0"))
	assert.Equal(t, "5.59", formatMetric("5.590"))
	assert.Equal(t, "1,626", formatMetric("1626"))
	assert.Equal(t, "3.007", formatMetric("3.007"))
}
