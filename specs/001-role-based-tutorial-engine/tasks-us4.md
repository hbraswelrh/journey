# Tasks: US4 — Reusable Content Transformation

**Input**: `specs/001-role-based-tutorial-engine/plan-us4.md`
**Prerequisites**: plan-us4.md (required), spec.md US4
section (required)

## Format: `[ID] [P?] [US4] Description`

---

## Phase 1: Content Block Data Model (FR-005)

**Purpose**: Define the ContentBlock type, block ID
generation, content hashing, and section-to-category mapping

- [x] T301 [US4] Create `internal/blocks/blocks.go` with
  SPDX header, package declaration, and `ContentBlock` struct
  (ID, SourceTutorial, SourceSection, SchemaVersion, Layer,
  Category, Content, ContentHash, ExtractedAt, LastVerified)
- [x] T302 [US4] Add `BlockIndex` type with `Blocks` slice
  and lookup methods: `ByID(id string) *ContentBlock`,
  `ByLayer(layer int) []ContentBlock`,
  `ByCategory(cat string) []ContentBlock`
- [x] T303 [US4] Add `GenerateBlockID(tutorialPath,
  sectionHeading string) string` — deterministic ID from
  source file base name + section heading slug
- [x] T304 [US4] Add `HashContent(content string) string` —
  SHA-256 hex digest of the content body
- [x] T305 [US4] Add `CategorizeSection(heading string)
  string` — map section headings to category constants using
  keyword matching (pattern, validation_step,
  naming_convention, schema_structure, cross_reference)
- [x] T306 [US4] Write `internal/blocks/blocks_test.go` —
  test `GenerateBlockID` determinism, `HashContent`
  correctness, `CategorizeSection` for all heading patterns,
  `BlockIndex` lookups

**Checkpoint**: ContentBlock type compiles, all unit tests
pass

---

## Phase 2: Block Extraction Engine (FR-005)

**Purpose**: Extract ContentBlocks from tutorial Markdown
files using the existing ParseSections function

- [x] T307 [US4] Add `ExtractBlocks(tutorials
  []tutorials.Tutorial, schemaVersion string)
  (*BlockIndex, error)` — iterate tutorials, call
  `tutorials.ParseSections` per file, create ContentBlock
  per section with category, layer, hash, timestamp
- [x] T308 [US4] Handle edge cases in `ExtractBlocks`:
  tutorials with no sections (skip), empty section bodies
  (skip), duplicate section headings across tutorials
  (distinct IDs via tutorial name prefix)
- [x] T309 [US4] Write extraction tests in
  `internal/blocks/blocks_test.go` — extract from
  `testdata/valid/threat-assessment-guide.md`, verify 4
  blocks: scope definition (pattern), capability
  identification (pattern), threat identification (pattern),
  CUE validation (validation_step). Each block must carry
  source tutorial identity, schema version, and layer 2
- [x] T310 [US4] Write extraction test: extract from all 4
  testdata tutorials, verify total block count, layer
  distribution, and category distribution
- [x] T311 [US4] Write extraction test: empty tutorial
  directory produces zero blocks (no error)

**Checkpoint**: Block extraction from testdata tutorials works
correctly, all tests pass

---

## Phase 3: Extraction Manifest and Drift Detection (FR-006)

**Purpose**: Persist extraction state and detect when upstream
tutorials change relative to previously extracted blocks

- [x] T312 [US4] Create `internal/blocks/manifest.go` with
  SPDX header and `Manifest` struct: map of block ID to
  `ManifestEntry` (ContentHash, SourceTutorial,
  SchemaVersion, ExtractedAt)
- [x] T313 [US4] Add `SaveManifest(path string,
  m *Manifest) error` — write manifest as YAML
- [x] T314 [US4] Add `LoadManifest(path string)
  (*Manifest, error)` — read manifest from YAML; return
  empty manifest for nonexistent file
- [x] T315 [US4] Add `DriftResult` struct: BlockID, DriftType
  (added, modified, removed), OldHash, NewHash
- [x] T316 [US4] Add `DetectDrift(current *BlockIndex,
  previous *Manifest) []DriftResult` — compare content
  hashes; return added (in current but not manifest),
  modified (hash differs), removed (in manifest but not
  current)
- [x] T317 [US4] Write `internal/blocks/manifest_test.go` —
  test save/load round-trip, detect drift for added block,
  modified block, removed block, and no-drift case
- [x] T318 [US4] Write drift detection integration test:
  extract blocks, save manifest, modify a tutorial section's
  content (in temp dir), re-extract, detect drift, verify
  the modified block is flagged

