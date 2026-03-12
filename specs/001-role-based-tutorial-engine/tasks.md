# Tasks: Role-Based Tutorial Engine — US1 (P1)

**Input**: Design documents from
`/specs/001-role-based-tutorial-engine/`
**Prerequisites**: plan.md (required), spec.md (required)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no
  dependencies)
- **[Story]**: US1 for all tasks in this file
- Exact file paths included in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project structure, build tooling, and OpenCode
configuration

- [ ] T001 [US1] Restructure project: move entry point to
  `cmd/pacman/main.go`, create `internal/` package directories
  (`consts/`, `mcp/`, `fallback/`, `session/`, `cli/`)
- [ ] T002 [US1] Create `Makefile` with targets: `build`, `test`,
  `lint`, `fmt`, `schema-check`, `clean`
- [ ] T003 [P] [US1] Create `internal/consts/consts.go` with
  centralized constants: MCP server URLs, tool names
  (`get_lexicon`, `validate_gemara_artifact`, `get_schema_docs`),
  default schema version, default tutorials directory path,
  Gemara repository URL
- [ ] T004 [P] [US1] Create `.golangci.yml` with lint rules per
  constitution coding standards
- [ ] T005 [P] [US1] Create `.pre-commit-config.yaml` with hooks
  for `gofmt`, `goimports`, Gitleaks, DCO sign-off verification
- [ ] T006 [P] [US1] Add SPDX license headers
  (`// SPDX-License-Identifier: Apache-2.0`) to all existing
  and new source files
- [ ] T007 [P] [US1] Create OpenCode project configuration:
  `AGENTS.md` at repo root encoding project context, and
  `.opencode/rules/` with constitution-derived rules

**Checkpoint**: `make build` and `make lint` pass with zero
errors. Project structure matches plan.

---

## Phase 2: MCP Detection and Client (FR-028, FR-030)

**Purpose**: Detect and connect to an existing Gemara MCP server

### Tests for Phase 2

> **Write these tests FIRST, ensure they FAIL before
> implementation**

- [ ] T008 [P] [US1] Write test
  `internal/mcp/detect_test.go`: MCP binary found in PATH
  returns `(detected, method=binary)`
- [ ] T009 [P] [US1] Write test
  `internal/mcp/detect_test.go`: Docker container
  `gemara-mcp` running returns `(detected, method=docker)`
- [ ] T010 [P] [US1] Write test
  `internal/mcp/detect_test.go`: Neither binary nor Docker
  found returns `(not detected)`
- [ ] T011 [P] [US1] Write test
  `internal/mcp/client_test.go`: Health check succeeds when
  server is responsive
- [ ] T012 [P] [US1] Write test
  `internal/mcp/client_test.go`: Health check fails with
  timeout when server is unresponsive
- [ ] T013 [P] [US1] Write test
  `internal/mcp/client_test.go`: Mid-session disconnection
  is detected and reported without panic

### Implementation for Phase 2

- [ ] T014 [US1] Implement `internal/mcp/detect.go`:
  `Detect() (DetectionResult, error)` — check PATH for
  `gemara-mcp` binary, check Docker for running container,
  return detection result with installation method
- [ ] T015 [US1] Implement `internal/mcp/client.go`:
  `NewClient(config) (*Client, error)` — MCP client with
  `Connect()`, `HealthCheck()`, `GetLexicon()`,
  `ValidateArtifact()`, `GetSchemaDocs()`, `Close()` methods.
  Must handle connection lifecycle and detect mid-session
  disconnection

**Checkpoint**: `make test` passes for `internal/mcp/`. MCP
detection correctly identifies binary, Docker, and not-found
states.

---

## Phase 3: MCP Installation Guidance (FR-026, FR-027)

**Purpose**: Guide users through MCP server installation when
not detected

### Tests for Phase 3

- [ ] T016 [P] [US1] Write test
  `internal/mcp/install_test.go`: Binary installation
  generates platform-appropriate instructions for Linux
- [ ] T017 [P] [US1] Write test
  `internal/mcp/install_test.go`: Binary installation
  generates platform-appropriate instructions for macOS
