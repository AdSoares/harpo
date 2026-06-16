// Package vault implements the Provider interface for HashiCorp Vault by
// shelling out to the locally installed `vault` CLI.
//
// The MVP adapter brokers static reads from the KV secrets engine
// (`vault kv get -field=<field> <path>`), which prints the raw value. Vault
// also natively supports dynamic secrets, rotation and audit devices; brokering
// those is out of scope for this adapter.
//
// Security contract (mirrors the other adapters; see MVP spec §10.2, §15):
//   - It never lists the vault on behalf of an agent.
//   - It never writes secret values to stdout/stderr or logs.
//   - VAULT_TOKEN is never advertised as safe to inherit; the runner strips it
//     from child processes.
package vault

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/harpo-sh/harpo/internal/provider"
)

// TypeName is the provider type string used in harpo.yml.
const TypeName = "hashicorp-vault"

// Provider adapts the `vault` CLI to provider.Provider.
type Provider struct {
	id  string
	bin string // path/name of the vault binary
}

// New returns a HashiCorp Vault provider with the given config id.
func New(id string) *Provider {
	return &Provider{id: id, bin: "vault"}
}

func (p *Provider) ID() string   { return p.id }
func (p *Provider) Type() string { return TypeName }

// Capabilities for HashiCorp Vault. SupportsScopedAccess is true: access is
// governed by the token's policies, enforced server-side per path — real scope,
// not the merely-logical scope of a personal vault.
func (p *Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		CanList:                true,
		CanReadByRef:           true,
		SupportsScopedAccess:   true,
		SupportsAudit:          false,
		SupportsRotation:       false,
		SupportsDynamicSecrets: false,
	}
}

// Status probes the vault CLI and reports auth state via `vault token lookup`,
// which validates the current token and prints metadata but no secrets.
func (p *Provider) Status() (provider.Status, error) {
	st := provider.Status{Type: TypeName, SafeForInheritance: false}
	if _, err := exec.LookPath(p.bin); err != nil {
		st.CLIFound = false
		st.Vault = provider.StateNotFound
		st.Detail = "`vault` not found in PATH"
		return st, nil
	}
	st.CLIFound = true

	if _, err := p.run("token", "lookup"); err != nil {
		st.Vault = provider.StateUnauthed
		st.Detail = "no valid token (set VAULT_ADDR and run `vault login`)"
		return st, nil
	}
	st.Vault = provider.StateUnlocked
	return st, nil
}

// kvGetArgs builds the `vault kv get` arguments for reading a single field from
// a KV path. It is pure so the command construction can be unit-tested without
// the vault CLI.
func kvGetArgs(path, field string) []string {
	return []string{"kv", "get", "-field=" + field, path}
}

// Resolve returns the secret value for the given ref/field. The value is
// sensitive and must never be logged or printed by callers.
//
// The ref is the KV secret path (e.g. "secret/myapp") and the field is the key
// within it. `vault kv get -field=<field> <path>` prints the raw value and
// auto-detects KV v1/v2.
func (p *Provider) Resolve(ref provider.Ref) (string, error) {
	field := ref.Field
	if field == "" {
		field = "password"
	}
	if ref.Ref == "" {
		return "", errors.New("vault ref must be a KV secret path (e.g. \"secret/myapp\")")
	}
	out, err := p.run(kvGetArgs(ref.Ref, field)...)
	if err != nil {
		return "", err
	}
	value := strings.TrimRight(out, "\r\n")
	if value == "" {
		return "", fmt.Errorf("vault returned an empty value for field %q", field)
	}
	return value, nil
}

// Test resolves the secret and reports metadata only — never the value.
func (p *Provider) Test(ref provider.Ref) (provider.TestResult, error) {
	value, err := p.Resolve(ref)
	if err != nil {
		return provider.TestResult{Resolved: false}, err
	}
	return provider.TestResult{
		Resolved:    true,
		Length:      len(value),
		Fingerprint: provider.Fingerprint(value),
	}, nil
}

// run executes `vault <args...>` and returns trimmed stdout. stderr is captured
// only for error context and is not surfaced to general output.
func (p *Provider) run(args ...string) (string, error) {
	cmd := exec.Command(p.bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New("vault " + args[0] + ": " + msg)
	}
	return stdout.String(), nil
}
