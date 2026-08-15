# domainsearch-go

A simple Go command-line tool that checks the availability of domain names by reading a
wordlist file and appending a top-level domain (TLD) to each entry. It uses the
[`github.com/haccer/available`](https://github.com/haccer/available) package to query
whether each resulting domain is available for registration.

## Features

- Reads wordlist entries from a text file (one word per line).
- Appends a configurable TLD to each word.
- Checks each domain for availability via the `available` package.
- Prints status for each domain to standard output.

## Requirements

- [Go](https://go.dev/) 1.21 or later (any recent version works).

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/emreozudogru/domainsearch-go.git
   cd domainsearch-go
   ```

2. Download dependencies:

   ```bash
   go mod tidy
   ```

3. Build the binary:

   ```bash
   go build -o ds_go
   ```

## Usage

The program reads words from `wtzl.txt` (located in the working directory) and appends the
hardcoded TLD `.us` to each word. It then checks and prints the availability for every
generated domain.

```bash
./ds_go
```

Expected output (per domain):

```
example.us looking...
example.us is EMPTY
```

### Wordlist Format

The wordlist (`wtzl.txt`) is a text file with one word per line. Each line is read
sequentially. Blank lines are included as-is. To customize the list, replace the contents of
`wtzl.txt` or modify the path in `ds_go.go`.

### Changing the TLD

The TLD is configured in `ds_go.go`:

```go
ext := []string{".us"}
```

To check multiple TLDs, extend the slice:

```go
ext := []string{".us", ".com", ".net"}
```

Then rebuild:

```bash
go build -o ds_go
```

## Project Structure

| File       | Description                                              |
|------------|----------------------------------------------------------|
| `ds_go.go` | Main program: reads wordlist, checks domain availability. |
| `wtzl.txt` | Wordlist file (one word per line).                       |
| `go.mod`   | Go module definition.                                    |
| `go.sum`   | Dependency checksums.                                    |
| `AGENTS.md`| Guidelines for automated agents operating in this repo.   |
| `LICENSE`  | GNU General Public License v3.0.                         |

## Dependencies

- [`github.com/haccer/available`](https://github.com/haccer/available) — checks domain
  availability.

## Known Limitations

- Only one hardcoded TLD is used by default.
- Domain checks run sequentially (no concurrency).
- No rate limiting on external lookups.
- No structured output format (JSON, CSV, or file output).
- File path and options are hardcoded; no CLI flags.

## Development

Contributions are welcome. To develop locally:

1. Fork and clone the repository.
2. Build with `go build`.
3. Run with `./ds_go`.
4. After making changes, build and verify the output.

Follow the commit conventions described in [AGENTS.md](AGENTS.md). Commit directly to
`master` (no feature branches) and push after each task.

## License

This program is free software: you can redistribute it and/or modify it under the terms of
the GNU General Public License as published by the Free Software Foundation, either version
3 of the License, or (at your option) any later version.

See [LICENSE](LICENSE) for the full license text.
