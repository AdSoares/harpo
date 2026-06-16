// Package policy validates requests against the configured security mode and
// policies before any secret is resolved or any session is created. See MVP
// spec §10.3 and §8.
package policy

import (
	"fmt"
	"time"

	"github.com/harpo-sh/harpo/internal/config"
)

// Engine enforces policy for a loaded config.
type Engine struct {
	cfg *config.Config
}

// New returns a policy engine bound to the given config.
func New(cfg *config.Config) *Engine { return &Engine{cfg: cfg} }

// CheckTTL validates a requested TTL against policy. In strict mode a TTL is
// mandatory; in any mode it must not exceed max_ttl. Returns the effective TTL.
func (e *Engine) CheckTTL(requested time.Duration) (time.Duration, error) {
	if requested <= 0 {
		if e.cfg.Mode == config.ModeStrict {
			return 0, fmt.Errorf("a TTL is required in strict mode")
		}
		requested = e.cfg.Policies.DefaultTTL.Duration()
	}
	if max := e.cfg.Policies.MaxTTL.Duration(); max > 0 && requested > max {
		return 0, fmt.Errorf("requested TTL %s exceeds max_ttl %s", requested, max)
	}
	return requested, nil
}

// CheckReveal returns an error if revealing secrets is not allowed. Reveal is
// always denied in strict mode regardless of the policy flag. See MVP spec §8.1.
func (e *Engine) CheckReveal() error {
	if e.cfg.Mode == config.ModeStrict {
		return fmt.Errorf("reveal is disabled in strict mode")
	}
	if !e.cfg.Policies.AllowReveal {
		return fmt.Errorf("reveal is disabled by policy (allow_reveal: false)")
	}
	return nil
}

// CheckDotenv returns an error if rendering a .env file is not allowed.
func (e *Engine) CheckDotenv() error {
	if e.cfg.Mode == config.ModeStrict {
		return fmt.Errorf(".env rendering is disabled in strict mode")
	}
	if !e.cfg.Policies.AllowDotenv {
		return fmt.Errorf(".env rendering is disabled by policy (allow_dotenv: false)")
	}
	return nil
}

// CheckSecretAuthorized verifies the alias exists in config and resolves its
// provider. Returns the secret definition for resolution.
func (e *Engine) CheckSecretAuthorized(alias string) (config.Secret, error) {
	s, ok := e.cfg.Secrets[alias]
	if !ok {
		return config.Secret{}, fmt.Errorf("secret alias %q is not defined in %s", alias, config.FileName)
	}
	if _, ok := e.cfg.Providers[s.Provider]; !ok {
		return config.Secret{}, fmt.Errorf("secret %q references unknown provider %q", alias, s.Provider)
	}
	return s, nil
}

// ResolveProfile returns the named profile or an error.
func (e *Engine) ResolveProfile(name string) (config.Profile, error) {
	p, ok := e.cfg.Profiles[name]
	if !ok {
		return config.Profile{}, fmt.Errorf("profile %q is not defined in %s", name, config.FileName)
	}
	return p, nil
}

// SuspiciousAliasWarning returns a non-empty warning when an alias looks like a
// production credential. This is advisory, not blocking. See MVP spec §15.
func SuspiciousAliasWarning(alias string) string {
	for _, kw := range []string{"prod", "production", "root", "admin"} {
		if containsWord(alias, kw) {
			return fmt.Sprintf("alias %q looks production-like (%q) — double-check before granting to an agent", alias, kw)
		}
	}
	return ""
}

func containsWord(s, sub string) bool {
	// simple case-insensitive substring check; aliases are lowercase by convention
	for i := 0; i+len(sub) <= len(s); i++ {
		if equalFold(s[i:i+len(sub)], sub) {
			return true
		}
	}
	return false
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
