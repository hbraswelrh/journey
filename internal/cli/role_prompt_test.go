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
	"github.com/hbraswelrh/pacman/internal/roles"
)

// mockFreeTextPrompter implements FreeTextPrompter for tests.
type mockFreeTextPrompter struct {
	choices   []int
	texts     []string
	choiceIdx int
	textIdx   int
}

func (m *mockFreeTextPrompter) Ask(
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

func (m *mockFreeTextPrompter) AskText(
	_ string,
) (string, error) {
	if m.textIdx >= len(m.texts) {
		return "", errors.New("no more texts")
	}
	text := m.texts[m.textIdx]
	m.textIdx++
	return text, nil
}

// T221: Selecting a predefined role proceeds to activity
// probing with that role.
func TestRoleDiscovery_SelectPredefinedRole(
	t *testing.T,
) {
	var buf bytes.Buffer

	// Select Security Engineer (index 0), then provide
	// activity description.
	cfg := &cli.RolePromptConfig{
		Prompter: &mockFreeTextPrompter{
			choices: []int{0},
			texts: []string{
				"CI/CD pipeline management and " +
					"dependency management",
			},
		},
		TutorialsDir:  filepath.Join("..", "tutorials", "testdata", "valid"),
		SchemaVersion: "v0.18.0",
	}

	result, err := cli.RunRoleDiscovery(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if result.Profile.Role == nil {
		t.Fatal("expected non-nil role in profile")
	}
	if result.Profile.Role.Name !=
		consts.RoleSecurityEngineer {
		t.Errorf(
			"expected %s, got %s",
			consts.RoleSecurityEngineer,
			result.Profile.Role.Name,
		)
	}

	output := buf.String()
	if !strings.Contains(output, "Selected role") {
		t.Errorf(
			"expected 'Selected role' in output, "+
				"got: %s",
			output,
		)
	}
}

// T222: Selecting "My role isn't listed" accepts free-text
// input and shows partial matches.
func TestRoleDiscovery_CustomRolePartialMatch(
	t *testing.T,
) {
	var buf bytes.Buffer

	// Select "My role isn't listed" (last option = 7),
	// enter "Product Security Engineer", then provide
	// activities.
	predefined := roles.PredefinedRoles()
	customIdx := len(predefined)

	cfg := &cli.RolePromptConfig{
		Prompter: &mockFreeTextPrompter{
			choices: []int{
				customIdx, // "My role isn't listed"
				2,         // "Both layers" for
				//   ambiguous "evidence collection"
			},
			texts: []string{
				"Product Security Engineer",
				"evidence collection and audit " +
					"interviews",
			},
		},
		TutorialsDir:  filepath.Join("..", "tutorials", "testdata", "valid"),
		SchemaVersion: "v0.18.0",
	}

	result, err := cli.RunRoleDiscovery(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Profile == nil {
		t.Fatal("expected non-nil profile")
	}

	output := buf.String()
	if !strings.Contains(output, "partially matches") {
		t.Errorf(
			"expected partial match message, got: %s",
			output,
		)
	}
}

// T223: Entering a custom role with no partial match proceeds
// with extracted keywords only.
func TestRoleDiscovery_CustomRoleNoMatch(t *testing.T) {
	var buf bytes.Buffer

	predefined := roles.PredefinedRoles()
	customIdx := len(predefined)

	cfg := &cli.RolePromptConfig{
		Prompter: &mockFreeTextPrompter{
			choices: []int{
				customIdx, // "My role isn't listed"
				0,         // Category selection
			},
			texts: []string{
				"Underwater Basket Weaver",
				"", // Empty description -> categories
			},
		},
		TutorialsDir:  filepath.Join("..", "tutorials", "testdata", "valid"),
		SchemaVersion: "v0.18.0",
	}

	result, err := cli.RunRoleDiscovery(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Profile == nil {
		t.Fatal("expected non-nil profile")
	}

	output := buf.String()
	if !strings.Contains(
		output, "No predefined role",
	) {
		t.Errorf(
			"expected no-match message, got: %s",
			output,
		)
	}
}

// T233: No recognizable keywords in activity description
// presents full activity category list.
func TestRoleDiscovery_NoCategoryFallback(t *testing.T) {
	var buf bytes.Buffer

	cfg := &cli.RolePromptConfig{
		Prompter: &mockFreeTextPrompter{
			choices: []int{
				0, // Security Engineer
				0, // First category
			},
			texts: []string{
				"I like to go hiking on weekends",
			},
		},
		TutorialsDir:  filepath.Join("..", "tutorials", "testdata", "valid"),
		SchemaVersion: "v0.18.0",
	}

	result, err := cli.RunRoleDiscovery(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Profile == nil {
		t.Fatal("expected non-nil profile")
	}

	output := buf.String()
	if !strings.Contains(
		output, "narrow down",
	) {
		t.Errorf(
			"expected category fallback message, "+
				"got: %s",
			output,
		)
	}
}

// T234: "Secure Software Development professional" extracts
// "SDLC" keyword.
func TestRoleDiscovery_SDLCKeywordExtraction(
	t *testing.T,
) {
	var buf bytes.Buffer

	predefined := roles.PredefinedRoles()
	customIdx := len(predefined)

	cfg := &cli.RolePromptConfig{
		Prompter: &mockFreeTextPrompter{
			choices: []int{customIdx},
			texts: []string{
				"Secure Software Development " +
					"professional",
				"SDLC and threat modeling work",
			},
		},
		TutorialsDir:  filepath.Join("..", "tutorials", "testdata", "valid"),
		SchemaVersion: "v0.18.0",
	}

	result, err := cli.RunRoleDiscovery(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Profile == nil {
		t.Fatal("expected non-nil profile")
	}

	// Check that Layer 2 was resolved.
	layers := result.Profile.UniqueLayerNumbers()
	hasL2 := false
	for _, l := range layers {
		if l == consts.LayerThreatsControls {
			hasL2 = true
			break
		}
	}
	if !hasL2 {
		t.Errorf(
			"expected Layer 2 from SDLC keywords, "+
				"got layers: %v",
			layers,
		)
	}
}

// Tutorials are loaded and version mismatches detected.
func TestRoleDiscovery_LoadsTutorials(t *testing.T) {
	var buf bytes.Buffer

	cfg := &cli.RolePromptConfig{
		Prompter: &mockFreeTextPrompter{
			choices: []int{0},
			texts: []string{
				"CI/CD pipeline management",
			},
		},
		TutorialsDir:  filepath.Join("..", "tutorials", "testdata", "valid"),
		SchemaVersion: "v0.18.0",
	}

	result, err := cli.RunRoleDiscovery(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tutorials) == 0 {
		t.Error("expected tutorials to be loaded")
	}

	// Control Catalog Guide is at v0.20.0, should
	// mismatch v0.18.0.
	if len(result.VersionMismatches) == 0 {
		t.Error("expected version mismatches")
	}
}

// Ambiguous keywords trigger clarification question.
func TestRoleDiscovery_AmbiguousKeywords(t *testing.T) {
	var buf bytes.Buffer

	cfg := &cli.RolePromptConfig{
		Prompter: &mockFreeTextPrompter{
			choices: []int{
				0, // Security Engineer
				2, // "Both layers" for clarification
			},
			texts: []string{
				"evidence collection for my team",
			},
		},
		TutorialsDir:  filepath.Join("..", "tutorials", "testdata", "valid"),
		SchemaVersion: "v0.18.0",
	}

	result, err := cli.RunRoleDiscovery(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Profile == nil {
		t.Fatal("expected non-nil profile")
	}

	output := buf.String()
	if !strings.Contains(output, "multiple") {
		t.Errorf(
			"expected ambiguity message, got: %s",
			output,
		)
	}
}
