# Harpo — Near-term Roadmap

**Date:** 2026-06-17
**Scope:** sequences the two designed features (managed unlock, proxy / MCP mode)
and defers the remaining providers.

This is the execution order for what is specced but not yet built. Full designs:
[`specs/managed-unlock.md`](specs/managed-unlock.md) and
[`specs/proxy-mcp-mode.md`](specs/proxy-mcp-mode.md). The long-term vision lives
in [`market-ready-spec.md`](market-ready-spec.md).

## Sequencing rationale

```
M1  Managed unlock  ──▶  M2  Proxy / MCP mode  ──▶  (later) remaining providers
```

- **Managed unlock comes first.** It delivers value on its own (removes the
  long-lived `BW_SESSION` from the shell) and gives Harpo **ownership of the
  session**. The MCP brokered exec benefits directly from that: the MCP server
  can hold/refresh the session instead of relying on an ambient `BW_SESSION`.
- **Proxy / MCP mode comes second.** It builds on M1 and reuses existing runtime
  pieces (`runner.RunWith`, `redact`, `policy`, `audit`).
- **Remaining providers are deferred.** The adapter pattern is established and
  mechanical; new providers don't block the UX work and can land anytime after.

Both features ship **off by default** (opt-in), so existing flows are unchanged.

## M1 — Managed unlock

Spec: [`specs/managed-unlock.md`](specs/managed-unlock.md). Reference provider:
Bitwarden Password Manager.

| Step | Description | Notes |
|---|---|---|
| M1.1 | `Unlocker` interface + `Capabilities.SupportsUnlock` in `internal/provider` | optional interface; only unlock-capable providers implement it |
| M1.2 | Bitwarden `Unlock` via secure prompt → stdin (`bw unlock --raw`); session held in memory; threaded into the adapter's own subprocess env | replaces reliance on ambient `BW_SESSION` |
| M1.3 | `harpo unlock [provider]` and `harpo lock` commands | secure no-echo prompt (`golang.org/x/term`) |
| M1.4 | Auto-unlock in `harpo run` / `harpo exec` when provider is `locked` (TTY only; clear error otherwise) | |
| M1.5 | OS keychain cache (TTL, capped by `max_ttl`); `harpo lock` evicts | **decision point:** add `github.com/zalando/go-keyring` (first runtime dep beyond cobra/yaml/charm) |
| M1.6 | `harpo.yml` policy knobs (`manage_unlock`, `unlock_cache`, `unlock_cache_ttl`); audit events; leak-guard tests | `manage_unlock` defaults false |

Suggested split: **M1a** = M1.1–M1.4 (in-memory, no cache), **M1b** = M1.5–M1.6
(keychain cache + policy). M1a is shippable without the new dependency.

**Status:** M1 is **implemented**.
- M1a: `Unlocker` interface + `SupportsUnlock`, Bitwarden `Unlock` (stdin) with
  the session held in memory, and auto-unlock in `run`/`exec`/`secret test`
  gated by `policies.manage_unlock` and a real TTY check.
- M1b: OS keychain session cache (`github.com/zalando/go-keyring`) with a TTL
  capped by `max_ttl`, the `unlock_cache`/`unlock_cache_ttl` policy knobs, the
  standalone `harpo unlock` / `harpo lock` commands, and `vault.unlocked` /
  `vault.unlock.cache_hit` / `vault.locked` audit events.

Risk/size: medium. Main risks: master-secret custody (secure prompt + stdin,
never args/env/logs) and Go GC zeroization limits (see
[`adr/ADR-0001-stack-mvp-go.md`](adr/ADR-0001-stack-mvp-go.md)).

## M2 — Proxy / MCP mode

Spec: [`specs/proxy-mcp-mode.md`](specs/proxy-mcp-mode.md). Depends on M1
(session ownership) — softly; read-only tools (M2.1) can land before M1b.

| Step | Description | Notes |
|---|---|---|
| M2.1 | MCP server skeleton over stdio + read-only tools: `harpo_session_status`, `harpo_secret_available`, `harpo_audit_tail` | **decision point:** Go MCP library choice; no value path — low risk |
| M2.2 | `harpo_exec` brokered exec: allowlist + shell-wrapper denial + alias authorization; reuse `runner.RunWith` + `redact.Writer` + `audit` | the key tool; value never returned |
| M2.3 | `harpo agent setup ... --mcp`: write `.mcp.json` (Claude) / Codex MCP config; keep deny rules; update the managed `CLAUDE.md`/`AGENTS.md` block + skill | safe path alongside the existing guardrail |
| M2.4 | `harpo.yml` policy (`policies.mcp`, `policies.proxy`); audit events; leak guards | `mcp.enabled` defaults false; `reveal` stays off |

**Status:** M2.1 + M2.2 are **implemented**.
- M2.1: `harpo mcp` runs an MCP stdio server
  (`github.com/modelcontextprotocol/go-sdk`) exposing the read-only,
  value-free tools `harpo_session_status`, `harpo_secret_available` and
  `harpo_audit_tail`, gated by `policies.mcp.enabled` (default false).
- M2.2: the `harpo_exec` brokered-exec tool — `policies.proxy.exec_allowlist`
  (empty denies all), shell-interpreter denial, profile-scoped alias
  authorization, secret resolution reusing provider/unlock, redacted output
  capture, and `mcp.exec` / `mcp.tool.denied` audit events. The secret value
  is never returned to the agent ("use without sight").

- M2.3: `harpo agent setup <claude|codex> --mcp --profile <p>` wires the MCP
  server (Claude `.mcp.json`; Codex config snippet in `AGENTS.md` since its
  config is global), enables `policies.mcp.enabled`, and adds an "MCP tools"
  section to the managed block — keeping the deny rules.

- M2.4: leak-guard tests prove the brokered-exec output is redacted and the
  secret value never reaches the tool result or the audit log; the policy
  surface (`mcp.enabled`, `proxy.exec_allowlist`) is complete; `reveal` and the
  loopback API proxy are intentionally left unimplemented. The proxy/MCP spec is
  marked implemented.

**M2 is complete.** Remaining work is the deferred providers (Bitwarden Secrets
Manager, AWS Secrets Manager, Infisical, Doppler) and any optional extensions
(loopback API proxy, `policies.mcp.tools` selection).

Deferred within M2: the loopback **API proxy** (§4b of the spec) and
`harpo_secret_reveal` (kept disabled).

Risk/size: medium–large. Main risks: brokered-command exfiltration (mitigated by
allowlist + no shell wrappers + optional human approval) and the new local
surface (stdio, no socket).

## Deferred — remaining providers

After M1/M2. Each follows the established adapter pattern (new package +
factory + docs; env denylist/deny rules as needed) and is small/mechanical:

- Bitwarden Secrets Manager (scoped machine identity)
- AWS Secrets Manager
- Infisical
- Doppler

## Cross-cutting

- **Rust/zeroize reevaluation.** M1 (master secret + session in memory) and M2
  (brokering resolved secrets) raise the value of deterministic memory
  zeroization. This strengthens the trigger in
  [`adr/ADR-0001-stack-mvp-go.md`](adr/ADR-0001-stack-mvp-go.md) for revisiting a
  Rust core if security positioning becomes the competitive axis.
- **Defaults stay safe.** Every new capability is opt-in; strict mode keeps
  `reveal` off, requires TTL, and never caches by default.
