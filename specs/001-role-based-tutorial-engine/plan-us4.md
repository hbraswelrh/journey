# Implementation Plan: US4 — Reusable Content Transformation
# (P4)

**Branch**: `feat/adds-specification` | **Date**: 2026-03-13
**Spec**: [spec.md](spec.md) — User Story 4
**Depends on**: US1 (MCP Server Setup), US2 (Schema Version
Selection), US3 (Role & Activity Discovery) — all completed

## Summary

US4 extracts modular, reusable content blocks from Gemara
tutorials. Each block represents an actionable pattern (naming
conventions, validation steps, schema structures, cross-
referencing techniques) tagged with source identity, schema
version, and Gemara layer. The system detects when upstream
tutorials change and flags affected blocks for review. Users
can query blocks by goal or layer and receive adaptation
instructions.

This plan covers FR-005, FR-006, FR-019 (schema version
consistency for extraction), and the US4 acceptance scenarios
1-3.

## Technical Context

**Language/Version**: Go 1.26.1
**Dependencies**: Existing `internal/tutorials/` (loader,
Tutorial struct), `internal/consts/` (centralized constants),
`internal/cli/` (TUI styles, prompter interface),
`internal/roles/` (ActivityProfile for goal-based queries)
**Storage**: Content blocks persisted as YAML in
`~/.config/pacman/blocks/`; manifest file tracks extraction
state for drift detection
**Testing**: `go test ./...` via `make test`; TDD per
constitution
**Constraints**: Section-boundary extraction (heading-based
parsing, no NLP); blocks tagged with content categories
(pattern, validation step, naming convention, schema structure);
drift detection via content hashing

## Constitution Check

| Principle | Status | Notes |
|:---|:---|:---|
| I. Schema Conformance | Pass | Blocks tagged with schema version; extraction respects selected version (FR-019) |
| II. Gemara Layer Fidelity | Pass | Each block tagged with its source layer |
| III. TDD | Pass | Tests before implementation per phase |
| IV. Tutorial-First Design | Pass | US4 makes tutorials modular and evergreen |
| V. Incremental Delivery | Pass | US4 independently usable after US1-US3 |
| VI. Decision Documentation | N/A | No new ADRs anticipated |
| VII. Centralized Constants | Pass | Block categories, cache paths in consts |
| VIII. Composability | Pass | Extraction and query are independent operations |
| IX. Convention Over Configuration | Pass | Default extraction from all tutorials in configured dir |

## Source Code

```text
internal/
├── blocks/
│   ├── model.go              # ContentBlock, Manifest,
│   │                         #   BlockCategory types
│   ├── model_test.go
│   ├── extractor.go          # Extract blocks from tutorials
│   │                         #   by section heading parsing
│   ├── extractor_test.go
│   ├── drift.go              # Detect upstream changes via
│   │                         #   content hashing, flag blocks
│   ├── drift_test.go
│   ├── store.go              # Persist/load blocks and
│   │                         #   manifest as YAML
│   ├── store_test.go
│   ├── query.go              # Query blocks by layer, goal,
│   │                         #   or category; generate
│   │                         #   adaptation instructions
│   └── query_test.go
├── cli/
│   ├── blocks_prompt.go      # CLI flow: extract, list,
│   │                         #   query, drift check
│   └── blocks_prompt_test.go
├── consts/
│   └── consts.go             # Update: block categories,
│                             #   block cache dir, manifest
│                             #   filename
└── tutorials/
    └── loader.go             # Update: add BodySections()
                              #   method for full content
                              #   extraction
```

## Implementation Phases

### Phase 1: Content Block Data Model (FR-005)

- Define `ContentBlock` struct: ID, source tutorial path,
  source tutorial title, source section heading, Gemara
  schema version, Gemara layer (int), content category
  (pattern | validation_step | naming_convention |
  schema_structure | cross_reference), content body (string),
  content hash (SHA-256 of body for drift detection),
  extracted timestamp.
- Define `BlockCategory` type with the five categories.
- Define `Manifest` struct: maps tutorial path -> list of
  block IDs with content hashes, extraction timestamp,
  schema version used. Used by drift detection.
- Add block-related constants to `internal/consts/consts.go`:
  category names, block cache directory path
  (`~/.config/pacman/blocks/`), manifest filename.
