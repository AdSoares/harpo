# Harpo — Market-Ready Specification

**Status:** Draft v0.1  
**Date:** 2026-06-16  
**Type:** Open source product with commercial potential for secrets security in development workflows with AI agents  
**Name:** Harpo  
**Main tagline:** _A local secret broker for AI coding agents._  
**Proposed category:** Agentic Secrets Security / AI Agent Secret Broker

---

## 1. Product vision

**Harpo** is an open source tool that creates a secure, local, and auditable layer between **AI coding agents** and **secrets stored in existing vaults**.

The product is born to solve an emerging market pain: developers are using agents like Claude Code, Codex, Cursor, Windsurf, Gemini CLI, and others to perform real tasks, but they still have to provide credentials, tokens, and access keys in an improvised way.

Harpo should position itself as:

> **The local secret broker for agentic development.**

Or, in English:

> **Harpo is the local secret broker that lets AI coding agents use credentials without seeing your vault, your raw secrets, or your long-lived environment.**

---

## 2. Market thesis

The adoption of AI agents in software development creates a new risk surface:

- agents run commands;
- agents read files;
- agents edit code;
- agents call CLIs;
- agents access APIs;
- agents can be influenced by prompt injection in README files, issues, logs, dependencies, and repository files.

At the same time, the productive workflow requires access to secrets:

- private repositories;
- registries;
- cloud providers;
- development databases;
- internal APIs;
- observability tools;
- CI/CD platforms.

The market already has secrets vaults, but the specific problem is different:

```text
How can a coding agent use temporary, limited credentials without giving it direct access to the vault and without forcing the developer to paste tokens into the prompt?
```

Harpo occupies that space.

---

## 3. Positioning

### 3.1 Category

```text
Agentic Secrets Security
```

or

```text
AI Agent Secret Broker
```

### 3.2 Short sentence

```text
Harpo gives AI coding agents scoped, temporary access to secrets from your existing vaults.
```

### 3.3 Long sentence

```text
Harpo is an open source local secret broker for AI coding agents. It lets developers and teams grant temporary, scoped, auditable access to credentials from existing vaults without pasting tokens into prompts, committing .env files, or giving agents direct access to the vault.
```

---

## 4. Why open source?

Harpo should be open source because:

- security requires trust and auditability;
- developers want to understand how secrets are handled;
- integration with agents and CLIs requires community;
- vault providers change frequently;
- bottom-up adoption in engineering teams is more likely with a free CLI;
- companies can adopt the core before purchasing advanced features.

### 4.1 Recommended model

```text
Open source core + commercial extensions opcionais
```

Open source core:

- local CLI;
- main providers;
- session grants;
- local policy;
- local audit;
- agent setup;
- basic plugin SDK.

Possible future commercial offerings:

- web console for teams;
- centralized policy management;
- centralized audit log;
- SIEM integration;
- enterprise SSO;
- remote approval;
- rotation management;
- compliance reports;
- policy marketplace;
- enterprise support.

---

## 5. Differentiation

Harpo should not compete directly with Bitwarden, 1Password, HashiCorp Vault, AWS Secrets Manager, or Azure Key Vault.

Harpo should position itself as a complementary layer.

### 5.1 It is not a password manager

Harpo does not store your digital life.

### 5.2 It is not a cloud secrets manager

Harpo does not try to replace corporate secrets infrastructure.

### 5.3 It is not just a `.env` manager

Harpo avoids `.env` as the default and uses `.env` only as a compatibility mode.

### 5.4 It is not just a Bitwarden CLI wrapper

Harpo adds:

- session grants;
- TTL;
- profiles;
- policy-as-code;
- agent setup;
- audit;
- redaction;
- controlled child process execution;
- provider abstraction;
- patterns specific to AI agents.

---

## 6. Strategic principle

Harpo should run the agent whenever possible.

Preferred flow:

```bash
harpo run --profile graces-dev -- claude
```

Less secure but supportable flow:

