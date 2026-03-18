// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// OpenCodeConfig represents the OpenCode configuration file
// structure, focused on the MCP server entries.
type OpenCodeConfig struct {
	Schema string                      `json:"$schema,omitempty"`
	MCP    map[string]OpenCodeMCPEntry `json:"mcp,omitempty"`
	// Extra preserves unknown top-level fields during
	// read-modify-write.
	Extra map[string]json.RawMessage `json:"-"`
}

// OpenCodeMCPEntry represents a single MCP server entry in
// the OpenCode configuration. The format follows the OpenCode
// MCP specification where command is an array containing the
// binary path and all arguments.
type OpenCodeMCPEntry struct {
	Type    string   `json:"type"`
	Command []string `json:"command"`
	Enabled *bool    `json:"enabled,omitempty"`
}

// ReadOpenCodeConfig reads and parses the OpenCode
// configuration file at the given path. If the file does not
// exist, it returns an empty config.
func ReadOpenCodeConfig(
	path string,
) (*OpenCodeConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &OpenCodeConfig{
			MCP: make(map[string]OpenCodeMCPEntry),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config OpenCodeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if config.MCP == nil {
		config.MCP = make(map[string]OpenCodeMCPEntry)
	}

	return &config, nil
}

// WriteOpenCodeConfig writes the OpenCode configuration to
// the given path with indented formatting.
func WriteOpenCodeConfig(
	path string,
	config *OpenCodeConfig,
) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	// Append newline per constitution (end-of-file rule).
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

// EnsureMCPEntry adds or updates the gemara-mcp entry in the
// OpenCode configuration for a source-built binary. The
// command array contains the binary path followed by the
// serve subcommand and mode flag.
func EnsureMCPEntry(
	config *OpenCodeConfig,
	binaryPath string,
	mode string,
) {
	if mode == "" {
		mode = consts.MCPModeDefault
	}
	enabled := true
	config.MCP[consts.MCPServerName] = OpenCodeMCPEntry{
		Type: "local",
		Command: []string{
			binaryPath,
			"serve", consts.MCPModeFlag, mode,
		},
		Enabled: &enabled,
	}
}

// ParseMCPMode extracts the server mode from an existing
// MCP config entry's command array. Returns the default
// mode if no --mode flag is found.
func ParseMCPMode(entry OpenCodeMCPEntry) string {
	for i, arg := range entry.Command {
		if arg == consts.MCPModeFlag &&
			i+1 < len(entry.Command) {
			return entry.Command[i+1]
		}
	}
	return consts.MCPModeDefault
}

// MCPBinaryPath extracts the binary path from an MCP
// config entry's command array. Returns empty string if
// the command array is empty.
func MCPBinaryPath(entry OpenCodeMCPEntry) string {
	if len(entry.Command) > 0 {
		return entry.Command[0]
	}
	return ""
}
