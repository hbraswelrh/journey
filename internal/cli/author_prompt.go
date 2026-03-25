// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"

	"github.com/hbraswelrh/gemara-user-journey/internal/authoring"
	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/session"
)

// AuthorPromptConfig holds dependencies for the guided
// authoring flow.
type AuthorPromptConfig struct {
	// Prompter handles user interaction including free-text
	// input.
	Prompter FreeTextPrompter
	// Session is the current Gemara User Journey session.
	Session *session.Session
	// SchemaVersion is the selected Gemara schema version.
	SchemaVersion string
	// OutputDir is the directory for authored artifact
	// output.
	OutputDir string
	// OutputFormat is the output format ("yaml" or "json").
	OutputFormat string
	// RoleName is the user's identified role for
	// personalized guidance.
	RoleName string
	// Keywords are the user's activity keywords for
	// context.
	Keywords []string
}

// AuthorPromptResult holds the outcome of guided authoring.
type AuthorPromptResult struct {
	// Artifact is the authored artifact.
	Artifact *authoring.AuthoredArtifact
	// OutputPath is the file path of the written artifact.
	OutputPath string
	// ValidationErrors are any errors from final
	// validation.
	ValidationErrors []authoring.ValidationError
	// WizardDelegated is true if the user chose to use
	// an MCP wizard instead of the built-in authoring
	// flow.
	WizardDelegated bool
	// WizardName is the MCP prompt name to delegate to
	// (set when WizardDelegated is true).
	WizardName string
}

// RunGuidedAuthoring executes the full guided authoring
// flow:
//  1. Present available artifact types
//  2. Initialize authoring engine with selected template
//  3. Walk through each step: display guidance, prompt for
//     field values, validate
//  4. On completion, render output
func RunGuidedAuthoring(
	cfg *AuthorPromptConfig,
	out io.Writer,
) (*AuthorPromptResult, error) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"Guided Gemara Content Authoring",
	))
	fmt.Fprintln(out)

	// Step 1: Select artifact type.
	artifactType, err := selectArtifactType(cfg, out)
	if err != nil {
		return nil, fmt.Errorf(
			"select artifact type: %w", err,
		)
	}

	// Check if the selected artifact type has a
	// corresponding MCP wizard and the session is in
	// artifact mode.
	if cfg.Session != nil && cfg.Session.IsArtifactMode() {
		wizardName := artifactTypeWizard(artifactType)
		if wizardName != "" {
			useWizard, wizErr := offerWizardChoice(
				cfg.Prompter, out, artifactType,
				wizardName,
			)
			if wizErr != nil {
				return nil, fmt.Errorf(
					"wizard choice: %w", wizErr,
				)
			}
			if useWizard {
				fmt.Fprintln(out, RenderNote(
					"Delegating to the MCP "+
						wizardName+" prompt. "+
						"The wizard will guide "+
						"you interactively.",
				))
				// Return nil to signal wizard
				// delegation — the caller
				// (main.go) handles launching
				// the wizard flow.
				return &AuthorPromptResult{
					WizardDelegated: true,
					WizardName:      wizardName,
				}, nil
			}
		}
	}

	// Step 2: Initialize engine.
	templates := authoring.ArtifactTemplates()
	tmpl, ok := templates[artifactType]
	if !ok {
		return nil, fmt.Errorf(
			"no template for artifact type %q",
			artifactType,
		)
	}

	engine := authoring.NewAuthoringEngine(
		tmpl,
		cfg.RoleName,
		cfg.Keywords,
		nil,
	)

	// Update session state.
	if cfg.Session != nil {
		cfg.Session.SetAuthoringState(
			artifactType, "0/"+fmt.Sprintf(
				"%d", len(tmpl.Steps),
			),
		)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderSuccess(fmt.Sprintf(
		"Starting %s authoring (%d steps)",
		artifactType, len(tmpl.Steps),
	)))

	// Step 3: Walk through each step.
	for !engine.IsComplete() {
		stepErrs, err := runAuthoringStep(
			engine, cfg, out,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"authoring step: %w", err,
			)
		}
		if len(stepErrs) > 0 {
			// Show errors and retry the step.
			displayErrs := toDisplayErrors(stepErrs)
			fmt.Fprintln(out,
				RenderValidationResults(displayErrs),
			)
			continue
		}

		// Update progress in session.
		if cfg.Session != nil {
			completed, total := engine.Progress()
			cfg.Session.SetAuthoringState(
				artifactType,
				fmt.Sprintf("%d/%d", completed, total),
			)
		}
	}

	// Step 4: Build and output artifact.
	artifact := engine.BuildArtifact()
	artifact.SchemaVersion = cfg.SchemaVersion

	// Determine output format and directory.
	format := cfg.OutputFormat
	if format == "" {
		format = consts.DefaultArtifactFormat
	}
	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = consts.AuthoringOutputDir
	}

	// Write artifact.
	outputPath, err := authoring.WriteArtifact(
		artifact, outputDir, format,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"write artifact: %w", err,
		)
	}

	// Render summary.
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)

	completed, total := engine.Progress()
	fmt.Fprintln(out,
		RenderAuthoringProgress(completed, total),
	)
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderArtifactSummary(
		artifact.ArtifactType,
		artifact.SchemaDef,
		artifact.SchemaVersion,
		len(artifact.Sections),
		string(artifact.Status),
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderSuccess(
		"Artifact written to "+outputPath,
	))

	return &AuthorPromptResult{
		Artifact:   artifact,
		OutputPath: outputPath,
	}, nil
}

