// SPDX-License-Identifier: Apache-2.0

package fallback_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hbraswelrh/gemara-user-journey/internal/fallback"
)

func TestLoadCachedDocs_Success(t *testing.T) {
	dir := t.TempDir()
	version := "v0.20.0"
	docPath := filepath.Join(
		dir,
		"schema-docs-v0.20.0.md",
	)
	content := []byte("# Schema Docs\nVersion v0.20.0\n")
	if err := os.WriteFile(
		docPath, content, 0o644,
	); err != nil {
		t.Fatalf("setup: %v", err)
	}

	docs, err := fallback.LoadCachedDocs(dir, version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docs.Version != version {
		t.Fatalf(
			"expected version %s, got %s",
			version,
			docs.Version,
		)
	}
	if string(docs.Content) != string(content) {
		t.Fatalf(
			"expected content %q, got %q",
			string(content),
			string(docs.Content),
		)
	}
}

func TestLoadCachedDocs_MissingCache(t *testing.T) {
	dir := t.TempDir()

	_, err := fallback.LoadCachedDocs(dir, "v0.20.0")
	if err == nil {
		t.Fatal("expected error for missing cache")
	}
}
