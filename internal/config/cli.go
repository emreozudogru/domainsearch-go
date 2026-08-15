package config

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the cobra root command with all flags attached. The provided
// Config is populated from the parsed flags, and run is invoked when the command
// executes. Passing run in avoids an import cycle with the main package.
func NewRootCmd(cfg *Config, run func(*Config) error) *cobra.Command {
	var tldsFlag string

	cmd := &cobra.Command{
		Use:   "domainsearch [wordlist]",
		Short: "Check domain name availability from a wordlist",
		Long: `domainsearch checks the availability of domain names by reading a wordlist
file and appending one or more top-level domains (TLDs) to each entry.

It uses github.com/haccer/available for the actual lookups and runs checks
concurrently with an optional rate limit.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				cfg.InputPath = args[0]
			}
			cfg.Tlds = parseTLDs(tldsFlag)
			cfg.Validate()
			return run(cfg)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&cfg.InputPath, "input", "i", "assets/wtzl.txt", "path to the wordlist file (one word per line)")
	flags.StringVarP(&tldsFlag, "tlds", "t", ".us", "comma-separated list of TLDs, e.g. .us,.com,.net")
	flags.StringVarP(&cfg.OutputPath, "output", "o", "", "output file path (default: stdout)")
	flags.StringVarP(&cfg.Format, "format", "f", "text", "output format: text, json, csv")
	flags.IntVarP(&cfg.Rate, "rate", "r", 5, "maximum domain checks per second")
	flags.IntVarP(&cfg.Workers, "workers", "w", 10, "number of concurrent workers")
	flags.BoolVar(&cfg.NoProgress, "no-progress", false, "disable the progress bar on stderr")
	flags.BoolVarP(&cfg.Verbose, "verbose", "v", false, "enable verbose logging on stderr")
	flags.StringVar(&cfg.CachePath, "cache", "", "path to result cache file (enables caching/resume; empty disables)")
	flags.DurationVar(&cfg.CacheTTL, "cache-ttl", 24*time.Hour, "cache entry freshness window")
	flags.IntVar(&cfg.Retries, "retries", 2, "number of retries on a lookup timeout")
	flags.DurationVar(&cfg.Timeout, "timeout", 15*time.Second, "per-lookup timeout")
	flags.BoolVarP(&cfg.AvailableOnly, "available-only", "a", false, "only write available domains to output")
	return cmd
}

// parseTLDs splits a comma-separated list of TLDs, ensuring each has a leading dot.
func parseTLDs(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		t := strings.TrimSpace(part)
		if t == "" {
			continue
		}
		if !strings.HasPrefix(t, ".") {
			t = "." + t
		}
		out = append(out, t)
	}
	return out
}
