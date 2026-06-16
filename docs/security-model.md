# Security model

Harpo is a security tool and aims to be honest about both what it protects and
what it does not. This document explains the model in depth; for the quick
summary and vulnerability reporting see [`../SECURITY.md`](../SECURITY.md), and
for the full design rationale see [`mvp-spec.md`](mvp-spec.md) §7 and §15.

## The core idea

> The agent never receives vault access. It receives only temporary, limited,
> auditable access to the specific credentials you explicitly authorized.

Harpo sits between your existing vault and the AI coding agent. It resolves
only the secrets a profile authorizes, injects them into the child process it
launches, and records the use — without ever storing or printing the value.

## Non-negotiable invariants

These are enforced and guarded by tests:

1. Secret values are never printed by default.
2. Secret values are never written to the audit log.
3. Secret values are never stored in `harpo.yml`.
4. `BW_SESSION` (and other vault session tokens) are never passed to a child
   process started by `harpo run`.
5. TTL is mandatory for agent profiles in strict mode.
6. `.harpo/` is always gitignored.

## What Harpo protects against

- **Vault exposure to the agent.** Harpo never passes `BW_SESSION` to the child
  process, and `harpo agent setup` adds deny rules for `bw`/`op`/`vault`.
- **Secrets pasted into prompts.** Secrets are injected into the process
  environment; nothing is pasted into the chat.
- **Secrets in shell history / stdout.** Harpo never prints secret values by
  default; `harpo exec` redacts command output.
- **Secrets committed to Git.** `.env` is opt-in, written only inside the
  gitignored `.harpo/` directory.
- **Long-lived access.** Sessions carry a mandatory TTL in strict mode.

## What Harpo does NOT fully protect against

Being explicit here is part of the design:

- **Environment variables are plaintext inside the child process.** A secret
  delivered to an agent as an env var can be read by that process. Harpo
  reduces *scope and lifetime*; it does not make a running agent a perfect
  sandbox.
- **An agent with shell access can try to print variables it received.**
  Mitigated with deny rules, sandboxing and strict mode — not eliminated.
- **A rendered `.env` is plaintext on disk.** Allowed only as an explicit,
  confirmed, TTL-bound action inside `.harpo/`.
- **Redaction in interactive TUIs is not guaranteed.** `harpo run` wires the
  terminal directly to the agent and does not promise redaction; only
  `harpo exec` redacts captured output.

## Provider scope caveat

With Bitwarden Password Manager, the CLI has no fine-grained per-item scope: an
unlocked personal vault grants broad access in the user's context. Harpo
applies only *logical* scope and advertises this via provider capabilities
(`SupportsScopedAccess: false`). For strong scoping, a Secrets Manager provider
is the post-MVP path. See [providers](providers.md).

## Security modes

- **strict** (recommended for agents): TTL mandatory, `run` only, no `reveal`,
  no `.env` by default, `BW_SESSION` never inherited, audit mandatory.
- **balanced** (solo dev): `.env` and `reveal` allowed with explicit
  confirmation; TTL configurable.

See [policies](policies.md) for the full set of knobs, and the
[threat model](threat-model.md) for the threat-by-threat breakdown.
