# Cheap goroutines, expensive cache misses

Code and benchmarks for the talk **"Goroutines baratas, cache misses caros: concurrencia en
Go al nivel del silicio"** (*Cheap goroutines, expensive cache misses: Go concurrency down
at the silicon*), presented at [Nerdearla 2026](https://nerdearla.com).

Writing concurrent code in Go is deceptively simple: `go func()` and you are done. But
between that line and the cores that run it there is a whole world that decides whether
your code actually runs in parallel or quietly fights itself. This repository contains the
experiments run live during the talk, so you can reproduce them on your own machine.

## The experiment

Two goroutines that share no state, hold no lock, pass `-race` clean... and are twice as
slow as a single goroutine doing exactly the same amount of work.

```go
type Counters struct {
	A atomic.Int64 // incremented only by goroutine 1
	B atomic.Int64 // incremented only by goroutine 2
}
```

`A` and `B` are 8 bytes apart, so they live on the same **cache line**. To write its own
field, each core needs to own the whole line, and the line bounces between cores on every
increment. That is *false sharing*: the data is independent, but the hardware cannot tell.

The fix does not touch the program's logic: push the fields onto separate lines.

```go
type PaddedCounters struct {
	A atomic.Int64
	_ [cacheLineSize - 8]byte
	B atomic.Int64
	_ [cacheLineSize - 8]byte
}
```

`cacheLineSize` is resolved with build tags: **64 bytes on amd64, 128 on arm64** (Apple
Silicon). 

## Running it

Requires **Go 1.26** or newer.

```bash
git clone https://github.com/diegolatorre16/nerdearla-2026-mex.git
cd nerdearla-2026-mex

# Tests: every variant performs the same amount of work, and the padding
# really keeps the fields on separate cache lines.
go test ./...

# The hook: one goroutine vs two (the padded variant stays hidden)
go test -bench 'OneWorker|TwoWorkers$' -benchtime 3x -count 5 ./falsesharing

# No data race, no trick
go test -race -bench . -benchtime 1x ./falsesharing

# The reveal: all three variants
go test -bench . -benchtime 3x -count 5 ./falsesharing
```

Besides the standard `ns/op`, each benchmark reports `Madds/s` (millions of atomic
increments per second), so a slower run reads as a smaller number.

### Reference results

Apple M4, 10 cores, 128-byte cache lines:

```
BenchmarkOneWorker-10            16.4 ms/op    604 Madds/s
BenchmarkTwoWorkers-10           31.8 ms/op    314 Madds/s   <- two goroutines, half the speed
BenchmarkTwoWorkersPadded-10     12.3 ms/op    813 Madds/s
```

The exact milliseconds depend on your CPU, but the ratios hold: `TwoWorkers` is around 2x
slower than `OneWorker`, and `TwoWorkersPadded` is around 2.5x to 3x faster than
`TwoWorkers`.

### Measuring tips

- The first run of a session is always slow (CPU frequency scaling, cold caches). Use
  `-count 3` or more and read the steady-state rows.
- On laptops, run plugged in. On Apple Silicon, P-core/E-core scheduling shifts the numbers
  depending on the power state.
- Close your browser and other heavy applications before measuring.
- Expect about 10% run-to-run drift. To compare variants with statistical rigor, use
  [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

  ```bash
  go test -bench . -benchtime 3x -count 10 ./falsesharing | tee bench.txt
  benchstat bench.txt
  ```

## Repository layout

```
falsesharing/
  counters.go           Counters, PaddedCounters and the three worker variants
  counters_test.go      benchmarks, plus tests that keep the comparison honest
  cacheline_amd64.go    cacheLineSize = 64
  cacheline_arm64.go    cacheLineSize = 128
  cacheline_other.go    fallback to 64
```

## Further reading

- Scott Meyers, [*CPU Caches and Why You Care*](https://www.youtube.com/watch?v=WDIkqP4JbkE):
  the source of the line "if performance matters, the total memory you have is the total
  amount of cache".
- Ulrich Drepper, [*What Every Programmer Should Know About Memory*](https://people.freebsd.org/~lstewart/articles/cpumemory.pdf):
  the reference text on the memory hierarchy and cache coherence.
- [MESI protocol](https://en.wikipedia.org/wiki/MESI_protocol): how cores negotiate who
  owns each line.
- [`golang.org/x/sys/cpu.CacheLinePad`](https://pkg.go.dev/golang.org/x/sys/cpu#CacheLinePad):
  portable padding with the right size per architecture.
- [The Go Memory Model](https://go.dev/ref/mem): what the language guarantees and what is
  left to the hardware.

## Author

Diego Latorre ([@diegolatorre16](https://github.com/diegolatorre16))
