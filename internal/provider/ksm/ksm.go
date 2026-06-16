// Package ksm implements the Provider interface for Keeper Secrets Manager
// (KSM) by shelling out to the locally installed `ksm` CLI.
//
// Unlike Keeper Commander (a full personal vault), KSM is a machine-identity
// secrets manager: an Application is bound to specific shared folders, so
// access is genuinely scoped (least privilege). This adapter therefore reports
// SupportsScopedAccess: true.
//
// Security contract (mirrors the other adapters; see MVP spec §10.2, §15):
//   - It never lists the vault on behalf of an agent.
//   - It never writes secret values to stdout/stderr or logs.
//   - The KSM config/token (KSM_CONFIG/KSM_TOKEN/KSM_CLI_TOKEN) is never
//     advertised as safe to inherit; the runner strips it from child processes.
package ksm

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
const TypeName = "keeper-secrets-manager"

// Provider adapts the `ksm` CLI to provider.Provider.
type Provider struct {
	id  string
	bin string // path/name of the ksm binary
}

// New returns a Keeper Secrets Manager provider with the given config id.
func New(id string) *Provider {
	return &Provider{id: id, bin: "ksm"}
}

func (p *Provider) ID() string   { return p.id }
func (p *Provider) Type() string { return TypeName }

// Capabilities for KSM. SupportsScopedAccess is true: a KSM Application is bound
// to specific shared folders, so the access Harpo brokers is genuinely scoped
// rather than the broad logical scope of a personal vault.
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

// Status probes the ksm CLI and reports readiness via `ksm profile list`, which
// shows configured profile names but no secrets. KSM has no lock/unlock
// concept: a configured, working profile maps to "unlocked"; the absence of one
// maps to "unauthenticated".
func (p *Provider) Status() (provider.Status, error) {
	st := provider.Status{Type: TypeName, SafeForInheritance: false}
	if _, err := exec.LookPath(p.bin); err != nil {
		st.CLIFound = false
		st.Vault = provider.StateNotFound
		st.Detail = "`ksm` not found in PATH"
		return st, nil
	}
	st.CLIFound = true

	if _, err := p.run("profile", "list"); err != nil {
		st.Vault = provider.StateUnauthed
		st.Detail = "no KSM profile configured (run `ksm profile init` with a one-time token)"
		return st, nil
	}
	st.Vault = provider.StateUnlocked
	return st, nil
}

// keeperUIDRe matches a Keeper record UID: a 22-character base64url string
// (e.g. "SNzjw8tM1HsXEzXERCJrNQ").
var keeperUIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{22}$`)

func looksLikeKeeperUID(s string) bool { return keeperUIDRe.MatchString(s) }

// Resolve returns the secret value for the given ref/field. The value is
// sensitive and must never be logged or printed by callers.
//
// It resolves the ref to a single record UID (a ref that already looks like a
// UID is used directly; otherwise it is matched against `ksm secret list
// --json` by title and disambiguated), then reads the field with Keeper
// notation (`ksm secret notation keeper://<UID>/field/<field>`), which returns
// the raw value — avoiding the masked/tabular output of `ksm secret get`.
func (p *Provider) Resolve(ref provider.Ref) (string, error) {
	field := ref.Field
	if field == "" {
		field = "password"
	}
	uid, err := p.resolveUID(ref.Ref)
	if err != nil {
		return "", err
	}
	notation := fmt.Sprintf("keeper://%s/field/%s", uid, field)
	out, err := p.run("secret", "notation", notation)
	if err != nil {
		return "", err
	}
	value := strings.TrimRight(out, "\r\n")
	if value == "" {
		return "", fmt.Errorf("keeper secrets manager returned an empty value for field %q", field)
	}
	return value, nil
}

// ksmRecord is the subset of `ksm secret list --json` output we need to
// disambiguate matches.
type ksmRecord struct {
	UID   string `json:"uid"`
	Title string `json:"title"`
}

// resolveUID turns a ref into a single record UID. A ref that already looks like
// a Keeper UID is used directly; otherwise it is treated as a title and
// disambiguated via pickRecord.
func (p *Provider) resolveUID(ref string) (string, error) {
	if looksLikeKeeperUID(ref) {
		return ref, nil
	}
	out, err := p.run("secret", "list", "--json")
	if err != nil {
		return "", err
	}
	var records []ksmRecord
	if err := json.Unmarshal([]byte(out), &records); err != nil {
		return "", fmt.Errorf("parsing ksm secret list json: %w", err)
	}
	rec, err := pickRecord(records, ref)
	if err != nil {
		return "", err
	}
	return rec.UID, nil
}

// pickRecord selects exactly one record for a ref. With multiple matches it
// prefers a unique exact title match; otherwise it fails with a clear,
// value-free error. It is pure so the rules can be tested without the ksm CLI.
func pickRecord(records []ksmRecord, ref string) (ksmRecord, error) {
	var matches []ksmRecord
	for _, r := range records {
		if strings.EqualFold(r.Title, ref) {
			matches = append(matches, r)
		}
	}
	switch len(matches) {
	case 0:
		return ksmRecord{}, fmt.Errorf("no KSM record matches title %q; use the record UID", ref)
	case 1:
		return matches[0], nil
	default:
		return ksmRecord{}, fmt.Errorf("title %q matches %d KSM records; use the record UID", ref, len(matches))
	}
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

// run executes `ksm <args...>` and returns trimmed stdout. stderr is captured
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
		return "", errors.New("ksm " + args[0] + ": " + msg)
	}
	return stdout.String(), nil
}
