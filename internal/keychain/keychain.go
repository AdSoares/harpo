// Package keychain caches an unlocked provider session in the OS keychain
// (Windows Credential Manager, macOS Keychain, Linux Secret Service) with a
// TTL. It stores only the session token — never a master password — and treats
// an expired entry as absent. It is the single place that imports go-keyring.
package keychain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/zalando/go-keyring"
)

// service is the keychain service name under which Harpo stores sessions.
const service = "harpo"

// Entry is a cached provider session plus its Harpo-imposed expiry.
type Entry struct {
	Name      string    `json:"name"` // env var the provider CLI needs (e.g. BW_SESSION)
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Save stores the entry under the given provider id.
func Save(id string, e Entry) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return keyring.Set(service, id, string(data))
}

// Load returns the cached entry if present and unexpired. An expired or corrupt
// entry is deleted and reported as not found (ok == false).
func Load(id string) (Entry, bool, error) {
	s, err := keyring.Get(service, id)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return Entry{}, false, nil
		}
		return Entry{}, false, err
	}
	var e Entry
	if err := json.Unmarshal([]byte(s), &e); err != nil {
		_ = keyring.Delete(service, id)
		return Entry{}, false, nil
	}
	if !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt) {
		_ = keyring.Delete(service, id)
		return Entry{}, false, nil
	}
	return e, true, nil
}

// Delete removes the cached entry for the given provider id. Absence is not an
// error.
func Delete(id string) error {
	err := keyring.Delete(service, id)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
