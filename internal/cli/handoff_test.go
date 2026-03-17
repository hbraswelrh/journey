// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/schema"
	"github.com/hbraswelrh/pacman/internal/session"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// T007: BuildHandoffSummary produces correct summary for
// L2 tutorial step (ThreatCatalog + wizard).
func TestBuildHandoffSummary_L2Step(t *testing.T) {
	t.Parallel()

	step := &tutorials.PathStep{
		Tutorial: tutorials.Tutorial{
			Title: "Threat Assessment Tutorial",
			Layer: consts.LayerThreatsControls,
		},
		Layer: consts.LayerThreatsControls,
		PrimarySections: []string{
			"Scope Definition",
			"Capability Mapping",
		},
	}

	sess := session.NewSessionWithMCP("v0.20.0", "artifact")
	selRes := &schema.SelectionResult{
		SelectedTag:         "v0.20.0",
		ExperimentalSchemas: []string{"base"},
	}

	summary := BuildHandoffSummary(step, sess, selRes)

	if summary.ArtifactType != consts.ArtifactThreatCatalog {
		t.Errorf(
			"expected ThreatCatalog, got %s",
			summary.ArtifactType,
		)
	}
	if summary.SchemaDef != consts.SchemaThreatCatalog {
		t.Errorf(
			"expected %s, got %s",
			consts.SchemaThreatCatalog,
			summary.SchemaDef,
		)
	}
	if summary.MCPPrompt != consts.WizardThreatAssessment {
		t.Errorf(
			"expected threat_assessment prompt, "+
				"got %s", summary.MCPPrompt,
		)
	}
	if !summary.MCPConfigured {
		t.Error("expected MCPConfigured to be true")
	}
	if summary.SchemaVersion != "v0.20.0" {
		t.Errorf(
			"expected v0.20.0, got %s",
			summary.SchemaVersion,
		)
	}
	if len(summary.ExperimentalSchemas) == 0 {
		t.Error(
			"expected experimental schemas from " +
				"selection result",
		)
	}
	if len(summary.MCPResources) == 0 {
		t.Error("expected MCPResources to be populated")
	}
	if len(summary.MCPTools) == 0 {
		t.Error("expected MCPTools to be populated")
	}
	if len(summary.PreparationChecklist) == 0 {
		t.Error("expected preparation checklist")
	}
}

// T007: BuildHandoffSummary produces correct summary for
// L1 tutorial step (GuidanceCatalog, no wizard).
func TestBuildHandoffSummary_L1Step(t *testing.T) {
	t.Parallel()

	step := &tutorials.PathStep{
		Tutorial: tutorials.Tutorial{
			Title: "Guidance Catalog Tutorial",
			Layer: consts.LayerGuidance,
		},
		Layer:           consts.LayerGuidance,
		PrimarySections: []string{"Standards"},
	}

	sess := session.NewSessionWithMCP("v0.20.0", "artifact")

	summary := BuildHandoffSummary(step, sess, nil)

	if summary.ArtifactType != consts.ArtifactGuidanceCatalog {
		t.Errorf(
			"expected GuidanceCatalog, got %s",
			summary.ArtifactType,
		)
	}
	if summary.MCPPrompt != "" {
		t.Errorf(
			"expected no wizard for GuidanceCatalog, "+
				"got %s", summary.MCPPrompt,
		)
	}
}

// T007: BuildHandoffSummary handles L4 (no artifacts).
func TestBuildHandoffSummary_L4NoArtifacts(t *testing.T) {
	t.Parallel()

	step := &tutorials.PathStep{
		Tutorial: tutorials.Tutorial{
			Title: "Pipeline Security",
			Layer: consts.LayerSensitiveActivity,
		},
		Layer: consts.LayerSensitiveActivity,
	}

	sess := session.NewSessionWithoutMCP("v0.20.0")

	summary := BuildHandoffSummary(step, sess, nil)

	if summary.ArtifactType != "" {
		t.Errorf(
			"expected empty artifact type for L4, "+
				"got %s", summary.ArtifactType,
		)
	}
}

// T007: BuildHandoffSummary handles MCP not configured.
func TestBuildHandoffSummary_MCPNotConfigured(
	t *testing.T,
) {
	t.Parallel()

	step := &tutorials.PathStep{
		Tutorial: tutorials.Tutorial{
			Title: "Threat Assessment Tutorial",
			Layer: consts.LayerThreatsControls,
		},
		Layer: consts.LayerThreatsControls,
	}

	sess := session.NewSessionWithoutMCP("v0.20.0")

	summary := BuildHandoffSummary(step, sess, nil)

	if summary.MCPConfigured {
		t.Error(
			"expected MCPConfigured to be false " +
				"when session has no MCP",
		)
	}
}

