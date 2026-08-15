package checker

import (
	"errors"
	"testing"
	"time"
)

// fakeLookup blocks when it returns false (slow path) so we can simulate timeouts.
func makeLookup(blockCount int, delay time.Duration, avail bool) Lookup {
	calls := 0
	return func(domain string) bool {
		calls++
		if calls <= blockCount {
			time.Sleep(delay) // exceeds short timeouts
		}
		return avail
	}
}

func TestCheckerRetryOnTimeoutThenSuccess(t *testing.T) {
	lookup := makeLookup(3, 150*time.Millisecond, true) // blocks first 3 calls
	c := NewChecker(50*time.Millisecond, 3, lookup)
	res := c.Check("x.us", ".us")
	if !res.Available {
		t.Errorf("expected available=true after retries, got %+v", res)
	}
	if res.Err != nil {
		t.Errorf("expected nil error, got %v", res.Err)
	}
}

func TestCheckerRetriesExhausted(t *testing.T) {
	lookup := makeLookup(99, 150*time.Millisecond, true) // always slow
	c := NewChecker(50*time.Millisecond, 2, lookup)
	res := c.Check("x.us", ".us")
	if res.Available {
		t.Error("expected available=false on exhausted retries")
	}
	if !errors.Is(res.Err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got %v", res.Err)
	}
}

func TestCheckerFastLookupSucceeds(t *testing.T) {
	lookup := makeLookup(0, 0, true)
	c := NewChecker(time.Second, 3, lookup)
	res := c.Check("x.us", ".us")
	if !res.Available || res.Err != nil {
		t.Errorf("expected success, got %+v", res)
	}
}
