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
| `JSONCodec Encode (64 B)` | 389.4 | 144 | 2 |
| `JSONCodec Encode (1 KiB)` | 3,866 | 1,200 | 2 |
| `JSONCodec Encode (16 KiB)` | 58,321 | 18,495 | 2 |
| `JSONCodec Decode (64 B)` | 633.1 | 336 | 7 |
| `JSONCodec Decode (1 KiB)` | 3,475 | 1,296 | 7 |
| `JSONCodec Decode (16 KiB)` | 48,446 | 16,656 | 7 |
| `Message struct alloc` | 1.239 | 0 | 0 |
| `Router Dispatch (1 route, 0 mw)` | 18.84 | 0 | 0 |
| `Router Dispatch (1 route, 3 mw)` | 23.14 | 0 | 0 |
| `Router Dispatch (10 routes, 0 mw)` | 28.92 | 0 | 0 |
| `Router Dispatch (10 routes, 3 mw)` | 32.47 | 0 | 0 |
| `Router Dispatch (100 routes, 0 mw)` | 30.11 | 0 | 0 |
| `Router Dispatch (100 routes, 3 mw)` | 34.08 | 0 | 0 |
<!-- benchsync:core:end -->
