// SPDX-License-Identifier: Apache-2.0

package session_test

import (
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/session"
)

func TestNewSessionWithMCP_ArtifactMode(t *testing.T) {
	s := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)

	if s.GetMCPStatus() != session.MCPConnected {
		t.Fatal("expected MCPConnected")
	}
	if s.IsFallback() {
		t.Fatal("expected fallback to be false")
	}
	if s.GetServerMode() != consts.MCPModeArtifact {
		t.Fatalf(
			"expected mode %q, got %q",
			consts.MCPModeArtifact,
			s.GetServerMode(),
		)
	}

	// Tool available.
	if !s.Capabilities.Tools.ValidateArtifact {
		t.Fatal("expected ValidateArtifact available")
	}

	// Resources available.
	if !s.Capabilities.Resources.Lexicon {
		t.Fatal("expected Lexicon resource available")
	}
	if !s.Capabilities.Resources.SchemaDefinitions {
		t.Fatal(
			"expected SchemaDefinitions resource " +
				"available",
		)
	}

	// Prompts available in artifact mode.
	if !s.Capabilities.Prompts.ThreatAssessment {
		t.Fatal(
			"expected ThreatAssessment prompt available",
		)
	}
	if !s.Capabilities.Prompts.ControlCatalog {
		t.Fatal(
			"expected ControlCatalog prompt available",
		)
	}

	if s.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected version v0.20.0, got %s",
			s.SchemaVersion,
		)
	}
}

func TestNewSessionWithMCP_AdvisoryMode(t *testing.T) {
	s := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeAdvisory,
	)

	if s.GetServerMode() != consts.MCPModeAdvisory {
		t.Fatalf(
			"expected mode %q, got %q",
			consts.MCPModeAdvisory,
			s.GetServerMode(),
		)
	}

	// Tool and resources available in advisory mode.
	if !s.Capabilities.Tools.ValidateArtifact {
		t.Fatal("expected ValidateArtifact available")
	}
	if !s.Capabilities.Resources.Lexicon {
		t.Fatal("expected Lexicon resource available")
	}
	if !s.Capabilities.Resources.SchemaDefinitions {
		t.Fatal(
			"expected SchemaDefinitions resource " +
				"available",
		)
	}

	// Prompts NOT available in advisory mode.
	if s.Capabilities.Prompts.ThreatAssessment {
		t.Fatal(
			"expected ThreatAssessment prompt " +
				"unavailable in advisory mode",
		)
	}
	if s.Capabilities.Prompts.ControlCatalog {
		t.Fatal(
			"expected ControlCatalog prompt " +
				"unavailable in advisory mode",
		)
	}
	if s.IsArtifactMode() {
		t.Fatal(
			"expected IsArtifactMode false for " +
				"advisory",
		)
	}
	if s.HasPrompts() {
		t.Fatal(
			"expected HasPrompts false for advisory",
		)
	}
}

func TestNewSessionWithMCP_DefaultMode(t *testing.T) {
	s := session.NewSessionWithMCP("v0.20.0", "")

	// Empty string defaults to artifact mode.
	if s.GetServerMode() != consts.MCPModeArtifact {
		t.Fatalf(
			"expected default mode %q, got %q",
			consts.MCPModeArtifact,
			s.GetServerMode(),
		)
	}
	if !s.IsArtifactMode() {
		t.Fatal("expected IsArtifactMode true for default")
	}
	if !s.HasPrompts() {
		t.Fatal("expected HasPrompts true for default")
	}
}

