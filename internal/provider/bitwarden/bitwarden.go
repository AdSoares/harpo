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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/harpo-sh/harpo/internal/provider"
)

// TypeName is the provider type string used in harpo.yml.
const TypeName = "bitwarden-password-manager"

// Provider adapts the `bw` CLI to the provider.Provider interface.
type Provider struct {
	id      string
	bin     string // path/name of the bw binary
	session string // managed BW_SESSION, set by Unlock; empty = use ambient env
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
		SupportsUnlock:         true,
	}
}

// Unlock unlocks the vault with the given master password and holds the
// resulting BW_SESSION in memory for this provider instance. The password is
// passed over stdin (never as an argument) and is not retained after the call;
// the session is never logged or passed to child processes.
func (p *Provider) Unlock(master string) (provider.Session, error) {
	cmd := exec.Command(p.bin, "unlock", "--raw")
	cmd.Stdin = strings.NewReader(master + "\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return provider.Session{}, errors.New("bw unlock: " + msg)
	}
	session := strings.TrimSpace(stdout.String())
	if session == "" {
		return provider.Session{}, errors.New("bw unlock returned an empty session")
	}
	p.session = session
	return provider.Session{Name: "BW_SESSION", Value: session}, nil
}

// envWithSession returns base with BW_SESSION forced to session (replacing any
// inherited value), so the managed session — not an ambient one — is used.
func envWithSession(base []string, session string) []string {
	out := make([]string, 0, len(base)+1)
	for _, kv := range base {
		if strings.HasPrefix(kv, "BW_SESSION=") {
			continue
		}
		out = append(out, kv)
	}
	return append(out, "BW_SESSION="+session)
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

// bwItem is the subset of a Bitwarden item we need to disambiguate matches.
type bwItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func looksLikeUUID(s string) bool { return uuidRe.MatchString(s) }

// Resolve returns the secret value for the given ref/field. The value is
// sensitive and must never be logged or printed by callers.
//
// It first resolves the ref to a single, unambiguous item id (so a search
// string that matches multiple items fails with a clear message instead of a
// raw `bw` error), then reads the requested field from that exact item.
func (p *Provider) Resolve(ref provider.Ref) (string, error) {
	field := ref.Field
	if field == "" {
		field = "password"
	}
	id, err := p.resolveItemID(ref.Ref)
	if err != nil {
		return "", err
	}
	out, err := p.run("get", field, id)
	if err != nil {
		return "", err
	}
	value := strings.TrimRight(out, "\r\n")
	if value == "" {
		return "", fmt.Errorf("bitwarden returned an empty value for field %q", field)
	}
	return value, nil
}

// resolveItemID turns a ref into a single item id. A ref that already looks
// like an item UUID is used directly; otherwise it is treated as a search
// string and disambiguated via pickItem.
func (p *Provider) resolveItemID(ref string) (string, error) {
	if looksLikeUUID(ref) {
		return ref, nil
	}
	out, err := p.run("list", "items", "--search", ref)
	if err != nil {
		return "", err
	}
	var items []bwItem
	if err := json.Unmarshal([]byte(out), &items); err != nil {
		return "", fmt.Errorf("parsing bitwarden item list: %w", err)
	}
	item, err := pickItem(items, ref)
	if err != nil {
		return "", err
	}
	return item.ID, nil
}

// pickItem selects exactly one item for a ref. With multiple matches it prefers
// a unique exact name match; otherwise it fails with a clear, value-free error.
// It is pure so the disambiguation rules can be tested without the bw CLI.
func pickItem(items []bwItem, ref string) (bwItem, error) {
	switch len(items) {
	case 0:
		return bwItem{}, fmt.Errorf("no vault item matches ref %q", ref)
	case 1:
		return items[0], nil
	default:
		var exact []bwItem
		for _, it := range items {
			if it.Name == ref {
				exact = append(exact, it)
			}
		}
		if len(exact) == 1 {
			return exact[0], nil
		}
		return bwItem{}, fmt.Errorf("ref %q matches %d vault items; use a more specific reference or the item id", ref, len(items))
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

// run executes `bw <args...>` and returns trimmed stdout. stderr is captured
// only for error context and is not surfaced to general output.
func (p *Provider) run(args ...string) (string, error) {
	cmd := exec.Command(p.bin, args...)
	if p.session != "" {
		cmd.Env = envWithSession(os.Environ(), p.session)
	}
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