```bash
claude
# agente chama harpo depois
```

The difference is important:

```text
Quando Harpo executa o agente, Harpo controla o ambiente desde o início.
Quando o agente chama Harpo depois, parte do controle já foi perdida.
```

---

## 7. Target architecture

```text
┌────────────────────────────────────┐
│ Developer / Team                   │
└─────────────────┬──────────────────┘
                  │
                  ▼
┌────────────────────────────────────┐
│ Harpo CLI                          │
│ - commands                         │
│ - TUI approval                     │
│ - agent setup                      │
└─────────────────┬──────────────────┘
                  │
                  ▼
┌────────────────────────────────────┐
│ Harpo Core                         │
│ - policy engine                    │
│ - session grants                   │
│ - runner                           │
│ - redactor                         │
│ - audit                            │
└─────────────────┬──────────────────┘
                  │
        ┌─────────┴─────────┐
        ▼                   ▼
┌─────────────────┐ ┌──────────────────────┐
│ Provider SDK    │ │ Agent Integration     │
│ - Bitwarden     │ │ - Claude Code         │
│ - 1Password     │ │ - Codex CLI           │
│ - Vault         │ │ - Cursor/Windsurf     │
│ - AWS/GCP/Azure │ │ - MCP server          │
└─────────────────┘ └──────────────────────┘
        │                   │
        ▼                   ▼
┌─────────────────┐ ┌──────────────────────┐
│ Existing Vaults │ │ AI Coding Agents      │
└─────────────────┘ └──────────────────────┘

Optional future:

┌────────────────────────────────────┐
│ Harpo Team Console                 │
│ - central policies                 │
│ - audit aggregation                │
│ - approvals                        │
│ - compliance                       │
└────────────────────────────────────┘
```

---

## 8. Target features

## 8.1 Core CLI

Main commands:

```bash
harpo init
harpo provider add
harpo provider login
harpo provider status
harpo secret map
harpo secret list
harpo secret test
harpo profile create
harpo profile add-secret
harpo session start
harpo session status
harpo session revoke
harpo run
harpo exec
harpo env render
harpo audit list
harpo policy validate
harpo doctor
harpo agent setup
harpo scan
```

---

## 8.2 Provider SDK

Harpo should support providers via plugin.

### Priority providers

| Provider | Priority | Note |
|---|---:|---|
| Bitwarden Password Manager | High | great for solo dev |
| Bitwarden Secrets Manager | High | best for professional use and machine accounts |
| 1Password | High | strong adoption by devs and teams |
| HashiCorp Vault | High | standard in companies and infra |
| AWS Secrets Manager | Medium | relevant for AWS cloud |
| GCP Secret Manager | Medium | relevant for GCP cloud |
| Azure Key Vault | Medium | relevant for Microsoft companies |
| Doppler | Medium | popular as a developer-first secrets manager |
| Infisical | Medium | open source and developer-first |
| Local encrypted provider | Low/Medium | useful for demo, testing, and offline dev |

### Provider capabilities

Each provider should expose capabilities:

```json
{
  "supportsScopedAccess": true,
  "supportsAudit": true,
  "supportsRotation": false,
  "supportsDynamicSecrets": false,
  "supportsMachineIdentity": true,
  "supportsReadByAlias": true,
  "supportsList": true
}
```

This allows Harpo to warn the user when actual security depends only on local logical scope.

---

## 8.3 Policy-as-code

The `harpo.yml` file should evolve into a complete policy.

Example:

```yaml
version: 1

project:
  name: graces-api
  environments: [dev, staging]
  protected_environments: [production]

providers:
  bw-personal:
    type: bitwarden-password-manager

secrets:
  gitlab.ad.read:
    provider: bw-personal
    ref: "gitlab.com | ad | PAT | claude-code | read_api"
    field: password
    classification: internal
    allowed_agents: [claude, codex]
    allowed_destinations:
      - env:GITLAB_TOKEN
    max_ttl: 2h

profiles:
  graces-dev:
    mode: strict
    environment: dev
    agents: [claude, codex]
    ttl: 2h
    secrets:
      - gitlab.ad.read -> env:GITLAB_TOKEN

policies:
  require_ttl: true
  max_ttl: 8h
  allow_reveal: false
  allow_dotenv: false
  require_audit: true
  deny_production_by_default: true
  deny_direct_vault_cli_for_agents: true
  deny_env_printing: true
```

