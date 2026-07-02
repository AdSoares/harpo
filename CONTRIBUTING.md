# Contributing to Harpo

Thanks for your interest in Harpo - a local secret broker for AI coding agents.
Contributions of all kinds are welcome: bug reports, fixes, new providers,
docs, and tests.

Harpo is a **security tool**, so contributions are reviewed with that lens.
Please read the security invariants below before changing anything that touches
secrets, the runner, or the audit log.

## Ground rules

- Be respectful and constructive. Assume good intent. This project follows our
  [Code of Conduct](CODE_OF_CONDUCT.md).
- All commits, code, comments, and documents are written in **English**.
- Discuss non-trivial changes in an issue before opening a large PR.

## Prerequisites

- Go 1.26+ (the toolchain version is pinned in `go.mod`).
- For provider work against a real vault: the relevant CLI installed (the MVP
  uses the Bitwarden CLI, `bw`).

## Development workflow

```bash
# Build everything
go build ./...

# Static checks
go vet ./...
gofmt -l .        # must print nothing; run `gofmt -w .` to fix

# Run the full test suite
go test ./...

# Run a single package or test
go test ./internal/runner/
go test ./internal/runner/ -run TestBuildChildEnvStripsBWSession -v
```

Please make sure `go build`, `go vet`, `gofmt -l .`, and `go test ./...` are all
clean before opening a PR. CI runs the same checks (see
`.github/workflows/ci.yml`).

## Project layout

The CLI is a single Go module. Core logic lives under `internal/`:

- `config` - `harpo.yml` types and loading.
- `provider` (+ `provider/bitwarden`) - the vault abstraction and adapters.
- `policy` - security-mode invariants (TTL, reveal, dotenv, authorization).
- `session` - metadata-only session grants.
- `runner` - building the child-process environment.
- `audit` - JSONL audit logging.
- `redact` - masking of secret values and token formats.
- `cli` - the cobra command tree.

See [`CLAUDE.md`](CLAUDE.md) for the bigger-picture architecture.

## Security invariants (non-negotiable)

Any change must preserve these. They are guarded by tests; do not weaken them:

1. Secret values are never printed by default.
2. Secret values are never written to the audit log.
3. Secret values are never stored in `harpo.yml`.
4. `BW_SESSION` (and other vault session tokens) are never passed to a child
   process started by `harpo run`.
5. TTL is mandatory for agent profiles in strict mode.
6. `.harpo/` is always gitignored.

If your change could affect any of these, add or update a test that proves the
invariant still holds. See [`SECURITY.md`](SECURITY.md) for the full threat
model.

## Adding a provider

Providers are the main extension point. To add one:

1. Implement the `provider.Provider` interface (`internal/provider/provider.go`):
   `ID`, `Type`, `Status`, `Resolve`, `Test`, and `Capabilities`.
2. Declare honest `Capabilities` - in particular, set `SupportsScopedAccess`
   to `false` when access is only logically scoped, so Harpo can warn the user.
3. Never list the vault on behalf of an agent and never write secret values to
   stdout/stderr or logs.
4. Register the new type in the provider factory (`newProvider` in
   `internal/cli/provider.go`).
5. Add tests for the pure logic (e.g. ref resolution/disambiguation) so the
   adapter can be verified without the vault CLI.

## Commit and PR conventions

- Follow **[Conventional Commits](https://www.conventionalcommits.org/)**
  (e.g. `feat:`, `fix:`, `docs:`, `ci:`, `refactor:`, `test:`).
- Keep PRs focused; one logical change per PR where possible.
- Include tests for behavior changes and update docs when behavior changes.
- Branch from `main`; open the PR against `main`.

## Reporting security vulnerabilities

Do **not** open a public issue for a security vulnerability. Report it privately
as described in [`SECURITY.md`](SECURITY.md).

## License

By contributing, you agree that your contributions are licensed under the
project's [Apache-2.0](LICENSE) license.
