// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hbraswelrh/pacman/internal/cli"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// tutorialMockPrompter provides predefined choices for
// tutorial player testing.
type tutorialMockPrompter struct {
	choices   []int
	texts     []string
	choiceIdx int
	textIdx   int
}

func (m *tutorialMockPrompter) Ask(
	_ string,
	opts []string,
) (int, error) {
	if m.choiceIdx >= len(m.choices) {
		// Return last option (typically "Back" or "Exit")
		// to prevent infinite loops.
		return len(opts) - 1, nil
	}
	choice := m.choices[m.choiceIdx]
	m.choiceIdx++
	return choice, nil
}

func (m *tutorialMockPrompter) AskText(
	_ string,
) (string, error) {
	if m.textIdx >= len(m.texts) {
		return "", nil
	}
	text := m.texts[m.textIdx]
	m.textIdx++
	return text, nil
}

func testLearningPath() *tutorials.LearningPath {
	return &tutorials.LearningPath{
		TargetRole: "Security Engineer",
		Steps: []tutorials.PathStep{
			{
				Tutorial: tutorials.Tutorial{
					Title:    "Threat Assessment Guide",
					FilePath: "../../internal/tutorials/testdata/valid/threat-assessment-guide.md",
					Layer:    2,
					Sections: []string{
						"Scope Definition",
						"Capability Identification",
						"Threat Identification",
						"CUE Validation",
					},
				},
				Layer:         2,
				WhyAnnotation: "Essential for your work",
			},
			{
				Tutorial: tutorials.Tutorial{
					Title:    "Guidance Catalog Guide",
					FilePath: "../../internal/tutorials/testdata/valid/guidance-catalog-guide.md",
					Layer:    1,
					Sections: []string{
						"Creating a Guidance Catalog",
						"Metadata Setup",
					},
				},
				Layer:         1,
				WhyAnnotation: "Foundational knowledge",
			},
		},
		CompletedSteps: make(map[int]bool),
	}
}

// TestRunTutorialPlayer_SelectAndWalk verifies that
// selecting a tutorial walks through its sections.
func TestRunTutorialPlayer_SelectAndWalk(t *testing.T) {
	t.Parallel()

	prompter := &tutorialMockPrompter{
		choices: []int{
			0, // Select first tutorial
			0, // Next section (Scope -> Capability)
			0, // Next section (Capability -> Threat)
			0, // Next section (Threat -> CUE)
			// At last section, opts are:
			//   [Previous, Mark complete, Back]
			1, // Mark complete
			2, // Back to main menu (from step list)
		},
	}

	cfg := &cli.TutorialPromptConfig{
		Prompter:     prompter,
		LearningPath: testLearningPath(),
		TutorialsDir: "../../internal/tutorials/testdata/valid",
		RoleName:     "Security Engineer",
	}

	var buf bytes.Buffer
	result, err := cli.RunTutorialPlayer(cfg, &buf)
	if err != nil {
		t.Fatalf("RunTutorialPlayer: %v", err)
	}

	// First tutorial should be marked complete.
	if !result.CompletedSteps[0] {
		t.Error("expected step 0 to be completed")
	}

	output := buf.String()
	// Should contain section headings.
	if !strings.Contains(output, "Scope Definition") {
		t.Error("expected Scope Definition in output")
	}
	if !strings.Contains(output, "CUE Validation") {
		t.Error("expected CUE Validation in output")
	}
	// Should show completion.
	if !strings.Contains(output, "Completed") {
		t.Error("expected completion message")
	}
}

// TestRunTutorialPlayer_BackWithoutComplete verifies
// exiting a tutorial without marking it complete.
func TestRunTutorialPlayer_BackWithoutComplete(
	t *testing.T,
) {
	t.Parallel()

	prompter := &tutorialMockPrompter{
		choices: []int{
			0, // Select first tutorial
			1, // Back to tutorial list (from first
			//    section, only options are "Next" and
			//    "Back" so index 1 = Back)
			2, // Back to main menu
		},
	}

	cfg := &cli.TutorialPromptConfig{
		Prompter:     prompter,
		LearningPath: testLearningPath(),
		TutorialsDir: "../../internal/tutorials/testdata/valid",
		RoleName:     "Security Engineer",
	}

	var buf bytes.Buffer
	result, err := cli.RunTutorialPlayer(cfg, &buf)
	if err != nil {
		t.Fatalf("RunTutorialPlayer: %v", err)
	}

	// Should NOT be marked complete.
	if result.CompletedSteps[0] {
		t.Error("step 0 should not be completed")
	}
}

// TestRunTutorialPlayer_EmptyPath handles no tutorials.
func TestRunTutorialPlayer_EmptyPath(t *testing.T) {
	t.Parallel()

	cfg := &cli.TutorialPromptConfig{
		Prompter: &tutorialMockPrompter{},
		LearningPath: &tutorials.LearningPath{
			Steps:          nil,
			CompletedSteps: make(map[int]bool),
		},
		RoleName: "Security Engineer",
	}

	var buf bytes.Buffer
	result, err := cli.RunTutorialPlayer(cfg, &buf)
	if err != nil {
		t.Fatalf("RunTutorialPlayer: %v", err)
	}
	if len(result.CompletedSteps) != 0 {
		t.Error("expected no completed steps")
	}

	output := buf.String()
	if !strings.Contains(output, "No tutorials") {
		t.Error("expected no tutorials message")
	}
}

// TestRunTutorialPlayer_RoleContext verifies role-specific
// context is displayed.
func TestRunTutorialPlayer_RoleContext(t *testing.T) {
	t.Parallel()

	prompter := &tutorialMockPrompter{
		choices: []int{
			0, // Select first tutorial
			1, // Back to tutorial list
			2, // Back to main menu
		},
	}

	cfg := &cli.TutorialPromptConfig{
		Prompter:     prompter,
		LearningPath: testLearningPath(),
		TutorialsDir: "../../internal/tutorials/testdata/valid",
		RoleName:     "Security Engineer",
	}

	var buf bytes.Buffer
	_, err := cli.RunTutorialPlayer(cfg, &buf)
	if err != nil {
		t.Fatalf("RunTutorialPlayer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Security Engineer") {
		t.Error(
			"expected role name in output for " +
				"personalized context",
		)
	}
}
