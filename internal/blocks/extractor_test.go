// SPDX-License-Identifier: Apache-2.0

package blocks

import (
	"path/filepath"
	"testing"

	"github.com/hbraswelrh/pacman/internal/tutorials"
)

func loadTestTuts(t *testing.T) []tutorials.Tutorial {
	t.Helper()
	dir := filepath.Join(
		"..", "tutorials", "testdata", "valid",
	)
	tuts, err := tutorials.LoadTutorials(dir)
	if err != nil {
		t.Fatalf("load tutorials: %v", err)
	}
	return tuts
}

func findTutorial(
	tuts []tutorials.Tutorial, title string,
) *tutorials.Tutorial {
	for i := range tuts {
		if tuts[i].Title == title {
			return &tuts[i]
		}
	}
	return nil
}

// T309: ExtractBlocks from Threat Assessment Guide yields
// blocks for all four sections (SC-003, US4-SC1).
func TestExtractBlocksThreatAssessment(t *testing.T) {
	t.Parallel()

	tuts := loadTestTuts(t)
	tut := findTutorial(tuts, "Threat Assessment Guide")
	if tut == nil {
		t.Fatal("Threat Assessment Guide not found")
	}

	sections, err := tutorials.ParseSections(
		tut.FilePath,
	)
	if err != nil {
		t.Fatalf("parse sections: %v", err)
	}

	blocks := ExtractBlocks(*tut, sections, "v0.18.0")

	if len(blocks) < 4 {
		t.Fatalf(
			"expected at least 4 blocks, got %d",
			len(blocks),
		)
	}

	// Verify expected sections are present.
	sectionNames := make(map[string]bool)
	for _, b := range blocks {
		sectionNames[b.SourceSection] = true
	}

	expected := []string{
		"Scope Definition",
		"Capability Identification",
		"Threat Identification",
		"CUE Validation",
	}
	for _, name := range expected {
		if !sectionNames[name] {
			t.Errorf(
				"expected block for section %q",
				name,
			)
		}
	}
}

// T310: Each extracted block has metadata.
func TestExtractBlocksMetadata(t *testing.T) {
	t.Parallel()

	tuts := loadTestTuts(t)
	tut := findTutorial(tuts, "Threat Assessment Guide")
	if tut == nil {
		t.Fatal("Threat Assessment Guide not found")
	}

	sections, err := tutorials.ParseSections(
		tut.FilePath,
	)
	if err != nil {
		t.Fatalf("parse sections: %v", err)
	}

	blocks := ExtractBlocks(*tut, sections, "v0.18.0")

	for _, b := range blocks {
		if b.SourceTutorialTitle != "Threat Assessment Guide" {
			t.Errorf(
				"expected title 'Threat Assessment "+
					"Guide', got %s",
				b.SourceTutorialTitle,
			)
		}
		if b.SchemaVersion != "v0.18.0" {
			t.Errorf(
				"expected v0.18.0, got %s",
				b.SchemaVersion,
			)
		}
		if b.Layer != 2 {
			t.Errorf(
				"expected layer 2, got %d", b.Layer,
			)
		}
		if b.ContentHash == "" {
			t.Error("expected non-empty hash")
		}
	}
}

// T311: Empty section body produces no block.
func TestExtractBlocksEmptySection(t *testing.T) {
	t.Parallel()

	tut := tutorials.Tutorial{
		Title:         "Test",
		FilePath:      "/test.md",
		Layer:         1,
		SchemaVersion: "v0.18.0",
	}

	sections := []tutorials.SectionContent{
		{Heading: "Filled", Body: "Some content."},
		{Heading: "Empty", Body: ""},
		{Heading: "Whitespace", Body: "   \n\t  "},
	}

	blocks := ExtractBlocks(tut, sections, "v0.18.0")

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].SourceSection != "Filled" {
		t.Errorf(
			"expected 'Filled', got %s",
			blocks[0].SourceSection,
		)
	}
}

