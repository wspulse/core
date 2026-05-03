# Benchmarks

The table below is a checked-in baseline for `wspulse/core`, last refreshed
locally by a maintainer running `make bench-sync` on the hardware noted in
the table. The CI workflow runs the same benchmarks on every PR and uploads
the raw `bench.txt` as an artefact; download it from the run page if you
need to compare specific numbers between branches. CI does not regenerate
or commit this file.

Variance between machines is expected — these baselines are a regression
sanity check, not a portability claim. Single runs at `-benchtime=3s -count=1`
are noisy at the few-percent level; a one-off `bench-sync` is enough for
order-of-magnitude regression detection but not micro-optimisation.

The matrix covers:

- `JSONCodec Encode` / `Decode` — wire-format encoding cost across
  `messageSize ∈ {64 B, 1 KiB, 16 KiB}`. Shared by every wspulse SDK that
  speaks the JSON protocol; allocation count especially matters.
- `Message struct alloc` — bare struct construction + payload assignment.
  Watchdog bench: should report 0 allocs (Go escape analysis keeps the value
  on the stack); a regression to non-zero is a red flag.
- `Router Dispatch` — event routing cost across `handlers ∈ {1, 10, 100}`
  (router map size, one terminal handler per event) and `middleware ∈ {0, 3}`
  (global middleware depth prepended to every chain).

<!-- benchsync:core:start -->
Measured on `darwin/arm64` (`Apple M1 Max`).

| Operation | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| `JSONCodec Encode (64 B)` | 390.3 | 144 | 2 |
| `JSONCodec Encode (1 KiB)` | 3,833 | 1,200 | 2 |
| `JSONCodec Encode (16 KiB)` | 58,294 | 18,496 | 2 |
| `JSONCodec Decode (64 B)` | 633 | 336 | 7 |
| `JSONCodec Decode (1 KiB)` | 3,477 | 1,296 | 7 |
| `JSONCodec Decode (16 KiB)` | 48,459 | 16,656 | 7 |
| `Message struct alloc` | 0.3164 | 0 | 0 |
| `Router Dispatch (1 handlers, 0 mw)` | 18.88 | 0 | 0 |
| `Router Dispatch (1 handlers, 3 mw)` | 23.27 | 0 | 0 |
| `Router Dispatch (10 handlers, 0 mw)` | 28.86 | 0 | 0 |
| `Router Dispatch (10 handlers, 3 mw)` | 32.39 | 0 | 0 |
| `Router Dispatch (100 handlers, 0 mw)` | 33.05 | 0 | 0 |
| `Router Dispatch (100 handlers, 3 mw)` | 33.73 | 0 | 0 |
<!-- benchsync:core:end -->
