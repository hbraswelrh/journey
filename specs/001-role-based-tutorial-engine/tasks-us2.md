# Tasks: US2 — Schema Version Selection (P2)

**Input**: `plan-us2.md`, `spec.md` (User Story 2)
**Prerequisites**: US1 completed (MCP detection, client,
version compatibility, session management)

---

## Phase 1: Release Fetching (FR-016)

### Tests

- [x] T101 [P] [US2] Write test
  `internal/schema/releases_test.go`: Fetch releases returns
  parsed list of versions with tags and dates
- [x] T102 [P] [US2] Write test
  `internal/schema/releases_test.go`: Fetch releases handles
  empty release list gracefully
- [x] T103 [P] [US2] Write test
  `internal/schema/releases_test.go`: Fetch releases handles
  API error (network unreachable) gracefully
- [x] T104 [P] [US2] Write test
  `internal/schema/releases_test.go`: Parse schema status
  correctly identifies Stable vs Experimental schemas at a
  given version

### Implementation

- [x] T105 [US2] Add GitHub API release list URL constant to
  `internal/consts/consts.go`
- [x] T106 [US2] Implement `internal/schema/releases.go`:
  `FetchReleases(ctx) ([]Release, error)` — query GitHub
  releases API for gemaraproj/gemara, return parsed release
  list with version tag, commit SHA, release date, and schema
  status map (schema name to Stable/Experimental)

**Checkpoint**: Release fetching works with mock API responses.

---

## Phase 2: Local Cache (FR-018)

### Tests

- [x] T107 [P] [US2] Write test
  `internal/schema/cache_test.go`: Write cache creates valid
  JSON file with timestamp
- [x] T108 [P] [US2] Write test
  `internal/schema/cache_test.go`: Read cache returns cached
  data with last-fetched timestamp
- [x] T109 [P] [US2] Write test
  `internal/schema/cache_test.go`: Stale cache is refreshed
  when upstream is available
- [x] T110 [P] [US2] Write test
  `internal/schema/cache_test.go`: Offline with cache returns
  cached data and informs user
- [x] T111 [P] [US2] Write test
  `internal/schema/cache_test.go`: Offline without cache
  returns informative error

### Implementation

- [x] T112 [US2] Add cache file name and directory constants
  to `internal/consts/consts.go`
- [x] T113 [US2] Implement `internal/schema/cache.go`:
  `WriteCache(path, releases, timestamp) error` and
  `ReadCache(path) (*CachedReleases, error)` — JSON
  serialization with `last_fetched` timestamp.
  `RefreshOrCache(ctx, fetcher, cachePath)
  (*CachedReleases, error)` — try upstream first, fall back
  to cache if unreachable

**Checkpoint**: Cache read/write works. Offline fallback works.

---

## Phase 3: Version Selection (FR-017, FR-019)

### Tests

- [x] T114 [P] [US2] Write test
  `internal/schema/selector_test.go`: DetermineVersions
  identifies correct Stable version (most recent with core
  schemas Stable)
- [x] T115 [P] [US2] Write test
  `internal/schema/selector_test.go`: DetermineVersions
  identifies correct Latest version (most recent tag)
- [x] T116 [P] [US2] Write test
  `internal/schema/selector_test.go`: User selects Stable —
  session schema version set correctly, Experimental schemas
  listed
- [x] T117 [P] [US2] Write test
  `internal/schema/selector_test.go`: User selects Latest —
  session schema version set correctly, warning about
  Experimental schemas displayed
- [x] T118 [P] [US2] Write test
  `internal/schema/selector_test.go`: User selects Latest
  with MCP installed — version compatibility check triggered,
  mismatch warning displayed
- [x] T119 [P] [US2] Write test
  `internal/schema/selector_test.go`: Mid-session version
  switch requires explicit confirmation

### Implementation

- [x] T120 [US2] Implement `internal/schema/selector.go`:
  `DetermineVersions(releases) (*VersionChoice, error)` —
  identify Stable and Latest versions from release list.
  `SelectVersion(choice, session) error` — record user's
  selection in session, trigger MCP compatibility check if
  applicable.
  `VersionChoice` struct with `StableVersion`,
  `LatestVersion`, `StableSchemaStatus`, `LatestSchemaStatus`
  fields

**Checkpoint**: Version selection logic correct for all
scenarios. Session updated with selected version.

---

## Phase 4: CLI Integration and Polish (FR-020)

### Tests

- [x] T121 [P] [US2] Write test
  `internal/cli/version_prompt_test.go`: Version prompt
  displays Stable and Latest options with version numbers
- [x] T122 [P] [US2] Write test
  `internal/cli/version_prompt_test.go`: Newer version
  available upstream triggers notification
- [x] T123 [P] [US2] Write test
  `internal/cli/version_prompt_test.go`: Offline mode uses
  cached version and informs user

### Implementation

- [x] T124 [US2] Implement `internal/cli/version_prompt.go`:
  `RunVersionSelection(ctx, session, prompter, out)
  error` — fetch or cache releases, determine versions,
  prompt user, record selection, display warnings
- [x] T125 [US2] Update `internal/cli/setup.go` to call
  `RunVersionSelection` after MCP setup completes, passing
  the session from SetupResult
- [x] T126 [US2] Integration test: full flow from MCP setup
  through version selection — verify session has correct
  schema version
- [x] T127 [US2] Verify `make build` and `make test` pass
  with zero errors

**Checkpoint**: US2 is fully functional. User can select
schema version during the setup flow. Version persists in
session for all subsequent operations.

---

## Dependencies & Execution Order

- Phase 1 -> Phase 2 -> Phase 3 -> Phase 4 (sequential)
- Phase 3 integrates with existing `internal/mcp/version.go`
  (already implemented in US1 Phase 5)
- Phase 4 integrates with existing `internal/cli/setup.go`

---

## Notes

- TDD: write tests first, confirm they fail, then implement
- All files get SPDX headers and pass `make lint`
- Commit after each phase per Conventional Commits
- US2 tasks numbered T101-T127 to avoid conflicts with US1
  (T001-T054)
