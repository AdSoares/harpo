# Claude Code

Harpo is designed to run Claude Code with credentials already in place, so you
never paste tokens into the chat.

## Recommended usage

Let Harpo launch the agent, so it controls the environment from the start:

```bash
harpo run --profile dev -- claude
```

Harpo resolves the profile's authorized secrets, injects them as environment
variables, strips `BW_SESSION` from the child process, and records the use in
the audit log.

## Setup

```bash
harpo agent setup claude
```

This creates or updates:

- **`CLAUDE.md`** - a managed "Secrets policy" block telling the agent to use
  `harpo run` / `harpo exec` and never to run `bw`, `op`, `vault`, `env`,
  `printenv`, `cat .env`, `harpo reveal`, etc. The block is delimited by markers
  so re-running setup updates it in place without touching the rest of the file.
- **`.claude/settings.local.json`** - deny rules that block those same commands
  and reads of `.env` / `.harpo/`. Only written if absent, so it won't clobber
  existing settings.
- **`.gitignore`** - ensures `.harpo/` and `.env*` are ignored.

## MCP tools (runtime, optional)

Add `--mcp --profile <name>` to also wire Harpo's MCP server:

```bash
harpo agent setup claude --mcp --profile dev
```

This additionally writes `.mcp.json` (registering the `harpo` server as
`harpo mcp --profile dev`), enables `policies.mcp.enabled`, and adds an "MCP
tools" section to the managed `CLAUDE.md` block. The agent can then call the
value-free tools `harpo_session_status`, `harpo_secret_available` and
`harpo_exec` at runtime - receiving brokered command output, never raw secret
values. `harpo_exec` runs only commands in `policies.proxy.exec_allowlist`
(empty by default = denied). The deny rules stay; MCP is the safe path, the deny
rules block the unsafe one. See [proxy / MCP mode](../specs/proxy-mcp-mode.md).

## Why deny rules, not just instructions

Instructions in `CLAUDE.md` are advisory: an agent can be influenced by prompt
injection. The deny rules in `.claude/settings.local.json` are the real
guardrail, because Claude Code's permission system enforces them regardless of
what the prompt says. Harpo writes both, and treats the deny rules as the
primary control. See the [security model](../security-model.md).

## One-off commands

For a single command rather than an interactive session:

```bash
harpo exec --with gitlab.ad.read:GITLAB_TOKEN -- glab repo list
```

`harpo exec` redacts known secret values and token formats from the command's
output; `harpo run` (interactive) does not promise redaction.
