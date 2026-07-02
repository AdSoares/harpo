# Harpo - MVP Specification

**Status:** Draft v0.1  
**Date:** 2026-06-16  
**Type:** Local open source CLI for secure secret brokering between vaults and AI coding agents  
**Name:** Harpo  
**Tagline:** _A local secret broker for AI coding agents._

---

## 1. MVP Vision

**Harpo** is a local CLI that lets the user authorize, in a temporary and limited way, the use of credentials stored in a vault by tools such as **Claude Code**, **Codex CLI**, local scripts, third-party CLIs, and development automations.

The goal of the MVP is not to replace Bitwarden, 1Password, Vault, or other secret managers. The goal is to create a local, simple, and secure layer between:

```text
AI agent / local process
        ↓
      Harpo
        ↓
Existing vault provider
        ↓
Credential authorized by the user
```

The core principle of the MVP is:

> **The agent does not receive access to the vault. The agent receives only temporary, limited, and auditable access to specific credentials that the user has authorized.**

---

## 2. The Problem Harpo Solves

Developers who use AI coding agents frequently need to grant access to:

- GitLab, GitHub, or Azure DevOps tokens;
- AWS, GCP, or Azure credentials;
- API keys;
- connection data for development databases;
- tokens for internal tools;
- credentials for temporary environments.

The common flow is either insecure or unproductive:

```text
1. User creates a token.
2. Saves it in a text file, .env, or loose notes.
3. Copies and pastes it into the agent's prompt.
4. The secret may end up in history, transcript, shell, Git, logs, or screenshots.
```

Harpo proposes a better flow:

```text
1. User saves the secret in the vault.
2. User maps a local alias in Harpo.
3. User authorizes a session with TTL and scope.
4. Harpo starts the agent or process with only the allowed secrets.
5. Harpo audits usage without storing the secret value.
```

---

## 3. MVP Target Audience

### Primary persona

**Solo or senior developer using local coding agents**, especially:

- Claude Code;
- Codex CLI;
- GitHub Copilot CLI/agents;
- Cursor/Windsurf via terminal;
- local automation scripts;
- CLIs such as `glab`, `gh`, `aws`, `az`, `gcloud`.

### Secondary persona

**Small development teams** that want to standardize secure credential usage across projects without initially adopting a complex enterprise solution.

---

## 4. MVP Scope

The MVP should deliver a useful and secure flow with few features, but very well polished.

### Included in the MVP

- Local `harpo` CLI.
- Initial provider for **Bitwarden Password Manager via the Bitwarden CLI (`bw`)**.
- Versionable `harpo.yml` file, without secret values.
- Mapping of local aliases to vault items/fields.
- Project profiles.
- Temporary sessions with TTL.
- Child process execution with secrets injected as environment variables.
- Optional rendering of a temporary `.env` inside `.harpo/`.
- Basic setup for Claude Code and Codex.
- Local auditing without storing the secret.
- Basic redaction for commands executed via Harpo.
- `strict` and `balanced` modes.

### Out of the MVP

- Desktop UI.
- Cloud service.
- Secret sharing between users.
- Automatic token rotation.
- Enterprise RBAC.
- SIEM/centralized logging.
- Complete provider for all vaults.
- Resident daemon.
- Remote policy synchronization.
- MCP server.
- Official IDE plugin.

---

## 5. Design Principles

### 5.1 Secure by default

Harpo must be born secure by default. The default operation must not print secrets, must not create a persistent `.env`, and must not expose the vault session to the agent.

### 5.2 Convenient by explicit choice

More convenient flows, such as creating a `.env`, revealing a secret in the terminal, or increasing the TTL, should exist but require an explicit choice by the user.

### 5.3 Agent-safe, not agent-blind

Harpo must accept that AI agents can execute commands, interpret files, and suffer prompt injection. Therefore, the control cannot rely solely on instructions in `CLAUDE.md` or `AGENTS.md`.

### 5.4 Least privilege

Each session must release only the authorized secrets, to the authorized destination, and for the authorized duration.

### 5.5 No vault access for agents

