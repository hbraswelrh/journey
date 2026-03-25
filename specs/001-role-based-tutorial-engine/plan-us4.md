# Implementation Plan: US4 — Reusable Content
# Transformation (P4)

**Branch**: `001-role-based-tutorial-engine` | **Date**: 2026-03-13
**Spec**: [spec.md](spec.md) — User Story 4
**Depends on**: US1 (MCP Server Setup), US2 (Schema Version
Selection), US3 (Role and Activity Discovery) — all completed

## Summary

US4 implements reusable content block extraction from Gemara
tutorials and upstream drift detection. The system reads
tutorial Markdown files, parses them into sections, and
extracts modular content blocks — each tagged with source
tutorial identity, Gemara schema version, Gemara layer, and
content category (pattern, validation step, naming convention,
schema structure, cross-reference). Blocks are persisted as
YAML in a local cache. When tutorials are updated upstream,
the system detects which blocks are affected by comparing
content hashes against a stored extraction manifest.

Users can request blocks relevant to their goal (via the
activity profile from US3) and receive context-adaptive
retrieval with adaptation instructions.

This plan covers FR-005, FR-006, FR-019 (content block
version alignment), and the US4 acceptance scenarios 1-3.

## Technical Context

**Language/Version**: Go 1.26.1
**Dependencies**: Existing `internal/tutorials/` (loader,
section parser), `internal/roles/` (activity profile, layer
mappings), `internal/consts/` (centralized constants),
`internal/session/` (schema version), `internal/cli/`
(TUI styles, prompter interface)
**Storage**: Local filesystem — content blocks stored as YAML
in `~/.config/gemara-user-journey/blocks/`; extraction manifest at
`~/.config/gemara-user-journey/blocks/manifest.yaml`
**Testing**: `go test ./...` via `make test`; TDD per
constitution
**Constraints**: Content extraction uses section-boundary
parsing from the existing `ParseSections` function in
`loader.go`. Block categories are assigned by keyword matching
against section headings. Drift detection uses SHA-256 content
hashes.

## Constitution Check

| Principle | Status | Notes |
|:---|:---|:---|
| I. Schema Conformance | Pass | Blocks reference schema version; no schema output produced directly |
| II. Gemara Layer Fidelity | Pass | Each block is tagged with its source Gemara layer |
| III. TDD | Pass | Tests written before implementation per phase |
| IV. Tutorial-First Design | Pass | US4 makes tutorial content modular and reusable |
| V. Incremental Delivery | Pass | US4 is independently usable after US1+US2+US3 |
| VI. Decision Documentation | N/A | No new ADRs anticipated; patterns consistent with prior stories |
| VII. Centralized Constants | Pass | Block categories, cache paths in `internal/consts/consts.go` |
| VIII. Composability | Pass | Block extraction is an independent operation |
| IX. Convention Over Configuration | Pass | Default cache paths; zero config for common case |

## Source Code

```text
internal/
├── blocks/
│   ├── blocks.go             # ContentBlock type, block
│   │                         #   extraction from tutorials
│   ├── blocks_test.go
│   ├── manifest.go           # Extraction manifest: store,
│   │                         #   load, drift detection
│   ├── manifest_test.go
│   ├── retrieval.go          # Context-adaptive block
│   │                         #   retrieval by profile/goal
│   └── retrieval_test.go
├── tutorials/
│   └── loader.go             # Already has ParseSections
│                             #   (no changes needed)
├── cli/
│   ├── blocks_prompt.go      # CLI flow: extract, detect
│   │                         #   drift, retrieve blocks
│   ├── blocks_prompt_test.go
│   └── styles.go             # Update: add block rendering
│                             #   styles
├── consts/
│   └── consts.go             # Already has block constants
│                             #   (BlockCacheDir,
│                             #   BlockManifestFile,
│                             #   Category* constants)
└── session/
    └── session.go            # Update: add content blocks
                              #   count to session state
```

## Implementation Phases

### Phase 1: Content Block Data Model (FR-005)

- Define the `ContentBlock` type in
  `internal/blocks/blocks.go`:
  - `ID string` — deterministic ID from source + section.
  - `SourceTutorial string` — tutorial file path.
  - `SourceSection string` — section heading.
  - `SchemaVersion string` — Gemara schema version at
    extraction time.
  - `Layer int` — Gemara layer number.
  - `Category string` — one of the category constants
    (pattern, validation_step, naming_convention,
    schema_structure, cross_reference).
  - `Content string` — section body text.
  - `ContentHash string` — SHA-256 hash for drift detection.
  - `ExtractedAt time.Time` — timestamp of extraction.
  - `LastVerified time.Time` — last drift-check timestamp.
- Define the `BlockIndex` type: a collection of blocks with
  lookup by ID, by layer, and by category.
