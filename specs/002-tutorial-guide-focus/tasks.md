# Tasks: Refocus Gemara User Journey as Tutorial Guide

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

- [x] T001 Add `ArtifactDescriptions` map with one-sentence descriptions for all 6 artifact types in `internal/consts/consts.go`
- [x] T002 Add `ArtifactWizards` map linking ThreatCatalog to `threat_assessment` and ControlCatalog to `control_catalog` in `internal/consts/consts.go`
- [x] T003 Add `ApproachWizard` and `ApproachCollaborative` authoring approach constants in `internal/consts/consts.go`
- [x] T004 Add `DefaultPreparationChecklists` map with per-artifact-type preparation items in `internal/consts/consts.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types and functions that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

### Tests

- [x] T005 [P] Write failing tests for `ArtifactRecommendation` type and `ArtifactRecommendations` function in `internal/roles/activities_test.go`: test with profile having strong L2 layers (expects ThreatCatalog + ControlCatalog), inferred L1 (expects GuidanceCatalog), empty layers (expects empty slice), L4 layer (expects no recommendations since L4 has no artifacts), and deduplication across layers
- [x] T006 [P] Write failing tests for `AutoSelectLatest` function in `internal/schema/selector_test.go`: test with valid releases (expects latest tag set on session), empty releases (expects `ErrNoVersionAvailable`), cache fallback when fetcher fails, and experimental schema detection in returned `SelectionResult`
- [x] T007 [P] Write failing tests for `HandoffSummary` struct, `BuildHandoffSummary`, and `RenderHandoffSummary` in `internal/cli/handoff_test.go`: test with L2 step (expects ThreatCatalog + wizard + MCPResources populated), L1 step (expects GuidanceCatalog + no wizard), L4 step (expects empty artifact type), MCP configured vs not configured, version mismatch warning, render output contains "OpenCode" and "gemara-mcp" references, and render output lists available tools/resources/prompts

### Implementation

- [x] T008 [P] Add `ArtifactRecommendation` struct and `ArtifactRecommendations` function to `internal/roles/activities.go` per data-model.md and contracts/cli-flow.md: iterate `ResolvedLayers`, look up `consts.LayerArtifacts`, construct recommendations with descriptions from `consts.ArtifactDescriptions`, deduplicate by artifact type keeping highest confidence
- [x] T009 [P] Add `AutoSelectLatest` function to `internal/schema/selector.go` per contracts/cli-flow.md: wrap `RefreshOrCache` + `DetermineVersions` + `SelectVersion(SelectionLatest, sess, nil, nil)`, return `*SelectionResult`
- [x] T010 [P] Create `internal/cli/handoff.go` with `HandoffSummary` struct (including `MCPResources`, `MCPTools`, `MCPConfigured` fields), `BuildHandoffSummary`, and `RenderHandoffSummary` functions per data-model.md and contracts/cli-flow.md: use existing `stepBarStyle`, `annotationLabelStyle`, `codeBlockStyle`, and `RenderWarning` from `styles.go`. Handoff must direct users to OpenCode with gemara-mcp, listing available tools, resources, and prompts. All output must be sleek and accessible for non-technical audiences (FR-018)
- [x] T011 Add `Recommendations` field (`[]ArtifactRecommendation`) to `ActivityProfile` struct in `internal/roles/activities.go`
- [x] T012 Verify all tests from T005-T007 pass with `make test`

**Checkpoint**: Foundation ready — all new types and functions exist and are tested. User story implementation can now begin.

---

## Phase 3: User Story 1 — Activity and Output Identification (Priority: P1) MVP

**Goal**: Users describe their role and activities, receive tailored layer mappings, artifact type recommendations with descriptions, and a recommended learning path.

**Independent Test**: Run Gemara User Journey, state a role and activities, verify the output includes relevant Gemara layers, recommended artifact types with descriptions and MCP wizard/collaborative labels, and a learning path.

### Tests

- [x] T013 Write failing test in `internal/cli/setup_test.go` verifying that after `RunRoleDiscovery` completes, the returned `ActivityProfile` has a populated `Recommendations` field with artifact types matching the resolved layers
- [x] T014 [P] Write failing test in `internal/cli/role_prompt_test.go` verifying that the artifact recommendation rendering output includes artifact type names, descriptions, and MCP wizard names where applicable

### Implementation

- [x] T015 [US1] Modify `RunRoleDiscovery` in `internal/cli/role_prompt.go` to call `roles.ArtifactRecommendations(profile)` after `ResolveLayerMappings` returns and populate `profile.Recommendations`
- [x] T016 [US1] Add artifact recommendation rendering to `RunRoleDiscovery` in `internal/cli/role_prompt.go`: after displaying resolved layers, render each recommendation with artifact type, description, and authoring approach (wizard name or "Collaborative authoring with MCP resources")
- [x] T017 [US1] Update `RenderSessionStatus` in `internal/cli/styles.go` (or equivalent rendering function) to include the count of recommended artifact types in the session summary
- [x] T018 [US1] Verify tests from T013-T014 pass and run `make test` to confirm no regressions

**Checkpoint**: User Story 1 is fully functional — users can identify activities and see artifact recommendations.

---

## Phase 4: User Story 3 — Latest Release Auto-Selection (Priority: P3)

**Goal**: Schema version is auto-selected to latest during setup with no user prompt. Placed before US2 because the tutorial walkthrough (US2) benefits from having the schema version already set.

**Independent Test**: Run Gemara User Journey setup and verify no version selection prompt appears, the session has a schema version set, and the version is displayed to the user.

### Tests

- [x] T019 Write failing test in `internal/cli/setup_test.go` verifying that `RunSetup` with a `VersionFetcher` configured calls `AutoSelectLatest` instead of `RunVersionSelection` and sets `Session.SchemaVersion` to the latest release tag
- [x] T020 [P] Write failing test in `internal/cli/setup_test.go` verifying that when `AutoSelectLatest` fails (network + no cache), the setup flow continues with `SchemaVersion` empty and displays a warning

### Implementation

- [x] T021 [US3] Replace the `RunVersionSelection` call in `RunSetup` in `internal/cli/setup.go` with a call to `schema.AutoSelectLatest(ctx, cfg.VersionFetcher, cfg.VersionCachePath, result.Session)`. Add a code comment referencing ADR-0003 explaining the bypass.
- [x] T022 [US3] Add version display output after auto-selection in `internal/cli/setup.go`: show the selected version tag, whether it was from cache, and any experimental schema warnings from the `SelectionResult`
- [x] T023 [US3] Handle `AutoSelectLatest` error in `internal/cli/setup.go`: on failure, log a warning ("Schema version could not be resolved; proceeding without version constraint"), set `Session.SchemaVersion` to empty, and continue setup
- [x] T024 [US3] Verify tests from T019-T020 pass and run `make test` to confirm no regressions

**Checkpoint**: User Story 3 is fully functional — version auto-selects with no prompt.

---

## Phase 5: User Story 2 — Tutorial Walkthrough Before Authoring (Priority: P2)

**Goal**: Tutorial sections are presented with role-tailored Why/How/What annotations, activity-keyword sections are highlighted, and users understand the authoring procedure before using the MCP server.

**Independent Test**: Navigate through a tutorial for a specific artifact type and verify sections include role-tailored annotations, relevant sections are highlighted, and a post-tutorial completion summary appears.

### Tests

- [x] T025 Write failing test in `internal/cli/tutorial_prompt_test.go` verifying that when a tutorial is marked complete, `BuildHandoffSummary` is called with the completed step and session, and `RenderHandoffSummary` produces output containing the artifact type and schema definition

### Implementation

- [x] T026 [US2] Add `Session` and `SelectionResult` fields to `TutorialPlayerConfig` (or equivalent config struct) in `internal/cli/tutorial_prompt.go` so the tutorial player has access to session state and version selection results for building handoff summaries
- [x] T027 [US2] Modify the tutorial completion handler in `internal/cli/tutorial_prompt.go` (the `navComplete` case): after rendering the existing "Completed: <title>" success message, call `BuildHandoffSummary(step, cfg.Session, cfg.SelectionResult)` and `RenderHandoffSummary(summary, out)`
- [x] T028 [US2] Update the caller of `RunTutorialPlayer` in `cmd/gemara-user-journey/main.go` to pass the `Session` and `SelectionResult` through the config
- [x] T029 [US2] Verify test from T025 passes and run `make test` to confirm no regressions

**Checkpoint**: User Story 2 is fully functional — tutorials end with a handoff summary directing users to OpenCode with the gemara-mcp server.

---

## Phase 6: User Story 4 — Clear Handoff to OpenCode with gemara-mcp (Priority: P4)

**Goal**: Post-tutorial handoff summary directs users to open an OpenCode session where the gemara-mcp server provides tools (`validate_gemara_artifact`), resources (`gemara://lexicon`, `gemara://schema/definitions`), and wizard prompts (`threat_assessment`, `control_catalog`) for authoring. When gemara-mcp is not configured, setup instructions referencing `./gemara-user-journey --doctor` are shown.