The agent must not receive `BW_SESSION`, the vault token, the master password, the provider API key, or any ability to list the vault.

### 5.6 Auditable without leaking

Harpo must log what was used, when, by which profile, and in which project, without logging the secret value.

---

## 6. Security References Used in the Design

The MVP design should align with recognized best practices:

- OWASP Secrets Management Cheat Sheet: centralization, rotation, auditing, and least privilege for secrets.
- Bitwarden CLI: the `unlocked` state depends on an active key in `BW_SESSION`; the `locked` state indicates the absence of that active session.
- Claude Code: permissions and sandbox can block access to restricted resources; deny rules help prevent attempts to access sensitive files and commands.
- Codex CLI: sandbox and approval policy are different and complementary controls; local mode can operate with `workspace-write` and `on-request` approval.

Links:

- https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html
- https://bitwarden.com/help/cli/
- https://code.claude.com/docs/en/permissions
- https://code.claude.com/docs/en/sandboxing
- https://developers.openai.com/codex/agent-approvals-security
- https://developers.openai.com/codex/concepts/sandboxing
- https://developers.openai.com/codex/cli/reference

---

## 7. MVP Threat Model

### 7.1 Threats considered

| Threat | Mitigation in the MVP |
|---|---|
| Agent tries to list the vault | Harpo does not expose `BW_SESSION` to the child process; setup recommends denying `bw`, `op`, `vault` |
| Secret ends up in prompt/transcript | Harpo does not print the secret by default |
| Secret ends up in shell history | Harpo avoids secrets in command arguments and stdout |
| Secret is committed in `.env` | The temporary `.env` lives in `.harpo/`; Harpo adds `.harpo/` to `.gitignore` |
| Prompt injection instructs reading `.env` | Setup generates rules/instructions to block reading; Harpo recommends sandbox/deny rules |
| Session lasts too long | TTL mandatory for agent profiles |
| Improper reuse in another project | The session can be bound to the project path |
| User authorizes the wrong secret | Interactive flow shows alias, provider, field, scope, and destination before confirming |
| Command output leaks a secret | Basic redaction for commands executed via `harpo exec` |

### 7.2 Threats not fully resolved in the MVP

| Limitation | Note |
|---|---|
| Environment variables can be read by the child process itself | Harpo reduces scope, but env vars remain plaintext inside the process |
| An agent with shell access may try to print variables it received | Mitigate with permissions, instructions, deny rules, and `strict` mode; there is no absolute guarantee |
| The temporary `.env` is plaintext on disk | Allowed only in explicit mode, with a warning, TTL, and a secure path |
| The Bitwarden Password Manager provider does not offer fine-grained per-item scope in the CLI itself | Harpo applies logical scope; for strong scope, use Bitwarden Secrets Manager in the future |
| Redaction in an interactive TUI may not be reliable | The MVP should avoid promising full redaction in the `harpo run -- claude` mode |

---

## 8. Security Modes

### 8.1 `strict`

Recommended mode for AI agents.

Rules:

- TTL mandatory.
- `harpo run` as the primary mode.
- No `reveal`.
- No `.env` by default.
- No secret wildcards.
- Does not inherit `BW_SESSION` to the child process.
- Mandatory auditing.
- Interactive confirmation for new secrets.

### 8.2 `balanced`

Recommended mode for the solo developer.

Rules:

- TTL recommended, but configurable.
- Temporary `.env` allowed inside `.harpo/`.
- `reveal` allowed with strong confirmation.
- Mandatory auditing.
- Clear warnings for dangerous actions.

### 8.3 Out of the MVP: `convenient`

Extreme-convenience mode, to be evaluated later.

---

## 9. MVP Architecture

```text
┌─────────────────────────┐
│ User                     │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Harpo CLI                │
│ - commands               │
│ - interactive UX         │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Policy Engine            │
│ - mode                   │
│ - TTL                    │
│ - profile                │
│ - authorized path        │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Session Manager          │
│ - grants                 │
│ - expiration             │
│ - metadata               │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Provider Adapter         │
│ - Bitwarden CLI          │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Existing vault           │
│ - Bitwarden              │
└─────────────────────────┘

┌─────────────────────────┐
│ Runner                   │
│ - starts Claude/Codex    │
│ - injects filtered env   │
│ - strips BW_SESSION      │
└─────────────────────────┘

┌─────────────────────────┐
│ Audit Logger             │
│ - records access         │
│ - never records value    │
└─────────────────────────┘
```

