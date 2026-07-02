# Harpo - Managed Unlock Specification

**Status:** Draft v0.1
**Date:** 2026-06-16
**Type:** Feature spec - Harpo manages the vault unlock instead of relying on a user-run unlock

---

## 1. Summary

Today the user unlocks the vault out-of-band (e.g. `bw unlock`), which leaves the
provider session token (`BW_SESSION`) in the **shell environment**, and Harpo
reads from that already-unlocked state. This feature lets **Harpo own the unlock
step**: it prompts for the master secret via a secure prompt, unlocks the
provider itself, holds the resulting session **in process memory** (optionally
cached in the **OS keychain** with a TTL), uses it only for its own resolution
calls, and never lets it touch the user's shell or the child agent.

This is a UX improvement and a **security improvement** - it removes the
long-lived `BW_SESSION` from the shell, shrinking the exposure window that
exists today.

## 2. Context & problem

Current flow:

```bash
bw unlock                              # sets BW_SESSION in the shell
harpo run --profile dev -- claude      # Harpo inherits BW_SESSION, strips it from the child
```

Problems:

- `BW_SESSION` lives in the **shell** for the whole session; any process the
  user launches there can read it - including `claude` if run directly (without
  `harpo run`), which would then inherit a live vault session.
- The unlock UX is provider-specific and outside Harpo, so Harpo can't guide it
  or enforce policy on it.

The MVP deliberately scoped this out ("Harpo does not unlock the vault") to keep
the trust surface small. This spec adds it as a post-MVP capability.

## 3. Goals / Non-goals

**Goals**

- Let Harpo perform the provider unlock using a secure master-secret prompt.
- Keep the resulting session out of the user's shell and out of the child agent.
- Optionally cache the session in the OS keychain with a TTL to avoid
  re-prompting on every run.
- Auto-detect a locked vault in `harpo run` / `harpo exec` and offer to unlock.

**Non-goals**

- Managing the provider **login/registration** (e.g. `bw login`, `vault`
  auth methods, account credentials). This spec only covers **unlock** of an
  already-registered account.
- 1Password desktop-app / biometric integration internals.
- Persisting the master secret anywhere, ever.
- Replacing the existing ambient-session behavior - it remains a fallback.

## 4. Principles

1. **Secure by default.** The master secret is read via a no-echo prompt and
   passed to the provider CLI over **stdin** - never as a command-line argument
   (process listing / shell history leak) and never as a lingering env var.
2. **Session in memory; never plaintext on disk.** The session token lives in
   Harpo's process memory. The only persistence allowed is the **OS keychain**,
   and only when explicitly enabled.
3. **Never inherited by the child.** The session is used only for Harpo's own
   provider subprocesses; the runner continues to strip it from the agent's env
   (it is already in `DangerousEnvVars`).
4. **Least lifetime.** Cached sessions carry a TTL; expiry forces a re-prompt.
5. **Honest about custody.** Harpo now touches the master secret; this is a
   higher-value asset and is treated with the same discipline as secret values
   (never logged, never audited, cleared after use).

## 5. Provider support matrix

"Managed unlock" only applies to providers with an interactive unlock step:

| Provider | Managed unlock | Mechanism |
|---|---|---|
| Bitwarden Password Manager (`bw`) | ✅ | `bw unlock --raw` via stdin → `BW_SESSION` |
| Keeper Commander (`keeper`) | ◑ | persistent-login / master password; provider-dependent |
| 1Password (`op`) | ✕ (delegated) | relies on desktop app / biometric / service-account token |
| HashiCorp Vault (`vault`) | ✕ (login, not unlock) | `vault login` / token - different model |
| Keeper Secrets Manager (`ksm`) | ✕ (no concept) | machine config; always "ready" |

Bitwarden Password Manager is the reference implementation.

## 6. Provider interface extension

Add an **optional** capability and interface in `internal/provider`:

- `Capabilities.SupportsUnlock bool` - advertised per provider so the CLI knows
  whether to offer unlock.
- A new optional interface implemented only by unlock-capable providers:

```go
// Unlocker is implemented by providers that have an interactive unlock step.
// Providers without one (e.g. KSM) do not implement it.
type Unlocker interface {
    // Unlock consumes a master secret (read via a secure prompt) and returns an
    // opaque session the provider uses for subsequent reads. The session value
    // is sensitive: never log it, never write it to disk in plaintext, never
    // pass it to child processes.
    Unlock(master string) (Session, error)
}

type Session struct {
    Name      string    // env var the provider CLI needs (e.g. "BW_SESSION")
    Value     string    // the session token (sensitive)
    ExpiresAt time.Time // provider/policy-derived; zero means unknown
}
```

The provider stores the active session internally and includes `Name=Value` in
the environment of its own `Resolve`/`Test`/list subprocesses - replacing
today's reliance on an ambient `BW_SESSION`.

### Session resolution order (per `Resolve`/`Status`)

1. In-memory session on the provider instance, if present and unexpired.
2. OS-keychain cached session, if enabled and unexpired.
3. If managed unlock is enabled and a TTY is available: prompt for the master
   secret, call `Unlock`, cache (if enabled), use.
