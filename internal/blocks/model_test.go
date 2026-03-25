// SPDX-License-Identifier: Apache-2.0

package blocks

import (
	"testing"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
)

// T301: ContentBlock contains required fields.
func TestContentBlockRequiredFields(t *testing.T) {
	t.Parallel()

	block := NewBlock(
		"/path/to/tutorial.md",
		"Threat Assessment Guide",
		"Scope Definition",
		"v0.18.0",
		consts.LayerThreatsControls,
		Pattern,
		"Define the scope of your assessment.",
	)

	if block.ID == "" {
		t.Error("expected non-empty ID")
	}
	if block.SourceTutorialPath == "" {
		t.Error("expected non-empty SourceTutorialPath")
	}
	if block.SourceTutorialTitle == "" {
		t.Error("expected non-empty SourceTutorialTitle")
	}
	if block.SourceSection == "" {
		t.Error("expected non-empty SourceSection")
	}
	if block.SchemaVersion == "" {
		t.Error("expected non-empty SchemaVersion")
	}
	if block.Layer == 0 {
		t.Error("expected non-zero Layer")
	}
	if block.Category == "" {
		t.Error("expected non-empty Category")
	}
	if block.Body == "" {
		t.Error("expected non-empty Body")
	}
	if block.ContentHash == "" {
		t.Error("expected non-empty ContentHash")
	}
	if block.ExtractedAt.IsZero() {
		t.Error("expected non-zero ExtractedAt")
	}
}

// T302: BlockCategory validates five categories.
func TestBlockCategoryValidation(t *testing.T) {
	t.Parallel()

	valid := []BlockCategory{
		Pattern,
		ValidationStep,
		NamingConv,
		SchemaStruct,
		CrossRef,
	}

	for _, cat := range valid {
		if !IsValidCategory(cat) {
			t.Errorf(
				"expected %s to be valid", cat,
			)
		}
	}

	if IsValidCategory("bogus") {
		t.Error("expected 'bogus' to be invalid")
	}

	if len(ValidCategories) != 5 {
		t.Errorf(
			"expected 5 categories, got %d",
			len(ValidCategories),
		)
	}
}

// T303: ComputeHash is deterministic SHA-256.
func TestComputeHashDeterministic(t *testing.T) {
	t.Parallel()

	body := "Define the scope of your assessment."
	h1 := ComputeHash(body)
	h2 := ComputeHash(body)

	if h1 != h2 {
		t.Errorf(
			"hashes differ: %s vs %s", h1, h2,
		)
	}
	if len(h1) != 64 {
		t.Errorf(
			"expected 64-char hex, got %d chars",
			len(h1),
		)
	}

	// Different input produces different hash.
	h3 := ComputeHash("Something else")
	if h1 == h3 {
		t.Error("different inputs produced same hash")
	}
}

// T304: Manifest maps tutorial paths to block entries.
func TestManifestStructure(t *testing.T) {
	t.Parallel()

	m := NewManifest("v0.18.0")

	if m.Tutorials == nil {
		t.Fatal("expected non-nil Tutorials map")
	}
	if m.SchemaVersion != "v0.18.0" {
		t.Errorf(
			"expected v0.18.0, got %s",
			m.SchemaVersion,
		)
	}
	if m.ExtractedAt.IsZero() {
		t.Error("expected non-zero ExtractedAt")
	}

	// Add entries.
	m.Tutorials["/path/tutorial.md"] = []ManifestEntry{
		{
			BlockID:     "Guide/Scope",
			Section:     "Scope",
			ContentHash: "abc123",
		},
	}

	entries := m.Tutorials["/path/tutorial.md"]
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].BlockID != "Guide/Scope" {
		t.Errorf(
			"expected 'Guide/Scope', got %s",
			entries[0].BlockID,
		)
	}
}

// NewBlock generates correct ID format.
func TestNewBlockIDFormat(t *testing.T) {
	t.Parallel()

	block := NewBlock(
		"/p", "My Tutorial", "My Section",
		"v1.0.0", 2, Pattern, "body",
	)

	expected := "My Tutorial/My Section"
	if block.ID != expected {
		t.Errorf(
			"expected ID %q, got %q",
			expected, block.ID,
		)
	}
}

// T302: BlockIndex lookups by ID, layer, and category.
func TestBlockIndexLookups(t *testing.T) {
	t.Parallel()

	b1 := NewBlock(
		"/a.md", "TutA", "Scope",
		"v0.18.0", 1, Pattern, "scope body",
	)
	b2 := NewBlock(
		"/a.md", "TutA", "Validation",
		"v0.18.0", 1, ValidationStep, "val body",
	)
	b3 := NewBlock(
		"/b.md", "TutB", "Structure",
		"v0.18.0", 2, SchemaStruct, "struct body",
	)

	idx := NewBlockIndex([]ContentBlock{b1, b2, b3})

	// ByID.
	got := idx.ByID("TutA/Scope")
	if got == nil {
		t.Fatal("expected block for TutA/Scope")
	}
	if got.Category != Pattern {
		t.Errorf("expected Pattern, got %s", got.Category)
	}

	if idx.ByID("nonexistent") != nil {
		t.Error("expected nil for nonexistent ID")
	}

	// ByLayer.
	l1 := idx.ByLayer(1)
	if len(l1) != 2 {
		t.Errorf("expected 2 layer-1 blocks, got %d",
			len(l1))
	}
	l2 := idx.ByLayer(2)
	if len(l2) != 1 {
		t.Errorf("expected 1 layer-2 block, got %d",
			len(l2))
	}
	l3 := idx.ByLayer(3)
	if len(l3) != 0 {
		t.Errorf("expected 0 layer-3 blocks, got %d",
			len(l3))
	}

	// ByCategory.
	patterns := idx.ByCategory(Pattern)
	if len(patterns) != 1 {
		t.Errorf("expected 1 Pattern block, got %d",
			len(patterns))
	}
	vals := idx.ByCategory(ValidationStep)
	if len(vals) != 1 {
		t.Errorf("expected 1 ValidationStep, got %d",
			len(vals))
	}

	// All.
	all := idx.All()
	if len(all) != 3 {
		t.Errorf("expected 3 total blocks, got %d",
			len(all))
	}
}
