// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// CompatStatus represents the compatibility between the
// installed gemara-mcp version and the user's selected Gemara
// schema version.
type CompatStatus int

const (
	// CompatOK means versions are compatible.
	CompatOK CompatStatus = iota
	// CompatMismatch means versions do not match.
	CompatMismatch
	// CompatUnknown means the server does not expose version
	// metadata.
	CompatUnknown
)

// CompatResult holds the result of a version compatibility
// check.
type CompatResult struct {
	// Status is the compatibility determination.
	Status CompatStatus
	// MCPVersion is the gemara-mcp server version.
	MCPVersion string
	// MCPSchemaVersion is the Gemara schema version the MCP
	// server was built against.
	MCPSchemaVersion string
	// MCPMode is the server's operating mode (advisory or
	// artifact), if reported.
	MCPMode string
	// SelectedVersion is the user's selected schema version.
	SelectedVersion string
	// Recommendation is an actionable message for the user
	// when versions are mismatched or unknown.
	Recommendation string
}

// VersionInfo represents the version metadata returned by the
// MCP server.
type VersionInfo struct {
	ServerVersion string `json:"server_version"`
	SchemaVersion string `json:"schema_version"`
	Mode          string `json:"mode,omitempty"`
}

// VersionFetcher abstracts how version info is retrieved from
// the MCP server for testing.
type VersionFetcher func(
	ctx context.Context,
	client *Client,
) (*VersionInfo, error)

// DefaultVersionFetcher queries the MCP server for version
// metadata via a tool call.
func DefaultVersionFetcher(
	ctx context.Context,
	client *Client,
) (*VersionInfo, error) {
	resp, err := client.callTool(
		ctx, "get_version", nil,
	)
	if err != nil {
		return nil, err
	}

	var info VersionInfo
	if err := json.Unmarshal(resp, &info); err != nil {
		return nil, fmt.Errorf(
			"parse version info: %w", err,
		)
	}
	return &info, nil
}

// CheckCompatibility verifies that the installed gemara-mcp
// version is compatible with the user's selected Gemara schema
// version.
func CheckCompatibility(
	ctx context.Context,
	fetcher VersionFetcher,
	client *Client,
	selectedVersion string,
) (*CompatResult, error) {
	info, err := fetcher(ctx, client)
	if err != nil {
		// Server does not expose version metadata.
		return &CompatResult{
			Status:          CompatUnknown,
			SelectedVersion: selectedVersion,
			Recommendation: "The installed gemara-mcp does " +
				"not expose version metadata. Update to " +
				"a version that supports version reporting " +
				"to enable compatibility verification.",
		}, nil
	}

	if info.SchemaVersion == "" {
		return &CompatResult{
			Status:          CompatUnknown,
			MCPVersion:      info.ServerVersion,
			SelectedVersion: selectedVersion,
			Recommendation: "The installed gemara-mcp does " +
				"not report its schema version. " +
				"Compatibility cannot be verified.",
		}, nil
	}

	if info.SchemaVersion == selectedVersion {
		return &CompatResult{
			Status:           CompatOK,
			MCPVersion:       info.ServerVersion,
			MCPSchemaVersion: info.SchemaVersion,
			SelectedVersion:  selectedVersion,
		}, nil
	}

	return &CompatResult{
		Status:           CompatMismatch,
		MCPVersion:       info.ServerVersion,
		MCPSchemaVersion: info.SchemaVersion,
		SelectedVersion:  selectedVersion,
		Recommendation: fmt.Sprintf(
			"The installed gemara-mcp (built against %s) "+
				"may produce inaccurate results for schema "+
				"%s. Either update gemara-mcp to a version "+
				"built against %s, or select schema version "+
				"%s to match your installed gemara-mcp.",
			info.SchemaVersion,
			selectedVersion,
			selectedVersion,
			info.SchemaVersion,
		),
	}, nil
}
