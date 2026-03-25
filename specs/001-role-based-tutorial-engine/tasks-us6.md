# Tasks: US6 — Guided Gemara Content Authoring (P6)

**Input**: `plan-us6.md`, `spec.md` (User Story 6)
**Prerequisites**: US1-US4 completed (MCP setup, schema
version selection, role and activity discovery, reusable
content transformation)

---

## Phase 1: Authored Artifact Data Model (FR-009)

**Purpose**: Define authored artifact, authoring step,
artifact section, step field, and artifact template types

### Tests

- [x] T501 [P] [US6] Write test
  `internal/authoring/model_test.go`: NewAuthoredArtifact
  creates an artifact with the given type, schema
  definition, schema version, and empty sections
- [x] T502 [P] [US6] Write test
  `internal/authoring/model_test.go`:
  ArtifactTypeToSchema maps each supported artifact type
  to its CUE schema definition (e.g., "ThreatCatalog" ->
  "#ThreatCatalog")
- [x] T503 [P] [US6] Write test
  `internal/authoring/model_test.go`:
  SupportedArtifactTypes returns the six artifact types
  that have published CUE schemas
- [x] T504 [P] [US6] Write test
  `internal/authoring/model_test.go`: AddSection appends
  an ArtifactSection with the given name and empty field
  values
- [x] T505 [P] [US6] Write test
  `internal/authoring/model_test.go`:
  SetFieldValue records a value for a named field within
  a section; returns error for unknown section
- [x] T506 [P] [US6] Write test
  `internal/authoring/model_test.go`:
  ValidationStatus transitions: NotValidated -> Partial
  (after step validation) -> Valid (after full validation)
  or Invalid (on failure)
- [x] T507 [P] [US6] Write test
  `internal/authoring/model_test.go`:
  StepField with Required=true and empty value is
  reported by IncompleteFields
- [x] T508 [P] [US6] Write test
  `internal/authoring/model_test.go`:
  ArtifactTemplate defines ordered steps for a
  ThreatCatalog with correct section names (metadata,
  scope, capabilities, threats)

### Implementation

- [x] T509 [US6] Add authoring constants to
  `internal/consts/consts.go`: `AuthoringOutputDir`
  (`artifacts`), `DefaultArtifactFormat` (`yaml`),
  section name constants (`SectionMetadata`,
  `SectionScope`, `SectionCapabilities`,
  `SectionThreats`, `SectionControls`,
  `SectionGuidanceItems`, `SectionPolicyCriteria`,
  `SectionMappings`, `SectionEvaluations`),
  `ValidationStatusNotValidated`,
  `ValidationStatusPartial`, `ValidationStatusValid`,
  `ValidationStatusInvalid`
- [x] T510 [US6] Create `internal/authoring/model.go` with
  SPDX header:
  - `ValidationStatus` type (string enum: not_validated,
    partial, valid, invalid)
  - `StepField` struct (Name, Description, FieldType,
    Required, ExampleValue, HelpText)
  - `AuthoringStep` struct (Name, Description,
    RoleExplanation, Fields []StepField, Completed)
  - `ArtifactSection` struct (Name, Fields
    map[string]string, Step *AuthoringStep)
  - `AuthoredArtifact` struct (ArtifactType, SchemaDef,
    SchemaVersion, Sections []ArtifactSection,
    ValidationStatus, AuthoringRole, CreatedAt,
    UpdatedAt)
  - `ArtifactTemplate` struct (ArtifactType, Steps
    []AuthoringStep, Layer int, TutorialRefs []string)
  - Constructor functions and helper methods:
    `NewAuthoredArtifact`, `AddSection`,
    `SetFieldValue`, `IncompleteFields`,
    `ArtifactTypeToSchema`, `SupportedArtifactTypes`

**Checkpoint**: Data model compiles, all unit tests pass.

---

## Phase 2: Authoring Step Engine (FR-009, FR-010)

**Purpose**: Implement step sequencing, field guidance,
value suggestions, and artifact assembly

### Tests

- [x] T511 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  NewAuthoringEngine creates an engine with correct
  template, initial step index at 0, and role context
- [x] T512 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  CurrentStep returns the first step with
  role-personalized explanation
- [x] T513 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  SetFieldValue records value for a field in the current
  step; returns error for unknown field name
- [x] T514 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  CompleteStep advances to next step after all required
  fields are filled
