// Package session creates and manages temporary grants. A session stores only
// metadata (who/what/where/when) — never secret values. See MVP spec §10.4
// and §16.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DirName is the sessions subdirectory inside .harpo.
const DirName = "sessions"

// Grant authorizes one secret alias to a destination for the session.
type Grant struct {
	Alias       string `json:"alias"`
	Destination string `json:"destination"` // e.g. "env:GITLAB_TOKEN"
}

// Session is a time-bound authorization. It never contains secret values.
type Session struct {
	ID          string    `json:"id"`
	Profile     string    `json:"profile"`
	Agent       string    `json:"agent"`
	ProjectPath string    `json:"project_path"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Grants      []Grant   `json:"grants"`
}

// Expired reports whether the session has passed its TTL.
func (s *Session) Expired() bool { return time.Now().After(s.ExpiresAt) }

// Remaining returns the time left before expiry (zero if expired).
func (s *Session) Remaining() time.Duration {
	d := time.Until(s.ExpiresAt)
	if d < 0 {
		return 0
	}
	return d
}

// Manager persists sessions as JSON files under <harpoDir>/sessions.
type Manager struct {
	dir string
}

// NewManager returns a session manager rooted at <harpoDir>/sessions.
func NewManager(harpoDir string) *Manager {
	return &Manager{dir: filepath.Join(harpoDir, DirName)}
}

func newID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "sess_" + hex.EncodeToString(b)
}

// Create builds and persists a new session bound to the given project path.
func (m *Manager) Create(profile, agent, projectPath string, ttl time.Duration, grants []Grant) (*Session, error) {
	now := time.Now()
	s := &Session{
		ID:          newID(),
		Profile:     profile,
		Agent:       agent,
		ProjectPath: projectPath,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
		Grants:      grants,
	}
	if err := m.save(s); err != nil {
		return nil, err
	}
	return s, nil
}

func (m *Manager) save(s *Session) error {
	if err := os.MkdirAll(m.dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(m.dir, s.ID+".json"), data, 0o600)
}

// List returns all persisted sessions.
func (m *Manager) List() ([]*Session, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*Session
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.dir, e.Name()))
		if err != nil {
			continue
		}
		var s Session
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		out = append(out, &s)
	}
	return out, nil
}

// Current returns the most recent non-expired session bound to projectPath.
func (m *Manager) Current(projectPath string) (*Session, error) {
	sessions, err := m.List()
	if err != nil {
		return nil, err
	}
	var current *Session
	for _, s := range sessions {
		if s.ProjectPath != projectPath || s.Expired() {
			continue
		}
		if current == nil || s.CreatedAt.After(current.CreatedAt) {
			current = s
		}
	}
	if current == nil {
		return nil, fmt.Errorf("no active session for %s", projectPath)
	}
	return current, nil
}

// Revoke deletes a session's metadata file by id.
func (m *Manager) Revoke(id string) error {
	return os.Remove(filepath.Join(m.dir, id+".json"))
}
