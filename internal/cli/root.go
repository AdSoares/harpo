// Package cli wires the Harpo command tree (cobra) to the core packages.
// Command surface follows MVP spec §12.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/config"
)

// Version is set via -ldflags at build time; "dev" during local development.
var Version = "dev"

// project bundles the resolved config and the relevant paths for a command.
type project struct {
	cfg      *config.Config
	cfgPath  string // absolute path to harpo.yml
	root     string // project root (dir containing harpo.yml)
	harpoDir string // <root>/.harpo
}

// loadProject locates harpo.yml from the current working directory upward and
// loads it. Most commands need this; `init` does not.
func loadProject() (*project, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	cfgPath, err := config.Find(wd)
	if err != nil {
		return nil, fmt.Errorf("%w\nRun `harpo init` to create one", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	root := filepath.Dir(cfgPath)
	return &project{
		cfg:      cfg,
		cfgPath:  cfgPath,
		root:     root,
		harpoDir: filepath.Join(root, config.Dir),
	}, nil
}

// NewRootCmd builds the root command and registers all subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "harpo",
		Short:         "A local secret broker for AI coding agents",
		Long:          "Harpo grants AI coding agents temporary, scoped, auditable access to\nsecrets from your existing vaults — without exposing the vault itself.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newInitCmd(),
		newProviderCmd(),
		newSecretCmd(),
		newProfileCmd(),
		newSessionCmd(),
		newRunCmd(),
		newExecCmd(),
		newUnlockCmd(),
		newLockCmd(),
		newEnvCmd(),
		newAuditCmd(),
		newAgentCmd(),
		newVersionCmd(),
	)
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Harpo version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("harpo %s\n", Version)
		},
	}
}

// Execute runs the root command and returns a process exit code.
func Execute() int {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
