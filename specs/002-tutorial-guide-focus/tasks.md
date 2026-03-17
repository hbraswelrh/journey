# Tasks: Refocus Pac-Man as Tutorial Guide

**Input**: Design documents from `/specs/002-tutorial-guide-focus/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-flow.md

**Tests**: Included per Constitution Principle III (TDD required). Tests MUST be written and confirmed failing before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add new constants and shared types that multiple user stories depend on

- [ ] T001 Add `ArtifactDescriptions` map with one-sentence descriptions for all 6 artifact types in `internal/consts/consts.go`
- [ ] T002 Add `ArtifactWizards` map linking ThreatCatalog to `threat_assessment` and ControlCatalog to `control_catalog` in `internal/consts/consts.go`
- [ ] T003 Add `ApproachWizard` and `ApproachCollaborative` authoring approach constants in `internal/consts/consts.go`
- [ ] T004 Add `DefaultPreparationChecklists` map with per-artifact-type preparation items in `internal/consts/consts.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types and functions that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

### Tests

- [ ] T005 [P] Write failing tests for `ArtifactRecommendation` type and `ArtifactRecommendations` function in `internal/roles/activities_test.go`: test with profile having strong L2 layers (expects ThreatCatalog + ControlCatalog), inferred L1 (expects GuidanceCatalog), empty layers (expects empty slice), L4 layer (expects no recommendations since L4 has no artifacts), and deduplication across layers
- [ ] T006 [P] Write failing tests for `AutoSelectLatest` function in `internal/schema/selector_test.go`: test with valid releases (expects latest tag set on session), empty releases (expects `ErrNoVersionAvailable`), cache fallback when fetcher fails, and experimental schema detection in returned `SelectionResult`
- [ ] T007 [P] Write failing tests for `HandoffSummary` struct, `BuildHandoffSummary`, and `RenderHandoffSummary` in `internal/cli/handoff_test.go`: test with L2 step (expects ThreatCatalog + wizard + MCPResources populated), L1 step (expects GuidanceCatalog + no wizard), L4 step (expects empty artifact type), MCP configured vs not configured, version mismatch warning, render output contains "OpenCode" and "gemara-mcp" references, and render output lists available tools/resources/prompts

### Implementation

- [ ] T008 [P] Add `ArtifactRecommendation` struct and `ArtifactRecommendations` function to `internal/roles/activities.go` per data-model.md and contracts/cli-flow.md: iterate `ResolvedLayers`, look up `consts.LayerArtifacts`, construct recommendations with descriptions from `consts.ArtifactDescriptions`, deduplicate by artifact type keeping highest confidence
- [ ] T009 [P] Add `AutoSelectLatest` function to `internal/schema/selector.go` per contracts/cli-flow.md: wrap `RefreshOrCache` + `DetermineVersions` + `SelectVersion(SelectionLatest, sess, nil, nil)`, return `*SelectionResult`
- [ ] T010 [P] Create `internal/cli/handoff.go` with `HandoffSummary` struct (including `MCPResources`, `MCPTools`, `MCPConfigured` fields), `BuildHandoffSummary`, and `RenderHandoffSummary` functions per data-model.md and contracts/cli-flow.md: use existing `stepBarStyle`, `annotationLabelStyle`, `codeBlockStyle`, and `RenderWarning` from `styles.go`. Handoff must direct users to OpenCode with gemara-mcp, listing available tools, resources, and prompts. All output must be sleek and accessible for non-technical audiences (FR-018)
- [ ] T011 Add `Recommendations` field (`[]ArtifactRecommendation`) to `ActivityProfile` struct in `internal/roles/activities.go`
- [ ] T012 Verify all tests from T005-T007 pass with `make test`

**Checkpoint**: Foundation ready — all new types and functions exist and are tested. User story implementation can now begin.

---

## Phase 3: User Story 1 — Activity and Output Identification (Priority: P1) MVP

**Goal**: Users describe their role and activities, receive tailored layer mappings, artifact type recommendations with descriptions, and a recommended learning path.

**Independent Test**: Run Pac-Man, state a role and activities, verify the output includes relevant Gemara layers, recommended artifact types with descriptions and MCP wizard/collaborative labels, and a learning path.

### Tests

- [ ] T013 Write failing test in `internal/cli/setup_test.go` verifying that after `RunRoleDiscovery` completes, the returned `ActivityProfile` has a populated `Recommendations` field with artifact types matching the resolved layers
- [ ] T014 [P] Write failing test in `internal/cli/role_prompt_test.go` verifying that the artifact recommendation rendering output includes artifact type names, descriptions, and MCP wizard names where applicable

### Implementation

