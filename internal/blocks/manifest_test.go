// SPDX-License-Identifier: Apache-2.0

package blocks

import (
	"os"
	"path/filepath"
	"testing"
)

// T313: SaveManifest and LoadManifest round-trip.
func TestManifestSaveLoadRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")

	m := NewManifest("v0.18.0")
	m.Tutorials["/tut/a.md"] = []ManifestEntry{
		{
			BlockID:     "TutA/Scope",
			Section:     "Scope",
			ContentHash: "abc123",
		},
		{
			BlockID:     "TutA/Validation",
			Section:     "Validation",
			ContentHash: "def456",
		},
	}

	if err := SaveManifest(path, m); err != nil {
		t.Fatalf("save: %v", err)
	}

	// File should exist.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("manifest file not created: %v", err)
	}

	loaded, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.SchemaVersion != "v0.18.0" {
		t.Errorf(
			"expected v0.18.0, got %s",
			loaded.SchemaVersion,
		)
	}

	entries := loaded.Tutorials["/tut/a.md"]
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d",
			len(entries))
	}
	if entries[0].BlockID != "TutA/Scope" {
		t.Errorf("expected TutA/Scope, got %s",
			entries[0].BlockID)
	}
	if entries[1].ContentHash != "def456" {
		t.Errorf("expected def456, got %s",
			entries[1].ContentHash)
	}
}

// T314: LoadManifest returns empty for nonexistent file.
func TestLoadManifestNonexistent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nope.yaml")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil manifest")
	}
	if len(m.Tutorials) != 0 {
		t.Errorf("expected empty tutorials, got %d",
			len(m.Tutorials))
	}
}

// T316: DetectDrift finds added, modified, and removed
// blocks.
func TestDetectDrift(t *testing.T) {
	t.Parallel()

	// Previous manifest: two blocks.
	prev := NewManifest("v0.18.0")
	prev.Tutorials["/tut/a.md"] = []ManifestEntry{
		{
			BlockID:     "TutA/Scope",
			Section:     "Scope",
			ContentHash: "hash_old_scope",
		},
		{
			BlockID:     "TutA/Removed",
			Section:     "Removed",
			ContentHash: "hash_removed",
		},
	}

	// Current extraction: modified Scope, removed
	// "Removed", added "New".
	current := NewBlockIndex([]ContentBlock{
		{
			ID:                 "TutA/Scope",
			ContentHash:        "hash_new_scope",
			SourceTutorialPath: "/tut/a.md",
		},
		{
			ID:                 "TutA/New",
			ContentHash:        "hash_new",
			SourceTutorialPath: "/tut/a.md",
		},
	})

	results := DetectDrift(current, prev)

	// Expect 3 drift results: modified, removed, added.
	if len(results) != 3 {
		t.Fatalf("expected 3 drift results, got %d",
			len(results))
	}

	types := make(map[DriftType]int)
	for _, r := range results {
		types[r.Type]++
	}

	if types[DriftModified] != 1 {
		t.Errorf("expected 1 modified, got %d",
			types[DriftModified])
	}
	if types[DriftRemoved] != 1 {
		t.Errorf("expected 1 removed, got %d",
			types[DriftRemoved])
	}
	if types[DriftAdded] != 1 {
		t.Errorf("expected 1 added, got %d",
			types[DriftAdded])
	}
}

// T317: No drift when nothing changed.
func TestDetectDriftNone(t *testing.T) {
	t.Parallel()

	hash := ComputeHash("body content")

	prev := NewManifest("v0.18.0")
	prev.Tutorials["/tut/a.md"] = []ManifestEntry{
		{
			BlockID:     "TutA/Scope",
			Section:     "Scope",
			ContentHash: hash,
		},
	}

	current := NewBlockIndex([]ContentBlock{
		{
			ID:                 "TutA/Scope",
			ContentHash:        hash,
			SourceTutorialPath: "/tut/a.md",
		},
	})

	results := DetectDrift(current, prev)

	if len(results) != 0 {
		t.Errorf("expected 0 drift results, got %d",
			len(results))
	}
}