---

## 10. Components

### 10.1 CLI Core

Responsible for command parsing, UX, validations, error messages, and integration between modules.

### 10.2 Provider Adapter

Interface for integration with vaults.

MVP:

```text
bitwarden-password-manager
```

The initial implementation may call the locally installed `bw` binary.

Responsibilities:

- check whether `bw` is installed;
- check `bw status`;
- trigger `bw sync` when requested;
- fetch a specific field of the authorized item;
- never list the vault for the agent;
- never return data directly to stdout.

### 10.3 Policy Engine

Validates:

- profile;
- security mode;
- TTL;
- project path;
- allowed destination;
- whether `.env` is allowed;
- whether `reveal` is allowed;
- whether the requested secret is authorized.

### 10.4 Session Manager

Creates and manages temporary grants.

The session stores only metadata:

```json
{
  "id": "sess_abc123",
  "created_at": "2026-06-16T10:00:00-03:00",
  "expires_at": "2026-06-16T12:00:00-03:00",
  "project_path": "C:/dev/graces-api",
  "agent": "claude",
  "profile": "graces-dev",
  "allowed_secrets": [
    {
      "alias": "gitlab.ad.read",
      "destination": "env:GITLAB_TOKEN"
    }
  ]
}
```

It must not store secret values.

### 10.5 Runner

Executes the child process with a controlled environment:

```text
harpo run --profile graces-dev -- claude
```

Responsibilities:

- resolve the authorized secrets;
- assemble the child process environment;
- remove dangerous variables from the inherited environment, especially `BW_SESSION`;
- inject only the authorized variables;
- record the audit;
- end the session at the end if configured.

### 10.6 Env Renderer

Renders a temporary `.env`.

Rules:

- default path: `.harpo/.env.session`;
- require confirmation;
- create `.harpo/` if it does not exist;
- add `.harpo/` to `.gitignore`;
- show the TTL;
- delete the file when the session is revoked, when possible;
- record the audit.

### 10.7 Audit Logger

Records events without secrets.

Example:

```json
{
  "time": "2026-06-16T14:22:31-03:00",
  "event": "secret.injected",
  "profile": "graces-dev",
  "agent": "claude",
  "project": "graces-api",
  "secret_alias": "gitlab.ad.read",
  "destination": "env:GITLAB_TOKEN",
  "mode": "run",
  "ttl_seconds": 7200,
  "result": "success"
}
```

### 10.8 Redactor

MVP:

- redaction for commands executed via `harpo exec`;
- best-effort redaction for Harpo errors;
- do not promise full redaction for an interactive TUI.

---

## 11. File Structure

### 11.1 `harpo.yml`

Versionable file, without secrets.

```yaml
version: 1

project:
  name: graces-api
  allowed_paths:
    - .

mode: strict

providers:
  bitwarden-personal:
    type: bitwarden-password-manager

secrets:
  gitlab.ad.read:
    provider: bitwarden-personal
    ref: "gitlab.com | ad | PAT | claude-code | read_api"
    field: password
    default_env: GITLAB_TOKEN
    tags: [gitlab, dev, agent-safe]

profiles:
  graces-dev:
    ttl: 2h
    agent: claude
    secrets:
      - secret: gitlab.ad.read
        env: GITLAB_TOKEN

policies:
  allow_dotenv: false
  allow_reveal: false
  default_ttl: 2h
  max_ttl: 8h
```

### 11.2 `.harpo/`

Local, non-versioned directory.

```text
.harpo/
  sessions/
  audit.log.jsonl
  .env.session
```

### 11.3 `.gitignore`

Harpo must ensure:

```gitignore
.harpo/
.env
.env.*
!.env.example
```

---

## 12. MVP Commands

### 12.1 `harpo init`

Initializes the project.

