# Implementation Plan: Role-Based Tutorial Engine — US1 (P1)

**Branch**: `001-role-based-tutorial-engine` | **Date**: 2026-03-12
**Spec**: [spec.md](spec.md)
**Input**: Feature specification from
`/specs/001-role-based-tutorial-engine/spec.md`, User Story 1

## Summary

User Story 1 (Gemara MCP Server Setup) is the highest-priority
deliverable. When a user launches Pac-Man through OpenCode, the
system detects whether the Gemara MCP server is installed,
offers installation via binary or Docker if it is not, and
configures the MCP client connection. If the user declines, the
system falls back to local CUE tooling and bundled lexicon data.
The MCP server provides three tools (`get_lexicon`,
`validate_gemara_artifact`, `get_schema_docs`) that enhance
every subsequent feature.

This plan covers the implementation of US1 acceptance scenarios
1-6 and the functional requirements that directly support them:
FR-026, FR-027, FR-028, FR-029, FR-030, FR-031, FR-032, FR-033,
FR-034.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: CUE Go SDK (`cuelang.org/go`),
Gemara schema repository (`github.com/gemaraproj/gemara`),
Gemara Go SDK (`github.com/gemaraproj/go-gemara`), Gemara MCP
server (`github.com/gemaraproj/gemara-mcp`)
**Storage**: Local filesystem (cached lexicon, schema docs,
version info)
**Testing**: `go test ./...` via `make test`; positive and
negative fixtures per Constitution Principle III
**Target Platform**: Linux and macOS (CLI)
**Project Type**: CLI tool
**Constraints**: Must function fully offline with degraded
capabilities when MCP server is unavailable; must detect
mid-session MCP disconnection without data loss
**AI Harness**: OpenCode (`https://opencode.ai`) per ADR-0001

## Constitution Check

| Principle | Status | Notes |
|:---|:---|:---|
| I. Schema Conformance | Pass | MCP server provides `validate_gemara_artifact`; local `cue vet` as fallback |
| II. Gemara Layer Fidelity | Pass | US1 is infrastructure; does not produce layer-specific artifacts |
| III. Test-Driven Development | Pass | Tests written before implementation; positive/negative fixtures for MCP detection, installation, fallback |
| IV. Tutorial-First Design | Pass | OpenCode guides users through setup; tutorial content deferred to US3 |
| V. Incremental Delivery | Pass | US1 is independently usable — user gets MCP setup or confirmed fallback |
| VI. Decision Documentation | Pass | ADR-0001 created for OpenCode adoption |
| VII. Centralized Constants | Pass | MCP server URLs, tool names, version strings defined in `internal/consts/` |
| VIII. Composability | Pass | MCP detection and setup are independent subcommands; output to stdout, errors to stderr |
| IX. Convention Over Configuration | Pass | Default to auto-detection; zero-config for the common case (MCP already installed) |

## Project Structure

### Documentation (this feature)

```text
specs/001-role-based-tutorial-engine/
├── spec.md
├── plan.md              # This file
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 output (next step)

docs/adrs/
└── ADR-0001-opencode-as-ai-harness.md
```

### Source Code (repository root)

```text
internal/
├── consts/
│   └── consts.go            # Centralized constants (MCP URLs,
│                            #   tool names, schema types)
├── mcp/
│   ├── client.go            # MCP client: connect, health
│   │                        #   check, tool invocation
│   ├── client_test.go       # Positive/negative MCP client
│   │                        #   tests
│   ├── detect.go            # Auto-detection: is MCP server
│   │                        #   installed and running?
│   ├── detect_test.go       # Detection tests (binary found,
│   │                        #   Docker running, neither)
│   ├── install.go           # Installation guidance: binary
│   │                        #   and Docker flows
│   ├── install_test.go      # Installation flow tests
│   ├── version.go           # Version compatibility checks
│   │                        #   (gemara-mcp vs schema version)
│   └── version_test.go      # Compatibility check tests
├── fallback/
│   ├── lexicon.go           # Bundled lexicon data for offline
│   │                        #   use
│   ├── lexicon_test.go
│   ├── validator.go         # Local CUE validation wrapper
│   ├── validator_test.go
│   ├── schemadocs.go        # Cached schema documentation
│   └── schemadocs_test.go
├── session/
│   ├── session.go           # Session state: MCP connection
│   │                        #   status, selected schema
│   │                        #   version, fallback mode
│   └── session_test.go
└── cli/
    ├── setup.go             # CLI entry point for MCP setup
    │                        #   flow (first-launch prompt)
    ├── setup_test.go
    └── root.go              # Root command (existing or new)

cmd/
└── pacman/
    └── main.go              # Entry point (replaces or wraps
                             #   current main.go)

testdata/
├── lexicon_valid.yaml       # Positive fixture: valid lexicon
├── lexicon_invalid.yaml     # Negative fixture: invalid
│                            #   lexicon
└── mcp_version_response.json # Mock MCP version response
```

