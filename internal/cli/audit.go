package cli

import (
	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/audit"
	"github.com/harpo-sh/harpo/internal/ui"
)

func newAuditCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "audit", Short: "Inspect the local audit log (never shows values)"}
	cmd.AddCommand(newAuditListCmd())
	return cmd
}

func newAuditListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List audit events",
		RunE: func(_ *cobra.Command, _ []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			events, err := audit.NewLogger(proj.harpoDir).List()
			if err != nil {
				return err
			}
			if len(events) == 0 {
				ui.Dim("No audit events yet.")
				return nil
			}
			for _, e := range events {
				ui.Info("%s  %-16s  %-20s  %s  %s",
					e.Time, e.Event, e.SecretAlias, e.Destination, e.Result)
			}
			return nil
		},
	}
}
