package redact

import (
	"strings"
	"testing"
)

func TestRedactsKnownValue(t *testing.T) {
	r := New("glpat-abcdef123456")
	out := r.Redact("token is glpat-abcdef123456 here")
	if strings.Contains(out, "glpat-abcdef123456") {
		t.Fatalf("known value not redacted: %q", out)
	}
	if !strings.Contains(out, Mask) {
		t.Fatalf("expected mask in output: %q", out)
	}
}

func TestRedactsTokenFormats(t *testing.T) {
	cases := []string{
		"ghp_0123456789abcdefghij",
		"AKIAIOSFODNN7EXAMPLE",
		"sk-0123456789abcdefghij",
	}
	r := New() // no literal values; rely on format patterns
	for _, c := range cases {
		out := r.Redact("leaked " + c)
		if strings.Contains(out, c) {
			t.Fatalf("token format %q not redacted: %q", c, out)
		}
	}
}

func TestShortValuesNotMasked(t *testing.T) {
	// Very short values are ignored to avoid over-masking ordinary text.
	r := New("ab")
	out := r.Redact("the cab is ab here")
	if strings.Contains(out, Mask) {
		t.Fatalf("short value should not trigger masking: %q", out)
	}
}
