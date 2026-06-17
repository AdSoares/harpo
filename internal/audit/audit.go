// Package audit appends structured events to a local JSONL log. It records what
// was used, when, by which profile and project — but NEVER a secret value.
// See MVP spec §10.7 and §15.
package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// LogFileName is the audit log file inside the .harpo directory.
const LogFileName = "audit.log.jsonl"

// Event is a single audit record. It must never carry a secret value.
type Event struct {
	Time        string `json:"time"`
	Event       string `json:"event"`
	Profile     string `json:"profile,omitempty"`
	Agent       string `json:"agent,omitempty"`
	Project     string `json:"project,omitempty"`
	Provider    string `json:"provider,omitempty"`
	Cache       string `json:"cache,omitempty"`
	SecretAlias string `json:"secret_alias,omitempty"`
	Destination string `json:"destination,omitempty"`
	Mode        string `json:"mode,omitempty"`
	TTLSeconds  int    `json:"ttl_seconds,omitempty"`
	Result      string `json:"result,omitempty"`
}

// Logger appends events to a JSONL file under the project's .harpo directory.
type Logger struct {
	path string
}

// NewLogger returns a logger writing to <harpoDir>/audit.log.jsonl.
func NewLogger(harpoDir string) *Logger {
	return &Logger{path: filepath.Join(harpoDir, LogFileName)}
}

// Log appends an event, stamping the current time if unset.
func (l *Logger) Log(e Event) error {
	if e.Time == "" {
		e.Time = time.Now().Format(time.RFC3339)
	}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

// List reads all events from the log. Returns an empty slice if no log exists.
func (l *Logger) List() ([]Event, error) {
	f, err := os.Open(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var events []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue // skip malformed lines rather than fail the whole listing
		}
		events = append(events, e)
	}
	return events, sc.Err()
}
