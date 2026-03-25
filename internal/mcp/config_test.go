// SPDX-License-Identifier: Apache-2.0

package mcp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hbraswelrh/journey/internal/consts"
	"github.com/hbraswelrh/journey/internal/mcp"
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
		"/home/user/.local/share/journey/gemara-mcp/"+
			"bin/gemara-mcp",
		consts.MCPModeArtifact,
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

	// Verify command[0] is the binary path.
	expectedCmd := "/home/user/.local/share/journey/" +
		"gemara-mcp/bin/gemara-mcp"
	gotPath := mcp.MCPBinaryPath(entry)
	if gotPath != expectedCmd {
		t.Fatalf(
			"expected command[0] %q, got %q",
			expectedCmd, gotPath,
		)
	}

	// Verify command contains serve, --mode, artifact.
	wantCmd := []string{
		expectedCmd, "serve", "--mode", "artifact",
	}
	if len(entry.Command) != len(wantCmd) {
		t.Fatalf(
			"expected command %v, got %v",
			wantCmd, entry.Command,
		)
	}
	for i, want := range wantCmd {
		if entry.Command[i] != want {
			t.Fatalf(
				"command[%d] = %q, want %q",
				i, entry.Command[i], want,
			)
		}
	}

	// Verify type is "local".
	if entry.Type != "local" {
		t.Fatalf(
			"expected type local, got %q",
			entry.Type,
		)
	}
}

func TestEnsureMCPEntry_PreservesExistingEntries(
	t *testing.T,
) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	// Write a config with an existing MCP entry using
	// the correct OpenCode format.
	existing := map[string]any{
		"mcp": map[string]any{
			"other-server": map[string]any{
				"type":    "local",
				"command": []string{"other-binary"},
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

	config, err := mcp.ReadOpenCodeConfig(path)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	mcp.EnsureMCPEntry(
		config,
		"/opt/gemara-mcp/bin/gemara-mcp",
		consts.MCPModeArtifact,
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

	if _, ok := readBack.MCP["other-server"]; !ok {
		t.Fatal("existing other-server entry was lost")
	}
	if _, ok := readBack.MCP[consts.MCPServerName]; !ok {
		t.Fatal("gemara-mcp entry was not added")
	}
}

// Verify JSON output matches the OpenCode MCP config format.
func TestMCPConfigOutputFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	config := &mcp.OpenCodeConfig{
		MCP: make(map[string]mcp.OpenCodeMCPEntry),
	}
	mcp.EnsureMCPEntry(
		config, "/usr/local/bin/gemara-mcp",
		consts.MCPModeArtifact,
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

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Should have "mcp" key (not "mcpServers").
	if _, ok := raw["mcp"]; !ok {
		t.Fatal("expected 'mcp' key in output")
	}

	// Parse the servers.
	var servers map[string]map[string]any
	if err := json.Unmarshal(
		raw["mcp"], &servers,
	); err != nil {
		t.Fatalf("parse servers: %v", err)
	}

	entry, ok := servers[consts.MCPServerName]
	if !ok {
		t.Fatal("expected gemara-mcp server entry")
	}

	// type should be "local".
	if entry["type"] != "local" {
		t.Errorf("type = %v, want local", entry["type"])
	}

	// command should be an array.
	cmd, ok := entry["command"].([]any)
	if !ok {
		t.Fatalf(
			"command should be array, got %T",
			entry["command"],
		)
	}
	wantCmd := []string{
		"/usr/local/bin/gemara-mcp",
		"serve", "--mode", "artifact",
	}
	if len(cmd) != len(wantCmd) {
		t.Fatalf(
			"expected %d command elements, got %v",
			len(wantCmd), cmd,
		)
	}
	for i, want := range wantCmd {
		if cmd[i] != want {
			t.Errorf(
				"command[%d] = %v, want %q",
				i, cmd[i], want,
			)
		}
	}
}

func TestEnsureMCPEntry_DefaultMode(t *testing.T) {
	config := &mcp.OpenCodeConfig{
		MCP: make(map[string]mcp.OpenCodeMCPEntry),
	}
	mcp.EnsureMCPEntry(
		config, "/usr/local/bin/gemara-mcp", "",
	)

	entry := config.MCP[consts.MCPServerName]
	// command = [binary, serve, --mode, artifact]
	if len(entry.Command) != 4 {
		t.Fatalf(
			"expected 4 command elements, got %v",
			entry.Command,
		)
	}
	if entry.Command[3] != consts.MCPModeArtifact {
		t.Fatalf(
			"expected default mode %q, got %q",
			consts.MCPModeArtifact, entry.Command[3],
		)
	}
}

func TestEnsureMCPEntry_AdvisoryMode(t *testing.T) {
	config := &mcp.OpenCodeConfig{
		MCP: make(map[string]mcp.OpenCodeMCPEntry),
	}
	mcp.EnsureMCPEntry(
		config, "/usr/local/bin/gemara-mcp",
		consts.MCPModeAdvisory,
	)

	entry := config.MCP[consts.MCPServerName]
	wantCmd := []string{
		"/usr/local/bin/gemara-mcp",
		"serve", "--mode", "advisory",
	}
	if len(entry.Command) != len(wantCmd) {
		t.Fatalf(
			"expected command %v, got %v",
			wantCmd, entry.Command,
		)
	}
	for i, want := range wantCmd {
		if entry.Command[i] != want {
			t.Fatalf(
				"command[%d] = %q, want %q",
				i, entry.Command[i], want,
			)
		}
	}
}

func TestParseMCPMode_Found(t *testing.T) {
	entry := mcp.OpenCodeMCPEntry{
		Type: "local",
		Command: []string{
			"/usr/local/bin/gemara-mcp",
			"serve", "--mode", "advisory",
		},
	}
	mode := mcp.ParseMCPMode(entry)
	if mode != consts.MCPModeAdvisory {
		t.Fatalf(
			"expected %q, got %q",
			consts.MCPModeAdvisory, mode,
		)
	}
}

func TestParseMCPMode_NotFound(t *testing.T) {
	entry := mcp.OpenCodeMCPEntry{
		Type: "local",
		Command: []string{
			"/usr/local/bin/gemara-mcp",
			"serve",
		},
	}
	mode := mcp.ParseMCPMode(entry)
	if mode != consts.MCPModeDefault {
		t.Fatalf(
			"expected default %q, got %q",
			consts.MCPModeDefault, mode,
		)
	}
}

func TestMCPBinaryPath(t *testing.T) {
	entry := mcp.OpenCodeMCPEntry{
		Type: "local",
		Command: []string{
			"/opt/gemara-mcp/bin/gemara-mcp",
			"serve", "--mode", "artifact",
		},
	}
	path := mcp.MCPBinaryPath(entry)
	if path != "/opt/gemara-mcp/bin/gemara-mcp" {
		t.Fatalf("expected binary path, got %q", path)
	}
}

func TestMCPBinaryPath_Empty(t *testing.T) {
	entry := mcp.OpenCodeMCPEntry{
		Type:    "local",
		Command: []string{},
	}
	path := mcp.MCPBinaryPath(entry)
	if path != "" {
		t.Fatalf("expected empty, got %q", path)
	}
}
