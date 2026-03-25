# Tasks: US5 — Cross-Functional Collaboration View (P5)

**Input**: `plan-us5.md`, `spec.md` (User Story 5)
**Prerequisites**: US1-US3 completed (MCP setup, schema
version selection, role and activity discovery)

---

## Phase 1: Team and Collaboration Data Model (FR-008)

**Purpose**: Define team configuration, collaboration view,
handoff point, and artifact flow types

### Tests

- [x] T401 [P] [US5] Write test
  `internal/team/team_test.go`: NewTeamConfig creates a
  team with a name and empty member list
- [x] T402 [P] [US5] Write test
  `internal/team/team_test.go`: AddMember adds a team
  member with role name and resolved layer mappings
- [x] T403 [P] [US5] Write test
  `internal/team/team_test.go`: AddMember with duplicate
  role name returns error
- [x] T404 [P] [US5] Write test
  `internal/team/team_test.go`: RemoveMember removes a
  member by name and returns true; removing nonexistent
  member returns false
- [x] T405 [P] [US5] Write test
  `internal/team/team_test.go`: TeamMember wraps a Role
  with resolved layers from ActivityProfile

### Implementation

- [x] T406 [US5] Add team configuration constants to
  `internal/consts/consts.go`: `TeamConfigDir`
  (`gemara-user-journey/teams`), `TeamConfigExt` (`.yaml`), and
  `LayerArtifacts` map (layer number to artifact type
  names)
- [x] T407 [US5] Create `internal/team/team.go` with SPDX
  header: `TeamMember` struct (Name, RoleName, Layers,
  Keywords), `TeamConfig` struct (Name, Members),
  `NewTeamConfig(name string) *TeamConfig`,
  `AddMember(member TeamMember) error`,
  `RemoveMember(name string) bool`

**Checkpoint**: Team data model compiles, all unit tests
pass.

---

## Phase 2: Handoff Detection and Artifact Flow

**Purpose**: Compute handoff points between team members,
map artifact flows, and detect layer coverage gaps

### Tests

- [x] T408 [P] [US5] Write test
  `internal/team/handoffs_test.go`: Three-role team
  (Security Engineer L1-L2, Compliance Officer L3+L5,
  Developer L4+L5) produces correct handoff points:
  Security Engineer -> Compliance Officer at L2->L3,
  Compliance Officer -> Developer at L3->L4
- [x] T409 [P] [US5] Write test
  `internal/team/handoffs_test.go`: Handoff point includes
  correct artifact types (e.g., ControlCatalog flows from
  L2 to L3)
- [x] T410 [P] [US5] Write test
  `internal/team/handoffs_test.go`: Coverage gaps for
  three-role team correctly identifies L6 and L7 as
  uncovered
- [x] T411 [P] [US5] Write test
  `internal/team/handoffs_test.go`: Adding a fourth member
  (Auditor at L5+L3) updates handoff points to include
  the new member's connections
- [x] T412 [P] [US5] Write test
  `internal/team/handoffs_test.go`: Two members sharing
  the same layer (e.g., both at L5) does not produce a
  handoff between them for that layer
- [x] T413 [P] [US5] Write test
  `internal/team/handoffs_test.go`: Single-member team
  produces no handoff points
- [x] T414 [P] [US5] Write test
  `internal/team/handoffs_test.go`: HandoffPoint includes
  tutorial references for both producing and consuming
  roles

### Implementation

- [x] T415 [US5] Create `internal/team/handoffs.go` with
  SPDX header:
  `HandoffPoint` struct (ProducerName, ProducerRole,
  ConsumerName, ConsumerRole, ProducerLayer, ConsumerLayer,
  ArtifactTypes []string, ProducerTutorials []string,
  ConsumerTutorials []string).
  `ArtifactFlow` struct (ArtifactType, SourceLayer,
  TargetLayer, Description).
  `CollaborationView` struct (TeamName, Members
  []TeamMember, Handoffs []HandoffPoint, CoverageGaps
  []int).
  `GenerateView(team *TeamConfig,
  tutorials []Tutorial) *CollaborationView` — compute
  handoff points by finding layer boundary pairs between
  team members, map artifact flows using LayerArtifacts
  constants, detect coverage gaps, attach tutorial
  references.
  `DetectCoverageGaps(team *TeamConfig) []int` — compare
  union of member layers against layers 1-7.
  `ArtifactFlows() []ArtifactFlow` — return the defined
  layer-to-layer artifact relationships

**Checkpoint**: Handoff detection works for the spec's
example team. Coverage gaps are correctly identified.

---

## Phase 3: Team Configuration Persistence

**Purpose**: Save and load team configurations as YAML

### Tests

- [x] T416 [P] [US5] Write test
  `internal/team/team_test.go`: SaveTeam writes a valid
  YAML file with team name and member list
- [x] T417 [P] [US5] Write test
  `internal/team/team_test.go`: LoadTeam reads a saved
  team and returns correct TeamConfig
