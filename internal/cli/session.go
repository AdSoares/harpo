package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/config"
	"github.com/harpo-sh/harpo/internal/policy"
	"github.com/harpo-sh/harpo/internal/session"
	"github.com/harpo-sh/harpo/internal/ui"
)

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "session", Short: "Manage temporary session grants"}
	cmd.AddCommand(newSessionStartCmd(), newSessionStatusCmd(), newSessionListCmd(), newSessionRevokeCmd())
	return cmd
}

func newSessionStartCmd() *cobra.Command {
	var profileName, ttl string
	cmd := &cobra.Command{
		Use:   "start --profile <name>",
		Short: "Create an explicit session grant (metadata only)",
		RunE: func(_ *cobra.Command, _ []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			eng := policy.New(proj.cfg)
			prof, err := eng.ResolveProfile(profileName)
			if err != nil {
				return err
			}
			requested := prof.TTL.Duration()
			if ttl != "" {
				d, err := config.ParseTTL(ttl)
				if err != nil {
					return err
				}
				requested = d.Duration()
			}
			effective, err := eng.CheckTTL(requested)
			if err != nil {
				return err
			}
			var grants []session.Grant
			for _, ps := range prof.Secrets {
				env := ps.Env
				if env == "" {
					env = proj.cfg.Secrets[ps.Secret].DefaultEnv
				}
				grants = append(grants, session.Grant{Alias: ps.Secret, Destination: "env:" + env})
			}
			sm := session.NewManager(proj.harpoDir)
			sess, err := sm.Create(profileName, prof.Agent, proj.root, effective, grants)
			if err != nil {
				return err
			}
			ui.Success("Started session %s (expires in %s)", sess.ID, effective)
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile to use (required)")
	cmd.Flags().StringVar(&ttl, "ttl", "", "override the profile TTL")
	_ = cmd.MarkFlagRequired("profile")
	return cmd
}

func printSession(s *session.Session) {
	ui.Title("Session: " + s.ID)
	ui.Info("Profile: %s", s.Profile)
	ui.Info("Agent: %s", s.Agent)
	ui.Info("Project: %s", s.ProjectPath)
	ui.Info("Expires in: %s", s.Remaining().Round(time.Second))
	ui.Info("Secrets: %d", len(s.Grants))
	for _, g := range s.Grants {
		ui.Info("- %s -> %s", g.Alias, g.Destination)
	}
}

func newSessionStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current active session for this project",
		RunE: func(_ *cobra.Command, _ []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			sm := session.NewManager(proj.harpoDir)
			s, err := sm.Current(proj.root)
			if err != nil {
				ui.Dim("No active session.")
				return nil
			}
			printSession(s)
			return nil
		},
	}
}

func newSessionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List sessions",
		RunE: func(_ *cobra.Command, _ []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			sm := session.NewManager(proj.harpoDir)
			sessions, err := sm.List()
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				ui.Dim("No sessions.")
				return nil
			}
			for _, s := range sessions {
				state := "active"
				if s.Expired() {
					state = "expired"
				}
				ui.Info("%s  %-16s  %-8s  %s", s.ID, s.Profile, state, s.Remaining().Round(time.Second))
			}
			return nil
		},
	}
}

func newSessionRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke [id|current]",
		Short: "Revoke a session and delete its metadata (and any .env it rendered)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			sm := session.NewManager(proj.harpoDir)
			target := "current"
			if len(args) == 1 {
				target = args[0]
			}
			id := target
			if target == "current" {
				s, err := sm.Current(proj.root)
				if err != nil {
					return err
				}
				id = s.ID
			}
			if err := sm.Revoke(id); err != nil {
				return fmt.Errorf("revoke %s: %w", id, err)
			}
			ui.Success("Revoked session %s", id)
			return nil
		},
	}
}
