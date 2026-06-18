package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harpo-sh/harpo/internal/config"
	"github.com/harpo-sh/harpo/internal/ui"
)

// harpoBlockMarker delimits the Harpo-managed section in agent instruction
// files so `agent setup` is idempotent.
const harpoBlockMarker = "<!-- harpo:secrets-policy -->"

// claudePolicyBody / codexPolicyBody are the managed-block contents (without
// the delimiting markers, which block() adds).
const claudePolicyBody = `# Secrets policy (managed by Harpo)

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
- harpo reveal`

const codexPolicyBody = `# Secrets policy (managed by Harpo)

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
- approval: on-request`

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

// mcpAddendumClaude is appended to the managed block when MCP is wired.
func mcpAddendumClaude() string {
	return `## MCP tools (Harpo)

Harpo also exposes MCP tools — prefer them for credentials:
- harpo_session_status — the active session (no values)
- harpo_secret_available — the credentials you may use (aliases only, never values)
- harpo_exec — run an allowlisted command with a credential injected by Harpo;
  you receive the output, never the secret value

Never expect or request a raw secret value.`
}

// mcpAddendumCodex includes the snippet to add to the global Codex config,
// since Codex has no project-local MCP config file.
func mcpAddendumCodex(profile string) string {
	return `## MCP tools (Harpo)

Harpo exposes MCP tools — prefer them for credentials: harpo_session_status,
harpo_secret_available, harpo_exec. You receive command output, never raw
secret values.

To enable, add to ~/.codex/config.toml:

    [mcp_servers.harpo]
    command = "harpo"
    args = ["mcp", "--profile", "` + profile + `"]`
}

// block wraps body in the managed-section markers.
func block(body string) string {
	return harpoBlockMarker + "\n" + body + "\n" + harpoBlockMarker
}

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "agent", Short: "Generate agent-safety configuration"}
	cmd.AddCommand(newAgentSetupCmd())
	return cmd
}

func newAgentSetupCmd() *cobra.Command {
	var withMCP bool
	var profileName string
	cmd := &cobra.Command{
		Use:       "setup <claude|codex>",
		Short:     "Write agent instructions and deny rules for secrets",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"claude", "codex"},
		RunE: func(_ *cobra.Command, args []string) error {
			proj, err := loadProject()
			if err != nil {
				return err
			}
			if withMCP && profileName == "" {
				return fmt.Errorf("--mcp requires --profile <name> (the profile the MCP server exposes)")
			}

			switch args[0] {
			case "claude":
				body := claudePolicyBody
				if withMCP {
					body += "\n\n" + mcpAddendumClaude()
				}
				if err := writeManagedBlock(filepath.Join(proj.root, "CLAUDE.md"), block(body)); err != nil {
					return err
				}
				ui.Success("Updated CLAUDE.md secrets policy")
				if err := writeIfAbsent(filepath.Join(proj.root, ".claude", "settings.local.json"), claudeDenySettings); err != nil {
					return err
				}
				if withMCP {
					created, err := upsertMCPServer(filepath.Join(proj.root, ".mcp.json"), "harpo", mcpServerEntry{
						Command: "harpo",
						Args:    []string{"mcp", "--profile", profileName},
					})
					if err != nil {
						return err
					}
					if created {
						ui.Success("Created .mcp.json with the harpo MCP server")
					} else {
						ui.Success("Updated .mcp.json (harpo MCP server)")
					}
				}
			case "codex":
				body := codexPolicyBody
				if withMCP {
					body += "\n\n" + mcpAddendumCodex(profileName)
				}
				if err := writeManagedBlock(filepath.Join(proj.root, "AGENTS.md"), block(body)); err != nil {
					return err
				}
				ui.Success("Updated AGENTS.md secrets policy")
				if withMCP {
					ui.Dim("Codex MCP config is global — add the snippet from AGENTS.md to ~/.codex/config.toml.")
				}
			default:
				return fmt.Errorf("unknown agent %q: use claude or codex", args[0])
			}

			if withMCP && !proj.cfg.Policies.MCP.Enabled {
				proj.cfg.Policies.MCP.Enabled = true
				if err := proj.cfg.Save(proj.cfgPath); err != nil {
					return err
				}
				ui.Success("Enabled policies.mcp.enabled in %s", config.FileName)
				if len(proj.cfg.Policies.Proxy.ExecAllowlist) == 0 {
					ui.Dim("harpo_exec is denied until you set policies.proxy.exec_allowlist.")
				}
			}

			if _, err := ensureGitignore(proj.root); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&withMCP, "mcp", false, "also wire Harpo's MCP server and add MCP guidance")
	cmd.Flags().StringVar(&profileName, "profile", "", "profile the MCP server exposes (required with --mcp)")
	return cmd
}

// mcpServerEntry is one entry under "mcpServers" in a Claude Code .mcp.json.
type mcpServerEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// upsertMCPServer adds or replaces the named server in a .mcp.json file,
// preserving any other servers and top-level keys. Returns whether the file was
// newly created.
func upsertMCPServer(path, name string, entry mcpServerEntry) (bool, error) {
	root := map[string]json.RawMessage{}
	created := false
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &root); err != nil {
			return false, fmt.Errorf("parse %s: %w", path, err)
		}
	} else if os.IsNotExist(err) {
		created = true
	} else {
		return false, err
	}

	servers := map[string]json.RawMessage{}
	if raw, ok := root["mcpServers"]; ok {
		_ = json.Unmarshal(raw, &servers)
	}
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return false, err
	}
	servers[name] = entryJSON
	serversJSON, err := json.Marshal(servers)
	if err != nil {
		return false, err
	}
	root["mcpServers"] = serversJSON

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, err
	}
	return created, os.WriteFile(path, append(out, '\n'), 0o644)
}

// writeManagedBlock inserts or replaces the Harpo-managed block in a markdown
// file, leaving the rest of the file untouched.
func writeManagedBlock(path, blk string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(existing)
	if start := strings.Index(content, harpoBlockMarker); start >= 0 {
		if end := strings.LastIndex(content, harpoBlockMarker); end > start {
			content = content[:start] + blk + content[end+len(harpoBlockMarker):]
			return os.WriteFile(path, []byte(content), 0o644)
		}
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if content != "" {
		content += "\n"
	}
	content += blk + "\n"
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
