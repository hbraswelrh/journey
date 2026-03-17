// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/schema"
	"github.com/hbraswelrh/pacman/internal/session"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// HandoffSummary is a structured transition context
// presented to the user after completing a tutorial in
// the Pac-Man terminal. It bridges the learn-to-author
// transition by directing the user to OpenCode with the
// gemara-mcp server.
type HandoffSummary struct {
	// ArtifactType is the target artifact type for
	// authoring (e.g., "ThreatCatalog").
	ArtifactType string
	// SchemaDef is the CUE schema definition for
	// validation (e.g., "#ThreatCatalog").
	SchemaDef string
	// MCPPrompt is the MCP wizard prompt name, or empty
	// if no wizard exists for this artifact type.
	MCPPrompt string
	// MCPResources lists available MCP resources
	// (e.g., "gemara://lexicon").
	MCPResources []string
	// MCPTools lists available MCP tools
	// (e.g., "validate_gemara_artifact").
	MCPTools []string
	// MCPConfigured is true when the gemara-mcp server
	// is configured in opencode.json.
	MCPConfigured bool
	// ServerMode is the MCP server operating mode.
	ServerMode string
	// SchemaVersion is the auto-selected schema version.
	SchemaVersion string
	// ExperimentalSchemas lists schemas with experimental
	// status at the selected version.
	ExperimentalSchemas []string
	// VersionMismatch is true when the MCP server version
	// differs from the selected schema version.
	VersionMismatch bool
	// KeyDecisions are decisions the user should have
	// answers for based on the tutorial.
	KeyDecisions []string
	// PreparationChecklist contains items the user should
	// prepare before beginning authoring.
	PreparationChecklist []string
	// TutorialTitle is the title of the completed
	// tutorial.
	TutorialTitle string
	// Layer is the Gemara layer of the completed
	// tutorial.
	Layer int
}

// artifactSchemaMap maps artifact type identifiers to their
// CUE schema definition names. This is used by
// BuildHandoffSummary to look up schema defs without
// importing the authoring package.
var artifactSchemaMap = map[string]string{
	consts.ArtifactGuidanceCatalog: consts.SchemaGuidanceCatalog,
	consts.ArtifactControlCatalog:  consts.SchemaControlCatalog,
	consts.ArtifactThreatCatalog:   consts.SchemaThreatCatalog,
	consts.ArtifactPolicy:          consts.SchemaPolicy,
	consts.ArtifactMappingDocument: consts.SchemaMappingDocument,
	consts.ArtifactEvaluationLog:   consts.SchemaEvaluationLog,
}

// BuildHandoffSummary creates a HandoffSummary from a
// completed tutorial step, session state, and version
// selection result.
func BuildHandoffSummary(
	step *tutorials.PathStep,
	sess *session.Session,
	selRes *schema.SelectionResult,
) *HandoffSummary {
	summary := &HandoffSummary{
		TutorialTitle: step.Tutorial.Title,
		Layer:         step.Layer,
		SchemaVersion: sess.SchemaVersion,
		ServerMode:    sess.GetServerMode(),
		MCPResources: []string{
			consts.ResourceLexicon,
			consts.ResourceSchemaDefinitions,
		},
		MCPTools: []string{
			consts.ToolValidateArtifact,
		},
	}

	// Determine MCP configuration status from session.
	mcpStatus := sess.GetMCPStatus()
	summary.MCPConfigured = mcpStatus == session.MCPConnected

	// Resolve artifact type from layer.
	artifacts := consts.LayerArtifacts[step.Layer]
	if len(artifacts) > 0 {
		summary.ArtifactType = artifacts[0]
		summary.SchemaDef = artifactSchemaMap[artifacts[0]]
		summary.MCPPrompt =
			consts.ArtifactWizards[artifacts[0]]
	}

	// Load preparation checklist.
	if summary.ArtifactType != "" {
		summary.PreparationChecklist =
			consts.DefaultPreparationChecklists[summary.ArtifactType]
	}

	// Derive key decisions from primary sections.
	if len(step.PrimarySections) > 0 {
		summary.KeyDecisions = make(
			[]string, len(step.PrimarySections),
		)
		copy(
			summary.KeyDecisions,
			step.PrimarySections,
		)
	}

	// Copy experimental schemas from selection result.
	if selRes != nil {
		summary.ExperimentalSchemas =
			selRes.ExperimentalSchemas
		summary.VersionMismatch =
			selRes.CompatWarning != ""
	}

	return summary
}