---

## 8.4 Session grants

The session grant is the most important product concept.

A grant should answer:

```text
Quem pode usar?
Qual agente/processo?
Qual projeto?
Qual secret?
Qual destino?
Por quanto tempo?
Com qual modo de segurança?
O uso foi auditado?
```

Example:

```json
{
  "id": "sess_20260616_abc123",
  "subject": "local-user",
  "agent": "claude",
  "project": "graces-api",
  "project_path": "C:/dev/graces-api",
  "profile": "graces-dev",
  "expires_at": "2026-06-16T12:00:00-03:00",
  "grants": [
    {
      "secret": "gitlab.ad.read",
      "destination": "env:GITLAB_TOKEN",
      "mode": "run-only"
    }
  ]
}
```

---

## 8.5 Delivery modes

### 8.5.1 `run` — recommended mode

```bash
harpo run --profile graces-dev -- claude
```

Harpo starts the child process and controls the environment.

### 8.5.2 `exec` — one-off command

```bash
harpo exec --with gitlab.ad.read:GITLAB_TOKEN -- glab repo list
```

### 8.5.3 `env render` — compatibility

```bash
harpo env render --profile graces-dev --out .harpo/.env.session --ttl 30m
```

### 8.5.4 `proxy` — advanced version

Instead of delivering the token to the process, Harpo can act as a local proxy:

```bash
harpo proxy start --profile gitlab-dev
```

The agent calls:

```bash
harpo gitlab projects list
```

Or accesses an authorized local endpoint:

```text
http://127.0.0.1:<port>/gitlab/projects
```

In this mode, the token never enters the agent's environment.

### 8.5.5 `fd/stdin` — advanced version

For compatible tools:

```bash
harpo exec --secret gitlab.ad.read --stdin -- some-command
```

Or use of a temporary file descriptor to avoid env vars.

---

## 8.6 Agent integrations

### Claude Code

Harpo should generate:

- `CLAUDE.md`;
- `.claude/settings.local.json`, when applicable;
- deny rules for `.env`, `.harpo/`, `bw`, `op`, `vault`, `printenv`, `env`, `set`;
- usage instructions for `harpo`;
- recommended commands.

Claude Code has permission and sandbox mechanisms that can block access to restricted resources and limit commands. Harpo should integrate with these mechanisms whenever possible.

### Codex CLI

Harpo should generate:

- `AGENTS.md`;
- sandbox recommendations;
- approval policy recommendations;
- local rules when available;
- recommended commands such as:

```bash
harpo run --profile graces-dev -- codex --sandbox workspace-write --ask-for-approval on-request
```

Codex technically separates sandbox and approval policy: the sandbox defines technical limits; the approval policy defines when the agent needs to ask for confirmation.

### Cursor, Windsurf, and others

Initial support via documentation and a temporary `.env`; advanced support via plugins or MCP.

---

## 8.7 Local MCP server

The market version should include an optional local MCP server:

```bash
harpo mcp start --profile graces-dev
```

The MCP should not expose raw secrets by default.

Safe MCP tools:

```text
harpo.session.status
harpo.secret.available
harpo.exec.approved
harpo.gitlab.projects.list
harpo.aws.identity.get
```

Dangerous tools, such as `harpo.secret.reveal`, should be disabled by default.

---

## 8.8 Authorization TUI

An interactive terminal interface should allow:

