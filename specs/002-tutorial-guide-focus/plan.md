# Implementation Plan: Refocus Gemara User Journey as Tutorial Guide

**Branch**: `002-tutorial-guide-focus` | **Date**: 2026-03-25 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-tutorial-guide-focus/spec.md`

## Summary

Refocus Gemara User Journey from a full authoring engine to a role-based tutorial guide that helps users identify their activities, discover relevant Gemara tutorials, walk through them section-by-section, and hand off to OpenCode with the gemara-mcp server for artifact authoring. The implementation auto-selects the latest Gemara schema version (bypassing version selection prompts), retains version switching code for future re-enablement, and rewrites the README as a concise landing page with dependency links, a web UI screenshot, and a user journey narrative. Detailed documentation is moved from the README to dedicated files in `docs/`.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: `charm.land/huh/v2` (TUI forms), `charm.land/lipgloss/v2` (terminal styling), `github.com/charmbracelet/glamour` (markdown rendering), `gopkg.in/yaml.v3` (YAML), React 19 + Vite 8 (web frontend)
**Storage**: File-based caching (`~/.config/journey/`), YAML profiles (`~/.config/journey/roles/`), upstream tutorial clone (`~/.local/share/journey/gemara/`)
**Testing**: Go standard `testing` package with `-race` flag; `go test -race -count=1 ./...`
**Target Platform**: Linux and macOS (Windows explicitly out of scope per constitution)
**Project Type**: CLI tool (Go binary) + Web SPA (React/Vite) + OpenCode integration (AGENTS.md)
**Performance Goals**: N/A — interactive CLI with sub-second response for role matching, tutorial loading, and rendering
**Constraints**: Offline-capable (cached releases, bundled lexicon); all output must pass `cue vet` against Gemara CUE schemas; Makefile is single entry point
**Scale/Scope**: 7 predefined roles, 60+ activity keywords, 7 Gemara layers, 12 artifact types, 4 upstream tutorials

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Schema Conformance | PASS | No schema-producing changes in this feature. README and docs are Markdown, not Gemara artifacts. Existing validation remains intact. |
| II. Gemara Layer Fidelity | PASS | Layer model already correctly implemented in `internal/roles/activities.go` and `internal/tutorials/path.go`. No layer boundary changes. |
| III. Test-Driven Development | PASS | US1-US5 already have tests. US6 is documentation-only (README rewrite); acceptance criteria verified by visual inspection and link testing, not unit tests. |
| IV. Tutorial-First Design | PASS | This feature strengthens tutorial-first design by making the tutorial walkthrough the primary user experience. |
| V. Incremental Delivery | PASS | US6 (README) is independently deliverable. US1-US5 are already implemented and tested. |
| VI. Decision Documentation | PASS | ADR-0003 already documents version selection deferral. No new non-trivial decisions required. |
| VII. Centralized Constants | PASS | All constants remain in `internal/consts/consts.go`. No new magic strings introduced. |
| VIII. Composability | PASS | No changes to CLI subcommand structure. |
| IX. Convention Over Configuration | PASS | Auto-selecting latest release reduces user configuration decisions, directly supporting this principle. |
| Repository Standard Files | PASS | README.md restructuring maintains the file. CONTRIBUTING.md, LICENSE, SECURITY.md, CODE_OF_CONDUCT.md, .github/ all exist. |
| SPDX License Headers | PASS | No new source files created (only Markdown documentation). |
| Makefile as Entry Point | PASS | No Makefile changes needed. |
| Tool Installation (Homebrew) | PASS | README will include Homebrew as preferred installation method per constitution, with alternative methods documented. |

**Gate Result: ALL PASS — proceed to Phase 0.**

## Project Structure

### Documentation (this feature)

```text
specs/002-tutorial-guide-focus/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CLI contract)
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
# Existing structure — no new source directories needed
cmd/journey/
  main.go                  # Entry point (US1-US5 already wired)
internal/
  cli/                     # Setup flows, role prompts, tutorial player, handoff rendering
  consts/                  # Centralized constants
  roles/                   # Role matching, keyword extraction, layer mapping
  tutorials/               # Tutorial loading, path generation, section scoring
  schema/                  # Release fetching, version selection (bypassed), auto-select
  session/                 # In-memory session state
  mcp/                     # MCP server detection, installation, client
  fallback/                # Local alternatives when MCP unavailable
  blocks/                  # Content block extraction, drift detection
  team/                    # Team collaboration, handoff detection
  authoring/               # Guided authoring engine (retained, not in active flow)
web/
  src/
    components/            # React components (RoleSelection, ActivityProbe, Results, etc.)
    generated/             # Auto-generated TypeScript from Go constants
    lib/                   # Client-side role matching logic

# New files for US6 (README restructuring)
docs/
  images/
    web-ui-preview.png     # Manually captured screenshot of web UI (NEW)
  project-structure.md     # Moved from README (NEW)
  layer-reference.md       # Moved from README (NEW)
  mcp-update-guide.md      # Moved from README (NEW)
  adrs/                    # Existing ADRs (unchanged)
  tutorials/               # Existing tailored tutorials (unchanged)
README.md                  # Rewritten as concise landing page (MODIFIED)
```

**Structure Decision**: The existing Go project structure is retained in full. US1-US5 require no structural changes — all code is already implemented. US6 adds documentation files under `docs/` and rewrites `README.md`. No new Go source files, packages, or directories are created.

## Constitution Check — Post-Design Re-evaluation

*All principles re-evaluated after Phase 1 design completion.*

| Principle | Pre-Design | Post-Design | Delta |
|-----------|-----------|-------------|-------|
| I. Schema Conformance | PASS | PASS | No change — no schema-producing work |
| II. Gemara Layer Fidelity | PASS | PASS | Layer reference content preserved in `docs/` |
| III. Test-Driven Development | PASS | PASS | US6 verified by visual inspection (SC-010/SC-011) |
| IV. Tutorial-First Design | PASS | PASS | README strengthens tutorial-first narrative |
| V. Incremental Delivery | PASS | PASS | US6 independently deliverable |
| VI. Decision Documentation | PASS | PASS | R7 documents README restructuring rationale |
| VII. Centralized Constants | PASS | PASS | No new constants for README |
| VIII. Composability | PASS | PASS | No CLI changes |
| IX. Convention Over Configuration | PASS | PASS | Hyperlinked deps reduce config burden |
| Repository Standard Files | PASS | PASS | All required files preserved |
| Tool Installation (Homebrew) | PASS | PASS | Official pages include Homebrew; detailed commands move to `docs/` |

**Post-design gate: ALL PASS. No new violations introduced.**

## Complexity Tracking

> No Constitution Check violations. No complexity justifications needed.
