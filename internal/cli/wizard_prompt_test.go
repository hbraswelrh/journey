// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hbraswelrh/pacman/internal/cli"
	"github.com/hbraswelrh/pacman/internal/consts"
)

// wizardMockPrompter implements WizardPrompter for tests.
type wizardMockPrompter struct {
	choices      []int
	texts        []string
	multiSelects [][]int
	confirms     []bool
	choiceIdx    int
	textIdx      int
	multiIdx     int
	confirmIdx   int
}

func (m *wizardMockPrompter) Ask(
	_ string,
	opts []string,
) (int, error) {
	if m.choiceIdx >= len(m.choices) {
		return len(opts) - 1, nil
	}
	choice := m.choices[m.choiceIdx]
	m.choiceIdx++
	return choice, nil
}

func (m *wizardMockPrompter) AskText(
	_ string,
) (string, error) {
	if m.textIdx >= len(m.texts) {
		return "", nil
	}
	text := m.texts[m.textIdx]
	m.textIdx++
	return text, nil
}

func (m *wizardMockPrompter) AskMultiSelect(
	_ string,
	_ []string,
	defaults []int,
) ([]int, error) {
	if m.multiIdx >= len(m.multiSelects) {
		return defaults, nil
	}
	selected := m.multiSelects[m.multiIdx]
	m.multiIdx++
	return selected, nil
}

func (m *wizardMockPrompter) AskConfirm(
	_ string,
) (bool, error) {
	if m.confirmIdx >= len(m.confirms) {
		return true, nil
	}
	val := m.confirms[m.confirmIdx]
	m.confirmIdx++
	return val, nil
}

func (m *wizardMockPrompter) AskTextWithDefault(
	_ string,
	defaultValue string,
) (string, error) {
	if m.textIdx >= len(m.texts) {
		return defaultValue, nil
	}
	text := m.texts[m.textIdx]
	m.textIdx++
	if text == "" {
		return defaultValue, nil
	}
	return text, nil
}

// TestThreatWizard_FullFlow walks the threat assessment
// wizard with FINOS CCC import.
func TestThreatWizard_FullFlow(t *testing.T) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{
			0, // Select Threat Assessment Wizard
			0, // Yes, use FINOS CCC Core
		},
		texts: []string{
			"container runtime",      // Component name
			"Manages OCI containers", // Description
			"Alice Smith",            // Author name
			"alice.smith",            // Author ID
			"ACME.PLAT.CR",           // ID prefix
		},
		multiSelects: [][]int{
			{0, 1, 2, 3, 4}, // Capabilities
		},
		confirms: []bool{
			true, // Confirm metadata
			true, // Enable MITRE ATT&CK
		},
	}

	cfg := &cli.WizardPromptConfig{
		Prompter:     prompter,
		MCPAvailable: true,
		RoleName:     consts.RoleSecurityEngineer,
	}

	var buf bytes.Buffer
	result, err := cli.RunWizardLauncher(cfg, &buf)
	if err != nil {
		t.Fatalf("RunWizardLauncher: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.WizardName !=
		consts.WizardThreatAssessment {
		t.Errorf(
			"WizardName = %q", result.WizardName,
		)
	}
	if result.Component != "container runtime" {
		t.Errorf(
			"Component = %q", result.Component,
		)
	}
	if result.CatalogRef != "FINOS CCC Core" {
		t.Errorf(
			"CatalogRef = %q", result.CatalogRef,
		)
	}
	if result.AuthorName != "Alice Smith" {
		t.Errorf(
			"AuthorName = %q", result.AuthorName,
		)
	}
	if result.IDPrefix != "ACME.PLAT.CR" {
		t.Errorf(
			"IDPrefix = %q", result.IDPrefix,
		)
	}
	if len(result.Capabilities) != 5 {
		t.Errorf(
			"Capabilities = %d, want 5",
			len(result.Capabilities),
		)
	}
	if !result.IncludeMITRE {
		t.Error("expected IncludeMITRE = true")
	}
	if result.LaunchCommand == "" {
		t.Error("expected non-empty LaunchCommand")
	}

	output := buf.String()
	if !strings.Contains(output, "FINOS CCC Core") {
		t.Error(
			"output should mention FINOS CCC Core",
		)
	}
	if !strings.Contains(output, "Catalog Import") {
		t.Error(
			"output should contain step 1 title",
		)
	}
	if !strings.Contains(output, "Wizard Summary") {
		t.Error(
			"output should contain wizard summary",
		)
	}
}

// TestControlWizard_FullFlow walks the control catalog
// wizard with guideline framework selection.
func TestControlWizard_FullFlow(t *testing.T) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{
			1, // Select Control Catalog Wizard
			0, // Yes, use FINOS CCC Core
		},
		texts: []string{
			"API gateway",         // Component name
			"Routes API requests", // Description
			"Bob Jones",           // Author name
			"bob.jones",           // Author ID
			"ACME.PROJ.GW",        // ID prefix
		},
		multiSelects: [][]int{
			{0, 3},       // CSF + NIST 800-53
			{0, 1, 2, 3}, // All families
		},
		confirms: []bool{
			true, // Confirm metadata
		},
	}

	cfg := &cli.WizardPromptConfig{
		Prompter:     prompter,
		MCPAvailable: true,
	}

	var buf bytes.Buffer
	result, err := cli.RunWizardLauncher(cfg, &buf)
	if err != nil {
		t.Fatalf("RunWizardLauncher: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.WizardName != consts.WizardControlCatalog {
		t.Errorf(
			"WizardName = %q", result.WizardName,
		)
	}
	if result.Component != "API gateway" {
		t.Errorf(
			"Component = %q", result.Component,
		)
	}
	if len(result.GuidelineFrameworks) != 2 {
		t.Errorf(
			"GuidelineFrameworks = %d, want 2",
			len(result.GuidelineFrameworks),
		)
	}

	output := buf.String()
	if !strings.Contains(output, "Control Families") {
		t.Error(
			"output should contain families step",
		)
	}
	if !strings.Contains(
		output, "ControlCatalog",
	) {
		t.Error(
			"output should reference ControlCatalog",
		)
	}
}

