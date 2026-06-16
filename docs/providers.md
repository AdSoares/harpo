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

## Planned providers

The interface is designed to grow. Planned adapters include:

- Bitwarden Secrets Manager (machine identities, finer scope)
- 1Password
- HashiCorp Vault
- AWS Secrets Manager, GCP Secret Manager, Azure Key Vault
- Infisical, Doppler

See [`market-ready-spec.md`](market-ready-spec.md) §8.2 for priorities, and
[CONTRIBUTING](../CONTRIBUTING.md) for how to add a provider.
