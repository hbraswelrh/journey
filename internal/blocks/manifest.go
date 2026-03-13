// SPDX-License-Identifier: Apache-2.0

package blocks

import (
	"fmt"
	"os"

	"github.com/hbraswelrh/pacman/internal/tutorials"
	"gopkg.in/yaml.v3"
)

// DriftType classifies how a content block changed between
// extractions.
type DriftType string

const (
	// DriftAdded means the block exists in the current
	// extraction but not in the previous manifest.
	DriftAdded DriftType = "added"
	// DriftModified means the block's content hash
	// changed between extractions.
	DriftModified DriftType = "modified"
	// DriftRemoved means the block was in the previous
	// manifest but not in the current extraction.
	DriftRemoved DriftType = "removed"
)

// DriftResult records a single block's drift status.
type DriftResult struct {
	// BlockID identifies the affected block.
	BlockID string
	// Type is the kind of drift detected.
	Type DriftType
	// OldHash is the content hash from the manifest
	// (empty for DriftAdded).
	OldHash string
	// NewHash is the content hash from the current
	// extraction (empty for DriftRemoved).
	NewHash string
}

// SaveManifest writes a manifest to a YAML file.
func SaveManifest(path string, m *Manifest) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(
		path, data, 0o644,
	); err != nil {
		return fmt.Errorf(
			"write manifest %s: %w", path, err,
		)
	}
	return nil
}

// LoadManifest reads a manifest from a YAML file. Returns an
// empty manifest (not an error) if the file does not exist.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{
				Tutorials: make(
					map[string][]ManifestEntry,
				),
			}, nil
		}
		return nil, fmt.Errorf(
			"read manifest %s: %w", path, err,
		)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf(
			"unmarshal manifest %s: %w", path, err,
		)
	}
	if m.Tutorials == nil {
		m.Tutorials = make(
			map[string][]ManifestEntry,
		)
	}
	return &m, nil
}

// DetectDrift compares a current extraction (BlockIndex)
// against a previous manifest and returns drift results for
// any added, modified, or removed blocks.
func DetectDrift(
	current *BlockIndex,
	previous *Manifest,
) []DriftResult {
	var results []DriftResult

	// Build a lookup of previous block hashes.
	prevHashes := make(map[string]string)
	for _, entries := range previous.Tutorials {
		for _, e := range entries {
			prevHashes[e.BlockID] = e.ContentHash
		}
	}

	// Build a set of current block IDs.
	currentIDs := make(map[string]bool)
	for _, b := range current.All() {
		currentIDs[b.ID] = true

		oldHash, existed := prevHashes[b.ID]
		if !existed {
			results = append(results, DriftResult{
				BlockID: b.ID,
				Type:    DriftAdded,
				NewHash: b.ContentHash,
			})
		} else if oldHash != b.ContentHash {
			results = append(results, DriftResult{
				BlockID: b.ID,
				Type:    DriftModified,
				OldHash: oldHash,
				NewHash: b.ContentHash,
			})
		}
	}

	// Check for removed blocks.
	for _, entries := range previous.Tutorials {
		for _, e := range entries {
			if !currentIDs[e.BlockID] {
				results = append(results, DriftResult{
					BlockID: e.BlockID,
					Type:    DriftRemoved,
					OldHash: e.ContentHash,
				})
			}
		}
	}

	return results
}

// loadTutsFromDir is a helper that loads tutorials from a
// directory.
func loadTutsFromDir(
	dir string,
) ([]tutorials.Tutorial, error) {
	return tutorials.LoadTutorials(dir)
}
