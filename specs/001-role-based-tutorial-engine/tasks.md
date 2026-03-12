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
  `internal/mcp/detect_test.go`: Podman container
  `gemara-mcp` running returns `(detected, method=podman)`
- [ ] T010 [P] [US1] Write test
  `internal/mcp/detect_test.go`: Neither binary nor Podman
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
  `gemara-mcp` binary, check Podman for running container,
  return detection result with installation method
- [ ] T015 [US1] Implement `internal/mcp/client.go`:
  `NewClient(config) (*Client, error)` — MCP client with
  `Connect()`, `HealthCheck()`, `GetLexicon()`,
  `ValidateArtifact()`, `GetSchemaDocs()`, `Close()` methods.
  Must handle connection lifecycle and detect mid-session
  disconnection

**Checkpoint**: `make test` passes for `internal/mcp/`. MCP
detection correctly identifies binary, Podman, and not-found
states.

---

## Phase 3: MCP Automated Installation (FR-026, FR-027)

**Purpose**: Automate gemara-mcp installation (clone, build,
configure) when not detected

### Tests for Phase 3

- [ ] T016 [P] [US1] Write test
  `internal/mcp/install_test.go`: Resolve latest release
  and retrieve SHA256 commit digest from upstream
  gemara-mcp repository
- [ ] T017 [P] [US1] Write test
  `internal/mcp/install_test.go`: Clone via SSH succeeds
  and checks out correct commit by SHA256 digest
- [ ] T018 [P] [US1] Write test
  `internal/mcp/install_test.go`: Clone via HTTPS succeeds
  and checks out correct commit by SHA256 digest
- [ ] T019 [P] [US1] Write test
  `internal/mcp/install_test.go`: `make build` produces
  expected binary in cloned directory
- [ ] T020 [P] [US1] Write test
  `internal/mcp/install_test.go`: Podman installation
  generates correct Podman run configuration
- [ ] T021 [P] [US1] Write test
  `internal/mcp/install_test.go`: Post-installation
  verification confirms server responds to health check
- [ ] T022 [P] [US1] Write test
  `internal/mcp/config_test.go`: Writing new `opencode.json`
  creates valid config with gemara-mcp local server entry
  and correct binary path in command array
- [ ] T023 [P] [US1] Write test
  `internal/mcp/config_test.go`: Updating existing
  `opencode.json` preserves other MCP entries and adds
  gemara-mcp entry
- [ ] T024 [P] [US1] Write test
  `internal/cli/setup_test.go`: First-launch prompt
  explains three MCP tools and offers installation
- [ ] T025 [P] [US1] Write test
  `internal/cli/setup_test.go`: User declines installation;
  system informs of degraded capabilities and continues
- [ ] T026 [P] [US1] Write test
  `internal/cli/setup_test.go`: Previously declined user
  is re-offered installation when requesting enhanced
  capability

### Implementation for Phase 3

- [ ] T027 [US1] Implement `internal/mcp/install.go`:
  `ResolveLatestRelease() (ReleaseInfo, error)` — query
  upstream gemara-mcp repository for latest release and
  retrieve the SHA256 commit digest.
  `CloneAndBuild(cloneMethod, digest, destDir) (string,
  error)` — clone repo via SSH or HTTPS, check out the
  pinned commit by SHA256 digest (not mutable tag), run
  `make build`, return path to built binary.
  `InstallPodman() error` — Podman alternative.
- [ ] T028 [US1] Implement `internal/mcp/config.go`:
  `ReadOpenCodeConfig(path) (*OpenCodeConfig, error)` and
  `WriteOpenCodeConfig(path, config) error` — read/write
  `opencode.json`. `EnsureMCPEntry(config, binaryPath)
  *OpenCodeConfig` — add or update the gemara-mcp local
  MCP server entry with the built binary path in the
  command array (e.g., `["path/to/gemara-mcp"]`)
- [ ] T029 [US1] Implement `internal/cli/setup.go`:
  `RunSetup(session) error` — first-launch setup flow:
  verify required tools (CUE, Gitleaks) are installed
  with Homebrew as the preferred installation method
  (FR-035), explain MCP tools, offer automated source
  build (SSH/HTTPS) or Podman or decline for gemara-mcp,
  execute chosen installation method, configure
  `opencode.json`, record user choices in session state

**Checkpoint**: Full automated installation flow testable
end-to-end. User can accept source build (SSH or HTTPS) or
Podman or decline, with `opencode.json` correctly configured.