```text
Select secrets for this agent session:

[ ] aws.prod.root
[x] aws.dev.goritek
[x] gitlab.ad.read
[ ] database.prod.graces
[x] database.dev.graces

TTL: 2h
Mode: strict
Agent: claude
Project: graces-api

Confirm session grant? [y/N]
```

---

## 8.9 Advanced redaction

Harpo should be able to mask:

- exact secrets;
- common prefixes;
- tokens known by format;
- sensitive env vars;
- outputs of `harpo exec`;
- internal logs;
- error messages.

Known types:

```text
GITLAB_TOKEN
GITHUB_TOKEN
gh[pousr]_
AKIA...
AWS_SECRET_ACCESS_KEY
GOOGLE_APPLICATION_CREDENTIALS
AZURE_CLIENT_SECRET
DATABASE_URL
JWT
OPENAI_API_KEY
ANTHROPIC_API_KEY
```

---

## 8.10 Secret scanning

Harpo should include a simple local scanner:

```bash
harpo scan
harpo scan --staged
harpo scan --history
```

Goal:

- warn about a tracked `.env`;
- detect tokens in files;
- detect secrets in `harpo.yml`;
- install an optional pre-commit hook.

---

## 8.11 Rotation assistant

Harpo may not rotate all secrets automatically, but it can help:

```bash
harpo rotate plan gitlab.ad.read
```

Output:

```text
Secret: gitlab.ad.read
Provider: bitwarden
Target: GitLab PAT
Suggested rotation:
1. Create new GitLab PAT with same/minimal scopes.
2. Update Bitwarden item.
3. Run harpo secret test gitlab.ad.read.
4. Revoke old token in GitLab.
5. Confirm rotation complete.
```

For providers with a rotation API, implement:

```bash
harpo rotate gitlab.ad.read
```

---

## 8.12 Audit and compliance

Local audit:

```json
{
  "time": "2026-06-16T14:22:31-03:00",
  "event": "secret.injected",
  "user": "local-user",
  "agent": "claude",
  "project": "graces-api",
  "secret": "gitlab.ad.read",
  "destination": "env:GITLAB_TOKEN",
  "ttl": "2h",
  "policy": "strict",
  "result": "success"
}
```

The enterprise version can export to:

- JSONL;
- OpenTelemetry;
- SIEM;
- Datadog;
- Splunk;
- CloudWatch;
- Elastic;
- Grafana Loki.

---

## 9. Target security

### 9.1 Mandatory requirements

- Secrets must never be written to logs.
- Secrets must never be printed by default.
- Versionable configuration must never contain secrets.
- Sessions must have a TTL.
- Agent profiles must deny production by default.
- Harpo must remove the provider tokens from the agent's environment.
- `.env` must be an explicit exception, not the default.
- Redaction must protect stdout/stderr when Harpo controls the process.
- Harpo must be compatible with the agents' deny/sandbox.
- Harpo must have regression tests against leakage.

### 9.2 Security by level

#### Level 1 — Solo Developer

- personal provider;
- local policy;
- local audit;
- `run`, `exec`, temporary `.env`.

#### Level 2 — Team

- shareable profiles;
- policy-as-code;
- secrets manager providers;
- exportable audit;
- pre-commit scanning;
- approval per environment.

#### Level 3 — Enterprise

- SSO;
- RBAC;
- centralized policy;
- SIEM;
- corporate secrets manager;
- approval workflow;
- compliance reporting;
- dynamic secrets;
- central revocation.

---

## 10. Installation experience

### 10.1 Desired installation

```bash
# macOS/Linux
curl -fsSL https://harpo.dev/install.sh | sh

# Windows PowerShell
iwr https://harpo.dev/install.ps1 -useb | iex

# Homebrew
brew install harpo

# Winget
winget install harpo

# Chocolatey
choco install harpo

# Docker, para automações
 docker run --rm harpo/harpo version
```

### 10.2 Ideal first use

```bash
harpo init
harpo provider add bitwarden
harpo secret import --interactive
harpo profile create dev-agent
harpo agent setup claude
harpo run --profile dev-agent -- claude
```

