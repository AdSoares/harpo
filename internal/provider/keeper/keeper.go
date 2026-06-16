// Package keeper implements the Provider interface for Keeper Security via
// Keeper Commander, by shelling out to the locally installed `keeper` CLI.
//
// Security contract (mirrors the Bitwarden adapter; see MVP spec §10.2, §15):
//   - It never lists the vault on behalf of an agent.
//   - It never writes secret values to stdout/stderr or logs.
//   - The user's Keeper login (persistent-login config / master password) is
//     never advertised as safe to inherit; the runner strips Keeper credential
//     env vars from child processes.
//
// Keeper Commander is a full personal-vault CLI (analogous to Bitwarden
// Password Manager), so access is broad in the user's context and Harpo applies
// only logical scope (SupportsScopedAccess: false).
package keeper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/harpo-sh/harpo/internal/provider"
)

// TypeName is the provider type string used in harpo.yml.
const TypeName = "keeper-commander"

// Provider adapts the `keeper` (Commander) CLI to provider.Provider.
type Provider struct {
	id  string
	bin string // path/name of the keeper binary
}

// New returns a Keeper Commander provider with the given config id.
func New(id string) *Provider {
	return &Provider{id: id, bin: "keeper"}
}

func (p *Provider) ID() string   { return p.id }
func (p *Provider) Type() string { return TypeName }

// Capabilities for Keeper Commander. SupportsScopedAccess is false: it is a
// full personal vault, so a logged-in session has broad access in the user's
// context and Harpo applies only logical scope. (For fine-grained, scoped
// machine access, a future Keeper Secrets Manager adapter is the path.)
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

// Status probes the keeper CLI and reports login state via `keeper whoami`,
// which prints account info but no secrets. Keeper Commander has no lock/unlock
// concept, so a successful whoami maps to "unlocked" and a failure (typically
// not logged in) maps to "unauthenticated".
func (p *Provider) Status() (provider.Status, error) {
	st := provider.Status{Type: TypeName, SafeForInheritance: false}
	if _, err := exec.LookPath(p.bin); err != nil {
		st.CLIFound = false
		st.Vault = provider.StateNotFound
		st.Detail = "`keeper` not found in PATH"
		return st, nil
	}
	st.CLIFound = true

	if _, err := p.run("whoami"); err != nil {
		st.Vault = provider.StateUnauthed
		st.Detail = "not logged in (run `keeper login`, ideally with persistent login)"
		return st, nil
	}
	st.Vault = provider.StateUnlocked
	return st, nil
}

// keeperUIDRe matches a Keeper record UID: a 22-character base64url string
// (e.g. "rvwIBG_ban2VTH64OsnzLn"). This is intentionally different from the
// Bitwarden UUID shape.
var keeperUIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{22}$`)

func looksLikeKeeperUID(s string) bool { return keeperUIDRe.MatchString(s) }

// Resolve returns the secret value for the given ref/field. The value is
// sensitive and must never be logged or printed by callers.
//
// For the password field it uses `keeper find-password`, which accepts a record
// UID or path/title and returns just the password — no parsing required. For
// any other field it reads the full record as JSON (`keeper get <ref> --format
// json --unmask`) and extracts the requested field.
func (p *Provider) Resolve(ref provider.Ref) (string, error) {
	field := ref.Field
	if field == "" {
		field = "password"
	}

	if strings.EqualFold(field, "password") {
		out, err := p.run("find-password", ref.Ref)
		if err != nil {
			return "", err
		}
		value := strings.TrimRight(out, "\r\n")
		if value == "" {
			return "", errors.New("keeper returned an empty password for ref")
		}
		return value, nil
	}

	out, err := p.run("get", ref.Ref, "--format", "json", "--unmask")
	if err != nil {
		return "", err
	}
	value, err := extractField([]byte(out), field)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("keeper returned an empty value for field %q", field)
	}
	return value, nil
}

// keeperRecord is the subset of `keeper get --format json` output we read.
// Keeper typed records expose standard fields plus user-defined custom fields,
// each with a type, an optional label, and a value array.
type keeperRecord struct {
	Fields []keeperField `json:"fields"`
	Custom []keeperField `json:"custom"`
}

type keeperField struct {
	Type  string            `json:"type"`
	Label string            `json:"label"`
	Value []json.RawMessage `json:"value"`
}

// extractField finds a field by type or label (case-insensitive) across the
// record's standard and custom fields and returns its first string value. It is
// pure so the parsing can be unit-tested without the keeper CLI.
func extractField(data []byte, field string) (string, error) {
	var rec keeperRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return "", fmt.Errorf("parsing keeper record json: %w", err)
	}
	for _, f := range append(rec.Fields, rec.Custom...) {
		if strings.EqualFold(f.Type, field) || strings.EqualFold(f.Label, field) {
			if len(f.Value) == 0 {
				return "", nil
			}
			var s string
			if err := json.Unmarshal(f.Value[0], &s); err != nil {
				return "", fmt.Errorf("field %q has a non-string value; not supported", field)
			}
			return s, nil
		}
	}
	return "", fmt.Errorf("field %q not found in keeper record", field)
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

// run executes `keeper <args...>` and returns trimmed stdout. stderr is captured
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
		return "", errors.New("keeper " + args[0] + ": " + msg)
	}
	return stdout.String(), nil
}
