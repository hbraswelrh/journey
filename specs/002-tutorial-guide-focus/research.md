# Research: Refocus Pac-Man as Tutorial Guide

**Branch**: `002-tutorial-guide-focus`
**Date**: 2026-03-17
**Status**: Complete

## R1: Bypassing Version Selection Without Breaking Setup

### Decision

Replace the interactive `RunVersionSelection` call in
`RunSetup` with a non-interactive auto-selection path that
calls `schema.RefreshOrCache`, `schema.DetermineVersions`,
and `schema.SelectVersion` with `SelectionLatest` and no
confirmer.

### Rationale

The existing `SelectVersion` function already accepts `nil`
for both `mcpClient` and `confirmer` parameters. When
`sess.SchemaVersion` is empty (initial selection), the
mid-session switch logic is skipped entirely. The three-call
sequence (`RefreshOrCache` -> `DetermineVersions` ->
`SelectVersion`) is the minimal path that resolves a version
dynamically, handles cache fallback, and sets
`sess.SchemaVersion` atomically.

The alternative of calling `DetermineLatestVersion` directly
and assigning `sess.SchemaVersion = latest.Tag` was
considered but rejected because it bypasses the
`SelectionResult` which provides `ExperimentalSchemas`
warnings needed for the handoff summary (FR-016).

### Alternatives Considered

1. **Direct field assignment**: `sess.SchemaVersion = tag`.
   Simplest but loses experimental schema detection and
   cache management. Rejected.
2. **Pass version at session construction**: Change
   `NewSessionWithMCP("", mode)` to pass a pre-resolved
   tag. Rejected because release fetching depends on
   network, which should happen after session creation
   so fallback mode can be set if network fails.
3. **Add `AutoSelectLatest` wrapper function**: New function
   in `internal/schema/` that encapsulates the three-call
   sequence. Chosen approach — clean, testable, reusable.

### Implementation Detail

New function in `internal/schema/selector.go`:

```
AutoSelectLatest(ctx, fetcher, cachePath, sess)
    -> (*SelectionResult, error)
```

This wraps `RefreshOrCache` + `DetermineVersions` +
`SelectVersion(choice, SelectionLatest, sess, nil, nil)`.

In `internal/cli/setup.go`, replace the
`RunVersionSelection` call with `schema.AutoSelectLatest`.
The `VersionPromptConfig` struct and `RunVersionSelection`
function remain in `version_prompt.go` untouched.

### Downstream Impact

All downstream code that reads `Session.SchemaVersion`
continues to work because the field is set by the same
`SelectVersion` function. Specifically:
- Tutorial version compat checks in `role_prompt.go`
- Learning path mismatch detection in `path.go`
- Session status rendering in `main.go`
- Authored artifact stamping in `author_prompt.go`

## R2: Artifact Recommendations from Activity Profile

### Decision

Add an `ArtifactRecommendations` function to
`internal/roles/activities.go` that takes an
`*ActivityProfile` and returns a list of recommended artifact
types with descriptions, derived from the existing
`consts.LayerArtifacts` mapping.

### Rationale

The `consts.LayerArtifacts` map already defines which
artifact types belong to each layer, and the
`ActivityProfile` already contains `ResolvedLayers` with
confidence levels. The missing piece is a function that
combines them and provides user-facing descriptions.

The descriptions should be defined as constants in
`internal/consts/consts.go` to comply with Principle VII
(Centralized Constants).

### Alternatives Considered

1. **Inline the logic in the CLI layer**: Would work but
   violates composability (Principle VIII) and makes the
   logic untestable without UI dependencies.
2. **Add to tutorial path generation**: The learning path
   already maps layers to tutorials; adding artifact
   recommendations would conflate tutorial navigation with
   output identification.

### Implementation Detail

New type `ArtifactRecommendation`:
- `ArtifactType string`
- `SchemaDef string`
- `Description string`
- `Layer int`
- `Confidence Confidence`
- `MCPWizard string` (empty if no wizard available)

New constant map `ArtifactDescriptions` in `consts.go`
mapping each artifact type to a one-sentence description.

New constant map `ArtifactWizards` in `consts.go` mapping
artifact types to MCP wizard names (only ThreatCatalog and
ControlCatalog have wizards).

## R3: Handoff Summary Structure

### Decision

Introduce a `HandoffSummary` struct in a new file
`internal/cli/handoff.go` that captures the transition
context from tutorial completion to MCP server authoring.
Render it using the existing wizard summary pattern
(divider, card with label-value pairs, next-steps with
code block).

### Rationale

