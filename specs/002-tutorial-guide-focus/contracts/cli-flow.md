# Contract: CLI Flow Changes

**Branch**: `002-tutorial-guide-focus`
**Date**: 2026-03-25

## Overview

Gemara User Journey is a CLI tool used within OpenCode sessions. Its
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
  instructs user to run `./journey --doctor` and set up
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

## Documentation Contract (US6)

### README.md Structure

The README MUST follow this section order:

```text
# Gemara User Journey
  │
  ├─ One-paragraph summary (distinguish from MCP server)
  │
  ├─ Screenshot (docs/images/web-ui-preview.png)
  │
  ├─ ## User Journey
  │    ├─ 1. Discover (role + activity identification)
  │    ├─ 2. Learn (tutorial walkthrough)
  │    └─ 3. Author (handoff to OpenCode + gemara-mcp)
  │
  ├─ ## Prerequisites
  │    └─ Hyperlinked dependency list (Go, CUE,
  │       OpenCode, Git, gemara-mcp)
  │
  ├─ ## Getting Started
  │    ├─ Step 1: Clone and build
  │    ├─ Step 2: Verify environment
  │    ├─ Step 3: Launch OpenCode
  │    └─ Step 4: Tell OpenCode your role
  │
  ├─ ## Upstream Projects
  │    └─ Compact table (Gemara, gemara-mcp)
  │
  ├─ ## Learn More
  │    ├─ Link: docs/layer-reference.md
  │    ├─ Link: docs/project-structure.md
  │    ├─ Link: docs/mcp-update-guide.md
  │    └─ Link: CONTRIBUTING.md
  │
  └─ ## License
       └─ One-liner: Apache 2.0
```

**Constraints**:
- Total length: ~120-150 lines
- No inline platform-specific install commands (link to
  official pages instead)
- No `<details>` collapsible HTML sections
- Screenshot referenced via relative path
- All dependency names MUST be hyperlinked to official
  installation pages

### docs/ File Contracts

Each displaced file MUST:
- Include a title and brief context paragraph
- Contain the full content from the original README section
- Be self-contained (readable without the README)
- Include a link back to the README for navigation

| File | Source | Minimum Content |
|------|--------|-----------------|
| `docs/layer-reference.md` | README lines 142-161 | 7-layer table with layer names and purposes |
| `docs/project-structure.md` | README lines 407-452 | Directory tree with descriptions |
| `docs/mcp-update-guide.md` | README lines 329-388 | Sync instructions for both clone and fork workflows |
| `docs/images/web-ui-preview.png` | Manual capture | Screenshot of Results view from web UI |
