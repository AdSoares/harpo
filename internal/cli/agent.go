package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/ui"
)

// harpoBlockMarker delimits the Harpo-managed section in agent instruction
// files so `agent setup` is idempotent.
const harpoBlockMarker = "<!-- harpo:secrets-policy -->"

const claudePolicy = harpoBlockMarker + `
# Secrets policy (managed by Harpo)

Use Harpo for secrets. Never paste tokens into the prompt, never print secrets,
never write secrets to versioned files.

Allowed:
- harpo run
- harpo exec
- harpo session status

Forbidden:
- bw, op, vault, keeper, ksm
- env, printenv, set, Get-ChildItem Env:
- cat .env, type .env
- git add .env, git add .harpo
- harpo reveal
` + harpoBlockMarker

const codexPolicy = harpoBlockMarker + `
# Secrets policy (managed by Harpo)

Use Harpo for credentials.

When you need credentials:
- harpo session status
- harpo exec with approved commands

Do not run directly:
- bw, op, vault, keeper, ksm
- printenv, env
- cat .env, type .env

Recommended mode for Codex:
- sandbox: workspace-write
- approval: on-request
` + harpoBlockMarker

// claudeDenySettings is a minimal .claude/settings.local.json that denies the
// commands an agent must never run. Control must not rely on instructions
// alone (MVP spec §5.3), so these deny rules are the real guardrail.
const claudeDenySettings = `{
  "permissions": {
    "deny": [
      "Bash(bw:*)",
      "Bash(op:*)",
      "Bash(vault:*)",
      "Bash(keeper:*)",
      "Bash(ksm:*)",
      "Bash(env:*)",
      "Bash(printenv:*)",
      "Read(./.env)",
      "Read(./.harpo/**)"
    ]
  }
}
`

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "agent", Short: "Generate agent-safety configuration"}
	cmd.AddCommand(newAgentSetupCmd())
	return cmd
}

func newAgentSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "setup <claude|codex>",
		Short:     "Write agent instructions and deny rules for secrets",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"claude", "codex"},
		RunE: func(_ *cobra.Command, args []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			switch args[0] {
			case "claude":
				if err := writeManagedBlock(filepath.Join(proj.root, "CLAUDE.md"), claudePolicy); err != nil {
					return err
				}
				ui.Success("Updated CLAUDE.md secrets policy")
				if err := writeIfAbsent(filepath.Join(proj.root, ".claude", "settings.local.json"), claudeDenySettings); err != nil {
					return err
				}
			case "codex":
				if err := writeManagedBlock(filepath.Join(proj.root, "AGENTS.md"), codexPolicy); err != nil {
					return err
				}
				ui.Success("Updated AGENTS.md secrets policy")
			default:
				return fmt.Errorf("unknown agent %q: use claude or codex", args[0])
			}
			if _, err := ensureGitignore(proj.root); err != nil {
				return err
			}
			return nil
		},
	}
}

// writeManagedBlock inserts or replaces the Harpo-managed block in a markdown
// file, leaving the rest of the file untouched.
func writeManagedBlock(path, block string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(existing)
	if start := strings.Index(content, harpoBlockMarker); start >= 0 {
		if end := strings.LastIndex(content, harpoBlockMarker); end > start {
			content = content[:start] + block + content[end+len(harpoBlockMarker):]
			return os.WriteFile(path, []byte(content), 0o644)
		}
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if content != "" {
		content += "\n"
	}
	content += block + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

// writeIfAbsent writes content only if the file does not already exist, to
// avoid clobbering an agent's existing settings. Warns when it skips.
func writeIfAbsent(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		ui.Warn("%s already exists — review deny rules manually", path)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	ui.Success("Wrote %s", path)
	return nil
}
