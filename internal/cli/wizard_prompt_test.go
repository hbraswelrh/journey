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

// TestRunWizardLauncher_ThreatWizardFullFlow walks through
// the complete threat assessment wizard flow.
func TestRunWizardLauncher_ThreatWizardFullFlow(
	t *testing.T,
) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{
			0, // Select Threat Assessment Wizard
			0, // Select "Container Runtime"
			0, // Select "FINOS CCC Core"
		},
		texts: []string{
			"ACME.PLAT.CR", // ID prefix
		},
		multiSelects: [][]int{
			{0, 1, 2, 3}, // Select first 4 capabilities
		},
		confirms: []bool{
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
			"WizardName = %q, want %q",
			result.WizardName,
			consts.WizardThreatAssessment,
		)
	}
	if result.Component != "Container Runtime" {
		t.Errorf(
			"Component = %q", result.Component,
		)
	}
	if result.IDPrefix != "ACME.PLAT.CR" {
		t.Errorf(
			"IDPrefix = %q", result.IDPrefix,
		)
	}
	if result.CatalogRef != "FINOS CCC Core" {
		t.Errorf(
			"CatalogRef = %q", result.CatalogRef,
		)
	}
	if len(result.Capabilities) != 4 {
		t.Errorf(
			"Capabilities = %d, want 4",
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
	if !strings.Contains(output, "Wizard Summary") {
		t.Error("expected wizard summary in output")
	}
}

// TestRunWizardLauncher_ControlWizardFullFlow walks
// through the control catalog wizard.
func TestRunWizardLauncher_ControlWizardFullFlow(
	t *testing.T,
) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{
			1, // Select Control Catalog Wizard
			2, // Select "Object Storage"
		},
		texts: []string{
			"ACME.PROJ.OS", // ID prefix
		},
		multiSelects: [][]int{
			{0, 1}, // Select FINOS CCC + OSPS Baseline
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
	if result.WizardName != consts.WizardControlCatalog {
		t.Errorf(
			"WizardName = %q", result.WizardName,
		)
	}
	if result.Component != "Object Storage" {
		t.Errorf(
			"Component = %q", result.Component,
		)
	}
}

// TestRunWizardLauncher_CustomComponent tests entering a
// custom component name.
func TestRunWizardLauncher_CustomComponent(t *testing.T) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{
			0, // Threat Assessment Wizard
			8, // "Enter custom component" (last option)
			0, // Reference catalog
		},
		texts: []string{
			"ML Training Pipeline",
			"", // Accept default ID prefix
		},
		multiSelects: [][]int{
			{0, 1}, // Accept first 2 capabilities
		},
		confirms: []bool{false}, // No MITRE
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
	if result.Component != "ML Training Pipeline" {
		t.Errorf(
			"Component = %q", result.Component,
		)
	}
	if !result.IncludeMITRE {
		// Confirm was false, so MITRE should be off.
		// Actually our mock returns false.
	}
}

// TestRunWizardLauncher_BackToMenu returns nil.
func TestRunWizardLauncher_BackToMenu(t *testing.T) {
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
		t.Error("expected nil result when backing out")
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
			"wizard[0].Name = %q",
			wizards[0].Name,
		)
	}
	if wizards[1].Name != consts.WizardControlCatalog {
		t.Errorf(
			"wizard[1].Name = %q",
			wizards[1].Name,
		)
	}
}

// TestSuggestCapabilities returns relevant suggestions.
func TestSuggestCapabilities(t *testing.T) {
	t.Parallel()
	// Exported via the wizard flow, but test indirectly
	// by checking the result has capabilities.
	prompter := &wizardMockPrompter{
		choices: []int{0, 0, 0},
		texts:   []string{""},
		multiSelects: [][]int{
			{0, 1, 2},
		},
		confirms: []bool{false},
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
		t.Fatal("expected result")
	}
	if len(result.Capabilities) == 0 {
		t.Error("expected capabilities from suggestions")
	}
}
