package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harpo-sh/harpo/internal/audit"
	"github.com/harpo-sh/harpo/internal/config"
)

// TestHelperProcess is not a real test. When GO_WANT_HELPER=1 it prints the
// injected secret to stdout and stderr so the leak guard can verify the
// brokered-exec output is redacted. (Standard Go subprocess-test idiom.)
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER") != "1" {
		return
	}
	v := os.Getenv("LEAKVAR")
	os.Stdout.WriteString("stdout has " + v + " here\n")
	os.Stderr.WriteString("stderr has " + v + " here\n")
	os.Exit(0)
}

// TestBrokeredExecRedactsAndNeverLeaks runs a brokered command that prints the
// injected secret, and asserts the value never reaches the tool output or the
// audit log.
func TestBrokeredExecRedactsAndNeverLeaks(t *testing.T) {
	const secret = "SUPER-SECRET-VALUE-123"
	self := os.Args[0]
	base := commandBase(self)

	cfg := config.Default("demo", config.ModeBalanced, "claude")
	cfg.Policies.Proxy.ExecAllowlist = []string{base}
	cfg.Secrets["tok"] = config.Secret{Provider: "bw", DefaultEnv: "LEAKVAR"}
	cfg.Profiles["dev"] = config.Profile{Secrets: []config.ProfileSecret{{Secret: "tok", Env: "LEAKVAR"}}}
	dir := t.TempDir()
	s := New(cfg, "dev", dir, dir, "test", func(string) (string, error) { return secret, nil })

	t.Setenv("GO_WANT_HELPER", "1")
	out, err := s.execData(context.Background(), execInput{
		Command: self,
		Args:    []string{"-test.run=TestHelperProcess"},
		With:    []execWith{{Alias: "tok", Env: "LEAKVAR"}},
	})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}

	if strings.Contains(out.Stdout, secret) || strings.Contains(out.Stderr, secret) {
		t.Fatalf("secret leaked into tool output:\nstdout=%q\nstderr=%q", out.Stdout, out.Stderr)
	}
	if !strings.Contains(out.Stdout, "[redacted]") {
		t.Fatalf("expected the redaction marker in stdout, got %q", out.Stdout)
	}

	if raw, _ := os.ReadFile(filepath.Join(dir, audit.LogFileName)); strings.Contains(string(raw), secret) {
		t.Fatal("secret value leaked into the audit log")
	}
}
