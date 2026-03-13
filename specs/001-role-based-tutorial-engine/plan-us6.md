# Implementation Plan: US6 — Guided Gemara Content
# Authoring (P6)

**Branch**: `001-role-based-tutorial-engine` | **Date**: 2026-03-13
**Spec**: [spec.md](spec.md) — User Story 6
**Depends on**: US1 (MCP Server Setup), US2 (Schema Version
Selection), US3 (Role and Activity Discovery), US4 (Reusable
Content Transformation) — all completed

## Summary

US6 implements guided authoring for Gemara artifacts. Based on
the user's role and the selected artifact type, the system
provides step-by-step guidance that mirrors the structure of
the corresponding Gemara tutorial. Each authoring step explains
the field being authored, suggests values based on the user's
stated scope, and validates the in-progress artifact against
the Gemara CUE schema. The final output is a YAML (or JSON)
document that passes full `cue vet` validation.

This is the capstone feature — it turns learning into
production artifacts. The authoring flow draws on:
- Role and activity context from US3 (personalizing
  explanations and example values)
- Reusable content blocks from US4 (providing patterns,
  naming conventions, and schema structure references)
- Schema validation via MCP (US1) or local fallback
- The user's selected schema version from US2

This plan covers FR-009, FR-010, FR-012, FR-019 (authoring
portion), and the US6 acceptance scenarios 1-3.

## Technical Context

**Language/Version**: Go 1.26.1
**Dependencies**: Existing `internal/blocks/` (content block
retrieval for contextual help), `internal/roles/` (role
context for personalization), `internal/fallback/` (local CUE
validation), `internal/mcp/` (MCP-based validation),
`internal/consts/` (artifact type and schema constants),
`internal/cli/` (TUI styles, prompter interfaces),
`internal/session/` (session state)
**Storage**: Authored artifacts written to the current working
directory or a user-specified output path
**Testing**: `go test ./...` via `make test`; TDD per
constitution
**Constraints**: The authoring flow is role-aware but does not
modify the user's role profile. Validation uses the session's
selected schema version consistently (FR-019). Output in both
YAML and JSON per FR-012.

## Constitution Check

| Principle | Status | Notes |
|:---|:---|:---|
| I. Schema Conformance | Pass | Every step validates against the Gemara CUE schema; final output passes `cue vet` |
| II. Gemara Layer Fidelity | Pass | Artifact types map to specific layers; guidance references correct layer context |
| III. TDD | Pass | Tests written before implementation per phase |
| IV. Tutorial-First Design | Pass | Authoring steps mirror tutorial structure; content blocks provide contextual examples |
| V. Incremental Delivery | Pass | US6 is independently usable after US1-US4 |
| VI. Decision Documentation | N/A | No new ADRs anticipated; patterns consistent with prior stories |
| VII. Centralized Constants | Pass | Artifact types, schema defs already in `internal/consts/` |
| VIII. Composability | Pass | Authoring engine is an independent operation composable with blocks and roles |
| IX. Convention Over Configuration | Pass | Default artifact structures follow Gemara tutorial patterns |

## Source Code

```text
internal/
├── authoring/
│   ├── model.go              # AuthoredArtifact,
│   │                         #   ArtifactSection,
│   │                         #   AuthoringStep,
│   │                         #   StepField types
│   ├── model_test.go
│   ├── engine.go             # AuthoringEngine: step
│   │                         #   sequencing, field
│   │                         #   guidance, value
│   │                         #   suggestions
│   ├── engine_test.go
│   ├── validate.go           # Step-level and full
│   │                         #   artifact validation
│   │                         #   (MCP or local fallback)
│   ├── validate_test.go
│   ├── output.go             # YAML/JSON rendering,
│   │                         #   file writing
│   └── output_test.go
├── cli/
│   ├── author_prompt.go      # CLI flow: artifact type
│   │                         #   selection, step-by-step
│   │                         #   guided authoring,
│   │                         #   validation display,
│   │                         #   final output
│   ├── author_prompt_test.go
│   └── styles.go             # Update: add authoring
│                             #   step rendering styles
├── consts/
│   └── consts.go             # Update: add authoring
│                             #   constants (output dir,
│                             #   artifact section names)
└── session/
    └── session.go            # Update: add authoring
                              #   state to session
```

