// Package mcp serves Harpo's value-free tools to an AI agent over the Model
// Context Protocol (stdio). The tools return metadata and (later) brokered
// results — never raw secret values. See docs/specs/proxy-mcp-mode.md.
//
// M2.1 ships the read-only tools: session status, available secrets, and an
// audit tail. Brokered exec (harpo_exec) is added in M2.2.
package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/harpo-sh/harpo/internal/audit"
	"github.com/harpo-sh/harpo/internal/config"
	"github.com/harpo-sh/harpo/internal/session"
)

// Server wires Harpo's value-free tools to an MCP stdio server.
type Server struct {
	cfg      *config.Config
	profile  string
	root     string
	harpoDir string
	version  string
}

// New returns an MCP server bound to the given project/profile.
func New(cfg *config.Config, profile, root, harpoDir, version string) *Server {
	return &Server{cfg: cfg, profile: profile, root: root, harpoDir: harpoDir, version: version}
}

// Run registers the tools and serves over stdio until the client disconnects.
func (s *Server) Run(ctx context.Context) error {
	srv := sdk.NewServer(&sdk.Implementation{Name: "harpo", Version: s.version}, nil)
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "harpo_session_status",
		Description: "Show the current Harpo session for this project (profile, expiry, authorized aliases). Never returns secret values.",
	}, s.sessionStatus)
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "harpo_secret_available",
		Description: "List the secret aliases the active profile authorizes, with their destination env vars. Never returns secret values.",
	}, s.secretAvailable)
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "harpo_audit_tail",
		Description: "Return recent Harpo audit events. Audit records never contain secret values.",
	}, s.auditTail)
	return srv.Run(ctx, &sdk.StdioTransport{})
}

// --- harpo_session_status ---

type statusInput struct{}

type grantInfo struct {
	Alias       string `json:"alias"`
	Destination string `json:"destination"`
}

type statusOutput struct {
	Active           bool        `json:"active"`
	SessionID        string      `json:"session_id,omitempty"`
	Profile          string      `json:"profile,omitempty"`
	Agent            string      `json:"agent,omitempty"`
	Project          string      `json:"project,omitempty"`
	ExpiresInSeconds int         `json:"expires_in_seconds,omitempty"`
	Grants           []grantInfo `json:"grants,omitempty"`
}

func (s *Server) sessionStatus(_ context.Context, _ *sdk.CallToolRequest, _ statusInput) (*sdk.CallToolResult, statusOutput, error) {
	return nil, s.sessionStatusData(), nil
}

func (s *Server) sessionStatusData() statusOutput {
	cur, err := session.NewManager(s.harpoDir).Current(s.root)
	if err != nil {
		return statusOutput{Active: false}
	}
	out := statusOutput{
		Active:           true,
		SessionID:        cur.ID,
		Profile:          cur.Profile,
		Agent:            cur.Agent,
		Project:          s.cfg.Project.Name,
		ExpiresInSeconds: int(cur.Remaining().Seconds()),
	}
	for _, g := range cur.Grants {
		out.Grants = append(out.Grants, grantInfo{Alias: g.Alias, Destination: g.Destination})
	}
	return out
}

// --- harpo_secret_available ---

type availInput struct {
	Tag string `json:"tag,omitempty" jsonschema:"optional tag to filter the aliases"`
}

type secretInfo struct {
	Alias       string   `json:"alias"`
	Destination string   `json:"destination"`
	Tags        []string `json:"tags,omitempty"`
}

type availOutput struct {
	Secrets []secretInfo `json:"secrets"`
}

func (s *Server) secretAvailable(_ context.Context, _ *sdk.CallToolRequest, in availInput) (*sdk.CallToolResult, availOutput, error) {
	return nil, s.secretAvailableData(in.Tag), nil
}

func (s *Server) secretAvailableData(tag string) availOutput {
	out := availOutput{Secrets: []secretInfo{}}
	prof, ok := s.cfg.Profiles[s.profile]
	if !ok {
		return out
	}
	for _, ps := range prof.Secrets {
		sec := s.cfg.Secrets[ps.Secret]
		if tag != "" && !hasTag(sec.Tags, tag) {
			continue
		}
		env := ps.Env
		if env == "" {
			env = sec.DefaultEnv
		}
		out.Secrets = append(out.Secrets, secretInfo{
			Alias:       ps.Secret,
			Destination: "env:" + env,
			Tags:        sec.Tags,
		})
	}
	return out
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

// --- harpo_audit_tail ---

type auditInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"max number of most-recent events to return (0 = all)"`
}

type auditOutput struct {
	Events []audit.Event `json:"events"`
}

func (s *Server) auditTail(_ context.Context, _ *sdk.CallToolRequest, in auditInput) (*sdk.CallToolResult, auditOutput, error) {
	out, err := s.auditTailData(in.Limit)
	return nil, out, err
}

func (s *Server) auditTailData(limit int) (auditOutput, error) {
	events, err := audit.NewLogger(s.harpoDir).List()
	if err != nil {
		return auditOutput{}, err
	}
	if limit > 0 && limit < len(events) {
		events = events[len(events)-limit:]
	}
	if events == nil {
		events = []audit.Event{}
	}
	return auditOutput{Events: events}, nil
}
