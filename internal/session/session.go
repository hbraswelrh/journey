// SPDX-License-Identifier: Apache-2.0

// Package session manages Gemara User Journey session state, tracking MCP
// connection status, schema version selection, fallback mode,
// and available capabilities (tools, resources, prompts).
package session

import (
	"sync"

	"github.com/hbraswelrh/journey/internal/consts"
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

// AvailableCapabilities tracks which MCP capabilities are
// accessible in the current session, organized by MCP
// protocol category.
type AvailableCapabilities struct {
	// Tools tracks callable MCP tools.
	Tools AvailableToolSet
	// Resources tracks readable MCP resources.
	Resources AvailableResourceSet
	// Prompts tracks available MCP prompts (wizards).
	Prompts AvailablePromptSet
}

// AvailableToolSet tracks which MCP tools are accessible.
type AvailableToolSet struct {
	ValidateArtifact bool
}

// AvailableResourceSet tracks which MCP resources are
// accessible.
type AvailableResourceSet struct {
	Lexicon           bool
	SchemaDefinitions bool
}

// AvailablePromptSet tracks which MCP prompts are
// accessible. Prompts are only available in artifact mode.
type AvailablePromptSet struct {
	ThreatAssessment bool
	ControlCatalog   bool
}

// Session holds the state for a Gemara User Journey session.
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

	// ServerMode is the MCP server operating mode
	// ("advisory" or "artifact"). Empty when MCP is not
	// installed.
	ServerMode string

	// Capabilities tracks which MCP capabilities are
	// available, organized by protocol category.
	Capabilities AvailableCapabilities

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

	// RecommendedArtifacts is the number of artifact types
	// recommended for this user based on their resolved
	// layers.
	RecommendedArtifacts int

	// ContentBlocksCount is the number of content blocks
	// extracted from tutorials.
	ContentBlocksCount int

	// TeamName is the name of the configured team, if any.
	TeamName string

	// TeamMemberCount is the number of members in the
	// configured team.
	TeamMemberCount int

	// AuthoringArtifactType is the artifact type being
	// authored, if any.
	AuthoringArtifactType string

	// AuthoringProgress tracks the authoring progress
	// (e.g., "2/4 steps").
	AuthoringProgress string

	// CompletedTutorials tracks which tutorial titles
	// have been marked as complete.
	CompletedTutorials map[string]bool
}