## Implementation Phases

### Phase 1: Authored Artifact Data Model (FR-009)

- Define the `ArtifactType` type: a string enum referencing
  the artifact types from `internal/consts/` that have CUE
  schemas (GuidanceCatalog, ControlCatalog, ThreatCatalog,
  Policy, MappingDocument, EvaluationLog).
- Define the `StepField` type: field name, field description,
  field type (string, list, map, enum), required flag, example
  value, help text sourced from content blocks.
- Define the `AuthoringStep` type: step name, step
  description, role-specific explanation (why this matters
  for your role), list of `StepField` entries, validation
  schema subset (the CUE path for partial validation),
  completion status.
- Define the `ArtifactSection` type: section name (e.g.,
  "metadata", "scope", "capabilities", "threats"),
  associated authoring step, completed field values.
- Define the `AuthoredArtifact` type: artifact type (from
  consts), target schema definition (e.g., `#ThreatCatalog`),
  target schema version, list of `ArtifactSection` entries,
  validation status (not validated, partial, valid, invalid),
  authoring role name, created/updated timestamps.
- Define the `ArtifactTemplate` type: artifact type,
  ordered list of `AuthoringStep` definitions (the recipe
  for building this artifact type), Gemara layer, related
  tutorial references.
- Add authoring constants to `internal/consts/consts.go`:
  `AuthoringOutputDir` (default output subdirectory),
  artifact section name constants.
- Tests: AuthoredArtifact construction, section completion
  tracking, validation status transitions, artifact type
  to schema definition mapping.

### Phase 2: Authoring Step Engine (FR-009, FR-010)

- Implement `ArtifactTemplates()` — return the predefined
  authoring templates for each supported artifact type.
  Each template defines the ordered sequence of steps,
  the fields within each step, and example values drawn
  from the Gemara tutorials. Templates are defined as data
  (not code logic) so they can be extended.
- Implement `NewAuthoringEngine(template ArtifactTemplate,
  role string, keywords []string,
  blocks *BlockIndex) *AuthoringEngine` — initialize an
  engine with role-specific context.
- Implement `CurrentStep() *AuthoringStep` — return the
  current authoring step with role-personalized guidance.
- Implement `SetFieldValue(field string,
  value string) error` — record a field value within the
  current step.
- Implement `CompleteStep() ([]ValidationError,
  error)` — mark the current step as complete, trigger
  step-level validation, advance to the next step.
- Implement `GetSuggestions(field string) []string` —
  return suggested values for a field based on the user's
  role, keywords, and relevant content blocks.
- Implement `Progress() (completed int,
  total int)` — return authoring progress.
- Implement `BuildArtifact() *AuthoredArtifact` — assemble
  the completed sections into an authored artifact.
- Tests: engine initialization with each artifact type,
  step progression, field value recording, suggestion
  generation, progress tracking, artifact assembly.

### Phase 3: Step-Level and Full Validation (FR-010)

- Define `ValidationError` type: field path, error message,
  fix suggestion.
- Define `Validator` interface with `ValidatePartial(
  artifact *AuthoredArtifact,
  step int) ([]ValidationError, error)` and
  `ValidateFull(
  artifact *AuthoredArtifact) ([]ValidationError, error)`.
- Implement `MCPValidator` — uses the MCP client's
  `ValidateArtifact` tool for validation when the MCP
  server is available. Translates MCP validation responses
  into `ValidationError` entries with actionable fix
  suggestions.
- Implement `LocalValidator` — uses the fallback
  `ValidateLocal` (cue vet) when MCP is unavailable.
  Parses CUE error output into `ValidationError` entries.
- Implement `NewValidator(session *Session) Validator` —
  factory that returns MCPValidator or LocalValidator based
  on session state.