**Structure Decision**: Go standard project layout with
`internal/` for private packages and `cmd/` for the binary
entry point. The existing `main.go` at root will be migrated
to `cmd/pacman/main.go` per Go conventions. The `pacman/`
directory (existing) will be evaluated for reuse or migration
into `internal/`.

## US1 Implementation Phases

### Phase 1: Setup (Shared Infrastructure)

- Makefile with `build`, `test`, `lint`, `schema-check` targets
- Project restructure: `cmd/pacman/main.go`, `internal/`
  packages
- `internal/consts/consts.go` with centralized MCP constants
- Pre-commit hook configuration (`.pre-commit-config.yaml`)
- `.golangci.yml` lint configuration
- SPDX license headers on all source files
- OpenCode configuration (`.opencode/rules/`, `AGENTS.md`)

### Phase 2: MCP Detection and Client (FR-030, FR-028)

- `internal/mcp/detect.go`: Detect whether gemara-mcp binary
  is in PATH or a Docker container named `gemara-mcp` is
  running. Return detection result with method (binary/Docker/
  not found).
- `internal/mcp/client.go`: MCP client that connects to the
  detected server, performs health checks, and invokes the
  three tools (`get_lexicon`, `validate_gemara_artifact`,
  `get_schema_docs`). Must handle connection timeouts and
  mid-session disconnection gracefully.
- Tests: MCP server found via binary, found via Docker, not
  found, found but unresponsive, disconnects mid-session.

### Phase 3: MCP Installation Guidance (FR-026, FR-027)

- `internal/mcp/install.go`: Provide platform-appropriate
  installation instructions for binary (Linux/macOS) and
  Docker. Guide the user through installation steps. After
  installation, verify the server is accessible and responds
  to a health check.
- `internal/cli/setup.go`: First-launch prompt that explains
  the three MCP tools, offers installation, and handles
  accept/decline flow. If declined, inform user of degraded
  capabilities.
- Tests: User accepts binary install, user accepts Docker
  install, user declines, re-offer on subsequent capability
  request.

### Phase 4: Local Fallback (FR-029)

- `internal/fallback/lexicon.go`: Load bundled lexicon data
  from embedded YAML when MCP server is unavailable.
- `internal/fallback/validator.go`: Wrap local `cue vet`
  invocation for schema validation when MCP
  `validate_gemara_artifact` is unavailable.
- `internal/fallback/schemadocs.go`: Serve cached schema
  documentation when MCP `get_schema_docs` is unavailable.
- Tests: Fallback activates when MCP not installed, fallback
  activates when MCP disconnects mid-session, bundled data
  is valid.

### Phase 5: Version Compatibility (FR-031, FR-032)

- `internal/mcp/version.go`: Query installed gemara-mcp for
  its version and the Gemara schema version it was built
  against. Compare with the user's selected schema version.
  Warn on mismatch with actionable recommendations.
- Tests: Versions match, versions mismatch, server does not
  expose version metadata.

### Phase 6: Session Management

- `internal/session/session.go`: Session state object that
  tracks MCP connection status, selected schema version,
  fallback mode, and available tools. Handles transitions
  (MCP connected -> disconnected -> reconnected) without
  losing in-progress work.
- Tests: Session initialization with MCP, session
  initialization without MCP, mid-session fallback transition.

### Phase 7: Integration and Polish

- End-to-end integration test: launch Pac-Man, detect MCP
  status, offer installation, configure session, confirm
  tools are available.
- CLI help text and error messages using Gemara lexicon terms.
- OpenCode rules encoding the setup flow for guided
  onboarding.

## Dependencies & Execution Order

```text
Phase 1 (Setup)
    │
    ▼
Phase 2 (MCP Detection & Client)
    │
    ├──────────────────┐
    ▼                  ▼
Phase 3 (Install)   Phase 4 (Fallback)
    │                  │
    └──────┬───────────┘
           ▼
    Phase 5 (Version Compat)
           │
           ▼
    Phase 6 (Session Mgmt)
           │
           ▼
    Phase 7 (Integration)
```

- Phases 3 and 4 can proceed in parallel after Phase 2.
- Phase 5 depends on both the MCP client (Phase 2) and the
  session concept (feeds into Phase 6).
- Phase 6 integrates all preceding work.
- Phase 7 is the final validation pass.

## Complexity Tracking

No constitution violations identified. No complexity
justifications required.
