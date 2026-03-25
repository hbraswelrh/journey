// SPDX-License-Identifier: Apache-2.0

package fallback_test

import (
	"errors"
	"testing"

	"github.com/hbraswelrh/journey/internal/fallback"
)

func TestLoadBundledLexicon_ValidYAML(t *testing.T) {
	loader := func() ([]byte, error) {
		return []byte(`entries:
  - term: "Control"
    definition: "A safeguard or countermeasure."
  - term: "Threat"
    definition: "A potential cause of an incident."
`), nil
	}

	lexicon, err := fallback.LoadBundledLexicon(loader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lexicon.Entries) != 2 {
		t.Fatalf(
			"expected 2 entries, got %d",
			len(lexicon.Entries),
		)
	}
	if lexicon.Entries[0].Term != "Control" {
		t.Fatalf(
			"expected term Control, got %s",
			lexicon.Entries[0].Term,
		)
	}
}

func TestLoadBundledLexicon_InvalidYAML(t *testing.T) {
	loader := func() ([]byte, error) {
		return []byte(`not: [valid: yaml: {{`), nil
	}

	_, err := fallback.LoadBundledLexicon(loader)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadBundledLexicon_EmptyEntries(t *testing.T) {
	loader := func() ([]byte, error) {
		return []byte(`entries: []`), nil
	}

	_, err := fallback.LoadBundledLexicon(loader)
	if err == nil {
		t.Fatal("expected error for empty entries")
	}
}

func TestLoadBundledLexicon_LoaderError(t *testing.T) {
	loader := func() ([]byte, error) {
		return nil, errors.New("file not found")
	}

	_, err := fallback.LoadBundledLexicon(loader)
	if err == nil {
		t.Fatal("expected error from loader")
	}
}
