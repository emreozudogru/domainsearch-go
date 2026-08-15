package worker

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"ds_go/internal/checker"
)

// Pool is a fixed-size pool of workers that check domains concurrently,
// respecting a shared rate limiter.
type Pool struct {
	jobs      <-chan string
	results   chan<- checker.Result
	workers   int
	check     checker.CheckFunc
	limiter   *rate.Limiter
	waitGroup sync.WaitGroup
}

// New creates a Pool that reads words from jobs, checking each word appended
// with every TLD, and writes results to the results channel.
func New(jobs <-chan string, results chan<- checker.Result, workers, ratePerSec int, check checker.CheckFunc) *Pool {
	if workers < 1 {
		workers = 1
	}
	if ratePerSec < 1 {
		ratePerSec = 1
	}
	if check == nil {
		def := checker.NewChecker(15*time.Second, 2, checker.DefaultLookup)
		check = def.Check
	}
	return &Pool{
		jobs:    jobs,
		results: results,
		workers: workers,
		check:   check,
		limiter: rate.NewLimiter(rate.Every(time.Second), ratePerSec),
	}
}

// Run starts the worker goroutines.
func (p *Pool) Run(ctx context.Context, tlds []string) {
	for w := 0; w < p.workers; w++ {
		p.waitGroup.Add(1)
		go p.worker(ctx, tlds)
	}
}

// Wait blocks until all workers have finished, then closes the results channel.
func (p *Pool) Wait() {
	p.waitGroup.Wait()
	close(p.results)
}

func (p *Pool) worker(ctx context.Context, tlds []string) {
	defer p.waitGroup.Done()
	for word := range p.jobs {
		for _, tld := range tlds {
			domain := word + tld
			if err := p.limiter.Wait(ctx); err != nil {
				sendResult(ctx, p.results, checker.Result{Domain: domain, TLD: tld, Err: err})
				return
			}
			sendResult(ctx, p.results, p.check(domain, tld))
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

// sendResult delivers a result unless the context was cancelled first.
func sendResult(ctx context.Context, ch chan<- checker.Result, r checker.Result) {
	select {
	case <-ctx.Done():
	case ch <- r:
	}
}
