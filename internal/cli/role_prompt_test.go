// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hbraswelrh/gemara-user-journey/internal/cli"
	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/roles"
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

// T257: Integration test — Security Engineer + CI/CD
// activities produces Layer 2 learning path.
func TestRoleDiscovery_IntegrationCICD(t *testing.T) {
	var buf bytes.Buffer

	cfg := &cli.RolePromptConfig{
		Prompter: &mockFreeTextPrompter{
			choices: []int{0}, // Security Engineer
			texts: []string{
				"CI/CD pipeline management, " +
					"dependency management, and " +
					"coding with upstream " +
					"open-source components",
			},
		},
		TutorialsDir:  filepath.Join("..", "tutorials", "testdata", "valid"),
		SchemaVersion: "v0.18.0",
	}

	result, err := cli.RunRoleDiscovery(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify profile targets Layer 2.
	if result.Profile == nil {
		t.Fatal("expected non-nil profile")
	}
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
			"expected Layer 2 for CI/CD activities, "+
				"got layers: %v",
			layers,
		)
	}

	// Verify role is Security Engineer.
	if result.Profile.Role.Name !=
		consts.RoleSecurityEngineer {
		t.Errorf(
			"expected role %s, got %s",
			consts.RoleSecurityEngineer,
			result.Profile.Role.Name,
		)
	}
}

// T258: Integration test — custom "Product Security
// Engineer" + audit activities produces Layer 1/3 path.
func TestRoleDiscovery_IntegrationAudit(t *testing.T) {
	var buf bytes.Buffer

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
				"evidence collection, audit " +
					"interviews, and defining " +
					"compliance scope",
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

	// Verify partial match was identified.
	output := buf.String()
	if !strings.Contains(output, "partially matches") {
		t.Errorf(
			"expected partial match message, got: %s",
			output,
		)
	}

	// Verify layers include 1 and/or 3.
	layers := result.Profile.UniqueLayerNumbers()
	hasL1 := false
	hasL3 := false
	for _, l := range layers {
		if l == consts.LayerGuidance {
			hasL1 = true
		}
		if l == consts.LayerRiskPolicy {
			hasL3 = true
		}
	}
	if !hasL1 && !hasL3 {
		t.Errorf(
			"expected Layers 1 and/or 3 for audit "+
				"activities, got layers: %v",
			layers,
		)
	}
}

// T014: Artifact recommendation rendering includes type
// names, descriptions, and MCP wizard names where applicable.
func TestRoleDiscovery_ArtifactRecommendationRendering(
	t *testing.T,
) {
	t.Run("L2_wizard_artifacts", func(t *testing.T) {
		var buf bytes.Buffer

		// Security Engineer with CI/CD activities
		// resolves Layer 2, producing ThreatCatalog
		// (wizard: threat_assessment) and ControlCatalog
		// (wizard: control_catalog).
		cfg := &cli.RolePromptConfig{
			Prompter: &mockFreeTextPrompter{
				choices: []int{0}, // Security Engineer
				texts: []string{
					"CI/CD pipeline management " +
						"and threat modeling",
				},
			},
			TutorialsDir: filepath.Join(
				"..", "tutorials", "testdata", "valid",
			),
			SchemaVersion: "v0.20.0",
		}

		result, err := cli.RunRoleDiscovery(cfg, &buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Profile == nil {
			t.Fatal("expected non-nil profile")
		}

		// Verify recommendations were populated.
		if len(result.Profile.Recommendations) == 0 {
			t.Fatal(
				"expected non-empty recommendations",
			)
		}

		output := buf.String()

		// Verify artifact type names appear in output.
		if !strings.Contains(
			output, consts.ArtifactThreatCatalog,
		) {
			t.Errorf(
				"expected %q in output, got: %s",
				consts.ArtifactThreatCatalog, output,
			)
		}
		if !strings.Contains(
			output, consts.ArtifactControlCatalog,
		) {
			t.Errorf(
				"expected %q in output, got: %s",
				consts.ArtifactControlCatalog, output,
			)
		}

		// Verify descriptions appear in output.
		threatDesc :=
			consts.ArtifactDescriptions[consts.ArtifactThreatCatalog]
		if !strings.Contains(output, threatDesc) {
			t.Errorf(
				"expected threat catalog description "+
					"in output, got: %s",
				output,
			)
		}
		controlDesc :=
			consts.ArtifactDescriptions[consts.ArtifactControlCatalog]
		if !strings.Contains(output, controlDesc) {
			t.Errorf(
				"expected control catalog description "+
					"in output, got: %s",
				output,
			)
		}

		// Verify MCP wizard names appear in output.
		if !strings.Contains(
			output, consts.WizardThreatAssessment,
		) {
			t.Errorf(
				"expected wizard name %q in output, "+
					"got: %s",
				consts.WizardThreatAssessment, output,
			)
		}
		if !strings.Contains(
			output, consts.WizardControlCatalog,
		) {
			t.Errorf(
				"expected wizard name %q in output, "+
					"got: %s",
				consts.WizardControlCatalog, output,
			)
		}
	})

	t.Run("L3_collaborative_artifacts", func(t *testing.T) {
		var buf bytes.Buffer

		// CISO with policy activities resolves Layer 3,
		// producing Policy recommendation — no wizard,
		// uses collaborative authoring.
		predefined := roles.PredefinedRoles()
		cisoIdx := -1
		for i, r := range predefined {
			if r.Name == consts.RoleCISO {
				cisoIdx = i
				break
			}
		}
		if cisoIdx < 0 {
			t.Fatal("CISO role not found in predefined")
		}

		cfg := &cli.RolePromptConfig{
			Prompter: &mockFreeTextPrompter{
				choices: []int{
					cisoIdx, // CISO role
					2,       // "Both layers" for
					//   ambiguous "adherence"
				},
				texts: []string{
					"create policy and timeline " +
						"for adherence",
				},
			},
			TutorialsDir: filepath.Join(
				"..", "tutorials", "testdata", "valid",
			),
			SchemaVersion: "v0.20.0",
		}

		result, err := cli.RunRoleDiscovery(cfg, &buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Profile == nil {
			t.Fatal("expected non-nil profile")
		}

		// Verify Policy recommendation exists.
		hasPolicy := false
		for _, rec := range result.Profile.Recommendations {
			if rec.ArtifactType ==
				consts.ArtifactPolicy {
				hasPolicy = true
				break
			}
		}
		if !hasPolicy {
			t.Errorf(
				"expected Policy in recommendations, "+
					"got: %v",
				result.Profile.Recommendations,
			)
		}

		output := buf.String()

		// Verify Policy artifact type appears.
		if !strings.Contains(
			output, consts.ArtifactPolicy,
		) {
			t.Errorf(
				"expected %q in output, got: %s",
				consts.ArtifactPolicy, output,
			)
		}

		// Verify Policy description appears.
		policyDesc :=
			consts.ArtifactDescriptions[consts.ArtifactPolicy]
		if !strings.Contains(output, policyDesc) {
			t.Errorf(
				"expected policy description in "+
					"output, got: %s",
				output,
			)
		}

		// Verify collaborative approach is shown
		// (Policy has no wizard).
		if !strings.Contains(
			output, "Collaborative authoring",
		) {
			t.Errorf(
				"expected 'Collaborative authoring' "+
					"in output for Policy, got: %s",
				output,
			)
		}
	})
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