- Add the `CategorizeSection` function: maps section headings
  to content categories using keyword matching against heading
  text. Categories per the spec:
  - Pattern: "scope definition", "capability identification",
    "threat identification", "policy structure",
    "implementation plan", "creating a guidance catalog".
  - Validation step: "cue validation", "validation".
  - Naming convention: "naming", "metadata setup".
  - Schema structure: "structure", "catalog structure",
    "control catalog structure".
  - Cross-reference: "cross-references", "mapping documents",
    "importing external catalogs", "osps baseline",
    "finos ccc", "integration".
- Tests: Block construction, ID generation determinism,
  category assignment from headings, content hashing.

### Phase 2: Block Extraction Engine (FR-005)

- `internal/blocks/blocks.go`:
  - `ExtractBlocks(tutorials []Tutorial,
    schemaVersion string) (*BlockIndex, error)` — iterate
    tutorials, call `ParseSections` for each, create
    `ContentBlock` per section with appropriate category,
    layer from tutorial metadata, and content hash.
  - Handle edge cases: tutorials with no sections, empty
    section bodies, duplicate section headings across
    tutorials.
- Tests per acceptance scenario 1 (Threat Assessment Guide):
  extract blocks from testdata, verify blocks for scope
  definition pattern, capability identification pattern,
  threat identification pattern, CUE validation pattern.
  Each block must have source tutorial identity and schema
  version metadata.

### Phase 3: Extraction Manifest and Drift Detection (FR-006)

- `internal/blocks/manifest.go`:
  - `Manifest` type: maps block IDs to content hashes plus
    extraction metadata (tutorial path, schema version,
    extraction timestamp).
  - `SaveManifest(path string, m *Manifest) error` — write
    YAML.
  - `LoadManifest(path string) (*Manifest, error)` — read
    YAML.
  - `DetectDrift(current *BlockIndex,
    previous *Manifest) []DriftResult` — compare content
    hashes between the current extraction and the stored
    manifest. Return a list of blocks whose content has
    changed (added, modified, removed).
  - `DriftResult` type: block ID, drift type (added,
    modified, removed), old hash, new hash.
- Tests per acceptance scenario 2: extract blocks at v0.17.0,
  modify a tutorial section, re-extract, verify drift
  detection flags the changed block.

### Phase 4: Context-Adaptive Block Retrieval (FR-005,
acceptance scenario 3)

- `internal/blocks/retrieval.go`:
  - `RetrieveBlocks(index *BlockIndex,
    profile *ActivityProfile, goal string)
    []RetrievalResult` — filter blocks by the user's
    resolved layers and goal keywords. Return blocks sorted
    by relevance (strong layer match first, then category
    relevance).
  - `RetrievalResult` type: the block plus adaptation
    instructions explaining how to adapt the block to the
    user's context.
  - `GenerateAdaptation(block *ContentBlock,
    goal string) string` — produce adaptation instructions
    based on the block's category and the user's stated goal.
- Tests per acceptance scenario 3: user goal "create my own
  guidance document" retrieves Layer 1 blocks with adaptation
  instructions.

### Phase 5: CLI Integration and TUI Rendering

- `internal/cli/blocks_prompt.go`:
  - `RunBlockExtraction(cfg *BlocksConfig,
    out io.Writer) (*BlocksResult, error)` — orchestrates
    the extract → manifest → display flow.
  - `RunDriftCheck(cfg *BlocksConfig,
    out io.Writer) ([]DriftResult, error)` — loads the
    previous manifest, re-extracts, compares, displays
    results.
  - `RunBlockRetrieval(cfg *BlocksConfig,
    profile *ActivityProfile, goal string,
    out io.Writer) error` — retrieves and displays
    relevant blocks.
- `internal/cli/styles.go`: Add rendering functions for
  content blocks (block card with left-bar accent, category
  badge, layer badge, drift indicators).
- Update `internal/cli/setup.go`: wire block extraction into
  the setup flow after learning path generation (optional —
  user may invoke separately).
- Update `internal/session/session.go`: add
  `ContentBlocksCount int` field, `SetContentBlocks` method.
- Integration test: full flow from tutorial loading → block
  extraction → manifest save → drift detection → block
  retrieval.
- Verify `make build`, `make test` pass.

## Dependencies & Execution Order

```text
Phase 1 (Data Model)
        │
        ▼
Phase 2 (Extraction Engine)
        │
        ▼
Phase 3 (Manifest & Drift Detection)
        │
        ▼
Phase 4 (Context-Adaptive Retrieval)
        │
        ▼
Phase 5 (CLI Integration & TUI)
```

All phases are sequential. Each phase depends on the
preceding one.

## Complexity Tracking

No constitution violations identified. No complexity
justifications required.