func TestNewSessionWithoutMCP(t *testing.T) {
	s := session.NewSessionWithoutMCP("v0.20.0")

	if s.GetMCPStatus() != session.MCPNotInstalled {
		t.Fatal("expected MCPNotInstalled")
	}
	if !s.IsFallback() {
		t.Fatal("expected fallback to be true")
	}
	if s.GetServerMode() != "" {
		t.Fatalf(
			"expected empty mode, got %q",
			s.GetServerMode(),
		)
	}

	// All capabilities unavailable.
	if s.Capabilities.Tools.ValidateArtifact {
		t.Fatal("expected ValidateArtifact unavailable")
	}
	if s.Capabilities.Resources.Lexicon {
		t.Fatal("expected Lexicon unavailable")
	}
	if s.Capabilities.Resources.SchemaDefinitions {
		t.Fatal(
			"expected SchemaDefinitions unavailable",
		)
	}
	if s.Capabilities.Prompts.ThreatAssessment {
		t.Fatal(
			"expected ThreatAssessment unavailable",
		)
	}
	if len(s.DegradedCapabilities) == 0 {
		t.Fatal("expected degraded capabilities listed")
	}
}

func TestSession_HandleDisconnection(t *testing.T) {
	s := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)

	// Verify starts connected.
	if s.GetMCPStatus() != session.MCPConnected {
		t.Fatal("expected MCPConnected initially")
	}

	s.HandleDisconnection()

	if s.GetMCPStatus() != session.MCPDisconnected {
		t.Fatal(
			"expected MCPDisconnected after disconnect",
		)
	}
	if !s.IsFallback() {
		t.Fatal("expected fallback after disconnect")
	}
	if s.Capabilities.Tools.ValidateArtifact {
		t.Fatal(
			"expected ValidateArtifact unavailable " +
				"after disconnect",
		)
	}
	if s.Capabilities.Resources.Lexicon {
		t.Fatal(
			"expected Lexicon unavailable after " +
				"disconnect",
		)
	}
	if s.Capabilities.Prompts.ThreatAssessment {
		t.Fatal(
			"expected ThreatAssessment unavailable " +
				"after disconnect",
		)
	}
	// Schema version should be preserved (no data loss).
	if s.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"schema version lost: expected v0.20.0, "+
				"got %s", s.SchemaVersion,
		)
	}
}

func TestSession_HandleReconnection_ArtifactMode(
	t *testing.T,
) {
	s := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)
	s.HandleDisconnection()

	// Verify disconnected state.
	if !s.IsFallback() {
		t.Fatal("expected fallback after disconnect")
	}

	s.HandleReconnection()

	if s.GetMCPStatus() != session.MCPConnected {
		t.Fatal("expected MCPConnected after reconnect")
	}
	if s.IsFallback() {
		t.Fatal("expected fallback false after reconnect")
	}
	if !s.Capabilities.Tools.ValidateArtifact {
		t.Fatal(
			"expected ValidateArtifact available " +
				"after reconnect",
		)
	}
	if !s.Capabilities.Resources.Lexicon {
		t.Fatal(
			"expected Lexicon available after reconnect",
		)
	}
	if !s.Capabilities.Resources.SchemaDefinitions {
		t.Fatal(
			"expected SchemaDefinitions available " +
				"after reconnect",
		)
	}
	// Artifact mode: prompts restored.
	if !s.Capabilities.Prompts.ThreatAssessment {
		t.Fatal(
			"expected ThreatAssessment restored after " +
				"reconnect in artifact mode",
		)
	}
	if s.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"schema version lost: expected v0.20.0, "+
				"got %s", s.SchemaVersion,
		)
	}
}

func TestSession_HandleReconnection_AdvisoryMode(
	t *testing.T,
) {
	s := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeAdvisory,
	)
	s.HandleDisconnection()
	s.HandleReconnection()

	// Advisory mode: prompts stay false.
	if s.Capabilities.Prompts.ThreatAssessment {
		t.Fatal(
			"expected ThreatAssessment unavailable " +
				"after reconnect in advisory mode",
		)
	}
	if s.Capabilities.Prompts.ControlCatalog {
		t.Fatal(
			"expected ControlCatalog unavailable " +
				"after reconnect in advisory mode",
		)
	}
	// Tools and resources restored.
	if !s.Capabilities.Tools.ValidateArtifact {
		t.Fatal(
			"expected ValidateArtifact available " +
				"after reconnect",
		)
	}
	if !s.Capabilities.Resources.Lexicon {
		t.Fatal(
			"expected Lexicon available after reconnect",
		)
	}
}

