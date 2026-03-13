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
	// Prompter handles user interaction with multi-select
	// and confirmation support.
	Prompter WizardPrompter
	// MCPAvailable indicates whether the MCP server is
	// connected.
	MCPAvailable bool
	// RoleName is the user's role for context.
	RoleName string
}

// WizardPromptResult holds the outcome of wizard execution.
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
	// CatalogRef is the selected reference catalog.
	CatalogRef string
	// Capabilities are the selected/approved capabilities.
	Capabilities []string
	// IncludeMITRE indicates whether MITRE ATT&CK linking
	// was requested.
	IncludeMITRE bool
}

// RunWizardLauncher presents the available MCP wizards
// and walks the user through a structured multi-step flow
// with selections, multi-choice, and pre-filled values.
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
			"You can still configure the wizard "+
				"and generate the launch command for "+
				"use once the MCP server is available.",
		))
		fmt.Fprintln(out)
	}

	wizards := AvailableWizards()

	// Display wizard cards.
	for _, w := range wizards {
		fmt.Fprintln(out, RenderWizardCard(
			w.Title, w.Description, w.Layer,
		))
	}

	// Step 1: Select wizard.
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

	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		selected.Title,
	))
	fmt.Fprintln(out)

	// Route to the specific wizard flow.
	switch selected.Name {
	case consts.WizardThreatAssessment:
		return runThreatWizard(cfg, selected, out)
	case consts.WizardControlCatalog:
		return runControlWizard(cfg, selected, out)
	default:
		return nil, fmt.Errorf(
			"unknown wizard: %s", selected.Name,
		)
	}
}

// --- Threat Assessment Wizard ---

func runThreatWizard(
	cfg *WizardPromptConfig,
	wizard WizardInfo,
	out io.Writer,
) (*WizardPromptResult, error) {
	result := &WizardPromptResult{
		WizardName: wizard.Name,
	}

	// Step 1: Component selection.
	fmt.Fprintln(out, renderWizardStep(
		1, 5, "Component Selection",
		"What component or technology are you "+
			"assessing?",
	))

	componentOpts := []string{
		"Container Runtime",
		"API Gateway",
		"Object Storage",
		"Database System",
		"CI/CD Pipeline",
		"Identity Provider",
		"Message Queue",
		"Load Balancer",
	}

	compChoice, err := cfg.Prompter.Ask(
		"Select a component type (or type your own):",
		append(componentOpts, "Enter custom component"),
	)
	if err != nil {
		return nil, err
	}

	if compChoice < len(componentOpts) {
		result.Component = componentOpts[compChoice]
	} else {
		custom, err := cfg.Prompter.AskText(
			"Component name:",
		)
		if err != nil {
			return nil, err
		}
		result.Component = custom
	}

	if result.Component == "" {
		return nil, nil
	}

	fmt.Fprintln(out, RenderSuccess(
		"Component: "+result.Component,
	))

	// Step 2: ID Prefix with suggestion.
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		2, 5, "ID Prefix",
		"Set the identifier prefix for this "+
			"artifact in ORG.PROJECT.COMPONENT format.",
	))

	suggestedPrefix := suggestIDPrefix(result.Component)
	prefix, err := cfg.Prompter.AskTextWithDefault(
		"ID prefix:",
		suggestedPrefix,
	)
	if err != nil {
		return nil, err
	}
	result.IDPrefix = prefix

	fmt.Fprintln(out, RenderSuccess(
		"ID prefix: "+result.IDPrefix,
	))

	// Step 3: Reference catalog selection.
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		3, 5, "Reference Catalog",
		"Select a reference catalog to import "+
			"capabilities and threats from.",
	))

	catalogOpts := []string{
		"FINOS CCC Core (recommended)",
		"OWASP Top 10",
		"CIS Benchmarks",
		"NIST SP 800-53",
		"No reference catalog (start from scratch)",
	}

	catChoice, err := cfg.Prompter.Ask(
		"Reference catalog:", catalogOpts,
	)
	if err != nil {
		return nil, err
	}

	switch catChoice {
	case 0:
		result.CatalogRef = "FINOS CCC Core"
	case 1:
		result.CatalogRef = "OWASP Top 10"
	case 2:
		result.CatalogRef = "CIS Benchmarks"
	case 3:
		result.CatalogRef = "NIST SP 800-53"
	default:
		result.CatalogRef = ""
	}

	if result.CatalogRef != "" {
		fmt.Fprintln(out, RenderSuccess(
			"Reference: "+result.CatalogRef,
		))
	}

	// Step 4: Capability selection (multi-select).
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		4, 5, "Capabilities",
		"Select the capabilities of "+
			result.Component+
			" that threats may target.",
	))

	capOpts := suggestCapabilities(result.Component)
	defaultCaps := make([]int, len(capOpts))
	for i := range capOpts {
		defaultCaps[i] = i
	}

	selectedCaps, err := cfg.Prompter.AskMultiSelect(
		"Select capabilities (all selected by default):",
		capOpts,
		defaultCaps,
	)
	if err != nil {
		return nil, err
	}

	for _, idx := range selectedCaps {
		if idx < len(capOpts) {
			result.Capabilities = append(
				result.Capabilities, capOpts[idx],
			)
		}
	}

	fmt.Fprintln(out, RenderSuccess(fmt.Sprintf(
		"%d capabilities selected",
		len(result.Capabilities),
	)))

	// Step 5: MITRE ATT&CK and launch.
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		5, 5, "Options & Launch",
		"Configure additional options and launch "+
			"the wizard.",
	))

	mitre, err := cfg.Prompter.AskConfirm(
		"Link threats to MITRE ATT&CK techniques?",
	)
	if err != nil {
		return nil, err
	}
	result.IncludeMITRE = mitre

	// Generate launch command.
	result.LaunchCommand = generateLaunchCommand(
		wizard.Name, result.Component, result.IDPrefix,
	)

	// Display summary.
	renderWizardSummary(result, out)

	return result, nil
}

