package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/config"
	"github.com/harpo-sh/harpo/internal/policy"
	"github.com/harpo-sh/harpo/internal/ui"
)

func newEnvCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "env", Short: "Render secrets to a temporary .env (compatibility mode)"}
	cmd.AddCommand(newEnvRenderCmd())
	return cmd
}

func newEnvRenderCmd() *cobra.Command {
	var profileName, out, ttl string
	cmd := &cobra.Command{
		Use:   "render --profile <name>",
		Short: "Write a temporary .env with plaintext secrets (requires confirmation)",
		RunE: func(_ *cobra.Command, _ []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			// .env is an explicit opt-out from secure-by-default; policy must allow it.
			if err := policy.New(proj.cfg).CheckDotenv(); err != nil {
				return err
			}
			if out == "" {
				out = filepath.Join(config.Dir, ".env.session")
			}
			// Refuse to write outside .harpo/ (MVP spec §15).
			if !strings.HasPrefix(filepath.ToSlash(out), config.Dir+"/") {
				return fmt.Errorf("refusing to write .env outside %s/ (got %q)", config.Dir, out)
			}

			ui.Warn("This will write plaintext secrets to disk.")
			ui.Info("Path: %s", out)
			ui.Info("This path is ignored by Git.")
			if ttl != "" {
				ui.Info("Expires in: %s", ttl)
			}
			if !ui.Confirm("Confirm writing plaintext secrets?") {
				return fmt.Errorf("aborted")
			}

			r, _, err := resolveProfile(proj, profileName)
			if err != nil {
				return err
			}
			if _, err := ensureGitignore(proj.root); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(filepath.Join(proj.root, out)), 0o700); err != nil {
				return err
			}

			keys := make([]string, 0, len(r.inject))
			for k := range r.inject {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			var b strings.Builder
			for _, k := range keys {
				fmt.Fprintf(&b, "%s=%s\n", k, r.inject[k])
			}
			if err := os.WriteFile(filepath.Join(proj.root, out), []byte(b.String()), 0o600); err != nil {
				return err
			}
			ui.Success("Wrote %d secret(s) to %s", len(r.inject), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile to use (required)")
	cmd.Flags().StringVar(&out, "out", "", "output path (must be inside .harpo/)")
	cmd.Flags().StringVar(&ttl, "ttl", "", "informational TTL for the rendered file")
	_ = cmd.MarkFlagRequired("profile")
	return cmd
}
