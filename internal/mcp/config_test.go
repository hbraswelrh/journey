// SPDX-License-Identifier: Apache-2.0

package mcp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/mcp"
)

func TestReadOpenCodeConfig_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	config, err := mcp.ReadOpenCodeConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.MCP == nil {
		t.Fatal("expected non-nil MCP map")
	}
	if len(config.MCP) != 0 {
		t.Fatalf(
			"expected empty MCP map, got %d entries",
			len(config.MCP),
		)
	}
}

func TestWriteAndReadOpenCodeConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	config := &mcp.OpenCodeConfig{
		MCP: make(map[string]mcp.OpenCodeMCPEntry),
	}

	mcp.EnsureMCPEntry(
		config,
		"/home/user/.local/share/pacman/gemara-mcp/"+
			"bin/gemara-mcp",
	)

	if err := mcp.WriteOpenCodeConfig(
		path, config,
	); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	readBack, err := mcp.ReadOpenCodeConfig(path)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	entry, ok := readBack.MCP[consts.MCPServerName]
	if !ok {
		t.Fatalf(
			"expected %q entry in MCP config",
			consts.MCPServerName,
		)
	}

	// Verify command is the absolute binary path.
	expectedCmd := "/home/user/.local/share/pacman/" +
		"gemara-mcp/bin/gemara-mcp"
	if entry.Command != expectedCmd {
		t.Fatalf(
			"expected command %q, got %q",
			expectedCmd, entry.Command,
		)
	}

	// Verify args includes "serve".
	if len(entry.Args) != 1 || entry.Args[0] != "serve" {
		t.Fatalf(
			"expected args [serve], got %v",
			entry.Args,
		)
	}
}

func TestEnsureMCPEntryPodman(t *testing.T) {
	config := &mcp.OpenCodeConfig{
		MCP: make(map[string]mcp.OpenCodeMCPEntry),
	}

	mcp.EnsureMCPEntryPodman(config, "docker")

	entry, ok := config.MCP[consts.MCPServerName]
	if !ok {
		t.Fatal("expected gemara-mcp entry")
	}
	if entry.Command != "docker" {
		t.Fatalf(
			"expected command docker, got %q",
			entry.Command,
		)
	}
	if len(entry.Args) < 4 {
		t.Fatalf(
			"expected at least 4 args, got %v",
			entry.Args,
		)
	}
	// Verify args: run --rm -i <image> serve
	if entry.Args[0] != "run" {
		t.Errorf("args[0] = %q, want run", entry.Args[0])
	}
	if entry.Args[1] != "--rm" {
		t.Errorf(
			"args[1] = %q, want --rm", entry.Args[1],
		)
	}
	if entry.Args[2] != "-i" {
		t.Errorf("args[2] = %q, want -i", entry.Args[2])
	}
	if entry.Args[len(entry.Args)-1] != "serve" {
		t.Errorf(
			"last arg = %q, want serve",
			entry.Args[len(entry.Args)-1],
		)
	}
}

func TestEnsureMCPEntry_PreservesExistingEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	// Write a config with an existing MCP entry.
	existing := map[string]any{
		"mcpServers": map[string]any{
			"other-server": map[string]any{
				"command": "other-binary",
				"args":    []string{"run"},
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	data = append(data, '\n')
	if err := os.WriteFile(
		path, data, 0o644,
	); err != nil {
		t.Fatalf("write setup: %v", err)
	}

	// Read, add gemara-mcp, write back.
	config, err := mcp.ReadOpenCodeConfig(path)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	mcp.EnsureMCPEntry(
		config,
		"/opt/gemara-mcp/bin/gemara-mcp",
	)

	if err := mcp.WriteOpenCodeConfig(
		path, config,
	); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	// Verify both entries exist.
	readBack, err := mcp.ReadOpenCodeConfig(path)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	if _, ok := readBack.MCP["other-server"]; !ok {
		t.Fatal("existing other-server entry was lost")
	}
	if _, ok := readBack.MCP[consts.MCPServerName]; !ok {
		t.Fatal("gemara-mcp entry was not added")
	}
}

// TestMCPConfigOutputFormat verifies the JSON output matches
// the expected MCP client format.
func TestMCPConfigOutputFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	config := &mcp.OpenCodeConfig{
		MCP: make(map[string]mcp.OpenCodeMCPEntry),
	}
	mcp.EnsureMCPEntry(
		config, "/usr/local/bin/gemara-mcp",
	)

	if err := mcp.WriteOpenCodeConfig(
		path, config,
	); err != nil {
		t.Fatalf("write error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	// Parse as generic JSON to verify structure.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Should have mcpServers key.
	if _, ok := raw["mcpServers"]; !ok {
		t.Fatal("expected mcpServers key in output")
	}

	// Parse the servers.
	var servers map[string]map[string]any
	if err := json.Unmarshal(
		raw["mcpServers"], &servers,
	); err != nil {
		t.Fatalf("parse servers: %v", err)
	}

	entry, ok := servers[consts.MCPServerName]
	if !ok {
		t.Fatal("expected gemara-mcp server entry")
	}

	// command should be a string, not an array.
	cmd, ok := entry["command"].(string)
	if !ok {
		t.Fatalf(
			"command should be string, got %T",
			entry["command"],
		)
	}
	if cmd != "/usr/local/bin/gemara-mcp" {
		t.Errorf("command = %q", cmd)
	}

	// args should be an array containing "serve".
	args, ok := entry["args"].([]any)
	if !ok {
		t.Fatalf(
			"args should be array, got %T",
			entry["args"],
		)
	}
	if len(args) != 1 || args[0] != "serve" {
		t.Errorf("args = %v, want [serve]", args)
	}
}
