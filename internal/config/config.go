// Package config loads and persists harpo.yml, the versionable, secret-free
// project configuration. No secret VALUE is ever stored here — only aliases,
// vault references and policy. See MVP spec §11.1 and §16.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// FileName is the canonical config file name.
	FileName = "harpo.yml"
	// SchemaVersion is the current harpo.yml schema version.
	SchemaVersion = 1
	// Dir is the local, untracked working directory.
	Dir = ".harpo"
)

// Mode is a security mode. See MVP spec §8.
type Mode string

const (
	// ModeStrict is recommended for AI agents: TTL mandatory, no reveal,
	// no .env by default, never inherit BW_SESSION.
	ModeStrict Mode = "strict"
	// ModeBalanced is recommended for solo developers: .env and reveal
	// allowed with explicit confirmation.
	ModeBalanced Mode = "balanced"
)

// Config is the root of harpo.yml.
type Config struct {
	Version   int                 `yaml:"version"`
	Project   Project             `yaml:"project"`
	Mode      Mode                `yaml:"mode"`
	Providers map[string]Provider `yaml:"providers"`
	Secrets   map[string]Secret   `yaml:"secrets"`
	Profiles  map[string]Profile  `yaml:"profiles"`
	Policies  Policies            `yaml:"policies"`
}

// Project identifies the project and the paths a session may be bound to.
type Project struct {
	Name         string   `yaml:"name"`
	AllowedPaths []string `yaml:"allowed_paths"`
}

// Provider declares a configured vault provider (no credentials here).
type Provider struct {
	Type string `yaml:"type"`
}

// Secret maps a local alias to an item/field in a vault. The value is never
// stored — only the reference needed to resolve it at runtime.
type Secret struct {
	Provider   string   `yaml:"provider"`
	Ref        string   `yaml:"ref"`
	Field      string   `yaml:"field"`
	DefaultEnv string   `yaml:"default_env"`
	Tags       []string `yaml:"tags,omitempty"`
}

// ProfileSecret binds a secret alias to a destination env var within a profile.
type ProfileSecret struct {
	Secret string `yaml:"secret"`
	Env    string `yaml:"env"`
}

// Profile is a named, reusable session template.
type Profile struct {
	TTL     Duration        `yaml:"ttl"`
	Agent   string          `yaml:"agent"`
	Secrets []ProfileSecret `yaml:"secrets"`
}

// Policies are the local policy knobs enforced by the policy engine.
type Policies struct {
	AllowDotenv    bool        `yaml:"allow_dotenv"`
	AllowReveal    bool        `yaml:"allow_reveal"`
	ManageUnlock   bool        `yaml:"manage_unlock"`
	UnlockCache    string      `yaml:"unlock_cache,omitempty"`     // "keychain" | "none" (default none)
	UnlockCacheTTL Duration    `yaml:"unlock_cache_ttl,omitempty"` // effective TTL is capped by max_ttl
	MCP            MCPPolicy   `yaml:"mcp,omitempty"`
	Proxy          ProxyPolicy `yaml:"proxy,omitempty"`
	DefaultTTL     Duration    `yaml:"default_ttl"`
	MaxTTL         Duration    `yaml:"max_ttl"`
}

// MCPPolicy governs the optional MCP server. It is off by default.
type MCPPolicy struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

// ProxyPolicy governs brokered execution (the harpo_exec MCP tool). An empty
// allowlist denies every command, so brokered exec is opt-in per command.
type ProxyPolicy struct {
	ExecAllowlist []string `yaml:"exec_allowlist,omitempty"`
}

// Default returns a secure-by-default config for a new project.
func Default(projectName string, mode Mode, agent string) *Config {
	if mode == "" {
		mode = ModeStrict
	}
	return &Config{
		Version:   SchemaVersion,
		Project:   Project{Name: projectName, AllowedPaths: []string{"."}},
		Mode:      mode,
		Providers: map[string]Provider{},
		Secrets:   map[string]Secret{},
		Profiles:  map[string]Profile{},
		Policies: Policies{
			AllowDotenv: mode == ModeBalanced,
			AllowReveal: false,
			DefaultTTL:  Duration(2 * 60 * 60 * 1e9), // 2h
			MaxTTL:      Duration(8 * 60 * 60 * 1e9), // 8h
		},
	}
}

// Load reads and parses a harpo.yml file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &c, nil
}

// Save writes the config to disk as YAML.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Find walks up from startDir looking for harpo.yml, returning its full path.
func Find(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, FileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no %s found from %s upward", FileName, startDir)
		}
		dir = parent
	}
}
