// Package bitwarden implements the Provider interface for Bitwarden Password
// Manager by shelling out to the locally installed `bw` CLI.
//
// Security contract (MVP spec §10.2, §15):
//   - It never lists the vault on behalf of an agent.
//   - It never writes secret values to stdout/stderr or logs.
//   - The BW_SESSION token is the user's; it is never advertised as safe to
//     inherit and the runner strips it from child processes.
package bitwarden

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"github.com/harpo-sh/harpo/internal/provider"
)

// TypeName is the provider type string used in harpo.yml.
const TypeName = "bitwarden-password-manager"

// Provider adapts the `bw` CLI to the provider.Provider interface.
type Provider struct {
	id  string
	bin string // path/name of the bw binary
}

// New returns a Bitwarden provider with the given config id.
func New(id string) *Provider {
	return &Provider{id: id, bin: "bw"}
}

func (p *Provider) ID() string   { return p.id }
func (p *Provider) Type() string { return TypeName }

// Capabilities for Bitwarden Password Manager. SupportsScopedAccess is false:
// Harpo applies only logical scope; an unlocked personal vault grants broad
// access in the user's context. See MVP spec §17.
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

// Status probes the bw CLI and reports vault state. It never inherits or
// exposes BW_SESSION as safe.
func (p *Provider) Status() (provider.Status, error) {
	st := provider.Status{Type: TypeName, SafeForInheritance: false}
	if _, err := exec.LookPath(p.bin); err != nil {
		st.CLIFound = false
		st.Vault = provider.StateNotFound
		st.Detail = "`bw` not found in PATH"
		return st, nil
	}
	st.CLIFound = true

	out, err := p.run("status")
	if err != nil {
		st.Vault = provider.StateUnknown
		st.Detail = "could not read `bw status`"
		return st, nil
	}
	// `bw status` returns JSON like {"status":"unlocked"|"locked"|"unauthenticated"}.
	switch {
	case strings.Contains(out, `"unlocked"`):
		st.Vault = provider.StateUnlocked
	case strings.Contains(out, `"locked"`):
		st.Vault = provider.StateLocked
	case strings.Contains(out, `"unauthenticated"`):
		st.Vault = provider.StateUnauthed
	default:
		st.Vault = provider.StateUnknown
	}
	return st, nil
}

// Resolve returns the secret value for the given ref/field. The value is
// sensitive and must never be logged or printed by callers.
//
// MVP-initial implementation: uses `bw get <field> <ref>`. Ref resolution will
// be hardened (item id lookup, disambiguation) in a follow-up — see TODO.
func (p *Provider) Resolve(ref provider.Ref) (string, error) {
	field := ref.Field
	if field == "" {
		field = "password"
	}
	// TODO(mvp): support fields beyond bw's `get` verbs and disambiguate
	// multiple matches for the search string instead of failing.
	out, err := p.run("get", field, ref.Ref)
	if err != nil {
		return "", err
	}
	value := strings.TrimRight(out, "\r\n")
	if value == "" {
		return "", errors.New("bitwarden returned an empty value for ref")
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

// run executes `bw <args...>` and returns trimmed stdout. stderr is captured
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
		return "", errors.New("bw " + args[0] + ": " + msg)
	}
	return stdout.String(), nil
}