**Independent Test**: Complete a tutorial when gemara-mcp is configured (verify OpenCode launch instructions, wizard name, and available tools/resources shown) and when not configured (verify `./gemara-user-journey --doctor` reference and `cue vet` command shown).

### Tests

- [x] T030 Write failing test in `internal/cli/handoff_test.go` verifying `RenderHandoffSummary` output includes "OpenCode", "gemara-mcp", the wizard prompt name, and lists `validate_gemara_artifact`, `gemara://lexicon`, and `gemara://schema/definitions` when `MCPConfigured` is true
- [x] T031 [P] Write failing test in `internal/cli/handoff_test.go` verifying `RenderHandoffSummary` output includes `./gemara-user-journey --doctor` reference, `opencode.json` setup instructions, and `cue vet` command when `MCPConfigured` is false

### Implementation

- [x] T032 [US4] Enhance `RenderHandoffSummary` in `internal/cli/handoff.go` to render the configured path: show "Available in OpenCode" section listing tools, resources, and prompts from the `HandoffSummary` fields; show instructions to launch `opencode` and use the wizard prompt with the gemara-mcp server
- [x] T033 [US4] Enhance `RenderHandoffSummary` in `internal/cli/handoff.go` to render the not-configured path: show a clear note that gemara-mcp is not yet configured, reference `./gemara-user-journey --doctor` for environment verification, explain how to configure `opencode.json`, and show `cue vet -c -d '<SchemaDef>' github.com/gemaraproj/gemara@latest artifact.yaml` as a manual validation alternative
- [x] T034 [US4] Add version mismatch warning rendering to `RenderHandoffSummary` in `internal/cli/handoff.go`: when `VersionMismatch` is true, render a `RenderWarning` noting the discrepancy and recommending post-authoring validation
- [x] T035 [US4] Verify tests from T030-T031 pass and run `make test` to confirm no regressions