The current tutorial completion shows only a checkmark and
title. The wizard launcher collects information that maps
to tutorial context but is currently disconnected. The
handoff summary bridges this gap by surfacing:
- What artifact type to create
- Which MCP prompt or tool to use
- Key decisions from the tutorial
- Schema definition for validation
- MCP server availability status

The wizard summary rendering pattern in `wizard_prompt.go`
(divider, `stepBarStyle` card, `annotationLabelStyle`
labels, `codeBlockStyle` for commands) is the established
visual language for transition points.

### Alternatives Considered

1. **Extend the existing `TutorialPromptResult`**: Would
   mix tutorial navigation concerns with handoff concerns.
   Rejected for composability.
2. **Use the existing `WizardPromptResult`**: Similar fields
   but the wizard result is for MCP wizard input, not
   tutorial completion output. Different lifecycle.
3. **Put in `internal/authoring/`**: The handoff is a CLI
   presentation concern, not an authoring engine concern.
   Rejected.

### Implementation Detail

`HandoffSummary` struct:
- `ArtifactType string`
- `SchemaDef string`
- `MCPPrompt string` (e.g., "threat_assessment" or "")
- `MCPAvailable bool`
- `SchemaVersion string`
- `ExperimentalSchemas []string`
- `VersionMismatch bool` (MCP vs selected version)
- `KeyDecisions []string` (from tutorial content)
- `PreparationChecklist []string`

`RenderHandoffSummary(summary *HandoffSummary, out io.Writer)`
uses the existing styles: `stepBarStyle` for the card,
`annotationLabelStyle` for labels, `codeBlockStyle` for the
MCP prompt command, `RenderWarning` for version mismatches.

## R4: Bypassing Code Without Deleting It

### Decision

Use a conditional bypass in `RunSetup` that skips the
`RunVersionSelection` call. Add a code comment with the
bypass rationale and reference the ADR. Do not introduce
feature flags or build tags.

### Rationale

Go does not have a native feature flag mechanism. Build
tags add complexity and create two code paths that must
both be tested. The simplest approach is to replace the
call site with the auto-selection logic and leave the
`RunVersionSelection` function and `VersionPromptConfig`
struct intact.

The bypass is documented in three places:
1. A code comment at the bypass point in `setup.go`
2. ADR-0003 explaining the decision
3. A note in `version_prompt.go` explaining the function
   is retained for planned re-enablement

### Alternatives Considered

1. **Build tags (`//go:build versionselect`)**: Would
   compile-gate the code but prevents test coverage of the
   bypassed path. Rejected.
2. **Feature flag in config**: Over-engineered for a
   planned future enhancement. Rejected.
3. **Delete the code entirely**: Violates FR-014 which
   requires the code to be preserved for re-enablement.
   Rejected.

## R5: ADR-0003 Version Selection Deferral

### Decision

Create ADR-0003 documenting the intentional deferral of
schema version selection as a user-facing feature.

### Rationale

Constitution Principle VI requires an ADR for every
non-trivial technical decision. Removing a user-facing
feature (even temporarily) is non-trivial because it
changes the setup flow and affects how users understand
schema versioning.

### Content Summary

- **Context**: Version selection adds a decision point that
  creates friction during onboarding. The current Gemara
  schema repository may not yet have distinct Stable vs
  Latest versions that warrant user choice. Pac-Man's focus
  is tutorial guidance, not configuration management.
- **Decision**: Auto-select the latest release. Preserve
  version selection code for future re-enablement.
- **Consequences**: Simpler onboarding; users working with
  experimental schemas won't be warned before selection
  (only after, in the handoff summary); re-enabling requires
  rewiring `RunSetup` to call `RunVersionSelection` again.

## R6: Tutorial Completion Gap Analysis

### Decision

Extend the tutorial completion flow in `tutorial_prompt.go`
to call `RenderHandoffSummary` after marking a tutorial
complete.

### Rationale

The current tutorial completion shows only:
```
✓ Completed: <title>
```
followed by a count:
```
2 of 3 tutorials completed this session
```

There is no content-aware summary and no transition to
authoring. The handoff summary fills this gap by using the
tutorial's layer to determine the artifact type and MCP
prompt, and presenting preparation context.

The handoff summary should be generated from:
1. The completed tutorial's `Layer` field (maps to artifact
   types via `consts.LayerArtifacts`)
2. The session's `SchemaVersion` (set by auto-selection)
3. The session's MCP status (`MCPConnected` or not)
4. The tutorial's section relevance scores (to identify
   key focus areas as "key decisions")

### Alternatives Considered

1. **Show handoff only after all tutorials complete**:
   Users may want to author after a single tutorial.
   Per-tutorial handoff is more flexible. Chosen.
2. **Require explicit "Ready to author?" confirmation**:
   Adds friction. The handoff is informational, not a
   gate. Rejected.