// Tailored Policy Writing tutorial extracts 11 blocks
// with correct categories.
func TestExtractBlocks_TailoredPolicyWriting(
	t *testing.T,
) {
	t.Parallel()

	tuts := loadTestTuts(t)
	tut := findTutorial(tuts, "Tailored Policy Writing")
	if tut == nil {
		t.Fatal("Tailored Policy Writing not found")
	}

	sections, err := tutorials.ParseSections(
		tut.FilePath,
	)
	if err != nil {
		t.Fatalf("parse sections: %v", err)
	}

	blocks := ExtractBlocks(*tut, sections, "v0.20.0")

	if len(blocks) != 11 {
		t.Fatalf(
			"expected 11 blocks, got %d", len(blocks),
		)
	}

	// Verify category assignments for key sections.
	catMap := make(map[string]BlockCategory)
	for _, b := range blocks {
		catMap[b.SourceSection] = b.Category
	}

	// "Metadata and Naming Conventions" should be
	// NamingConv (contains "naming" and "convention").
	if catMap["Metadata and Naming Conventions"] !=
		NamingConv {
		t.Errorf(
			"Metadata and Naming Conventions: "+
				"expected NamingConv, got %s",
			catMap["Metadata and Naming Conventions"],
		)
	}

	// "RACI Contacts Structure" should be SchemaStruct
	// (contains "structure").
	if catMap["RACI Contacts Structure"] !=
		SchemaStruct {
		t.Errorf(
			"RACI Contacts Structure: expected "+
				"SchemaStruct, got %s",
			catMap["RACI Contacts Structure"],
		)
	}

	// "CUE Validation" should be ValidationStep.
	if catMap["CUE Validation"] != ValidationStep {
		t.Errorf(
			"CUE Validation: expected "+
				"ValidationStep, got %s",
			catMap["CUE Validation"],
		)
	}

	// "Cross-References to Other Layers" should be
	// CrossRef.
	if catMap["Cross-References to Other Layers"] !=
		CrossRef {
		t.Errorf(
			"Cross-References: expected "+
				"CrossRef, got %s",
			catMap["Cross-References to Other Layers"],
		)
	}

	// All blocks should have Layer 3.
	for _, b := range blocks {
		if b.Layer != 3 {
			t.Errorf(
				"block %q: expected layer 3, got %d",
				b.SourceSection, b.Layer,
			)
		}
	}
}

// T312: ExtractAll processes multiple tutorials and returns
// manifest.
func TestExtractAll(t *testing.T) {
	t.Parallel()

	tuts := loadTestTuts(t)
	dir := filepath.Join(
		"..", "tutorials", "testdata", "valid",
	)

	blocks, manifest, err := ExtractAll(
		tuts, dir, "v0.18.0",
	)
	if err != nil {
		t.Fatalf("extract all: %v", err)
	}

	if len(blocks) == 0 {
		t.Fatal("expected blocks to be extracted")
	}
	if manifest == nil {
		t.Fatal("expected non-nil manifest")
	}
	if len(manifest.Tutorials) == 0 {
		t.Error("expected manifest to have entries")
	}
	if manifest.SchemaVersion != "v0.18.0" {
		t.Errorf(
			"expected v0.18.0, got %s",
			manifest.SchemaVersion,
		)
	}
}

// T313: Block categories assigned correctly.
func TestCategorizeSectionKeywords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		heading  string
		expected BlockCategory
	}{
		{"CUE Validation", ValidationStep},
		{"Scope Definition", Pattern},
		{"Naming Conventions", NamingConv},
		{"Schema Structure", SchemaStruct},
		{"Cross-References", CrossRef},
		{"Mapping Documents", CrossRef},
		{"Metadata Setup", Pattern},
		{"Capability Identification", Pattern},
	}

	for _, tt := range tests {
		got := CategorizeSection(tt.heading)
		if got != tt.expected {
			t.Errorf(
				"CategorizeSection(%q): "+
					"expected %s, got %s",
				tt.heading, tt.expected, got,
			)
		}
	}
}
