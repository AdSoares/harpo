// Package redact masks secret values and well-known token formats in text that
// Harpo controls (output of `harpo exec`, Harpo's own error messages). It is
// best-effort: Harpo does not promise full redaction inside interactive TUIs
// such as `harpo run -- claude`. See MVP spec §10.8 and §15.
package redact

import (
	"regexp"
	"strings"
)

// Mask is the placeholder substituted for redacted content.
const Mask = "[redacted]"

// tokenPatterns matches common credential shapes by format. These complement
// exact-value redaction for cases where the value is not known in advance.
var tokenPatterns = []*regexp.Regexp{
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{16,}`),       // GitHub tokens
	regexp.MustCompile(`glpat-[A-Za-z0-9_\-]{16,}`),        // GitLab PAT
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),                 // AWS access key id
	regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),              // OpenAI-style keys
	regexp.MustCompile(`xox[baprs]-[A-Za-z0-9\-]{10,}`),    // Slack tokens
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
