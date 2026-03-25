// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/roles"
	"github.com/hbraswelrh/gemara-user-journey/internal/tutorials"
)

// FreeTextPrompter extends UserPrompter with free-text input.
type FreeTextPrompter interface {
	UserPrompter
	// AskText presents a question and returns free-text
	// input from the user.
	AskText(question string) (string, error)
}

// WizardPrompter extends FreeTextPrompter with multi-select
// and confirmation capabilities for wizard flows.
type WizardPrompter interface {
	FreeTextPrompter
	// AskMultiSelect presents a question with options and
	// returns the indices of all selected options.
	AskMultiSelect(
		question string,
		options []string,
		defaults []int,
	) ([]int, error)
	// AskConfirm presents a yes/no question and returns
	// the answer.
	AskConfirm(question string) (bool, error)
	// AskTextWithDefault presents a question with a
	// pre-filled default value.
	AskTextWithDefault(
		question string,
		defaultValue string,
	) (string, error)
}

// RolePromptConfig holds dependencies for the role discovery
// flow.
type RolePromptConfig struct {
	// Prompter handles user interaction including
	// free-text input.
	Prompter FreeTextPrompter
	// TutorialsDir is the path to the Gemara tutorials
	// directory.
	TutorialsDir string
	// SchemaVersion is the user's selected schema version.
	SchemaVersion string
	// CustomProfiles are previously saved custom roles to
	// include in the selection list.
	CustomProfiles []roles.Role
}

// RolePromptResult holds the outcome of the role discovery
// flow.
type RolePromptResult struct {
	// Profile is the resolved activity profile.
	Profile *roles.ActivityProfile
	// Tutorials are the loaded tutorials.
	Tutorials []tutorials.Tutorial
	// VersionMismatches are tutorials whose schema version
	// differs from the selected version.
	VersionMismatches []tutorials.VersionMismatch
}

// RunRoleDiscovery executes the two-phase role discovery flow:
//
//  1. Phase 1 — Role Identification: present predefined roles,
//     handle "My role isn't listed" with free-text input,
//     partial match detection.
//  2. Phase 2 — Activity Probing: ask the user to describe
//     their activities, extract keywords, resolve layer
//     mappings, handle ambiguous keywords.
//
// Returns the resolved ActivityProfile and loaded tutorials.
func RunRoleDiscovery(
	cfg *RolePromptConfig,
	out io.Writer,
) (*RolePromptResult, error) {
	// Load tutorials.
	tutDir := cfg.TutorialsDir
	if tutDir == "" {
		tutDir = consts.DefaultTutorialsDir
	}

	loadedTutorials, err := tutorials.LoadTutorials(tutDir)
	if err != nil {
		// Non-fatal: proceed without tutorials but warn.
		fmt.Fprintln(out, RenderWarning(
			"Could not load tutorials: "+err.Error(),
		))
	}

	// Check version compatibility.
	var mismatches []tutorials.VersionMismatch
	if len(loadedTutorials) > 0 &&
		cfg.SchemaVersion != "" {
		mismatches = tutorials.CheckVersionCompat(
			loadedTutorials, cfg.SchemaVersion,
		)
		if len(mismatches) > 0 {
			fmt.Fprintln(out, RenderNote(fmt.Sprintf(
				"%d tutorial(s) reference a different "+
					"schema version than your "+
					"selection (%s).",
				len(mismatches), cfg.SchemaVersion,
			)))
		}
	}

	// Phase 1: Role Identification.
	fmt.Fprintln(out, RenderRoleHeader())

	selectedRole, err := runRoleSelection(cfg, out)
	if err != nil {
		return nil, fmt.Errorf("role selection: %w", err)
	}

	// Phase 2: Activity Probing.
	profile, err := runActivityProbing(
		cfg, selectedRole, out,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"activity probing: %w", err,
		)
	}

	// Populate artifact recommendations from resolved
	// layers and display them.
	profile.Recommendations =
		roles.ArtifactRecommendations(profile)
	if len(profile.Recommendations) > 0 {
		displayArtifactRecommendations(
			profile.Recommendations, out,
		)
	}

	return &RolePromptResult{
		Profile:           profile,
		Tutorials:         loadedTutorials,
		VersionMismatches: mismatches,
	}, nil
}

