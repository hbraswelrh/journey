// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/hbraswelrh/pacman/internal/cli"
	"github.com/hbraswelrh/pacman/internal/consts"
)

// wizardMockPrompter provides predefined choices for wizard
// testing.
type wizardMockPrompter struct {
	choices   []int
	texts     []string
	choiceIdx int
	textIdx   int
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
		return "", errors.New("no more texts")
	}
	text := m.texts[m.textIdx]
	m.textIdx++
	return text, nil
}

// TestRunWizardLauncher_SelectAndCollectArgs verifies that
// selecting a wizard and providing arguments produces a
// launch command.
func TestRunWizardLauncher_SelectAndCollectArgs(
	t *testing.T,
) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{0}, // Select first wizard
		texts: []string{
			"container runtime",
			"ACME.PLAT.CR",
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
	if result.Component != "container runtime" {
		t.Errorf(
			"Component = %q", result.Component,
		)
	}
	if result.IDPrefix != "ACME.PLAT.CR" {
		t.Errorf(
			"IDPrefix = %q", result.IDPrefix,
		)
	}
	if result.LaunchCommand == "" {
		t.Error("expected non-empty LaunchCommand")
	}

	output := buf.String()
	if !strings.Contains(output, "container runtime") {
		t.Error(
			"output should mention component name",
		)
	}
	if !strings.Contains(output, "ACME.PLAT.CR") {
		t.Error(
			"output should mention ID prefix",
		)
	}
}

// TestRunWizardLauncher_MCPUnavailable shows warning when
// MCP is not connected.
func TestRunWizardLauncher_MCPUnavailable(t *testing.T) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{0},
		texts: []string{
			"API gateway",
			"ACME.PROJ.GW",
		},
	}

	cfg := &cli.WizardPromptConfig{
		Prompter:     prompter,
		MCPAvailable: false,
		RoleName:     consts.RoleSecurityEngineer,
	}

	var buf bytes.Buffer
	_, err := cli.RunWizardLauncher(cfg, &buf)
	if err != nil {
		t.Fatalf("RunWizardLauncher: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "not connected") {
		t.Error(
			"expected MCP unavailable warning",
		)
	}
}

// TestRunWizardLauncher_BackToMenu returns nil when user
// selects "Back to main menu".
func TestRunWizardLauncher_BackToMenu(t *testing.T) {
	t.Parallel()

	prompter := &wizardMockPrompter{
		choices: []int{2}, // Back to main menu (3rd option)
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
