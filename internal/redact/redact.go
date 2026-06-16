// Package redact masks secret values and well-known token formats in text that
// Harpo controls (output of `harpo exec`, Harpo's own error messages). It is
// best-effort: Harpo does not promise full redaction inside interactive TUIs
// such as `harpo run -- claude`. See MVP spec §10.8 and §15.
package redact

import (
	"bytes"
	"io"
	"regexp"
	"strings"
)

// Mask is the placeholder substituted for redacted content.
const Mask = "[redacted]"

// tokenPatterns matches common credential shapes by format. These complement
// exact-value redaction for cases where the value is not known in advance.
var tokenPatterns = []*regexp.Regexp{
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{16,}`),    // GitHub tokens
	regexp.MustCompile(`glpat-[A-Za-z0-9_\-]{16,}`),     // GitLab PAT
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),              // AWS access key id
	regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),           // OpenAI-style keys
	regexp.MustCompile(`xox[baprs]-[A-Za-z0-9\-]{10,}`), // Slack tokens
}

// Redactor masks a known set of literal secret values plus generic token
// formats. Construct it with the values resolved for the current session so
// their exact strings are scrubbed from controlled output.
type Redactor struct {
	values []string
}

// New returns a Redactor that masks the given literal secret values (in
// addition to format-based patterns). Empty and very short values are ignored
// to avoid over-masking.
func New(values ...string) *Redactor {
	kept := make([]string, 0, len(values))
	for _, v := range values {
		if len(v) >= 4 {
			kept = append(kept, v)
		}
	}
	return &Redactor{values: kept}
}

// Redact returns s with known secret values and recognized token formats
// replaced by Mask.
func (r *Redactor) Redact(s string) string {
	for _, v := range r.values {
		if v != "" {
			s = strings.ReplaceAll(s, v, Mask)
		}
	}
	for _, re := range tokenPatterns {
		s = re.ReplaceAllString(s, Mask)
	}
	return s
}

// Writer wraps an io.Writer and redacts data as it streams through. It buffers
// partial lines and redacts each complete line before forwarding, so secrets
// are never written to the underlying writer unmasked. Call Close to flush any
// trailing partial line. This is the mechanism behind `harpo exec` redaction
// (MVP spec §10.8); it is line-oriented and does not catch a secret split
// across newlines.
type Writer struct {
	r   *Redactor
	dst io.Writer
	buf []byte
}

// NewWriter returns a redacting Writer that forwards to dst.
func (r *Redactor) NewWriter(dst io.Writer) *Writer {
	return &Writer{r: r, dst: dst}
}

// Write buffers p and flushes every complete line through the redactor. It
// always reports len(p) consumed, per the io.Writer contract.
func (w *Writer) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		line := string(w.buf[:i+1])
		if _, err := io.WriteString(w.dst, w.r.Redact(line)); err != nil {
			return 0, err
		}
		w.buf = w.buf[i+1:]
	}
	return len(p), nil
}

// Close flushes any remaining buffered (newline-less) content, redacted.
func (w *Writer) Close() error {
	if len(w.buf) == 0 {
		return nil
	}
	_, err := io.WriteString(w.dst, w.r.Redact(string(w.buf)))
	w.buf = nil
	return err
}