- [ ] T018 [P] [US1] Write test
  `internal/mcp/install_test.go`: Docker installation
  generates correct Docker run configuration
- [ ] T019 [P] [US1] Write test
  `internal/mcp/install_test.go`: Post-installation
  verification confirms server responds to health check
- [ ] T020 [P] [US1] Write test
  `internal/cli/setup_test.go`: First-launch prompt
  explains three MCP tools and offers installation
- [ ] T021 [P] [US1] Write test
  `internal/cli/setup_test.go`: User declines installation;
  system informs of degraded capabilities and continues
- [ ] T022 [P] [US1] Write test
  `internal/cli/setup_test.go`: Previously declined user
  is re-offered installation when requesting enhanced
  capability

### Implementation for Phase 3

- [ ] T023 [US1] Implement `internal/mcp/install.go`:
  `InstallBinary(platform) error` and
  `InstallDocker() error` — generate installation
  instructions, execute or guide user through steps,
  verify installation succeeds via health check
- [ ] T024 [US1] Implement `internal/cli/setup.go`:
  `RunSetup(session) error` — first-launch MCP setup
  flow: explain tools, offer binary/Docker/decline,
  handle each path, record user choice in session state

**Checkpoint**: Full installation flow testable end-to-end.
User can accept (binary or Docker) or decline, with
appropriate follow-up behavior.

---

## Phase 4: Local Fallback (FR-029)

**Purpose**: Ensure Pac-Man functions when MCP server is
unavailable

### Tests for Phase 4

- [ ] T025 [P] [US1] Write test
  `internal/fallback/lexicon_test.go`: Bundled lexicon loads
  successfully and contains expected terms
- [ ] T026 [P] [US1] Write test
  `internal/fallback/lexicon_test.go`: Bundled lexicon data
  is valid YAML conforming to expected structure
- [ ] T027 [P] [US1] Write test
  `internal/fallback/validator_test.go`: Local `cue vet`
  validates a known-good artifact successfully
- [ ] T028 [P] [US1] Write test
  `internal/fallback/validator_test.go`: Local `cue vet`
  rejects a known-bad artifact with actionable error
- [ ] T029 [P] [US1] Write test
  `internal/fallback/schemadocs_test.go`: Cached schema docs
  load successfully when present
- [ ] T030 [P] [US1] Write test
  `internal/fallback/schemadocs_test.go`: Missing cache
  returns informative error, not panic

### Implementation for Phase 4

- [ ] T031 [P] [US1] Implement `internal/fallback/lexicon.go`:
  `LoadBundledLexicon() (*Lexicon, error)` — load embedded
  lexicon YAML, return structured lexicon data
- [ ] T032 [P] [US1] Implement
  `internal/fallback/validator.go`:
  `ValidateLocal(artifact, schemaType, schemaVersion) error`
  — wrap `cue vet -c -d '#<SchemaType>'` invocation
- [ ] T033 [P] [US1] Implement
  `internal/fallback/schemadocs.go`:
  `LoadCachedDocs(version) (*SchemaDocs, error)` — load
  cached schema documentation from local filesystem
- [ ] T034 [US1] Create test fixtures in `testdata/`:
  `lexicon_valid.yaml`, `lexicon_invalid.yaml`,
  `artifact_valid.yaml`, `artifact_invalid.yaml`

**Checkpoint**: All fallback paths functional. System operates
without MCP server using local data and tooling.

---

## Phase 5: Version Compatibility (FR-031, FR-032)

**Purpose**: Detect and warn about gemara-mcp / schema version
mismatches

### Tests for Phase 5

- [ ] T035 [P] [US1] Write test
  `internal/mcp/version_test.go`: gemara-mcp version matches
  selected schema version — no warning
- [ ] T036 [P] [US1] Write test
  `internal/mcp/version_test.go`: gemara-mcp built against
  older schema than user selected — warning with
  recommendations
- [ ] T037 [P] [US1] Write test
  `internal/mcp/version_test.go`: gemara-mcp does not expose
  version metadata — warning that compatibility cannot be
  verified

### Implementation for Phase 5