// RenderHandoffSummary renders a handoff summary to the
// given writer using the established Pac-Man visual style.
// Output is designed to be user-friendly and accessible
// for all audiences.
func RenderHandoffSummary(
	summary *HandoffSummary,
	out io.Writer,
) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())

	// Build the card content.
	var content strings.Builder

	// Header.
	if summary.ArtifactType != "" {
		content.WriteString(
			headingStyle.Render(
				"Ready to Author: " +
					summary.ArtifactType,
			),
		)
	} else {
		content.WriteString(
			headingStyle.Render(
				"Tutorial Complete: " +
					summary.TutorialTitle,
			),
		)
	}
	content.WriteString("\n\n")

	// Schema and version info.
	if summary.SchemaDef != "" {
		content.WriteString(
			annotationLabelStyle.Render("Schema:  ") +
				summary.SchemaDef + "\n",
		)
	}
	if summary.SchemaVersion != "" {
		content.WriteString(
			annotationLabelStyle.Render("Version: ") +
				summary.SchemaVersion + "\n",
		)
	}
	if summary.MCPPrompt != "" {
		content.WriteString(
			annotationLabelStyle.Render("Wizard:  ") +
				summary.MCPPrompt + "\n",
		)
	}
	content.WriteString("\n")

	// Available in OpenCode section.
	if summary.MCPConfigured {
		content.WriteString(
			annotationLabelStyle.Render(
				"Available in OpenCode:",
			) + "\n",
		)
		if len(summary.MCPTools) > 0 {
			content.WriteString(
				"  " +
					annotationLabelStyle.Render(
						"Tools:     ",
					) +
					strings.Join(
						summary.MCPTools, ", ",
					) + "\n",
			)
		}
		if len(summary.MCPResources) > 0 {
			for i, res := range summary.MCPResources {
				if i == 0 {
					content.WriteString(
						"  " +
							annotationLabelStyle.Render(
								"Resources: ",
							) +
							res + "\n",
					)
				} else {
					content.WriteString(
						"             " + res + "\n",
					)
				}
			}
		}
		if summary.MCPPrompt != "" {
			content.WriteString(
				"  " +
					annotationLabelStyle.Render(
						"Prompts:   ",
					) +
					summary.MCPPrompt + "\n",
			)
		}
		content.WriteString("\n")
	}

	// Key decisions.
	if len(summary.KeyDecisions) > 0 {
		content.WriteString(
			annotationLabelStyle.Render(
				"Key Decisions:",
			) + "\n",
		)
		for i, decision := range summary.KeyDecisions {
			content.WriteString(
				fmt.Sprintf(
					"  %d. %s\n", i+1, decision,
				),
			)
		}
		content.WriteString("\n")
	}

	// Preparation checklist.
	if len(summary.PreparationChecklist) > 0 {
		content.WriteString(
			annotationLabelStyle.Render(
				"Preparation Checklist:",
			) + "\n",
		)
		for _, item := range summary.PreparationChecklist {
			content.WriteString(
				"  • " + item + "\n",
			)
		}
		content.WriteString("\n")
	}

	// Version mismatch warning.
	if summary.VersionMismatch {
		content.WriteString(
			RenderWarning(
				"The MCP server schema version differs "+
					"from the selected version ("+
					summary.SchemaVersion+"). "+
					"Validate your artifact after "+
					"authoring.",
			) + "\n\n",
		)
	}

	// Next steps section.
	if summary.MCPConfigured {
		content.WriteString(
			headingStyle.Render(
				"Next: Open an OpenCode Session",
			) + "\n\n",
		)
		content.WriteString(
			subtleStyle.Render(
				"Launch opencode and use the gemara-mcp",
			) + "\n",
		)
		content.WriteString(
			subtleStyle.Render(
				"server to begin guided authoring.",
			) + "\n\n",
		)

		var cmdBlock string
		if summary.MCPPrompt != "" {
			cmdBlock = "$ opencode\n\n" +
				"Then tell the AI:\n" +
				"\"Run the " +
				summary.MCPPrompt +
				" wizard for my component\""
		} else {
			cmdBlock = "$ opencode\n\n" +
				"Then tell the AI:\n" +
				"\"Help me create a " +
				summary.ArtifactType +
				" using the gemara-mcp resources\""
		}
		content.WriteString(
			codeBlockStyle.Render(cmdBlock) + "\n",
		)
	} else {
		content.WriteString(
			headingStyle.Render(
				"Next: Set Up the MCP Server",
			) + "\n\n",
		)
		content.WriteString(
			subtleStyle.Render(
				"The gemara-mcp server is not yet "+
					"configured.",
			) + "\n",
		)
		content.WriteString(
			subtleStyle.Render(
				"Run ./pacman --doctor to verify your "+
					"environment,",
			) + "\n",
		)
		content.WriteString(
			subtleStyle.Render(
				"then configure the MCP server in "+
					"opencode.json.",
			) + "\n\n",
		)

		if summary.SchemaDef != "" {
			cueCmd := "cue vet -c " +
				"-d '" + summary.SchemaDef + "' " +
				"github.com/gemaraproj/gemara@latest " +
				"artifact.yaml"
			content.WriteString(
				annotationLabelStyle.Render(
					"Manual validation:",
				) + "\n",
			)
			content.WriteString(
				codeBlockStyle.Render(cueCmd) + "\n",
			)
		}
	}

	// Render the card.
	fmt.Fprintln(out, stepBarStyle.Render(content.String()))
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)
}
