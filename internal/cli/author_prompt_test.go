// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hbraswelrh/pacman/internal/cli"
	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/session"
)

// authorMockPrompter implements FreeTextPrompter for
// authoring tests.
type authorMockPrompter struct {
	choices   []int
	texts     []string
	choiceIdx int
	textIdx   int
}

func (m *authorMockPrompter) Ask(
	_ string,
	_ []string,
) (int, error) {
	if m.choiceIdx >= len(m.choices) {
		return 0, errors.New("no more choices")
	}
	choice := m.choices[m.choiceIdx]
	m.choiceIdx++
	return choice, nil
}

func (m *authorMockPrompter) AskText(
	_ string,
) (string, error) {
	if m.textIdx >= len(m.texts) {
		return "", errors.New("no more texts")
	}
	text := m.texts[m.textIdx]
	m.textIdx++
	return text, nil
}

// T540: RunGuidedAuthoring presents artifact type
// selection and returns selected type.
func TestRunGuidedAuthoring_TypeSelection(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	// Select ThreatCatalog (index 4 in
	// SupportedArtifactTypes order).
	// Then provide values for all required fields in
	// each step of ThreatCatalog template:
	// Step 1 (metadata): name, description, version
	// Step 2 (scope): scope, boundary
	// Step 3 (capabilities): capability_name,
	//   capability_description
	// Step 4 (threats): threat_id, threat_description,
	//   target_capability
	prompter := &authorMockPrompter{
		choices: []int{4}, // ThreatCatalog
		texts: []string{
			// metadata fields
			"ACME.WEB.THR01",
			"Test threat catalog",
			"1.0.0",
			// scope fields
			"Web application",
			"Third-party SaaS",
			// capabilities fields
			"Authentication",
			"User identity verification",
			// threats fields
			"THR-001",
			"SQL injection via unvalidated input",
			"Authentication",
		},
	}

	sess := session.NewSessionWithoutMCP("v0.20.0")

	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      consts.RoleSecurityEngineer,
		Keywords: []string{
			"threat modeling",
		},
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	if result.Artifact == nil {
		t.Fatal("expected Artifact in result")
	}
	if result.Artifact.ArtifactType !=
		consts.ArtifactThreatCatalog {
		t.Errorf(
			"ArtifactType = %q, want %q",
			result.Artifact.ArtifactType,
			consts.ArtifactThreatCatalog,
		)
	}
}

// T541: RunGuidedAuthoring walks through all steps of a
// ThreatCatalog template.
func TestRunGuidedAuthoring_AllSteps(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	prompter := &authorMockPrompter{
		choices: []int{4}, // ThreatCatalog
		texts: []string{
			"ACME.WEB.THR01",
			"Test threat catalog",
			"1.0.0",
			"Web application",
			"",
			"Authentication",
			"User identity verification",
			"THR-001",
			"SQL injection",
			"Authentication",
		},
	}

	sess := session.NewSessionWithoutMCP("v0.20.0")
	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      consts.RoleSecurityEngineer,
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	// Should have sections for all template steps.
	if len(result.Artifact.Sections) == 0 {
		t.Error("expected sections in artifact")
	}

	output := buf.String()
	// Output should mention step progression.
	if !strings.Contains(output, "Step") {
		t.Error("expected step indicators in output")
	}
}

// T542: Validation errors are displayed with fix
// suggestions (not directly testable without triggering
// validation, but we verify the output path exists).
func TestRunGuidedAuthoring_OutputExists(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	prompter := &authorMockPrompter{
		choices: []int{4},
		texts: []string{
			"ACME.WEB.THR01",
			"Test threat catalog",
			"",
			"Web application",
			"",
			"Auth",
			"Identity",
			"THR-001",
			"SQL injection",
			"Auth",
		},
	}

	sess := session.NewSessionWithoutMCP("v0.20.0")
	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      consts.RoleSecurityEngineer,
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	// Output file should exist.
	if result.OutputPath == "" {
		t.Fatal("expected OutputPath in result")
	}
	if _, err := os.Stat(
		result.OutputPath,
	); err != nil {
		t.Fatalf(
			"output file not found: %v", err,
		)
	}
}

