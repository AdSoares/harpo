package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/config"
	"github.com/harpo-sh/harpo/internal/ui"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "profile", Short: "Manage session profiles"}
	cmd.AddCommand(newProfileCreateCmd(), newProfileAddSecretCmd(), newProfileListCmd())
	return cmd
}

func newProfileCreateCmd() *cobra.Command {
	var ttl, agent string
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a session profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			proj, err := loadProject()
			if err != nil {
				return err
			}
			d, err := config.ParseTTL(ttl)
			if err != nil {
				return err
			}
			if proj.cfg.Profiles == nil {
				proj.cfg.Profiles = map[string]config.Profile{}
			}
			if _, exists := proj.cfg.Profiles[name]; exists {
				return fmt.Errorf("profile %q already exists", name)
			}
			proj.cfg.Profiles[name] = config.Profile{TTL: d, Agent: agent}
			if err := proj.cfg.Save(proj.cfgPath); err != nil {
				return err
			}
			ui.Success("Created profile %q (ttl: %s, agent: %s)", name, d, agent)
			return nil
		},
	}
	cmd.Flags().StringVar(&ttl, "ttl", "2h", "session TTL (e.g. 2h, 30m)")
	cmd.Flags().StringVar(&agent, "agent", "", "agent for this profile: claude|codex")
	return cmd
}

func newProfileAddSecretCmd() *cobra.Command {
	var env string
	cmd := &cobra.Command{
		Use:   "add-secret <profile> <alias>",
		Short: "Add a mapped secret to a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			profileName, alias := args[0], args[1]
			proj, err := loadProject()
			if err != nil {
				return err
			}
			prof, ok := proj.cfg.Profiles[profileName]
			if !ok {
				return fmt.Errorf("profile %q does not exist", profileName)
			}
			sec, ok := proj.cfg.Secrets[alias]
			if !ok {
				return fmt.Errorf("secret alias %q is not mapped", alias)
			}
			if env == "" {
				env = sec.DefaultEnv
			}
			for _, ps := range prof.Secrets {
				if ps.Secret == alias {
					return fmt.Errorf("profile %q already includes %q", profileName, alias)
				}
			}
			prof.Secrets = append(prof.Secrets, config.ProfileSecret{Secret: alias, Env: env})
			proj.cfg.Profiles[profileName] = prof
			if err := proj.cfg.Save(proj.cfgPath); err != nil {
				return err
			}
			ui.Success("Added %q -> env:%s to profile %q", alias, env, profileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&env, "env", "", "destination env var (defaults to the secret's default_env)")
	return cmd
}

func newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		RunE: func(_ *cobra.Command, _ []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			if len(proj.cfg.Profiles) == 0 {
				ui.Dim("No profiles defined yet.")
				return nil
			}
			names := make([]string, 0, len(proj.cfg.Profiles))
			for n := range proj.cfg.Profiles {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, n := range names {
				p := proj.cfg.Profiles[n]
				ui.Info("%-20s  ttl=%s  agent=%s  secrets=%d", n, p.TTL, p.Agent, len(p.Secrets))
			}
			return nil
		},
	}
}
