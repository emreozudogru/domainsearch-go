package output

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"ds_go/internal/checker"
)

// Writer writes check results in a specific format.
type Writer interface {
	Write(checker.Result) error
	Flush() error
}

// New returns a Writer for the given format ("text", "json", or "csv").
func New(format string, w io.Writer) (Writer, error) {
	switch format {
	case "text":
		return NewText(w), nil
	case "json":
		return NewJSON(w), nil
	case "csv":
		return NewCSV(w), nil
	default:
		return nil, fmt.Errorf("unknown format %q (use text, json, csv)", format)
	}
}

// Text writer: one human-readable line per result.
type Text struct {
	w  *bufio.Writer
	mu sync.Mutex
}

func NewText(w io.Writer) *Text { return &Text{w: bufio.NewWriter(w)} }

func (t *Text) Write(r checker.Result) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	status := "TAKEN"
	if r.Available {
		status = "AVAILABLE"
	}
	extra := ""
	if r.Err != nil {
		extra = fmt.Sprintf(" (error: %s)", r.Err.Error())
	}
	_, err := fmt.Fprintf(t.w, "%s\t%s%s\n", r.Domain, status, extra)
	return err
}

func (t *Text) Flush() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Flush()
}

// JSON writer: a streaming JSON array.
type JSON struct {
	w     *bufio.Writer
	mu    sync.Mutex
	count int
}

type jsonRec struct {
	Domain    string `json:"domain"`
	TLD       string `json:"tld"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

func NewJSON(w io.Writer) *JSON { return &JSON{w: bufio.NewWriter(w)} }

func (j *JSON) Write(r checker.Result) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	rec := jsonRec{Domain: r.Domain, TLD: r.TLD, Available: r.Available}
	if r.Err != nil {
		rec.Error = r.Err.Error()
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	if j.count == 0 {
		fmt.Fprint(j.w, "[")
	} else {
		fmt.Fprint(j.w, ",")
	}
	j.count++
	_, err = j.w.Write(b)
	return err
}

func (j *JSON) Flush() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.count == 0 {
		_, err := fmt.Fprint(j.w, "[]\n")
		if err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprint(j.w, "]\n"); err != nil {
			return err
		}
	}
	return j.w.Flush()
}

// CSV writer: a header row followed by one row per result.
type CSV struct {
	w    *csv.Writer
	mu   sync.Mutex
	head bool
}

func NewCSV(w io.Writer) *CSV { return &CSV{w: csv.NewWriter(w)} }

func (c *CSV) Write(r checker.Result) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.head {
		if err := c.w.Write([]string{"domain", "tld", "available", "error"}); err != nil {
			return err
		}
		c.head = true
	}
	avail := "false"
	if r.Available {
		avail = "true"
	}
	errStr := ""
	if r.Err != nil {
		errStr = r.Err.Error()
	}
	return c.w.Write([]string{r.Domain, r.TLD, avail, errStr})
}

func (c *CSV) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.w.Flush()
	return c.w.Error()
}
