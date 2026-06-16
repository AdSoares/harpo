package runner

import (
	"strings"
	"testing"
)

// TestBuildChildEnvStripsBWSession guards the non-negotiable invariant: a child
// process started by Harpo must never inherit BW_SESSION (MVP spec §15.4).
func TestBuildChildEnvStripsBWSession(t *testing.T) {
	base := []string{
		"PATH=/usr/bin",
		"BW_SESSION=super-secret-session-token",
		"HOME=/home/ad",
	}
	env := BuildChildEnv(base, map[string]string{"GITLAB_TOKEN": "glpat-xyz"})

	for _, kv := range env {
		if strings.HasPrefix(kv, "BW_SESSION=") {
			t.Fatalf("BW_SESSION leaked into child env: %q", kv)
		}
		if strings.Contains(kv, "super-secret-session-token") {
			t.Fatalf("BW_SESSION value leaked into child env: %q", kv)
		}
	}
	if !containsKV(env, "GITLAB_TOKEN=glpat-xyz") {
		t.Fatal("injected GITLAB_TOKEN missing from child env")
	}
	if !containsKV(env, "PATH=/usr/bin") {
		t.Fatal("benign PATH should be preserved")
	}
}

func TestBuildChildEnvStripsOtherVaultTokens(t *testing.T) {
	base := []string{
		"OP_SESSION_my=abc",
		"OP_SERVICE_ACCOUNT_TOKEN=ops_tok",
		"VAULT_TOKEN=hvs.tok",
		"KEEPER_PASSWORD=master-pw",
		"KSM_CONFIG=base64cfg",
		"KSM_TOKEN=onetime",
		"KEEP=ok",
	}
	env := BuildChildEnv(base, nil)
	for _, kv := range env {
		name := kv[:strings.IndexByte(kv, '=')]
		if isDangerous(name) {
			t.Fatalf("dangerous var %q leaked", name)
		}
	}
	if !containsKV(env, "KEEP=ok") {
		t.Fatal("benign var should be preserved")
	}
}

// TestInjectedValueOverridesInherited ensures an injected secret wins over any
// inherited variable of the same name (no stale value remains).
func TestInjectedValueOverridesInherited(t *testing.T) {
	base := []string{"GITLAB_TOKEN=old-value"}
	env := BuildChildEnv(base, map[string]string{"GITLAB_TOKEN": "new-value"})
	count := 0
	for _, kv := range env {
		if strings.HasPrefix(kv, "GITLAB_TOKEN=") {
			count++
			if kv != "GITLAB_TOKEN=new-value" {
				t.Fatalf("expected injected value to win, got %q", kv)
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one GITLAB_TOKEN entry, got %d", count)
	}
}

func containsKV(env []string, want string) bool {
	for _, kv := range env {
		if kv == want {
			return true
		}
	}
	return false
}
