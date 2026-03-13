// SPDX-License-Identifier: Apache-2.0

// Package fallback provides local alternatives for Gemara MCP
// server capabilities when the server is unavailable. This
// includes bundled lexicon data, local CUE validation, and
// cached schema documentation.
package fallback

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// LexiconEntry represents a single term in the Gemara lexicon.
type LexiconEntry struct {
	Term       string `yaml:"term"`
	Definition string `yaml:"definition"`
}

// Lexicon holds the bundled Gemara lexicon data.
type Lexicon struct {
	Entries []LexiconEntry `yaml:"entries"`
}

// LexiconLoader abstracts lexicon data loading for testing.
type LexiconLoader func() ([]byte, error)

// LoadBundledLexicon loads the bundled lexicon data using the
// provided loader and returns parsed lexicon entries.
func LoadBundledLexicon(
	loader LexiconLoader,
) (*Lexicon, error) {
	data, err := loader()
	if err != nil {
		return nil, fmt.Errorf("load lexicon data: %w", err)
	}

	var lexicon Lexicon
	if err := yaml.Unmarshal(data, &lexicon); err != nil {
		return nil, fmt.Errorf("parse lexicon: %w", err)
	}

	if len(lexicon.Entries) == 0 {
		return nil, fmt.Errorf(
			"bundled lexicon contains no entries",
		)
	}

	return &lexicon, nil
}
