// SPDX-License-Identifier: Apache-2.0

// Package tutorials handles loading tutorial metadata from the
// Gemara tutorials directory and generating tailored learning
// paths based on user role and activity profiles.
package tutorials

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Tutorial represents a single Gemara tutorial with its
// metadata.
type Tutorial struct {
	// Title is the tutorial's display name.
	Title string
	// FilePath is the absolute or relative path to the
	// tutorial file.
	FilePath string
	// Layer is the Gemara layer number (1-7) this tutorial
	// belongs to.
	Layer int
	// Sections are the major section headings in the
	// tutorial.
	Sections []string
	// SchemaVersion is the Gemara schema version this
	// tutorial references (e.g., "v0.18.0").
	SchemaVersion string
}

// VersionMismatch records a tutorial whose schema version
// differs from the user's selected version.
type VersionMismatch struct {
	// Tutorial is the mismatched tutorial.
	Tutorial Tutorial
	// TutorialVersion is the version the tutorial
	// references.
	TutorialVersion string
	// SelectedVersion is the user's selected schema
	// version.
	SelectedVersion string
}

// LoadTutorials scans a directory for tutorial files (Markdown
// with YAML front matter) and returns a structured index. It
// parses the front matter for title, layer, schema_version,
// and sections fields.
//
// Returns an empty slice (not an error) for an empty directory.
// Returns an error with resolution guidance for a nonexistent
// directory.
func LoadTutorials(dir string) ([]Tutorial, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"tutorials directory not found: %s — "+
					"clone the Gemara repository or "+
					"update the configured tutorials "+
					"path",
				dir,
			)
		}
		return nil, fmt.Errorf(
			"access tutorials directory %s: %w",
			dir, err,
		)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf(
			"%s is not a directory", dir,
		)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf(
			"read tutorials directory %s: %w",
			dir, err,
		)
	}

	var tutorials []Tutorial
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		tut, err := parseTutorialFile(path)
		if err != nil {
			// Skip files that cannot be parsed.
			continue
		}
		tutorials = append(tutorials, *tut)
	}

	return tutorials, nil
}

// CheckVersionCompat identifies tutorials whose schema version
// references differ from the selected version.
func CheckVersionCompat(
	tutorials []Tutorial,
	selectedVersion string,
) []VersionMismatch {
	var mismatches []VersionMismatch
	for _, tut := range tutorials {
		if tut.SchemaVersion != "" &&
			tut.SchemaVersion != selectedVersion {
			mismatches = append(
				mismatches,
				VersionMismatch{
					Tutorial:        tut,
					TutorialVersion: tut.SchemaVersion,
					SelectedVersion: selectedVersion,
				},
			)
		}
	}
	return mismatches
}

// parseTutorialFile reads a Markdown file with YAML front
// matter and extracts tutorial metadata.
func parseTutorialFile(path string) (*Tutorial, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Look for front matter delimiters (---).
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty file: %s", path)
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return nil, fmt.Errorf(
			"no front matter in %s", path,
		)
	}

	// Parse front matter fields.
	tut := &Tutorial{FilePath: path}
	var inSections bool
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			break // End of front matter.
		}

		if inSections {
			if strings.HasPrefix(trimmed, "- ") {
				section := strings.TrimPrefix(
					trimmed, "- ",
				)
				tut.Sections = append(
					tut.Sections, section,
				)
				continue
			}
			inSections = false
		}

		key, value, ok := parseFrontMatterLine(line)
		if !ok {
			continue
		}

		switch key {
		case "title":
			tut.Title = value
		case "layer":
			layer, err := strconv.Atoi(value)
			if err == nil {
				tut.Layer = layer
			}
		case "schema_version":
			tut.SchemaVersion = value
		case "sections":
			inSections = true
		}
	}

	if tut.Title == "" {
		return nil, fmt.Errorf(
			"no title in front matter: %s", path,
		)
	}

	return tut, nil
}

// parseFrontMatterLine extracts a key: value pair from a YAML
// front matter line.
func parseFrontMatterLine(
	line string,
) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}

	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])

	return key, value, true
}

// SectionContent holds a heading and its body text parsed
// from a tutorial's Markdown content.
type SectionContent struct {
	// Heading is the section heading text (without ##).
	Heading string
	// Body is the full text content of the section.
	Body string
}

// ParseSections reads a tutorial file and returns the body
// content split by `## ` headings. Content before the first
// heading is ignored (typically the H1 title). Returns nil
// for files with no sections.
func ParseSections(path string) ([]SectionContent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf(
			"open tutorial %s: %w", path, err,
		)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Skip past front matter.
	inFrontMatter := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			if inFrontMatter {
				break // End of front matter.
			}
			inFrontMatter = true
		}
	}

	var sections []SectionContent
	var current *SectionContent

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## ") {
			// Save previous section.
			if current != nil &&
				strings.TrimSpace(current.Body) != "" {
				sections = append(sections, *current)
			}
			heading := strings.TrimPrefix(line, "## ")
			heading = strings.TrimSpace(heading)
			current = &SectionContent{Heading: heading}
			continue
		}

		if current != nil {
			current.Body += line + "\n"
		}
	}

	// Save last section.
	if current != nil &&
		strings.TrimSpace(current.Body) != "" {
		sections = append(sections, *current)
	}

	return sections, nil
}
