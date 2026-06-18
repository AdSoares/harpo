package mcp

import (
	"testing"

	"github.com/harpo-sh/harpo/internal/audit"
	"github.com/harpo-sh/harpo/internal/config"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	cfg := config.Default("demo", config.ModeStrict, "claude")
	cfg.Secrets["gitlab.ad.read"] = config.Secret{Provider: "bw", DefaultEnv: "GITLAB_TOKEN", Tags: []string{"gitlab", "dev"}}
	cfg.Secrets["aws.dev"] = config.Secret{Provider: "bw", DefaultEnv: "AWS_ACCESS_KEY", Tags: []string{"aws"}}
	cfg.Profiles["dev"] = config.Profile{
		Agent: "claude",
		Secrets: []config.ProfileSecret{
			{Secret: "gitlab.ad.read", Env: "GITLAB_TOKEN"},
			{Secret: "aws.dev", Env: "AWS_ACCESS_KEY"},
		},
	}
	dir := t.TempDir()
	return New(cfg, "dev", dir, dir, "test")
}

func TestSecretAvailableData(t *testing.T) {
	s := testServer(t)

	all := s.secretAvailableData("")
	if len(all.Secrets) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(all.Secrets))
	}
	for _, sec := range all.Secrets {
		if sec.Destination == "" || sec.Alias == "" {
			t.Fatalf("alias/destination must be set: %+v", sec)
		}
	}

	// Tag filter.
	aws := s.secretAvailableData("aws")
	if len(aws.Secrets) != 1 || aws.Secrets[0].Alias != "aws.dev" {
		t.Fatalf("tag filter failed: %+v", aws.Secrets)
	}
}

func TestSecretAvailableNeverLeaksValues(t *testing.T) {
	// secretInfo has no value field by construction; assert the destination is
	// an env reference, not a value.
	s := testServer(t)
	for _, sec := range s.secretAvailableData("").Secrets {
		if got := sec.Destination; got[:4] != "env:" {
			t.Fatalf("destination should be an env reference, got %q", got)
		}
	}
}

func TestSessionStatusNoActive(t *testing.T) {
	s := testServer(t)
	if st := s.sessionStatusData(); st.Active {
		t.Fatal("expected no active session in a fresh dir")
	}
}

func TestAuditTailData(t *testing.T) {
	s := testServer(t)
	log := audit.NewLogger(s.harpoDir)
	for i := 0; i < 5; i++ {
		_ = log.Log(audit.Event{Event: "secret.injected", Result: "success"})
	}
	out, err := s.auditTailData(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Events) != 2 {
		t.Fatalf("expected 2 most-recent events, got %d", len(out.Events))
	}

	all, _ := s.auditTailData(0)
	if len(all.Events) != 5 {
		t.Fatalf("expected all 5 events, got %d", len(all.Events))
	}
}
