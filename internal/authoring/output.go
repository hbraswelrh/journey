// SPDX-License-Identifier: Apache-2.0

package authoring

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hbraswelrh/pacman/internal/consts"
	"gopkg.in/yaml.v3"
)

// RenderYAML serializes an authored artifact to YAML
// following Gemara naming conventions.
func RenderYAML(
	artifact *AuthoredArtifact,
) ([]byte, error) {
	data := buildOutputMap(artifact)
	return yaml.Marshal(data)
}

// RenderJSON serializes an authored artifact to JSON
// following Gemara naming conventions.
func RenderJSON(
	artifact *AuthoredArtifact,
) ([]byte, error) {
	data := buildOutputMap(artifact)
	return json.MarshalIndent(data, "", "  ")
}

// buildOutputMap converts an authored artifact into a map
// structure suitable for YAML/JSON serialization.
func buildOutputMap(
	artifact *AuthoredArtifact,
) map[string]interface{} {
	data := make(map[string]interface{})
	for _, section := range artifact.Sections {
		sectionData := make(map[string]interface{})
		for k, v := range section.Fields {
			sectionData[k] = v
		}
		data[section.Name] = sectionData
	}
	return data
}

// GenerateFilename generates a filename for an authored
// artifact following Gemara naming conventions. The filename
// is based on the artifact type and the name field from the
// metadata section.
func GenerateFilename(
	artifact *AuthoredArtifact,
) string {
	// Try to use the name from metadata.
	name := extractMetadataName(artifact)
	if name != "" {
		return sanitizeFilename(name) + ".yaml"
	}

	// Fall back to artifact type.
	typeName := strings.ToLower(artifact.ArtifactType)
	return typeName + "-artifact.yaml"
}

// extractMetadataName looks for a "name" field in the
// metadata section of the artifact.
func extractMetadataName(
	artifact *AuthoredArtifact,
) string {
	for _, section := range artifact.Sections {
		if section.Name == consts.SectionMetadata {
			if name, ok := section.Fields["name"]; ok {
				return name
			}
		}
	}
	return ""
}

// sanitizeFilename removes or replaces characters that are
// not safe for filenames.
func sanitizeFilename(name string) string {
	// Replace spaces and special chars with hyphens.
	replacer := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(name)
}

// WriteArtifact writes the authored artifact to disk in
// the specified format. Creates the output directory if
// it does not exist. Returns the output file path.
func WriteArtifact(
	artifact *AuthoredArtifact,
	outputDir string,
	format string,
) (string, error) {
	// Create output directory.
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf(
			"create output dir %s: %w",
			outputDir, err,
		)
	}

	// Generate filename with correct extension.
	filename := GenerateFilename(artifact)
	if format == "json" {
		filename = strings.TrimSuffix(
			filename, ".yaml",
		) + ".json"
	}

	outputPath := filepath.Join(outputDir, filename)

	// Serialize content.
	var data []byte
	var err error
	switch format {
	case "json":
		data, err = RenderJSON(artifact)
	default:
		data, err = RenderYAML(artifact)
	}
	if err != nil {
		return "", fmt.Errorf(
			"serialize artifact: %w", err,
		)
	}

	// Write file.
	if err := os.WriteFile(
		outputPath, data, 0o644,
	); err != nil {
		return "", fmt.Errorf(
			"write artifact file: %w", err,
		)
	}

	return outputPath, nil
}
