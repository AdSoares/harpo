package cli

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/audit"
	"github.com/harpo-sh/harpo/internal/keychain"
	"github.com/harpo-sh/harpo/internal/provider"
	"github.com/harpo-sh/harpo/internal/ui"
)

// unlockTargets returns the provider names to act on: the named one, or all
// unlock-capable providers when none is given.
func unlockTargets(proj *project, args []string) ([]string, error) {
	if len(args) == 1 {
		if _, ok := proj.cfg.Providers[args[0]]; !ok {
			return nil, fmt.Errorf("provider %q is not configured", args[0])
		}
		return []string{args[0]}, nil
	}
	var names []string
	for name, pc := range proj.cfg.Providers {
		p, err := newProvider(name, pc.Type)
		if err != nil {
			continue
		}
		if _, ok := p.(provider.Unlocker); ok {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no unlock-capable providers configured")
	}
	sort.Strings(names)
	return names, nil
}

func newUnlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unlock [provider]",
		Short: "Unlock an unlock-capable vault and cache the session (OS keychain)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			if !cacheEnabled(proj.cfg) {
				return fmt.Errorf("standalone unlock needs a session cache; set `policies.unlock_cache: keychain` (otherwise use `harpo run`, which unlocks in-process)")
			}
			if !isInteractive() {
				return fmt.Errorf("no terminal is available to prompt for the master password")
			}
			targets, err := unlockTargets(proj, args)
			if err != nil {
				return err
			}
			ttl := unlockCacheTTL(proj.cfg)
			for _, name := range targets {
				p, err := newProvider(name, proj.cfg.Providers[name].Type)
				if err != nil {
					return err
				}
				u, ok := p.(provider.Unlocker)
				if !ok {
					ui.Warn("provider %q does not support unlock; skipping", name)
					continue
				}
				master, err := ui.Password(fmt.Sprintf("Master password for provider %q:", name))
				if err != nil {
					return err
				}
				sess, err := u.Unlock(master)
				if err != nil {
					return err
				}
				if err := keychain.Save(name, keychain.Entry{Name: sess.Name, Value: sess.Value, ExpiresAt: time.Now().Add(ttl)}); err != nil {
					return fmt.Errorf("caching session: %w", err)
				}
				_ = audit.NewLogger(proj.harpoDir).Log(audit.Event{Event: "vault.unlocked", Project: proj.cfg.Project.Name, Provider: name, Cache: "keychain", Result: "success"})
				ui.Success("Unlocked %q; session cached (expires in %s). Not exported to the shell or agents.", name, ttl)
			}
			return nil
		},
	}
}

func newLockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lock [provider]",
		Short: "Forget Harpo's cached session for a provider",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			targets, err := unlockTargets(proj, args)
			if err != nil {
				return err
			}
			for _, name := range targets {
				if err := keychain.Delete(name); err != nil {
					return fmt.Errorf("evicting %q: %w", name, err)
				}
				_ = audit.NewLogger(proj.harpoDir).Log(audit.Event{Event: "vault.locked", Project: proj.cfg.Project.Name, Provider: name, Result: "success"})
				ui.Success("Forgot cached session for %q.", name)
			}
			return nil
		},
	}
}
