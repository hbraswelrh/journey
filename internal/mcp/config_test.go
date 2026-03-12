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
		Schema: "https://opencode.ai/config.json",
		MCP:    make(map[string]mcp.OpenCodeMCPEntry),
	}

	mcp.EnsureMCPEntry(config, "/usr/local/bin/gemara-mcp")

	if err := mcp.WriteOpenCodeConfig(path, config); err != nil {
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
	if entry.Type != "local" {
		t.Fatalf("expected type local, got %s", entry.Type)
	}
	if len(entry.Command) != 1 ||
		entry.Command[0] != "/usr/local/bin/gemara-mcp" {
		t.Fatalf(
			"expected command [/usr/local/bin/gemara-mcp], "+
				"got %v",
			entry.Command,
		)
	}
	if entry.Enabled == nil || !*entry.Enabled {
		t.Fatal("expected enabled to be true")
	}
}

func TestEnsureMCPEntry_PreservesExistingEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	// Write a config with an existing MCP entry.
	existing := map[string]any{
		"$schema": "https://opencode.ai/config.json",
		"mcp": map[string]any{
			"other-server": map[string]any{
				"type":    "remote",
				"url":     "https://example.com/mcp",
				"enabled": true,
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

	mcp.EnsureMCPEntry(config, "/opt/gemara-mcp/gemara-mcp")

	if err := mcp.WriteOpenCodeConfig(path, config); err != nil {
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
