# Harpo - Proxy / MCP Mode Specification

**Status:** Draft v0.1
**Date:** 2026-06-16
**Type:** Feature spec - let an agent interact with Harpo at runtime without secrets entering the agent's context

---

## 1. Summary

`harpo run` injects secrets into the agent's environment at launch. That is the
secure baseline, but it has friction: a secret added mid-session is not visible
to the already-running agent (process env is immutable), so you must restart.

**Proxy / MCP mode** lets the agent talk to Harpo **at runtime** through a small
set of **value-free tools**. The agent can discover what credentials exist and,
crucially, ask Harpo to **run a command with a credential injected** - receiving
only the (redacted) result. The secret value never enters the agent's
environment, prompt, or transcript.

The defining property is **use without sight**: the agent can *use* a credential
to get work done but never *possesses* the value - so it cannot paste it into a
prompt, commit it, or be prompt-injected into exfiltrating a value it never has.

## 2. Context & problem

From the env-var model:

- New/rotated secret mid-session ⇒ restart `harpo run`, or fall back to per-call
  `harpo exec` in a separate shell.
- Any path where the agent *reads* the value (e.g. `eval $(harpo export ...)`)
  puts the secret into the agent's context - the exact leak Harpo prevents.

Proxy / MCP mode removes the restart friction while keeping the value out of the
agent. It is the runtime, agent-driven complement to `harpo run`.

This realizes the roadmap items "MCP server" (`docs/market-ready-spec.md` §8.7),
"proxy mode" (§8.5.4) and "command broker" (Phase 4).

## 3. Core principle

> Tools return **metadata** and **brokered results** - never raw secret values.

- No tool returns a secret value by default. `reveal` is **disabled** by default
  (allowed only in `balanced`/`convenient` modes with explicit policy + strong
  confirmation, never to the model silently).
- The MCP/proxy server runs inside Harpo's process boundary; it uses the
  resolved secrets internally and strips them from anything it returns.
- The agent still must not run vault CLIs or read `.env` directly - the existing
  deny rules remain the guardrail. MCP/proxy is the **safe path**; deny rules
  block the **unsafe path**. They are complementary.

## 4. Delivery surfaces

### 4a. MCP server (`harpo mcp`)

For MCP-aware agents (Claude Code, Codex, others). Harpo runs a local MCP server
over **stdio** (no network surface) exposing the tool catalog in §5.

```bash
harpo mcp --profile dev      # serves over stdio; started by the agent's MCP config
```

stdio is preferred because it has no listening socket. The server inherits the
current session/profile and never exposes the session token.

### 4b. Local proxy (`harpo proxy`)

For tools/agents that are not MCP-aware. Two flavors:

- **Brokered subcommand:** the agent runs `harpo exec ...` (already exists) - a
  per-command broker that injects the secret into the child only.
- **Loopback API proxy (advanced):** `harpo proxy start --profile gitlab-dev`
  binds `127.0.0.1:<port>` and proxies specific upstream API calls with auth
  injected, so the agent calls a local endpoint and the token never enters its
  environment. Requires per-upstream adapters; this is the most advanced tier
  and is specified here only at the contract level.

The MVP of this feature is **4a (MCP) + brokered exec**; the API proxy (§4b
loopback) is a later extension.

## 5. Safe tool catalog

MCP tool names use `snake_case`; shown here with their contracts. None returns a
secret value.

### `harpo_session_status`

- **Input:** none.
- **Output:** `{ session_id, profile, agent, project, expires_in_seconds, secret_count, grants: [{ alias, destination }] }`.
- Mirrors `harpo session status`. No values.

### `harpo_secret_available`

- **Input:** `{ tag?: string }` (optional filter).
- **Output:** `[{ alias, destination, tags }]` - the aliases the current profile
  authorizes. **No values.**
- Lets the agent discover what it may use without guessing env var names.

### `harpo_exec` - the brokered exec (key tool)

- **Input:** `{ command: string, args: string[], with: [{ alias, env }] }`.
- **Behavior:** Harpo validates (see §6), resolves the secret(s) internally,
  runs `command args...` with the secret(s) injected into *that child's* env
  (vault session vars stripped), captures output through the redactor, audits.
