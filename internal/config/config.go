package config

// Config holds all runtime options for a domain search run.
type Config struct {
	InputPath  string   // path to the wordlist file (one word per line)
	OutputPath string   // output file path (stdout when empty)
	Tlds       []string // TLDs to check, e.g. [".us", ".com"]
	Format     string   // "text", "json", or "csv"
	Rate       int      // max domain checks per second (rate limit)
	Workers    int      // number of concurrent worker goroutines
	NoProgress bool     // disable the progress bar on stderr
	Verbose    bool     // enable verbose (debug) logging on stderr
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
}
