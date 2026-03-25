// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
)

// WizardInfo describes an available MCP wizard.
type WizardInfo struct {
	Name        string
	Title       string
	Description string
	Layer       int
	Args        []string
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

// WizardPromptConfig holds dependencies for the wizard.
type WizardPromptConfig struct {
	Prompter     WizardPrompter
	MCPAvailable bool
	ServerMode   string
	RoleName     string
}

// WizardPromptResult holds the outcome of wizard execution.
type WizardPromptResult struct {
	WizardName    string
	Component     string
	IDPrefix      string
	LaunchCommand string
	CatalogRef    string
	CatalogURL    string
	Description   string
	AuthorName    string
	AuthorID      string
	Capabilities  []string
	IncludeMITRE  bool
	// GuidelineFrameworks is for control catalog wizard.
	GuidelineFrameworks []string
}

// RunWizardLauncher presents the wizards and runs the
// selected one through its structured steps.
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
	} else if cfg.ServerMode != consts.MCPModeArtifact {
		fmt.Fprintln(out, RenderWarning(
			"The MCP server is running in advisory "+
				"mode. Wizards require artifact mode. "+
				"You can reconfigure the server mode "+
				"in opencode.json by changing --mode "+
				"to artifact.",
		))
		fmt.Fprintln(out)
		fmt.Fprintln(out, RenderNote(
			"You can still configure the wizard "+
				"and generate the launch command for "+
				"use once artifact mode is enabled.",
		))
		fmt.Fprintln(out)
	}

	wizards := AvailableWizards()
	for _, w := range wizards {
		fmt.Fprintln(out, RenderWizardCard(
			w.Title, w.Description, w.Layer,
		))
	}

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
// Mirrors gemara-mcp threat_assessment prompt exactly.