- [x] T515 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  CompleteStep with missing required fields returns
  validation errors listing the missing fields
- [x] T516 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  GetSuggestions returns example values from the template
  and relevant content blocks for the current role
- [x] T517 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  Progress returns correct completed/total counts
  as steps are completed
- [x] T518 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  BuildArtifact assembles all completed sections into
  an AuthoredArtifact with correct metadata
- [x] T519 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  ArtifactTemplates returns templates for all six
  supported artifact types with non-empty step lists
- [x] T520 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  ThreatCatalog template has steps for metadata, scope,
  capabilities, and threats sections in order
- [x] T521 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  Engine with role "Security Engineer" personalizes step
  explanations to reference security concerns
- [x] T522 [P] [US6] Write test
  `internal/authoring/engine_test.go`:
  CompleteStep on the last step sets IsComplete flag

### Implementation

- [x] T523 [US6] Create `internal/authoring/engine.go` with
  SPDX header:
  - `AuthoringEngine` struct (template ArtifactTemplate,
    currentStep int, artifact *AuthoredArtifact,
    roleName string, keywords []string,
    blocks *BlockIndex, isComplete bool)
  - `ArtifactTemplates() map[string]ArtifactTemplate` —
    return templates for GuidanceCatalog,
    ControlCatalog, ThreatCatalog, Policy,
    MappingDocument, EvaluationLog
  - `NewAuthoringEngine(template ArtifactTemplate,
    roleName string, keywords []string,
    blocks *BlockIndex) *AuthoringEngine`
  - `CurrentStep() *AuthoringStep`
  - `SetFieldValue(field string, value string) error`
  - `CompleteStep() ([]ValidationError, error)`
  - `GetSuggestions(field string) []string`
  - `Progress() (completed int, total int)`
  - `BuildArtifact() *AuthoredArtifact`
  - `IsComplete() bool`
  - `personalizeExplanation(step *AuthoringStep,
    roleName string) string` — add role-specific
    context to step explanations

**Checkpoint**: Engine correctly sequences steps, records
field values, provides suggestions, and builds artifacts.

---

## Phase 3: Step-Level and Full Validation (FR-010)

**Purpose**: Validate in-progress and completed artifacts
against the Gemara CUE schema using MCP or local fallback

### Tests

- [x] T524 [P] [US6] Write test
  `internal/authoring/validate_test.go`:
  ValidationError contains field path, error message,
  and fix suggestion
- [x] T525 [P] [US6] Write test
  `internal/authoring/validate_test.go`:
  MCPValidator calls ValidateArtifact on the MCP client
  and translates response to ValidationError entries
- [x] T526 [P] [US6] Write test
  `internal/authoring/validate_test.go`:
  LocalValidator calls cue vet via CUERunner and parses
  error output into ValidationError entries
- [x] T527 [P] [US6] Write test
  `internal/authoring/validate_test.go`:
  ValidatePartial serializes completed fields and
  validates; returns errors for invalid fields only
- [x] T528 [P] [US6] Write test
  `internal/authoring/validate_test.go`:
  ValidateFull validates the complete artifact; returns
  empty errors for a valid artifact
- [x] T529 [P] [US6] Write test
  `internal/authoring/validate_test.go`:
  ValidateFull with missing required fields returns
  errors with actionable fix suggestions
- [x] T530 [P] [US6] Write test
  `internal/authoring/validate_test.go`:
  NewValidator returns MCPValidator when session has MCP
  connected, LocalValidator when in fallback mode
- [x] T531 [P] [US6] Write test
  `internal/authoring/validate_test.go`:
  Validation uses session's selected schema version
  (FR-019 compliance)

### Implementation

- [x] T532 [US6] Create `internal/authoring/validate.go`
  with SPDX header:
  - `ValidationError` struct (FieldPath, Message,
    FixSuggestion)
  - `Validator` interface with `ValidatePartial(
    artifact *AuthoredArtifact,
    stepIdx int) ([]ValidationError, error)` and
    `ValidateFull(
    artifact *AuthoredArtifact) ([]ValidationError,
    error)`
  - `MCPValidator` struct (client MCPClient) implementing
    Validator
  - `LocalValidator` struct (runner CUERunner)
    implementing Validator
  - `NewValidator(mcpAvailable bool,
    mcpClient MCPClient,
    runner CUERunner) Validator`
  - `parseValidationOutput(output string)
    []ValidationError` — parse cue vet or MCP error
    output into structured errors with fix suggestions

