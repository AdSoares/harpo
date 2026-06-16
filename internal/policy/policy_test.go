package policy

import (
	"testing"
	"time"

	"github.com/harpo-sh/harpo/internal/config"
)

func strictCfg() *config.Config { return config.Default("demo", config.ModeStrict, "claude") }

func TestRevealBlockedInStrict(t *testing.T) {
	eng := New(strictCfg())
	if err := eng.CheckReveal(); err == nil {
		t.Fatal("reveal must be blocked in strict mode")
	}
}

func TestDotenvBlockedInStrict(t *testing.T) {
	eng := New(strictCfg())
	if err := eng.CheckDotenv(); err == nil {
		t.Fatal(".env must be blocked in strict mode")
	}
}

func TestTTLRequiredInStrict(t *testing.T) {
	eng := New(strictCfg())
	if _, err := eng.CheckTTL(0); err == nil {
		t.Fatal("a TTL must be required in strict mode")
	}
}

func TestTTLExceedsMax(t *testing.T) {
	eng := New(strictCfg())
	if _, err := eng.CheckTTL(24 * time.Hour); err == nil {
		t.Fatal("TTL above max_ttl must be rejected")
	}
}

func TestTTLWithinBounds(t *testing.T) {
	eng := New(strictCfg())
	got, err := eng.CheckTTL(2 * time.Hour)
	if err != nil {
		t.Fatalf("valid TTL rejected: %v", err)
	}
	if got != 2*time.Hour {
		t.Fatalf("effective TTL = %v, want 2h", got)
	}
}

func TestSuspiciousAliasWarning(t *testing.T) {
	if SuspiciousAliasWarning("aws.prod.root") == "" {
		t.Fatal("expected warning for production-like alias")
	}
	if SuspiciousAliasWarning("gitlab.ad.read") != "" {
		t.Fatal("did not expect warning for benign alias")
	}
}
