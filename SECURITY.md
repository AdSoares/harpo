# Security Policy

Harpo is a security tool. We aim to be honest about both what it protects and
what it does not. This document summarizes the threat model; the full design
rationale lives in [`harpo-mvp-spec.md`](harpo-mvp-spec.md) §7 and §15.

## What Harpo protects against

| Threat | Mitigation |
|---|---|
| Agent tries to read the whole vault | Harpo never passes `BW_SESSION` to the child; setup adds deny rules for `bw`/`op`/`vault` |
| Secret pasted into a prompt/transcript | Harpo injects secrets into the process env; nothing is pasted |
| Secret printed to stdout/history | Harpo never prints secret values by default |
| Secret committed in `.env` | `.env` is opt-in, written only inside `.harpo/`, which is gitignored |
| Session lives too long | TTL is mandatory for agent profiles in strict mode |
| Wrong secret authorized | Interactive confirmation shows alias, provider, field and destination |

## What Harpo does NOT fully protect against

Being explicit here is part of the design:

- **Environment variables are plaintext inside the child process.** A secret
  delivered to an agent as an env var can be read by that process. Harpo
  reduces *scope and lifetime*; it does not make the running agent a perfect
  sandbox.
- **An agent with shell access may try to print variables it received.**
  Mitigated with deny rules, sandboxing and strict mode — not eliminated.
- **A rendered `.env` is plaintext on disk.** Allowed only as an explicit,
  confirmed, TTL-bound action inside `.harpo/`.
- **Bitwarden Password Manager has no fine-grained per-item scope in the CLI.**
  An unlocked personal vault grants broad access in the user's context; Harpo
  applies only *logical* scope. For strong scope, use a Secrets Manager
  provider (post-MVP).

## Core invariants

These are enforced and guarded by tests:

1. Secret values are never printed by default.
2. Secret values are never written to the audit log.
3. Secret values are never stored in `harpo.yml`.
4. `BW_SESSION` (and other vault session tokens) are never passed to a child
   process started by `harpo run`.
5. TTL is mandatory for agent profiles in strict mode.
6. `.harpo/` is always gitignored.

## Reporting a vulnerability

Please report security issues privately to **adnilson.soares@gmail.com** rather
than opening a public issue. We will acknowledge receipt and work with you on a
coordinated disclosure timeline.
