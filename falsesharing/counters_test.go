package falsesharing

import (
	"testing"
	"unsafe"
)

// increments is how many times each counter is bumped per benchmark operation.
const increments = 5_000_000

// reportThroughput turns the result into millions of atomic adds per second, so
// that a slower run reads as a smaller number.
func reportThroughput(b *testing.B, ops int) {
	b.ReportMetric(float64(ops)/b.Elapsed().Seconds()/1e6, "Madds/s")
}

func BenchmarkOneWorker(b *testing.B) {
	var c Counters
	ops := 0
	for b.Loop() {
		oneWorker(&c, increments)
		ops += 2 * increments
	}
	reportThroughput(b, ops)
}

func BenchmarkTwoWorkers(b *testing.B) {
	var c Counters
	ops := 0
	for b.Loop() {
		twoWorkers(&c, increments)
		ops += 2 * increments
	}
	reportThroughput(b, ops)
}

func BenchmarkTwoWorkersPadded(b *testing.B) {
	var c PaddedCounters
	ops := 0
	for b.Loop() {
		twoWorkersPadded(&c, increments)
		ops += 2 * increments
	}
	reportThroughput(b, ops)
}

// TestWorkersAgree guards the honesty of the comparison: every variant must
// perform exactly the same number of increments.
func TestWorkersAgree(t *testing.T) {
	const n = 1000

	var one, two Counters
	oneWorker(&one, n)
	twoWorkers(&two, n)

	var padded PaddedCounters
	twoWorkersPadded(&padded, n)

	for _, got := range []int64{
		one.A.Load(), one.B.Load(),
		two.A.Load(), two.B.Load(),
		padded.A.Load(), padded.B.Load(),
	} {
		if got != n {
			t.Errorf("counter = %d, want %d", got, n)
		}
	}
}

// TestPaddingSeparatesCacheLines fails if the padding stops doing its job, for
// example after a change to cacheLineSize or to the size of atomic.Int64.
func TestPaddingSeparatesCacheLines(t *testing.T) {
	var p PaddedCounters

	if got, want := unsafe.Sizeof(p), uintptr(2*cacheLineSize); got != want {
		t.Errorf("sizeof(PaddedCounters) = %d, want %d", got, want)
	}

	gap := unsafe.Offsetof(p.B) - unsafe.Offsetof(p.A)
	if gap != cacheLineSize {
		t.Errorf("A and B are %d bytes apart, want %d", gap, cacheLineSize)
	}

	var c Counters
	if gap := unsafe.Offsetof(c.B) - unsafe.Offsetof(c.A); gap >= cacheLineSize {
		t.Errorf("unpadded counters are %d bytes apart: the demo has no false sharing left", gap)
	}
}
