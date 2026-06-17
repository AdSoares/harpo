package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/x/term"

	"github.com/harpo-sh/harpo/internal/audit"
	"github.com/harpo-sh/harpo/internal/provider"
	"github.com/harpo-sh/harpo/internal/ui"
)

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
	st, _ := p.Status()
	if st.Vault != provider.StateLocked {
		return nil
	}
	if !isInteractive() {
		return fmt.Errorf("vault for provider %q is locked and no terminal is available to prompt for the master password", p.ID())
	}
	master, err := ui.Password(fmt.Sprintf("Master password for provider %q:", p.ID()))
	if err != nil {
		return err
	}
	if _, err := u.Unlock(master); err != nil {
		return err
	}
	_ = audit.NewLogger(proj.harpoDir).Log(audit.Event{
		Event:   "vault.unlocked",
		Project: proj.cfg.Project.Name,
		Result:  "success",
	})
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
