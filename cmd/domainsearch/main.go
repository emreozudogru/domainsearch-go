package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"

	"ds_go/internal/cache"
	"ds_go/internal/checker"
	"ds_go/internal/config"
	"ds_go/internal/output"
	"ds_go/internal/worker"
)

func main() {
	cfg := &config.Config{}
	cmd := config.NewRootCmd(cfg, run)
	// Default logger writes only errors to stderr; run() raises the level in verbose mode.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// run executes the domain search described by cfg.
func run(cfg *config.Config) error {
	if cfg.Verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}
	slog.Debug("starting",
		"input", cfg.InputPath,
		"tlds", cfg.Tlds,
		"workers", cfg.Workers,
		"rate", cfg.Rate,
		"format", cfg.Format,
		"cache", cfg.CachePath,
		"timeout", cfg.Timeout,
		"retries", cfg.Retries,
	)

	words, err := readWordlist(cfg.InputPath)
	if err != nil {
		return fmt.Errorf("read wordlist %q: %w", cfg.InputPath, err)
	}
	if len(words) == 0 {
		return fmt.Errorf("wordlist %q is empty", cfg.InputPath)
	}

	out, cleanup, err := buildWriter(cfg)
	if err != nil {
		return fmt.Errorf("open output: %w", err)
	}
	defer func() {
		_ = out.Flush()
		cleanup()
	}()

	// Optional result cache (also provides resume across runs).
	var store *cache.Store
	if cfg.CachePath != "" {
		store = cache.NewStore(cfg.CachePath, cfg.CacheTTL)
		if err := store.Load(); err != nil {
			slog.Warn("cache load failed; continuing without cache", "err", err)
			store = nil
		}
	}

	// Build the lookup checker with timeout and retries.
	ch := checker.NewChecker(cfg.Timeout, cfg.Retries, checker.DefaultLookup)
	check := checker.CheckFunc(ch.Check)
	if store != nil {
		check = cache.Wrap(check, store)
	}

	const bufSize = 256
	jobs := make(chan string, bufSize)
	results := make(chan checker.Result, bufSize)
	p := worker.New(jobs, results, cfg.Workers, cfg.Rate, check)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var bar *progressbar.ProgressBar
	if !cfg.NoProgress {
		total := len(words) * len(cfg.Tlds)
		bar = progressbar.NewOptions(total,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetDescription("Checking domains"),
			progressbar.OptionSetWidth(20),
		)
	}

	// Feeder: push words into the jobs channel.
	go func() {
		defer close(jobs)
		for _, w := range words {
			select {
			case <-ctx.Done():
				return
			case jobs <- w:
			}
		}
	}()

	// Consumer: drain results into the output writer.
	var stats stats
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for res := range results {
			stats.record(res)
			// available-only filter: skip emitting taken/error results.
			if !(cfg.AvailableOnly && !res.Available) {
				_ = out.Write(res)
			}
			if cfg.Verbose {
				slog.Debug("checked", "domain", res.Domain, "available", res.Available, "err", res.Err)
			}
			if bar != nil {
				_ = bar.Add(1)
			}
		}
	}()

	start := time.Now()
	p.Run(ctx, cfg.Tlds)
	p.Wait() // workers finish, results channel closed
	wg.Wait()

	if store != nil {
		if err := store.Save(); err != nil {
			slog.Warn("cache save failed", "err", err)
		}
	}

	elapsed := time.Since(start)
	slog.Debug("done", "elapsed", elapsed, "checked", stats.checked, "available", stats.available)
	printSummary(stats, elapsed)
	return nil
}

// stats tracks per-run counters.
type stats struct {
	checked   int64
	available int64
	taken     int64
	failed    int64
}

func (s *stats) record(r checker.Result) {
	atomic.AddInt64(&s.checked, 1)
	if r.Err != nil {
		atomic.AddInt64(&s.failed, 1)
		return
	}
	if r.Available {
		atomic.AddInt64(&s.available, 1)
	} else {
		atomic.AddInt64(&s.taken, 1)
	}
}

// printSummary writes the run summary to stderr.
func printSummary(s stats, elapsed time.Duration) {
	fmt.Fprintf(os.Stderr,
		"\nSummary: checked=%d  available=%d  taken=%d  errors=%d  in %s\n",
		s.checked, s.available, s.taken, s.failed, elapsed.Round(time.Millisecond),
	)
}

// readWordlist reads non-empty, trimmed lines from path.
func readWordlist(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var words []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		words = append(words, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return words, nil
}

// buildWriter constructs the output Writer based on config.
func buildWriter(cfg *config.Config) (output.Writer, func(), error) {
	var w io.Writer
	var cleanup func()

	if cfg.OutputPath == "" {
		w = os.Stdout
		cleanup = func() {}
	} else {
		f, err := os.Create(cfg.OutputPath)
		if err != nil {
			return nil, nil, err
		}
		w = f
		cleanup = func() { _ = f.Close() }
	}

	out, err := output.New(cfg.Format, w)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return out, cleanup, nil
}