- [ ] T015 [US1] Modify `RunRoleDiscovery` in `internal/cli/role_prompt.go` to call `roles.ArtifactRecommendations(profile)` after `ResolveLayerMappings` returns and populate `profile.Recommendations`
- [ ] T016 [US1] Add artifact recommendation rendering to `RunRoleDiscovery` in `internal/cli/role_prompt.go`: after displaying resolved layers, render each recommendation with artifact type, description, and authoring approach (wizard name or "Collaborative authoring with MCP resources")
- [ ] T017 [US1] Update `RenderSessionStatus` in `internal/cli/styles.go` (or equivalent rendering function) to include the count of recommended artifact types in the session summary
- [ ] T018 [US1] Verify tests from T013-T014 pass and run `make test` to confirm no regressions

**Checkpoint**: User Story 1 is fully functional — users can identify activities and see artifact recommendations.

---

## Phase 4: User Story 3 — Latest Release Auto-Selection (Priority: P3)

**Goal**: Schema version is auto-selected to latest during setup with no user prompt. Placed before US2 because the tutorial walkthrough (US2) benefits from having the schema version already set.

**Independent Test**: Run Pac-Man setup and verify no version selection prompt appears, the session has a schema version set, and the version is displayed to the user.

### Tests

- [ ] T019 Write failing test in `internal/cli/setup_test.go` verifying that `RunSetup` with a `VersionFetcher` configured calls `AutoSelectLatest` instead of `RunVersionSelection` and sets `Session.SchemaVersion` to the latest release tag
- [ ] T020 [P] Write failing test in `internal/cli/setup_test.go` verifying that when `AutoSelectLatest` fails (network + no cache), the setup flow continues with `SchemaVersion` empty and displays a warning

### Implementation

- [ ] T021 [US3] Replace the `RunVersionSelection` call in `RunSetup` in `internal/cli/setup.go` with a call to `schema.AutoSelectLatest(ctx, cfg.VersionFetcher, cfg.VersionCachePath, result.Session)`. Add a code comment referencing ADR-0003 explaining the bypass.
- [ ] T022 [US3] Add version display output after auto-selection in `internal/cli/setup.go`: show the selected version tag, whether it was from cache, and any experimental schema warnings from the `SelectionResult`
- [ ] T023 [US3] Handle `AutoSelectLatest` error in `internal/cli/setup.go`: on failure, log a warning ("Schema version could not be resolved; proceeding without version constraint"), set `Session.SchemaVersion` to empty, and continue setup
- [ ] T024 [US3] Verify tests from T019-T020 pass and run `make test` to confirm no regressions

**Checkpoint**: User Story 3 is fully functional — version auto-selects with no prompt.

---

## Phase 5: User Story 2 — Tutorial Walkthrough Before Authoring (Priority: P2)

**Goal**: Tutorial sections are presented with role-tailored Why/How/What annotations, activity-keyword sections are highlighted, and users understand the authoring procedure before using the MCP server.

**Independent Test**: Navigate through a tutorial for a specific artifact type and verify sections include role-tailored annotations, relevant sections are highlighted, and a post-tutorial completion summary appears.

### Tests

- [ ] T025 Write failing test in `internal/cli/tutorial_prompt_test.go` verifying that when a tutorial is marked complete, `BuildHandoffSummary` is called with the completed step and session, and `RenderHandoffSummary` produces output containing the artifact type and schema definition

### Implementation

- [ ] T026 [US2] Add `Session` and `SelectionResult` fields to `TutorialPlayerConfig` (or equivalent config struct) in `internal/cli/tutorial_prompt.go` so the tutorial player has access to session state and version selection results for building handoff summaries
- [ ] T027 [US2] Modify the tutorial completion handler in `internal/cli/tutorial_prompt.go` (the `navComplete` case): after rendering the existing "Completed: <title>" success message, call `BuildHandoffSummary(step, cfg.Session, cfg.SelectionResult)` and `RenderHandoffSummary(summary, out)`
- [ ] T028 [US2] Update the caller of `RunTutorialPlayer` in `cmd/pacman/main.go` to pass the `Session` and `SelectionResult` through the config
- [ ] T029 [US2] Verify test from T025 passes and run `make test` to confirm no regressions

**Checkpoint**: User Story 2 is fully functional — tutorials end with a handoff summary directing users to OpenCode with the gemara-mcp server.

---

## Phase 6: User Story 4 — Clear Handoff to OpenCode with gemara-mcp (Priority: P4)

**Goal**: Post-tutorial handoff summary directs users to open an OpenCode session where the gemara-mcp server provides tools (`validate_gemara_artifact`), resources (`gemara://lexicon`, `gemara://schema/definitions`), and wizard prompts (`threat_assessment`, `control_catalog`) for authoring. When gemara-mcp is not configured, setup instructions referencing `./pacman --doctor` are shown.

