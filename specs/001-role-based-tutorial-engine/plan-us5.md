# Implementation Plan: US5 — Cross-Functional
# Collaboration View (P5)

**Branch**: `001-role-based-tutorial-engine` | **Date**: 2026-03-13
**Spec**: [spec.md](spec.md) — User Story 5
**Depends on**: US1 (MCP Server Setup), US2 (Schema Version
Selection), US3 (Role and Activity Discovery) — all completed

## Summary

US5 implements the cross-functional collaboration view that
maps team members' roles to Gemara layers, identifies handoff
points between roles, shows which artifacts flow across layer
boundaries, and detects coverage gaps where no team member
owns a layer. This transforms Pac-Man from an individual
learning tool into a team coordination tool.

The collaboration view is built from existing `Role` and
`ActivityProfile` types (US3) and the artifact type constants
(US1). A team is configured by adding roles (predefined or
custom), each with their resolved layer mappings. The system
then computes handoff points — boundaries where one role
produces artifacts that another role consumes — and links
these handoffs to relevant tutorials for both the producing
and consuming roles.

This plan covers FR-008, FR-024 (team assignment), and the
US5 acceptance scenarios 1-3.

## Technical Context

**Language/Version**: Go 1.26.1
**Dependencies**: Existing `internal/roles/` (Role type,
PredefinedRoles, ActivityProfile, custom profiles),
`internal/tutorials/` (Tutorial type, LearningPath),
`internal/consts/` (artifact type constants, layer constants),
`internal/cli/` (TUI styles, prompter interfaces),
`internal/session/` (session state)
**Storage**: Local filesystem — team configurations stored as
YAML in `~/.config/pacman/teams/`
**Testing**: `go test ./...` via `make test`; TDD per
constitution
**Constraints**: Team view is a read-only visualization; it
does not modify role profiles or learning paths. Handoff
points are derived from layer adjacency and artifact type
associations. Coverage gaps compare team layer coverage
against the full seven-layer model.

## Constitution Check

| Principle | Status | Notes |
|:---|:---|:---|
| I. Schema Conformance | Pass | No schema output produced; artifact types referenced by constant |
| II. Gemara Layer Fidelity | Pass | Layer boundaries and handoffs respect the seven-layer model |
| III. TDD | Pass | Tests written before implementation per phase |
| IV. Tutorial-First Design | Pass | Handoff points link to relevant tutorials for both roles |
| V. Incremental Delivery | Pass | US5 is independently usable after US3 |
| VI. Decision Documentation | N/A | No new ADRs anticipated; patterns consistent with prior stories |
| VII. Centralized Constants | Pass | Artifact types, layer constants already in `internal/consts/` |
| VIII. Composability | Pass | Collaboration view is an independent operation |
| IX. Convention Over Configuration | Pass | Default layer-to-artifact mapping requires zero config |

## Source Code

```text
internal/
├── team/
│   ├── team.go               # TeamConfig, TeamMember,
│   │                         #   CollaborationView,
│   │                         #   HandoffPoint types
│   ├── team_test.go
│   ├── handoffs.go           # Handoff detection, artifact
│   │                         #   flow mapping, coverage gaps
│   └── handoffs_test.go
├── cli/
│   ├── team_prompt.go        # CLI flow: configure team,
│   │                         #   add members, display view
│   ├── team_prompt_test.go
│   └── styles.go             # Update: add collaboration
│                             #   view rendering styles
├── consts/
│   └── consts.go             # Update: add team config
│                             #   constants, artifact-layer
│                             #   associations
└── session/
    └── session.go            # Update: add team config
                              #   to session state
```

## Implementation Phases

### Phase 1: Team and Collaboration Data Model (FR-008)

- Define the `TeamMember` type: role name, layer mappings,
  activity keywords, source (predefined or custom profile).
  This wraps an existing `Role` or `RoleProfile` with the
  member's resolved layer mappings.
- Define the `TeamConfig` type: team name, list of
  `TeamMember` entries. Methods for adding and removing
  members.
- Define the `CollaborationView` type: role-to-layer
  mappings (map of member name to layers), handoff points
  (list of `HandoffPoint`), coverage gaps (layers with no
  assigned member).
- Define the `HandoffPoint` type: producing role, consuming
  role, layer boundary (producer layer, consumer layer),
  artifact types that flow across the boundary, tutorial
  references for both roles.
- Define the `ArtifactFlow` type: artifact type name
  (from consts), source layer, target layer, description.