// --- Control Catalog Wizard ---

func runControlWizard(
	cfg *WizardPromptConfig,
	wizard WizardInfo,
	out io.Writer,
) (*WizardPromptResult, error) {
	result := &WizardPromptResult{
		WizardName: wizard.Name,
	}

	// Step 1: Component selection.
	fmt.Fprintln(out, renderWizardStep(
		1, 4, "Component Selection",
		"What component or technology are you "+
			"creating controls for?",
	))

	componentOpts := []string{
		"Container Runtime",
		"API Gateway",
		"Object Storage",
		"Database System",
		"CI/CD Pipeline",
		"Identity Provider",
		"Message Queue",
		"Load Balancer",
	}

	compChoice, err := cfg.Prompter.Ask(
		"Select a component type (or type your own):",
		append(componentOpts, "Enter custom component"),
	)
	if err != nil {
		return nil, err
	}

	if compChoice < len(componentOpts) {
		result.Component = componentOpts[compChoice]
	} else {
		custom, err := cfg.Prompter.AskText(
			"Component name:",
		)
		if err != nil {
			return nil, err
		}
		result.Component = custom
	}

	if result.Component == "" {
		return nil, nil
	}

	fmt.Fprintln(out, RenderSuccess(
		"Component: "+result.Component,
	))

	// Step 2: ID Prefix.
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		2, 4, "ID Prefix",
		"Set the identifier prefix for this "+
			"artifact in ORG.PROJECT.COMPONENT format.",
	))

	suggestedPrefix := suggestIDPrefix(result.Component)
	prefix, err := cfg.Prompter.AskTextWithDefault(
		"ID prefix:",
		suggestedPrefix,
	)
	if err != nil {
		return nil, err
	}
	result.IDPrefix = prefix

	fmt.Fprintln(out, RenderSuccess(
		"ID prefix: "+result.IDPrefix,
	))

	// Step 3: Control source selection.
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		3, 4, "Control Sources",
		"Select control frameworks to import from.",
	))

	sourceOpts := []string{
		"FINOS CCC Core (recommended)",
		"OSPS Baseline",
		"CIS Benchmarks",
		"NIST SP 800-53",
		"Custom controls only",
	}

	sourceDefaults := []int{0}
	selectedSources, err := cfg.Prompter.AskMultiSelect(
		"Select control sources:",
		sourceOpts,
		sourceDefaults,
	)
	if err != nil {
		return nil, err
	}

	var sources []string
	for _, idx := range selectedSources {
		if idx < len(sourceOpts) {
			sources = append(sources, sourceOpts[idx])
		}
	}

	fmt.Fprintln(out, RenderSuccess(fmt.Sprintf(
		"%d source(s) selected", len(sources),
	)))

	// Step 4: Launch.
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		4, 4, "Launch",
		"Review and launch the wizard.",
	))

	result.LaunchCommand = generateLaunchCommand(
		wizard.Name, result.Component, result.IDPrefix,
	)

	renderWizardSummary(result, out)

	return result, nil
}

// --- Helpers ---

