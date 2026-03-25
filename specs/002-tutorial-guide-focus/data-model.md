# Data Model: Refocus Gemara User Journey as Tutorial Guide

**Branch**: `002-tutorial-guide-focus`
**Date**: 2026-03-25

## Entity Overview

This feature introduces two new entities, modifies two
existing entities, and defines a documentation structure
for US6 (README restructuring). No entities are removed.

```text
                    ┌──────────────────┐
                    │  ActivityProfile  │ (existing, extended)
                    │                  │
                    │ + Recommendations│─────┐
                    └────────┬─────────┘     │
                             │               │
                    resolves to               │
                             │               │
                    ┌────────▼─────────┐     │
                    │   LearningPath   │     │
                    │                  │     │
                    │   Steps[]        │     │
                    └────────┬─────────┘     │
                             │               │
                    each step has            │
                             │               │
                    ┌────────▼─────────┐     │
                    │    PathStep      │     │
                    │  (existing)      │     │
                    │ + HandoffInfo    │─────┤
                    └────────┬─────────┘     │
                             │               │
                    on completion             │
                             │               │
                    ┌────────▼─────────┐     │
                    │ HandoffSummary   │◄────┘
                    │  (NEW)           │
                    └────────┬─────────┘
                             │
                    references
                             │
                    ┌────────▼─────────┐
                    │ArtifactRecommend.│
                    │  (NEW)           │
                    └──────────────────┘


    ┌──────────────────────────────────────────────────┐
    │        Documentation Structure (US6)              │
    │                                                  │
    │  README.md (landing page)                        │
    │    ├── links to docs/layer-reference.md           │
    │    ├── links to docs/project-structure.md         │
    │    ├── links to docs/mcp-update-guide.md          │
    │    ├── links to CONTRIBUTING.md                   │
    │    └── embeds docs/images/web-ui-preview.png      │
    └──────────────────────────────────────────────────┘
```

## New Entities

### ArtifactRecommendation

Represents a recommended artifact type for the user based
on their resolved Gemara layers.

**Package**: `internal/roles`

| Field | Type | Description | Source |
|-------|------|-------------|--------|
| ArtifactType | string | Artifact type identifier (e.g., "ThreatCatalog") | `consts.LayerArtifacts[layer]` |
| SchemaDef | string | CUE schema definition (e.g., "#ThreatCatalog") | `consts.Schema*` constants |
| Description | string | One-sentence user-facing description | New `consts.ArtifactDescriptions` map |
| Layer | int | Gemara layer number (1-7) | From `ActivityProfile.ResolvedLayers` |
| Confidence | Confidence | Inferred or Strong, from the layer mapping | From `LayerMapping.Confidence` |
| MCPWizard | string | MCP wizard prompt name, or empty | New `consts.ArtifactWizards` map |
| AuthoringApproach | string | "wizard" or "collaborative" | Derived from MCPWizard presence |

**Validation Rules**:
- ArtifactType MUST be one of the 6 values in
  `consts.Artifact*` constants
- SchemaDef MUST be the corresponding value from
  `authoring.ArtifactTypeToSchema()`
- Layer MUST be between 1 and 7
- Confidence MUST be `ConfidenceInferred` or
  `ConfidenceStrong`
- MCPWizard MUST be empty or one of
  `consts.WizardThreatAssessment`,
  `consts.WizardControlCatalog`

**Relationships**:
- Derived from `ActivityProfile.ResolvedLayers`
- Referenced by `HandoffSummary`
- Maps to `consts.LayerArtifacts` and
  `consts.ArtifactTypeSections`

### HandoffSummary

A structured transition context presented to the user after
completing a tutorial in the Gemara User Journey terminal. Bridges the
learn-to-author transition by directing the user to
OpenCode with the gemara-mcp server.

**Package**: `internal/cli`