// TestWizard_AlternativeCatalog tests providing a custom
// catalog URL with unverified source warning.
func TestWizard_AlternativeCatalog(t *testing.T) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{
			0, // Threat Assessment Wizard
			1, // Provide alternative catalog
		},
		texts: []string{
			"https://example.com/my-catalog.yaml",
			"web server",
			"Serves HTTP requests",
			"Carol",
			"carol",
			"ACME.WEB.SRV",
		},
		multiSelects: [][]int{
			{0, 1},
		},
		confirms: []bool{
			true,  // Confirm metadata
			false, // No MITRE
		},
	}

	cfg := &cli.WizardPromptConfig{
		Prompter:     prompter,
		MCPAvailable: true,
	}

	var buf bytes.Buffer
	result, err := cli.RunWizardLauncher(cfg, &buf)
	if err != nil {
		t.Fatalf("RunWizardLauncher: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CatalogRef != "Custom" {
		t.Errorf(
			"CatalogRef = %q", result.CatalogRef,
		)
	}

	output := buf.String()
	if !strings.Contains(output, "unverified") {
		t.Error(
			"expected unverified source warning",
		)
	}
}

// TestWizard_BackToMenu returns nil.
func TestWizard_BackToMenu(t *testing.T) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{2}, // Back to main menu
	}

	cfg := &cli.WizardPromptConfig{
		Prompter:     prompter,
		MCPAvailable: true,
	}

	var buf bytes.Buffer
	result, err := cli.RunWizardLauncher(cfg, &buf)
	if err != nil {
		t.Fatalf("RunWizardLauncher: %v", err)
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

// TestWizard_AdvisoryModeWarning verifies that when the MCP
// server is in advisory mode, the wizard warns and shows an
// advisory mode message.
func TestWizard_AdvisoryModeWarning(t *testing.T) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{2}, // Back to main menu
	}

	cfg := &cli.WizardPromptConfig{
		Prompter:     prompter,
		MCPAvailable: true,
		ServerMode:   consts.MCPModeAdvisory,
	}

	var buf bytes.Buffer
	_, err := cli.RunWizardLauncher(cfg, &buf)
	if err != nil {
		t.Fatalf("RunWizardLauncher: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "advisory") {
		t.Fatalf(
			"expected advisory mode warning, got: %s",
			output,
		)
	}
}

// TestWizard_ArtifactModeNoWarning verifies that artifact
// mode does not show the advisory mode warning.
func TestWizard_ArtifactModeNoWarning(t *testing.T) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{2}, // Back to main menu
	}

	cfg := &cli.WizardPromptConfig{
		Prompter:     prompter,
		MCPAvailable: true,
		ServerMode:   consts.MCPModeArtifact,
	}

	var buf bytes.Buffer
	_, err := cli.RunWizardLauncher(cfg, &buf)
	if err != nil {
		t.Fatalf("RunWizardLauncher: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "advisory mode") {
		t.Fatalf(
			"expected no advisory warning, got: %s",
			output,
		)
	}
}

// TestAvailableWizards returns the expected wizards.
func TestAvailableWizards(t *testing.T) {
	t.Parallel()
	wizards := cli.AvailableWizards()
	if len(wizards) != 2 {
		t.Fatalf(
			"expected 2 wizards, got %d",
			len(wizards),
		)
	}
	if wizards[0].Name != consts.WizardThreatAssessment {
		t.Errorf(
			"wizard[0].Name = %q", wizards[0].Name,
		)
	}
	if wizards[1].Name != consts.WizardControlCatalog {
		t.Errorf(
			"wizard[1].Name = %q", wizards[1].Name,
		)
	}
}
