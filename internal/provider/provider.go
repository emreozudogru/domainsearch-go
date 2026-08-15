package provider

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// maxTotalForProgress caps the total reported for the progress bar. Beyond this
// size the bar is disabled (the run could take impractically long anyway).
const maxTotalForProgress = 5_000_000

// Source produces a stream of domain labels (without a TLD) and the total count
// of labels that will be produced.
type Source interface {
	Open(ctx context.Context) (<-chan string, int, error)
}

// Wordlist reads labels from a file (one per line; blank lines are skipped).
type Wordlist struct {
	Path string
}

func (w Wordlist) Open(ctx context.Context) (<-chan string, int, error) {
	f, err := os.Open(w.Path)
	if err != nil {
		return nil, 0, fmt.Errorf("open wordlist %q: %w", w.Path, err)
	}
	words, err := scanLines(f)
	_ = f.Close()
	if err != nil {
		return nil, 0, fmt.Errorf("read wordlist %q: %w", w.Path, err)
	}

	out := make(chan string, 256)
	go func() {
		defer close(out)
		for _, w := range words {
			select {
			case <-ctx.Done():
				return
			case out <- w:
			}
		}
	}()
	return out, len(words), nil
}

// Charset generates every combination of characters from Alphabet with lengths
// in [MinLen, MaxLen] (shortest first). Labels are plain ASCII, so no
// internationalized domain names are produced.
type Charset struct {
	Alphabet string
	MinLen   int
	MaxLen   int
}

func (c Charset) Open(ctx context.Context) (<-chan string, int, error) {
	if c.Alphabet == "" {
		return nil, 0, fmt.Errorf("charset alphabet must not be empty")
	}
	min := c.MinLen
	if min < 1 {
		min = 1
	}
	max := c.MaxLen
	if max < min {
		max = min
	}

	total := countCombos(len(c.Alphabet), min, max)

	out := make(chan string, 256)
	go func() {
		defer close(out)
		chars := c.Alphabet
		for length := min; length <= max; length++ {
			combine("", length, chars, out, ctx)
		}
	}()
	return out, total, nil
}

// combine emits all len-deep combinations of chars as prefix+remaining.
func combine(prefix string, depth int, chars string, out chan<- string, ctx context.Context) {
	if depth == 0 {
		select {
		case <-ctx.Done():
		case out <- prefix:
		}
		return
	}
	for i := 0; i < len(chars); i++ {
		combine(prefix+string(chars[i]), depth-1, chars, out, ctx)
	}
}

// countCombos returns sum_{L=minLen..maxLen} base^L, or 0 on overflow / when the
// space is too large to report a meaningful progress total.
func countCombos(base, minLen, maxLen int) int {
	total := 0
	for length := minLen; length <= maxLen; length++ {
		p := 1
		for i := 0; i < length; i++ {
			p *= base
			if p > maxTotalForProgress {
				return 0
			}
		}
		total += p
		if total < 0 || total > maxTotalForProgress {
			return 0
		}
	}
	return total
}

func scanLines(f *os.File) ([]string, error) {
	var words []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		words = append(words, line)
	}
	return words, sc.Err()
}

// Preset alphabets.
const (
	AlphabetLetters = "abcdefghijklmnopqrstuvwxyz"
	AlphabetDigits  = "0123456789"
	AlphabetSymbols = "-_"
)

// ParseAlphabet resolves a charset flag value into a concrete alphabet.
// Recognized presets: letters, digits, lettersdigits, symbols, all.
// Any other (non-empty) string is treated as a literal alphabet.
func ParseAlphabet(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "letters":
		return AlphabetLetters
	case "digits":
		return AlphabetDigits
	case "lettersdigits":
		return AlphabetLetters + AlphabetDigits
	case "symbols":
		return AlphabetSymbols
	case "all":
		return AlphabetLetters + AlphabetDigits + AlphabetSymbols
	default:
		return raw
	}
}