// runRoleSelection presents the role list and handles selection.
func runRoleSelection(
	cfg *RolePromptConfig,
	out io.Writer,
) (*roles.Role, error) {
	// Build the role list: predefined + custom + "My role
	// isn't listed."
	predefined := roles.PredefinedRoles()
	allRoles := make([]roles.Role, 0,
		len(predefined)+len(cfg.CustomProfiles)+1,
	)
	allRoles = append(allRoles, predefined...)
	allRoles = append(allRoles, cfg.CustomProfiles...)

	options := make([]string, 0, len(allRoles)+1)
	for _, r := range allRoles {
		label := r.Name
		if r.Description != "" {
			label += " — " + r.Description
		}
		options = append(options, label)
	}
	options = append(options,
		consts.RoleCustom+" — Define your own role",
	)

	idx, err := cfg.Prompter.Ask(
		"Select your role:", options,
	)
	if err != nil {
		return nil, err
	}

	// "My role isn't listed" is the last option.
	if idx == len(allRoles) {
		return handleCustomRoleInput(cfg, out)
	}

	role := &allRoles[idx]
	fmt.Fprintln(out, RenderSuccess(
		"Selected role: "+role.Name,
	))
	return role, nil
}

// handleCustomRoleInput accepts free-text role input and
// performs partial matching.
func handleCustomRoleInput(
	cfg *RolePromptConfig,
	out io.Writer,
) (*roles.Role, error) {
	input, err := cfg.Prompter.AskText(
		"What is your role?",
	)
	if err != nil {
		return nil, err
	}

	result := roles.MatchRole(input)

	switch result.Type {
	case roles.MatchExact:
		fmt.Fprintln(out, RenderSuccess(
			"Matched role: "+result.Role.Name,
		))
		return result.Role, nil

	case roles.MatchPartial:
		fmt.Fprintln(out, RenderNote(fmt.Sprintf(
			"Your role %q partially matches %q. "+
				"We'll refine through activity "+
				"probing.",
			input, result.Role.Name,
		)))
		return result.Role, nil

	default:
		// No match — create a minimal custom role from
		// the input.
		fmt.Fprintln(out, RenderNote(
			"No predefined role matches your input. "+
				"We'll determine your learning path "+
				"from your activities.",
		))
		customRole := &roles.Role{
			Name:   input,
			Source: roles.SourceCustom,
		}
		return customRole, nil
	}
}

// runActivityProbing asks the user about their activities
// and resolves layer mappings.
func runActivityProbing(
	cfg *RolePromptConfig,
	role *roles.Role,
	out io.Writer,
) (*roles.ActivityProfile, error) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderActivityHeader())

	description, err := cfg.Prompter.AskText(
		"Describe what you do or what problem you are " +
			"trying to solve (or press Enter to " +
			"select from categories):",
	)
	if err != nil {
		return nil, err
	}

	var keywords []string

	if strings.TrimSpace(description) != "" {
		keywords = roles.ExtractKeywords(description)
	}

	// If no keywords extracted, offer category selection.
	if len(keywords) == 0 {
		profile, err := handleCategorySelection(
			cfg, role, description, out,
		)
		if err != nil {
			return nil, err
		}
		return profile, nil
	}

	// Handle ambiguous keywords.
	ambiguous := roles.ClarificationNeeded(keywords)
	if len(ambiguous) > 0 {
		keywords, err = handleAmbiguousKeywords(
			cfg, keywords, ambiguous, out,
		)
		if err != nil {
			return nil, err
		}
	}

	profile := roles.ResolveLayerMappings(
		role, keywords, description,
	)

	// Display resolved layers.
	displayResolvedLayers(profile, out)

	return profile, nil
}

// handleCategorySelection presents activity categories for
// manual selection when no keywords were extracted.
func handleCategorySelection(
	cfg *RolePromptConfig,
	role *roles.Role,
	description string,
	out io.Writer,
) (*roles.ActivityProfile, error) {
	fmt.Fprintln(out, RenderNote(
		"Let's narrow down your activities. "+
			"Select the categories that best "+
			"describe your work:",
	))

	categories := roles.ActivityCategories()
	options := make([]string, len(categories))
	for i, cat := range categories {
		options[i] = fmt.Sprintf(
			"%s — %s", cat.Name, cat.Description,
		)
	}

	idx, err := cfg.Prompter.Ask(
		"Select your primary activity category:",
		options,
	)
	if err != nil {
		return nil, err
	}

	selected := categories[idx]
	keywords := selected.Keywords

	profile := roles.ResolveLayerMappings(
		role, keywords, description,
	)

	displayResolvedLayers(profile, out)
	return profile, nil
}

