package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/config"
	"github.com/harpo-sh/harpo/internal/ui"
)

func newInitCmd() *cobra.Command {
	var mode, agent string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Harpo in the current project",
		Long:  "Creates harpo.yml, the .harpo/ working directory, and the required .gitignore entries.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runInit(config.Mode(mode), agent)
		},
	}
	cmd.Flags().StringVar(&mode, "mode", string(config.ModeStrict), "security mode: strict|balanced")
	cmd.Flags().StringVar(&agent, "agent", "", "default agent for the starter profile: claude|codex")
	return cmd
}

func runInit(mode config.Mode, agent string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	cfgPath := filepath.Join(wd, config.FileName)
	if _, err := os.Stat(cfgPath); err == nil {
		return fmt.Errorf("%s already exists in this directory", config.FileName)
	}

	if mode != config.ModeStrict && mode != config.ModeBalanced {
		return fmt.Errorf("invalid mode %q: use strict or balanced", mode)
	}

	cfg := config.Default(filepath.Base(wd), mode, agent)
	if err := cfg.Save(cfgPath); err != nil {
		return err
	}

	harpoDir := filepath.Join(wd, config.Dir)
	if err := os.MkdirAll(filepath.Join(harpoDir, "sessions"), 0o700); err != nil {
		return err
	}

	added, err := ensureGitignore(wd)
	if err != nil {
		return err
	}

	ui.Success("Created %s (mode: %s)", config.FileName, mode)
	ui.Success("Created %s/", config.Dir)
	if len(added) > 0 {
		ui.Success("Updated .gitignore (%d entries)", len(added))
	}
	fmt.Println()
	ui.Dim("Next steps:")
	ui.Dim("  harpo provider add bitwarden-personal --type bitwarden-password-manager")
	ui.Dim("  harpo secret map <alias> --provider bitwarden-personal --ref \"...\" --field password --env ENV_NAME")
	ui.Dim("  harpo profile create dev --ttl 2h --agent claude")
	if agent != "" {
		ui.Dim("  harpo agent setup %s", agent)
	}
	ui.Dim("  harpo run --profile dev -- claude")
	return nil
}