- [x] T418 [P] [US5] Write test
  `internal/team/team_test.go`: ListTeams returns all
  saved teams from the teams directory
- [x] T419 [P] [US5] Write test
  `internal/team/team_test.go`: LoadTeam with nonexistent
  file returns informative error
- [x] T420 [P] [US5] Write test
  `internal/team/team_test.go`: ListTeams with missing
  directory returns empty list (no error)

### Implementation

- [x] T421 [US5] Add persistence methods to
  `internal/team/team.go`:
  `SaveTeam(dir string, team *TeamConfig) error`,
  `LoadTeam(path string) (*TeamConfig, error)`,
  `ListTeams(dir string) ([]TeamConfig, error)` — YAML
  serialization following the same pattern as
  `internal/roles/profiles.go`

**Checkpoint**: Team persistence works. Save/load
round-trip preserves all fields.

---

## Phase 4: CLI Integration and TUI Rendering

**Purpose**: Wire team operations into the CLI, add styled
output, update session state

### Tests

- [x] T422 [P] [US5] Write test
  `internal/cli/team_prompt_test.go`: RunTeamSetup prompts
  for team name and member selection, returns TeamConfig
  with correct members
- [x] T423 [P] [US5] Write test
  `internal/cli/team_prompt_test.go`: RunTeamSetup with
  3 predefined roles generates collaboration view with
  correct handoff points
- [x] T424 [P] [US5] Write test
  `internal/cli/team_prompt_test.go`: Adding a member to
  an existing team updates the collaboration view
- [x] T425 [P] [US5] Write test
  `internal/cli/team_prompt_test.go`: RunHandoffInspection
  displays artifact types and tutorial references for a
  selected handoff point

### Implementation

- [x] T426 [US5] Add collaboration view rendering to
  `internal/cli/styles.go`: `RenderCollaborationView`
  (team overview with member-layer grid),
  `RenderHandoffPoint` (card with producer/consumer roles,
  artifact types, tutorial links), `RenderCoverageGaps`
  (warning for unassigned layers),
  `RenderTeamMember` (member card with layer badges)
- [x] T427 [US5] Create `internal/cli/team_prompt.go` with
  SPDX header: `TeamPromptConfig` struct (Prompter,
  TutorialsDir, SchemaVersion, TeamConfigDir),
  `TeamPromptResult` struct (Team *TeamConfig,
  View *CollaborationView).
  `RunTeamSetup(cfg *TeamPromptConfig,
  out io.Writer) (*TeamPromptResult, error)` — prompt
  for team name, add members by selecting from predefined
  roles and saved custom profiles (via MergeWithPredefined),
  generate and display collaboration view.
  `RunHandoffInspection(view *CollaborationView,
  idx int, out io.Writer) error` — display detailed
  handoff info
- [x] T428 [US5] Update `internal/session/session.go`: add
  `TeamName string` and `TeamMemberCount int` fields with
  `SetTeamInfo(name string, count int)` method
- [x] T429 [US5] Integration test: full flow — configure a
  3-role team (Security Engineer, Compliance Officer,
  Developer), verify collaboration view shows correct
  layer mappings, handoff points at L2->L3 and L3->L4,
  and coverage gaps at L6-L7
- [x] T430 [US5] Integration test: add Auditor as 4th member,
  verify view updates with new handoff points
- [x] T431 [US5] Verify `make build`, `make test`, `make lint`
  pass with zero errors and zero warnings

**Checkpoint**: US5 is fully functional. A team of 3+ roles
can generate a collaboration view that maps roles to layers,
identifies handoff points with artifact flows, detects
coverage gaps, and allows inspection of individual handoff
points.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** (Data Model): No US5-internal dependencies —
  start immediately (depends on US3 being complete)
- **Phase 2** (Handoffs): Depends on Phase 1 (uses
  TeamConfig and TeamMember types)
- **Phase 3** (Persistence): Depends on Phase 1 (uses
  TeamConfig type)
- **Phase 4** (CLI Integration): Depends on Phases 2 and 3

### Parallel Opportunities

```text
Phase 1 (Data Model)
    │
    ├───────────────────┐
    ▼                   ▼
Phase 2 (Handoffs)   Phase 3 (Persistence)  ← parallel
    │                   │
    └───────┬───────────┘
            ▼
    Phase 4 (CLI Integration)
```

Within each phase, all tasks marked `[P]` can run in
parallel. All test tasks within a phase can run in parallel.

---

## Notes

- All test tasks follow TDD: write test, confirm it fails,
  then implement.
- Each task produces files with SPDX headers and passes
  `make lint`.
- Commit after each phase per Conventional Commits format.
- US5 tasks numbered T401-T431 to avoid conflicts with US1
  (T001-T054), US2 (T101-T127), US3 (T201-T260), and US4
  (T301-T331).
- Key invariant: the spec example team (Security Engineer,
  Compliance Officer, Developer) MUST produce the correct
  layer mappings and handoff points per acceptance scenario
  1 (SC-006).