- **Output:** `{ exit_code, stdout, stderr, truncated }` - stdout/stderr are
  **redacted** (known values + token formats). The secret value is never
  returned.

### `harpo_audit_tail` (optional)

- **Input:** `{ limit?: number }`.
- **Output:** recent audit events (no values). Read-only visibility.

### `harpo_secret_reveal` - **disabled by default**

- Returns a raw value. **Not registered** unless `policies.mcp.expose_reveal`
  is true (only in non-strict modes), and even then gated by confirmation. This
  exists for completeness; the recommended posture is to never expose it.

## 6. Brokered exec flow

```
Agent ──tool call: harpo_exec{command,args,with}──▶ Harpo MCP server
                                                     │
                                                     ├─ 1. session active & unexpired?
                                                     ├─ 2. command ∈ allowlist & not a shell wrapper?
                                                     ├─ 3. each alias authorized for this profile/agent?
                                                     ├─ 4. resolve secret(s) from provider (in-process)
                                                     ├─ 5. runner.RunWith(cmd, args, injected-env)
                                                     │      (DangerousEnvVars stripped; output via redact.Writer)
                                                     ├─ 6. audit: secret.injected + command + result (no values)
                                                     ▼
Agent ◀──{exit_code, stdout(redacted), stderr(redacted)}── Harpo
```

Reuses existing components: `runner.RunWith` (controlled child env),
`redact.Writer` (output masking), `policy` (authorization), `audit`.

### Validation rules (§6.2)

- **Command allowlist:** only commands in `policies.proxy.exec_allowlist` may be
  brokered (e.g. `gh`, `glab`, `aws`, `gcloud`, `az`, `kubectl`). Anything else
  is denied (or, with policy, escalated to human approval).
- **No shell wrappers:** `bash`, `sh`, `zsh`, `pwsh`, `cmd`, `python`, `node`,
  `ruby`, etc. are denied as the brokered command, because they would let the
  agent run arbitrary code with the credential present.
- **Alias authorization:** each `alias` must be mapped and included in the
  active profile; the destination env must be allowed.
- **Session binding:** brokered exec requires an active, unexpired session bound
  to the project path.

## 7. Security model

**What it guarantees**

- The secret value never enters the agent's environment, prompt, or transcript
  ("use without sight").
- A prompt-injected agent cannot exfiltrate a value it never receives.
- New/rotated secrets are usable live (each brokered call re-resolves) without a
  restart.

**Residual risks (honest)**

| Risk | Mitigation |
|---|---|
| A brokered command can still *act* with the credential (e.g. push to an attacker target) | Curated allowlist; no shell wrappers; optional human-in-the-loop for sensitive commands. The value is never readable, but it is usable by the allowlisted command. |
| Output redaction is best-effort | Don't allowlist commands that print the secret; redaction covers known values + token formats; document the limit. |
| MCP/proxy is a new local surface | Prefer stdio (no socket); the loopback API proxy binds `127.0.0.1` only with a per-session bearer token; the server never returns the session token. |
| Injection asks to broker an arbitrary command | Allowlist denies it; non-allowlisted ⇒ deny or explicit human approval. |
| `reveal` re-introduces leakage | Disabled by default; never exposed in strict mode. |

The child-agent guarantee from `harpo run` is unchanged and reused: brokered
children never inherit vault session vars (`runner.DangerousEnvVars`).

## 8. Changes to `agent setup`

`agent setup` gains an opt-in `--mcp` that wires the safe path **in addition to**
the existing deny rules (which stay).

### Claude Code (`harpo agent setup claude --mcp`)

- Writes/updates the project MCP config (`.mcp.json`) with the Harpo server:
  ```json
  {
    "mcpServers": {
      "harpo": { "command": "harpo", "args": ["mcp", "--profile", "dev"] }
    }
  }
  ```
- Keeps `.claude/settings.local.json` deny rules (still block `bw`/`op`/`vault`/
  `keeper`/`ksm`, `env`/`printenv`, reading `.env`/`.harpo`).