// renderWizardStep renders a wizard step header with
// progress.
func renderWizardStep(
	current int,
	total int,
	title string,
	description string,
) string {
	progress := fmt.Sprintf(
		"Step %d of %d", current, total,
	)
	var lines []string
	lines = append(lines,
		stepNumStyle.Render(progress)+"  "+
			headingStyle.Render(title),
	)
	lines = append(lines, "")
	lines = append(lines,
		subtleStyle.Render(description),
	)
	return stepBarStyle.Render(
		strings.Join(lines, "\n"),
	)
}

// suggestIDPrefix generates a suggested ID prefix from the
// component name.
func suggestIDPrefix(component string) string {
	// Convert component name to uppercase abbreviation.
	words := strings.Fields(
		strings.ToUpper(component),
	)
	if len(words) == 0 {
		return "ORG.PROJ.COMP"
	}
	var abbrev string
	for _, w := range words {
		if len(w) > 0 {
			abbrev += string(w[0])
			if len(w) > 1 {
				abbrev += string(w[1])
			}
		}
	}
	if len(abbrev) > 6 {
		abbrev = abbrev[:6]
	}
	return "ORG.PROJ." + abbrev
}

// suggestCapabilities returns suggested capabilities based
// on the component type.
func suggestCapabilities(component string) []string {
	lower := strings.ToLower(component)

	// Common capabilities across all components.
	common := []string{
		"Authentication",
		"Authorization",
		"Logging and Monitoring",
		"Data Encryption",
	}

	switch {
	case strings.Contains(lower, "container"):
		return append(common,
			"Image Management",
			"Network Isolation",
			"Resource Limits",
			"Runtime Security",
		)
	case strings.Contains(lower, "api") ||
		strings.Contains(lower, "gateway"):
		return append(common,
			"Rate Limiting",
			"Input Validation",
			"TLS Termination",
			"Request Routing",
		)
	case strings.Contains(lower, "storage") ||
		strings.Contains(lower, "database"):
		return append(common,
			"Access Control",
			"Backup and Recovery",
			"Data Integrity",
			"Encryption at Rest",
		)
	case strings.Contains(lower, "ci") ||
		strings.Contains(lower, "pipeline"):
		return append(common,
			"Pipeline Integrity",
			"Artifact Signing",
			"Secret Management",
			"Dependency Scanning",
		)
	case strings.Contains(lower, "identity") ||
		strings.Contains(lower, "idp"):
		return append(common,
			"Multi-Factor Authentication",
			"Session Management",
			"Federation",
			"Credential Storage",
		)
	default:
		return append(common,
			"Configuration Management",
			"Network Security",
			"Availability",
		)
	}
}

// generateLaunchCommand produces the OpenCode prompt to
// invoke the MCP wizard.
func generateLaunchCommand(
	wizardName string,
	component string,
	idPrefix string,
) string {
	return fmt.Sprintf(
		"Use the %s MCP prompt with component=%q "+
			"id_prefix=%q",
		wizardName,
		component,
		idPrefix,
	)
}

// renderWizardSummary displays the wizard configuration
// summary and launch instructions.
func renderWizardSummary(
	result *WizardPromptResult,
	out io.Writer,
) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)

	var lines []string
	lines = append(lines,
		headingStyle.Render("Wizard Summary"),
	)
	lines = append(lines, "")
	lines = append(lines,
		annotationLabelStyle.Render("Wizard: ")+
			result.WizardName,
	)
	lines = append(lines,
		annotationLabelStyle.Render("Component: ")+
			result.Component,
	)
	lines = append(lines,
		annotationLabelStyle.Render("ID Prefix: ")+
			result.IDPrefix,
	)
	if result.CatalogRef != "" {
		lines = append(lines,
			annotationLabelStyle.Render("Reference: ")+
				result.CatalogRef,
		)
	}
	if len(result.Capabilities) > 0 {
		lines = append(lines,
			annotationLabelStyle.Render(
				"Capabilities: ",
			)+fmt.Sprintf(
				"%d selected", len(result.Capabilities),
			),
		)
	}
	if result.IncludeMITRE {
		lines = append(lines,
			annotationLabelStyle.Render("MITRE ATT&CK: ")+
				successStyle.Render("enabled"),
		)
	}

	fmt.Fprintln(out, stepBarStyle.Render(
		strings.Join(lines, "\n"),
	))

	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderSuccess(
		"Wizard configured",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		"Run the following in OpenCode to start "+
			"the wizard:",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out,
		"  "+headingStyle.Render(result.LaunchCommand),
	)
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"The wizard will use the MCP server's "+
			"lexicon, schema docs, and validation "+
			"tools to guide you through creating a "+
			"valid Gemara artifact step by step.",
	))
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