func runThreatWizard(
	cfg *WizardPromptConfig,
	wizard WizardInfo,
	out io.Writer,
) (*WizardPromptResult, error) {
	result := &WizardPromptResult{
		WizardName: wizard.Name,
	}

	// Step 1: Catalog Import
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		1, 5, "Catalog Import",
		"The FINOS CCC Core catalog provides "+
			"pre-built capabilities and threats you "+
			"can import rather than redefine.",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"  Catalog: https://github.com/finos/"+
			"common-cloud-controls/releases",
	))
	fmt.Fprintln(out)

	catChoice, err := cfg.Prompter.Ask(
		"Import from FINOS CCC Core?",
		[]string{
			"Yes, use FINOS CCC Core (recommended)",
			"Provide an alternative catalog URL",
			"No reference catalog (start from scratch)",
		},
	)
	if err != nil {
		return nil, err
	}

	switch catChoice {
	case 0:
		result.CatalogRef = "FINOS CCC Core"
		result.CatalogURL = "https://github.com/finos/" +
			"common-cloud-controls/releases/download/" +
			"v2025.10/CCC.Core_v2025.10.yaml"
		fmt.Fprintln(out, RenderSuccess(
			"Using FINOS CCC Core as reference catalog",
		))
	case 1:
		url, err := cfg.Prompter.AskText(
			"Catalog URL or file path:",
		)
		if err != nil {
			return nil, err
		}
		result.CatalogRef = "Custom"
		result.CatalogURL = url
		// Warn if not from trusted source.
		if !strings.Contains(url, "github.com/finos") &&
			!strings.Contains(
				url, "github.com/gemaraproj",
			) {
			fmt.Fprintln(out, RenderWarning(
				"Source is unverified. Confirm the "+
					"catalog is from a trusted source.",
			))
		}
		fmt.Fprintln(out, RenderSuccess(
			"Using custom catalog: "+url,
		))
	default:
		fmt.Fprintln(out, RenderNote(
			"No reference catalog. All capabilities "+
				"and threats will be custom.",
		))
	}

	// Step 2: Scope and Metadata
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		2, 5, "Scope and Metadata",
		"Define the component, author, and scope "+
			"for your Threat Catalog.",
	))

	component, err := cfg.Prompter.AskText(
		"Component name (e.g., 'container runtime', " +
			"'API gateway'):",
	)
	if err != nil {
		return nil, err
	}
	if component == "" {
		return nil, nil
	}
	result.Component = component

	description, err := cfg.Prompter.AskText(
		"Short description of what " + component +
			" does:",
	)
	if err != nil {
		return nil, err
	}
	result.Description = description

	authorName, err := cfg.Prompter.AskText(
		"Author name:",
	)
	if err != nil {
		return nil, err
	}
	result.AuthorName = authorName

	authorID, err := cfg.Prompter.AskTextWithDefault(
		"Author identifier:",
		strings.ToLower(
			strings.ReplaceAll(authorName, " ", "."),
		),
	)
	if err != nil {
		return nil, err
	}
	result.AuthorID = authorID

	suggestedPrefix := suggestIDPrefix(component)
	prefix, err := cfg.Prompter.AskTextWithDefault(
		"ID prefix (ORG.PROJECT.COMPONENT format):",
		suggestedPrefix,
	)
	if err != nil {
		return nil, err
	}
	result.IDPrefix = prefix

	// Show metadata YAML preview.
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"  Generated metadata:",
	))
	fmt.Fprintln(out, renderMetadataPreview(
		result, "ThreatCatalog",
	))

	confirmed, err := cfg.Prompter.AskConfirm(
		"Confirm metadata?",
	)
	if err != nil {
		return nil, err
	}
	if !confirmed {
		fmt.Fprintln(out, RenderNote("Metadata skipped."))
	} else {
		fmt.Fprintln(out, RenderSuccess(
			"Metadata confirmed",
		))
	}

	// Step 3: Identify Capabilities
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		3, 5, "Identify Capabilities",
		"What are the core functions or features "+
			"of "+component+"?",
	))
	fmt.Fprintln(out)

	if result.CatalogRef == "FINOS CCC Core" {
		fmt.Fprintln(out, faintStyle.Render(
			"  Capabilities from FINOS CCC Core are "+
				"available for import. Select which "+
				"to include.",
		))
		fmt.Fprintln(out)
	}

	capOpts := suggestCapabilities(component)

	// Present as a table like the gemara-mcp wizard.
	fmt.Fprintln(out, faintStyle.Render(
		"  Proposed capabilities:",
	))
	for i, cap := range capOpts {
		letter := string(rune('a' + i))
		source := "Custom"
		if i < 4 && result.CatalogRef != "" {
			source = "Import from " + result.CatalogRef
		}
		fmt.Fprintf(out,
			"  %s) %s — %s\n",
			letter, cap, faintStyle.Render(source),
		)
	}
	fmt.Fprintln(out)

	allCaps := make([]int, len(capOpts))
	for i := range capOpts {
		allCaps[i] = i
	}
	selectedCaps, err := cfg.Prompter.AskMultiSelect(
		"Select capabilities to include:",
		capOpts,
		allCaps,
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

	// Step 4: Identify Threats
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		4, 5, "Identify Threats",
		"For each capability, what could go wrong?",
	))
	fmt.Fprintln(out)

	mitre, err := cfg.Prompter.AskConfirm(
		"Link threats to MITRE ATT&CK techniques? " +
			"This adds structured vectors entries " +
			"referencing the ATT&CK Enterprise matrix.",
	)
	if err != nil {
		return nil, err
	}
	result.IncludeMITRE = mitre

	if mitre {
		fmt.Fprintln(out, RenderSuccess(
			"MITRE ATT&CK linking enabled",
		))
		fmt.Fprintln(out, faintStyle.Render(
			"  Reference: https://attack.mitre.org/"+
				"techniques/enterprise/",
		))
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"  Threat identification will be completed "+
			"interactively in the MCP wizard session. "+
			"The wizard proposes threats per capability "+
			"with table-based approval.",
	))

	// Step 5: Launch
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		5, 5, "Assemble and Validate",
		"The MCP wizard will assemble the complete "+
			"Threat Catalog YAML and validate it "+
			"against #ThreatCatalog using "+
			"validate_gemara_artifact.",
	))

	result.LaunchCommand = generateLaunchCommand(
		wizard.Name, result.Component, result.IDPrefix,
	)

	renderWizardSummary(result, out)
	return result, nil
}

// --- Control Catalog Wizard ---
// Mirrors gemara-mcp control_catalog prompt exactly.

