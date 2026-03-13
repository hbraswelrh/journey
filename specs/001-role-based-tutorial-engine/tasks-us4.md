# Tasks: US4 — Reusable Content Transformation (P4)

**Input**: `plan-us4.md`, `spec.md` (User Story 4)
**Prerequisites**: US1-US3 completed

---

## Phase 1: Content Block Data Model (FR-005)

### Tests

- [ ] T301 [P] [US4] Write test
  `internal/blocks/model_test.go`: ContentBlock struct
  contains required fields (ID, source tutorial path,
  source section, schema version, layer, category,
  content body, content hash, extracted timestamp)
- [ ] T302 [P] [US4] Write test
  `internal/blocks/model_test.go`: BlockCategory
  validates the five defined categories (pattern,
  validation_step, naming_convention, schema_structure,
  cross_reference)
- [ ] T303 [P] [US4] Write test
  `internal/blocks/model_test.go`: ContentHash computes
  a deterministic SHA-256 from the content body
- [ ] T304 [P] [US4] Write test
  `internal/blocks/model_test.go`: Manifest maps tutorial
  paths to block entries with hashes

### Implementation

- [ ] T305 [US4] Add content block constants to
  `internal/consts/consts.go`: category names, block
  cache directory (`~/.config/pacman/blocks/`), manifest
  filename, category heading keywords
- [ ] T306 [US4] Implement `internal/blocks/model.go`:
  `ContentBlock` struct, `BlockCategory` type with five
  values, `Manifest` struct, `ComputeHash(body string)
  string`, `NewBlock(...)` constructor

**Checkpoint**: Data model compiles. Block types and categories
defined. Hash computation deterministic.

---

## Phase 2: Content Extraction (FR-005)

### Tests

- [ ] T307 [P] [US4] Write test
  `internal/tutorials/loader_test.go`: ParseSections
  returns section headings with their full body content
  from a tutorial file
- [ ] T308 [P] [US4] Write test
  `internal/tutorials/loader_test.go`: ParseSections
  handles tutorial with no body sections (front matter
  only) gracefully
- [ ] T309 [P] [US4] Write test
  `internal/blocks/extractor_test.go`: ExtractBlocks
  from Threat Assessment Guide yields blocks for scope
  definition, capability identification, threat
  identification, and CUE validation (SC-003, US4-SC1)
- [ ] T310 [P] [US4] Write test
  `internal/blocks/extractor_test.go`: Each extracted
  block has source tutorial identity, schema version,
  and layer metadata
- [ ] T311 [P] [US4] Write test
  `internal/blocks/extractor_test.go`: Empty section
  body produces no block
- [ ] T312 [P] [US4] Write test
  `internal/blocks/extractor_test.go`: ExtractAll
  processes multiple tutorials and returns manifest
- [ ] T313 [P] [US4] Write test
  `internal/blocks/extractor_test.go`: Block categories
  are assigned correctly: "CUE Validation" section ->
  validation_step, "Scope Definition" -> pattern

### Implementation

- [ ] T314 [US4] Add `SectionContent` type and
  `ParseSections(path string) ([]SectionContent, error)`
  to `internal/tutorials/loader.go`: reads tutorial body
  after front matter, splits by `## ` headings, returns
  heading + body pairs
- [ ] T315 [US4] Implement `internal/blocks/extractor.go`:
  `CategorizeSection(heading string) BlockCategory` —
  match heading keywords to categories.
  `ExtractBlocks(tutorial Tutorial,
  sections []SectionContent, schemaVersion string)
  []ContentBlock` — create one block per non-empty
  section with computed hash.
  `ExtractAll(tutorials []Tutorial, dir string,
  schemaVersion string) ([]ContentBlock, *Manifest,
  error)` — batch extraction with manifest generation

**Checkpoint**: Tutorials are split into content blocks by
section. Blocks have correct categories and metadata. Manifest
tracks all extracted blocks.

---

## Phase 3: Drift Detection (FR-006)

### Tests

- [ ] T316 [P] [US4] Write test
  `internal/blocks/drift_test.go`: Modified tutorial
  section produces DriftResult with change type
  "modified" and different hashes
- [ ] T317 [P] [US4] Write test
  `internal/blocks/drift_test.go`: Removed tutorial
  section produces DriftResult with change type "removed"
- [ ] T318 [P] [US4] Write test
  `internal/blocks/drift_test.go`: New tutorial section
  produces DriftResult with change type "added"
- [ ] T319 [P] [US4] Write test
  `internal/blocks/drift_test.go`: Unchanged tutorial
  produces zero DriftResults
