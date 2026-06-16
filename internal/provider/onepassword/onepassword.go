// Package onepassword implements the Provider interface for 1Password by
// shelling out to the locally installed `op` CLI (v2).
//
// Resolution uses 1Password secret references (`op read "op://<vault>/<item>/
// <field>"`), which return the raw value — the same scripting-oriented path
// 1Password itself recommends.
//
// Security contract (mirrors the other adapters; see MVP spec §10.2, §15):
//   - It never lists the vault on behalf of an agent.
//   - It never writes secret values to stdout/stderr or logs.
//   - The user's 1Password session / service-account token is never advertised
//     as safe to inherit; the runner strips OP_SESSION*, OP_SERVICE_ACCOUNT_TOKEN
//     and OP_CONNECT_TOKEN from child processes.
package onepassword

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/harpo-sh/harpo/internal/provider"
)

// TypeName is the provider type string used in harpo.yml.
const TypeName = "1password"

// Provider adapts the `op` CLI to provider.Provider.
type Provider struct {
	id  string
	bin string // path/name of the op binary
}

// New returns a 1Password provider with the given config id.
func New(id string) *Provider {
	return &Provider{id: id, bin: "op"}
}

func (p *Provider) ID() string   { return p.id }
func (p *Provider) Type() string { return TypeName }

// Capabilities for 1Password. SupportsScopedAccess is false by default: a
// regular `op signin` grants broad access across the user's vaults, so Harpo
// applies only logical scope. (A 1Password service account scoped to specific
// vaults provides real scope; that is an operational choice, not guaranteed by
// the adapter, so we conservatively report false to keep the user warned.)
func (p *Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		CanList:                true,
		CanReadByRef:           true,
		SupportsScopedAccess:   false,
		SupportsAudit:          false,
		SupportsRotation:       false,
		SupportsDynamicSecrets: false,
	}
}

// Status probes the op CLI and reports sign-in state via `op whoami`, which
// prints the account/identity but no secrets.
func (p *Provider) Status() (provider.Status, error) {
	st := provider.Status{Type: TypeName, SafeForInheritance: false}
	if _, err := exec.LookPath(p.bin); err != nil {
		st.CLIFound = false
		st.Vault = provider.StateNotFound
		st.Detail = "`op` not found in PATH"
		return st, nil
	}
	st.CLIFound = true

	if _, err := p.run("whoami"); err != nil {
		st.Vault = provider.StateUnauthed
		st.Detail = "not signed in (run `op signin`)"
		return st, nil
	}
	st.Vault = provider.StateUnlocked
	return st, nil
}

// Resolve returns the secret value for the given ref/field. The value is
// sensitive and must never be logged or printed by callers.
//
// The ref is the 1Password item path "<vault>/<item>" (optionally including a
// section, "<vault>/<item>/<section>"); combined with the field it forms a
// secret reference read with `op read`, which returns the raw value.
func (p *Provider) Resolve(ref provider.Ref) (string, error) {
	field := ref.Field
	if field == "" {
		field = "password"
	}
	reference, err := buildReference(ref.Ref, field)
	if err != nil {
		return "", err
	}
	out, err := p.run("read", reference)
	if err != nil {
		return "", err
	}
	value := strings.TrimRight(out, "\r\n")
	if value == "" {
		return "", fmt.Errorf("1password returned an empty value for field %q", field)
	}
	return value, nil
}

// buildReference assembles a 1Password secret reference (op://vault/item/field)
// from the harpo ref and field. It tolerates a ref that already carries the
// "op://" scheme or a trailing slash. It is pure so it can be unit-tested
// without the op CLI.
func buildReference(ref, field string) (string, error) {
	path := strings.TrimSuffix(strings.TrimPrefix(ref, "op://"), "/")
	if path == "" {
		return "", errors.New("1password ref must be a \"vault/item\" path")
	}
	if field == "" {
		return "", errors.New("1password requires a field")
	}
	return "op://" + path + "/" + field, nil
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

// run executes `op <args...>` and returns trimmed stdout. stderr is captured
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
		return "", errors.New("op " + args[0] + ": " + msg)
	}
	return stdout.String(), nil
}
