# Threat model

This is the threat-by-threat breakdown for the Harpo MVP. It mirrors
[`mvp-spec.md`](mvp-spec.md) §7. For the broader explanation see the
[security model](security-model.md).

## Assumptions

- The user trusts their own vault and the machine Harpo runs on.
- The AI coding agent is **not** fully trusted: it can run commands, read
  files, and may be influenced by prompt injection from a README, issue, log,
  dependency, or other repository content.
- Harpo controls the child process environment only when it launches the
  process (the `harpo run` path).

## Threats considered and mitigations

| Threat | Mitigation in the MVP |
|---|---|
| Agent tries to list the vault | Harpo does not expose `BW_SESSION` to the child; setup recommends denying `bw`, `op`, `vault` |
| Secret ends up in a prompt/transcript | Harpo does not print secrets by default |
| Secret ends up in shell history | Harpo avoids secrets in command arguments and stdout |
| Secret committed in `.env` | Temporary `.env` lives in `.harpo/`; Harpo adds `.harpo/` to `.gitignore` |
| Prompt injection tells the agent to read `.env` | Setup generates deny rules; Harpo recommends sandbox/deny rules |
| Session lives too long | TTL mandatory for agent profiles in strict mode |
| Reuse in another project | A session can be bound to the project path |
| Wrong secret authorized | Interactive flow shows alias, provider, field, scope and destination before confirming |
| Command output leaks a secret | Best-effort redaction for commands run via `harpo exec` |

## Threats not fully resolved in the MVP

| Limitation | Note |
|---|---|
| Env vars can be read by the child process itself | Harpo reduces scope, but env vars are plaintext inside the process |
| An agent with shell access may print variables it received | Mitigated with permissions, instructions, deny rules and strict mode — not absolute |
| A temporary `.env` is plaintext on disk | Allowed only in an explicit mode, with a warning, TTL, and safe path |
| Bitwarden Password Manager has no fine-grained per-item scope in the CLI | Harpo applies logical scope; for strong scope, use a Secrets Manager provider in the future |
| Redaction in an interactive TUI may be unreliable | The MVP does not promise full redaction for `harpo run -- claude` |