**Checkpoint**: Manifest persistence and drift detection work
correctly, all tests pass

---

## Phase 4: Context-Adaptive Block Retrieval

**Purpose**: Filter and return blocks relevant to the user's
activity profile and stated goal, with adaptation instructions

- [x] T319 [US4] Create `internal/blocks/retrieval.go` with
  SPDX header and `RetrievalResult` struct: Block
  (*ContentBlock), AdaptationInstructions (string),
  RelevanceScore (int)
- [x] T320 [US4] Add `RetrieveBlocks(index *BlockIndex,
  layers []int, goalKeywords []string)
  []RetrievalResult` — filter blocks by layers, score by
  category relevance to goal keywords, sort by score
  descending
- [x] T321 [US4] Add `GenerateAdaptation(
  block *ContentBlock, goal string) string` — produce
  adaptation instructions based on block category and goal
  text (e.g., "Adapt this scope definition pattern to
  define the scope of your guidance document")
- [x] T322 [US4] Write `internal/blocks/retrieval_test.go` —
  test retrieval for Layer 1 goal "create my own guidance
  document" returns Layer 1 blocks with adaptation
  instructions; test retrieval with empty layers returns no
  blocks; test retrieval with no matching goal still returns
  layer-matched blocks
- [x] T323 [US4] Write retrieval test: verify blocks are
  sorted by relevance (strong layer match first, matching
  category second)

**Checkpoint**: Context-adaptive retrieval works correctly,
all tests pass

---

## Phase 5: CLI Integration and TUI Rendering

**Purpose**: Wire block operations into the CLI, add styled
output, update session state

- [x] T324 [US4] Add block rendering styles to
  `internal/cli/styles.go`: `RenderContentBlock` (left-bar
  accent card with category badge, layer badge, source info),
  `RenderDriftResult` (drift indicator with block ID and
  drift type), `RenderBlockSummary` (count by category and
  layer)
- [x] T325 [US4] Create `internal/cli/blocks_prompt.go` with
  SPDX header, `BlocksConfig` struct (TutorialsDir,
  SchemaVersion, CacheDir), and `BlocksResult` struct
  (BlockCount, DriftResults)
- [x] T326 [US4] Add `RunBlockExtraction(cfg *BlocksConfig,
  out io.Writer) (*BlocksResult, error)` — load tutorials,
  extract blocks, save manifest, render summary
- [x] T327 [US4] Add `RunDriftCheck(cfg *BlocksConfig,
  out io.Writer) ([]DriftResult, error)` — load manifest,
  re-extract, detect drift, render results
- [x] T328 [US4] Add `RunBlockRetrieval(cfg *BlocksConfig,
  layers []int, goal string, out io.Writer) error` — load
  manifest or extract fresh, retrieve blocks, render with
  adaptation instructions
- [x] T329 [US4] Update `internal/session/session.go`: add
  `ContentBlocksCount int` field, add
  `SetContentBlocks(count int)` method
- [x] T330 [US4] Write `internal/cli/blocks_prompt_test.go` —
  test `RunBlockExtraction` with testdata tutorials, verify
  output contains block count and category summary; test
  `RunDriftCheck` with known drift; test
  `RunBlockRetrieval` with a Layer 2 goal
- [x] T331 [US4] Verify `make build` and `make test` pass
  with all new code

**Checkpoint**: Full US4 functionality works end-to-end,
all tests pass, build succeeds

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** (Data Model): No dependencies — start
  immediately
- **Phase 2** (Extraction): Depends on Phase 1
- **Phase 3** (Manifest & Drift): Depends on Phase 2
- **Phase 4** (Retrieval): Depends on Phase 2 (uses
  BlockIndex)
- **Phase 5** (CLI Integration): Depends on Phases 3 and 4

### Within Each Phase

- Tests MUST be written and FAIL before implementation
- Types before functions
- Core logic before CLI wiring
- Commit after each phase

### Parallel Opportunities

- Phases 3 and 4 can run in parallel after Phase 2 (they
  share BlockIndex but have no other dependencies)
- All test tasks marked [P] within a phase can run in
  parallel

---

## Notes

- Task numbering starts at T301 per US4 convention
  (US1=T001-T054, US2=T101-T127, US3=T201-T260)
- Constants already exist in `internal/consts/consts.go`:
  `BlockCacheDir`, `BlockManifestFile`, `CategoryPattern`,
  `CategoryValidationStep`, `CategoryNamingConv`,
  `CategorySchemaStruct`, `CategoryCrossRef`
- The existing `tutorials.ParseSections` function handles
  Markdown section parsing — no duplication needed
- Commit after each phase; push after all phases complete
