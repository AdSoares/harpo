# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project is

**Harpo** is an open-source, local CLI **secret broker for AI coding agents**. It sits between an AI agent (or any local process) and an existing secrets vault, granting *temporary, scoped, auditable* access to specific credentials — without giving the agent access to the vault itself, without pasting tokens into prompts, and without committing `.env` files.

Core principle that governs every design decision:

> The agent never receives vault access. It receives only temporary, limited, auditable access to specific credentials the user explicitly authorized.

## Current state (read this first)

This repository currently contains **only specification documents — no code, no build, no tests, no chosen stack committed.** Do not assume any toolchain exists yet. The two specs are the source of truth:

- `docs/mvp-spec.md` — the **MVP** to build first. Scoped down to one provider (Bitwarden CLI), `run`/`exec`/`env render`, agent setup for Claude/Codex, local audit. **Build against this spec.**
- `docs/market-ready-spec.md` — the long-term v1.0 / market vision (more providers, MCP server, proxy mode, TUI, policy-as-code, scanning, rotation, commercial tiers). Treat as roadmap context, **not** as MVP scope.

When implementing, the MVP spec's "Fora do MVP" (out of scope) list is binding — do not pull market-ready features into the MVP without explicit decision from Ad.

## Stack (decided: Go for the MVP)

**The MVP is built in Go.** Decision recorded in `docs/adr/ADR-0001-stack-mvp-go.md` (2026-06-16).

Rationale in short: the MVP's value is in broker correctness and DX (policy, session grants, stripping `BW_SESSION`, audit-without-value), and Go's `os/exec` + environment handling is idiomatic for the Runner — the most central component. Go also gives trivial cross-compilation/single-binary distribution and matches the founder's existing skillset. The dominant secret-leak surface is the env var handed to the child process (MVP spec §7.2), which neither Go nor Rust eliminates — so Rust's main edge (deterministic in-memory zeroization via `zeroize`/`secrecy`) buys little at this stage.

**Rust is the post-traction reevaluation**, not a rejected option: if security positioning becomes the competitive axis, rewriting the core in Rust is justified (and a good public story). Do not let the language choice become a bottleneck before there is a real user.

Implementation notes for Go:
- CLI framework: `cobra` (or stdlib `flag` if kept minimal) — pick when scaffolding.
- Config (`harpo.yml`): a YAML lib such as `gopkg.in/yaml.v3`.
- Terminal UX: `lipgloss`/`bubbletea` (Charm) if rich UX is needed; keep MVP lean.
- Even with GC, treat secret values as short-lived `[]byte`, avoid putting them in `string`s that linger, and never place them in any type with a `String()`/`Error()` that could be logged.

Must run cross-platform on **Windows (PowerShell), Linux, and macOS**.

## Architecture (planned)

The CLI decomposes into these components (see MVP spec §9–§10). The data flow is strictly one-directional toward the vault and back into a controlled child-process environment:

```
CLI Core → Policy Engine → Session Manager → Provider Adapter → Vault
                                                  ↓
                                Runner (spawns agent w/ filtered env)
                                                  ↓
                                          Audit Logger
```

- **CLI Core** — command parsing, interactive UX, validation, error messages.
- **Provider Adapter** — pluggable interface to vaults. MVP = `bitwarden-password-manager`, implemented by shelling out to the `bw` binary (`bw status`, `bw sync`, field lookup). Must never list the vault for the agent and never dump secrets to stdout. Capabilities are declared per provider (`canList`, `supportsScopedAccess`, etc.) so Harpo can warn when "security" is only logical local scoping (true for Bitwarden Password Manager).
- **Policy Engine** — validates profile, security mode, TTL, project path, destination, whether `.env`/`reveal` are allowed, and whether the requested secret is authorized.
- **Session Manager** — creates time-bound grants. **Stores metadata only, never secret values** (id, timestamps, project path, agent, profile, allowed-secret aliases + destinations).
- **Runner** — the primary delivery path. Resolves authorized secrets, builds the child environment, **strips dangerous inherited vars (especially `BW_SESSION`)**, injects only authorized vars, audits, optionally ends the session on exit.
- **Env Renderer** — opt-in `.env` materialization to `.harpo/.env.session`; requires explicit confirmation, ensures `.harpo/` is gitignored, has a TTL, deletes on revoke.
- **Audit Logger** — JSONL events recording what/when/which profile/project — **never the value**.
- **Redactor** — masks known secret values/token formats in `harpo exec` output and Harpo's own errors. Does **not** promise full redaction inside interactive TUIs (e.g. `harpo run -- claude`).

### Security modes
- `strict` (recommended for agents): TTL mandatory, `run` only, no `reveal`, no `.env` by default, no wildcards, never inherit `BW_SESSION`, audit mandatory, confirm new secrets.
- `balanced` (solo dev): TTL configurable, `.env` allowed within `.harpo/`, `reveal` with strong confirmation, audit mandatory.

## Non-negotiable security invariants

These cut across every component. Any change that could violate one is a bug, and the test suite is expected to guard them (MVP spec §15, §19):

1. Never print a secret value by default.
2. Never write a secret value to the audit log.
3. Never store a secret value in `harpo.yml` (config is versionable and secret-free).
4. Never pass `BW_SESSION` (or other vault session tokens) to a child process started by `harpo run`.
5. TTL is mandatory for agent profiles.
6. `.harpo/` must be gitignored; `.env` materialization requires explicit confirmation.
7. `reveal` is disabled in `strict` mode.
8. `secret test` shows only length + partial fingerprint, never the value.
9. Errors must redact detected sensitive values.

When adding any feature, ask: does the agent gain a path to the raw vault, a long-lived session, or an unaudited secret? If yes, it is wrong by design.

## Key files & config formats

- `harpo.yml` — versioned, secret-free project config: providers, secret aliases (alias → vault ref + field + default env, with tags), profiles (TTL + agent + secret→env mappings), and `policies`. See MVP spec §11.1 / §16.
- `.harpo/` — local, **untracked**: `sessions/`, `audit.log.jsonl`, `.env.session`.
- `harpo init` must also seed `.gitignore` with `.harpo/`, `.env`, `.env.*`, `!.env.example`.
- `harpo agent setup claude|codex` generates `CLAUDE.md` / `AGENTS.md` plus (for Claude) `.claude/settings.local.json` deny rules blocking `bw`, `op`, `vault`, `env`, `printenv`, `cat .env`, `harpo reveal`, etc. Note the agent-safety principle: control must **not** rely on instructions in `CLAUDE.md`/`AGENTS.md` alone, because agents can be prompt-injected.

## Conventions

- **All commits and all repository documents are written in English** (ADRs, READMEs, design docs, code comments). This is a public open-source product, so docs are part of the product surface. This **overrides** the Company OS default (`F:\02-company-os\CLAUDE.md`), which sets documentation to Portuguese (BR) — that default does not apply here.
- Commits follow **Conventional Commits**.
- This project lives under `marca-pessoal/` (founder's personal projects). It is positioned as a public open-source product (target license **Apache-2.0**).
- Project tagline: *A local secret broker for AI coding agents.*
- The specs live in `docs/mvp-spec.md` and `docs/market-ready-spec.md` (translated to English from the original PT-BR drafts).