**Checkpoint**: Both MCP and local validators correctly
identify errors and provide actionable fix suggestions.

---

## Phase 4: YAML/JSON Output Generation (FR-012)

**Purpose**: Serialize authored artifacts to YAML and JSON,
write to disk with Gemara naming conventions

### Tests

- [x] T533 [P] [US6] Write test
  `internal/authoring/output_test.go`:
  RenderYAML produces valid YAML with correct structure
  for a ThreatCatalog artifact
- [x] T534 [P] [US6] Write test
  `internal/authoring/output_test.go`:
  RenderJSON produces valid JSON with correct structure
  for a ThreatCatalog artifact
- [x] T535 [P] [US6] Write test
  `internal/authoring/output_test.go`:
  GenerateFilename produces Gemara-convention filename
  (e.g., contains artifact type abbreviation)
- [x] T536 [P] [US6] Write test
  `internal/authoring/output_test.go`:
  WriteArtifact creates a file at the expected path
  with correct content
- [x] T537 [P] [US6] Write test
  `internal/authoring/output_test.go`:
  WriteArtifact with format "json" writes JSON output
- [x] T538 [P] [US6] Write test
  `internal/authoring/output_test.go`:
  RenderYAML output round-trips through YAML
  parse/marshal without data loss

### Implementation

- [x] T539 [US6] Create `internal/authoring/output.go`
  with SPDX header:
  - `RenderYAML(
    artifact *AuthoredArtifact) ([]byte, error)`
  - `RenderJSON(
    artifact *AuthoredArtifact) ([]byte, error)`
  - `GenerateFilename(
    artifact *AuthoredArtifact) string`
  - `WriteArtifact(artifact *AuthoredArtifact,
    outputDir string,
    format string) (string, error)`

**Checkpoint**: Artifacts serialize correctly to both YAML
and JSON. Files are written with correct naming conventions.

---

## Phase 5: CLI Integration and TUI Rendering

**Purpose**: Wire guided authoring into the CLI, add styled
output, update session state

### Tests

- [x] T540 [P] [US6] Write test
  `internal/cli/author_prompt_test.go`:
  RunGuidedAuthoring presents artifact type selection
  and returns selected type
- [x] T541 [P] [US6] Write test
  `internal/cli/author_prompt_test.go`:
  RunGuidedAuthoring walks through all steps of a
  ThreatCatalog template, prompting for each field
- [x] T542 [P] [US6] Write test
  `internal/cli/author_prompt_test.go`:
  Validation errors at each step are displayed with fix
  suggestions
- [x] T543 [P] [US6] Write test
  `internal/cli/author_prompt_test.go`:
  Completed authoring produces a valid artifact and
  writes output file
- [x] T544 [P] [US6] Write test
  `internal/cli/author_prompt_test.go`:
  Artifact type list is filtered by role when role
  context is available

### Implementation

- [x] T545 [US6] Add authoring rendering to
  `internal/cli/styles.go`:
  `RenderAuthoringStep` (step card with progress,
  field list, role-specific explanation),
  `RenderFieldPrompt` (field input with description
  and example), `RenderValidationResults` (error
  display with fix suggestions),
  `RenderArtifactSummary` (completed artifact overview),
  `RenderAuthoringProgress` (progress indicator)
- [x] T546 [US6] Create
  `internal/cli/author_prompt.go` with SPDX header:
  `AuthorPromptConfig` struct (Prompter, Session,
  TutorialsDir, SchemaVersion, OutputDir, OutputFormat,
  BlockIndex), `AuthorPromptResult` struct (Artifact,
  OutputPath, ValidationErrors).
  `RunGuidedAuthoring(cfg *AuthorPromptConfig,
  out io.Writer) (*AuthorPromptResult, error)` — full
  flow: artifact type selection, step-by-step prompting,
  validation at each step, final output.
  `selectArtifactType(cfg *AuthorPromptConfig,
  out io.Writer) (string, error)` — present available
  types filtered by role context.
  `runAuthoringStep(engine *AuthoringEngine,
  cfg *AuthorPromptConfig,
  out io.Writer) ([]ValidationError, error)` —
  prompt for field values, validate, display results
- [x] T547 [US6] Update `internal/session/session.go`: add
  `AuthoringArtifactType string` and
  `AuthoringProgress string` fields with
  `SetAuthoringState(artifactType string,
  progress string)` and `GetAuthoringState() (string,
  string)` methods