---

## 11. Developer experience

Harpo should be:

- predictable;
- easy to explain;
- secure by default;
- fast;
- cross-platform;
- agent-friendly;
- human-friendly;
- scriptable;
- auditable.

### 11.1 Good error messages

Bad:

```text
Error: provider failed
```

Good:

```text
Bitwarden vault is locked.
Run:
  bw unlock
Then retry:
  harpo run --profile graces-dev -- claude

Harpo will not pass BW_SESSION to the agent process.
```

---

## 12. Target command surface

```bash
harpo init
harpo doctor
harpo version

harpo provider add <name> --type <type>
harpo provider login <name>
harpo provider status [name]
harpo provider list
harpo provider remove <name>

harpo secret map <alias>
harpo secret import --interactive
harpo secret list
harpo secret show <alias> --metadata
harpo secret test <alias>
harpo secret remove <alias>

harpo profile create <name>
harpo profile list
harpo profile show <name>
harpo profile add-secret <profile> <secret>
harpo profile remove-secret <profile> <secret>

harpo session start --profile <name>
harpo session status
harpo session revoke [id|current]
harpo session list

harpo run --profile <name> -- <command>
harpo exec --with <secret:ENV> -- <command>
harpo env render --profile <name> --out <path>
harpo reveal <secret>

harpo audit list
harpo audit export

harpo policy validate
harpo policy explain

harpo scan
harpo scan --staged
harpo scan --history

harpo agent setup claude
harpo agent setup codex
harpo agent setup cursor
harpo agent setup windsurf

harpo mcp start --profile <name>
```

---

## 13. Open source governance

### 13.1 Repository

Suggestion:

```text
github.com/harpo-sh/harpo
```

Alternatives:

```text
github.com/harpo-dev/harpo
github.com/harpocrates-dev/harpo
github.com/adnilson/harpo
```

### 13.2 Repository structure

```text
harpo/
  README.md
  LICENSE
  SECURITY.md
  CONTRIBUTING.md
  CODE_OF_CONDUCT.md
  docs/
    getting-started.md
    security-model.md
    providers.md
    agents/
      claude-code.md
      codex.md
    policies.md
    threat-model.md
  src/
  tests/
  examples/
    bitwarden-claude/
    bitwarden-codex/
    onepassword/
  scripts/
```

### 13.3 Mandatory documents

- `README.md`
- `SECURITY.md`
- `CONTRIBUTING.md`
- `docs/security-model.md`
- `docs/threat-model.md`
- `docs/providers.md`
- `docs/agents/claude-code.md`
- `docs/agents/codex.md`

### 13.4 License

Recommendation: **Apache-2.0**.

Reason:

- low friction for corporate adoption;
- good acceptance in infrastructure;
- explicit language about patents;
- allows monetization of services and extensions.

Strategic alternative:

- **MPL-2.0** if you want to better protect the core against direct closed forks.

Avoid initially:

- AGPL, if the goal is broad adoption in companies that avoid strong copyleft licenses.

---

## 14. Differentiation strategy for the market

### 14.1 Central theme

```text
AI agents should not need vault access to get work done.
```

### 14.2 Narrative

Before:

```text
Developers pasted secrets into prompts or stored them in .env files.
```

After:

```text
Developers grant scoped, temporary, auditable access to secrets through Harpo.
```

### 14.3 Key messages

- “Stop pasting tokens into AI coding agents.”
- “Use your existing vault. Add agent-safe access.”
- “Temporary secrets for agentic development.”
- “Give agents what they need, not your whole vault.”
- “Policy-as-code for secrets used by coding agents.”

---

## 15. Possible documentation/marketing pages

### Home

Title:

```text
Harpo — Agent-safe secrets for AI coding workflows
```

Subtitle:

```text
Grant Claude Code, Codex and other coding agents temporary access to credentials from your existing vaults without exposing your vault, pasting tokens or committing .env files.
```

