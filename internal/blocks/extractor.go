// SPDX-License-Identifier: Apache-2.0

package blocks

import (
	"fmt"
	"strings"

	"github.com/hbraswelrh/gemara-user-journey/internal/tutorials"
)

// categoryKeywords maps heading keywords to block categories.
// Longest matches are checked first by CategorizeSection.
var categoryKeywords = map[string]BlockCategory{
	"cross-reference": CrossRef,
	"cross reference": CrossRef,
	"mapping":         CrossRef,
	"link":            CrossRef,
	"validation":      ValidationStep,
	"cue":             ValidationStep,
	"vet":             ValidationStep,
	"naming":          NamingConv,
	"convention":      NamingConv,
	"identifier":      NamingConv,
	"schema":          SchemaStruct,
	"structure":       SchemaStruct,
	"artifact":        SchemaStruct,
}

// CategorizeSection determines the block category for a
// section based on its heading text. If no specific keyword
// matches, defaults to Pattern.
func CategorizeSection(heading string) BlockCategory {
	lower := strings.ToLower(heading)
	for keyword, cat := range categoryKeywords {
		if strings.Contains(lower, keyword) {
			return cat
		}
	}
	return Pattern
}

// ExtractBlocks creates content blocks from a tutorial's
// parsed sections. One block is created per non-empty section.
// Empty sections are skipped.
func ExtractBlocks(
	tutorial tutorials.Tutorial,
	sections []tutorials.SectionContent,
	schemaVersion string,
) []ContentBlock {
	var blocks []ContentBlock

	for _, sec := range sections {
		body := strings.TrimSpace(sec.Body)
		if body == "" {
			continue
		}

		category := CategorizeSection(sec.Heading)
		block := NewBlock(
			tutorial.FilePath,
			tutorial.Title,
			sec.Heading,
			schemaVersion,
			tutorial.Layer,
			category,
			body,
		)
		blocks = append(blocks, block)
	}

	return blocks
}

// ExtractAll processes all tutorials in a directory and
// returns the extracted blocks with a manifest for drift
// detection.
func ExtractAll(
	tuts []tutorials.Tutorial,
	dir string,
	schemaVersion string,
) ([]ContentBlock, *Manifest, error) {
	manifest := NewManifest(schemaVersion)
	var allBlocks []ContentBlock

	for _, tut := range tuts {
		sections, err := tutorials.ParseSections(
			tut.FilePath,
		)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"parse sections for %s: %w",
				tut.Title, err,
			)
		}

		blocks := ExtractBlocks(
			tut, sections, schemaVersion,
		)

		var entries []ManifestEntry
		for _, b := range blocks {
			entries = append(entries, ManifestEntry{
				BlockID:     b.ID,
				Section:     b.SourceSection,
				ContentHash: b.ContentHash,
			})
		}

		if len(entries) > 0 {
			manifest.Tutorials[tut.FilePath] = entries
		}

		allBlocks = append(allBlocks, blocks...)
	}

	return allBlocks, manifest, nil
}