**Independent Test**: Complete a tutorial when gemara-mcp is configured (verify OpenCode launch instructions, wizard name, and available tools/resources shown) and when not configured (verify `./pacman --doctor` reference and `cue vet` command shown).

### Tests

- [ ] T030 Write failing test in `internal/cli/handoff_test.go` verifying `RenderHandoffSummary` output includes "OpenCode", "gemara-mcp", the wizard prompt name, and lists `validate_gemara_artifact`, `gemara://lexicon`, and `gemara://schema/definitions` when `MCPConfigured` is true
- [ ] T031 [P] Write failing test in `internal/cli/handoff_test.go` verifying `RenderHandoffSummary` output includes `./pacman --doctor` reference, `opencode.json` setup instructions, and `cue vet` command when `MCPConfigured` is false

### Implementation

- [ ] T032 [US4] Enhance `RenderHandoffSummary` in `internal/cli/handoff.go` to render the configured path: show "Available in OpenCode" section listing tools, resources, and prompts from the `HandoffSummary` fields; show instructions to launch `opencode` and use the wizard prompt with the gemara-mcp server
- [ ] T033 [US4] Enhance `RenderHandoffSummary` in `internal/cli/handoff.go` to render the not-configured path: show a clear note that gemara-mcp is not yet configured, reference `./pacman --doctor` for environment verification, explain how to configure `opencode.json`, and show `cue vet -c -d '<SchemaDef>' github.com/gemaraproj/gemara@latest artifact.yaml` as a manual validation alternative
- [ ] T034 [US4] Add version mismatch warning rendering to `RenderHandoffSummary` in `internal/cli/handoff.go`: when `VersionMismatch` is true, render a `RenderWarning` noting the discrepancy and recommending post-authoring validation
- [ ] T035 [US4] Verify tests from T030-T031 pass and run `make test` to confirm no regressions

**Checkpoint**: User Story 4 is fully functional — handoff summary directs to OpenCode with gemara-mcp, adapts to configuration state and version.

---

## Phase 7: User Story 5 — Deprecate Version Switching as Future Work (Priority: P5)

**Goal**: Version selection code is preserved but bypassed, documented with an ADR and code comments, and the main menu no longer offers version switching.

**Independent Test**: Verify no version selection prompt appears, version switching code compiles and has passing tests, and ADR-0003 exists.

### Tests

- [ ] T036 Write failing test in `internal/cli/version_prompt_test.go` verifying that `RunVersionSelection` still compiles, accepts valid config, and functions correctly when called directly (proving it is preserved and functional, not broken by the bypass)

### Implementation

