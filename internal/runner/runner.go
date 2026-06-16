// Package runner spawns child processes (the AI agent or an arbitrary command)
// with a controlled environment: dangerous inherited variables — above all
// BW_SESSION — are stripped, and only authorized secrets are injected.
//
// This is the most security-critical component. See MVP spec §10.5 and §15.
package runner

import (
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// DangerousEnvVars are never passed to a child process started by Harpo. The
// list intentionally covers vault session tokens across common providers so
// that an agent can never inherit a live vault session. BW_SESSION is the
// non-negotiable case for the MVP (MVP spec §15.4).
var DangerousEnvVars = []string{
	"BW_SESSION",
	"BW_CLIENTID",
	"BW_CLIENTSECRET",
	"BW_PASSWORD",
	"OP_SESSION", // 1Password (prefix match handled below)
	"OP_SERVICE_ACCOUNT_TOKEN",
	"OP_CONNECT_TOKEN", // 1Password Connect server token
	"VAULT_TOKEN",      // HashiCorp Vault
	"KEEPER_PASSWORD",  // Keeper Commander master password, if provided via env
	"KSM_CONFIG",       // Keeper Secrets Manager machine config (base64 JSON)
	"KSM_TOKEN",        // Keeper Secrets Manager one-time/binding token
	"KSM_CLI_TOKEN",    // Keeper Secrets Manager CLI one-time token
}

// isDangerous reports whether an environment variable name (the part before '=')
// must be stripped. Matches exact names and the OP_SESSION_* family.
func isDangerous(name string) bool {
	for _, d := range DangerousEnvVars {
		if name == d {
			return true
		}
	}
	return strings.HasPrefix(name, "OP_SESSION_")
}

// BuildChildEnv returns the environment for a child process: it starts from
// base (typically os.Environ()), removes every dangerous variable, then
// applies the authorized injections. Injected variables always win over any
// inherited value of the same name.
//
// It is pure and deterministic to keep it easy to test against leaks.
func BuildChildEnv(base []string, inject map[string]string) []string {
	out := make([]string, 0, len(base)+len(inject))
	for _, kv := range base {
		name := kv
		if i := strings.IndexByte(kv, '='); i >= 0 {
			name = kv[:i]
		}
		if isDangerous(name) {
			continue
		}
		if _, overridden := inject[name]; overridden {
			continue // injected value replaces inherited one
		}
		out = append(out, kv)
	}
	keys := make([]string, 0, len(inject))
	for k := range inject {
		keys = append(keys, k)
	}
	sort.Strings(keys) // deterministic ordering
	for _, k := range keys {
		out = append(out, k+"="+inject[k])
	}
	return out
}

// RunWith executes command with args in a child process whose environment is
// built from the current environment minus dangerous variables, plus the
// injected secrets. The caller supplies stdio, which lets `harpo exec` route
// output through a redacting writer while `harpo run` uses the raw terminal.
func RunWith(command string, args []string, inject map[string]string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command(command, args...)
	cmd.Env = BuildChildEnv(os.Environ(), inject)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// Run executes command wired directly to the parent's stdio, so interactive
// agents (the `harpo run` path) work normally. Output is NOT redacted here —
// Harpo does not promise redaction inside interactive TUIs (MVP spec §10.8).
func Run(command string, args []string, inject map[string]string) error {
	return RunWith(command, args, inject, os.Stdin, os.Stdout, os.Stderr)
}