### Security model

Clearly explain:

- what Harpo protects;
- what Harpo does not protect;
- risks of env vars;
- risks of `.env`;
- provider capabilities;
- threat model.

### Providers

Document:

- Bitwarden Password Manager;
- Bitwarden Secrets Manager;
- 1Password;
- Vault;
- AWS;
- GCP;
- Azure.

### Agents

Document:

- Claude Code;
- Codex;
- Cursor;
- Windsurf;
- GitHub Copilot;
- Gemini CLI.

---

## 16. Success metrics

### Open source

- GitHub stars.
- Monthly installs.
- Issues opened by real users.
- External contributors.
- Providers contributed by the community.
- Mentions in blogs, newsletters, and AI coding communities.

### Product

- Number of profiles created.
- Number of sessions started.
- Number of secrets accessed via grants.
- Use of `harpo run` vs `.env`.
- Adoption by teams.
- Reduction of versioned `.env` files.
- Integrations with agents.

### Security

- Zero secrets in internal logs.
- Automated leak prevention tests.
- Security issues answered with an SLA.
- Reviewable audit.
- Public threat model.

---

## 17. Market roadmap

### Phase 1 — Open source MVP

Goal: prove value for the solo dev.

- Core CLI.
- Bitwarden Password Manager provider.
- `harpo run`.
- `harpo exec`.
- `harpo env render`.
- `harpo agent setup claude/codex`.
- local audit.
- strong docs.

### Phase 2 — Developer adoption

Goal: community adoption.

- Homebrew/Winget/Chocolatey.
- 1Password provider.
- Bitwarden Secrets Manager provider.
- `harpo doctor`.
- basic scanner.
- real examples.
- simple website.
- short videos.

### Phase 3 — Team-ready

Goal: small teams.

- robust policy-as-code;
- shared profiles;
- exportable audit;
- Vault provider;
- AWS/GCP/Azure;
- pre-commit hooks;
- `harpo.yml` templates per stack.

### Phase 4 — Agent-native platform

Goal: relevance in security for agents.

- MCP server;
- authorization TUI;
- proxy mode;
- dynamic secrets;
- command broker;
- advanced Claude/Codex integration;
- plugin SDK.

### Phase 5 — Enterprise/commercial

Goal: monetization.

- web console;
- central policy;
- SSO;
- RBAC;
- remote approvals;
- audit aggregation;
- SIEM integrations;
- compliance reports;
- enterprise support.

---

## 18. Product risks

| Risk | Mitigation |
|---|---|
| Being seen as a “bw wrapper” | Focus on session grants, agent integrations, policy, and audit |
| Security promising more than it delivers | Document limitations clearly |
| Temporary `.env` becoming an insecure default | `run` as default; `.env` with a warning and explicit mode |
| Too many providers before a mature product | Start with 1-2 very good providers |
| Agents changing APIs/permissions | Modular architecture per agent integration |
| Companies distrusting new OSS | Public threat model, tests, security policy, signed releases |
| Difficulty monetizing | OSS core + team/enterprise features |

---

## 19. Requirements to look serious in the market

Before broad publicity, Harpo should have:

- An excellent README with the problem, demo, and threat model.
- Simple installation.
- A demo GIF or short video.
- A clear `SECURITY.md`.
- Documentation of limitations.
- Automated tests.
- Signed releases or checksums.
- Examples with Claude Code and Codex.
- Real integration with Bitwarden.
- A “Why Harpo?” page comparing it with alternatives.
- A “Security model” page.
- A “What Harpo does not protect against” page.

---

## 20. Suggested initial README