- Updates the managed `CLAUDE.md` block / the `harpo-secrets` skill to say:
  *use the `harpo_*` MCP tools (or `harpo exec`) for credentials; never expect
  or request raw values; never read env or vault CLIs.*

### Codex (`harpo agent setup codex --mcp`)

- Writes the equivalent MCP server entry to Codex's config and updates the
  managed `AGENTS.md` block with the same guidance, alongside the recommended
  sandbox/approval settings.

### Relationship

```
deny rules        → block the unsafe path (direct vault CLIs, env, .env)
MCP/proxy tools   → provide the safe path (status, available, brokered exec)
harpo run         → still the baseline for launch-time injection
```

## 9. Configuration & policy (`harpo.yml`)

```yaml
policies:
  mcp:
    enabled: true
    expose_reveal: false
    tools: [session_status, secret_available, exec, audit_tail]
  proxy:
    exec_allowlist: [gh, glab, aws, gcloud, az, kubectl]
    deny_shell_wrappers: true
    require_human_approval_outside_allowlist: true
```

Mode defaults: `strict` ⇒ `expose_reveal: false`, allowlist required,
shell wrappers denied. `balanced` ⇒ same, with optional `reveal` behind
confirmation.

## 10. Transport details

- **MCP:** JSON-RPC over **stdio**, launched by the agent's MCP config. No
  listening socket; lifetime tied to the agent session.
- **Loopback API proxy (advanced):** binds `127.0.0.1:<random-port>`; every
  request carries a per-session bearer token minted at `proxy start`; refuses
  non-loopback origins; never serves the session token or raw secrets.

## 11. Audit

```json
{"time":"...","event":"mcp.exec","profile":"dev","command":"glab","secret_alias":"gitlab.ad.read","destination":"env:GITLAB_TOKEN","result":"success"}
{"time":"...","event":"mcp.tool.denied","tool":"harpo_exec","reason":"command not in allowlist","result":"denied"}
```

Never record secret values, the session token, or full command output.

## 12. Out of scope

- Returning raw secret values to the agent by default (`reveal` stays off).
- Per-upstream API proxy adapters beyond the contract in §4b (a later tier).
- Managing the unlock (covered by `docs/specs/managed-unlock.md`).
- Non-stdio remote MCP transports.

## 13. Acceptance criteria

- With the MCP server registered, the agent can call `harpo_session_status` and
  `harpo_secret_available` and receive **no** secret values.
- `harpo_exec` runs an allowlisted command with the secret injected and returns
  redacted output; the secret value never appears in the tool result, the
  transcript, or the audit log.
- A non-allowlisted command (or a shell wrapper) is denied with a clear reason
  and an audit event.
- A newly mapped/rotated secret is usable via `harpo_exec` without restarting
  the agent.
- `reveal` is absent from the tool list in strict mode.

## 14. Testing

- Tool-contract tests: each tool's output schema; assert no value-bearing field.
- Brokered exec: secret injected into the child env, stripped from the returned
  output; allowlist + shell-wrapper denial; alias authorization enforced.
- Leak guards: scan tool results and audit events to ensure no secret value or
  session token is ever present.
- Redaction: brokered output containing a known value is masked.

## 15. Rollout & relationship to the roadmap

**Status: implemented** (roadmap M2.1–M2.4). The `harpo mcp` stdio server
exposes `harpo_session_status`, `harpo_secret_available`, `harpo_audit_tail` and
`harpo_exec`; `harpo agent setup --mcp` wires it; `policies.mcp.enabled` (default
false) and `policies.proxy.exec_allowlist` (empty = deny all) gate it; shell
interpreters are always denied; output is redacted and the secret value never
reaches the agent or the audit log (covered by leak-guard tests).

Intentionally **not** implemented: `harpo_secret_reveal` (no raw-value tool
exists) and the loopback **API proxy** (§4b) - both remain future/optional.

- Phase 4 in `docs/market-ready-spec.md`. Shipped after the provider set and
  managed unlock, since brokered exec benefits from Harpo owning the session.
- `policies.mcp.enabled` defaults to **false**; opt-in via
  `harpo agent setup ... --mcp`.
