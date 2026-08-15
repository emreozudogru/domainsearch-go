package config

import "time"

// Config holds all runtime options for a domain search run. It is populated by
// the cobra command in cli.go.
type Config struct {
	InputPath     string        // path to the wordlist file (one word per line)
	OutputPath    string        // output file path (stdout when empty)
	Tlds          []string      // TLDs to check, e.g. [".us", ".com"]
	Format        string        // "text", "json", or "csv"
	Rate          int           // max domain checks per second (rate limit)
	Workers       int           // number of concurrent worker goroutines
	NoProgress    bool          // disable the progress bar on stderr
	Verbose       bool          // enable verbose (debug) logging on stderr
	CachePath     string        // path to the result cache file (empty disables caching)
	CacheTTL      time.Duration // cache entry freshness window
	Retries       int           // lookup retries on timeout
	Timeout       time.Duration // per-lookup timeout
	AvailableOnly bool          // only emit available domains
	Charset       string        // charset preset or literal alphabet (empty = use wordlist)
	MinLen        int           // shortest label to generate in charset mode
	MaxLen        int           // longest label to generate in charset mode
}

// Validate applies defaults and basic validation to the Config.
func (c *Config) Validate() {
	if len(c.Tlds) == 0 {
		c.Tlds = []string{".us"}
	}
	if c.Rate < 1 {
		c.Rate = 1
	}
	if c.Workers < 1 {
		c.Workers = 1
	}
	if c.CacheTTL <= 0 {
		c.CacheTTL = 24 * time.Hour
	}
	if c.Retries < 0 {
		c.Retries = 0
	}
	if c.Timeout <= 0 {
		c.Timeout = 15 * time.Second
	}
	if c.MinLen < 1 {
		c.MinLen = 1
	}
	if c.MaxLen < c.MinLen {
		c.MaxLen = c.MinLen
	}
}
