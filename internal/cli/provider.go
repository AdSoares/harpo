package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/config"
	"github.com/harpo-sh/harpo/internal/provider"
	"github.com/harpo-sh/harpo/internal/provider/bitwarden"
	"github.com/harpo-sh/harpo/internal/provider/keeper"
	"github.com/harpo-sh/harpo/internal/provider/ksm"
	"github.com/harpo-sh/harpo/internal/provider/onepassword"
	"github.com/harpo-sh/harpo/internal/provider/vault"
	"github.com/harpo-sh/harpo/internal/ui"
)

func newProviderCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "provider", Short: "Manage vault providers"}
	cmd.AddCommand(newProviderAddCmd(), newProviderStatusCmd())
	return cmd
}

// newProvider constructs a provider adapter from a configured type.
func newProvider(id, providerType string) (provider.Provider, error) {
	switch providerType {
	case bitwarden.TypeName:
		return bitwarden.New(id), nil
	case keeper.TypeName:
		return keeper.New(id), nil
	case ksm.TypeName:
		return ksm.New(id), nil
	case onepassword.TypeName:
		return onepassword.New(id), nil
	case vault.TypeName:
		return vault.New(id), nil
	default:
		return nil, fmt.Errorf("unsupported provider type %q (supported: %q, %q, %q, %q, %q)", providerType, bitwarden.TypeName, keeper.TypeName, ksm.TypeName, onepassword.TypeName, vault.TypeName)
	}
}

func newProviderAddCmd() *cobra.Command {
	var providerType string
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a vault provider to harpo.yml",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			proj, err := loadProject()
			if err != nil {
				return err
			}
			if _, err := newProvider(name, providerType); err != nil {
				return err
			}
			// Validate connectivity before persisting.
			p, _ := newProvider(name, providerType)
			st, _ := p.Status()
			if !st.CLIFound {
				ui.Warn("%s", st.Detail)
			}

			if proj.cfg.Providers == nil {
				proj.cfg.Providers = map[string]config.Provider{}
			}
			proj.cfg.Providers[name] = config.Provider{Type: providerType}
			if err := proj.cfg.Save(proj.cfgPath); err != nil {
				return err
			}
			ui.Success("Added provider %q (type: %s)", name, providerType)
			return nil
		},
	}
	cmd.Flags().StringVar(&providerType, "type", bitwarden.TypeName, "provider type")
	return cmd
}

func newProviderStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [name]",
		Short: "Show provider and vault status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			if len(proj.cfg.Providers) == 0 {
				return fmt.Errorf("no providers configured; run `harpo provider add`")
			}
			for name, pc := range proj.cfg.Providers {
				if len(args) == 1 && args[0] != name {
					continue
				}
				p, err := newProvider(name, pc.Type)
				if err != nil {
					ui.Error("%s: %v", name, err)
					continue
				}
				st, _ := p.Status()
				ui.Title("Provider: " + name)
				ui.Info("Type: %s", st.Type)
				ui.Info("CLI: %s", boolToFound(st.CLIFound))
				ui.Info("Vault status: %s", st.Vault)
				ui.Info("Safe for agent inheritance: %s", yesNo(st.SafeForInheritance))
				if st.Detail != "" {
					ui.Dim("  %s", st.Detail)
				}
			}
			return nil
		},
	}
}

func boolToFound(b bool) string {
	if b {
		return "found"
	}
	return "not found"
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
