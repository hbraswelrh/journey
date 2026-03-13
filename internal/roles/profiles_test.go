// SPDX-License-Identifier: Apache-2.0

package roles

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// T246: SaveProfile writes a valid YAML file.
func TestSaveProfile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	profile := &RoleProfile{
		Name: "Custom Security Lead",
		ActivityKeywords: []string{
			"threat modeling", "CI/CD",
		},
		LayerMappings: []int{
			consts.LayerThreatsControls,
			consts.LayerGuidance,
		},
		Description: "A custom security leadership role",
		CreatedAt:   time.Now(),
	}

	if err := SaveProfile(dir, profile); err != nil {
		t.Fatalf("save profile: %v", err)
	}

	// Verify file was created.
	expectedFile := filepath.Join(
		dir, "custom-security-lead.yaml",
	)
	if _, err := os.Stat(expectedFile); err != nil {
		t.Fatalf(
			"expected file %s to exist: %v",
			expectedFile, err,
		)
	}
}

// T247: LoadProfile reads a saved profile correctly.
func TestLoadProfile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	original := &RoleProfile{
		Name: "Test Role",
		ActivityKeywords: []string{
			"evidence collection", "audit",
		},
		LayerMappings: []int{
			consts.LayerEvaluation,
			consts.LayerRiskPolicy,
		},
		Description: "A test role for unit testing",
		CreatedAt:   time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC),
	}

	if err := SaveProfile(dir, original); err != nil {
		t.Fatalf("save: %v", err)
	}

	path := filepath.Join(dir, "test-role.yaml")
	loaded, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf(
			"Name: expected %s, got %s",
			original.Name, loaded.Name,
		)
	}
	if len(loaded.ActivityKeywords) !=
		len(original.ActivityKeywords) {
		t.Errorf(
			"ActivityKeywords: expected %d, got %d",
			len(original.ActivityKeywords),
			len(loaded.ActivityKeywords),
		)
	}
	if len(loaded.LayerMappings) !=
		len(original.LayerMappings) {
		t.Errorf(
			"LayerMappings: expected %d, got %d",
			len(original.LayerMappings),
			len(loaded.LayerMappings),
		)
	}
	if loaded.Description != original.Description {
		t.Errorf(
			"Description: expected %s, got %s",
			original.Description, loaded.Description,
		)
	}
}

// T248: ListProfiles returns all saved profiles.
func TestListProfiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	profiles := []*RoleProfile{
		{
			Name:             "Role A",
			ActivityKeywords: []string{"ci/cd"},
			LayerMappings: []int{
				consts.LayerThreatsControls,
			},
		},
		{
			Name:             "Role B",
			ActivityKeywords: []string{"audit"},
			LayerMappings: []int{
				consts.LayerEvaluation,
			},
		},
	}

	for _, p := range profiles {
		if err := SaveProfile(dir, p); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	listed, err := ListProfiles(dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(listed) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(listed))
	}
}

// T249: Saved custom profiles appear in role selection list
// via MergeWithPredefined.
func TestMergeWithPredefined(t *testing.T) {
	t.Parallel()

	predefined := PredefinedRoles()
	custom := []RoleProfile{
		{
			Name: "Custom DevSecOps",
			ActivityKeywords: []string{
				"CI/CD", "SDLC",
			},
			LayerMappings: []int{
				consts.LayerThreatsControls,
			},
		},
	}

	merged := MergeWithPredefined(predefined, custom)

	if len(merged) != len(predefined)+1 {
		t.Fatalf(
			"expected %d roles, got %d",
			len(predefined)+1, len(merged),
		)
	}

	// Last role should be the custom one.
	last := merged[len(merged)-1]
	if last.Name != "Custom DevSecOps" {
		t.Errorf(
			"expected 'Custom DevSecOps', got %s",
			last.Name,
		)
	}
	if last.Source != SourceCustom {
		t.Error("expected SourceCustom for merged role")
	}
}

// ListProfiles on nonexistent directory returns nil.
func TestListProfilesNonexistent(t *testing.T) {
	t.Parallel()

	profiles, err := ListProfiles(
		"/tmp/nonexistent-pacman-test-dir",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profiles != nil {
		t.Error("expected nil for nonexistent directory")
	}
}

// ProfileFromActivityProfile converts correctly.
func TestProfileFromActivityProfile(t *testing.T) {
	t.Parallel()

	role := &Role{
		Name:        "Test Role",
		Description: "A test role",
	}
	ap := &ActivityProfile{
		ExtractedKeywords: []string{"ci/cd", "sdlc"},
		ResolvedLayers: []LayerMapping{
			{
				Layer:      consts.LayerThreatsControls,
				Confidence: ConfidenceStrong,
			},
		},
		Role: role,
	}

	rp := ProfileFromActivityProfile(ap)

	if rp.Name != role.Name {
		t.Errorf(
			"expected name %s, got %s",
			role.Name, rp.Name,
		)
	}
	if len(rp.ActivityKeywords) != 2 {
		t.Errorf(
			"expected 2 keywords, got %d",
			len(rp.ActivityKeywords),
		)
	}
	if len(rp.LayerMappings) != 1 {
		t.Errorf(
			"expected 1 layer mapping, got %d",
			len(rp.LayerMappings),
		)
	}
}
