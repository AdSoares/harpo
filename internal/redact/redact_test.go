package redact

import (
	"bytes"
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

func TestWriterRedactsCompleteLines(t *testing.T) {
	var dst bytes.Buffer
	w := New("glpat-abcdef123456").NewWriter(&dst)

	// Write the secret split across two Writes, completed by a newline.
	if _, err := w.Write([]byte("token glpat-")); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("abcdef123456 done\n")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	out := dst.String()
	if strings.Contains(out, "glpat-abcdef123456") {
		t.Fatalf("secret leaked through writer: %q", out)
	}
	if !strings.Contains(out, Mask) {
		t.Fatalf("expected mask in writer output: %q", out)
	}
}

func TestWriterFlushesTrailingPartialLine(t *testing.T) {
	var dst bytes.Buffer
	w := New("secretvalue").NewWriter(&dst)
	// No trailing newline; Close must flush and redact.
	if _, err := w.Write([]byte("leaked secretvalue")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(dst.String(), "secretvalue") {
		t.Fatalf("trailing secret not redacted on close: %q", dst.String())
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
