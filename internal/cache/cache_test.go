package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ds_go/internal/checker"
)

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "cache.jsonl")
	s := NewStore(file, 24*time.Hour)

	orig := checker.Result{Domain: "a.us", TLD: ".us", Available: true}
	s.Put("a.us", orig)

	// Not saved yet -> loading from disk sees nothing.
	s2 := NewStore(file, 24*time.Hour)
	if err := s2.Load(); err != nil {
		t.Fatal(err)
	}
	if r, ok := s2.Get("a.us"); ok {
		t.Fatalf("expected no cached result before save, got %+v", r)
	}

	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	s3 := NewStore(file, 24*time.Hour)
	if err := s3.Load(); err != nil {
		t.Fatal(err)
	}
	got, ok := s3.Get("a.us")
	if !ok {
		t.Fatal("expected cached result after save+load")
	}
	if !got.Available || got.Domain != "a.us" || got.TLD != ".us" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestStoreTTLExpiry(t *testing.T) {
	file := filepath.Join(t.TempDir(), "cache.jsonl")
	s := NewStore(file, 10*time.Millisecond)
	s.Put("a.us", checker.Result{Domain: "a.us", Available: true})
	if _, ok := s.Get("a.us"); !ok {
		t.Fatal("expected fresh entry to be cached")
	}
	time.Sleep(20 * time.Millisecond)
	if _, ok := s.Get("a.us"); ok {
		t.Fatal("expected stale entry to be treated as a miss")
	}
}

func TestStoreLoadMissingFile(t *testing.T) {
	s := NewStore(filepath.Join(t.TempDir(), "does-not-exist.jsonl"), time.Hour)
	if err := s.Load(); err != nil {
		t.Fatalf("expected nil error on missing cache file, got %v", err)
	}
	if s.Count() != 0 {
		t.Fatalf("expected empty cache, got %d", s.Count())
	}
}

func TestWrapUsesCacheThenFallback(t *testing.T) {
	file := filepath.Join(t.TempDir(), "cache.jsonl")
	s := NewStore(file, time.Hour)

	calls := 0
	next := func(domain, tld string) checker.Result {
		calls++
		return checker.Result{Domain: domain, TLD: tld, Available: true}
	}
	wrapped := Wrap(next, s)

	// First call: cache miss -> calls next, stores result.
	r1 := wrapped("a.us", ".us")
	if !r1.Available || calls != 1 {
		t.Fatalf("expected first call to hit next: %+v calls=%d", r1, calls)
	}
	// Second call: cache hit -> does not call next.
	r2 := wrapped("a.us", ".us")
	if calls != 1 {
		t.Fatalf("expected cached call to skip next, calls=%d", calls)
	}
	if !r2.Available {
		t.Fatal("cached result should be available")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
