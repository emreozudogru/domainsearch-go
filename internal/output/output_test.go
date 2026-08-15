package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"ds_go/internal/checker"
)

func TestText(t *testing.T) {
	var buf bytes.Buffer
	w := NewText(&buf)
	if err := w.Write(checker.Result{Domain: "a.us", TLD: ".us", Available: true}); err != nil {
		t.Fatal(err)
	}
	if err := w.Write(checker.Result{Domain: "b.us", TLD: ".us", Available: false, Err: errors.New("timeout")}); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "a.us\tAVAILABLE") {
		t.Errorf("missing AVAILABLE line: %q", got)
	}
	if !strings.Contains(got, "b.us\tTAKEN (error: timeout)") {
		t.Errorf("missing taken+error line: %q", got)
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSON(&buf)
	if err := w.Write(checker.Result{Domain: "a.us", TLD: ".us", Available: true}); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	var out []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json %q: %v", buf.String(), err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 record, got %d", len(out))
	}
	if out[0]["available"] != true {
		t.Errorf("expected available true, got %v", out[0]["available"])
	}
}

func TestJSONEmpty(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSON(&buf)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Errorf("expected [], got %q", buf.String())
	}
}

func TestCSV(t *testing.T) {
	var buf bytes.Buffer
	w := NewCSV(&buf)
	if err := w.Write(checker.Result{Domain: "a.us", TLD: ".us", Available: true}); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.HasPrefix(s, "domain,tld,available,error") {
		t.Errorf("expected header, got %q", s)
	}
	if !strings.Contains(s, "a.us,.us,true") {
		t.Errorf("unexpected csv body: %q", s)
	}
}
