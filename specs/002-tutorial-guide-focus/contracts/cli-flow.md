# Contract: CLI Flow Changes

**Branch**: `002-tutorial-guide-focus`
**Date**: 2026-03-17

## Overview

Pac-Man is a CLI tool used within OpenCode sessions. Its
primary external interface is the interactive setup flow
that chains MCP setup, version resolution, role discovery,
and tutorial navigation. This contract documents the
changes to that flow.

## Current Setup Flow (Before)

```text
RunSetup
  │
  ├─ runMCPSetup
  │    └─ Detect/install MCP server → Session
  │
  ├─ RunVersionSelection          ← USER PROMPT
  │    ├─ Fetch/cache releases
  │    ├─ Present Stable/Latest options
  │    ├─ User selects version
  │    └─ SelectVersion → mutate Session.SchemaVersion
  │
  └─ RunRoleDiscovery
       ├─ Role identification
       ├─ Activity probing
       └─ generateLearningPath
```

## New Setup Flow (After)

```text
RunSetup
  │
  ├─ runMCPSetup
  │    └─ Detect/install MCP server → Session
  │
  ├─ AutoSelectLatest             ← NO PROMPT
  │    ├─ Fetch/cache releases
  │    ├─ DetermineVersions
  │    ├─ SelectVersion(Latest)
  │    ├─ Display: selected version + warnings
  │    └─ Mutate Session.SchemaVersion
  │
  └─ RunRoleDiscovery
       ├─ Role identification
       ├─ Activity probing
       ├─ ArtifactRecommendations ← NEW
       └─ generateLearningPath
```

## New Tutorial Completion Flow

```text
Tutorial player (existing)
  │
  ├─ Navigate sections (unchanged)
  │
  ├─ Mark tutorial complete
  │    ├─ ✓ Completed: <title>    (existing)
  │    │
  │    ├─ BuildHandoffSummary     ← NEW
  │    │    ├─ Map layer → artifact types
  │    │    ├─ Look up MCP wizard name
  │    │    ├─ Check MCP availability
  │    │    ├─ Collect key decisions from sections
  │    │    └─ Load preparation checklist
  │    │
  │    └─ RenderHandoffSummary    ← NEW
  │         ├─ Artifact type and schema def
  │         ├─ MCP prompt to use
  │         ├─ Key decisions list
  │         ├─ Preparation checklist
  │         ├─ Version warnings (if any)
  │         └─ Next steps instructions
  │
  └─ Return to tutorial list
```

## Function Contracts

### schema.AutoSelectLatest

```
AutoSelectLatest(
    ctx       context.Context,
    fetcher   ReleaseFetcherFn,
    cachePath string,
    sess      *session.Session,
) (*SelectionResult, error)
```

**Preconditions**:
- `fetcher` is a valid release fetcher function (may be
  nil if cache exists)
- `cachePath` is a valid file path for cache storage
- `sess` is a non-nil Session with SchemaVersion == ""

**Postconditions**:
- `sess.SchemaVersion` is set to the latest release tag
- Returns `SelectionResult` with `SelectedTag` and
  `ExperimentalSchemas`
- On network failure with valid cache: uses cached
  releases, returns successfully
- On network failure without cache: returns error

**Error Cases**:
- No releases available (upstream or cache): returns
  `ErrNoVersionAvailable`
- All releases fail parsing: returns wrapped error

### roles.ArtifactRecommendations

```
ArtifactRecommendations(
    profile *ActivityProfile,
) []ArtifactRecommendation
```

**Preconditions**:
- `profile` is non-nil with at least one entry in
  `ResolvedLayers`

**Postconditions**:
- Returns one `ArtifactRecommendation` per unique
  artifact type across all resolved layers
- Recommendations are ordered by layer confidence
  (Strong first), then by layer number
- Duplicate artifact types (same type from multiple
  layers) are deduplicated, keeping the highest
  confidence entry

**Edge Cases**:
- Layers with no artifacts (L4, L6, L7): produce no
  recommendations (silently skipped)
- Empty `ResolvedLayers`: returns empty slice

### cli.BuildHandoffSummary

```
BuildHandoffSummary(
    step    *tutorials.PathStep,
    sess    *session.Session,
    selRes  *schema.SelectionResult,
) *HandoffSummary
```

**Preconditions**:
- `step` is a completed PathStep
- `sess` has SchemaVersion set
- `selRes` may be nil (experimental schemas will be empty)

**Postconditions**:
- Returns a fully populated `HandoffSummary`
- `ArtifactType` is the first artifact type for the
  step's layer (from `consts.LayerArtifacts`)
- `MCPPrompt` is set if a wizard exists for the type
- `MCPResources` always includes `gemara://lexicon` and
  `gemara://schema/definitions`
- `MCPTools` always includes `validate_gemara_artifact`
- `MCPConfigured` reflects whether gemara-mcp is in
  `opencode.json`
- `KeyDecisions` are derived from the step's
  `PrimarySections` (sections with highest relevance)
- `PreparationChecklist` is loaded from
  `consts.DefaultPreparationChecklists`

**Edge Cases**:
- Layer with no artifacts (L4): `ArtifactType` is empty,
  summary notes "No artifact types are defined for this
  layer"
- MCP not configured: `MCPConfigured` is false, summary
  instructs user to run `./pacman --doctor` and set up
  the gemara-mcp server in `opencode.json`

### cli.RenderHandoffSummary

```
RenderHandoffSummary(
    summary *HandoffSummary,
    out     io.Writer,
)
```

**Preconditions**:
- `summary` is a non-nil `HandoffSummary`
- `out` is a valid writer

**Postconditions**:
- Renders the handoff summary using established styles:
  - Divider
  - Card header: "Ready to Author: <ArtifactType>"
  - Label-value pairs for schema, wizard, version
  - Available MCP resources and tools list
  - Key decisions as a numbered list
  - Preparation checklist as a bulleted list
  - Version mismatch warning (if applicable)
  - Next steps section directing to OpenCode
- All output MUST be visually polished and accessible for
  non-technical audiences (FR-018): use clear labels, avoid
  unexplained jargon, use consistent card styling

**Output Format** (visual structure):

```text
────────────────────────────────────────────
┃ Ready to Author: Threat Catalog
┃
┃ Schema:  #ThreatCatalog
┃ Version: v0.20.0
┃ Wizard:  threat_assessment
┃
┃ Available in OpenCode:
┃   Tools:     validate_gemara_artifact
┃   Resources: gemara://lexicon
┃              gemara://schema/definitions
┃   Prompts:   threat_assessment
┃
┃ Key Decisions:
┃   1. Component scope and boundaries
┃   2. MITRE ATT&CK alignment
┃   3. Import source selection
┃
┃ Preparation Checklist:
┃   • Identify the component to assess
┃   • Determine scope boundaries
┃   • Decide on catalog import source
┃   • Consider MITRE ATT&CK alignment
┃
┃ Next: Open an OpenCode Session
┃
┃   Launch opencode and use the
┃   threat_assessment prompt with the
┃   gemara-mcp server to begin guided
┃   authoring.
┃
┃   ┌──────────────────────────────────┐
┃   │ $ opencode                       │
┃   │                                  │
┃   │ Then tell the AI:                │
┃   │ "Run the threat_assessment       │
┃   │  wizard for my component"        │
┃   └──────────────────────────────────┘
────────────────────────────────────────────
```
