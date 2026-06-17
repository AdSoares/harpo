# Getting started

This guide walks through installing Harpo and setting up your first
agent-safe secret session. For the full design, see [`mvp-spec.md`](mvp-spec.md).

## Prerequisites

- **Go 1.26+** (to build from source; the toolchain version is pinned in `go.mod`).
- The **Bitwarden CLI** (`bw`) installed and logged in — the MVP's only provider.

## Install

From source:

```bash
go build -o harpo .
./harpo version
```

## First-time setup

```bash
# 1. Initialize Harpo in your project (secure-by-default: strict mode).
harpo init --mode strict --agent claude

# 2. Register your vault provider.
harpo provider add bw --type bitwarden-password-manager

# 3. Map a local alias to a vault item/field. No secret value is stored.
harpo secret map gitlab.ad.read \
  --provider bw \
  --ref "gitlab.com | ad | PAT | claude-code | read_api" \
  --field password \
  --env GITLAB_TOKEN

# 4. Create a reusable profile with a TTL, and add the secret to it.
harpo profile create dev --ttl 2h --agent claude
harpo profile add-secret dev gitlab.ad.read

# 5. Generate agent-safety configuration (deny rules + policy).
harpo agent setup claude

# 6. Run the agent with the authorized secret injected.
harpo run --profile dev -- claude
```

`harpo init` creates `harpo.yml`, the `.harpo/` working directory, and the
required `.gitignore` entries.

## Daily use

Once configured, your loop is just:

```bash
bw unlock                          # unlock your vault for the session
harpo run --profile dev -- claude  # Harpo injects only the authorized secrets
```

Harpo strips `BW_SESSION` from the child process, so the agent can use the
credentials but cannot reach your vault.

### Optional: let Harpo manage the unlock

With `policies.manage_unlock: true`, Harpo unlocks a locked vault itself
(prompting for the master password, never exporting the session to your shell).
Add `policies.unlock_cache: keychain` to cache the session in the OS keychain so
you are not re-prompted every run:

```bash
harpo unlock bw                    # prompt once; session cached with a TTL
harpo run --profile dev -- claude  # reuses the cached session
harpo lock bw                      # forget the cached session
```

See [policies](policies.md) for the knobs.

## One-off commands

For a single command (not a long agent session), use `exec`:

```bash
harpo exec --with gitlab.ad.read:GITLAB_TOKEN -- glab repo list
```

`exec` routes the command's output through a redactor, masking known secret
values and token formats.

## Inspecting and revoking

```bash
harpo session status     # show the active session for this project
harpo session list       # list all sessions
harpo session revoke current
harpo audit list         # see what was used, when (never the values)
```

## Next steps

- [Security model](security-model.md) — what Harpo protects and what it does not.
- [Policies](policies.md) — modes, TTLs and the `harpo.yml` policy knobs.
- [Providers](providers.md) — the vault abstraction and what is planned.
- Agents: [Claude Code](agents/claude-code.md), [Codex](agents/codex.md).
