# Implementation Plan: US2 — Schema Version Selection (P2)

**Branch**: `001-role-based-tutorial-engine` | **Date**: 2026-03-12
**Spec**: [spec.md](spec.md) — User Story 2
**Depends on**: US1 (MCP Server Setup) — completed

## Summary

US2 enables users to select which Gemara schema version they
work against for the entire session. The system fetches
available tagged releases from the upstream Gemara repository,
determines which schemas are Stable versus Experimental at each
version, and prompts the user to choose "Stable" or "Latest."
The selection governs all validation, content alignment, and
guided authoring. Version data is cached locally for offline
use.

This plan covers FR-016, FR-017, FR-018, FR-019, FR-020, and
integrates with FR-031/FR-032 (version compatibility, already
implemented in `internal/mcp/version.go`).

## Technical Context

**Language/Version**: Go 1.26.1
**Dependencies**: GitHub API (releases endpoint), existing
`internal/mcp/` (version compatibility), `internal/session/`
(schema version field), `internal/consts/`
**Storage**: Local filesystem cache (JSON) with timestamp
**Testing**: `go test ./...` via `make test`

## Constitution Check

| Principle | Status | Notes |
|:---|:---|:---|
| I. Schema Conformance | Pass | Version selection governs all validation |
| II. Gemara Layer Fidelity | Pass | Status attributes per layer reported to user |
| III. TDD | Pass | Tests before implementation |
| VII. Centralized Constants | Pass | Gemara repo URL, cache paths in consts |
| IX. Convention Over Configuration | Pass | Defaults to cached/previous selection |

## Source Code

```text
internal/
├── schema/
│   ├── releases.go         # Fetch releases from GitHub API,
│   │                       #   parse status attributes
│   ├── releases_test.go
│   ├── cache.go            # Local cache read/write with
│   │                       #   timestamp
│   ├── cache_test.go
│   ├── selector.go         # Version selection logic: present
│   │                       #   Stable vs Latest, record choice
│   └── selector_test.go
├── cli/
│   ├── version_prompt.go   # CLI prompt for version selection
│   └── version_prompt_test.go
├── session/
│   └── session.go          # Update: set SchemaVersion from
│                           #   selection
└── consts/
    └── consts.go           # Update: add GitHub API constants,
                            #   cache file names
```

## Implementation Phases

### Phase 1: Release Fetching (FR-016)

- `internal/schema/releases.go`: Query the GitHub releases API
  for `gemaraproj/gemara`, parse tagged releases, extract
  version tags.
- Determine which schemas at each version are marked
  `@status(Stable)` vs `@status(Experimental)` by inspecting
  CUE source files at each tag (or from cached metadata).
- Return a structured list of releases with status maps.
- Tests: API returns releases, API returns empty, API
  unreachable.

### Phase 2: Local Cache (FR-018)

- `internal/schema/cache.go`: Cache version data as JSON with
  a `last_fetched` timestamp. On launch, attempt to refresh
  from upstream; if unreachable, use cache and inform user.
- Tests: Cache write/read, stale cache with upstream available,
  offline with cache, offline without cache.

### Phase 3: Version Selection (FR-017, FR-019)

- `internal/schema/selector.go`: Given a list of releases,
  determine the "Stable" version (most recent where core
  schemas are Stable) and the "Latest" version (most recent
  tag). Present both with version numbers and schema status
  info.
- Integrate with `internal/mcp/version.go` when the user
  selects "Latest" and the MCP server is installed (FR-031).
- Record selection in session state.
- Tests: Stable selection, Latest selection, Latest with MCP
  mismatch warning, version switch mid-session.

### Phase 4: CLI Integration and Polish (FR-020)

- `internal/cli/version_prompt.go`: Wire the version selection
  into the CLI flow (after MCP setup, before role discovery).
  Notify user when a newer version is available upstream.
- Update `internal/cli/setup.go` to call version selection
  after MCP setup completes.
- Integration tests: full flow from MCP setup through version
  selection.

## Dependencies

```text
Phase 1 (Release Fetching)
    │
    ▼
Phase 2 (Local Cache)
    │
    ▼
Phase 3 (Version Selection)
    │
    ▼
Phase 4 (CLI Integration)
```

Sequential — each phase builds on the previous.
