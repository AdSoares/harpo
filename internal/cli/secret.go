package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/config"
	"github.com/harpo-sh/harpo/internal/policy"
	"github.com/harpo-sh/harpo/internal/provider"
	"github.com/harpo-sh/harpo/internal/ui"
)

func newSecretCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "secret", Short: "Map and inspect secret aliases (never values)"}
	cmd.AddCommand(newSecretMapCmd(), newSecretListCmd(), newSecretTestCmd())
	return cmd
}

func newSecretMapCmd() *cobra.Command {
	var providerName, ref, field, env string
	var tags []string
	cmd := &cobra.Command{
		Use:   "map <alias>",
		Short: "Map a local alias to a vault item/field (no value stored)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			alias := args[0]
			proj, err := loadProject()
			if err != nil {
				return err
			}
			if _, ok := proj.cfg.Providers[providerName]; !ok {
				return fmt.Errorf("provider %q is not configured", providerName)
			}
			if field == "" {
				field = "password"
			}

			ui.Title("Map secret alias")
			ui.Info("Alias: %s", alias)
			ui.Info("Provider: %s", providerName)
			ui.Info("Vault ref: %s", ref)
			ui.Info("Field: %s", field)
			ui.Info("Default destination: env:%s", env)
			fmt.Println()
			ui.Dim("No secret value will be stored in %s.", config.FileName)
			if w := policy.SuspiciousAliasWarning(alias); w != "" {
				ui.Warn("%s", w)
			}
			if !ui.Confirm("Confirm this mapping?") {
				return fmt.Errorf("aborted")
			}

			if proj.cfg.Secrets == nil {
				proj.cfg.Secrets = map[string]config.Secret{}
			}
			proj.cfg.Secrets[alias] = config.Secret{
				Provider:   providerName,
				Ref:        ref,
				Field:      field,
				DefaultEnv: env,
				Tags:       tags,
			}
			if err := proj.cfg.Save(proj.cfgPath); err != nil {
				return err
			}
			ui.Success("Mapped %q", alias)
			return nil
		},
	}
	cmd.Flags().StringVar(&providerName, "provider", "", "provider name (required)")
	cmd.Flags().StringVar(&ref, "ref", "", "vault item reference / search string (required)")
	cmd.Flags().StringVar(&field, "field", "password", "field within the item")
	cmd.Flags().StringVar(&env, "env", "", "default destination env var (required)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "optional tags")
	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("env")
	return cmd
}

func newSecretListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List secret aliases (never values)",
		RunE: func(_ *cobra.Command, _ []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			if len(proj.cfg.Secrets) == 0 {
				ui.Dim("No secrets mapped yet.")
				return nil
			}
			aliases := make([]string, 0, len(proj.cfg.Secrets))
			for a := range proj.cfg.Secrets {
				aliases = append(aliases, a)
			}
			sort.Strings(aliases)
			for _, a := range aliases {
				s := proj.cfg.Secrets[a]
				ui.Info("%-28s  provider=%s  env=%s", a, s.Provider, s.DefaultEnv)
			}
			return nil
		},
	}
}

func newSecretTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test <alias>",
		Short: "Verify a secret resolves (shows metadata only, never the value)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			eng := policy.New(proj.cfg)
			sec, err := eng.CheckSecretAuthorized(args[0])
			if err != nil {
				return err
			}
			p, err := newProvider(sec.Provider, proj.cfg.Providers[sec.Provider].Type)
			if err != nil {
				return err
			}
			res, err := p.Test(provider.Ref{Ref: sec.Ref, Field: sec.Field})
			if err != nil {
				return err
			}
			ui.Success("Secret resolved successfully.")
			ui.Info("Length: %d chars", res.Length)
			ui.Info("Fingerprint: %s", res.Fingerprint)
			ui.Info("Value: %s", "[redacted]")
			return nil
		},
	}
}
