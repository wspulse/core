# Benchmarks

The table below is a checked-in baseline for `wspulse/core`, last refreshed
locally by a maintainer running `make bench-sync` on the hardware noted in
the table. The CI workflow runs the same benchmarks on every code PR (the
workflow ignores Markdown / LICENSE / `.github/instructions/**` changes)
and uploads the raw `bench.txt` as an artefact; download it from the run
page if you need to compare specific numbers between branches. CI does not
regenerate or commit this file.

Variance between machines is expected — these baselines are a regression
sanity check, not a portability claim. Single runs at `-benchtime=3s -count=1`
are noisy at the few-percent level; a one-off `bench-sync` is enough for
order-of-magnitude regression detection but not micro-optimisation.

The matrix covers:

- `JSONCodec Encode` / `Decode` — wire-format encoding cost across
  `messageSize ∈ {64 B, 1 KiB, 16 KiB}`. Shared by every wspulse SDK that
  speaks the JSON protocol; allocation count especially matters.
- `Message struct alloc` — bare struct construction + payload assignment
  written into a package-level sink (so the compiler cannot elide the loop).
  Watchdog bench: should report 0 allocs (the sink is a fixed slot, not a
  per-iteration allocation, and the payload is just a slice-header copy);
  a regression to non-zero is a red flag.
- `Router Dispatch` — event routing cost across `routes ∈ {1, 10, 100}`
  (number of registered events, each with one terminal handler) and
  `middleware ∈ {0, 3}` (global middleware depth prepended to every chain).

<!-- benchsync:core:start -->
Measured on `darwin/arm64` (`Apple M1 Max`).

| Operation | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| `JSONCodec Encode (64 B)` | 392.5 | 144 | 2 |
| `JSONCodec Encode (1 KiB)` | 3,877 | 1,200 | 2 |
| `JSONCodec Encode (16 KiB)` | 59,985 | 18,496 | 2 |
| `JSONCodec Decode (64 B)` | 640.9 | 336 | 7 |
| `JSONCodec Decode (1 KiB)` | 3,520 | 1,296 | 7 |
| `JSONCodec Decode (16 KiB)` | 48,873 | 16,656 | 7 |
| `Message struct alloc` | 1.245 | 0 | 0 |
| `Router Dispatch (1 route, 0 mw)` | 19.69 | 0 | 0 |
| `Router Dispatch (1 route, 3 mw)` | 29.86 | 0 | 0 |
| `Router Dispatch (10 routes, 0 mw)` | 29.3 | 0 | 0 |
| `Router Dispatch (10 routes, 3 mw)` | 36.52 | 0 | 0 |
| `Router Dispatch (100 routes, 0 mw)` | 30.44 | 0 | 0 |
| `Router Dispatch (100 routes, 3 mw)` | 39.86 | 0 | 0 |
<!-- benchsync:core:end -->