- [ ] T038 [US1] Implement `internal/mcp/version.go`:
  `CheckCompatibility(client, selectedVersion)
  (*CompatResult, error)` — query MCP server for version
  info, compare against selected schema version, return
  compatibility status with actionable recommendations

**Checkpoint**: Version compatibility checks produce correct
warnings for all three scenarios (match, mismatch, unknown).

---

## Phase 6: Session Management

**Purpose**: Unified session state tracking MCP status and
fallback transitions

### Tests for Phase 6

- [ ] T039 [P] [US1] Write test
  `internal/session/session_test.go`: Session initializes
  with MCP connected — all three tools marked available
- [ ] T040 [P] [US1] Write test
  `internal/session/session_test.go`: Session initializes
  without MCP — fallback mode active, degraded capabilities
  listed
- [ ] T041 [P] [US1] Write test
  `internal/session/session_test.go`: Mid-session MCP
  disconnection transitions to fallback without data loss
- [ ] T042 [P] [US1] Write test
  `internal/session/session_test.go`: MCP reconnection
  after fallback restores full capabilities

### Implementation for Phase 6

- [ ] T043 [US1] Implement `internal/session/session.go`:
  `Session` struct with `MCPStatus`, `SchemaVersion`,
  `FallbackMode`, `AvailableTools` fields.
  `NewSession(mcpClient, fallbacks) *Session` constructor.
  `HandleDisconnection()` and `HandleReconnection()` methods
  for mid-session transitions

**Checkpoint**: Session correctly tracks MCP state across
all lifecycle transitions (connected, disconnected,
reconnected, never connected).

---

## Phase 7: Integration and Polish

**Purpose**: End-to-end validation and user-facing quality

- [ ] T044 [US1] Integration test: launch Pac-Man with MCP
  server available — verify detection, connection, all three
  tools respond, session shows connected status
- [ ] T045 [US1] Integration test: launch Pac-Man without MCP
  server — verify detection returns not-found, setup prompt
  appears, declining proceeds with fallback mode
- [ ] T046 [US1] Integration test: MCP server disconnects
  mid-session — verify fallback activates, user notified,
  in-progress work preserved
- [ ] T047 [US1] Verify all CLI help text and error messages
  use Gemara lexicon terms consistently (FR-011)
- [ ] T048 [US1] Verify `make build`, `make test`, `make lint`
  pass with zero errors and zero warnings
- [ ] T049 [US1] Update `README.md` to reference OpenCode as
  preferred harness and document MCP server setup

**Checkpoint**: US1 is fully functional and independently
testable. A user can launch Pac-Man, be guided through MCP
setup (or decline), and proceed with appropriate capabilities.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Detection & Client)**: Depends on Phase 1
  (project structure and constants)
- **Phase 3 (Installation)**: Depends on Phase 2 (needs MCP
  client for post-install verification)
- **Phase 4 (Fallback)**: Depends on Phase 1; can run in
  **parallel** with Phase 3
- **Phase 5 (Version Compat)**: Depends on Phase 2 (MCP
  client)
- **Phase 6 (Session)**: Depends on Phases 2, 4, and 5 (needs
  MCP client, fallback services, and version checks)
- **Phase 7 (Integration)**: Depends on all preceding phases

### Parallel Opportunities

```text
Phase 1 (Setup)
    │
    ▼
Phase 2 (Detection & Client)
    │
    ├────────────────────┐
    ▼                    ▼
Phase 3 (Install)    Phase 4 (Fallback)  ← parallel
    │                    │
    ├────────┬───────────┘
    ▼        │
Phase 5      │
    │        │
    └────────┘
         │
         ▼
    Phase 6 (Session)
         │
         ▼
    Phase 7 (Integration)
```

Within each phase, all tasks marked `[P]` can run in parallel.
All test tasks within a phase can run in parallel.

---

## Notes

- All test tasks follow TDD: write test, confirm it fails,
  then implement.
- Each task produces files with SPDX headers and passes
  `make lint`.
- Commit after each task or logical group per Conventional
  Commits format.
- US1 is independently deliverable — no dependency on US2-US6.
