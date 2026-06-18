package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/mcp"
	"github.com/harpo-sh/harpo/internal/provider"
)

func newMCPCmd() *cobra.Command {
	var profileName string
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Serve Harpo's value-free tools to an agent over MCP (stdio)",
		Long: "Runs a local MCP server (stdio) exposing read-only tools — session status,\n" +
			"available secrets and an audit tail. No tool returns a raw secret value.\n" +
			"Started by the agent's MCP config; enable with policies.mcp.enabled.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			if !proj.cfg.Policies.MCP.Enabled {
				return fmt.Errorf("MCP is disabled; set `policies.mcp.enabled: true` in harpo.yml")
			}
			// The resolver reuses provider creation + managed unlock (via the
			// keychain cache); values it returns are used to inject a brokered
			// command's environment and are never sent to the agent.
			pset := newProviderSet(proj)
			resolve := func(alias string) (string, error) {
				sec, ok := proj.cfg.Secrets[alias]
				if !ok {
					return "", fmt.Errorf("secret %q is not mapped", alias)
				}
				p, err := pset.get(sec.Provider)
				if err != nil {
					return "", err
				}
				return p.Resolve(provider.Ref{Ref: sec.Ref, Field: sec.Field})
			}
			srv := mcp.New(proj.cfg, profileName, proj.root, proj.harpoDir, Version, resolve)
			return srv.Run(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile to expose")
	return cmd
}