// handleAmbiguousKeywords asks clarifying questions for
// keywords that span multiple layers.
func handleAmbiguousKeywords(
	cfg *RolePromptConfig,
	keywords []string,
	ambiguous []string,
	out io.Writer,
) ([]string, error) {
	for _, kw := range ambiguous {
		fmt.Fprintln(out, RenderNote(fmt.Sprintf(
			"The term %q can apply to multiple "+
				"Gemara layers:",
			kw,
		)))

		options := []string{
			fmt.Sprintf(
				"Layer %d (Guidance) — defining "+
					"requirements and standards",
				consts.LayerGuidance,
			),
			fmt.Sprintf(
				"Layer %d (Risk & Policy) — "+
					"operationalizing in policy",
				consts.LayerRiskPolicy,
			),
			"Both layers apply to my work",
		}

		_, err := cfg.Prompter.Ask(
			fmt.Sprintf(
				"How does %q apply to your work?",
				kw,
			),
			options,
		)
		if err != nil {
			return nil, err
		}
		// The keywords list remains unchanged — the
		// ResolveLayerMappings function handles multi-
		// layer keywords. The clarification helps the
		// user understand the mapping.
	}

	return keywords, nil
}

// displayResolvedLayers shows the user which Gemara layers
// were identified from their profile using styled output.
func displayResolvedLayers(
	profile *roles.ActivityProfile,
	out io.Writer,
) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"Resolved Gemara Layers",
	))
	fmt.Fprintln(out)

	for _, lm := range profile.ResolvedLayers {
		badge := RenderLayerBadge(lm.Layer)
		conf := RenderConfidence(
			lm.Confidence == roles.ConfidenceStrong,
		)

		line := "  " + badge + "  " + conf
		if len(lm.Keywords) > 0 {
			line += "  " +
				RenderKeywordTags(lm.Keywords)
		}

		fmt.Fprintln(out, line)
	}
	fmt.Fprintln(out)
}

// displayArtifactRecommendations renders the recommended
// artifact types for the user's resolved layers. Each
// recommendation shows the artifact type, a user-facing
// description, and the authoring approach (MCP wizard or
// collaborative). Designed to be accessible for all
// audiences including non-technical stakeholders.
func displayArtifactRecommendations(
	recs []roles.ArtifactRecommendation,
	out io.Writer,
) {
	fmt.Fprintln(out, headingStyle.Render(
		"Recommended Artifact Outputs",
	))
	fmt.Fprintln(out)

	for _, rec := range recs {
		// Artifact type name with description.
		fmt.Fprintf(out, "  • %s\n",
			annotationLabelStyle.Render(rec.ArtifactType),
		)
		fmt.Fprintf(out, "    %s\n",
			subtleStyle.Render(rec.Description),
		)

		// Authoring approach.
		if rec.MCPWizard != "" {
			fmt.Fprintf(out, "    %s %s\n",
				successStyle.Render("→"),
				faintStyle.Render(
					"MCP Wizard: "+rec.MCPWizard,
				),
			)
		} else {
			fmt.Fprintf(out, "    %s %s\n",
				successStyle.Render("→"),
				faintStyle.Render(
					"Collaborative authoring with "+
						"MCP resources",
				),
			)
		}
		fmt.Fprintln(out)
	}
}

// RenderRoleHeader returns the styled role discovery header
// with a divider.
func RenderRoleHeader() string {
	return "\n" + RenderDivider() + "\n\n" +
		titleStyle.Render(
			" Role Discovery ",
		) + "\n"
}

// RenderActivityHeader returns the styled activity probing
// header with a divider.
func RenderActivityHeader() string {
	return RenderDivider() + "\n\n" +
		titleStyle.Render(
			" Activity Probing ",
		)
}
