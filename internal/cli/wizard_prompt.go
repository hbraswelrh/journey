// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// WizardInfo describes an available MCP wizard that can be
// launched through an AI agent like OpenCode.
type WizardInfo struct {
	// Name is the MCP prompt name.
	Name string
	// Title is the display title.
	Title string
	// Description explains what the wizard does.
	Description string
	// Layer is the Gemara layer this wizard targets.
	Layer int
	// Args are the required argument names.
	Args []string
}

// AvailableWizards returns the known gemara-mcp wizards.
// These are defined by the MCP server's prompts and are
// available when the MCP server is connected.
func AvailableWizards() []WizardInfo {
	return []WizardInfo{
		{
			Name:  consts.WizardThreatAssessment,
			Title: "Threat Assessment Wizard",
			Description: "Interactive wizard that guides " +
				"you through creating a Gemara-compatible " +
				"Threat Catalog (Layer 2) for your project",
			Layer: consts.LayerThreatsControls,
			Args:  []string{"component", "id_prefix"},
		},
		{
			Name:  consts.WizardControlCatalog,
			Title: "Control Catalog Wizard",
			Description: "Interactive wizard that guides " +
				"you through creating a Gemara-compatible " +
				"Control Catalog (Layer 2) for your project",
			Layer: consts.LayerThreatsControls,
			Args:  []string{"component", "id_prefix"},
		},
	}
}

// WizardPromptConfig holds dependencies for the wizard
// launcher.
type WizardPromptConfig struct {
	// Prompter handles user interaction.
	Prompter FreeTextPrompter
	// MCPAvailable indicates whether the MCP server is
	// connected.
	MCPAvailable bool
	// RoleName is the user's role for context.
	RoleName string
}

// WizardPromptResult holds the outcome of wizard selection.
type WizardPromptResult struct {
	// WizardName is the selected wizard's MCP prompt name.
	WizardName string
	// Component is the user-provided component name.
	Component string
	// IDPrefix is the user-provided ID prefix.
	IDPrefix string
	// LaunchCommand is the generated command to run the
	// wizard in OpenCode.
	LaunchCommand string
}

// RunWizardLauncher presents the available MCP wizards,
// collects arguments, and generates the launch command for
// the AI agent.
func RunWizardLauncher(
	cfg *WizardPromptConfig,
	out io.Writer,
) (*WizardPromptResult, error) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"Gemara MCP Wizards",
	))
	fmt.Fprintln(out)

	if !cfg.MCPAvailable {
		fmt.Fprintln(out, RenderWarning(
			"The MCP server is not connected. "+
				"Wizards require the gemara-mcp server "+
				"to be installed and configured.",
		))
		fmt.Fprintln(out)
		fmt.Fprintln(out, RenderNote(
			"You can still generate the launch "+
				"command and use it once the MCP "+
				"server is available.",
		))
		fmt.Fprintln(out)
	}

	wizards := AvailableWizards()

	// Filter by role context if available.
	if cfg.RoleName != "" {
		fmt.Fprintln(out, subtleStyle.Render(fmt.Sprintf(
			"Wizards relevant to your role (%s):",
			cfg.RoleName,
		)))
	} else {
		fmt.Fprintln(out, subtleStyle.Render(
			"Available wizards:",
		))
	}
	fmt.Fprintln(out)

	// Display wizard descriptions.
	for _, w := range wizards {
		fmt.Fprintln(out, RenderWizardCard(
			w.Title, w.Description, w.Layer,
		))
	}

	// Build selection options.
	options := make([]string, len(wizards)+1)
	for i, w := range wizards {
		options[i] = w.Title
	}
	options[len(wizards)] = "Back to main menu"

	choice, err := cfg.Prompter.Ask(
		"Select a wizard:", options,
	)
	if err != nil {
		return nil, fmt.Errorf("wizard selection: %w", err)
	}
	if choice >= len(wizards) {
		return nil, nil
	}

	selected := wizards[choice]

	// Collect wizard arguments.
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		selected.Title,
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		selected.Description,
	))
	fmt.Fprintln(out)

	component, err := cfg.Prompter.AskText(
		"Component name (e.g., 'container runtime', " +
			"'API gateway'):",
	)
	if err != nil {
		return nil, fmt.Errorf("component input: %w", err)
	}
	if component == "" {
		fmt.Fprintln(out, RenderWarning(
			"Component name is required.",
		))
		return nil, nil
	}

	idPrefix, err := cfg.Prompter.AskText(
		"ID prefix in ORG.PROJECT.COMPONENT format " +
			"(e.g., 'ACME.PLAT.GW'):",
	)
	if err != nil {
		return nil, fmt.Errorf("id prefix input: %w", err)
	}
	if idPrefix == "" {
		fmt.Fprintln(out, RenderWarning(
			"ID prefix is required.",
		))
		return nil, nil
	}

	// Generate the launch command.
	launchCmd := generateLaunchCommand(
		selected.Name, component, idPrefix,
	)

	// Display the result.
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderSuccess(
		"Wizard ready to launch",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		"Run the following in OpenCode to start "+
			"the wizard:",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  "+headingStyle.Render(launchCmd))
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"The wizard will guide you step by step "+
			"through creating a valid Gemara artifact. "+
			"It uses the MCP server's lexicon, schema "+
			"docs, and validation tools.",
	))
	fmt.Fprintln(out)

	return &WizardPromptResult{
		WizardName:    selected.Name,
		Component:     component,
		IDPrefix:      idPrefix,
		LaunchCommand: launchCmd,
	}, nil
}

// generateLaunchCommand produces the OpenCode prompt to
// invoke the MCP wizard.
func generateLaunchCommand(
	wizardName string,
	component string,
	idPrefix string,
) string {
	// The MCP prompt is invoked by asking the AI agent
	// to use it. The format references the prompt by name
	// with the collected arguments.
	return fmt.Sprintf(
		"Use the %s MCP prompt with component=%q "+
			"id_prefix=%q",
		wizardName,
		component,
		idPrefix,
	)
}

// RenderWizardCard renders a wizard description card.
func RenderWizardCard(
	title string,
	description string,
	layer int,
) string {
	var lines []string
	lines = append(lines,
		tutorialTitleStyle.Render(title)+"  "+
			RenderLayerBadge(layer),
	)
	lines = append(lines, "")
	lines = append(lines,
		subtleStyle.Render(description),
	)
	content := strings.Join(lines, "\n")
	return stepBarStyle.Render(content)
}
