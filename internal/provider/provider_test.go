package provider

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestParseAlphabetPresets(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"letters", AlphabetLetters},
		{"DIGITS", AlphabetDigits},
		{"lettersdigits", AlphabetLetters + AlphabetDigits},
		{"all", AlphabetLetters + AlphabetDigits + AlphabetSymbols},
		{"sym", ""}, // unknown preset -> literal (non-empty here)
	}
	for _, c := range cases {
		got := ParseAlphabet(c.in)
		if c.in == "sym" {
			if got != "sym" {
				t.Errorf("ParseAlphabet(%q) = %q, want literal %q", c.in, got, "sym")
			}
			continue
		}
		if got != c.want {
			t.Errorf("ParseAlphabet(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCountCombos(t *testing.T) {
	// base=2 (a,b), lengths 1..2: 2 + 4 = 6
	if got := countCombos(2, 1, 2); got != 6 {
		t.Errorf("countCombos(2,1,2) = %d, want 6", got)
	}
	// base=10, length 1: 10
	if got := countCombos(10, 1, 1); got != 10 {
		t.Errorf("countCombos(10,1,1) = %d, want 10", got)
	}
	// absurd size -> 0 (progress disabled)
	if got := countCombos(36, 1, 10); got != 0 {
		t.Errorf("expected 0 for huge space, got %d", got)
	}
}

func TestCharsetGeneratesAllCombos(t *testing.T) {
	alpha := "ab"
	cs := Charset{Alphabet: alpha, MinLen: 1, MaxLen: 2}
	ch, total, err := cs.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if total != 6 { // 2 (len1) + 4 (len2)
		t.Fatalf("total = %d, want 6", total)
	}
	var got []string
	for w := range ch {
		got = append(got, w)
	}
	want := []string{"a", "b", "aa", "ab", "ba", "bb"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCharsetEmptyAlphabetErrors(t *testing.T) {
	cs := Charset{Alphabet: "", MinLen: 1, MaxLen: 2}
	if _, _, err := cs.Open(context.Background()); err == nil {
		t.Fatal("expected error for empty alphabet")
	}
}

func TestCharsetCancelsOnContext(t *testing.T) {
	cs := Charset{Alphabet: "abcdef", MinLen: 1, MaxLen: 2}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, _, err := cs.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Consume a few, then cancel. The generator must wind down and close the
	// channel without deadlocking.
	for i := 0; i < 3; i++ {
		if _, ok := <-ch; !ok {
			return
		}
	}
	cancel()
	done := make(chan struct{})
	go func() {
		for range ch {
		}
		close(done)
	}()
	select {
	case <-done:
		// generator stopped and closed the channel
	case <-time.After(2 * time.Second):
		t.Fatal("channel did not close after cancel (possible deadlock)")
	}
}

func TestWordlistOpen(t *testing.T) {
	// Use the bundled wordlist: ensure it opens and has a sane total.
	wl := Wordlist{Path: "../../assets/wtzl.txt"}
	ch, total, err := wl.Open(context.Background())
	if err != nil {
		t.Skipf("assets file not reachable from test cwd: %v", err)
	}
	if total == 0 {
		t.Fatal("expected non-zero total for bundled wordlist")
	}
	first, ok := <-ch
	if !ok {
		t.Fatal("expected at least one word")
	}
	// Labels must be non-empty and printable ASCII without spaces.
	if first == "" || strings.ContainsAny(first, " \n\r\t") {
		t.Errorf("unexpected first label %q", first)
	}
	// Drain the rest so the producer goroutine exits cleanly.
	for range ch {
	}
}
