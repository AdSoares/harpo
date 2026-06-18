package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUpsertMCPServerCreatesAndMerges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mcp.json")

	// First call creates the file with the harpo server.
	created, err := upsertMCPServer(path, "harpo", mcpServerEntry{Command: "harpo", Args: []string{"mcp", "--profile", "dev"}})
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected the file to be created on first call")
	}

	// Simulate a pre-existing unrelated server, then upsert harpo again.
	var root map[string]map[string]map[string]any
	data, _ := os.ReadFile(path)
	_ = json.Unmarshal(data, &root)
	root["mcpServers"]["other"] = map[string]any{"command": "other-server"}
	merged, _ := json.Marshal(root)
	_ = os.WriteFile(path, merged, 0o644)

	created, err = upsertMCPServer(path, "harpo", mcpServerEntry{Command: "harpo", Args: []string{"mcp", "--profile", "prod"}})
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("expected created=false when the file already exists")
	}

	// Both servers must be present; harpo updated to the new profile.
	out, _ := os.ReadFile(path)
	var got struct {
		MCPServers map[string]mcpServerEntry `json:"mcpServers"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got.MCPServers["other"]; !ok {
		t.Fatal("the pre-existing 'other' server must be preserved")
	}
	h, ok := got.MCPServers["harpo"]
	if !ok {
		t.Fatal("harpo server missing")
	}
	if len(h.Args) != 3 || h.Args[2] != "prod" {
		t.Fatalf("harpo server not updated to new profile: %+v", h.Args)
	}
}
