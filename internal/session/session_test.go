// SPDX-License-Identifier: Apache-2.0

package session_test

import (
	"testing"

	"github.com/hbraswelrh/pacman/internal/session"
)

func TestNewSessionWithMCP(t *testing.T) {
	s := session.NewSessionWithMCP("v0.20.0")

	if s.GetMCPStatus() != session.MCPConnected {
		t.Fatal("expected MCPConnected")
	}
	if s.IsFallback() {
		t.Fatal("expected fallback to be false")
	}
	if !s.Tools.GetLexicon {
		t.Fatal("expected GetLexicon available")
	}
	if !s.Tools.ValidateArtifact {
		t.Fatal("expected ValidateArtifact available")
	}
	if !s.Tools.GetSchemaDocs {
		t.Fatal("expected GetSchemaDocs available")
	}
	if s.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected version v0.20.0, got %s",
			s.SchemaVersion,
		)
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
	if s.Tools.GetLexicon {
		t.Fatal("expected GetLexicon unavailable")
	}
	if s.Tools.ValidateArtifact {
		t.Fatal("expected ValidateArtifact unavailable")
	}
	if s.Tools.GetSchemaDocs {
		t.Fatal("expected GetSchemaDocs unavailable")
	}
	if len(s.DegradedCapabilities) == 0 {
		t.Fatal("expected degraded capabilities listed")
	}
}

func TestSession_HandleDisconnection(t *testing.T) {
	s := session.NewSessionWithMCP("v0.20.0")

	// Verify starts connected.
	if s.GetMCPStatus() != session.MCPConnected {
		t.Fatal("expected MCPConnected initially")
	}

	s.HandleDisconnection()

	if s.GetMCPStatus() != session.MCPDisconnected {
		t.Fatal("expected MCPDisconnected after disconnect")
	}
	if !s.IsFallback() {
		t.Fatal("expected fallback after disconnect")
	}
	if s.Tools.GetLexicon {
		t.Fatal(
			"expected GetLexicon unavailable after " +
				"disconnect",
		)
	}
	// Schema version should be preserved (no data loss).
	if s.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"schema version lost: expected v0.20.0, got %s",
			s.SchemaVersion,
		)
	}
}

func TestSession_HandleReconnection(t *testing.T) {
	s := session.NewSessionWithMCP("v0.20.0")
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
	if !s.Tools.GetLexicon {
		t.Fatal(
			"expected GetLexicon available after reconnect",
		)
	}
	if !s.Tools.ValidateArtifact {
		t.Fatal(
			"expected ValidateArtifact available after " +
				"reconnect",
		)
	}
	if !s.Tools.GetSchemaDocs {
		t.Fatal(
			"expected GetSchemaDocs available after " +
				"reconnect",
		)
	}
	if s.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"schema version lost: expected v0.20.0, got %s",
			s.SchemaVersion,
		)
	}
}