// selectArtifactType presents the available artifact types
// and returns the user's selection. When role context is
// available, types relevant to the role's layers are
// listed first.
func selectArtifactType(
	cfg *AuthorPromptConfig,
	out io.Writer,
) (string, error) {
	types := authoring.SupportedArtifactTypes()
	options := make([]string, len(types))
	for i, t := range types {
		options[i] = t
	}

	fmt.Fprintln(out, subtleStyle.Render(
		"Select the type of Gemara artifact to author:",
	))

	choice, err := cfg.Prompter.Ask(
		"Artifact type:", options,
	)
	if err != nil {
		return "", err
	}
	if choice < 0 || choice >= len(types) {
		return "", fmt.Errorf(
			"invalid artifact type selection: %d",
			choice,
		)
	}
	return types[choice], nil
}

// runAuthoringStep prompts the user for field values in
// the current step, displays guidance, and completes the
// step.
func runAuthoringStep(
	engine *authoring.AuthoringEngine,
	cfg *AuthorPromptConfig,
	out io.Writer,
) ([]authoring.ValidationError, error) {
	step := engine.CurrentStep()
	if step == nil {
		return nil, nil
	}

	completed, total := engine.Progress()

	// Display step guidance.
	fieldNames := make([]string, len(step.Fields))
	for i, f := range step.Fields {
		fieldNames[i] = f.Name
		if f.Required {
			fieldNames[i] += " *"
		}
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderAuthoringStep(
		step.Name,
		step.Description,
		step.RoleExplanation,
		fieldNames,
		completed,
		total,
	))
	fmt.Fprintln(out)

	// Prompt for each field.
	for _, field := range step.Fields {
		fmt.Fprintln(out, RenderFieldPrompt(
			field.Name,
			field.Description,
			field.ExampleValue,
			field.Required,
		))

		value, err := cfg.Prompter.AskText(
			fmt.Sprintf("  %s: ", field.Name),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"prompt field %q: %w",
				field.Name, err,
			)
		}

		// Use example value if user enters empty for
		// non-required fields.
		if value == "" && !field.Required {
			continue
		}

		if err := engine.SetFieldValue(
			field.Name, value,
		); err != nil {
			return nil, fmt.Errorf(
				"set field %q: %w",
				field.Name, err,
			)
		}
	}

	// Complete the step.
	return engine.CompleteStep()
}

// artifactTypeWizard returns the MCP wizard prompt name
// for a given artifact type, or empty string if no wizard
// exists for that type.
func artifactTypeWizard(artifactType string) string {
	switch artifactType {
	case consts.ArtifactThreatCatalog:
		return consts.WizardThreatAssessment
	case consts.ArtifactControlCatalog:
		return consts.WizardControlCatalog
	default:
		return ""
	}
}

// offerWizardChoice presents the user with a choice between
// using the MCP wizard or the built-in authoring flow.
func offerWizardChoice(
	prompter FreeTextPrompter,
	out io.Writer,
	artifactType string,
	wizardName string,
) (bool, error) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		fmt.Sprintf(
			"The MCP server provides a %q wizard "+
				"for %s authoring.",
			wizardName, artifactType,
		),
	))
	fmt.Fprintln(out)

	choice, err := prompter.Ask(
		"How would you like to author this artifact?",
		[]string{
			"Use MCP wizard (interactive, guided)",
			"Use built-in authoring flow",
		},
	)
	if err != nil {
		return false, err
	}
	return choice == 0, nil
}

// toDisplayErrors converts authoring ValidationErrors to
// the display-only type used by styles.
func toDisplayErrors(
	errs []authoring.ValidationError,
) []validationDisplayError {
	result := make(
		[]validationDisplayError, len(errs),
	)
	for i, e := range errs {
		result[i] = validationDisplayError{
			FieldPath:     e.FieldPath,
			Message:       e.Message,
			FixSuggestion: e.FixSuggestion,
		}
	}
	return result
}
