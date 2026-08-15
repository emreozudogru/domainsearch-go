package cache

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"ds_go/internal/checker"
)

// entry is one line in the JSON-lines cache file.
type entry struct {
	Domain    string `json:"domain"`
	TLD       string `json:"tld"`
	Available bool   `json:"available"`
	Err       string `json:"error,omitempty"`
	CheckedAt int64  `json:"checked_at"` // unix milliseconds
}

// Store is a file-backed cache of lookup results with a TTL. It is safe for
// concurrent use.
type Store struct {
	file string
	ttl  time.Duration
	mu   sync.Mutex
	data map[string]entry
}

// NewStore returns a Store that persists to file and treats entries older than
// ttl as stale.
func NewStore(file string, ttl time.Duration) *Store {
	return &Store{file: file, ttl: ttl, data: make(map[string]entry)}
}

// Load reads the cache file. A non-existent file is treated as an empty cache
// (first run) rather than an error.
func (s *Store) Load() error {
	f, err := os.Open(s.file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open cache %q: %w", s.file, err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e entry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			continue
		}
		s.data[e.Domain] = e
	}
	return sc.Err()
}

// Get returns a fresh cached result for domain. ok is false when the domain is
// missing or its entry is older than the TTL.
func (s *Store) Get(domain string) (checker.Result, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.data[domain]
	if !ok {
		return checker.Result{}, false
	}
	if time.Since(time.UnixMilli(e.CheckedAt)) > s.ttl {
		return checker.Result{}, false
	}
	var err error
	if e.Err != "" {
		err = errors.New(e.Err)
	}
	return checker.Result{Domain: e.Domain, TLD: e.TLD, Available: e.Available, Err: err}, true
}

// Put stores a result in the cache.
func (s *Store) Put(domain string, r checker.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	errStr := ""
	if r.Err != nil {
		errStr = r.Err.Error()
	}
	s.data[domain] = entry{
		Domain:    domain,
		TLD:       r.TLD,
		Available: r.Available,
		Err:       errStr,
		CheckedAt: time.Now().UnixMilli(),
	}
}

// Count returns the number of entries currently held in memory.
func (s *Store) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}

// Save persists the cache to disk.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Create(s.file)
	if err != nil {
		return fmt.Errorf("create cache %q: %w", s.file, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range s.data {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}

// Wrap returns a CheckFunc that serves cached results when available, falling
// back to next for misses and storing fresh results. This gives the tool an
// implicit resume capability across runs.
func Wrap(next checker.CheckFunc, s *Store) checker.CheckFunc {
	return func(domain, tld string) checker.Result {
		if r, ok := s.Get(domain); ok {
			return r
		}
		r := next(domain, tld)
		s.Put(domain, r)
		return r
	}
}
