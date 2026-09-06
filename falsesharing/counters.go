// Package falsesharing shows two goroutines that share no state and are still
// slower than a single goroutine doing the same amount of work.
package falsesharing

import (
	"sync"
	"sync/atomic"
)

// Counters holds two independent counters. Each worker owns exactly one field;
// neither ever reads or writes the other's.
type Counters struct {
	A atomic.Int64 // 8 bytes
	B atomic.Int64 // 8 bytes
}

// oneWorker performs 2*n increments from a single goroutine.
func oneWorker(c *Counters, n int) {
	for range n {
		c.A.Add(1)
	}
	for range n {
		c.B.Add(1)
	}
}

// twoWorkers performs the same 2*n increments, split across two goroutines that
// never touch the same field.
func twoWorkers(c *Counters, n int) {
	var wg sync.WaitGroup

	wg.Go(func() {
		for range n {
			c.A.Add(1)
		}
	})

	wg.Go(func() {
		for range n {
			c.B.Add(1)
		}
	})

	wg.Wait()
}

// PaddedCounters is Counters with each counter alone on its own cache line.
type PaddedCounters struct {
	A atomic.Int64
	_ [cacheLineSize - 8]byte
	B atomic.Int64
	_ [cacheLineSize - 8]byte
}

// twoWorkersPadded is twoWorkers, character for character, on PaddedCounters.
func twoWorkersPadded(c *PaddedCounters, n int) {
	var wg sync.WaitGroup

	wg.Go(func() {
		for range n {
			c.A.Add(1)
		}
	})

	wg.Go(func() {
		for range n {
			c.B.Add(1)
		}
	})

	wg.Wait()
}