4. Fallback: ambient env session (today's behavior).
5. Otherwise: a clear "vault is locked" error with the unlock command to run.

## 7. Master secret handling

- Read with a no-echo terminal prompt (e.g. `golang.org/x/term`'s
  `ReadPassword`); never from a flag or env var.
- Pass to the provider CLI via **stdin** (`bw unlock --raw --passwordenv` is
  *not* used; stdin avoids any env exposure).
- Clear the in-memory buffer as soon as the session is obtained. Note: Go's GC
  makes true zeroization best-effort (see `docs/adr/ADR-0001-stack-mvp-go.md`);
  a future Rust core would harden this with `zeroize`.
- Never written to the audit log, never echoed, never included in error text.

## 8. Session caching (OS keychain)

- Optional, off by default in `strict` mode; opt-in via policy.
- Backed by the OS keychain (Windows Credential Manager, macOS Keychain, Linux
  Secret Service). Candidate library: `github.com/zalando/go-keyring`
  (cross-platform). This would be the first runtime dependency beyond
  cobra/yaml/charm - a deliberate tradeoff to be confirmed at implementation.
- Cache key: `harpo/<provider-id>` (+ account where applicable). Stored value:
  the session token plus an expiry timestamp.
- **Never** cache the master secret - only the resulting session token.
- TTL governed by policy (`unlock_cache_ttl`), capped by `max_ttl`. On expiry,
  the entry is removed and a re-prompt is required.
- `harpo lock` (or `harpo unlock --forget`) removes the cached session.

This realizes the roadmap item "OS keychain for temporary cache"
(`docs/market-ready-spec.md`, v0.4).

## 9. Commands & UX

### `harpo unlock [provider]`

Prompts for the master secret and unlocks the named provider (or all
unlock-capable providers if omitted), caching per policy.

```text
$ harpo unlock bw
Master password for provider "bw": ********
✓ Unlocked. Session cached in the OS keychain (expires in 15m).
  BW_SESSION is held by Harpo and is not exported to your shell or to agents.
```

### Auto-unlock in `harpo run` / `harpo exec`

When the selected provider reports `locked` and managed unlock is enabled, Harpo
prompts inline before resolving:

```bash
harpo run --profile dev -- claude
# Vault is locked. Master password for "bw": ********
# ✓ Unlocked. Launching claude with 1 secret injected...
```

In non-interactive contexts (no TTY), Harpo does not prompt; it fails with the
explicit unlock instruction.

### `harpo lock`

Drops the in-memory session and removes any keychain-cached session.

## 10. Configuration & policy (`harpo.yml`)

```yaml
policies:
  manage_unlock: true        # may Harpo perform the unlock itself?
  unlock_cache: keychain     # keychain | none
  unlock_cache_ttl: 15m      # capped by max_ttl
```

Mode defaults:

- `strict`: `manage_unlock` allowed; `unlock_cache` defaults to `none`
  (re-prompt each run) unless explicitly set to `keychain`.
- `balanced`: `unlock_cache: keychain` permitted with a default TTL.

## 11. Security model

**What improves**

- `BW_SESSION` no longer needs to live in the shell; the exposure window for
  other shell processes (and a directly-run agent) shrinks or disappears.
- Unlock becomes policy-governed and auditable (event, not value).

**Residual risks (documented honestly)**

| Risk | Note / mitigation |
|---|---|
| Harpo now handles the master secret | Secure prompt + stdin only; never logged/persisted; cleared after use; Go GC limits zeroization (ADR-0001) |
| Cached session is a bearer token | Stored only in the OS keychain (OS-protected), with a TTL; never a plaintext file |
| Keychain strength varies by OS | Inherits the OS keychain's protections; Linux Secret Service requires a configured keyring |
| Re-prompt fatigue pushes long TTLs | `max_ttl` caps the cache TTL; strict mode defaults to no cache |

The child-agent guarantee is unchanged: the session is never passed to the
child (`runner.DangerousEnvVars` already strips `BW_SESSION` and peers).

## 12. Audit

Log unlock lifecycle events without secrets:

```json
{"time":"...","event":"vault.unlocked","provider":"bw","cache":"keychain","ttl_seconds":900,"result":"success"}
{"time":"...","event":"vault.unlock.cache_hit","provider":"bw","result":"success"}
{"time":"...","event":"vault.locked","provider":"bw","result":"success"}
```

Never record the master secret or the session token.

## 13. Acceptance criteria

- `harpo unlock bw` prompts (no echo), unlocks via stdin, and resolves secrets
  afterward without an ambient `BW_SESSION`.
- The master secret never appears in process arguments, stdout/stderr, the audit
  log, or any file.
- The session token is never written to disk except the OS keychain (when
  enabled) and never inherited by a child started by `harpo run`/`harpo exec`.
- With `unlock_cache: keychain`, a second `harpo run` within the TTL does not
  re-prompt; after expiry it does.
- `harpo lock` removes the cached session and a subsequent run re-prompts.
- In a non-TTY context Harpo fails with a clear unlock instruction instead of
  hanging on a prompt.

## 14. Testing

- Mock `Unlocker` provider: assert the master secret reaches `Unlock` only via
  the injected reader (never via args/env), and is absent from audit events.
- Regression: `runner.BuildChildEnv` still strips the session var.
- Keychain cache: TTL honored (hit before expiry, miss + re-prompt after), and
  `harpo lock` evicts.
- Leak guards: scan that no test ever observes the master secret or session in
  stdout/stderr/audit.

## 15. Rollout & relationship to the roadmap

- Lands after the provider set stabilizes; pairs naturally with the v0.4
  "OS keychain" roadmap item (`docs/market-ready-spec.md`).
- Bitwarden Password Manager first (reference); Keeper Commander next.
- `manage_unlock` defaults to **false** initially so existing flows are
  unchanged; users opt in.