- Step-level validation: after each step, serialize the
  completed fields so far and validate against the schema.
  Report errors with field-specific fix suggestions.
- Full validation: after all steps, validate the complete
  artifact. This is the final gate before output.
- Tests: MCPValidator with mock MCP client, LocalValidator
  with mock CUE runner, validation error parsing,
  step-level partial validation, full artifact validation.

### Phase 4: YAML/JSON Output Generation (FR-012)

- Implement `RenderYAML(
  artifact *AuthoredArtifact) ([]byte, error)` — serialize
  the authored artifact to YAML following Gemara naming
  conventions documented in the tutorials.
- Implement `RenderJSON(
  artifact *AuthoredArtifact) ([]byte, error)` — serialize
  the authored artifact to JSON.
- Implement `WriteArtifact(artifact *AuthoredArtifact,
  outputDir string, format string) (string, error)` — write
  the artifact to disk. File naming follows Gemara
  conventions (e.g., `ORG.PROJ.COMPONENT.THR##` for threat
  catalogs). Returns the output file path.
- Implement `GenerateFilename(
  artifact *AuthoredArtifact) string` — generate a
  filename following Gemara naming patterns based on the
  artifact type and metadata fields.
- Tests: YAML output matches expected structure, JSON output
  matches expected structure, file writing creates valid
  files, filename generation follows naming conventions,
  both formats produce content that round-trips correctly.

### Phase 5: CLI Integration and TUI Rendering

- Create `internal/cli/author_prompt.go`:
  - `AuthorPromptConfig` struct: Prompter (free text),
    Session, TutorialsDir, SchemaVersion, OutputDir,
    OutputFormat, BlockIndex (for contextual help).
  - `AuthorPromptResult` struct: Artifact
    *AuthoredArtifact, OutputPath string,
    ValidationErrors []ValidationError.
  - `RunGuidedAuthoring(cfg *AuthorPromptConfig,
    out io.Writer) (*AuthorPromptResult, error)` — full
    authoring flow:
    1. Present available artifact types (filtered by role
       if role context is available)
    2. Initialize authoring engine with selected template
    3. Walk through each step: display step guidance, prompt
       for field values, show suggestions, validate
    4. Display validation results with fix guidance
    5. On completion, validate full artifact and render
       output
  - `RunResumeAuthoring(cfg *AuthorPromptConfig,
    artifact *AuthoredArtifact,
    out io.Writer) (*AuthorPromptResult, error)` — resume
    an in-progress artifact from a saved state.
- Add rendering functions to `internal/cli/styles.go`:
  - `RenderAuthoringStep(step *AuthoringStep,
    current int, total int) string` — step card with
    progress indicator, field list, and role-specific
    explanation.
  - `RenderFieldPrompt(field *StepField) string` — field
    input prompt with description and example value.
  - `RenderValidationResults(
    errs []ValidationError) string` — validation error
    display with fix suggestions.
  - `RenderArtifactSummary(
    artifact *AuthoredArtifact) string` — completed
    artifact overview.
  - `RenderAuthoringProgress(completed int,
    total int) string` — progress bar or indicator.
- Update `internal/session/session.go`: add
  `AuthoringArtifactType string` and
  `AuthoringProgress string` fields with
  `SetAuthoringState(artifactType string,
  progress string)` method.
- Wire into setup flow: after role discovery and optional
  team configuration, offer guided authoring as the next
  step.
- Integration test: full authoring flow for a
  ThreatCatalog, verify step progression, validation at
  each step, and final YAML output.
- Verify `make build`, `make test`, `make lint` pass.

## Dependencies & Execution Order

```text
Phase 1 (Data Model)
        │
        ▼
Phase 2 (Authoring Engine)
        │
        ▼
Phase 3 (Validation)
        │
        ▼
Phase 4 (Output Generation)
        │
        ▼
Phase 5 (CLI Integration & TUI)
```

All phases are sequential. Each phase depends on the
preceding one.

## Complexity Tracking

No constitution violations identified. No complexity
justifications required.