// SetRoleProfile stores role discovery results in the session.
func (s *Session) SetRoleProfile(
	roleName string,
	keywords []string,
	layers []int,
	pathSteps int,
	recommendedArtifacts int,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.RoleName = roleName
	s.ActivityKeywords = keywords
	s.ResolvedLayers = layers
	s.LearningPathSteps = pathSteps
	s.RecommendedArtifacts = recommendedArtifacts
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
// connection. Capabilities are set based on the server mode.
func NewSessionWithMCP(
	schemaVersion string,
	mode string,
) *Session {
	if mode == "" {
		mode = consts.MCPModeDefault
	}
	caps := AvailableCapabilities{
		Tools: AvailableToolSet{
			ValidateArtifact: true,
		},
		Resources: AvailableResourceSet{
			Lexicon:           true,
			SchemaDefinitions: true,
		},
	}
	// Prompts only available in artifact mode.
	if mode == consts.MCPModeArtifact {
		caps.Prompts = AvailablePromptSet{
			ThreatAssessment: true,
			ControlCatalog:   true,
		}
	}
	return &Session{
		mcpStatus:          MCPConnected,
		SchemaVersion:      schemaVersion,
		FallbackMode:       false,
		ServerMode:         mode,
		Capabilities:       caps,
		CompletedTutorials: make(map[string]bool),
	}
}

// NewSessionWithoutMCP creates a session in fallback mode.
// Local alternatives are used for all capabilities.
func NewSessionWithoutMCP(schemaVersion string) *Session {
	return &Session{
		mcpStatus:          MCPNotInstalled,
		SchemaVersion:      schemaVersion,
		FallbackMode:       true,
		ServerMode:         "",
		Capabilities:       AvailableCapabilities{},
		CompletedTutorials: make(map[string]bool),
		DegradedCapabilities: []string{
			"Lexicon lookups use bundled data " +
				"(may not reflect latest upstream terms)",
			"Schema validation uses local cue vet " +
				"(requires CUE CLI installed)",
			"Schema documentation limited to " +
				"locally cached content",
			"Guided creation wizards unavailable " +
				"(requires MCP server in artifact mode)",
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
	s.Capabilities = AvailableCapabilities{}
	s.DegradedCapabilities = []string{
		"MCP server disconnected; using local " +
			"fallbacks",
		"Lexicon lookups use bundled data",
		"Schema validation uses local cue vet",
		"Schema documentation limited to cached " +
			"content",
		"Guided creation wizards unavailable",
	}
}

// HandlePartialFailure updates individual capability flags
// when a specific MCP resource or tool fails without a full
// disconnection. The capability parameter identifies which
// capability failed: "lexicon", "schema_definitions", or
// "validate_artifact".
func (s *Session) HandlePartialFailure(
	capability string,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var msg string
	switch capability {
	case "lexicon":
		s.Capabilities.Resources.Lexicon = false
		msg = "Lexicon lookups use bundled data " +
			"(MCP resource unavailable)"
	case "schema_definitions":
		s.Capabilities.Resources.SchemaDefinitions = false
		msg = "Schema documentation limited to " +
			"cached content (MCP resource unavailable)"
	case "validate_artifact":
		s.Capabilities.Tools.ValidateArtifact = false
		msg = "Schema validation uses local cue vet " +
			"(MCP tool unavailable)"
	default:
		return
	}
	s.DegradedCapabilities = append(
		s.DegradedCapabilities, msg,
	)
}

// HandleReconnection restores the session to full MCP
// capabilities when the server becomes available again.
func (s *Session) HandleReconnection() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mcpStatus = MCPConnected
	s.FallbackMode = false
	s.Capabilities = AvailableCapabilities{
		Tools: AvailableToolSet{
			ValidateArtifact: true,
		},
		Resources: AvailableResourceSet{
			Lexicon:           true,
			SchemaDefinitions: true,
		},
	}
	if s.ServerMode == consts.MCPModeArtifact {
		s.Capabilities.Prompts = AvailablePromptSet{
			ThreatAssessment: true,
			ControlCatalog:   true,
		}
	}
	s.DegradedCapabilities = nil
}

// SetTeamInfo stores team configuration info in the session.
func (s *Session) SetTeamInfo(name string, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TeamName = name
	s.TeamMemberCount = count
}

// GetTeamName returns the session's team name.
func (s *Session) GetTeamName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.TeamName
}

// GetTeamMemberCount returns the number of team members.
func (s *Session) GetTeamMemberCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.TeamMemberCount
}

// MarkTutorialComplete records a tutorial title as
// completed.
func (s *Session) MarkTutorialComplete(title string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CompletedTutorials == nil {
		s.CompletedTutorials = make(map[string]bool)
	}
	s.CompletedTutorials[title] = true
}

// IsTutorialComplete returns whether a tutorial has been
// marked as complete.
func (s *Session) IsTutorialComplete(
	title string,
) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CompletedTutorials[title]
}

// SetAuthoringState stores the current authoring state in
// the session.
func (s *Session) SetAuthoringState(
	artifactType string,
	progress string,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.AuthoringArtifactType = artifactType
	s.AuthoringProgress = progress
}

// GetAuthoringState returns the current authoring artifact
// type and progress.
func (s *Session) GetAuthoringState() (string, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.AuthoringArtifactType, s.AuthoringProgress
}

// IsFallback returns whether the session is currently operating
// in fallback mode.
func (s *Session) IsFallback() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.FallbackMode
}

// GetServerMode returns the MCP server's operating mode.
func (s *Session) GetServerMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ServerMode
}

// IsArtifactMode returns true if the MCP server is running
// in artifact mode (wizards available).
func (s *Session) IsArtifactMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ServerMode == consts.MCPModeArtifact
}

// HasPrompts returns true if MCP prompts (wizards) are
// available in the current session.
func (s *Session) HasPrompts() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Capabilities.Prompts.ThreatAssessment ||
		s.Capabilities.Prompts.ControlCatalog
}
