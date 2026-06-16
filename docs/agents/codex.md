# Codex CLI

Harpo runs Codex with authorized credentials injected, and recommends pairing
that with Codex's own sandbox and approval controls.

## Recommended usage

```bash
harpo run --profile dev -- codex --sandbox workspace-write --ask-for-approval on-request
```

Codex separates two distinct controls:

- **Sandbox** defines the technical limits (use `workspace-write` for local
  work).
- **Approval policy** defines when the agent must ask for confirmation (use
  `on-request`).

They are complementary; Harpo handles the credential side while these handle
what the agent may do.

## Setup

```bash
harpo agent setup codex
```

This creates or updates **`AGENTS.md`** with a managed "Secrets policy" block:
use Harpo for credentials (`harpo session status`, `harpo exec` with approved
commands), do not run `bw` / `op` / `vault` / `printenv` / `env` / `cat .env`
directly, and prefer `workspace-write` + `on-request`. The block is delimited by
markers, so re-running setup updates it in place. Setup also ensures the
required `.gitignore` entries.

## One-off commands

```bash
harpo exec --with gitlab.ad.read:GITLAB_TOKEN -- glab repo list
```

`harpo exec` redacts known secret values and token formats from the command's
output. See the [security model](../security-model.md) for what is and isn't
guaranteed.