- [ ] T320 [P] [US4] Write test
  `internal/blocks/drift_test.go`: Missing manifest
  returns empty drift (signals full re-extraction needed)

### Implementation

- [ ] T321 [US4] Implement `internal/blocks/drift.go`:
  `ChangeType` enum (modified, removed, added).
  `DriftResult` struct (block ID, tutorial path, section,
  change type, old hash, new hash).
  `DetectDrift(current []ContentBlock,
  manifest *Manifest) []DriftResult` — compare current
  block hashes against manifest entries, identify all
  three change types

**Checkpoint**: Drift detection identifies modified, removed,
and added sections. Unchanged tutorials produce no results.

---

## Phase 4: Block Storage and Query (FR-005, US4-SC3)

### Tests

- [ ] T322 [P] [US4] Write test
  `internal/blocks/store_test.go`: SaveBlocks writes
  YAML files to the blocks directory
- [ ] T323 [P] [US4] Write test
  `internal/blocks/store_test.go`: LoadBlocks reads
  saved blocks and returns correct ContentBlock structs
- [ ] T324 [P] [US4] Write test
  `internal/blocks/store_test.go`: SaveManifest writes
  manifest YAML, LoadManifest reads it back correctly
- [ ] T325 [P] [US4] Write test
  `internal/blocks/query_test.go`: QueryByLayer returns
  only blocks matching the given layer
- [ ] T326 [P] [US4] Write test
  `internal/blocks/query_test.go`: QueryByCategory
  returns only blocks matching the given category
- [ ] T327 [P] [US4] Write test
  `internal/blocks/query_test.go`: QueryByGoal matches
  goal keywords against block headings and content
- [ ] T328 [P] [US4] Write test
  `internal/blocks/query_test.go`: AdaptationInstructions
  returns non-empty string referencing user goal and
  source tutorial

### Implementation

- [ ] T329 [US4] Implement `internal/blocks/store.go`:
  `SaveBlocks`, `LoadBlocks`, `SaveManifest`,
  `LoadManifest` with YAML serialization
- [ ] T330 [US4] Implement `internal/blocks/query.go`:
  `QueryByLayer`, `QueryByCategory`, `QueryByGoal`,
  `AdaptationInstructions`

**Checkpoint**: Blocks persist across sessions. Queries
correctly filter by layer, category, and goal. Adaptation
instructions reference context.

---

## Phase 5: CLI Integration and Polish

- [ ] T331 [US4] Implement
  `internal/cli/blocks_prompt.go`:
  `RunExtraction(cfg, out)` — extract blocks from all
  tutorials, display progress with RenderStatus, display
  summary with block counts per layer and category using
  styled output
- [ ] T332 [US4] Implement
  `internal/cli/blocks_prompt.go`:
  `RunDriftCheck(cfg, out)` — compare current tutorials
  against saved manifest, display affected blocks with
  change type and guidance using RenderWarning
- [ ] T333 [US4] Implement
  `internal/cli/blocks_prompt.go`:
  `RunBlockQuery(cfg, out)` — prompt user for goal or
  layer, display matching blocks with left-bar styled
  cards, show adaptation instructions
- [ ] T334 [US4] Wire block operations into
  `cmd/pacman/main.go` as `--extract`, `--drift-check`,
  and `--query` flags (or post-setup menu options in
  demo mode)
- [ ] T335 [US4] Integration test: extract blocks from
  test tutorials, verify manifest written, modify a
  tutorial fixture, re-extract, verify drift detected
- [ ] T336 [US4] Integration test: extract blocks, query
  by Layer 2, verify only Layer 2 blocks returned with
  adaptation instructions
- [ ] T337 [US4] Verify `make build`, `make test` pass
  with zero errors

**Checkpoint**: US4 is fully functional. Users can extract
content blocks, detect upstream drift, and query blocks by
goal or layer with adaptation instructions.

---

## Dependencies & Execution Order

- **Phase 1 (Data Model)**: No US4-internal dependencies
- **Phase 2 (Extraction)**: Depends on Phase 1
- **Phase 3 (Drift)**: Depends on Phase 2. Can run in
  **parallel** with Phase 4
- **Phase 4 (Store/Query)**: Depends on Phase 2. Can run in
  **parallel** with Phase 3
- **Phase 5 (CLI)**: Depends on all preceding phases

## Notes

- US4 tasks numbered T301-T337 to avoid conflicts with
  US1 (T001-T054), US2 (T101-T127), US3 (T201-T260).
- Key invariant: extraction at a given schema version MUST
  tag all blocks with that version (FR-019, SC-010).
- 80% section coverage target per SC-003.