// T007: RenderHandoffSummary output contains OpenCode and
// gemara-mcp references.
func TestRenderHandoffSummary_OpenCodeReferences(
	t *testing.T,
) {
	t.Parallel()

	summary := &HandoffSummary{
		ArtifactType:         consts.ArtifactThreatCatalog,
		SchemaDef:            consts.SchemaThreatCatalog,
		MCPPrompt:            consts.WizardThreatAssessment,
		MCPConfigured:        true,
		SchemaVersion:        "v0.20.0",
		MCPResources:         []string{consts.ResourceLexicon, consts.ResourceSchemaDefinitions},
		MCPTools:             []string{consts.ToolValidateArtifact},
		TutorialTitle:        "Threat Assessment Tutorial",
		Layer:                consts.LayerThreatsControls,
		KeyDecisions:         []string{"Component scope"},
		PreparationChecklist: []string{"Identify component"},
	}

	var buf bytes.Buffer
	RenderHandoffSummary(summary, &buf)
	output := buf.String()

	if !strings.Contains(output, "OpenCode") &&
		!strings.Contains(output, "opencode") {
		t.Error(
			"expected output to reference OpenCode",
		)
	}
	if !strings.Contains(
		output, consts.WizardThreatAssessment,
	) {
		t.Error(
			"expected output to contain " +
				"threat_assessment",
		)
	}
	if !strings.Contains(
		output, consts.ToolValidateArtifact,
	) {
		t.Error(
			"expected output to contain " +
				"validate_gemara_artifact",
		)
	}
	if !strings.Contains(
		output, consts.ResourceLexicon,
	) {
		t.Error(
			"expected output to contain " +
				"gemara://lexicon",
		)
	}
}

// T007: RenderHandoffSummary includes doctor reference
// when MCP is not configured.
func TestRenderHandoffSummary_MCPNotConfigured(
	t *testing.T,
) {
	t.Parallel()

	summary := &HandoffSummary{
		ArtifactType:         consts.ArtifactThreatCatalog,
		SchemaDef:            consts.SchemaThreatCatalog,
		MCPPrompt:            consts.WizardThreatAssessment,
		MCPConfigured:        false,
		SchemaVersion:        "v0.20.0",
		MCPResources:         []string{consts.ResourceLexicon},
		MCPTools:             []string{consts.ToolValidateArtifact},
		TutorialTitle:        "Threat Assessment Tutorial",
		Layer:                consts.LayerThreatsControls,
		KeyDecisions:         []string{"Component scope"},
		PreparationChecklist: []string{"Identify component"},
	}

	var buf bytes.Buffer
	RenderHandoffSummary(summary, &buf)
	output := buf.String()

	if !strings.Contains(output, "--doctor") {
		t.Error(
			"expected --doctor reference when MCP " +
				"not configured",
		)
	}
	if !strings.Contains(output, "cue vet") {
		t.Error(
			"expected cue vet fallback when MCP " +
				"not configured",
		)
	}
}

// T007: RenderHandoffSummary shows version mismatch warning.
func TestRenderHandoffSummary_VersionMismatch(
	t *testing.T,
) {
	t.Parallel()

	summary := &HandoffSummary{
		ArtifactType:         consts.ArtifactThreatCatalog,
		SchemaDef:            consts.SchemaThreatCatalog,
		MCPConfigured:        true,
		SchemaVersion:        "v0.20.0",
		VersionMismatch:      true,
		MCPResources:         []string{consts.ResourceLexicon},
		MCPTools:             []string{consts.ToolValidateArtifact},
		TutorialTitle:        "Threat Assessment Tutorial",
		Layer:                consts.LayerThreatsControls,
		KeyDecisions:         []string{"Component scope"},
		PreparationChecklist: []string{"Identify component"},
	}

	var buf bytes.Buffer
	RenderHandoffSummary(summary, &buf)
	output := buf.String()

	if !strings.Contains(
		strings.ToLower(output), "mismatch",
	) && !strings.Contains(
		strings.ToLower(output), "warning",
	) {
		t.Error(
			"expected version mismatch warning in " +
				"output",
		)
	}
}

// T007: RenderHandoffSummary lists available tools and
// resources.
func TestRenderHandoffSummary_ListsToolsAndResources(
	t *testing.T,
) {
	t.Parallel()

	summary := &HandoffSummary{
		ArtifactType:  consts.ArtifactControlCatalog,
		SchemaDef:     consts.SchemaControlCatalog,
		MCPPrompt:     consts.WizardControlCatalog,
		MCPConfigured: true,
		SchemaVersion: "v0.20.0",
		MCPResources: []string{
			consts.ResourceLexicon,
			consts.ResourceSchemaDefinitions,
		},
		MCPTools:             []string{consts.ToolValidateArtifact},
		TutorialTitle:        "Control Catalog Tutorial",
		Layer:                consts.LayerThreatsControls,
		KeyDecisions:         []string{"Framework selection"},
		PreparationChecklist: []string{"Identify component"},
	}

	var buf bytes.Buffer
	RenderHandoffSummary(summary, &buf)
	output := buf.String()

	if !strings.Contains(
		output, consts.ResourceSchemaDefinitions,
	) {
		t.Error(
			"expected gemara://schema/definitions " +
				"in output",
		)
	}
	if !strings.Contains(
		output, consts.WizardControlCatalog,
	) {
		t.Error(
			"expected control_catalog in output",
		)
	}
}
