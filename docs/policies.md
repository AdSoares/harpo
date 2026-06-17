# Policies

Harpo's behavior is governed by a **security mode** plus a set of policy knobs
in `harpo.yml`. The policy engine validates every request before any secret is
resolved or any session is created.

## Security modes

### `strict` (recommended for AI agents)

- TTL is **mandatory**.
- `harpo run` is the primary path.
- `reveal` is disabled.
- `.env` rendering is disabled by default.
- No wildcard secrets.
- `BW_SESSION` is never inherited by the child process.
- Auditing is mandatory; new secrets require interactive confirmation.

### `balanced` (recommended for a solo developer)

- TTL is recommended but configurable.
- A temporary `.env` is allowed inside `.harpo/`.
- `reveal` is allowed with strong confirmation.
- Auditing is mandatory; dangerous actions print clear warnings.

(A `convenient` mode is intentionally out of scope for the MVP.)

## `harpo.yml` policy knobs

```yaml
mode: strict

policies:
  allow_dotenv: false    # may `harpo env render` write a plaintext .env?
  allow_reveal: false    # may secrets be revealed in the terminal?
  manage_unlock: false   # may Harpo unlock the vault itself (vs. you running `bw unlock`)?
  unlock_cache: none     # cache the unlocked session: keychain | none
  unlock_cache_ttl: 15m  # session-cache TTL (capped by max_ttl)
  default_ttl: 2h        # TTL used when none is given (non-strict)
  max_ttl: 8h            # hard ceiling for any session TTL
```

How they are enforced:

- **TTL** — in strict mode a TTL is required; in any mode it must not exceed
  `max_ttl`. When omitted in a non-strict context, `default_ttl` applies.
- **`allow_reveal`** — `reveal` is always denied in strict mode, regardless of
  this flag; in other modes it additionally requires `allow_reveal: true`.
- **`allow_dotenv`** — `.env` rendering is always denied in strict mode; in
  other modes it additionally requires `allow_dotenv: true`, and the file may
  only be written inside `.harpo/`.
- **`manage_unlock`** — when `true`, an unlock-capable provider (e.g. Bitwarden)
  that reports `locked` is unlocked in-process: Harpo prompts for the master
  password (no echo) only when a terminal is available, holds the session in
  memory, and never exports it to the shell or the agent. Defaults to `false`
  (you unlock the vault yourself).
- **`unlock_cache` / `unlock_cache_ttl`** — when `unlock_cache: keychain`, the
  unlocked session is cached in the **OS keychain** with a TTL (default 15m,
  capped by `max_ttl`), so `harpo run` reuses it across runs without
  re-prompting. Only the session token is cached — never the master password.
  `harpo unlock` populates the cache; `harpo lock` evicts it. Defaults to
  `none`. See [`specs/managed-unlock.md`](specs/managed-unlock.md).

## Advisory warnings

Beyond hard rules, Harpo surfaces non-blocking warnings — for example, when a
secret alias looks production-like (`prod`, `production`, `root`, `admin`),
prompting you to double-check before granting it to an agent.

See the [security model](security-model.md) and
[`mvp-spec.md`](mvp-spec.md) §8 and §15 for the rationale.
