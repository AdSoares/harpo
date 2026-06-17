package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/audit"
	"github.com/harpo-sh/harpo/internal/policy"
	"github.com/harpo-sh/harpo/internal/provider"
	"github.com/harpo-sh/harpo/internal/redact"
	"github.com/harpo-sh/harpo/internal/runner"
	"github.com/harpo-sh/harpo/internal/session"
	"github.com/harpo-sh/harpo/internal/ui"
)

// resolved holds the outcome of resolving a profile's secrets: the env to
// inject and the grant metadata to record in the session/audit.
type resolved struct {
	inject map[string]string
	grants []session.Grant
}

// resolveProfile resolves every secret in a profile to its destination env var.
// It returns the injection map (values are sensitive) and grant metadata.
func resolveProfile(proj *project, profileName string) (resolved, string, error) {
	eng := policy.New(proj.cfg)
	prof, err := eng.ResolveProfile(profileName)
	if err != nil {
		return resolved{}, "", err
	}
	pset := newProviderSet(proj)
	out := resolved{inject: map[string]string{}}
	for _, ps := range prof.Secrets {
		sec, err := eng.CheckSecretAuthorized(ps.Secret)
		if err != nil {
			return resolved{}, "", err
		}
		p, err := pset.get(sec.Provider)
		if err != nil {
			return resolved{}, "", err
		}
		value, err := p.Resolve(provider.Ref{Ref: sec.Ref, Field: sec.Field})
		if err != nil {
			return resolved{}, "", fmt.Errorf("resolving %q: %w", ps.Secret, err)
		}
		env := ps.Env
		if env == "" {
			env = sec.DefaultEnv
		}
		out.inject[env] = value
		out.grants = append(out.grants, session.Grant{Alias: ps.Secret, Destination: "env:" + env})
	}
	return out, prof.Agent, nil
}

func newRunCmd() *cobra.Command {
	var profileName string
	cmd := &cobra.Command{
		Use:                "run -- <command> [args...]",
		Short:              "Run a command/agent with authorized secrets injected (primary mode)",
		DisableFlagParsing: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command given; usage: harpo run --profile <name> -- <command>")
			}
			proj, err := loadProject()
			if err != nil {
				return err
			}
			r, agent, err := resolveProfile(proj, profileName)
			if err != nil {
				return err
			}

			// Enforce TTL policy and record a session (metadata only).
			eng := policy.New(proj.cfg)
			prof, _ := eng.ResolveProfile(profileName)
			ttl, err := eng.CheckTTL(prof.TTL.Duration())
			if err != nil {
				return err
			}
			sm := session.NewManager(proj.harpoDir)
			sess, err := sm.Create(profileName, agent, proj.root, ttl, r.grants)
			if err != nil {
				return err
			}

			log := audit.NewLogger(proj.harpoDir)
			for _, g := range r.grants {
				_ = log.Log(audit.Event{
					Event:       "secret.injected",
					Profile:     profileName,
					Agent:       agent,
					Project:     proj.cfg.Project.Name,
					SecretAlias: g.Alias,
					Destination: g.Destination,
					Mode:        string(proj.cfg.Mode),
					TTLSeconds:  int(ttl.Seconds()),
					Result:      "success",
				})
			}

			ui.Dim("Session %s — %d secret(s), expires in %s. BW_SESSION is not passed to the child.",
				sess.ID, len(r.grants), ttl)

			// runner.Run never prints secret values; it only wires stdio.
			return runner.Run(args[0], args[1:], r.inject)
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile to use (required)")
	_ = cmd.MarkFlagRequired("profile")
	return cmd
}

func newExecCmd() *cobra.Command {
	var with []string
	cmd := &cobra.Command{
		Use:   "exec --with <alias:ENV> -- <command> [args...]",
		Short: "Run a single command with specific secrets injected",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command given; usage: harpo exec --with alias:ENV -- <command>")
			}
			proj, err := loadProject()
			if err != nil {
				return err
			}
			eng := policy.New(proj.cfg)
			pset := newProviderSet(proj)
			inject := map[string]string{}
			var grants []session.Grant
			for _, spec := range with {
				alias, env, ok := strings.Cut(spec, ":")
				if !ok || alias == "" || env == "" {
					return fmt.Errorf("invalid --with %q; expected alias:ENV", spec)
				}
				sec, err := eng.CheckSecretAuthorized(alias)
				if err != nil {
					return err
				}
				p, err := pset.get(sec.Provider)
				if err != nil {
					return err
				}
				value, err := p.Resolve(provider.Ref{Ref: sec.Ref, Field: sec.Field})
				if err != nil {
					return fmt.Errorf("resolving %q: %w", alias, err)
				}
				inject[env] = value
				grants = append(grants, session.Grant{Alias: alias, Destination: "env:" + env})
			}

			log := audit.NewLogger(proj.harpoDir)
			for _, g := range grants {
				_ = log.Log(audit.Event{
					Event:       "secret.injected",
					Project:     proj.cfg.Project.Name,
					SecretAlias: g.Alias,
					Destination: g.Destination,
					Mode:        string(proj.cfg.Mode),
					Result:      "success",
				})
			}
			// Route output through a redacting writer so resolved secret values
			// (and known token formats) are masked in this command's
			// stdout/stderr. `exec` targets non-interactive commands, where
			// line-buffered redaction is appropriate (MVP spec §10.8).
			values := make([]string, 0, len(inject))
			for _, v := range inject {
				values = append(values, v)
			}
			red := redact.New(values...)
			outW := red.NewWriter(os.Stdout)
			errW := red.NewWriter(os.Stderr)
			runErr := runner.RunWith(args[0], args[1:], inject, os.Stdin, outW, errW)
			_ = outW.Close()
			_ = errW.Close()
			return runErr
		},
	}
	cmd.Flags().StringArrayVar(&with, "with", nil, "alias:ENV pair to inject (repeatable)")
	_ = cmd.MarkFlagRequired("with")
	return cmd
}
