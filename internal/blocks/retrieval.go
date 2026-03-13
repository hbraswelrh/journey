// SPDX-License-Identifier: Apache-2.0

package blocks

import (
	"fmt"
	"sort"
	"strings"
)

// RetrievalResult pairs a content block with adaptation
// instructions and a relevance score.
type RetrievalResult struct {
	// Block is the matched content block.
	Block ContentBlock
	// AdaptationInstructions explains how to adapt this
	// block to the user's context.
	AdaptationInstructions string
	// RelevanceScore ranks this result (higher = more
	// relevant).
	RelevanceScore int
}

// RetrieveBlocks filters blocks by the given layers and
// scores them by relevance to the goal keywords. Returns
// results sorted by relevance (highest first). Returns nil
// if layers is empty.
func RetrieveBlocks(
	index *BlockIndex,
	layers []int,
	goalKeywords []string,
) []RetrievalResult {
	if len(layers) == 0 {
		return nil
	}

	// Collect blocks matching any of the requested layers.
	layerSet := make(map[int]bool)
	for _, l := range layers {
		layerSet[l] = true
	}

	var results []RetrievalResult
	for _, b := range index.All() {
		if !layerSet[b.Layer] {
			continue
		}

		score := scoreBlock(b, goalKeywords)
		instructions := GenerateAdaptation(
			&b,
			strings.Join(goalKeywords, " "),
		)

		results = append(results, RetrievalResult{
			Block:                  b,
			AdaptationInstructions: instructions,
			RelevanceScore:         score,
		})
	}

	// Sort by relevance descending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore >
			results[j].RelevanceScore
	})

	return results
}

// scoreBlock computes a relevance score for a block based on
// how well its category and content match the goal keywords.
func scoreBlock(
	b ContentBlock,
	goalKeywords []string,
) int {
	score := 1 // Base score for layer match.

	if len(goalKeywords) == 0 {
		return score
	}

	lower := strings.ToLower(
		string(b.Category) + " " +
			b.SourceSection + " " +
			b.Body,
	)

	for _, kw := range goalKeywords {
		kwLower := strings.ToLower(kw)
		if strings.Contains(lower, kwLower) {
			score += 2
		}
	}

	return score
}

// categoryActionVerbs maps block categories to action verbs
// for adaptation instructions.
var categoryActionVerbs = map[BlockCategory]string{
	Pattern: "Adapt this pattern to define",
	ValidationStep: "Follow these validation steps " +
		"to verify",
	NamingConv: "Apply these naming conventions " +
		"when creating",
	SchemaStruct: "Use this schema structure as a " +
		"template for",
	CrossRef: "Use these cross-referencing " +
		"techniques to connect",
}

// GenerateAdaptation produces adaptation instructions for a
// content block based on its category and the user's stated
// goal.
func GenerateAdaptation(
	block *ContentBlock,
	goal string,
) string {
	verb, ok := categoryActionVerbs[block.Category]
	if !ok {
		verb = "Apply the concepts from this block to"
	}

	goalText := "your context"
	if goal != "" {
		goalText = goal
	}

	return fmt.Sprintf(
		"%s %s. Source: %s (Layer %d, %s).",
		verb,
		goalText,
		block.SourceTutorialTitle,
		block.Layer,
		block.SchemaVersion,
	)
}