- Add layer-to-artifact associations to `internal/consts/`:
  - Layer 1: GuidanceCatalog
  - Layer 2: ThreatCatalog, ControlCatalog
  - Layer 3: Policy
  - Layer 1-3: MappingDocument (cross-layer)
  - Layer 5: EvaluationLog
- Tests: TeamConfig construction, add/remove members,
  member layer mappings derived from Role defaults.

### Phase 2: Handoff Detection and Artifact Flow

- Implement handoff detection logic: for each pair of team
  members, determine if one member's layers produce
  artifacts that another member's layers consume. Handoffs
  occur at layer boundaries where:
  - Member A owns Layer N and Member B owns Layer N+1, or
  - Member A owns a layer that produces an artifact type
    consumed by a layer owned by Member B.
- Implement artifact flow mapping: map each handoff point
  to the specific Gemara artifact types that cross the
  boundary (e.g., ControlCatalog flows from L2 to L3 when
  Policy references controls).
- Implement coverage gap detection: compare the union of
  all team members' layers against layers 1-7, report
  unassigned layers.
- Implement tutorial linking: for each handoff point,
  identify relevant tutorials for both the producing and
  consuming roles using the existing tutorial index.
- Define artifact flow relationships:
  - L1 -> L2: GuidanceCatalog informs ThreatCatalog scope
  - L1 -> L3: GuidanceCatalog referenced by Policy
  - L2 -> L3: ControlCatalog/ThreatCatalog referenced by
    Policy evaluation criteria
  - L2 -> L4: ControlCatalog defines controls for
    sensitive activities
  - L3 -> L5: Policy drives EvaluationLog assessments
  - L1-L3: MappingDocument maps across guidance, controls,
    and policy layers
- Tests: handoff detection for the spec's example team
  (Security Engineer + Compliance Officer + Developer),
  correct artifact flow mapping, coverage gap detection
  (e.g., no team member owns L6-L7), adding a new member
  updates handoffs.

### Phase 3: Team Configuration Persistence

- Implement team configuration save/load as YAML files in
  `~/.config/pacman/teams/`.
- `SaveTeam(dir string, team *TeamConfig) error` — write
  YAML file named after team name.
- `LoadTeam(path string) (*TeamConfig, error)` — read and
  parse YAML.
- `ListTeams(dir string) ([]TeamConfig, error)` — scan
  directory and return all saved teams.
- Tests: save/load round-trip, list teams, handle missing
  directory.

### Phase 4: CLI Integration and TUI Rendering

- Create `internal/cli/team_prompt.go`:
  - `TeamPromptConfig` struct: Prompter, TutorialsDir,
    SchemaVersion, team config dir path.
  - `TeamPromptResult` struct: TeamConfig,
    CollaborationView.
  - `RunTeamSetup(cfg *TeamPromptConfig,
    out io.Writer) (*TeamPromptResult, error)` — prompt
    user for team name, add members by selecting from
    predefined roles and custom profiles, generate
    collaboration view, display results.
  - `RunHandoffInspection(view *CollaborationView,
    handoffIdx int, out io.Writer) error` — display
    detailed artifact flow and tutorial links for a
    specific handoff point (acceptance scenario 2).
- Add rendering functions to `internal/cli/styles.go`:
  - `RenderCollaborationView(view *CollaborationView,
    out io.Writer)` — table or structured layout showing
    each team member, their layers, and handoff indicators.
  - `RenderHandoffPoint(hp *HandoffPoint) string` — card
    with producing/consuming roles, artifact types, and
    tutorial links.
  - `RenderCoverageGaps(gaps []int) string` — warning
    display for unassigned layers.
  - `RenderTeamMember(name string, layers []int) string` —
    member card with role name and layer badges.
- Update session state: add optional `TeamConfig` field.
- Wire into the CLI: after individual role discovery, offer
  the option to configure a team collaboration view.
- Integration test: full flow with 3-role team, verify
  correct handoff points, verify adding a 4th role updates
  the view.
- Verify `make build`, `make test`, `make lint` pass.

## Dependencies & Execution Order

```text
Phase 1 (Data Model)
        │
        ▼
Phase 2 (Handoff Detection)
        │
        ▼
Phase 3 (Persistence)
        │
        ▼
Phase 4 (CLI Integration & TUI)
```

All phases are sequential. Each phase depends on the
preceding one.

## Complexity Tracking

No constitution violations identified. No complexity
justifications required.