// T318: Integration test — extract, save, modify, re-extract,
// detect drift.
func TestDriftDetectionIntegration(t *testing.T) {
	t.Parallel()

	// Copy testdata to temp dir so we can modify.
	srcDir := filepath.Join(
		"..", "tutorials", "testdata", "valid",
	)
	tmpDir := t.TempDir()

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Fatalf("read src dir: %v", err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(
			filepath.Join(srcDir, e.Name()),
		)
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		if err := os.WriteFile(
			filepath.Join(tmpDir, e.Name()),
			data, 0o644,
		); err != nil {
			t.Fatalf("write %s: %v", e.Name(), err)
		}
	}

	// First extraction.
	tuts, err := loadTutsFromDir(tmpDir)
	if err != nil {
		t.Fatalf("load tutorials: %v", err)
	}

	blocks1, manifest1, err := ExtractAll(
		tuts, tmpDir, "v0.18.0",
	)
	if err != nil {
		t.Fatalf("extract all 1: %v", err)
	}
	if len(blocks1) == 0 {
		t.Fatal("expected blocks from first extraction")
	}

	// Save manifest.
	manifestPath := filepath.Join(
		tmpDir, "manifest.yaml",
	)
	if err := SaveManifest(
		manifestPath, manifest1,
	); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	// Modify a tutorial section.
	threatPath := filepath.Join(
		tmpDir, "threat-assessment-guide.md",
	)
	data, err := os.ReadFile(threatPath)
	if err != nil {
		t.Fatalf("read threat file: %v", err)
	}
	modified := string(data)
	modified = replaceSection(
		modified,
		"## Scope Definition",
		"## Scope Definition\n\n"+
			"UPDATED: Define the scope with new "+
			"criteria.\n",
	)
	if err := os.WriteFile(
		threatPath, []byte(modified), 0o644,
	); err != nil {
		t.Fatalf("write modified: %v", err)
	}

	// Re-extract.
	tuts2, err := loadTutsFromDir(tmpDir)
	if err != nil {
		t.Fatalf("reload tutorials: %v", err)
	}
	blocks2, _, err := ExtractAll(
		tuts2, tmpDir, "v0.18.0",
	)
	if err != nil {
		t.Fatalf("extract all 2: %v", err)
	}

	// Load previous manifest and detect drift.
	prevManifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	idx2 := NewBlockIndex(blocks2)
	drifts := DetectDrift(idx2, prevManifest)

	// At least one modified drift (the Scope section).
	foundModified := false
	for _, d := range drifts {
		if d.Type == DriftModified &&
			d.BlockID ==
				"Threat Assessment Guide/"+
					"Scope Definition" {
			foundModified = true
		}
	}
	if !foundModified {
		t.Error("expected modified drift for " +
			"Scope Definition")
	}
}

// replaceSection replaces a section heading line and
// following content.
func replaceSection(
	content, heading, replacement string,
) string {
	idx := findSectionIndex(content, heading)
	if idx < 0 {
		return content
	}

	// Find next section or end.
	rest := content[idx+len(heading):]
	nextIdx := findNextSection(rest)
	if nextIdx < 0 {
		return content[:idx] + replacement
	}
	return content[:idx] + replacement +
		rest[nextIdx:]
}

func findSectionIndex(content, heading string) int {
	idx := 0
	for {
		i := indexOf(content[idx:], heading)
		if i < 0 {
			return -1
		}
		return idx + i
	}
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func findNextSection(s string) int {
	for i := 1; i+3 <= len(s); i++ {
		if s[i] == '\n' && i+4 <= len(s) &&
			s[i+1:i+4] == "## " {
			return i + 1
		}
	}
	return -1
}