---

## Phase 4: Local Fallback (FR-029)

**Purpose**: Ensure Pac-Man functions when MCP server is
unavailable

### Tests for Phase 4

- [ ] T030 [P] [US1] Write test
  `internal/fallback/lexicon_test.go`: Bundled lexicon loads
  successfully and contains expected terms
- [ ] T031 [P] [US1] Write test
  `internal/fallback/lexicon_test.go`: Bundled lexicon data
  is valid YAML conforming to expected structure
- [ ] T032 [P] [US1] Write test
  `internal/fallback/validator_test.go`: Local `cue vet`
  validates a known-good artifact successfully
- [ ] T033 [P] [US1] Write test
  `internal/fallback/validator_test.go`: Local `cue vet`
  rejects a known-bad artifact with actionable error
- [ ] T034 [P] [US1] Write test
  `internal/fallback/schemadocs_test.go`: Cached schema docs
  load successfully when present
- [ ] T035 [P] [US1] Write test
  `internal/fallback/schemadocs_test.go`: Missing cache
  returns informative error, not panic

### Implementation for Phase 4

- [ ] T036 [P] [US1] Implement `internal/fallback/lexicon.go`:
  `LoadBundledLexicon() (*Lexicon, error)` — load embedded
  lexicon YAML, return structured lexicon data
- [ ] T037 [P] [US1] Implement
  `internal/fallback/validator.go`:
  `ValidateLocal(artifact, schemaType, schemaVersion) error`
  — wrap `cue vet -c -d '#<SchemaType>'` invocation
- [ ] T038 [P] [US1] Implement
  `internal/fallback/schemadocs.go`:
  `LoadCachedDocs(version) (*SchemaDocs, error)` — load
  cached schema documentation from local filesystem
- [ ] T039 [US1] Create test fixtures in `testdata/`:
  `lexicon_valid.yaml`, `lexicon_invalid.yaml`,
  `artifact_valid.yaml`, `artifact_invalid.yaml`

**Checkpoint**: All fallback paths functional. System operates
without MCP server using local data and tooling.

---

## Phase 5: Version Compatibility (FR-031, FR-032)

**Purpose**: Detect and warn about gemara-mcp / schema version
mismatches

### Tests for Phase 5

- [ ] T040 [P] [US1] Write test
  `internal/mcp/version_test.go`: gemara-mcp version matches
  selected schema version — no warning
- [ ] T041 [P] [US1] Write test
  `internal/mcp/version_test.go`: gemara-mcp built against
  older schema than user selected — warning with
  recommendations
- [ ] T042 [P] [US1] Write test
  `internal/mcp/version_test.go`: gemara-mcp does not expose
  version metadata — warning that compatibility cannot be
  verified

### Implementation for Phase 5

- [ ] T043 [US1] Implement `internal/mcp/version.go`:
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

- [ ] T044 [P] [US1] Write test
  `internal/session/session_test.go`: Session initializes
  with MCP connected — all three tools marked available
- [ ] T045 [P] [US1] Write test
  `internal/session/session_test.go`: Session initializes
  without MCP — fallback mode active, degraded capabilities
  listed
- [ ] T046 [P] [US1] Write test
  `internal/session/session_test.go`: Mid-session MCP
  disconnection transitions to fallback without data loss
- [ ] T047 [P] [US1] Write test
  `internal/session/session_test.go`: MCP reconnection
  after fallback restores full capabilities

### Implementation for Phase 6

- [ ] T048 [US1] Implement `internal/session/session.go`:
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

- [ ] T049 [US1] Integration test: launch Pac-Man with MCP
  server available — verify detection, connection, all three
  tools respond, session shows connected status
- [ ] T050 [US1] Integration test: launch Pac-Man without MCP
  server — verify detection returns not-found, setup prompt
  appears, automated install completes, `opencode.json`
  configured, declining proceeds with fallback mode
- [ ] T051 [US1] Integration test: MCP server disconnects
  mid-session — verify fallback activates, user notified,
  in-progress work preserved
- [ ] T052 [US1] Verify all CLI help text and error messages
  use Gemara lexicon terms consistently (FR-011)
- [ ] T053 [US1] Verify `make build`, `make test`, `make lint`
  pass with zero errors and zero warnings
- [ ] T054 [US1] Update `README.md` to reference OpenCode as
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
