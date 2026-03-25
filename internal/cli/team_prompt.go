// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/roles"
	"github.com/hbraswelrh/gemara-user-journey/internal/team"
	"github.com/hbraswelrh/gemara-user-journey/internal/tutorials"
)

// TeamPromptConfig holds dependencies for the team setup
// flow.
type TeamPromptConfig struct {
	// Prompter handles user interaction.
	Prompter UserPrompter
	// TutorialsDir is the path to the Gemara tutorials.
	TutorialsDir string
	// SchemaVersion is the selected Gemara schema version.
	SchemaVersion string
	// TeamConfigDir is the directory for saved team
	// configs. Defaults to ~/.config/gemara-user-journey/teams/.
	TeamConfigDir string
}

// TeamPromptResult holds the outcome of team setup.
type TeamPromptResult struct {
	// Team is the configured team.
	Team *team.TeamConfig
	// View is the generated collaboration view.
	View *team.CollaborationView
}

// RunTeamSetup guides the user through configuring a team,
// adding members, and generating a collaboration view.
func RunTeamSetup(
	cfg *TeamPromptConfig,
	out io.Writer,
) (*TeamPromptResult, error) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"Team Collaboration Setup",
	))
	fmt.Fprintln(out)

	// Get team name.
	teamNamePrompter, ok :=
		cfg.Prompter.(FreeTextPrompter)
	if !ok {
		return nil, fmt.Errorf(
			"prompter does not support free text input",
		)
	}

	teamName, err := teamNamePrompter.AskText(
		"What is your team name?",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"prompt team name: %w", err,
		)
	}

	tc := team.NewTeamConfig(teamName)

	// Build role selection list.
	predefined := roles.PredefinedRoles()
	profileDir := resolveTeamProfileDir(cfg)
	customProfiles, _ := roles.ListProfiles(profileDir)
	allRoles := roles.MergeWithPredefined(
		predefined, customProfiles,
	)

	roleNames := make([]string, len(allRoles))
	for i, r := range allRoles {
		roleNames[i] = r.Name
	}

	// Add members loop.
	for {
		memberName, err := teamNamePrompter.AskText(
			"Enter team member name " +
				"(or empty to finish):",
		)
		if err != nil {
			return nil, fmt.Errorf(
				"prompt member name: %w", err,
			)
		}
		if memberName == "" {
			break
		}

		roleIdx, err := cfg.Prompter.Ask(
			fmt.Sprintf(
				"Select role for %s:", memberName,
			),
			roleNames,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"prompt role: %w", err,
			)
		}

		selectedRole := allRoles[roleIdx]
		member := team.TeamMember{
			Name:     memberName,
			RoleName: selectedRole.Name,
			Layers:   selectedRole.DefaultLayers,
			Keywords: selectedRole.DefaultKeywords,
		}

		if err := tc.AddMember(member); err != nil {
			fmt.Fprintln(out, RenderWarning(
				err.Error(),
			))
			continue
		}

		fmt.Fprintln(out, RenderSuccess(
			fmt.Sprintf(
				"Added %s as %s",
				memberName, selectedRole.Name,
			),
		))
	}

	if len(tc.Members) == 0 {
		fmt.Fprintln(out, RenderNote(
			"No members added. "+
				"Team setup cancelled.",
		))
		return &TeamPromptResult{Team: tc}, nil
	}

	// Load tutorials for handoff tutorial references.
	tuts, err := tutorials.LoadTutorials(
		cfg.TutorialsDir,
	)
	if err != nil {
		// Non-fatal: generate view without tutorials.
		tuts = nil
		fmt.Fprintln(out, RenderWarning(
			"Could not load tutorials: "+err.Error(),
		))
	}

	// Generate collaboration view.
	view := team.GenerateView(tc, tuts)
	RenderCollaborationView(view, out)

	// Save team config.
	teamDir := resolveTeamDir(cfg)
	if err := team.SaveTeam(teamDir, tc); err != nil {
		fmt.Fprintln(out, RenderWarning(
			"Could not save team config: "+
				err.Error(),
		))
	} else {
		fmt.Fprintln(out, RenderSuccess(
			"Team config saved",
		))
	}

	return &TeamPromptResult{
		Team: tc,
		View: view,
	}, nil
}

// RunHandoffInspection displays detailed information about
// a specific handoff point in the collaboration view.
func RunHandoffInspection(
	view *team.CollaborationView,
	idx int,
	out io.Writer,
) error {
	if idx < 0 || idx >= len(view.Handoffs) {
		return fmt.Errorf(
			"handoff index %d out of range (0-%d)",
			idx, len(view.Handoffs)-1,
		)
	}

	hp := view.Handoffs[idx]
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"Handoff Detail",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderHandoffPoint(hp))

	return nil
}

// resolveTeamDir returns the team config directory.
func resolveTeamDir(cfg *TeamPromptConfig) string {
	if cfg.TeamConfigDir != "" {
		return cfg.TeamConfigDir
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".", consts.TeamConfigDir)
	}
	return filepath.Join(configDir, consts.TeamConfigDir)
}

// resolveTeamProfileDir returns the role profiles directory.
func resolveTeamProfileDir(
	cfg *TeamPromptConfig,
) string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".", consts.RoleProfileDir)
	}
	return filepath.Join(
		configDir, consts.RoleProfileDir,
	)
}