// T254: Session stores role profile data.
func TestSessionSetRoleProfile(t *testing.T) {
	s := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)

	s.SetRoleProfile(
		"Security Engineer",
		[]string{"CI/CD", "SDLC"},
		[]int{2, 1},
		3,
		2,
	)

	if s.GetRoleName() != "Security Engineer" {
		t.Errorf(
			"expected 'Security Engineer', got %s",
			s.GetRoleName(),
		)
	}
	if len(s.ActivityKeywords) != 2 {
		t.Errorf(
			"expected 2 keywords, got %d",
			len(s.ActivityKeywords),
		)
	}
	if len(s.ResolvedLayers) != 2 {
		t.Errorf(
			"expected 2 layers, got %d",
			len(s.ResolvedLayers),
		)
	}
	if s.LearningPathSteps != 3 {
		t.Errorf(
			"expected 3 path steps, got %d",
			s.LearningPathSteps,
		)
	}
}

func TestSession_IsArtifactMode(t *testing.T) {
	artifact := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)
	if !artifact.IsArtifactMode() {
		t.Fatal("expected true for artifact mode")
	}

	advisory := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeAdvisory,
	)
	if advisory.IsArtifactMode() {
		t.Fatal("expected false for advisory mode")
	}

	noMCP := session.NewSessionWithoutMCP("v0.20.0")
	if noMCP.IsArtifactMode() {
		t.Fatal("expected false for no MCP")
	}
}

func TestSession_HandlePartialFailure_ResourceDown(
	t *testing.T,
) {
	s := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)

	// Lexicon resource fails but tool still works.
	s.HandlePartialFailure("lexicon")

	// Tool should still be available.
	if !s.Capabilities.Tools.ValidateArtifact {
		t.Fatal(
			"expected ValidateArtifact still available",
		)
	}
	// Lexicon resource should be degraded.
	if s.Capabilities.Resources.Lexicon {
		t.Fatal(
			"expected Lexicon unavailable after " +
				"partial failure",
		)
	}
	// Schema definitions should still work.
	if !s.Capabilities.Resources.SchemaDefinitions {
		t.Fatal(
			"expected SchemaDefinitions still available",
		)
	}
	// MCP is still connected (not full disconnection).
	if s.GetMCPStatus() != session.MCPConnected {
		t.Fatal("expected MCPConnected (partial failure)")
	}
	// Should not be in full fallback mode.
	if s.IsFallback() {
		t.Fatal(
			"expected not in full fallback for " +
				"partial failure",
		)
	}
	// Should list degraded capability.
	if len(s.DegradedCapabilities) == 0 {
		t.Fatal("expected degraded capabilities listed")
	}
}

func TestSession_HandlePartialFailure_SchemaDocsDown(
	t *testing.T,
) {
	s := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)

	s.HandlePartialFailure("schema_definitions")

	if !s.Capabilities.Resources.Lexicon {
		t.Fatal("expected Lexicon still available")
	}
	if s.Capabilities.Resources.SchemaDefinitions {
		t.Fatal(
			"expected SchemaDefinitions unavailable",
		)
	}
	if !s.Capabilities.Tools.ValidateArtifact {
		t.Fatal(
			"expected ValidateArtifact still available",
		)
	}
}

func TestSession_HasPrompts(t *testing.T) {
	artifact := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)
	if !artifact.HasPrompts() {
		t.Fatal("expected true for artifact mode")
	}

	advisory := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeAdvisory,
	)
	if advisory.HasPrompts() {
		t.Fatal("expected false for advisory mode")
	}
}
