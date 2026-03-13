// SPDX-License-Identifier: Apache-2.0

// Package blocks implements reusable content block extraction,
// drift detection, persistence, and querying for the Pac-Man
// tutorial engine (US4).
package blocks

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// BlockCategory classifies a content block by its function.
type BlockCategory string

const (
	Pattern        BlockCategory = BlockCategory(consts.CategoryPattern)
	ValidationStep BlockCategory = BlockCategory(consts.CategoryValidationStep)
	NamingConv     BlockCategory = BlockCategory(consts.CategoryNamingConv)
	SchemaStruct   BlockCategory = BlockCategory(consts.CategorySchemaStruct)
	CrossRef       BlockCategory = BlockCategory(consts.CategoryCrossRef)
)

// ValidCategories is the set of recognized block categories.
var ValidCategories = []BlockCategory{
	Pattern,
	ValidationStep,
	NamingConv,
	SchemaStruct,
	CrossRef,
}

// IsValidCategory returns true if the category is recognized.
func IsValidCategory(c BlockCategory) bool {
	for _, v := range ValidCategories {
		if v == c {
			return true
		}
	}
	return false
}

// ContentBlock is a modular, reusable unit of knowledge
// extracted from a Gemara tutorial section.
type ContentBlock struct {
	// ID uniquely identifies this block. Format:
	// "<tutorial-title>/<section-heading>".
	ID string `yaml:"id"`
	// SourceTutorialPath is the file path of the source
	// tutorial.
	SourceTutorialPath string `yaml:"source_tutorial_path"`
	// SourceTutorialTitle is the display title of the
	// source tutorial.
	SourceTutorialTitle string `yaml:"source_tutorial_title"`
	// SourceSection is the heading of the section this
	// block was extracted from.
	SourceSection string `yaml:"source_section"`
	// SchemaVersion is the Gemara schema version at the
	// time of extraction.
	SchemaVersion string `yaml:"schema_version"`
	// Layer is the Gemara layer number (1-7).
	Layer int `yaml:"layer"`
	// Category classifies the block's function.
	Category BlockCategory `yaml:"category"`
	// Body is the full text content of the block.
	Body string `yaml:"body"`
	// ContentHash is the SHA-256 hex digest of Body,
	// used for drift detection.
	ContentHash string `yaml:"content_hash"`
	// ExtractedAt is when this block was extracted.
	ExtractedAt time.Time `yaml:"extracted_at"`
}

// ManifestEntry records a single block's identity and hash
// within the manifest.
type ManifestEntry struct {
	// BlockID is the content block ID.
	BlockID string `yaml:"block_id"`
	// Section is the source section heading.
	Section string `yaml:"section"`
	// ContentHash is the SHA-256 at extraction time.
	ContentHash string `yaml:"content_hash"`
}

// Manifest tracks the extraction state for drift detection.
// It maps tutorial paths to their extracted block entries.
type Manifest struct {
	// Tutorials maps tutorial file paths to their block
	// entries.
	Tutorials map[string][]ManifestEntry `yaml:"tutorials"`
	// SchemaVersion is the schema version used during
	// extraction.
	SchemaVersion string `yaml:"schema_version"`
	// ExtractedAt is when the extraction was performed.
	ExtractedAt time.Time `yaml:"extracted_at"`
}

// NewManifest creates an empty manifest.
func NewManifest(schemaVersion string) *Manifest {
	return &Manifest{
		Tutorials:     make(map[string][]ManifestEntry),
		SchemaVersion: schemaVersion,
		ExtractedAt:   time.Now(),
	}
}

// ComputeHash returns the SHA-256 hex digest of the given
// content body. The hash is deterministic for identical input.
func ComputeHash(body string) string {
	h := sha256.Sum256([]byte(body))
	return fmt.Sprintf("%x", h)
}

// NewBlock creates a ContentBlock with a computed hash and
// the current timestamp.
func NewBlock(
	sourcePath string,
	sourceTitle string,
	section string,
	schemaVersion string,
	layer int,
	category BlockCategory,
	body string,
) ContentBlock {
	return ContentBlock{
		ID:                  sourceTitle + "/" + section,
		SourceTutorialPath:  sourcePath,
		SourceTutorialTitle: sourceTitle,
		SourceSection:       section,
		SchemaVersion:       schemaVersion,
		Layer:               layer,
		Category:            category,
		Body:                body,
		ContentHash:         ComputeHash(body),
		ExtractedAt:         time.Now(),
	}
}