```md
# Harpo

Harpo is a local secret broker for AI coding agents.

It lets developers grant Claude Code, Codex and other coding agents temporary, scoped access to credentials from existing vaults without pasting tokens into prompts, committing .env files, or giving agents direct access to the vault.

## Why

AI coding agents can run commands, edit files and call developer tools. But real work often requires credentials. Pasting secrets into prompts or storing them in plaintext .env files is risky.

Harpo keeps your vault as the source of truth and gives agents only the secrets you explicitly authorize for the current project and session.

## Example

```bash
bw unlock
harpo run --profile my-project-dev -- claude
```

## Security model

- Harpo does not replace your vault.
- Harpo does not expose your vault session to agents.
- Harpo does not print secrets by default.
- Harpo uses temporary session grants with TTL.
- Harpo writes audit logs without secret values.
```

---

## 21. Content strategy

Article topics:

1. “Stop pasting secrets into AI coding agents.”
2. “Why AI coding agents need scoped secret grants.”
3. “The new security problem created by agentic development.”
4. “How to use Bitwarden safely with Claude Code.”
5. “Why `.env` is not enough for AI-native development.”
6. “Designing a secret broker for coding agents.”
7. “Agentic Secrets Security: a new category for software teams.”

---

## 22. Important product decisions

### 22.1 `run` should be the main path

Because it allows controlling the process environment from the start.

### 22.2 `.env` should be compatibility, not default

Because `.env` is plaintext on disk.

### 22.3 The Password Manager provider should display a warning

Because an unlocked personal vault may have broad access.

### 22.4 Secrets Manager should be recommended for teams

Because it allows better separation between personal secrets and machine/project secrets.

### 22.5 The threat model should be honest

Harpo reduces risk, but it does not turn a local agent into a perfectly isolated environment.

---

## 23. Possible commercial model

### Open Source Core

Free:

- CLI;
- basic providers;
- local policies;
- local audit;
- agent setup;
- plugin SDK.

### Harpo Team

Paid:

- policy bundles;
- audit export;
- team profile templates;
- shared configuration validation;
- provider enterprise packs;
- support.

### Harpo Enterprise

Paid:

- central console;
- SSO/SAML/OIDC;
- RBAC;
- remote approvals;
- SIEM;
- compliance;
- dynamic secrets;
- customer support;
- private plugins.

### Services

- deployment consulting;
- hardening of workflows with Claude/Codex;
- creation of per-company policies;
- AI-native secure development training.

---

## 24. Stack choice for the market

### Option A — Go

Advantages:

- simple single binary;
- great for CLI;
- easy distribution;
- good performance;
- strong adoption in DevOps.

### Option B — Rust

Advantages:

- strong reputation in security;
- performant binary;
- fine-grained control;
- good image for a security tool.

### Option C — .NET

Advantages:

- excellent productivity for a creator with a .NET background;
- cross-platform;
- single-file publish;
- good ecosystem;
- Spectre.Console offers great UX.

Pragmatic recommendation:

```text
Começar em .NET se isso acelerar muito a entrega.
Reavaliar Go/Rust apenas se distribuição, performance ou percepção de mercado virarem gargalo.
```

A delivered and useful product is more important than the ideal stack.

---

## 25. Version 1.0: definition of market-ready

Harpo 1.0 should have:

- A stable CLI.
- Simple installation on Windows, macOS, and Linux.
- Bitwarden Password Manager provider.
- Bitwarden Secrets Manager provider.
- 1Password provider.
- A mature `harpo run`.
- A mature `harpo exec`.
- A secure `harpo env render`.
- `harpo agent setup claude`.
- `harpo agent setup codex`.
- Validatable policy-as-code.
- Local and exportable audit.
- A basic scanner.
- Security docs.
- A public threat model.
- Signed releases/checksums.
- Automated leak prevention tests.

---

## 26. North Star

Harpo's conceptual metric is:

```text
Quantas vezes um desenvolvedor deixou de colar um secret no prompt ou salvar um .env inseguro porque o Harpo tornou o caminho seguro mais fácil?
```

The ultimate goal is not just to protect secrets.

The goal is to make development with AI agents **more productive, more auditable, and more secure by default**.

---

## 27. Final positioning sentence

```text
Harpo is the missing security layer between AI coding agents and developer secrets.
```
