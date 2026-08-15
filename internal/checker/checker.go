package checker

import "github.com/haccer/available"

// Result represents the outcome of checking a single domain.
type Result struct {
	Domain    string // full domain checked, e.g. "example.us"
	TLD       string // TLD portion, e.g. ".us"
	Available bool   // whether the domain is available
	Err       error  // any error encountered
}

// Check queries whether domain is available for registration.
func Check(domain, tld string) Result {
	avail := available.Domain(domain)
	return Result{
		Domain:    domain,
		TLD:       tld,
		Available: avail,
	}
}
