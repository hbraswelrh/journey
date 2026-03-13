// SPDX-License-Identifier: Apache-2.0

package blocks

import (
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// T322: RetrieveBlocks returns Layer 1 blocks for guidance
// goal.
func TestRetrieveBlocksGuidanceGoal(t *testing.T) {
	t.Parallel()

	blocks := []ContentBlock{
		NewBlock(
			"/a.md", "Guidance Catalog Guide",
			"Creating a Guidance Catalog",
			"v0.18.0", consts.LayerGuidance,
			Pattern, "Create a guidance catalog.",
		),
		NewBlock(
			"/a.md", "Guidance Catalog Guide",
			"Cross-References",
			"v0.18.0", consts.LayerGuidance,
			CrossRef, "Link entries to frameworks.",
		),
		NewBlock(
			"/b.md", "Threat Assessment Guide",
			"Scope Definition",
			"v0.18.0", consts.LayerThreatsControls,
			Pattern, "Define the scope.",
		),
	}
	idx := NewBlockIndex(blocks)

	results := RetrieveBlocks(
		idx,
		[]int{consts.LayerGuidance},
		[]string{"guidance", "create"},
	)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d",
			len(results))
	}

	// All results should be Layer 1.
	for _, r := range results {
		if r.Block.Layer != consts.LayerGuidance {
			t.Errorf(
				"expected layer %d, got %d",
				consts.LayerGuidance,
				r.Block.Layer,
			)
		}
	}

	// Each result should have adaptation instructions.
	for _, r := range results {
		if r.AdaptationInstructions == "" {
			t.Errorf(
				"expected adaptation instructions "+
					"for block %s",
				r.Block.ID,
			)
		}
	}
}

// T322b: RetrieveBlocks with empty layers returns no blocks.
func TestRetrieveBlocksEmptyLayers(t *testing.T) {
	t.Parallel()

	blocks := []ContentBlock{
		NewBlock(
			"/a.md", "Guide", "Section",
			"v0.18.0", 1, Pattern, "body",
		),
	}
	idx := NewBlockIndex(blocks)

	results := RetrieveBlocks(idx, nil, nil)

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d",
			len(results))
	}
}

// T322c: RetrieveBlocks with no matching goal still returns
// layer-matched blocks.
func TestRetrieveBlocksNoGoalKeywords(t *testing.T) {
	t.Parallel()

	blocks := []ContentBlock{
		NewBlock(
			"/a.md", "Guide", "Section",
			"v0.18.0", 2, Pattern, "body",
		),
	}
	idx := NewBlockIndex(blocks)

	results := RetrieveBlocks(
		idx, []int{2}, nil,
	)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d",
			len(results))
	}
}

// T323: Blocks sorted by relevance (category match to goal
// keywords scored higher).
func TestRetrieveBlocksSortedByRelevance(t *testing.T) {
	t.Parallel()

	blocks := []ContentBlock{
		NewBlock(
			"/a.md", "Guide", "Cross-References",
			"v0.18.0", 1, CrossRef, "xref body",
		),
		NewBlock(
			"/a.md", "Guide",
			"Creating a Guidance Catalog",
			"v0.18.0", 1, Pattern, "pattern body",
		),
		NewBlock(
			"/a.md", "Guide", "CUE Validation",
			"v0.18.0", 1, ValidationStep,
			"validation body",
		),
	}
	idx := NewBlockIndex(blocks)

	// Goal keyword "validation" should boost the
	// validation_step block.
	results := RetrieveBlocks(
		idx, []int{1}, []string{"validation"},
	)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d",
			len(results))
	}

	// The validation block should be first (highest
	// relevance score).
	if results[0].Block.Category != ValidationStep {
		t.Errorf(
			"expected first result to be "+
				"ValidationStep, got %s",
			results[0].Block.Category,
		)
	}
}

// T321: GenerateAdaptation produces meaningful instructions.
func TestGenerateAdaptation(t *testing.T) {
	t.Parallel()

	block := NewBlock(
		"/a.md", "Guide", "Scope Definition",
		"v0.18.0", 2, Pattern,
		"Define the scope of your assessment.",
	)

	instructions := GenerateAdaptation(
		&block, "create my own guidance document",
	)

	if instructions == "" {
		t.Error("expected non-empty instructions")
	}
}
