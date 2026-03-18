// SPDX-License-Identifier: Apache-2.0

package fallback

import (
	"fmt"
	"os"
	"path/filepath"
)

// SchemaDocs holds cached schema documentation content.
type SchemaDocs struct {
	// Version is the Gemara schema version this documentation
	// was cached from.
	Version string
	// Content is the raw documentation content.
	Content []byte
}

// LoadCachedDocs is the fallback for the
// gemara://schema/definitions MCP resource when the server
// is unavailable. It loads cached schema documentation for
// the specified version from the given cache directory.
func LoadCachedDocs(
	cacheDir string,
	version string,
) (*SchemaDocs, error) {
	docPath := filepath.Join(
		cacheDir,
		fmt.Sprintf("schema-docs-%s.md", version),
	)

	content, err := os.ReadFile(docPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf(
			"no cached schema docs for version %s; "+
				"install the Gemara MCP server or run "+
				"with network access to fetch documentation",
			version,
		)
	}
	if err != nil {
		return nil, fmt.Errorf(
			"read cached docs: %w", err,
		)
	}

	return &SchemaDocs{
		Version: version,
		Content: content,
	}, nil
}
