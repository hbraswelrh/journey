// SPDX-License-Identifier: Apache-2.0

// Package session manages Pac-Man session state, tracking MCP
// connection status, schema version selection, fallback mode,
// and available tools.
package session

import (
	"sync"
)

// MCPStatus represents the MCP server connection state for
// the session.
type MCPStatus int

const (
	// MCPNotInstalled means the MCP server is not installed.
	MCPNotInstalled MCPStatus = iota
	// MCPConnected means the MCP server is connected and
	// available.
	MCPConnected
	// MCPDisconnected means the MCP server was connected but
	// lost connection mid-session.
	MCPDisconnected
)

// AvailableTools tracks which MCP tools are accessible in the
// current session.
type AvailableTools struct {
	GetLexicon       bool
	ValidateArtifact bool
	GetSchemaDocs    bool
}

// Session holds the state for a Pac-Man session.
type Session struct {
	mu sync.RWMutex

	// MCPStatus tracks the MCP server connection state.
	mcpStatus MCPStatus

	// SchemaVersion is the user's selected Gemara schema
	// version for this session.
	SchemaVersion string

	// FallbackMode is true when the session is operating
	// without the MCP server.
	FallbackMode bool

	// Tools tracks which MCP tools are available.
	Tools AvailableTools

	// DegradedCapabilities lists capabilities that are
	// unavailable or degraded in the current session.
	DegradedCapabilities []string

	// RoleName is the identified role name for this
	// session.
	RoleName string

	// ActivityKeywords are the extracted activity keywords
	// from the user's description.
	ActivityKeywords []string

	// ResolvedLayers are the Gemara layer numbers resolved
	// from role + activity probing.
	ResolvedLayers []int

	// LearningPathSteps is the number of steps in the
	// generated learning path.
	LearningPathSteps int

	// ContentBlocksCount is the number of content blocks
	// extracted from tutorials.
	ContentBlocksCount int
}

// SetRoleProfile stores role discovery results in the session.
func (s *Session) SetRoleProfile(
	roleName string,
	keywords []string,
	layers []int,
	pathSteps int,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.RoleName = roleName
	s.ActivityKeywords = keywords
	s.ResolvedLayers = layers
	s.LearningPathSteps = pathSteps
}

// GetRoleName returns the session's identified role name.
func (s *Session) GetRoleName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RoleName
}

// SetContentBlocks stores the content blocks count in the
// session.
func (s *Session) SetContentBlocks(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ContentBlocksCount = count
}

// GetContentBlocksCount returns the number of extracted
// content blocks.
func (s *Session) GetContentBlocksCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ContentBlocksCount
}

// NewSessionWithMCP creates a session with an active MCP
// connection. All three tools are marked as available.
func NewSessionWithMCP(schemaVersion string) *Session {
	return &Session{
		mcpStatus:     MCPConnected,
		SchemaVersion: schemaVersion,
		FallbackMode:  false,
		Tools: AvailableTools{
			GetLexicon:       true,
			ValidateArtifact: true,
			GetSchemaDocs:    true,
		},
	}
}

// NewSessionWithoutMCP creates a session in fallback mode.
// Local alternatives are used for all capabilities.
func NewSessionWithoutMCP(schemaVersion string) *Session {
	return &Session{
		mcpStatus:     MCPNotInstalled,
		SchemaVersion: schemaVersion,
		FallbackMode:  true,
		Tools: AvailableTools{
			GetLexicon:       false,
			ValidateArtifact: false,
			GetSchemaDocs:    false,
		},
		DegradedCapabilities: []string{
			"Lexicon lookups use bundled data (may not " +
				"reflect latest upstream terms)",
			"Schema validation uses local cue vet " +
				"(requires CUE CLI installed)",
			"Schema documentation limited to locally " +
				"cached content",
		},
	}
}

// GetMCPStatus returns the current MCP connection status.
func (s *Session) GetMCPStatus() MCPStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mcpStatus
}

// HandleDisconnection transitions the session to fallback mode
// when the MCP server becomes unavailable mid-session. This
// method is safe for concurrent use and does not discard any
// in-progress state.
func (s *Session) HandleDisconnection() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mcpStatus = MCPDisconnected
	s.FallbackMode = true
	s.Tools = AvailableTools{
		GetLexicon:       false,
		ValidateArtifact: false,
		GetSchemaDocs:    false,
	}
	s.DegradedCapabilities = []string{
		"MCP server disconnected; using local fallbacks",
		"Lexicon lookups use bundled data",
		"Schema validation uses local cue vet",
		"Schema documentation limited to cached content",
	}
}

// HandleReconnection restores the session to full MCP
// capabilities when the server becomes available again.
func (s *Session) HandleReconnection() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mcpStatus = MCPConnected
	s.FallbackMode = false
	s.Tools = AvailableTools{
		GetLexicon:       true,
		ValidateArtifact: true,
		GetSchemaDocs:    true,
	}
	s.DegradedCapabilities = nil
}

// IsFallback returns whether the session is currently operating
// in fallback mode.
func (s *Session) IsFallback() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.FallbackMode
}
