// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hbraswelrh/pacman/internal/cli"
	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/team"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// teamMockPrompter implements FreeTextPrompter for team
// tests.
type teamMockPrompter struct {
	choices   []int
	texts     []string
	choiceIdx int
	textIdx   int
}

func (m *teamMockPrompter) Ask(
	_ string,
	_ []string,
) (int, error) {
	if m.choiceIdx >= len(m.choices) {
		return 0, errors.New("no more choices")
	}
	choice := m.choices[m.choiceIdx]
	m.choiceIdx++
	return choice, nil
}

func (m *teamMockPrompter) AskText(
	_ string,
) (string, error) {
	if m.textIdx >= len(m.texts) {
		return "", errors.New("no more texts")
	}
	text := m.texts[m.textIdx]
	m.textIdx++
	return text, nil
}

// T422: RunTeamSetup with 3 roles generates collaboration
// view.
func TestRunTeamSetup_ThreeRoles(t *testing.T) {
	t.Parallel()

	tutDir := filepath.Join(
		"..", "tutorials", "testdata", "valid",
	)
	teamDir := t.TempDir()

	// Prompter sequence:
	// 1. AskText: team name "GRC Team"
	// 2. AskText: member name "Alice"
	// 3. Ask: role selection (0 = Security Engineer)
	// 4. AskText: member name "Bob"
	// 5. Ask: role selection (1 = Compliance Officer)
	// 6. AskText: member name "Carol"
	// 7. Ask: role selection (3 = Developer)
	// 8. AskText: empty string to finish
	prompter := &teamMockPrompter{
		choices: []int{0, 1, 3},
		texts: []string{
			"GRC Team",
			"Alice",
			"Bob",
			"Carol",
			"", // finish adding members
		},
	}

	cfg := &cli.TeamPromptConfig{
		Prompter:      prompter,
		TutorialsDir:  tutDir,
		SchemaVersion: "v0.20.0",
		TeamConfigDir: teamDir,
	}

	var buf bytes.Buffer
	result, err := cli.RunTeamSetup(cfg, &buf)
	if err != nil {
		t.Fatalf("RunTeamSetup: %v", err)
	}

	if result.Team == nil {
		t.Fatal("expected Team in result")
	}
	if len(result.Team.Members) != 3 {
		t.Errorf(
			"Members: got %d, want 3",
			len(result.Team.Members),
		)
	}

	if result.View == nil {
		t.Fatal("expected View in result")
	}

	// Should have handoff points.
	if len(result.View.Handoffs) == 0 {
		t.Error("expected handoff points in view")
	}

	output := buf.String()
	if !strings.Contains(output, "Collaboration") {
		t.Error(
			"expected collaboration heading in output",
		)
	}
}

// T423: Adding a member updates the collaboration view.
func TestRunTeamSetup_SingleMember(t *testing.T) {
	t.Parallel()

	tutDir := filepath.Join(
		"..", "tutorials", "testdata", "valid",
	)
	teamDir := t.TempDir()

	prompter := &teamMockPrompter{
		choices: []int{0},
		texts: []string{
			"Solo Team",
			"Alice",
			"", // finish
		},
	}

	cfg := &cli.TeamPromptConfig{
		Prompter:      prompter,
		TutorialsDir:  tutDir,
		SchemaVersion: "v0.20.0",
		TeamConfigDir: teamDir,
	}

	var buf bytes.Buffer
	result, err := cli.RunTeamSetup(cfg, &buf)
	if err != nil {
		t.Fatalf("RunTeamSetup: %v", err)
	}

	if len(result.Team.Members) != 1 {
		t.Errorf(
			"Members: got %d, want 1",
			len(result.Team.Members),
		)
	}

	// Single member = no handoffs.
	if result.View != nil &&
		len(result.View.Handoffs) != 0 {
		t.Error("expected no handoffs for single member")
	}
}

// T425: RunHandoffInspection displays handoff detail.
func TestRunHandoffInspection(t *testing.T) {
	t.Parallel()

	tc := team.NewTeamConfig("Test Team")
	_ = tc.AddMember(team.TeamMember{
		Name:     "Alice",
		RoleName: consts.RoleSecurityEngineer,
		Layers:   []int{consts.LayerThreatsControls},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:     "Bob",
		RoleName: consts.RoleComplianceOfficer,
		Layers:   []int{consts.LayerRiskPolicy},
	})

	tuts := []tutorials.Tutorial{
		{
			Title:    "Control Catalog Guide",
			FilePath: "control-catalog-guide.md",
			Layer:    consts.LayerThreatsControls,
		},
		{
			Title:    "Policy Guide",
			FilePath: "policy-guide.md",
			Layer:    consts.LayerRiskPolicy,
		},
	}

	view := team.GenerateView(tc, tuts)

	if len(view.Handoffs) == 0 {
		t.Fatal("expected handoffs in view")
	}

	var buf bytes.Buffer
	err := cli.RunHandoffInspection(view, 0, &buf)
	if err != nil {
		t.Fatalf("RunHandoffInspection: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Handoff") {
		t.Error("expected handoff heading in output")
	}
}