- [ ] T037 [US5] Add a bypass comment header to `RunVersionSelection` in `internal/cli/version_prompt.go`: document that this function is intentionally bypassed in the active flow per ADR-0003, retained for planned future re-enablement, and can be re-enabled by replacing the `AutoSelectLatest` call in `setup.go` with `RunVersionSelection`
- [ ] T038 [US5] Remove or disable the "Switch schema version" menu option in `cmd/pacman/main.go` (if one exists in the main menu): add a comment noting the deferral per ADR-0003
- [ ] T039 [US5] Create `docs/adrs/ADR-0003-version-selection-deferral.md` per research.md section R5: Context (friction during onboarding, Pac-Man's tutorial focus), Decision (auto-select latest, preserve code), Consequences (simpler onboarding, re-enablement path)
- [ ] T040 [US5] Verify test from T036 passes and run `make test` to confirm no regressions

**Checkpoint**: User Story 5 is complete — version switching is cleanly deferred with documentation.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, UX polish, cleanup, and cross-cutting improvements

- [ ] T041 Run `make lint` and fix any linting issues across all modified files
- [ ] T042 Run `make fmt` to ensure all files are formatted with `goimports`
- [ ] T043 Verify SPDX license headers on all new files: `internal/cli/handoff.go`, `internal/cli/handoff_test.go`, `docs/adrs/ADR-0003-version-selection-deferral.md`
- [ ] T044 Run full `make test` and verify zero failures across the entire test suite
- [ ] T045 Run `make build` and verify the binary builds with zero errors
- [ ] T046 [P] Run `./pacman --doctor` and verify it still functions correctly with no changes to its output or behavior (FR-017)
- [ ] T047 [P] Verify quickstart.md scenario: launch Pac-Man, confirm no version prompt, confirm artifact recommendations display, complete a tutorial, confirm handoff summary directs to OpenCode with gemara-mcp tools/resources listed
- [ ] T048 [P] Verify edge case: run setup with no network and no cache, confirm graceful degradation with warning message
- [ ] T049 Review all terminal output across every flow (activity identification, tutorial navigation, handoff summary) for consistent visual styling, clear spacing, scannable format, and accessibility for non-technical users (FR-018). Fix any rough edges in rendering functions in `internal/cli/styles.go` and `internal/cli/handoff.go`
- [ ] T050 Review all modified files to confirm no MCP authoring wizard replication (FR-010 compliance check)
- [ ] T051 Verify all handoff summaries reference OpenCode by name, list gemara-mcp tools/resources/prompts, and provide actionable next steps for each artifact type (FR-019)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Foundational (Phase 2) — can start after T012
- **US3 (Phase 4)**: Depends on Foundational (Phase 2) — can start after T012; independent of US1
- **US2 (Phase 5)**: Depends on Foundational (Phase 2) — benefits from US3 (schema version set) but not blocked by it
- **US4 (Phase 6)**: Depends on US2 (Phase 5) — extends the handoff summary rendering
- **US5 (Phase 7)**: Depends on US3 (Phase 4) — documents the bypass introduced in US3
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

```text
Phase 1 (Setup)
    │
Phase 2 (Foundational)
    │
    ├─── Phase 3 (US1: Activity ID)     ← MVP, start here
    │
    ├─── Phase 4 (US3: Auto-Select)     ← Can parallel with US1
    │         │
    │         └── Phase 7 (US5: Deprecate) ← Depends on US3
    │
    └─── Phase 5 (US2: Tutorial Walk)   ← Can parallel with US1/US3
              │
              └── Phase 6 (US4: Handoff) ← Depends on US2
                                          
Phase 8 (Polish) ← After all stories
```

### Within Each User Story

1. Tests MUST be written and FAIL before implementation
2. Types/structs before functions
3. Core functions before integration/wiring
4. Unit verification before moving to next story

### Parallel Opportunities

- **Phase 1**: T001-T004 all modify `consts.go` — execute sequentially
- **Phase 2**: T005-T007 (tests) can run in parallel; T008-T010 (implementations) can run in parallel
- **Phase 3 + Phase 4**: US1 and US3 can run in parallel after Phase 2
- **Phase 5**: US2 can run in parallel with US1/US3 after Phase 2
- **Phase 6 + Phase 7**: US4 depends on US2; US5 depends on US3 — these two can run in parallel with each other

---

## Parallel Example: Foundational Phase

```text
# Launch all foundational tests in parallel:
Task: T005 "Test ArtifactRecommendations in internal/roles/activities_test.go"
Task: T006 "Test AutoSelectLatest in internal/schema/selector_test.go"
Task: T007 "Test HandoffSummary in internal/cli/handoff_test.go"

# After tests written, launch all implementations in parallel:
Task: T008 "Implement ArtifactRecommendations in internal/roles/activities.go"
Task: T009 "Implement AutoSelectLatest in internal/schema/selector.go"
Task: T010 "Create handoff.go in internal/cli/handoff.go"
```

## Parallel Example: User Stories After Foundational

```text
# US1 and US3 can start simultaneously:
Developer A: T013 → T014 → T015 → T016 → T017 → T018 (US1)
Developer B: T019 → T020 → T021 → T022 → T023 → T024 (US3)

# US2 can start once foundational is done (independent of US1/US3):
Developer C: T025 → T026 → T027 → T028 → T029 (US2)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (constants)
2. Complete Phase 2: Foundational (types, functions, tests)
3. Complete Phase 3: User Story 1 (activity identification)
4. **STOP and VALIDATE**: Users can identify activities and see artifact recommendations
5. This alone delivers the core differentiator from the MCP server

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Test independently → MVP delivers activity/output identification
3. Add US3 → Test independently → Version auto-selects, setup simplified
4. Add US2 → Test independently → Tutorial walkthrough with OpenCode handoff
5. Add US4 → Test independently → OpenCode + gemara-mcp handoff with fallback
6. Add US5 → Test independently → Clean deferral documented
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With 2-3 developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (Phase 3) then US4 (Phase 6)
   - Developer B: US3 (Phase 4) then US5 (Phase 7)
   - Developer C: US2 (Phase 5)
3. Stories complete and integrate independently
4. Team reconvenes for Phase 8 (Polish)

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Constitution Principle III (TDD) requires tests before implementation
- Constitution Principle VI (Decision Documentation) requires ADR-0003
- Constitution Principle VII (Centralized Constants) requires all new strings in `consts.go`
- All new files require SPDX license headers per coding standards
- Commit after each task or logical group using conventional commits
- US3 is implemented before US2 (despite P3 vs P2 priority) because auto-version benefits the tutorial walkthrough
- The `--doctor` command is explicitly unchanged (FR-017) — verified in T046
- All terminal output must be sleek and accessible for all audiences (FR-018) — reviewed in T049
- Post-tutorial handoff directs to OpenCode + gemara-mcp (FR-019) — verified in T051
