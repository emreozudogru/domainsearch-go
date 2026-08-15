package config

import (
	"reflect"
	"testing"
)

func TestParseTLDs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{".us", []string{".us"}},
		{"us,com,net", []string{".us", ".com", ".net"}},
		{" us , com ", []string{".us", ".com"}},
		{"", nil},
	}
	for _, c := range cases {
		got := parseTLDs(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseTLDs(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestValidateSetsDefaults(t *testing.T) {
	c := &Config{}
	c.Validate()
	if !reflect.DeepEqual(c.Tlds, []string{".us"}) {
		t.Errorf("expected default tld [.us], got %v", c.Tlds)
	}
	if c.Rate != 1 {
		t.Errorf("expected rate 1, got %d", c.Rate)
	}
	if c.Workers != 1 {
		t.Errorf("expected workers 1, got %d", c.Workers)
	}
}