// T543: Completed authoring produces a valid artifact and
// writes output file.
func TestRunGuidedAuthoring_ValidOutput(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	prompter := &authorMockPrompter{
		choices: []int{4},
		texts: []string{
			"ACME.WEB.THR01",
			"Threat catalog for web app",
			"1.0.0",
			"Web application security",
			"External APIs",
			"Authentication",
			"User identity and session management",
			"THR-001",
			"SQL injection via search params",
			"Authentication",
		},
	}

	sess := session.NewSessionWithoutMCP("v0.20.0")
	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      consts.RoleSecurityEngineer,
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	// Artifact should have correct metadata.
	if result.Artifact.SchemaVersion != "v0.20.0" {
		t.Errorf(
			"SchemaVersion = %q, want %q",
			result.Artifact.SchemaVersion,
			"v0.20.0",
		)
	}
	if result.Artifact.ArtifactType !=
		consts.ArtifactThreatCatalog {
		t.Errorf(
			"ArtifactType = %q, want %q",
			result.Artifact.ArtifactType,
			consts.ArtifactThreatCatalog,
		)
	}

	// File should contain YAML content.
	data, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(
		string(data), "ACME.WEB.THR01",
	) {
		t.Error(
			"output should contain artifact name",
		)
	}
}

// T544: Artifact type list is filtered by role when role
// context is available (verified by checking output
// contains the role name).
func TestRunGuidedAuthoring_RoleContext(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	prompter := &authorMockPrompter{
		choices: []int{4},
		texts: []string{
			"ACME.WEB.THR01",
			"Test",
			"",
			"Web app",
			"",
			"Auth",
			"Identity",
			"THR-001",
			"SQL injection",
			"Auth",
		},
	}

	sess := session.NewSessionWithoutMCP("v0.20.0")
	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      consts.RoleSecurityEngineer,
		Keywords:      []string{"threat modeling"},
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	// Artifact should reference the authoring role.
	if result.Artifact.AuthoringRole !=
		consts.RoleSecurityEngineer {
		t.Errorf(
			"AuthoringRole = %q, want %q",
			result.Artifact.AuthoringRole,
			consts.RoleSecurityEngineer,
		)
	}

	output := buf.String()
	// Output should contain role-personalized content.
	if !strings.Contains(
		output, consts.RoleSecurityEngineer,
	) {
		t.Error(
			"output should contain role-specific " +
				"guidance",
		)
	}
}

// T548: Full integration test — authoring flow produces
// correct YAML structure.
func TestRunGuidedAuthoring_Integration(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	prompter := &authorMockPrompter{
		choices: []int{4},
		texts: []string{
			"ACME.WEB.THR01",
			"Web Application Threat Assessment",
			"1.0.0",
			"Web application attack surface",
			"Third-party SaaS integrations",
			"Authentication",
			"User identity verification and session " +
				"management",
			"THR-001",
			"SQL injection via unvalidated user " +
				"input in search parameters",
			"Authentication",
		},
	}

	sess := session.NewSessionWithoutMCP("v0.20.0")
	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      consts.RoleSecurityEngineer,
		Keywords: []string{
			"threat modeling",
			"penetration testing",
		},
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	// Verify artifact structure.
	if result.Artifact == nil {
		t.Fatal("expected artifact")
	}
	if result.Artifact.ArtifactType !=
		consts.ArtifactThreatCatalog {
		t.Fatalf(
			"ArtifactType = %q, want %q",
			result.Artifact.ArtifactType,
			consts.ArtifactThreatCatalog,
		)
	}

	// Verify sections.
	expectedSections := []string{
		consts.SectionMetadata,
		consts.SectionScope,
		consts.SectionCapabilities,
		consts.SectionThreats,
	}
	if len(result.Artifact.Sections) !=
		len(expectedSections) {
		t.Fatalf(
			"Sections = %d, want %d",
			len(result.Artifact.Sections),
			len(expectedSections),
		)
	}
	for i, name := range expectedSections {
		if result.Artifact.Sections[i].Name != name {
			t.Errorf(
				"Section[%d].Name = %q, want %q",
				i, result.Artifact.Sections[i].Name,
				name,
			)
		}
	}

	// Verify output file.
	if result.OutputPath == "" {
		t.Fatal("expected output path")
	}
	if !strings.HasSuffix(
		result.OutputPath, ".yaml",
	) {
		t.Errorf(
			"output should be .yaml, got %q",
			result.OutputPath,
		)
	}

	// Verify file content.
	data, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "ACME.WEB.THR01") {
		t.Error("output should contain artifact name")
	}
	if !strings.Contains(content, "THR-001") {
		t.Error("output should contain threat ID")
	}

	// Verify session was updated.
	artType, progress := sess.GetAuthoringState()
	if artType != consts.ArtifactThreatCatalog {
		t.Errorf(
			"session ArtifactType = %q, want %q",
			artType, consts.ArtifactThreatCatalog,
		)
	}
	if progress == "" {
		t.Error("session progress should be set")
	}

	// Verify output contains progress and summary.
	output := buf.String()
	if !strings.Contains(output, "Progress") {
		t.Error(
			"output should contain progress indicator",
		)
	}
	if !strings.Contains(output, "Artifact") {
		t.Error("output should contain artifact summary")
	}
}