```bash
harpo init
```

Creates:

- `harpo.yml`;
- `.harpo/`;
- an entry in `.gitignore`;
- initial instructions.

Options:

```bash
harpo init --mode strict
harpo init --agent claude
harpo init --agent codex
```

---

### 12.2 `harpo provider add`

Adds a provider.

```bash
harpo provider add bitwarden-personal --type bitwarden-password-manager
```

Validates:

- whether `bw` exists in the PATH;
- whether `bw status` is accessible;
- whether the vault is `unlocked` or `locked`.

---

### 12.3 `harpo provider status`

```bash
harpo provider status
```

Example output:

```text
Provider: bitwarden-personal
Type: bitwarden-password-manager
CLI: bw found
Vault status: unlocked
Safe for agent inheritance: no
```

---

### 12.4 `harpo secret map`

Maps a local alias to a vault item.

```bash
harpo secret map gitlab.ad.read \
  --provider bitwarden-personal \
  --ref "gitlab.com | ad | PAT | claude-code | read_api" \
  --field password \
  --env GITLAB_TOKEN
```

Must show a confirmation:

```text
Alias: gitlab.ad.read
Provider: bitwarden-personal
Vault ref: gitlab.com | ad | PAT | claude-code | read_api
Field: password
Default destination: env:GITLAB_TOKEN

No secret value will be stored in harpo.yml.
Confirm? [y/N]
```

---

### 12.5 `harpo secret list`

Lists aliases, never values.

```bash
harpo secret list
```

---

### 12.6 `harpo secret test`

Tests whether Harpo can resolve the secret, without printing the value.

```bash
harpo secret test gitlab.ad.read
```

Output:

```text
Secret resolved successfully.
Length: 20 chars
Fingerprint: sha256:ab12...9f
Value: [redacted]
```

---

### 12.7 `harpo profile create`

Creates a profile.

```bash
harpo profile create graces-dev --ttl 2h --agent claude
```

---

### 12.8 `harpo profile add-secret`

```bash
harpo profile add-secret graces-dev gitlab.ad.read --env GITLAB_TOKEN
```

---

### 12.9 `harpo session start`

Creates an explicit session.

```bash
harpo session start --profile graces-dev --ttl 2h
```

---

### 12.10 `harpo session status`

```bash
harpo session status
```

Output:

```text
Session: sess_abc123
Profile: graces-dev
Agent: claude
Project: C:/dev/graces-api
Expires in: 01:42:10
Secrets: 1
- gitlab.ad.read -> env:GITLAB_TOKEN
```

---

### 12.11 `harpo session revoke`

```bash
harpo session revoke current
```

Removes the local metadata and deletes the associated temporary `.env`, if it exists.

---

### 12.12 `harpo run`

Primary mode.

```bash
harpo run --profile graces-dev -- claude
```

```bash
harpo run --profile graces-dev -- codex --sandbox workspace-write --ask-for-approval on-request
```

Rules:

- Harpo resolves the authorized secrets.
- Harpo removes `BW_SESSION` from the environment inherited by the child process.
- Harpo injects only the profile's secrets.
- Harpo records the audit.
- Harpo does not print values.

---

### 12.13 `harpo exec`

Executes a specific command with a temporary secret.

```bash
harpo exec --with gitlab.ad.read:GITLAB_TOKEN -- glab repo list
```

Recommended for one-off commands.

---

### 12.14 `harpo env render`

Renders a temporary `.env`.

```bash
harpo env render --profile graces-dev --out .harpo/.env.session --ttl 30m
```

Must display a warning:

```text
This will write plaintext secrets to disk.
Path: .harpo/.env.session
This path is ignored by Git.
Expires in: 30m
Confirm? [y/N]
```

---

### 12.15 `harpo audit list`

```bash
harpo audit list
```

Does not show secrets.

---

### 12.16 `harpo agent setup claude`

Generates guidance files for Claude Code.

```bash
harpo agent setup claude
```

Can create/update:

- `CLAUDE.md`;
- `.claude/settings.local.json` when possible;
- `.gitignore`.

Suggested content for `CLAUDE.md`:

