# Providers

A **provider** is an adapter over an existing secrets vault. Harpo never
replaces your vault — it brokers scoped, temporary access to it. The provider
layer is pluggable so that, over time, Harpo can put access across *different*
vaults behind one consistent workflow.

## The provider interface

Every adapter implements `provider.Provider`
(`internal/provider/provider.go`):

- `ID()` / `Type()` — identity and type string.
- `Status()` — probe the CLI and report vault state (e.g. locked/unlocked).
- `Resolve(ref)` — return the secret value for a ref/field (sensitive; never
  logged or printed by callers).
- `Test(ref)` — verify a secret resolves, returning **metadata only** (length
  and a partial fingerprint), never the value.
- `Capabilities()` — declare what the provider can and cannot enforce.

## Capabilities

Each provider advertises capabilities so Harpo can warn when "security" is only
logical local scoping:

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

When `supportsScopedAccess` is `false`, the real boundary is whatever the
unlocked vault grants in the user's context — Harpo's scope is logical, not
enforced by the provider.

## MVP provider: Bitwarden Password Manager

Type: `bitwarden-password-manager`. Implemented by shelling out to the
`bw` CLI.

- Probes `bw status` to report `locked` / `unlocked` / `unauthenticated`.
- Resolves a ref to a **single item id** before reading the field: a search
  string that matches multiple items fails with a clear, value-free message
  (preferring a unique exact-name match); a ref that already looks like an item
  UUID is used directly.
- Never lists the vault on behalf of an agent and never writes values to stdout.
- `supportsScopedAccess` is `false`: an unlocked personal vault has broad access
  in the user's context, so Harpo applies only logical scope. For strong
  scoping, use a Secrets Manager provider (planned).
- Supports **managed unlock**: with `policies.manage_unlock: true`, Harpo can
  unlock a locked vault itself (master password prompt → stdin), holding the
  session in memory and never exporting it. See [policies](policies.md) and
  [`specs/managed-unlock.md`](specs/managed-unlock.md).

## Provider: Keeper Commander

Type: `keeper-commander`. Implemented by shelling out to the `keeper`
(Keeper Commander) CLI.

- Probes `keeper whoami` to report login state (`unlocked` when logged in,
  `unauthenticated` otherwise). Keeper Commander has no lock/unlock concept;
  it relies on the user's login, ideally a persistent-login config.
- Resolves the **password** field via `keeper find-password <ref>` (accepts a
  record UID or path/title, returns just the password). Other fields are read
  from `keeper get <ref> --format json --unmask` and extracted by field type or
  custom label.
- A Keeper record UID is a 22-character base64url string (e.g.
  `rvwIBG_ban2VTH64OsnzLn`) — different from the Bitwarden UUID shape.
- Never lists the vault on behalf of an agent and never writes values to stdout.
- `supportsScopedAccess` is `false`: like a personal vault, a logged-in session
  has broad access in the user's context, so Harpo applies only logical scope.
  For fine-grained, scoped machine access, a future **Keeper Secrets Manager**
  adapter is the path.

## Provider: Keeper Secrets Manager (KSM)

Type: `keeper-secrets-manager`. Implemented by shelling out to the `ksm` CLI.

- Probes `ksm profile list` to report readiness (`unlocked` when a profile is
  configured, `unauthenticated` otherwise). KSM has no lock/unlock concept; it
  uses a machine config bound to an Application.
- Resolves the field with Keeper notation
  (`ksm secret notation keeper://<UID>/field/<field>`), which returns the raw
  value — avoiding the masked/tabular output of `ksm secret get`. A ref that is
  not a UID is matched by title against `ksm secret list --json` and
  disambiguated (unique exact title match, else a value-free "use the UID" error).
- A KSM record UID is the same 22-character base64url shape as other Keeper UIDs.
- Never lists the vault on behalf of an agent and never writes values to stdout.
- **`supportsScopedAccess` is `true`** — unlike Keeper Commander, a KSM
  Application is bound to specific shared folders, so the access Harpo brokers
  is genuinely scoped (least privilege), not just logical scope. This makes KSM
  the recommended Keeper surface for agent-safe, machine-to-machine use.

## Provider: 1Password

Type: `1password`. Implemented by shelling out to the `op` CLI (v2).

- Probes `op whoami` to report sign-in state (`unlocked` when signed in,
  `unauthenticated` otherwise).
- Resolves the field with a 1Password **secret reference** via `op read
  "op://<vault>/<item>/<field>"`, which returns the raw value. The harpo `ref`
  is the `<vault>/<item>` path (a section may be included as
  `<vault>/<item>/<section>`); the field is appended automatically.
- Never lists the vault on behalf of an agent and never writes values to stdout.
- `supportsScopedAccess` is `false` by default: a regular `op signin` grants
  broad access across the user's vaults, so Harpo applies only logical scope. A
  1Password **service account** scoped to specific vaults provides real scope —
  an operational choice that the adapter reports conservatively as unscoped.

## Provider: HashiCorp Vault

Type: `hashicorp-vault`. Implemented by shelling out to the `vault` CLI.

- Probes `vault token lookup` to report auth state (`unlocked` when the token is
  valid, `unauthenticated` otherwise). Uses the standard `VAULT_ADDR` /
  `VAULT_TOKEN` environment.
- Resolves a field from the **KV** secrets engine via
  `vault kv get -field=<field> <path>`, which prints the raw value and
  auto-detects KV v1/v2. The harpo `ref` is the KV path (e.g. `secret/myapp`)
  and the field is the key within it.
- Never lists the vault on behalf of an agent and never writes values to stdout.
- **`supportsScopedAccess` is `true`** — access is governed by the token's
  policies and enforced server-side per path, so the scope is real, not merely
  logical. (A root token is broad; scope reflects the token you use.)
- The MVP adapter brokers static KV reads; Vault's native dynamic secrets,
  rotation and audit devices are out of scope for this adapter.

## Planned providers

The interface is designed to grow. Planned adapters include:

- Bitwarden Secrets Manager (machine identities, finer scope)
- AWS Secrets Manager, GCP Secret Manager, Azure Key Vault
- Infisical, Doppler

See [`market-ready-spec.md`](market-ready-spec.md) §8.2 for priorities, and
[CONTRIBUTING](../CONTRIBUTING.md) for how to add a provider.
