# Research: Refocus Gemara User Journey as Tutorial Guide

**Branch**: `002-tutorial-guide-focus`
**Date**: 2026-03-25
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
  Latest versions that warrant user choice. Gemara User Journey's focus
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

## R7: README Restructuring Strategy

### Decision

Rewrite `README.md` as a concise landing page. Move
detailed content to dedicated files under `docs/` and
link from the README.

### Rationale

The current README is 501 lines covering internal
implementation details (project structure, contributing
rules, MCP update procedures, layer reference, content
blocks, collaboration views, guided authoring). This
obscures the user journey and makes the README ineffective
as a GitHub landing page.

The "inverted pyramid" pattern (summary, visual, quick
start, links to details) is the established best practice
for open source project READMEs. Users need to understand
what the project does, what it looks like, and how to start
within a single scroll.

### Content Displacement Plan

| Current README Section (lines) | Destination | Action |
|-------------------------------|-------------|--------|
| Title + intro (1-17) | README | Rewrite: position as tutorial guide, distinguish from MCP server |
| Problem Statement (19-31) | README | Keep but shorten to 2-3 sentences |
| Capabilities (33-140) | README | Replace with User Journey section (3-step narrative) |
| Gemara Layer Reference (142-161) | `docs/layer-reference.md` | Move with link |
| Prerequisites (163-214) | README | Keep but convert to hyperlinked list (remove inline install commands) |
| Getting Started (216-302) | README | Keep but reduce to 4 steps max |
| Using MCP Wizard Prompts (304-327) | Remove | Already in AGENTS.md |
| Keeping gemara-mcp Up to Date (329-388) | `docs/mcp-update-guide.md` | Move with link |
| First Launch (390-405) | Remove | Outdated (references version prompt) |
| Project Structure (407-452) | `docs/project-structure.md` | Move with link |
| Contributing (454-473) | README | Reduce to one line linking `CONTRIBUTING.md` |
| Upstream Projects (475-488) | README | Keep as compact table |
| ADRs (490-496) | Remove | Move to `docs/project-structure.md` |
| License (498-501) | README | Keep as one-liner |

### Proposed README Structure

1. **Title** — `# Gemara User Journey`
2. **One-paragraph summary** — Role-based tutorial guide for
   Gemara GRC schemas. Distinguish from MCP server.
3. **Web UI screenshot** — `![Gemara User Journey Web UI](docs/images/web-ui-preview.png)`
4. **User Journey** — 3-step narrative:
   - Discover: Role and activity identification
   - Learn: Tailored tutorial walkthrough
   - Author: Handoff to OpenCode with gemara-mcp
5. **Prerequisites** — Hyperlinked dependency list
6. **Getting Started** — 4 steps: clone, build, verify, launch
7. **Upstream Projects** — Compact table
8. **Learn More** — Links to `docs/` files
9. **License** — One line

### Target: ~120-150 lines (down from 501)

### Alternatives Considered

1. **Keep all sections, shorten each**: Still too long;
   the problem is section count, not per-section verbosity.
   Rejected.
2. **Landing page only, delete detailed content**: Loses
   valuable contributor documentation. Rejected.
3. **Use `<details>` collapsible sections**: Adds HTML
   noise, renders inconsistently on GitHub mobile.
   Rejected.

## R8: Web UI Screenshot Selection

### Decision

Capture the Results view showing resolved layers and
artifact recommendations.

### Rationale

The Results view is the most visually distinctive screen
in the web UI — it shows the active layer map with
confidence highlighting, artifact recommendations with
checklists, and layer flows. This communicates the
project's value proposition more effectively than other
views:

- Role Selection: too generic (card grid)
- Activity Probe: shows input, not output
- Tutorial Suggestions: useful but less visually distinctive

### Storage

- File: `docs/images/web-ui-preview.png`
- Reference in README: `![Gemara User Journey Web UI](docs/images/web-ui-preview.png)`
- Committed to repository per clarification decision

## R9: Dependency Installation Links

### Decision

Use official installation page URLs for all dependencies.

### Rationale

Direct links to official installation pages ensure users
get current instructions and reduce Gemara User Journey's maintenance
burden when upstream installers change. The constitution
requires Homebrew as the preferred installation method but
also requires alternative methods to be documented.

### Links

| Dependency | URL | Version |
|------------|-----|---------|
| Go | https://go.dev/dl/ | 1.21+ |
| CUE | https://cuelang.org/docs/introduction/installation/ | v0.15.1+ |
| OpenCode | https://opencode.ai | latest |
| Git | https://git-scm.com/downloads | latest |
| gemara-mcp | https://github.com/gemaraproj/gemara-mcp | build from source |

The README will hyperlink each dependency name to its
installation page rather than inlining `brew install`
and `wget` commands. Platform-specific install commands
move to `docs/` or are omitted in favor of the official
pages.

## R10: AGENTS.md Minimal Updates

### Decision

Make minimal text updates to AGENTS.md to ensure
consistency with the spec's positioning.

### Rationale

AGENTS.md already reflects the "tutorial guide" positioning
with "Two Paths: Learn or Author." Updates needed:
- Ensure "Recent Changes" accurately reflects 002 changes
- Verify "Active Technologies" is current
- No structural changes required

### Impact

Text-only edits; no architectural changes.
