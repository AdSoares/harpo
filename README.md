# Harpo

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

**A local secret broker for AI coding agents.**

Harpo lets developers grant Claude Code, Codex and other coding agents
**temporary, scoped, auditable** access to credentials from existing vaults —
without pasting tokens into prompts, committing `.env` files, or giving the
agent direct access to the vault.

> The agent never receives vault access. It receives only temporary, limited,
> auditable access to the specific credentials you explicitly authorized.

## Why

AI coding agents can run commands, edit files and call developer tools. But
real work often needs credentials. Pasting secrets into prompts or storing them
in plaintext `.env` files is risky — they leak into history, transcripts,
logs, screenshots and Git.

Harpo keeps your vault as the source of truth and hands agents only the secrets
you authorized, for the current project and session, then audits the use
without ever storing the value.

## Example

```bash
bw unlock
harpo run --profile my-project-dev -- claude
```

## Status

Early MVP, under active development. Built in **Go** (decision recorded in
[`docs/adr/ADR-0001-stack-mvp-go.md`](docs/adr/ADR-0001-stack-mvp-go.md)). The
MVP targets a single provider — **Bitwarden Password Manager** via the `bw`
CLI. Specs:

- [`docs/mvp-spec.md`](docs/mvp-spec.md) — what is being built now.
- [`docs/market-ready-spec.md`](docs/market-ready-spec.md) — long-term vision.

## Install (from source)

Requires Go 1.26+.

```bash
go build -o harpo .
./harpo version
```

## Quick start

```bash
harpo init --mode strict --agent claude
harpo provider add bw --type bitwarden-password-manager
harpo secret map gitlab.ad.read \
  --provider bw \
  --ref "gitlab.com | ad | PAT | claude-code | read_api" \
  --field password \
  --env GITLAB_TOKEN
harpo profile create dev --ttl 2h --agent claude
harpo profile add-secret dev gitlab.ad.read
harpo agent setup claude
harpo run --profile dev -- claude
```

## Command surface (MVP)

```
harpo init                      Initialize Harpo in the current project
harpo provider add|status       Manage / probe vault providers
harpo secret map|list|test      Map aliases to vault items (never the value)
harpo profile create|add-secret Manage reusable session profiles
harpo session start|status|list|revoke
harpo run --profile <p> -- ...  Run an agent/command with secrets injected
harpo exec --with a:ENV -- ...  Run one command with specific secrets
harpo env render                Write a temporary .env (balanced mode only)
harpo audit list                Inspect the local audit log
harpo agent setup claude|codex  Generate agent-safety config
```

## Security model (summary)

- Harpo does **not** replace your vault.
- Harpo does **not** expose your vault session (`BW_SESSION`) to agents.
- Harpo does **not** print secrets by default.
- Harpo uses temporary session grants with a TTL.
- Harpo writes audit logs **without** secret values.

See [`SECURITY.md`](SECURITY.md) for the threat model and what Harpo does *not*
protect against.

## License

[Apache-2.0](LICENSE).
