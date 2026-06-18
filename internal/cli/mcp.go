package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/mcp"
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
			srv := mcp.New(proj.cfg, profileName, proj.root, proj.harpoDir, Version)
			return srv.Run(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile to expose")
	return cmd
}
