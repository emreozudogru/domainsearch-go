package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"ds_go/internal/checker"
)

func TestPoolAllJobsProcessedAndConcurrencyBounded(t *testing.T) {
	words := []string{"a", "b", "c", "d", "e"}
	var seen int64
	var inFlight int64
	var maxInFlight int64

	check := func(domain, tld string) checker.Result {
		n := atomic.AddInt64(&inFlight, 1)
		atomic.AddInt64(&seen, 1)
		for {
			m := atomic.LoadInt64(&maxInFlight)
			if n <= m {
				break
			}
			if atomic.CompareAndSwapInt64(&maxInFlight, m, n) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt64(&inFlight, -1)
		return checker.Result{Domain: domain, TLD: tld, Available: true}
	}

	jobs := make(chan string, len(words))
	results := make(chan checker.Result, len(words))
	p := New(jobs, results, 3, 1000, check)

	ctx := context.Background()
	p.Run(ctx, []string{".us"})

	go func() {
		defer close(jobs)
		for _, w := range words {
			jobs <- w
		}
	}()
	go func() {
		for range results {
		}
	}()

	p.Wait()

	if seen != int64(len(words)) {
		t.Errorf("expected %d checks, got %d", len(words), seen)
	}
	if maxInFlight > 3 {
		t.Errorf("concurrency %d exceeded worker count 3", maxInFlight)
	}
}

func TestPoolRespectsContextCancel(t *testing.T) {
	check := func(domain, tld string) checker.Result {
		time.Sleep(50 * time.Millisecond)
		return checker.Result{Domain: domain, TLD: tld}
	}

	jobs := make(chan string, 1)
	results := make(chan checker.Result, 1)
	p := New(jobs, results, 1, 1000, check)

	ctx, cancel := context.WithCancel(context.Background())
	p.Run(ctx, []string{".us"})
	jobs <- "a"
	cancel()

	done := make(chan struct{})
	go func() { p.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Wait did not return after context cancel (possible deadlock)")
	}
}
