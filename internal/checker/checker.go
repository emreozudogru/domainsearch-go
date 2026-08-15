package checker

import (
	"errors"
	"time"

	"github.com/haccer/available"
)

// ErrTimeout is returned when a lookup exceeds its configured timeout.
var ErrTimeout = errors.New("domain lookup timed out")

// Result represents the outcome of checking a single domain.
type Result struct {
	Domain    string // full domain checked, e.g. "example.us"
	TLD       string // TLD portion, e.g. ".us"
	Available bool   // whether the domain is available
	Err       error  // any error encountered
}

// CheckFunc resolves whether a domain (already suffixed with a TLD) is
// available. It is injectable so the Checker and tests can substitute
// implementations.
type CheckFunc func(domain, tld string) Result

// Lookup queries whether a domain is available for registration. It is
// injectable so the Checker can be unit tested without network access.
type Lookup func(domain string) bool

// Checker performs availability checks with a per-lookup timeout and
// retry-on-timeout behaviour.
type Checker struct {
	Timeout time.Duration
	Retries int
	Lookup  Lookup
}

// DefaultLookup is the production lookup implementation backed by
// github.com/haccer/available.
func DefaultLookup(domain string) bool {
	return available.Domain(domain)
}

// NewChecker builds a Checker with sane defaults for timeout and retries.
func NewChecker(timeout time.Duration, retries int, lookup Lookup) *Checker {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	if retries < 0 {
		retries = 0
	}
	if lookup == nil {
		lookup = DefaultLookup
	}
	return &Checker{Timeout: timeout, Retries: retries, Lookup: lookup}
}

// Check queries domain availability. On timeout it retries up to Retries times.
func (c *Checker) Check(domain, tld string) Result {
	b, err := c.lookupWithTimeout(domain)
	for attempt := 0; err != nil && attempt < c.Retries; attempt++ {
		b, err = c.lookupWithTimeout(domain)
	}
	return Result{Domain: domain, TLD: tld, Available: b, Err: err}
}

// lookupWithTimeout runs the lookup in a goroutine bounded by a timeout. If the
// timeout fires the lookup goroutine is abandoned and may leak until the
// underlying call returns; this is acceptable because real lookups are bounded
// by their own network timeouts.
func (c *Checker) lookupWithTimeout(domain string) (bool, error) {
	type res struct {
		b bool
	}
	ch := make(chan res, 1)
	go func() {
		ch <- res{b: c.Lookup(domain)}
	}()
	select {
	case r := <-ch:
		return r.b, nil
	case <-time.After(c.Timeout):
		return false, ErrTimeout
	}
}
