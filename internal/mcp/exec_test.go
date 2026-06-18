package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/harpo-sh/harpo/internal/config"
)

// execServer builds a server with an allowlist, a profile authorizing "tok",
// and a stub resolver that returns a known secret value.
func execServer(t *testing.T, allowlist []string) *Server {
	t.Helper()
	cfg := config.Default("demo", config.ModeBalanced, "claude")
	cfg.Policies.Proxy.ExecAllowlist = allowlist
	cfg.Secrets["tok"] = config.Secret{Provider: "bw", DefaultEnv: "TOK"}
	cfg.Profiles["dev"] = config.Profile{
		Agent:   "claude",
		Secrets: []config.ProfileSecret{{Secret: "tok", Env: "TOK"}},
	}
	dir := t.TempDir()
	resolve := func(alias string) (string, error) { return "super-secret-value", nil }
	return New(cfg, "dev", dir, dir, "test", resolve)
}

func TestExecDeniesShellWrapper(t *testing.T) {
	s := execServer(t, []string{"bash"}) // even if mistakenly allowlisted
	_, err := s.execData(context.Background(), execInput{Command: "bash", Args: []string{"-c", "echo hi"}})
	if err == nil {
		t.Fatal("a shell interpreter must be denied")
	}
}

func TestExecDeniesNonAllowlisted(t *testing.T) {
	s := execServer(t, []string{"gh"})
	_, err := s.execData(context.Background(), execInput{Command: "curl"})
	if err == nil {
		t.Fatal("a non-allowlisted command must be denied")
	}
}

func TestExecDeniesUnauthorizedAlias(t *testing.T) {
	s := execServer(t, []string{"go"})
	_, err := s.execData(context.Background(), execInput{
		Command: "go", Args: []string{"version"},
		With: []execWith{{Alias: "not-in-profile", Env: "X"}},
	})
	if err == nil {
		t.Fatal("an alias not authorized by the profile must be denied")
	}
}

func TestExecRunsAllowlistedCommand(t *testing.T) {
	// `go` is always present during `go test`.
	s := execServer(t, []string{"go"})
	out, err := s.execData(context.Background(), execInput{
		Command: "go", Args: []string{"version"},
		With: []execWith{{Alias: "tok", Env: "TOK"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", out.ExitCode, out.Stderr)
	}
	if !strings.Contains(out.Stdout, "go version") {
		t.Fatalf("stdout missing 'go version': %q", out.Stdout)
	}
	// The injected secret value must never appear in the returned output.
	if strings.Contains(out.Stdout, "super-secret-value") || strings.Contains(out.Stderr, "super-secret-value") {
		t.Fatal("secret value leaked into brokered exec output")
	}
}

func TestCommandBase(t *testing.T) {
	cases := map[string]string{
		"gh":            "gh",
		"GH.EXE":        "gh",
		`C:\bin\gh.exe`: "gh",
		"/usr/bin/glab": "glab",
	}
	for in, want := range cases {
		if got := commandBase(in); got != want {
			t.Fatalf("commandBase(%q) = %q, want %q", in, got, want)
		}
	}
}