- [x] T548 [US6] Integration test: full authoring flow
  for a ThreatCatalog — select artifact type, complete
  all steps with valid field values, verify final YAML
  output structure
- [x] T549 [US6] Integration test: authoring flow with
  validation errors — enter invalid field values,
  verify errors are displayed with fix suggestions,
  correct values, verify successful completion
- [x] T550 [US6] Verify `make build`, `make test`,
  `make lint` pass with zero errors and zero warnings

**Checkpoint**: US6 is fully functional. A user can select
an artifact type, walk through guided authoring steps with
role-specific explanations and value suggestions, receive
validation feedback at each step, and produce a valid YAML
artifact that passes `cue vet` validation.

---

## Phase 6 — MCP Wizard Delegation (spec delta, AS4/AS5)

These tasks were added after the spec was updated to support
MCP prompt-assisted authoring (FR-038). When the MCP server
is in artifact mode, the system offers MCP wizards as an
alternative to the built-in guided authoring flow for
ThreatCatalog and ControlCatalog.

- [x] T551 [US6] Write failing test
  `internal/cli/author_prompt_test.go`: when session is in
  artifact mode and user selects ThreatCatalog, system
  offers choice between MCP `threat_assessment` wizard and
  built-in authoring flow
- [x] T552 [US6] Write failing test
  `internal/cli/author_prompt_test.go`: when session is in
  artifact mode and user selects ControlCatalog, system
  offers choice between MCP `control_catalog` wizard and
  built-in authoring flow
- [x] T553 [US6] Write failing test
  `internal/cli/author_prompt_test.go`: when session is in
  advisory mode, ThreatCatalog authoring uses only built-in
  flow (no wizard offer)
- [x] T554 [US6] Write failing test
  `internal/cli/author_prompt_test.go`: for GuidanceCatalog
  (no corresponding MCP prompt), built-in flow is always
  used regardless of mode
- [x] T555 [P] [US6] Implement wizard-vs-builtin choice in
  `RunGuidedAuthoring` in `internal/cli/author_prompt.go`:
  check if artifact type has a corresponding MCP prompt and
  session is in artifact mode; if so, offer choice; delegate
  to `RunWizardLauncher` if wizard selected
- [x] T556 [US6] Verify `make build`, `make test`,
  `make lint` pass with zero errors and zero warnings

**Checkpoint**: US6 fully aligned with updated spec. Users in
artifact mode can choose between MCP wizard and built-in
authoring for ThreatCatalog and ControlCatalog.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** (Data Model): No US6-internal dependencies —
  start immediately (depends on US4 being complete)
- **Phase 2** (Engine): Depends on Phase 1 (uses
  AuthoredArtifact, ArtifactTemplate, StepField types)
- **Phase 3** (Validation): Depends on Phase 1 (uses
  AuthoredArtifact type) and Phase 2 (engine calls
  validation)
- **Phase 4** (Output): Depends on Phase 1 (uses
  AuthoredArtifact type)
- **Phase 5** (CLI Integration): Depends on Phases 2, 3,
  and 4

### Parallel Opportunities

```text
Phase 1 (Data Model)
    │
    ▼
Phase 2 (Engine)
    │
    ├───────────────────┐
    ▼                   ▼
Phase 3 (Validation) Phase 4 (Output)  ← parallel
    │                   │
    └───────┬───────────┘
            ▼
    Phase 5 (CLI Integration)
```

Within each phase, all tasks marked `[P]` can run in
parallel. All test tasks within a phase can run in parallel.

---

## Notes

- All test tasks follow TDD: write test, confirm it fails,
  then implement.
- Each task produces files with SPDX headers and passes
  `make lint`.
- Commit after each phase per Conventional Commits format.
- US6 tasks numbered T501-T550 to avoid conflicts with US1
  (T001-T054), US2 (T101-T127), US3 (T201-T260), US4
  (T301-T331), and US5 (T401-T431).
- Key invariant: a Security Engineer authoring a
  ThreatCatalog MUST receive role-specific guidance at
  each step (acceptance scenario 1), and the final YAML
  output MUST pass `cue vet -c -d '#ThreatCatalog'`
  (acceptance scenario 3).
- Validation at each step (acceptance scenario 2) uses MCP
  when available, local cue vet when not.