// T551: When session is in artifact mode and user selects
// ThreatCatalog, system offers choice between MCP wizard
// and built-in authoring flow.
func TestRunGuidedAuthoring_ArtifactMode_OffersWizard(
	t *testing.T,
) {
	t.Parallel()

	outputDir := t.TempDir()

	prompter := &authorMockPrompter{
		choices: []int{
			4, // ThreatCatalog
			1, // Built-in authoring (not wizard)
		},
		texts: []string{
			"ACME.WEB.THR01",
			"Test threat catalog",
			"1.0.0",
			"Web application",
			"Third-party SaaS",
			"Authentication",
			"User identity verification",
			"THR-001",
			"SQL injection via unvalidated input",
			"Authentication",
		},
	}

	sess := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)
	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      consts.RoleSecurityEngineer,
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	output := buf.String()
	// Should show wizard mention in the output.
	if !strings.Contains(output, "wizard") {
		t.Fatalf(
			"expected wizard mention in output, "+
				"got: %s",
			output,
		)
	}
	// Should still produce a valid artifact via built-in
	// (user chose option 1 = built-in).
	if result.Artifact == nil {
		t.Fatal("expected artifact from built-in flow")
	}
}

// T553: When session is in advisory mode, no wizard offer
// is presented for ThreatCatalog.
func TestRunGuidedAuthoring_AdvisoryMode_NoWizard(
	t *testing.T,
) {
	t.Parallel()

	outputDir := t.TempDir()

	prompter := &authorMockPrompter{
		choices: []int{4}, // ThreatCatalog (no wizard choice)
		texts: []string{
			"ACME.WEB.THR01",
			"Test",
			"",
			"Web app",
			"",
			"Auth",
			"Identity",
			"THR-001",
			"SQL injection",
			"Auth",
		},
	}

	sess := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeAdvisory,
	)
	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      consts.RoleSecurityEngineer,
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	output := buf.String()
	// Should NOT show wizard choice in advisory mode.
	if strings.Contains(output, "MCP wizard") {
		t.Fatalf(
			"expected no wizard offer in advisory "+
				"mode, got: %s",
			output,
		)
	}
	if result.Artifact == nil {
		t.Fatal("expected artifact")
	}
}

// T554: For GuidanceCatalog (no MCP prompt), built-in flow
// is always used regardless of mode.
func TestRunGuidedAuthoring_NoPromptArtifact_NoWizard(
	t *testing.T,
) {
	t.Parallel()

	outputDir := t.TempDir()

	prompter := &authorMockPrompter{
		// GuidanceCatalog is index 0 in
		// SupportedArtifactTypes.
		choices: []int{0},
		texts: []string{
			"ACME.WEB.GC01",
			"Guidance catalog",
			"1.0.0",
			"Web application",
			"",
			"Use HTTPS everywhere",
			"Ensure all connections use TLS 1.2+",
		},
	}

	sess := session.NewSessionWithMCP(
		"v0.20.0", consts.MCPModeArtifact,
	)
	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      consts.RoleSecurityEngineer,
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	output := buf.String()
	// GuidanceCatalog has no MCP prompt — no wizard offer.
	if strings.Contains(output, "MCP wizard") {
		t.Fatalf(
			"expected no wizard offer for "+
				"GuidanceCatalog, got: %s",
			output,
		)
	}
	if result.Artifact == nil {
		t.Fatal("expected artifact")
	}
}

// T549: JSON output format works correctly.
func TestRunGuidedAuthoring_JSONOutput(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	prompter := &authorMockPrompter{
		choices: []int{4},
		texts: []string{
			"ACME.WEB.THR01",
			"Test",
			"",
			"Web app",
			"",
			"Auth",
			"Identity",
			"THR-001",
			"SQL injection",
			"Auth",
		},
	}

	sess := session.NewSessionWithoutMCP("v0.20.0")
	cfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: "v0.20.0",
		OutputDir:     outputDir,
		OutputFormat:  "json",
		RoleName:      consts.RoleSecurityEngineer,
	}

	var buf bytes.Buffer
	result, err := cli.RunGuidedAuthoring(cfg, &buf)
	if err != nil {
		t.Fatalf("RunGuidedAuthoring: %v", err)
	}

	if !strings.HasSuffix(result.OutputPath, ".json") {
		t.Errorf(
			"output should be .json, got %q",
			result.OutputPath,
		)
	}

	// Verify it's valid JSON.
	data, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(data), "{") {
		t.Error("output should be JSON")
	}

	// Verify correct directory.
	dir := filepath.Dir(result.OutputPath)
	if dir != outputDir {
		t.Errorf(
			"output dir = %q, want %q",
			dir, outputDir,
		)
	}
}
