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
					Title:         "Threat Assessment Guide",
					FilePath:      "../../internal/tutorials/testdata/valid/threat-assessment-guide.md",
					Layer:         2,
					SchemaVersion: "v0.20.0",
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
					Title:         "Guidance Catalog Guide",
					FilePath:      "../../internal/tutorials/testdata/valid/guidance-catalog-guide.md",
					Layer:         1,
					SchemaVersion: "v0.20.0",
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

// TestRunTutorialPlayer_SelectAndWalk walks through all
// sections with focused questions at each step.
func TestRunTutorialPlayer_SelectAndWalk(t *testing.T) {
	t.Parallel()

	// For each section: question + follow-ups + nav
	prompter := &tutorialMockPrompter{
		choices: []int{
			0, // Select first tutorial
			// Section 1 (Scope Definition):
			0, // Answer focused question (no follow-up)
			0, // Continue to next section
			// Section 2 (Capability Identification):
			0, // "Import from FINOS CCC Core"
			0, // Follow-up: "also define custom?" -> Yes
			0, // Continue to next section
			// Section 3 (Threat Identification):
			0, // "Check for imported threats"
			0, // Follow-up: "also custom threats?" -> Yes
			0, // Follow-up: MITRE ATT&CK -> Yes
			0, // Continue to next section
			// Section 4 (CUE Validation):
			0, // Answer focused question (no follow-up)
			// Last section nav: [Go back, Mark complete,
			//   Back to list]
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

	if !result.CompletedSteps[0] {
		t.Error("expected step 0 to be completed")
	}

	output := buf.String()
	// Should contain section headings.
	if !strings.Contains(output, "Scope Definition") {
		t.Error("expected Scope Definition in output")
	}
	// Should contain focused questions.
	if !strings.Contains(output, "component") ||
		!strings.Contains(output, "assess") {
		t.Error("expected focused question about scope")
	}
	// Should show completion.
	if !strings.Contains(output, "Completed") {
		t.Error("expected completion message")
	}
}

// TestRunTutorialPlayer_BackWithoutComplete exits without
// marking complete.
func TestRunTutorialPlayer_BackWithoutComplete(
	t *testing.T,
) {
	t.Parallel()

	prompter := &tutorialMockPrompter{
		choices: []int{
			0, // Select first tutorial
			0, // Answer focused question
			// Section 1 nav: [Continue, Back to list]
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
	result, err := cli.RunTutorialPlayer(cfg, &buf)
	if err != nil {
		t.Fatalf("RunTutorialPlayer: %v", err)
	}

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
// questions and context.
func TestRunTutorialPlayer_RoleContext(t *testing.T) {
	t.Parallel()

	prompter := &tutorialMockPrompter{
		choices: []int{
			0, // Select first tutorial
			0, // Answer focused question
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
				"personalized questions",
		)
	}
}

// TestSplitSectionBody verifies progressive disclosure.
func TestSplitSectionBody(t *testing.T) {
	t.Parallel()

	body := "First paragraph intro.\n\n" +
		"Second paragraph with details.\n\n" +
		"Third paragraph with more."

	intro, detail := cli.SplitSectionBody(body)
	if intro != "First paragraph intro." {
		t.Errorf("intro = %q", intro)
	}
	if !strings.Contains(detail, "Second paragraph") {
		t.Errorf("detail = %q", detail)
	}
}

// TestSplitSectionBody_SingleParagraph returns no detail.
func TestSplitSectionBody_SingleParagraph(t *testing.T) {
	t.Parallel()

	intro, detail := cli.SplitSectionBody(
		"Just one paragraph.",
	)
	if intro != "Just one paragraph." {
		t.Errorf("intro = %q", intro)
	}
	if detail != "" {
		t.Errorf("expected empty detail, got %q", detail)
	}
}
