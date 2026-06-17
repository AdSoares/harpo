// Package provider defines the abstraction over secret vaults. The MVP ships a
// single adapter (Bitwarden Password Manager via the `bw` CLI). See MVP spec
// §10.2 and §17.
package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// VaultState describes whether a vault is reachable and unlocked.
type VaultState string

const (
	StateUnknown  VaultState = "unknown"
	StateUnlocked VaultState = "unlocked"
	StateLocked   VaultState = "locked"
	StateUnauthed VaultState = "unauthenticated"
	StateNotFound VaultState = "cli-not-found"
)

// Status is the result of probing a provider.
type Status struct {
	Type               string
	CLIFound           bool
	Vault              VaultState
	SafeForInheritance bool // whether the provider session is safe to inherit (always false for bw)
	Detail             string
}

// Capabilities advertises what a provider can and cannot enforce, so Harpo can
// warn when "security" is only logical local scoping. See MVP spec §17.
type Capabilities struct {
	CanList                bool
	CanReadByRef           bool
	SupportsScopedAccess   bool
	SupportsAudit          bool
	SupportsRotation       bool
	SupportsDynamicSecrets bool
	SupportsUnlock         bool
}

// Session is an unlocked provider session obtained via Unlock. The value is
// sensitive: never log it, never persist it in plaintext, never pass it to a
// child process.
type Session struct {
	Name      string    // env var the provider CLI needs (e.g. "BW_SESSION")
	Value     string    // the session token
	ExpiresAt time.Time // zero means unknown
}

// Unlocker is implemented by providers that have an interactive unlock step
// (e.g. a master password). Providers without one do not implement it.
type Unlocker interface {
	// Unlock consumes a master secret (read via a secure prompt) and returns the
	// session the provider uses for subsequent reads. Implementations must not
	// log the master secret or the session, persist them in plaintext, or pass
	// them to child processes.
	Unlock(master string) (Session, error)
}

// Ref locates a single secret value within a vault.
type Ref struct {
	Ref   string // provider-specific item reference / search string
	Field string // field within the item (e.g. "password")
}

// TestResult reports that a secret resolves, WITHOUT revealing its value.
// See MVP spec §12.6 and §15.
type TestResult struct {
	Resolved    bool
	Length      int
	Fingerprint string // e.g. "sha256:ab12...9f"
}

// Provider is the interface every vault adapter implements.
type Provider interface {
	ID() string
	Type() string
	Status() (Status, error)
	// Resolve returns the raw secret value. Callers MUST treat the result as
	// sensitive: never log it, never print it by default.
	Resolve(ref Ref) (string, error)
	Test(ref Ref) (TestResult, error)
	Capabilities() Capabilities
}

// Fingerprint returns a safe, partial SHA-256 fingerprint of a secret value
// for use in `secret test` output. It never exposes the value itself.
func Fingerprint(value string) string {
	sum := sha256.Sum256([]byte(value))
	hexed := hex.EncodeToString(sum[:])
	return "sha256:" + hexed[:4] + "..." + hexed[len(hexed)-2:]
}