func runControlWizard(
	cfg *WizardPromptConfig,
	wizard WizardInfo,
	out io.Writer,
) (*WizardPromptResult, error) {
	result := &WizardPromptResult{
		WizardName: wizard.Name,
	}

	// Step 1: Catalog Import (identical to threat wizard)
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		1, 5, "Catalog Import",
		"The FINOS CCC Core catalog provides "+
			"pre-built controls, families, and threat "+
			"mappings you can import rather than "+
			"redefine.",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"  Catalog: https://github.com/finos/"+
			"common-cloud-controls/releases",
	))
	fmt.Fprintln(out)

	catChoice, err := cfg.Prompter.Ask(
		"Import from FINOS CCC Core?",
		[]string{
			"Yes, use FINOS CCC Core (recommended)",
			"Provide an alternative catalog URL",
			"No reference catalog (start from scratch)",
		},
	)
	if err != nil {
		return nil, err
	}

	switch catChoice {
	case 0:
		result.CatalogRef = "FINOS CCC Core"
		result.CatalogURL = "https://github.com/finos/" +
			"common-cloud-controls/releases/download/" +
			"v2025.10/CCC.Core_v2025.10.yaml"
		fmt.Fprintln(out, RenderSuccess(
			"Using FINOS CCC Core as reference catalog",
		))
	case 1:
		url, err := cfg.Prompter.AskText(
			"Catalog URL or file path:",
		)
		if err != nil {
			return nil, err
		}
		result.CatalogRef = "Custom"
		result.CatalogURL = url
		if !strings.Contains(url, "github.com/finos") &&
			!strings.Contains(
				url, "github.com/gemaraproj",
			) {
			fmt.Fprintln(out, RenderWarning(
				"Source is unverified. Confirm the "+
					"catalog is from a trusted source.",
			))
		}
		fmt.Fprintln(out, RenderSuccess(
			"Using custom catalog: "+url,
		))
	default:
		fmt.Fprintln(out, RenderNote(
			"No reference catalog. All controls "+
				"will be custom.",
		))
	}

	// Step 2: Scope and Metadata
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		2, 5, "Scope and Metadata",
		"Define the component, author, guideline "+
			"frameworks, and scope for your Control "+
			"Catalog.",
	))

	component, err := cfg.Prompter.AskText(
		"Component name (e.g., 'container runtime', " +
			"'API gateway'):",
	)
	if err != nil {
		return nil, err
	}
	if component == "" {
		return nil, nil
	}
	result.Component = component

	description, err := cfg.Prompter.AskText(
		"Short description of what " + component +
			" does:",
	)
	if err != nil {
		return nil, err
	}
	result.Description = description

	authorName, err := cfg.Prompter.AskText(
		"Author name:",
	)
	if err != nil {
		return nil, err
	}
	result.AuthorName = authorName

	authorID, err := cfg.Prompter.AskTextWithDefault(
		"Author identifier:",
		strings.ToLower(
			strings.ReplaceAll(authorName, " ", "."),
		),
	)
	if err != nil {
		return nil, err
	}
	result.AuthorID = authorID

	suggestedPrefix := suggestIDPrefix(component)
	prefix, err := cfg.Prompter.AskTextWithDefault(
		"ID prefix (ORG.PROJECT.COMPONENT format):",
		suggestedPrefix,
	)
	if err != nil {
		return nil, err
	}
	result.IDPrefix = prefix

	// Guideline framework selection (control catalog
	// specific, from gemara-mcp system prompt).
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"  Select Layer 1 guideline frameworks to "+
			"map controls against:",
	))

	frameworkOpts := []string{
		"NIST Cybersecurity Framework (CSF)",
		"CSA Cloud Controls Matrix (CCM)",
		"ISO/IEC 27001",
		"NIST SP 800-53",
	}

	// Present as lettered table like gemara-mcp.
	for i, fw := range frameworkOpts {
		letter := string(rune('a' + i))
		fmt.Fprintf(out,
			"  %s) %s\n", letter, fw,
		)
	}
	fmt.Fprintln(out)

	selectedFw, err := cfg.Prompter.AskMultiSelect(
		"Select guideline frameworks:",
		frameworkOpts,
		[]int{0, 3}, // Default: CSF + NIST 800-53
	)
	if err != nil {
		return nil, err
	}

	for _, idx := range selectedFw {
		if idx < len(frameworkOpts) {
			result.GuidelineFrameworks = append(
				result.GuidelineFrameworks,
				frameworkOpts[idx],
			)
		}
	}

	fmt.Fprintln(out, RenderSuccess(fmt.Sprintf(
		"%d guideline framework(s) selected",
		len(result.GuidelineFrameworks),
	)))

	// Show metadata preview.
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"  Generated metadata:",
	))
	fmt.Fprintln(out, renderMetadataPreview(
		result, "ControlCatalog",
	))

	confirmed, err := cfg.Prompter.AskConfirm(
		"Confirm metadata?",
	)
	if err != nil {
		return nil, err
	}
	if !confirmed {
		fmt.Fprintln(out, RenderNote("Metadata skipped."))
	} else {
		fmt.Fprintln(out, RenderSuccess(
			"Metadata confirmed",
		))
	}

	// Step 3: Define Control Families
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		3, 5, "Define Control Families",
		"What logical groupings should your "+
			"controls fall into?",
	))
	fmt.Fprintln(out)

	familyOpts := suggestControlFamilies(component)

	fmt.Fprintln(out, faintStyle.Render(
		"  Proposed control families:",
	))
	for i, fam := range familyOpts {
		letter := string(rune('a' + i))
		fmt.Fprintf(out,
			"  %s) %s\n", letter, fam,
		)
	}
	fmt.Fprintln(out)

	allFamilies := make([]int, len(familyOpts))
	for i := range familyOpts {
		allFamilies[i] = i
	}
	_, err = cfg.Prompter.AskMultiSelect(
		"Select control families:",
		familyOpts,
		allFamilies,
	)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(out, RenderSuccess(
		"Control families confirmed",
	))

	// Step 4: Define Controls
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		4, 5, "Define Controls",
		"For each family, what risks need to be "+
			"reduced? Controls will be defined "+
			"interactively in the MCP wizard session "+
			"with threat mappings, guideline mappings, "+
			"and assessment requirements.",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"  The MCP wizard will walk through each "+
			"control with:\n"+
			"  - Risk-reduction objective\n"+
			"  - Threat mappings (table approval)\n"+
			"  - Guideline mappings (table approval)\n"+
			"  - Assessment requirements (testable "+
			"statements)",
	))

	// Step 5: Launch
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderWizardStep(
		5, 5, "Assemble and Validate",
		"The MCP wizard will assemble the complete "+
			"Control Catalog YAML and validate it "+
			"against #ControlCatalog using "+
			"validate_gemara_artifact.",
	))

	result.LaunchCommand = generateLaunchCommand(
		wizard.Name, result.Component, result.IDPrefix,
	)

	renderWizardSummary(result, out)
	return result, nil
}

