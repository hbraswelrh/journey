# Implementation Plan: Refocus Pac-Man as Tutorial Guide

**Branch**: `002-tutorial-guide-focus` | **Date**: 2026-03-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-tutorial-guide-focus/spec.md`

## Summary

Refocus Pac-Man from a combined tutorial-and-authoring tool
into a pure terminal-based tutorial guide that helps users
identify their activities, understand their expected artifact
outputs, and walk through relevant tutorials before
transitioning to OpenCode with the gemara-mcp server for
assisted authoring. The schema version selection prompt is
removed from the user flow (auto-selects latest), the guided
authoring engine is bypassed in favor of an OpenCode handoff,
and a new Handoff Summary entity is introduced to bridge the
learn-to-author transition. All terminal output must be
user-friendly and visually polished for all audiences. The
`--doctor` command remains fully functional and unchanged.

## Technical Context

**Language/Version**: Go 1.26.1, formatted with `goimports`
**Primary Dependencies**: `charm.land/huh/v2` (TUI prompts),
`charm.land/lipgloss/v2` (styling),
`github.com/charmbracelet/glamour` (markdown rendering),
`gopkg.in/yaml.v3` (YAML marshaling)
**Storage**: File-based caching (`~/.config/pacman/` for
releases, roles, blocks; `~/.local/share/pacman/` for MCP
install metadata)
**Testing**: `go test -race -count=1 ./...` via `make test`
**Target Platform**: Linux and macOS (POSIX-compatible)
**Project Type**: CLI tool used within OpenCode sessions
**Performance Goals**: Setup flow completes within 5 minutes
including role identification and learning path generation
(SC-001)
**Constraints**: Must not replicate MCP server authoring
wizards (FR-010); version selection code must be preserved
but bypassed (FR-014); `--doctor` command unchanged
(FR-017); terminal output must be sleek and accessible for
all audiences (FR-018); handoff must direct to OpenCode +
gemara-mcp specifically (FR-019)
**Scale/Scope**: 7 predefined roles, 5 Gemara layers with
tutorials (L1-L5), 6 artifact types, 34+ lexicon terms

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after
Phase 1 design.*

| # | Principle | Status | Notes |
|---|-----------|--------|-------|
| I | Schema Conformance | PASS | No artifact generation in Pac-Man; MCP server handles validation. Handoff summary references schema defs by name (from `consts`). |
| II | Gemara Layer Fidelity | PASS | Layer mapping unchanged. Activity-to-layer resolution uses existing `roles.ResolveLayerMappings`. Tutorial sections teach layers in isolation per existing path generation. |
| III | Test-Driven Development | GATE | All new functions must have failing tests before implementation. New entities (HandoffSummary, ReleaseResolution) must have positive and negative test fixtures. |
| IV | Tutorial-First Design | PASS | This feature elevates tutorials as the primary user experience. Handoff summary is the bridge to MCP authoring. |
| V | Incremental Delivery | PASS | Each user story is independently testable and deliverable per spec design. US1 (activity identification) works without US2-US5. |
| VI | Decision Documentation | GATE | Version switching deferral requires an ADR (ADR-0003). |
| VII | Centralized Constants | PASS | All new constants (handoff prompt names, bypass markers) go in `internal/consts/consts.go`. |
| VIII | Composability | PASS | No new CLI subcommands introduced; changes are within existing setup flow. |
| IX | Convention Over Configuration | PASS | Core change: removing a configuration decision (version selection) in favor of a default (latest). Directly aligned. |

**Gate Violations**: None. Two gates (III, VI) are process
requirements that apply during implementation, not design
blockers.

## Project Structure

### Documentation (this feature)

```text
specs/002-tutorial-guide-focus/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── pacman/
    └── main.go              # Remove version-switch menu item

internal/
├── consts/
│   └── consts.go            # Add handoff constants
├── cli/
│   ├── setup.go             # Bypass version prompt, wire auto-select
│   ├── version_prompt.go    # Add bypass flag, preserve for future
│   ├── tutorial_prompt.go   # Add post-tutorial handoff summary
│   ├── author_prompt.go     # Add skip-to-handoff path
│   ├── wizard_prompt.go     # No changes (MCP domain)
│   ├── doctor.go            # No changes (FR-017)
│   ├── styles.go            # UX polish for all audiences (FR-018)
│   └── handoff.go           # NEW: OpenCode handoff summary (FR-019)
├── schema/
│   ├── selector.go          # Add AutoSelectLatest function
│   ├── releases.go          # No changes
│   └── cache.go             # No changes
├── session/
│   └── session.go           # Add handoff state tracking
├── roles/
│   ├── activities.go        # Add ArtifactRecommendations fn
│   └── roles.go             # No changes
├── tutorials/
│   ├── path.go              # Add handoff metadata to PathStep
│   └── loader.go            # No changes
├── authoring/
│   └── engine.go            # No changes (bypassed, not deleted)
└── mcp/
    └── version.go           # No changes

docs/
└── adrs/
    └── ADR-0003-version-selection-deferral.md  # NEW
```

**Structure Decision**: This feature modifies the existing
single-project structure. No new top-level directories are
introduced. The primary change is adding `handoff.go` to
`internal/cli/` and a new function to `internal/roles/`.
The authoring engine (`internal/authoring/`) is left intact
but bypassed in the setup flow.