```md
# Secrets policy

Use Harpo for secrets.

Allowed:
- harpo run
- harpo exec
- harpo session status

Forbidden:
- bw
- op
- vault
- env
- printenv
- set
- Get-ChildItem Env:
- cat .env
- type .env
- git add .env
- git add .harpo
- harpo reveal

Never print secrets.
Never ask for tokens in the prompt.
Never write secrets to versioned files.
```

---

### 12.17 `harpo agent setup codex`

Generates an `AGENTS.md` with a secrets policy.

```bash
harpo agent setup codex
```

Suggested content:

```md
# Secrets policy

Use Harpo for credentials.

When you need credentials:
- harpo session status
- harpo exec with approved commands

Do not run directly:
- bw
- op
- vault
- printenv
- env
- cat .env
- type .env

Recommended mode for Codex:
- workspace-write
- approval on-request
```

---

## 13. Primary MVP UX

### 13.1 First use

```bash
harpo init --mode strict --agent claude
harpo provider add bitwarden-personal --type bitwarden-password-manager
harpo secret map gitlab.ad.read --provider bitwarden-personal --ref "gitlab.com | ad | PAT | claude-code | read_api" --field password --env GITLAB_TOKEN
harpo profile create graces-dev --ttl 2h --agent claude
harpo profile add-secret graces-dev gitlab.ad.read --env GITLAB_TOKEN
harpo agent setup claude
harpo run --profile graces-dev -- claude
```

### 13.2 Daily use

```bash
bw unlock
harpo run --profile graces-dev -- claude
```

### 13.3 One-off command

```bash
harpo exec --with gitlab.ad.read:GITLAB_TOKEN -- glab repo list
```

---

## 14. Technical Requirements

### 14.1 Cross-platform

The MVP must work on:

- Windows 10/11 with PowerShell;
- Linux;
- macOS.

### 14.2 Recommended stack

Recommended option for the project's creator:

- **.NET 10 Console** or **.NET 8 LTS** if you want LTS stability;
- `System.CommandLine` for the CLI;
- `Spectre.Console` for terminal UX;
- `YamlDotNet` for `harpo.yml`;
- tests with xUnit;
- packaging as single-file publish.

Strong alternatives for the market:

- Go: simple distribution, single binary, great for CLI.
- Rust: excellent for security and distribution, but a steeper curve.

As an MVP, the best choice is the one that reduces development friction and allows fast delivery.

---

## 15. MVP Security Requirements

### Mandatory

- Never print a secret by default.
- Never log a secret in the audit log.
- Never save the secret value in `harpo.yml`.
- Never pass `BW_SESSION` to a child process started by `harpo run`.
- TTL mandatory for agent profiles.
- `.harpo/` must be ignored by Git.
- The temporary `.env` must require confirmation.
- `reveal` must be disabled in `strict` mode.
- `secret test` must show only a partial fingerprint and safe metadata.
- Errors must redact sensitive values when detected.

### Recommended

- Detect whether `.env` is tracked in Git and alert.
- Detect whether `harpo.yml` contains token-like patterns.
- Alert if the TTL is longer than 8h.
- Alert if a secret alias looks like production, for example: `prod`, `production`, `root`, `admin`.
- Alert if the user tries to use `.env` outside `.harpo/`.

---

## 16. Internal Data Model

### Secret alias

```json
{
  "alias": "gitlab.ad.read",
  "provider": "bitwarden-personal",
  "ref": "gitlab.com | ad | PAT | claude-code | read_api",
  "field": "password",
  "default_env": "GITLAB_TOKEN",
  "tags": ["gitlab", "dev", "agent-safe"]
}
```

### Profile

```json
{
  "name": "graces-dev",
  "agent": "claude",
  "ttl": "2h",
  "secrets": [
    {
      "alias": "gitlab.ad.read",
      "destination": "env:GITLAB_TOKEN"
    }
  ]
}
```

### Session grant

```json
{
  "id": "sess_abc123",
  "profile": "graces-dev",
  "agent": "claude",
  "project_path": "C:/dev/graces-api",
  "created_at": "2026-06-16T10:00:00-03:00",
  "expires_at": "2026-06-16T12:00:00-03:00",
  "allowed_secrets": ["gitlab.ad.read"]
}
```

