package mcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/harpo-sh/harpo/internal/audit"
	"github.com/harpo-sh/harpo/internal/redact"
	"github.com/harpo-sh/harpo/internal/runner"
)

// maxOutput caps each captured stream so a brokered command cannot return an
// unbounded tool result.
const maxOutput = 64 * 1024

// shellWrappers are command interpreters that are always denied for brokered
// exec: allowing them would let the agent run arbitrary code with the
// credential present in the environment.
var shellWrappers = map[string]bool{
	"bash": true, "sh": true, "zsh": true, "fish": true, "ksh": true, "dash": true,
	"cmd": true, "command": true, "powershell": true, "pwsh": true,
	"python": true, "python3": true, "node": true, "nodejs": true, "deno": true,
	"bun": true, "ruby": true, "perl": true, "php": true, "env": true, "osascript": true,
}

type execWith struct {
	Alias string `json:"alias" jsonschema:"a secret alias authorized by the active profile"`
	Env   string `json:"env" jsonschema:"the environment variable to bind the secret to"`
}

type execInput struct {
	Command string     `json:"command" jsonschema:"the command to run; must be in the exec allowlist and must not be a shell interpreter"`
	Args    []string   `json:"args,omitempty"`
	With    []execWith `json:"with,omitempty" jsonschema:"secrets to inject for this command only"`
}

type execOutput struct {
	ExitCode  int    `json:"exit_code"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	Truncated bool   `json:"truncated,omitempty"`
}

func (s *Server) exec(ctx context.Context, _ *sdk.CallToolRequest, in execInput) (*sdk.CallToolResult, execOutput, error) {
	out, err := s.execData(ctx, in)
	return nil, out, err
}

func (s *Server) execData(_ context.Context, in execInput) (execOutput, error) {
	base := commandBase(in.Command)

	// 1. Command validation.
	if shellWrappers[base] {
		s.auditDenied(in.Command, "shell interpreter not allowed")
		return execOutput{}, fmt.Errorf("command %q is a shell interpreter and is not allowed", in.Command)
	}
	if !inAllowlist(s.cfg.Policies.Proxy.ExecAllowlist, base) {
		s.auditDenied(in.Command, "command not in exec allowlist")
		return execOutput{}, fmt.Errorf("command %q is not in policies.proxy.exec_allowlist", in.Command)
	}

	// 2. Authorize and resolve the requested secrets (profile-scoped).
	inject := map[string]string{}
	var values []string
	for _, w := range in.With {
		if w.Alias == "" || w.Env == "" {
			return execOutput{}, errors.New("each `with` entry needs both alias and env")
		}
		if !s.profileAuthorizes(w.Alias) {
			s.auditDenied(in.Command, "alias not authorized by profile")
			return execOutput{}, fmt.Errorf("alias %q is not authorized by profile %q", w.Alias, s.profile)
		}
		if s.resolve == nil {
			return execOutput{}, errors.New("secret resolution is not available")
		}
		val, err := s.resolve(w.Alias)
		if err != nil {
			return execOutput{}, fmt.Errorf("resolving %q: %w", w.Alias, err)
		}
		inject[w.Env] = val
		values = append(values, val)
	}

	// 3. Run with the secret(s) injected, capturing redacted output.
	red := redact.New(values...)
	outCap := &cappedWriter{max: maxOutput}
	errCap := &cappedWriter{max: maxOutput}
	ow := red.NewWriter(outCap)
	ew := red.NewWriter(errCap)
	runErr := runner.RunWith(in.Command, in.Args, inject, nil, ow, ew)
	_ = ow.Close()
	_ = ew.Close()

	// 4. Audit (no values).
	for _, w := range in.With {
		_ = audit.NewLogger(s.harpoDir).Log(audit.Event{
			Event:       "mcp.exec",
			Project:     s.cfg.Project.Name,
			Profile:     s.profile,
			SecretAlias: w.Alias,
			Destination: "env:" + w.Env,
			Command:     base,
			Result:      "success",
		})
	}

	code := exitCode(runErr)
	stderr := errCap.buf.String()
	if code == -1 && runErr != nil {
		// Surface a start failure (e.g. command not found) to the agent, redacted.
		stderr = strings.TrimSpace(stderr + "\n" + red.Redact(runErr.Error()))
	}
	return execOutput{
		ExitCode:  code,
		Stdout:    outCap.buf.String(),
		Stderr:    stderr,
		Truncated: outCap.truncated || errCap.truncated,
	}, nil
}

func (s *Server) profileAuthorizes(alias string) bool {
	prof, ok := s.cfg.Profiles[s.profile]
	if !ok {
		return false
	}
	for _, ps := range prof.Secrets {
		if ps.Secret == alias {
			return true
		}
	}
	return false
}

func (s *Server) auditDenied(command, reason string) {
	_ = audit.NewLogger(s.harpoDir).Log(audit.Event{
		Event:   "mcp.tool.denied",
		Project: s.cfg.Project.Name,
		Profile: s.profile,
		Command: commandBase(command),
		Reason:  reason,
		Result:  "denied",
	})
}

// commandBase normalizes a command to its base name for allowlist matching,
// stripping any directory and a Windows executable extension.
func commandBase(command string) string {
	base := strings.ToLower(filepath.Base(command))
	for _, ext := range []string{".exe", ".cmd", ".bat", ".com"} {
		base = strings.TrimSuffix(base, ext)
	}
	return base
}

func inAllowlist(allowlist []string, base string) bool {
	for _, a := range allowlist {
		if strings.ToLower(a) == base {
			return true
		}
	}
	return false
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// cappedWriter buffers up to max bytes and flags truncation past that.
type cappedWriter struct {
	buf       bytes.Buffer
	max       int
	truncated bool
}

func (w *cappedWriter) Write(p []byte) (int, error) {
	remaining := w.max - w.buf.Len()
	if remaining <= 0 {
		w.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		w.buf.Write(p[:remaining])
		w.truncated = true
		return len(p), nil
	}
	return w.buf.Write(p)
}
