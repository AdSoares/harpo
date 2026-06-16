package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration wraps time.Duration so harpo.yml can use human strings like "2h"
// or "30m". Go's time.ParseDuration already understands these forms.
type Duration time.Duration

// Duration returns the underlying time.Duration.
func (d Duration) Duration() time.Duration { return time.Duration(d) }

// String renders the duration in Go's compact form (e.g. "2h0m0s").
func (d Duration) String() string { return time.Duration(d).String() }

// MarshalYAML serializes the duration as a string.
func (d Duration) MarshalYAML() (any, error) { return time.Duration(d).String(), nil }

// UnmarshalYAML accepts a string such as "2h" and parses it into a duration.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("ttl must be a string like \"2h\": %w", err)
	}
	if s == "" {
		*d = 0
		return nil
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid ttl %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

// ParseTTL parses a TTL string into a Duration.
func ParseTTL(s string) (Duration, error) {
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid ttl %q: %w", s, err)
	}
	return Duration(parsed), nil
}
