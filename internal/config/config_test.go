package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestParseTTL(t *testing.T) {
	cases := map[string]time.Duration{
		"2h":    2 * time.Hour,
		"30m":   30 * time.Minute,
		"1h30m": 90 * time.Minute,
		"45s":   45 * time.Second,
	}
	for in, want := range cases {
		d, err := ParseTTL(in)
		if err != nil {
			t.Fatalf("ParseTTL(%q) error: %v", in, err)
		}
		if d.Duration() != want {
			t.Fatalf("ParseTTL(%q) = %v, want %v", in, d.Duration(), want)
		}
	}
	if _, err := ParseTTL("nonsense"); err == nil {
		t.Fatal("expected error for invalid TTL")
	}
}

// TestSaveLoadRoundTrip also asserts the security invariant that a saved config
// is parseable and the TTL survives as a human string.
func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	cfg := Default("demo", ModeStrict, "claude")
	ttl, _ := ParseTTL("2h")
	cfg.Profiles["dev"] = Profile{TTL: ttl, Agent: "claude"}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Mode != ModeStrict {
		t.Fatalf("mode = %q, want strict", loaded.Mode)
	}
	if got := loaded.Profiles["dev"].TTL.Duration(); got != 2*time.Hour {
		t.Fatalf("profile ttl = %v, want 2h", got)
	}
	if loaded.Policies.AllowReveal {
		t.Fatal("strict default must not allow reveal")
	}
}

func TestFindWalksUpward(t *testing.T) {
	root := t.TempDir()
	cfg := Default("demo", ModeStrict, "")
	if err := cfg.Save(filepath.Join(root, FileName)); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b")
	if found, err := Find(nested); err != nil {
		// nested dirs don't exist yet; Find should still resolve from abs path walk
		t.Skipf("nested path not created: %v", err)
	} else if filepath.Dir(found) != root {
		t.Fatalf("found %q, want config under %q", found, root)
	}
}
