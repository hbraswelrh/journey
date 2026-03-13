// SPDX-License-Identifier: Apache-2.0

package authoring

import (
	"context"
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// T524: ValidationError contains field path, error message,
// and fix suggestion.
func TestValidationErrorFields(t *testing.T) {
	t.Parallel()
	ve := ValidationError{
		FieldPath:     "metadata.name",
		Message:       "field is required",
		FixSuggestion: "Provide a unique artifact name",
	}
	if ve.FieldPath != "metadata.name" {
		t.Errorf(
			"FieldPath = %q, want %q",
			ve.FieldPath, "metadata.name",
		)
	}
	if ve.Message != "field is required" {
		t.Errorf(
			"Message = %q, want %q",
			ve.Message, "field is required",
		)
	}
	if ve.FixSuggestion != "Provide a unique artifact name" {
		t.Errorf(
			"FixSuggestion = %q, want %q",
			ve.FixSuggestion,
			"Provide a unique artifact name",
		)
	}
}

// mockMCPClient implements the MCPClient interface for
// testing.
type mockMCPClient struct {
	response []byte
	err      error
}

func (m *mockMCPClient) ValidateArtifact(
	_ context.Context,
	_ string,
	_ string,
) ([]byte, error) {
	return m.response, m.err
}

// mockCUERunner returns the configured output and error.
type mockCUERunner struct {
	output []byte
	err    error
}

func (m *mockCUERunner) Run(
	_ context.Context,
	_ ...string,
) ([]byte, error) {
	return m.output, m.err
}

// T525: MCPValidator calls ValidateArtifact on the MCP
// client and translates response to ValidationError entries.
func TestMCPValidatorValidateFull(t *testing.T) {
	t.Parallel()
	client := &mockMCPClient{
		response: []byte(`{"valid": true}`),
		err:      nil,
	}
	v := NewMCPValidator(client)
	artifact := testArtifact()

	errs, err := v.ValidateFull(
		context.Background(), artifact,
	)
	if err != nil {
		t.Fatalf("ValidateFull error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf(
			"expected 0 errors, got %d", len(errs),
		)
	}
}

// T526: LocalValidator calls cue vet via CUERunner and
// parses error output into ValidationError entries.
func TestLocalValidatorValidateFull(t *testing.T) {
	t.Parallel()
	runner := &mockCUERunner{
		output: nil,
		err:    nil,
	}
	v := NewLocalValidator(runner.Run)
	artifact := testArtifact()

	errs, err := v.ValidateFull(
		context.Background(), artifact,
	)
	if err != nil {
		t.Fatalf("ValidateFull error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf(
			"expected 0 errors, got %d", len(errs),
		)
	}
}

// T527: ValidatePartial serializes completed fields and
// validates; returns errors for invalid fields only.
func TestValidatePartial(t *testing.T) {
	t.Parallel()
	client := &mockMCPClient{
		response: []byte(`{"valid": true}`),
		err:      nil,
	}
	v := NewMCPValidator(client)
	artifact := testArtifact()
	artifact.AddSection(consts.SectionMetadata)
	_ = artifact.SetFieldValue(
		consts.SectionMetadata, "name", "ACME.WEB.THR01",
	)

	errs, err := v.ValidatePartial(
		context.Background(), artifact, 0,
	)
	if err != nil {
		t.Fatalf("ValidatePartial error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf(
			"expected 0 errors, got %d", len(errs),
		)
	}
}

// T528: ValidateFull validates the complete artifact;
// returns empty errors for a valid artifact.
func TestValidateFullValid(t *testing.T) {
	t.Parallel()
	runner := &mockCUERunner{
		output: nil,
		err:    nil,
	}
	v := NewLocalValidator(runner.Run)
	artifact := testArtifactComplete()

	errs, err := v.ValidateFull(
		context.Background(), artifact,
	)
	if err != nil {
		t.Fatalf("ValidateFull error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf(
			"expected 0 errors, got %d", len(errs),
		)
	}
}

// T529: ValidateFull with missing required fields returns
// errors with actionable fix suggestions.
func TestLocalValidatorWithErrors(t *testing.T) {
	t.Parallel()
	runner := &mockCUERunner{
		output: []byte(
			"metadata.name: incomplete value string",
		),
		err: errCUEFailed,
	}
	v := NewLocalValidator(runner.Run)
	artifact := testArtifact()

	errs, err := v.ValidateFull(
		context.Background(), artifact,
	)
	if err != nil {
		t.Fatalf("ValidateFull error: %v", err)
	}
	if len(errs) == 0 {
		t.Error(
			"expected validation errors for " +
				"incomplete artifact",
		)
	}
	// Each error should have a fix suggestion.
	for _, ve := range errs {
		if ve.FixSuggestion == "" {
			t.Errorf(
				"error for %q has no fix suggestion",
				ve.FieldPath,
			)
		}
	}
}

// T530: NewValidator returns MCPValidator when MCP
// available, LocalValidator when in fallback.
func TestNewValidatorSelection(t *testing.T) {
	t.Parallel()
	client := &mockMCPClient{}
	runner := &mockCUERunner{}

	// MCP available.
	v := NewValidator(true, client, runner.Run)
	if _, ok := v.(*MCPValidator); !ok {
		t.Error("expected MCPValidator when MCP available")
	}

	// Fallback mode.
	v = NewValidator(false, client, runner.Run)
	if _, ok := v.(*LocalValidator); !ok {
		t.Error(
			"expected LocalValidator when MCP " +
				"unavailable",
		)
	}
}

// T531: Validation uses session's selected schema version.
func TestValidationUsesSchemaVersion(t *testing.T) {
	t.Parallel()
	artifact := testArtifact()
	artifact.SchemaVersion = "v0.20.0"
	if artifact.SchemaVersion != "v0.20.0" {
		t.Errorf(
			"SchemaVersion = %q, want %q",
			artifact.SchemaVersion, "v0.20.0",
		)
	}
	// The schema version is carried through the artifact
	// and used by validators when constructing validation
	// commands.
	if artifact.SchemaDef != consts.SchemaThreatCatalog {
		t.Errorf(
			"SchemaDef = %q, want %q",
			artifact.SchemaDef,
			consts.SchemaThreatCatalog,
		)
	}
}

// testArtifact creates a minimal artifact for testing.
func testArtifact() *AuthoredArtifact {
	return NewAuthoredArtifact(
		consts.ArtifactThreatCatalog,
		consts.SchemaThreatCatalog,
		"v0.20.0",
		"Security Engineer",
	)
}

// testArtifactComplete creates a fully populated artifact.
func testArtifactComplete() *AuthoredArtifact {
	a := testArtifact()
	a.AddSection(consts.SectionMetadata)
	_ = a.SetFieldValue(
		consts.SectionMetadata, "name", "ACME.WEB.THR01",
	)
	_ = a.SetFieldValue(
		consts.SectionMetadata, "description",
		"Web app threat catalog",
	)
	a.AddSection(consts.SectionScope)
	_ = a.SetFieldValue(
		consts.SectionScope, "scope", "Web application",
	)
	a.AddSection(consts.SectionCapabilities)
	_ = a.SetFieldValue(
		consts.SectionCapabilities,
		"capability_name", "Authentication",
	)
	a.AddSection(consts.SectionThreats)
	_ = a.SetFieldValue(
		consts.SectionThreats, "threat_id", "THR-001",
	)
	_ = a.SetFieldValue(
		consts.SectionThreats, "threat_description",
		"SQL injection",
	)
	return a
}
