// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// T330a: RunBlockExtraction with testdata produces blocks.
func TestRunBlockExtraction(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(
		"..", "tutorials", "testdata", "valid",
	)
	cacheDir := t.TempDir()

	cfg := &BlocksConfig{
		TutorialsDir:  dir,
		SchemaVersion: "v0.18.0",
		CacheDir:      cacheDir,
	}

	var buf bytes.Buffer
	result, err := RunBlockExtraction(cfg, &buf)
	if err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	if result.BlockCount == 0 {
		t.Error("expected blocks to be extracted")
	}

	output := buf.String()
	if !strings.Contains(output, "Content Block") {
		t.Error("expected summary heading in output")
	}
	if !strings.Contains(output, "Total blocks:") {
		t.Error("expected total count in output")
	}

	// Manifest should be saved.
	manifestPath := filepath.Join(
		cacheDir, consts.BlockManifestFile,
	)
	if _, err := os.Stat(manifestPath); err != nil {
		t.Errorf("manifest not saved: %v", err)
	}
}

// T330b: RunDriftCheck with no changes reports no drift.
func TestRunDriftCheckNoDrift(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(
		"..", "tutorials", "testdata", "valid",
	)
	cacheDir := t.TempDir()

	cfg := &BlocksConfig{
		TutorialsDir:  dir,
		SchemaVersion: "v0.18.0",
		CacheDir:      cacheDir,
	}

	// First extraction to create manifest.
	var buf bytes.Buffer
	_, err := RunBlockExtraction(cfg, &buf)
	if err != nil {
		t.Fatalf("extraction: %v", err)
	}

	// Drift check — no changes.
	buf.Reset()
	drifts, err := RunDriftCheck(cfg, &buf)
	if err != nil {
		t.Fatalf("drift check: %v", err)
	}

	if len(drifts) != 0 {
		t.Errorf("expected 0 drifts, got %d",
			len(drifts))
	}

	output := buf.String()
	if !strings.Contains(output, "up to date") {
		t.Error("expected 'up to date' message")
	}
}

// T330c: RunDriftCheck with modified content detects drift.
func TestRunDriftCheckWithDrift(t *testing.T) {
	t.Parallel()

	// Copy testdata to temp dir.
	srcDir := filepath.Join(
		"..", "tutorials", "testdata", "valid",
	)
	tmpDir := t.TempDir()
	cacheDir := t.TempDir()

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Fatalf("read src dir: %v", err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(
			filepath.Join(srcDir, e.Name()),
		)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if err := os.WriteFile(
			filepath.Join(tmpDir, e.Name()),
			data, 0o644,
		); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	cfg := &BlocksConfig{
		TutorialsDir:  tmpDir,
		SchemaVersion: "v0.18.0",
		CacheDir:      cacheDir,
	}

	// Extract and save manifest.
	var buf bytes.Buffer
	_, err = RunBlockExtraction(cfg, &buf)
	if err != nil {
		t.Fatalf("extraction: %v", err)
	}

	// Modify a tutorial.
	threatPath := filepath.Join(
		tmpDir, "threat-assessment-guide.md",
	)
	data, err := os.ReadFile(threatPath)
	if err != nil {
		t.Fatalf("read threat: %v", err)
	}
	modified := strings.Replace(
		string(data),
		"Define the scope of your threat assessment.",
		"UPDATED scope definition content here.",
		1,
	)
	if err := os.WriteFile(
		threatPath, []byte(modified), 0o644,
	); err != nil {
		t.Fatalf("write modified: %v", err)
	}

	// Drift check.
	buf.Reset()
	drifts, err := RunDriftCheck(cfg, &buf)
	if err != nil {
		t.Fatalf("drift check: %v", err)
	}

	if len(drifts) == 0 {
		t.Error("expected drift to be detected")
	}

	output := buf.String()
	if !strings.Contains(output, "Drift Detected") {
		t.Error("expected drift heading in output")
	}
}

// T330d: RunBlockRetrieval with Layer 2 goal returns blocks.
func TestRunBlockRetrieval(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(
		"..", "tutorials", "testdata", "valid",
	)
	cacheDir := t.TempDir()

	cfg := &BlocksConfig{
		TutorialsDir:  dir,
		SchemaVersion: "v0.18.0",
		CacheDir:      cacheDir,
	}

	// Extract first.
	var buf bytes.Buffer
	_, err := RunBlockExtraction(cfg, &buf)
	if err != nil {
		t.Fatalf("extraction: %v", err)
	}

	// Retrieve Layer 2 blocks.
	buf.Reset()
	err = RunBlockRetrieval(
		cfg,
		[]int{consts.LayerThreatsControls},
		"threat modeling and scope",
		&buf,
	)
	if err != nil {
		t.Fatalf("retrieval: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Relevant Content") {
		t.Error("expected retrieval heading in output")
	}
	if !strings.Contains(output, "Adapt:") {
		t.Error("expected adaptation instructions")
	}
}

// T330e: Empty tutorials dir produces empty result.
func TestRunBlockExtractionEmpty(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(
		"..", "tutorials", "testdata", "empty",
	)
	cacheDir := t.TempDir()

	cfg := &BlocksConfig{
		TutorialsDir:  dir,
		SchemaVersion: "v0.18.0",
		CacheDir:      cacheDir,
	}

	var buf bytes.Buffer
	result, err := RunBlockExtraction(cfg, &buf)
	if err != nil {
		t.Fatalf("extraction: %v", err)
	}

	if result.BlockCount != 0 {
		t.Errorf("expected 0 blocks, got %d",
			result.BlockCount)
	}
}
