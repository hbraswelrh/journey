// SPDX-License-Identifier: Apache-2.0

package tutorials

import (
	"os"
	"path/filepath"
	"testing"
)

// T212: LoadTutorials from valid directory returns structured
// tutorial index with titles, layers, and section headings.
func TestLoadTutorialsValidDir(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("testdata", "valid")

	tutorials, err := LoadTutorials(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tutorials) == 0 {
		t.Fatal("expected tutorials to be loaded")
	}

	// Verify at least one tutorial has all required
	// fields.
	found := false
	for _, tut := range tutorials {
		if tut.Title != "" &&
			tut.Layer > 0 &&
			len(tut.Sections) > 0 &&
			tut.SchemaVersion != "" {
			found = true
			break
		}
	}

	if !found {
		t.Error(
			"expected at least one tutorial with " +
				"title, layer, sections, and " +
				"schema_version",
		)
	}

	// Verify specific tutorials are present.
	titles := make(map[string]bool)
	for _, tut := range tutorials {
		titles[tut.Title] = true
	}

	expected := []string{
		"Threat Assessment Guide",
		"Guidance Catalog Guide",
		"Policy Guide",
		"Control Catalog Guide",
	}
	for _, title := range expected {
		if !titles[title] {
			t.Errorf(
				"expected tutorial %q to be loaded",
				title,
			)
		}
	}
}

// T213: LoadTutorials from empty directory returns empty list
// with no error.
func TestLoadTutorialsEmptyDir(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("testdata", "empty")

	tutorials, err := LoadTutorials(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tutorials) != 0 {
		t.Errorf(
			"expected empty list, got %d tutorials",
			len(tutorials),
		)
	}
}

// T214: LoadTutorials from nonexistent directory returns
// informative error with expected path and resolution
// guidance.
func TestLoadTutorialsNonexistentDir(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("testdata", "nonexistent")

	tutorials, err := LoadTutorials(dir)
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
	if tutorials != nil {
		t.Error("expected nil tutorials on error")
	}

	errMsg := err.Error()
	if !contains(errMsg, "not found") {
		t.Errorf(
			"error should mention 'not found': %s",
			errMsg,
		)
	}
	if !contains(errMsg, "clone") ||
		!contains(errMsg, "tutorials path") {
		t.Errorf(
			"error should include resolution "+
				"guidance: %s",
			errMsg,
		)
	}
}

// T215: LoadTutorials detects tutorials referencing schemas
// unavailable in the selected version.
func TestCheckVersionCompat(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("testdata", "valid")

	tutorials, err := LoadTutorials(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Select v0.18.0 — the Control Catalog Guide at
	// v0.20.0 should be flagged.
	mismatches := CheckVersionCompat(
		tutorials, "v0.18.0",
	)

	if len(mismatches) == 0 {
		t.Fatal("expected at least one version mismatch")
	}

	found := false
	for _, mm := range mismatches {
		if mm.Tutorial.Title == "Control Catalog Guide" &&
			mm.TutorialVersion == "v0.20.0" {
			found = true
			break
		}
	}

	if !found {
		t.Error(
			"expected Control Catalog Guide (v0.20.0) " +
				"to be flagged as mismatched against " +
				"v0.18.0",
		)
	}
}

// LoadTutorials correctly parses layer numbers.
func TestLoadTutorialsLayerParsing(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("testdata", "valid")

	tutorials, err := LoadTutorials(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	layerMap := make(map[string]int)
	for _, tut := range tutorials {
		layerMap[tut.Title] = tut.Layer
	}

	checks := map[string]int{
		"Threat Assessment Guide": 2,
		"Guidance Catalog Guide":  1,
		"Policy Guide":            3,
		"Control Catalog Guide":   2,
	}

	for title, expectedLayer := range checks {
		if layer, ok := layerMap[title]; ok {
			if layer != expectedLayer {
				t.Errorf(
					"%s: expected layer %d, got %d",
					title, expectedLayer, layer,
				)
			}
		}
	}
}

// LoadTutorials populates file paths.
func TestLoadTutorialsFilePaths(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("testdata", "valid")

	tutorials, err := LoadTutorials(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, tut := range tutorials {
		if tut.FilePath == "" {
			t.Errorf(
				"tutorial %s has empty FilePath",
				tut.Title,
			)
		}
		if _, err := os.Stat(tut.FilePath); err != nil {
			t.Errorf(
				"tutorial %s FilePath is not "+
					"accessible: %v",
				tut.Title, err,
			)
		}
	}
}

// Helper: contains checks if a substring is present.
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		containsString(s, substr)
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