**Checkpoint**: User Story 4 is fully functional — handoff summary directs to OpenCode with gemara-mcp, adapts to configuration state and version.

---

## Phase 7: User Story 5 — Deprecate Version Switching as Future Work (Priority: P5)

**Goal**: Version selection code is preserved but bypassed, documented with an ADR and code comments, and the main menu no longer offers version switching.

**Independent Test**: Verify no version selection prompt appears, version switching code compiles and has passing tests, and ADR-0003 exists.

### Tests

- [x] T036 Write failing test in `internal/cli/version_prompt_test.go` verifying that `RunVersionSelection` still compiles, accepts valid config, and functions correctly when called directly (proving it is preserved and functional, not broken by the bypass)

### Implementation

- [x] T037 [US5] Add a bypass comment header to `RunVersionSelection` in `internal/cli/version_prompt.go`: document that this function is intentionally bypassed in the active flow per ADR-0003, retained for planned future re-enablement, and can be re-enabled by replacing the `AutoSelectLatest` call in `setup.go` with `RunVersionSelection`
- [x] T038 [US5] Remove or disable the "Switch schema version" menu option in `cmd/gemara-user-journey/main.go` (if one exists in the main menu): add a comment noting the deferral per ADR-0003 — N/A: no version switch menu option exists
- [x] T039 [US5] Create `docs/adrs/ADR-0003-version-selection-deferral.md` per research.md section R5: Context (friction during onboarding, Gemara User Journey's tutorial focus), Decision (auto-select latest, preserve code), Consequences (simpler onboarding, re-enablement path)
- [x] T040 [US5] Verify test from T036 passes and run `make test` to confirm no regressions

**Checkpoint**: User Story 5 is complete — version switching is cleanly deferred with documentation.

---

## Phase 8: User Story 6 — GitHub README Documentation (Priority: P6)

**Goal**: Rewrite README.md as a concise landing page with a web UI screenshot, hyperlinked dependency installation links, and a user journey narrative. Move detailed content to `docs/`. The README enables a new user to understand what Gemara User Journey does, install prerequisites, and begin the guided experience within minutes.

**Independent Test**: A new user reads only the README, follows the dependency links to install prerequisites, and confirms they can build and launch Gemara User Journey without consulting other documentation.

### Implementation

- [ ] T052 [P] [US6] Create `docs/images/` directory and add a manually captured screenshot of the web UI Results view (showing resolved layers and artifact recommendations) as `docs/images/web-ui-preview.png`. To capture: run `make web-dev`, navigate to the Results step in the browser, and take a screenshot.
- [ ] T053 [P] [US6] Create `docs/layer-reference.md` by extracting the Gemara Layer Reference content from the current `README.md` (lines 142-161). Include a title (`# Gemara Layer Reference`), a brief context paragraph, the full 7-layer table, and a link back to the README.
- [ ] T054 [P] [US6] Create `docs/project-structure.md` by extracting the Project Structure content from the current `README.md` (lines 407-452). Include a title (`# Project Structure`), a brief context paragraph, the full directory tree with descriptions, the ADRs section (lines 490-496), and a link back to the README.
- [ ] T055 [P] [US6] Create `docs/mcp-update-guide.md` by extracting the "Keeping gemara-mcp Up to Date" content from the current `README.md` (lines 329-388). Include a title (`# Keeping gemara-mcp Up to Date`), a brief context paragraph, sync instructions for both direct-clone and fork workflows, the verification steps, and a link back to the README.
- [ ] T056 [US6] Rewrite `README.md` as a concise landing page (~120-150 lines) following the structure defined in `contracts/cli-flow.md` Documentation Contract. The README MUST include: (1) a one-paragraph project summary positioning Gemara User Journey as a role-based tutorial guide and distinguishing it from the MCP server (FR-020), (2) the web UI screenshot via `![Gemara User Journey Web UI](docs/images/web-ui-preview.png)` (FR-022), (3) a User Journey section with 3-step narrative: Discover (role + activity identification), Learn (tailored tutorial walkthrough), Author (handoff to OpenCode with gemara-mcp for artifact authoring using Gemara schemas and MCP server tools) (FR-023), (4) a Prerequisites section with hyperlinked dependency names — [Go](https://go.dev/dl/) 1.21+, [CUE](https://cuelang.org/docs/introduction/installation/) v0.15.1+, [OpenCode](https://opencode.ai), [Git](https://git-scm.com/downloads), [gemara-mcp](https://github.com/gemaraproj/gemara-mcp) build from source (FR-021), (5) a Getting Started section with exactly 4 steps: clone and build, verify with `./gemara-user-journey --doctor`, launch `opencode`, tell OpenCode your role (FR-024), (6) an Upstream Projects table linking Gemara and gemara-mcp repos, (7) a Learn More section linking to `docs/layer-reference.md`, `docs/project-structure.md`, `docs/mcp-update-guide.md`, and `CONTRIBUTING.md` (FR-024), (8) a one-line License section. Do NOT include inline platform-specific install commands, `<details>` HTML, SDK references, or content that belongs in `docs/` files.
- [ ] T057 [US6] Verify all links in the rewritten `README.md` resolve correctly: `docs/images/web-ui-preview.png` displays, `docs/layer-reference.md` exists, `docs/project-structure.md` exists, `docs/mcp-update-guide.md` exists, `CONTRIBUTING.md` exists, `LICENSE` exists, all external URLs are valid (SC-011).
- [ ] T058 [US6] Verify `README.md` line count is between 100 and 160 lines (conciseness check per FR-024).

**Checkpoint**: User Story 6 is complete — README is a concise landing page with screenshot, dependency links, user journey narrative, and links to detailed docs.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, UX polish, cleanup, and cross-cutting improvements

- [x] T041 Run `make lint` and fix any linting issues across all modified files
- [x] T042 Run `make fmt` to ensure all files are formatted with `goimports`
- [x] T043 Verify SPDX license headers on all new files: `internal/cli/handoff.go`, `internal/cli/handoff_test.go`, `docs/adrs/ADR-0003-version-selection-deferral.md`
- [x] T044 Run full `make test` and verify zero failures across the entire test suite
- [x] T045 Run `make build` and verify the binary builds with zero errors
- [x] T046 [P] Run `./gemara-user-journey --doctor` and verify it still functions correctly with no changes to its output or behavior (FR-017)
- [x] T047 [P] Verify quickstart.md scenario: launch Gemara User Journey, confirm no version prompt, confirm artifact recommendations display, complete a tutorial, confirm handoff summary directs to OpenCode with gemara-mcp tools/resources listed
- [x] T048 [P] Verify edge case: run setup with no network and no cache, confirm graceful degradation with warning message
- [x] T049 Review all terminal output across every flow (activity identification, tutorial navigation, handoff summary) for consistent visual styling, clear spacing, scannable format, and accessibility for non-technical users (FR-018). Fix any rough edges in rendering functions in `internal/cli/styles.go` and `internal/cli/handoff.go`
- [x] T050 Review all modified files to confirm no MCP authoring wizard replication (FR-010 compliance check)
- [x] T051 Verify all handoff summaries reference OpenCode by name, list gemara-mcp tools/resources/prompts, and provide actionable next steps for each artifact type (FR-019)
- [ ] T059 Run `make build` and `make test` after US6 documentation changes to confirm no regressions
- [ ] T060 Visual inspection: push branch and verify README renders correctly on GitHub with screenshot displayed, all links working, and no broken images (SC-011)
- [ ] T061 Walkthrough test: follow only the README instructions as a new user — verify the 4-step Getting Started is sufficient to reach a first tutorial interaction within 10 minutes (SC-010)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately ✅ Complete
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) — BLOCKS all user stories ✅ Complete
- **US1 (Phase 3)**: Depends on Foundational (Phase 2) ✅ Complete
- **US3 (Phase 4)**: Depends on Foundational (Phase 2) ✅ Complete
- **US2 (Phase 5)**: Depends on Foundational (Phase 2) ✅ Complete
- **US4 (Phase 6)**: Depends on US2 (Phase 5) ✅ Complete
- **US5 (Phase 7)**: Depends on US3 (Phase 4) ✅ Complete
- **US6 (Phase 8)**: Depends on US1-US4 being defined (content accuracy) — can start now
- **Polish (Phase 9)**: Depends on all user stories being complete (including US6)

### User Story Dependencies

```text
Phase 1 (Setup)               ✅ Complete
    │
Phase 2 (Foundational)         ✅ Complete
    │
    ├─── Phase 3 (US1)         ✅ Complete
    ├─── Phase 4 (US3)         ✅ Complete
    │         │
    │         └── Phase 7 (US5) ✅ Complete
    └─── Phase 5 (US2)         ✅ Complete
              │
              └── Phase 6 (US4) ✅ Complete

Phase 8 (US6: README Docs)    ← START HERE (all dependencies met)
    │
Phase 9 (Polish)              ← After US6
```

### Within User Story 6

1. Create `docs/` files first (T052-T055) — can run in parallel
2. Rewrite README after docs files exist (T056) — depends on T052-T055
3. Verify links and line count (T057-T058) — depends on T056

### Parallel Opportunities

- **Phase 1-7**: All complete — no action needed
- **Phase 8 (US6)**: T052-T055 (create docs files + screenshot) can all run in parallel; T056 (rewrite README) depends on all four; T057-T058 (verification) depend on T056
- **Phase 9**: T059-T061 run after US6 is complete

---

## Parallel Example: User Story 6 (Active Phase)

```text
# Launch all docs file creation tasks in parallel:
Task: T052 "Capture web UI screenshot to docs/images/web-ui-preview.png"
Task: T053 "Create docs/layer-reference.md from README"
Task: T054 "Create docs/project-structure.md from README"
Task: T055 "Create docs/mcp-update-guide.md from README"

# After all docs files exist, rewrite README:
Task: T056 "Rewrite README.md as concise landing page"

# After README is rewritten, verify:
Task: T057 "Verify all README links resolve correctly"
Task: T058 "Verify README line count is 100-160 lines"
```

---

## Implementation Strategy

### Current State: US1-US5 Complete

Phases 1-7 are fully implemented and tested. The remaining work is US6 (README documentation) and final polish.

### Execution Plan for US6

1. Create all `docs/` files in parallel (T052-T055)
2. Rewrite README.md (T056) — the single largest task
3. Verify links and line count (T057-T058)
4. Run build/test validation (T059)
5. Visual inspection on GitHub (T060)
6. New-user walkthrough test (T061)

### Incremental Delivery (Full History)

1. ✅ Setup + Foundational → Foundation ready
2. ✅ US1 → Activity/output identification (MVP)
3. ✅ US3 → Version auto-selects, setup simplified
4. ✅ US2 → Tutorial walkthrough with OpenCode handoff
5. ✅ US4 → Handoff to OpenCode + gemara-mcp with fallback
6. ✅ US5 → Clean deferral documented
7. ⬜ US6 → README rewrite as landing page with screenshot, links, user journey
8. ⬜ Polish → Final validation and visual inspection

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Constitution Principle III (TDD) requires tests before implementation (US1-US5 code tasks)
- Constitution Principle VI (Decision Documentation) requires ADR-0003
- Constitution Principle VII (Centralized Constants) requires all new strings in `consts.go`
- All new source files require SPDX license headers per coding standards
- Commit after each task or logical group using conventional commits
- US3 is implemented before US2 (despite P3 vs P2 priority) because auto-version benefits the tutorial walkthrough
- The `--doctor` command is explicitly unchanged (FR-017) — verified in T046
- All terminal output must be sleek and accessible for all audiences (FR-018) — reviewed in T049
- Post-tutorial handoff directs to OpenCode + gemara-mcp (FR-019) — verified in T051
- US6 is documentation-only (Markdown files) — no Go code changes, no TDD required
- README must NOT reference an SDK — only Gemara schemas and MCP server tools (clarification)
- Screenshot stored in `docs/images/` and referenced via relative path (clarification)
- Detailed content displaced to `docs/` files, linked from README (FR-024)
