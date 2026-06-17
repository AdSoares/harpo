package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/x/term"

	"github.com/harpo-sh/harpo/internal/audit"
	"github.com/harpo-sh/harpo/internal/config"
	"github.com/harpo-sh/harpo/internal/keychain"
	"github.com/harpo-sh/harpo/internal/provider"
	"github.com/harpo-sh/harpo/internal/ui"
)

// cacheEnabled reports whether unlocked sessions should be cached in the OS
// keychain.
func cacheEnabled(cfg *config.Config) bool {
	return cfg.Policies.UnlockCache == "keychain"
}

// unlockCacheTTL returns the effective session-cache TTL: the configured value
// (or 15m default), capped by max_ttl.
func unlockCacheTTL(cfg *config.Config) time.Duration {
	ttl := cfg.Policies.UnlockCacheTTL.Duration()
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	if max := cfg.Policies.MaxTTL.Duration(); max > 0 && ttl > max {
		ttl = max
	}
	return ttl
}

// isInteractive reports whether stdin is a real terminal, so Harpo only prompts
// for a master password when a human can answer. A proper terminal check is
// required: a char-device test would wrongly treat /dev/null as interactive.
func isInteractive() bool {
	return term.IsTerminal(os.Stdin.Fd())
}

// ensureUnlocked performs an in-process unlock when managed unlock is enabled,
// the provider supports it, the vault is locked, and a terminal is available.
// The session lives only in this Harpo process (no persistence yet) and is
// never exported to the shell or to a child process.
func ensureUnlocked(proj *project, p provider.Provider) error {
	if !proj.cfg.Policies.ManageUnlock {
		return nil
	}
	u, ok := p.(provider.Unlocker)
	if !ok {
		return nil
	}
	log := audit.NewLogger(proj.harpoDir)

	// 1. Reuse a cached session from the OS keychain, if enabled and valid.
	if cacheEnabled(proj.cfg) {
		if e, ok, _ := keychain.Load(p.ID()); ok {
			u.SetSession(provider.Session{Name: e.Name, Value: e.Value})
			_ = log.Log(audit.Event{Event: "vault.unlock.cache_hit", Project: proj.cfg.Project.Name, Provider: p.ID(), Cache: "keychain", Result: "success"})
			return nil
		}
	}

	// 2. Already unlocked via an ambient session? Nothing to do.
	st, _ := p.Status()
	if st.Vault != provider.StateLocked {
		return nil
	}

	// 3. Prompt and unlock in-process.
	if !isInteractive() {
		return fmt.Errorf("vault for provider %q is locked and no terminal is available to prompt for the master password", p.ID())
	}
	master, err := ui.Password(fmt.Sprintf("Master password for provider %q:", p.ID()))
	if err != nil {
		return err
	}
	sess, err := u.Unlock(master)
	if err != nil {
		return err
	}

	cache := "none"
	if cacheEnabled(proj.cfg) {
		ttl := unlockCacheTTL(proj.cfg)
		if err := keychain.Save(p.ID(), keychain.Entry{Name: sess.Name, Value: sess.Value, ExpiresAt: time.Now().Add(ttl)}); err == nil {
			cache = "keychain"
		} else {
			ui.Warn("could not cache session in the OS keychain: %v", err)
		}
	}
	_ = log.Log(audit.Event{Event: "vault.unlocked", Project: proj.cfg.Project.Name, Provider: p.ID(), Cache: cache, Result: "success"})
	ui.Dim("Unlocked provider %q (session held by Harpo; not exported to the shell or the agent).", p.ID())
	return nil
}

// providerSet caches provider instances per config name within a single
// command, so a managed-unlock session obtained once is reused across every
// secret that shares the same provider.
type providerSet struct {
	proj  *project
	cache map[string]provider.Provider
}

func newProviderSet(proj *project) *providerSet {
	return &providerSet{proj: proj, cache: map[string]provider.Provider{}}
}

// get returns the provider for the given config name, creating it once and
// auto-unlocking it if managed unlock applies.
func (s *providerSet) get(name string) (provider.Provider, error) {
	if p, ok := s.cache[name]; ok {
		return p, nil
	}
	pc, ok := s.proj.cfg.Providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q is not configured", name)
	}
	p, err := newProvider(name, pc.Type)
	if err != nil {
		return nil, err
	}
	if err := ensureUnlocked(s.proj, p); err != nil {
		return nil, err
	}
	s.cache[name] = p
	return p, nil
}