---

## 17. Provider Interface

Conceptual interface:

```ts
interface SecretProvider {
  id: string;
  type: string;

  status(): Promise<ProviderStatus>;
  authenticate(): Promise<AuthResult>;
  resolveSecret(ref: SecretRef): Promise<SecretValue>;
  testSecret(ref: SecretRef): Promise<SecretTestResult>;
  capabilities(): ProviderCapabilities;
}
```

Capabilities:

```json
{
  "canList": true,
  "canReadByRef": true,
  "supportsScopedAccess": false,
  "supportsAudit": false,
  "supportsRotation": false,
  "supportsDynamicSecrets": false
}
```

For Bitwarden Password Manager via the CLI, `supportsScopedAccess` should be `false` because Harpo applies logical scope, but the unlocked vault may have broad access in the user's context.

---

## 18. MVP Acceptance Criteria

The MVP is considered functional when:

- A user can initialize a project with `harpo init`.
- A Bitwarden CLI provider can be configured.
- A secret alias can be mapped without saving the value to disk.
- A profile can be created with a TTL.
- `harpo run --profile X -- claude` injects `GITLAB_TOKEN` into the child process.
- `BW_SESSION` is not inherited by the child process.
- `harpo exec` works for a simple command.
- `harpo env render` creates `.harpo/.env.session` with confirmation.
- `harpo audit list` shows events without sensitive values.
- `harpo agent setup claude` creates useful instructions.
- `harpo agent setup codex` creates useful instructions.
- Automated tests cover secret resolution, policy, session, runner, and audit log.

---

## 19. Essential Tests

### Unit

- TTL parsing.
- `harpo.yml` validation.
- Profile validation.
- Blocking `reveal` in `strict` mode.
- Redaction of known values.
- Generating an audit log without a secret.
- Assembling an environment without `BW_SESSION`.

### Integration

- Mock Bitwarden provider.
- `harpo run` with a child process that prints the presence/absence of a variable.
- `harpo exec` with a simple command.
- `harpo env render` creating the file and `.gitignore`.

### Security

- Ensure the secret does not appear in stdout.
- Ensure the secret does not appear in stderr.
- Ensure the secret does not appear in the audit log.
- Ensure `harpo.yml` does not receive a secret value.
- Ensure `BW_SESSION` is not passed to the child process.

---

## 20. Post-MVP Roadmap

### v0.2

- Bitwarden Secrets Manager provider.
- 1Password provider.
- Better redaction.
- Per-agent policy templates.
- `harpo doctor`.

### v0.3

- AWS Secrets Manager provider.
- HashiCorp Vault provider.
- Basic secret scanner for the repository.
- Support for `.env.example` generated from the profile.

### v0.4

- Local MCP server.
- Authorization TUI.
- OS keychain for temporary cache.
- Per-organization/team policies.

---

## 21. Initial Open Source Project Positioning

Harpo should be presented as:

```text
A local secret broker for AI coding agents.
```

Short description:

```text
Harpo lets developers grant temporary, scoped access to secrets from their existing vaults without pasting credentials into prompts, committing .env files, or giving AI coding agents direct access to the vault.
```

---

## 22. Recommended License for the MVP

To maximize adoption in the software market:

- **Apache-2.0**: a good option for an infrastructure tool, allows commercial use and broad contributions.
- **MIT**: simple and permissive, but with less explicit patent language.
- **MPL-2.0**: an intermediate option if you want to ensure that changes to the core stay open without preventing commercial use.

Initial recommendation:

```text
Apache-2.0
```

Reason: it reduces friction for companies, allows broad adoption, and is familiar for infrastructure tools.

---

## 23. MVP Definition of Success

The MVP will be successful if it enables this flow without significant friction:

```bash
bw unlock
harpo run --profile my-project-dev -- claude
```

And if the user is able to say:

> “Now my agents can work with development credentials without me pasting a token into the chat, without scattering `.env` files, and without giving access to my entire vault.”
