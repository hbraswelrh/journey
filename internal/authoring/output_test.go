// SPDX-License-Identifier: Apache-2.0

package authoring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"gopkg.in/yaml.v3"
)

// T533: RenderYAML produces valid YAML with correct
// structure for a ThreatCatalog artifact.
func TestRenderYAML(t *testing.T) {
	t.Parallel()
	artifact := testArtifactComplete()

	data, err := RenderYAML(artifact)
	if err != nil {
		t.Fatalf("RenderYAML error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("RenderYAML produced empty output")
	}

	// Verify it is valid YAML.
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid YAML: %v", err)
	}

	// Should contain metadata section.
	if _, ok := parsed[consts.SectionMetadata]; !ok {
		t.Error(
			"YAML output missing metadata section",
		)
	}
}

// T534: RenderJSON produces valid JSON with correct
// structure for a ThreatCatalog artifact.
func TestRenderJSON(t *testing.T) {
	t.Parallel()
	artifact := testArtifactComplete()

	data, err := RenderJSON(artifact)
	if err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("RenderJSON produced empty output")
	}

	// Verify it is valid JSON.
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Should contain metadata section.
	if _, ok := parsed[consts.SectionMetadata]; !ok {
		t.Error(
			"JSON output missing metadata section",
		)
	}
}

// T535: GenerateFilename produces Gemara-convention
// filename.
func TestGenerateFilename(t *testing.T) {
	t.Parallel()
	artifact := testArtifactComplete()

	filename := GenerateFilename(artifact)
	if filename == "" {
		t.Fatal("GenerateFilename produced empty string")
	}

	// Should contain the artifact name from metadata
	// (which follows Gemara naming conventions with the
	// type abbreviation embedded, e.g., THR for threats).
	if !strings.Contains(filename, "THR") {
		t.Errorf(
			"filename should contain artifact type "+
				"abbreviation, got %q",
			filename,
		)
	}

	// Should have .yaml extension by default.
	if !strings.HasSuffix(filename, ".yaml") {
		t.Errorf(
			"filename should end with .yaml, got %q",
			filename,
		)
	}
}

// T536: WriteArtifact creates a file at the expected path.
func TestWriteArtifact(t *testing.T) {
	t.Parallel()
	artifact := testArtifactComplete()

	tmpDir := t.TempDir()
	path, err := WriteArtifact(
		artifact, tmpDir, consts.DefaultArtifactFormat,
	)
	if err != nil {
		t.Fatalf("WriteArtifact error: %v", err)
	}
	if path == "" {
		t.Fatal("WriteArtifact returned empty path")
	}

	// File should exist.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}

	// Verify content is valid YAML.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid YAML: %v", err)
	}
}

// T537: WriteArtifact with format "json" writes JSON.
func TestWriteArtifactJSON(t *testing.T) {
	t.Parallel()
	artifact := testArtifactComplete()

	tmpDir := t.TempDir()
	path, err := WriteArtifact(
		artifact, tmpDir, "json",
	)
	if err != nil {
		t.Fatalf("WriteArtifact error: %v", err)
	}

	// Should have .json extension.
	if !strings.HasSuffix(path, ".json") {
		t.Errorf(
			"path should end with .json, got %q",
			path,
		)
	}

	// Verify content is valid JSON.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

// T538: RenderYAML output round-trips through YAML
// parse/marshal without data loss.
func TestRenderYAMLRoundTrip(t *testing.T) {
	t.Parallel()
	artifact := testArtifactComplete()

	data, err := RenderYAML(artifact)
	if err != nil {
		t.Fatalf("RenderYAML error: %v", err)
	}

	// Parse and re-marshal.
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	remarshaled, err := yaml.Marshal(parsed)
	if err != nil {
		t.Fatalf("re-marshal YAML: %v", err)
	}

	// Parse both and compare.
	var original, roundTripped map[string]interface{}
	if err := yaml.Unmarshal(
		data, &original,
	); err != nil {
		t.Fatalf("parse original: %v", err)
	}
	if err := yaml.Unmarshal(
		remarshaled, &roundTripped,
	); err != nil {
		t.Fatalf("parse round-tripped: %v", err)
	}

	// Check metadata section preserved.
	origMeta, ok1 := original[consts.SectionMetadata]
	rtMeta, ok2 := roundTripped[consts.SectionMetadata]
	if !ok1 || !ok2 {
		t.Fatal("metadata section missing after round-trip")
	}
	origMap := origMeta.(map[string]interface{})
	rtMap := rtMeta.(map[string]interface{})
	if origMap["name"] != rtMap["name"] {
		t.Errorf(
			"name field lost in round-trip: "+
				"%v vs %v",
			origMap["name"], rtMap["name"],
		)
	}
}

// TestGenerateFilenameWithName verifies that GenerateFilename
// uses the artifact name field when available.
func TestGenerateFilenameWithName(t *testing.T) {
	t.Parallel()
	artifact := testArtifactComplete()

	filename := GenerateFilename(artifact)
	// Should incorporate the name from metadata.
	if !strings.Contains(filename, "ACME.WEB.THR01") {
		t.Errorf(
			"filename should contain artifact name, "+
				"got %q",
			filename,
		)
	}
}

// TestWriteArtifactCreatesDir verifies that WriteArtifact
// creates the output directory if it does not exist.
func TestWriteArtifactCreatesDir(t *testing.T) {
	t.Parallel()
	artifact := testArtifactComplete()

	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "sub", "dir")
	path, err := WriteArtifact(
		artifact, outputDir, consts.DefaultArtifactFormat,
	)
	if err != nil {
		t.Fatalf("WriteArtifact error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
}
