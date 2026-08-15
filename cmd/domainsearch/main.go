package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"

	"ds_go/internal/cache"
	"ds_go/internal/checker"
	"ds_go/internal/config"
	"ds_go/internal/output"
	"ds_go/internal/provider"
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
		"fmt", cfg.Format,
		"charset", cfg.Charset,
		"cache", cfg.CachePath,
		"timeout", cfg.Timeout,
		"retries", cfg.Retries,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// The lookup dependency writes diagnostic messages directly to os.Stderr.
	// Capture them into the logger (silenced unless verbose) so our own stderr
	// output (progress + summary) stays clean. Our own writers use the saved
	// realStderr.
	realStderr, restoreStderr := captureLibraryStderr()
	defer restoreStderr()

	// Choose the label source: a wordlist file, or a charset generator.
	var src provider.Source = provider.Wordlist{Path: cfg.InputPath}
	if cfg.Charset != "" {
		src = provider.Charset{
			Alphabet: provider.ParseAlphabet(cfg.Charset),
			MinLen:   cfg.MinLen,
			MaxLen:   cfg.MaxLen,
		}
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

	// Build the lookup checker with timeout and retry.
	ch := checker.NewChecker(cfg.Timeout, cfg.Retries, checker.DefaultLookup)
	check := checker.CheckFunc(ch.Check)
	if store != nil {
		check = cache.Wrap(check, store)
	}

	const bufSize = 256
	jobs := make(chan string, bufSize)
	results := make(chan checker.Result, bufSize)
	p := worker.New(jobs, results, cfg.Workers, cfg.Rate, check)

	stream, totalLabels, err := src.Open(ctx)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	totalChecks := totalLabels * len(cfg.Tlds)

	var bar *progressbar.ProgressBar
	if !cfg.NoProgress && totalChecks > 0 {
		bar = progressbar.NewOptions(totalChecks,
			progressbar.OptionSetWriter(realStderr),
			progressbar.OptionSetDescription("Checking domains"),
			progressbar.OptionSetWidth(20),
		)
	}

	// Feeder: push labels into the jobs channel.
	go func() {
		defer close(jobs)
		for w := range stream {
			select {
			case <-ctx.Done():
				return
			case jobs <- w:
			}
		}
	}()

	// Consumer: drain results into the output writer.
	var st stats
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for res := range results {
			st.record(res)
			// available-only filter: skip taken/error results from output.
			if !(cfg.AvailableOnly && !res.Available) {
				_ = out.Write(res)
			}
			slog.Debug("checked", "domain", res.Domain, "available", res.Available, "err", res.Err)
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
	slog.Debug("done", "elapsed", elapsed, "checked", st.checked, "available", st.available)
	printSummary(realStderr, st, elapsed, totalChecks)
	return nil
}

// stats tracks per-run counters.
type stats struct {
	checked, available, taken, failed int64
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

// printSummary writes the run summary to w (stderr).
func printSummary(w io.Writer, s stats, elapsed time.Duration, total int) {
	fmt.Fprintf(w,
		"\nSummary: checked=%d/%d  available=%d  taken=%d  errors=%d  in %s\n",
		s.checked, total, s.available, s.taken, s.failed, elapsed.Round(time.Millisecond),
	)
}

// captureLibraryStderr redirects writes to os.Stderr (used directly by the whois
// lookup dependency) into the logger, and returns the original stderr for our
// own progress/summary output. The restore func puts os.Stderr back and stops
// the drain goroutine.
func captureLibraryStderr() (realStderr *os.File, restore func()) {
	realStderr = os.Stderr
	pr, pw, err := os.Pipe()
	if err != nil {
		return realStderr, func() {}
	}
	os.Stderr = pw
	go func() {
		sc := bufio.NewScanner(pr)
		for sc.Scan() {
			slog.Warn(sc.Text())
		}
	}()
	restore = func() {
		_ = pw.Close()
		os.Stderr = realStderr
		_ = pr.Close()
	}
	return realStderr, restore
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