// --- Shared helpers ---

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

func suggestIDPrefix(component string) string {
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

func suggestCapabilities(component string) []string {
	lower := strings.ToLower(component)

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

func suggestControlFamilies(component string) []string {
	lower := strings.ToLower(component)

	families := []string{
		"Identity and Access Management",
		"Data Protection",
		"Logging and Monitoring",
	}

	switch {
	case strings.Contains(lower, "container"):
		return append(families,
			"Image Security",
			"Network Policy",
			"Runtime Protection",
		)
	case strings.Contains(lower, "api") ||
		strings.Contains(lower, "gateway"):
		return append(families,
			"Input Validation",
			"Rate Limiting",
			"Transport Security",
		)
	case strings.Contains(lower, "storage") ||
		strings.Contains(lower, "database"):
		return append(families,
			"Encryption",
			"Backup and Recovery",
			"Access Control",
		)
	default:
		return append(families,
			"Configuration Management",
			"Network Security",
			"Vulnerability Management",
		)
	}
}

func renderMetadataPreview(
	result *WizardPromptResult,
	artifactType string,
) string {
	var lines []string
	lines = append(lines,
		"    metadata:",
	)
	lines = append(lines, fmt.Sprintf(
		"      id: %s", result.IDPrefix,
	))
	lines = append(lines, fmt.Sprintf(
		"      type: %s", artifactType,
	))
	lines = append(lines,
		"      gemara-version: \"v0.20.0\"",
	)
	if result.Description != "" {
		lines = append(lines, fmt.Sprintf(
			"      description: %s", result.Description,
		))
	}
	lines = append(lines,
		"      version: 1.0.0",
	)
	lines = append(lines,
		"      author:",
	)
	if result.AuthorID != "" {
		lines = append(lines, fmt.Sprintf(
			"        id: %s", result.AuthorID,
		))
	}
	if result.AuthorName != "" {
		lines = append(lines, fmt.Sprintf(
			"        name: %s", result.AuthorName,
		))
	}
	lines = append(lines,
		"        type: Software Assisted",
	)
	if result.CatalogRef != "" {
		lines = append(lines,
			"      mapping-references:",
		)
		lines = append(lines, fmt.Sprintf(
			"        - id: %s", result.CatalogRef,
		))
		if result.CatalogURL != "" {
			lines = append(lines, fmt.Sprintf(
				"          url: %s", result.CatalogURL,
			))
		}
	}
	lines = append(lines, fmt.Sprintf(
		"    title: %s Security %s",
		result.Component, artifactType,
	))

	return faintStyle.Render(
		strings.Join(lines, "\n"),
	)
}

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
			annotationLabelStyle.Render("Catalog: ")+
				result.CatalogRef,
		)
	}
	if result.AuthorName != "" {
		lines = append(lines,
			annotationLabelStyle.Render("Author: ")+
				result.AuthorName+" ("+
				result.AuthorID+")",
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
	if len(result.GuidelineFrameworks) > 0 {
		lines = append(lines,
			annotationLabelStyle.Render(
				"Guidelines: ",
			)+strings.Join(
				result.GuidelineFrameworks, ", ",
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

	// OpenCode handoff instructions.
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"Next: Run in OpenCode",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		"The wizard prompts are MCP protocol "+
			"messages that require OpenCode as "+
			"the client. Copy and paste the "+
			"following into OpenCode:",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, codeBlockStyle.Render(
		result.LaunchCommand,
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		"If OpenCode is not running, start it "+
			"from your project directory:",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, codeBlockStyle.Render(
		"opencode",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"OpenCode will invoke the MCP server's "+
			result.WizardName+" prompt, guide "+
			"you through each step interactively, "+
			"and validate the final artifact using "+
			"the validate_gemara_artifact tool.",
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