- Tests: ContentBlock construction, category validation,
  manifest round-trip.

### Phase 2: Content Extraction (FR-005)

- Extend `internal/tutorials/loader.go` with a
  `ParseSections(path string) ([]SectionContent, error)`
  function that reads a tutorial file and returns the full
  body of each section (heading + text until next heading).
- `internal/blocks/extractor.go`:
  `ExtractBlocks(tutorial Tutorial,
  sections []SectionContent, schemaVersion string)
  []ContentBlock` — split each section into content blocks
  by identifying actionable patterns. Categorize each block
  by matching section heading keywords to categories:
  - "scope", "definition" -> pattern
  - "validation", "CUE", "vet" -> validation_step
  - "naming", "convention", "identifier" -> naming_convention
  - "schema", "structure", "artifact" -> schema_structure
  - "cross-reference", "mapping", "link" -> cross_reference
  - Default: pattern
- `ExtractAll(tutorials []Tutorial, dir string,
  schemaVersion string) ([]ContentBlock, *Manifest, error)`
  — batch extraction across all tutorials, produces manifest.
- Tests per SC-003: extraction from Threat Assessment Guide
  yields blocks for scope definition, capability ID, threat
  ID, CUE validation. At least 80% section coverage. Empty
  section body produces no block.

### Phase 3: Drift Detection (FR-006)

- `internal/blocks/drift.go`:
  `DetectDrift(current []ContentBlock, manifest *Manifest)
  []DriftResult` — compare current content hashes against
  manifest hashes. Return list of affected blocks with
  change type (modified, removed, new section).
- `DriftResult` struct: block ID, tutorial path, section,
  change type, old hash, new hash.
- Drift check runs automatically when tutorials are re-loaded
  and a previous manifest exists.
- Tests per SC-004: modified tutorial section detected,
  removed section detected, new section detected, unchanged
  tutorial produces no drift, manifest missing triggers full
  re-extraction.

### Phase 4: Block Storage and Query (FR-005, US4-SC3)

- `internal/blocks/store.go`: Persist blocks as YAML files
  in the blocks directory. Persist manifest as a separate
  YAML file. Load/list operations.
  - `SaveBlocks(dir string, blocks []ContentBlock) error`
  - `SaveManifest(dir string, manifest *Manifest) error`
  - `LoadBlocks(dir string) ([]ContentBlock, error)`
  - `LoadManifest(dir string) (*Manifest, error)`
- `internal/blocks/query.go`: Query blocks by criteria.
  - `QueryByLayer(blocks []ContentBlock, layer int)
    []ContentBlock`
  - `QueryByCategory(blocks []ContentBlock,
    category BlockCategory) []ContentBlock`
  - `QueryByGoal(blocks []ContentBlock, goal string)
    []ContentBlock` — match goal keywords against block
    section headings and content body.
  - `AdaptationInstructions(block ContentBlock,
    userGoal string) string` — generate context-specific
    instructions for adapting the block.
- Tests: query by layer returns correct blocks, query by
  goal matches relevant blocks, adaptation instructions
  reference user goal and source tutorial.

### Phase 5: CLI Integration and Polish

- `internal/cli/blocks_prompt.go`: CLI flows for content
  block operations.
  - `RunExtraction(cfg, out)` — extract blocks from all
    tutorials, display progress and summary.
  - `RunDriftCheck(cfg, out)` — compare current tutorials
    against manifest, display affected blocks.
  - `RunBlockQuery(cfg, out)` — prompt user for goal or
    layer, display matching blocks with adaptation
    instructions.
- Styled TUI output using existing lipgloss styles: left-bar
  accent for blocks, layer badges, category tags.
- Wire into main.go as subcommands or post-setup options.
- Integration tests: full extract -> drift check -> query
  flow.
- Verify `make build`, `make test` pass.

## Dependencies & Execution Order

```text
Phase 1 (Data Model)
    │
    ▼
Phase 2 (Extraction)
    │
    ├──────────────────┐
    ▼                  ▼
Phase 3 (Drift)    Phase 4 (Store/Query)
    │                  │
    └──────┬───────────┘
           ▼
    Phase 5 (CLI Integration)
```

- Phases 3 and 4 can proceed in parallel after Phase 2.
- Phase 5 integrates all preceding work.

## Complexity Tracking

No constitution violations identified. No complexity
justifications required.
