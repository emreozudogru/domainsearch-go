# domainsearch-go

A Go command-line tool that checks the availability of domain names by reading a wordlist
file and appending one or more top-level domains (TLDs) to each entry. It uses the
[`github.com/haccer/available`](https://github.com/haccer/available) package to determine
whether each resulting domain is available for registration.

## Features

- Reads wordlist entries from a text file (one word per line).
- Checks **multiple configurable TLDs** per word (e.g. `.us`, `.com`, `.net`).
- Runs checks **concurrently** with a configurable worker pool.
- **Rate-limited** to avoid overwhelming the lookup service.
- Reports **live progress** on stderr (available count, items checked).
- Supports **structured output formats**: plain text, JSON, and CSV.
- Writes results to a file or to stdout.
- Handles `SIGINT`/`SIGTERM` gracefully — cancels in-flight checks.

## Requirements

- [Go](https://go.dev/) 1.21 or later.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/emreozudogru/domainsearch-go.git
   cd domainsearch-go
   ```

2. Download dependencies and build:

   ```bash
   go mod tidy
   go build -o bin/domainsearch ./cmd/domainsearch
   ```

   Or use the provided build script:

   ```bash
   ./scripts/build.sh
   ```

## Usage

```bash
./bin/domainsearch [flags]
```

All flags are optional. Defaults read from `assets/wtzl.txt`, check `.us` domains, and
print results to stdout.

### Flags

| Flag / Shorthand        | Default              | Description                                                |
|-------------------------|----------------------|------------------------------------------------------------|
| `-i, --input`           | `assets/wtzl.txt`    | Path to the wordlist file (one word per line).             |
| `-t, --tlds`            | `.us`                | Comma-separated list of TLDs, e.g. `.us,.com,.net`.         |
| `-f, --format`          | `text`               | Output format: `text`, `json`, or `csv`.                    |
| `-o, --output`          | (stdout)             | Output file path. If empty, writes to stdout.               |
| `-r, --rate`            | `5`                  | Maximum domain checks per second (rate limit).              |
| `-w, --workers`         | `10`                 | Number of concurrent worker goroutines.                     |
| `-v, --verbose`         | `false`              | Enable verbose logging on stderr.                           |
| `--no-progress`         | `false`              | Disable the progress bar on stderr.                         |
| `--cache`               | (none)               | Result cache file (enables caching and resume across runs). |
| `--cache-ttl`           | `24h`                | Cache entry freshness window.                               |
| `--timeout`             | `15s`                | Per-lookup timeout.                                         |
| `--retries`             | `2`                  | Number of retries on a lookup timeout.                      |
| `-a, --available-only`  | `false`              | Only write available domains to output.                     |

### Examples

Check `.us` and `.com` for every word, print text results to stdout (with live progress
on stderr):

```bash
./bin/domainsearch --tlds .us,.com
```

Save JSON results to a file, no progress bar:

```bash
./bin/domainsearch --tlds .us,.com,.net --format json --output results.json --no-progress
```

Save CSV results with higher concurrency and a tighter rate limit:

```bash
./bin/domainsearch -t .com -f csv -o results.csv -w 20 -r 10
```

Only list available domains (to a file, with a 20-second per-lookup timeout and
2 retries on timeout):

```bash
./bin/domainsearch -t .us,.com,.net -a -f text -o available.txt --timeout 20s --retries 2
```

**Caching and resume** — persist results to a file so re-runs skip already-checked
domains (great for long wordlists interrupted mid-run):

```bash
# First run: performs the lookups and writes the cache.
./bin/domainsearch -t .us,.com -i words.txt --cache ~/.cache/domainsearch.jsonl --no-progress

# Later runs: cached domains are reused; only new/entries re-check.
./bin/domainsearch -t .us,.com -i words.txt --cache ~/.cache/domainsearch.jsonl --no-progress
```

A summary line is always printed to stderr, e.g.:

```
Summary: checked=2000  available=3  taken=1997  errors=0  in 12m34s
```

### Output Formats

- **text** — one line per domain: `domain\tSTATUS` (or `domain (error: ...)`).
- **json** — a streaming JSON array of objects
  `{"domain","tld","available","error"}`.
- **csv** — a header row followed by `domain,tld,available,error`.

### Wordlist Format

The wordlist is a text file with one word per line. Blank lines are skipped. Replace the
contents of `assets/wtzl.txt` or pass your own path with `--input`.

## Project Structure

```
domainsearch-go/
├── .github/
│   └── workflows/
│       └── ci.yml             # CI: build, vet, test, lint
├── AGENTS.md                  # Guidelines for automated agents
├── Makefile                   # Common dev targets (build, test, lint, …)
├── README.md                  # This file
├── LICENSE                    # GNU GPL v3
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── .gitignore                 # Ignores /bin/, binaries, etc.
├── .golangci.yml              # golangci-lint configuration
├── assets/
│   └── wtzl.txt               # Bundled wordlist (one word per line)
├── bin/                       # Build output (gitignored)
├── cmd/
│   └── domainsearch/
│       └── main.go            # Entry point + cobra CLI
├── internal/
│   ├── checker/               # Domain availability lookup (wraps haccer/available)
│   ├── config/                # cobra CLI flags + Config struct + validation
│   ├── output/                # Text/JSON/CSV result writers
│   └── worker/                # Concurrent worker pool + rate limiter
└── scripts/
    └── build.sh               # Convenience build script → bin/domainsearch
```

## Dependencies

- [`github.com/haccer/available`](https://github.com/haccer/available) — domain availability
  lookups.
- [`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate) — rate limiting.
- [`github.com/spf13/cobra`](https://github.com/spf13/cobra) +
  [`spf13/pflag`](https://github.com/spf13/pflag) — command-line interface.
- [`github.com/schollz/progressbar/v3`](https://github.com/schollz/progressbar) — progress bar.

## Known Limitations

- Lookups rely on the `github.com/haccer/available` package's own logic; non-timeout
  errors are reported per domain but not retried (only timeouts are retried).
- The cache is a local file; it is not shared across machines.
- Very large wordlists with many TLDs can still take a long time due to network latency.

## Development

Contributions are welcome. To develop locally:

1. Fork and clone the repository.
2. Install dependencies and build:
   ```bash
   go mod tidy
   make build      # or: go build ./cmd/domainsearch
   ```
3. Run unit tests and checks:
   ```bash
   make test       # go test ./...
   make vet        # go vet ./...
   make fmt        # gofmt -l . && go mod tidy
   make lint       # golangci-lint run
   ```
4. Run the tool:
   ```bash
   make run
   ```

Continuous integration (`.github/workflows/ci.yml`) runs the build, vet, test, and lint
steps on every push and pull request.

### Releases

Prebuilt cross-platform binaries can be produced with
[GoReleaser](https://goreleaser.com/) using the included `.goreleaser.yml`:

```bash
goreleaser release --clean
```

Prebuilt binaries are also available on the GitHub Releases page.

Follow the commit conventions described in [AGENTS.md](AGENTS.md). Commit directly to `master`
(no feature branches) and push after each task.

## License

This program is free software: you can redistribute it and/or modify it under the terms of
the GNU General Public License as published by the Free Software Foundation, either version
3 of the License, or (at your option) any later version.

See [LICENSE](LICENSE) for the full license text.