| Field | Type | Description | Source |
|-------|------|-------------|--------|
| ArtifactType | string | Target artifact type for authoring | Tutorial layer -> `consts.LayerArtifacts` |
| SchemaDef | string | CUE schema definition for validation | `consts.Schema*` constants |
| MCPPrompt | string | MCP wizard prompt name, or empty | `consts.ArtifactWizards` |
| MCPResources | []string | Available MCP resources (lexicon, schema docs) | `consts.Resource*` constants |
| MCPTools | []string | Available MCP tools (validate_gemara_artifact) | `consts.Tool*` constants |
| MCPConfigured | bool | Whether gemara-mcp is in opencode.json | `mcp.ReadOpenCodeConfig()` |
| ServerMode | string | MCP server mode (advisory/artifact) | `session.GetServerMode()` |
| SchemaVersion | string | Auto-selected schema version tag | `session.SchemaVersion` |
| ExperimentalSchemas | []string | Schemas with experimental status | From `SelectionResult` |
| VersionMismatch | bool | MCP version != selected version | From `mcp.CheckCompatibility` |
| KeyDecisions | []string | Decisions the user should have answers for | Derived from tutorial sections |
| PreparationChecklist | []string | Items to prepare before authoring | Static per artifact type |
| TutorialTitle | string | Title of the completed tutorial | From `PathStep.Tutorial.Title` |
| Layer | int | Gemara layer of the completed tutorial | From `PathStep.Layer` |

**Validation Rules**:
- ArtifactType MUST NOT be empty
- SchemaDef MUST correspond to ArtifactType
- If MCPPrompt is non-empty, it MUST be a known wizard name
- MCPResources MUST contain at least `gemara://lexicon`
  and `gemara://schema/definitions`
- MCPTools MUST contain at least `validate_gemara_artifact`
- SchemaVersion MUST NOT be empty (set by auto-selection)
- KeyDecisions SHOULD have at least one entry
- PreparationChecklist SHOULD have at least one entry

**Relationships**:
- Generated from a completed `PathStep`
- References `Session` for MCP status and schema version
- Contains `ArtifactRecommendation` data (flattened)
- Rendered by `RenderHandoffSummary` in `handoff.go`
- Directs user to OpenCode as the authoring environment

**State**: HandoffSummary is a value object with no state
transitions. It is created once at tutorial completion and
rendered immediately. It is not persisted.

## Modified Entities

### ActivityProfile (existing)

**Package**: `internal/roles`
**File**: `activities.go`

| Change | Field | Type | Description |
|--------|-------|------|-------------|
| ADD | Recommendations | []ArtifactRecommendation | Artifact types recommended for the user |

The `Recommendations` field is populated by the new
`ArtifactRecommendations()` function, which iterates
`ResolvedLayers`, looks up `consts.LayerArtifacts` for each
layer, and constructs an `ArtifactRecommendation` for each
artifact type found.

**Impact**: The `ActivityProfile` struct is used in:
- `roles.ResolveLayerMappings()` (constructor)
- `tutorials.GeneratePath()` (consumer)
- `roles.ProfileFromActivityProfile()` (serialization)
- `cli.RunRoleDiscovery()` (consumer)

Adding a new field does not break existing consumers
because Go struct fields are zero-valued by default
(empty slice). The `Recommendations` field is populated
after `ResolveLayerMappings` returns, by calling
`ArtifactRecommendations(profile)`.

### SelectionResult (existing)

**Package**: `internal/schema`
**File**: `selector.go`

No field changes needed. The existing `SelectionResult`
already contains:
- `SelectedTag string`
- `ExperimentalSchemas []string`
- `CompatWarning string`

These are sufficient for the handoff summary. The new
`AutoSelectLatest` function returns `*SelectionResult`
directly.

### Session (existing)

**Package**: `internal/session`
**File**: `session.go`

No structural changes needed. The existing session fields
are sufficient:
- `SchemaVersion` — set by auto-selection
- `mcpStatus` — checked for handoff summary
- `ServerMode` — checked for wizard availability
- `Capabilities` — checked for MCP prompt availability

## New Constants

**Package**: `internal/consts`
**File**: `consts.go`

### ArtifactDescriptions

Map of artifact type to one-sentence user-facing
description:

| Key | Description |
|-----|-------------|
| GuidanceCatalog | "A structured catalog of standards, best practices, and regulatory requirements that your organization follows." |
| ControlCatalog | "A catalog of security controls that mitigate identified threats, with assessment requirements and evidence criteria." |
| ThreatCatalog | "A catalog of threats to a specific component, organized by capability, with severity and likelihood assessments." |
| Policy | "An organizational policy document defining adherence requirements, timelines, and scope for a set of controls." |
| MappingDocument | "A cross-reference document that maps controls to guidance items, establishing traceability between layers." |
| EvaluationLog | "An assessment log recording control evaluations, evidence collected, and compliance findings." |

### ArtifactWizards

Map of artifact type to MCP wizard prompt name:

| Key | Wizard |
|-----|--------|
| ThreatCatalog | "threat_assessment" |
| ControlCatalog | "control_catalog" |

All other artifact types map to empty string (no wizard;
use collaborative authoring with MCP resources).

### AuthoringApproach Constants

| Constant | Value | Description |
|----------|-------|-------------|
| ApproachWizard | "wizard" | MCP wizard-guided authoring |
| ApproachCollaborative | "collaborative" | Collaborative authoring with MCP resources (lexicon, schema docs) and validation |

### DefaultPreparationChecklists

Static preparation checklists per artifact type,
containing the key items a user should have ready before
starting authoring:

**ThreatCatalog**:
- "Identify the component or system to assess"
- "Determine scope boundaries (what is in/out)"
- "Decide whether to import from an existing catalog
  (e.g., FINOS CCC Core)"
- "Consider MITRE ATT&CK alignment preference"

**ControlCatalog**:
- "Identify the component or system to protect"
- "Select the guideline framework(s) to align with"
- "Determine scope boundaries"
- "Decide whether to import from an existing catalog"

**GuidanceCatalog**:
- "Identify the standard, regulation, or best practice
  to codify"
- "Determine scope and applicability"
- "Gather source material (regulatory text, standard
  sections)"

**Policy**:
- "Identify the controls this policy governs"
- "Define the adherence timeline"
- "Determine compliance scope (teams, systems, regions)"
- "Establish non-compliance handling procedures"

**MappingDocument**:
- "Identify source and target catalogs to map"
- "Determine relationship types (implements, equivalent,
  subsumes)"
- "Gather entry references for both catalogs"

**EvaluationLog**:
- "Identify the controls to evaluate"
- "Gather evidence and assessment materials"
- "Determine evaluation criteria and scoring"

## Documentation Structure (US6)

### New Files

| File | Content Source | Purpose |
|------|---------------|---------|
| `docs/images/web-ui-preview.png` | Manually captured screenshot | Visual preview of web UI in README |
| `docs/layer-reference.md` | Moved from README lines 142-161 | Gemara seven-layer model reference |
| `docs/project-structure.md` | Moved from README lines 407-452 | Project directory layout |
| `docs/mcp-update-guide.md` | Moved from README lines 329-388 | Instructions for syncing gemara-mcp |

### Modified Files

| File | Change |
|------|--------|
| `README.md` | Rewritten as landing page (~120-150 lines); links to `docs/` files |

### README Data Model

The README is a static Markdown document with no runtime
data. Its content is derived from:

| Section | Data Source |
|---------|------------|
| Project summary | Manually authored; positions as tutorial guide |
| Screenshot | `docs/images/web-ui-preview.png` (relative path) |
| User Journey | Narrative describing: role discovery -> tutorial walkthrough -> OpenCode handoff |
| Prerequisites | Hyperlinks to official installation pages for Go, CUE, OpenCode, Git, gemara-mcp |
| Getting Started | 4-step instructions: clone, build, verify (`--doctor`), launch (`opencode`) |
| Upstream Projects | Table linking Gemara and gemara-mcp repositories |
| Learn More | Links to `docs/layer-reference.md`, `docs/project-structure.md`, `docs/mcp-update-guide.md`, `CONTRIBUTING.md` |
| License | Apache 2.0 one-liner |
