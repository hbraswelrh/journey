// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hbraswelrh/gemara-user-journey/internal/blocks"
	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/tutorials"
)

// BlocksConfig holds dependencies for block operations.
type BlocksConfig struct {
	// TutorialsDir is the path to the Gemara tutorials
	// directory.
	TutorialsDir string
	// SchemaVersion is the user's selected Gemara schema
	// version.
	SchemaVersion string
	// CacheDir is the directory for block cache storage.
	// Defaults to ~/.config/gemara-user-journey/blocks/.
	CacheDir string
}

// BlocksResult holds the outcome of block extraction.
type BlocksResult struct {
	// BlockCount is the number of blocks extracted.
	BlockCount int
	// DriftResults are any detected drift items.
	DriftResults []blocks.DriftResult
	// Index is the extracted block index.
	Index *blocks.BlockIndex
}

// resolveCacheDir returns the cache directory, creating it
// if needed.
func resolveCacheDir(cfg *BlocksConfig) (string, error) {
	dir := cfg.CacheDir
	if dir == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf(
				"user config dir: %w", err,
			)
		}
		dir = filepath.Join(
			configDir, consts.BlockCacheDir,
		)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf(
			"create cache dir %s: %w", dir, err,
		)
	}
	return dir, nil
}

// RunBlockExtraction loads tutorials, extracts content
// blocks, saves a manifest, and displays a summary.
func RunBlockExtraction(
	cfg *BlocksConfig,
	out io.Writer,
) (*BlocksResult, error) {
	// Load tutorials.
	tuts, err := tutorials.LoadTutorials(
		cfg.TutorialsDir,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load tutorials: %w", err,
		)
	}

	if len(tuts) == 0 {
		fmt.Fprintln(out, RenderNote(
			"No tutorials found in "+
				cfg.TutorialsDir,
		))
		return &BlocksResult{}, nil
	}

	// Extract blocks.
	allBlocks, manifest, err := blocks.ExtractAll(
		tuts, cfg.TutorialsDir, cfg.SchemaVersion,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"extract blocks: %w", err,
		)
	}

	// Save manifest.
	cacheDir, err := resolveCacheDir(cfg)
	if err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(
		cacheDir, consts.BlockManifestFile,
	)
	if err := blocks.SaveManifest(
		manifestPath, manifest,
	); err != nil {
		return nil, fmt.Errorf(
			"save manifest: %w", err,
		)
	}

	// Build summary counts.
	byCat := make(map[string]int)
	byLayer := make(map[int]int)
	for _, b := range allBlocks {
		byCat[string(b.Category)]++
		byLayer[b.Layer]++
	}

	// Render summary.
	RenderBlockSummary(
		len(allBlocks), byCat, byLayer, out,
	)

	fmt.Fprintln(out, RenderSuccess(fmt.Sprintf(
		"Manifest saved to %s", manifestPath,
	)))

	idx := blocks.NewBlockIndex(allBlocks)
	return &BlocksResult{
		BlockCount: len(allBlocks),
		Index:      idx,
	}, nil
}

// RunDriftCheck loads the previous manifest, re-extracts
// blocks, and detects changes.
func RunDriftCheck(
	cfg *BlocksConfig,
	out io.Writer,
) ([]blocks.DriftResult, error) {
	// Load previous manifest.
	cacheDir, err := resolveCacheDir(cfg)
	if err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(
		cacheDir, consts.BlockManifestFile,
	)
	prevManifest, err := blocks.LoadManifest(
		manifestPath,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load manifest: %w", err,
		)
	}

	// Re-extract.
	tuts, err := tutorials.LoadTutorials(
		cfg.TutorialsDir,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load tutorials: %w", err,
		)
	}

	allBlocks, _, err := blocks.ExtractAll(
		tuts, cfg.TutorialsDir, cfg.SchemaVersion,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"extract blocks: %w", err,
		)
	}

	idx := blocks.NewBlockIndex(allBlocks)
	drifts := blocks.DetectDrift(idx, prevManifest)

	// Display results.
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)

	if len(drifts) == 0 {
		fmt.Fprintln(out, RenderSuccess(
			"No content drift detected — all "+
				"blocks are up to date",
		))
	} else {
		fmt.Fprintln(out, headingStyle.Render(
			fmt.Sprintf(
				"Content Drift Detected (%d blocks "+
					"affected)",
				len(drifts),
			),
		))
		fmt.Fprintln(out)

		for _, d := range drifts {
			fmt.Fprintln(out, RenderDriftResult(
				d.BlockID, string(d.Type),
			))
		}
		fmt.Fprintln(out)
		fmt.Fprintln(out, RenderWarning(
			"Affected blocks should be reviewed "+
				"before presenting to users.",
		))
	}

	return drifts, nil
}

// RunBlockRetrieval retrieves blocks relevant to the given
// layers and goal, displaying them with adaptation
// instructions.
func RunBlockRetrieval(
	cfg *BlocksConfig,
	layers []int,
	goal string,
	out io.Writer,
) error {
	// Load or extract blocks.
	cacheDir, err := resolveCacheDir(cfg)
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(
		cacheDir, consts.BlockManifestFile,
	)

	var idx *blocks.BlockIndex

	// Try to use existing extraction if manifest exists.
	if _, statErr := os.Stat(manifestPath); statErr == nil {
		tuts, err := tutorials.LoadTutorials(
			cfg.TutorialsDir,
		)
		if err != nil {
			return fmt.Errorf(
				"load tutorials: %w", err,
			)
		}
		allBlocks, _, err := blocks.ExtractAll(
			tuts, cfg.TutorialsDir, cfg.SchemaVersion,
		)
		if err != nil {
			return fmt.Errorf(
				"extract blocks: %w", err,
			)
		}
		idx = blocks.NewBlockIndex(allBlocks)
	} else {
		// No manifest — extract fresh.
		result, err := RunBlockExtraction(cfg, out)
		if err != nil {
			return err
		}
		idx = result.Index
	}

	if idx == nil {
		fmt.Fprintln(out, RenderNote(
			"No blocks available for retrieval.",
		))
		return nil
	}

	// Parse goal into keywords.
	var goalKeywords []string
	if goal != "" {
		// Simple keyword split.
		for _, w := range splitGoalKeywords(goal) {
			if len(w) > 2 {
				goalKeywords = append(
					goalKeywords, w,
				)
			}
		}
	}

	results := blocks.RetrieveBlocks(
		idx, layers, goalKeywords,
	)

	if len(results) == 0 {
		fmt.Fprintln(out, RenderNote(
			"No content blocks match your layers "+
				"and goal.",
		))
		return nil
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		fmt.Sprintf(
			"Relevant Content Blocks (%d found)",
			len(results),
		),
	))
	fmt.Fprintln(out)

	for _, r := range results {
		fmt.Fprintln(out, RenderContentBlock(
			r.Block.ID,
			string(r.Block.Category),
			r.Block.Layer,
			r.Block.SourceTutorialTitle,
			r.Block.SchemaVersion,
			r.Block.Body,
		))
		fmt.Fprintln(out,
			"  "+annotationLabelStyle.Render(
				"Adapt: ",
			)+
				annotationTextStyle.Render(
					r.AdaptationInstructions,
				),
		)
		fmt.Fprintln(out)
	}

	return nil
}

// splitGoalKeywords splits a goal string into individual
// words, lowercased.
func splitGoalKeywords(goal string) []string {
	var words []string
	current := ""
	for _, c := range goal {
		if c == ' ' || c == ',' || c == ';' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
