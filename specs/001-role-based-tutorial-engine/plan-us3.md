# Implementation Plan: US3 — Role and Activity Discovery
# with Tailored Learning Path (P3)

**Branch**: `001-role-based-tutorial-engine` | **Date**: 2026-03-13
**Spec**: [spec.md](spec.md) — User Story 3
**Depends on**: US1 (MCP Server Setup), US2 (Schema Version
Selection) — both completed

## Summary

US3 implements the two-phase role discovery process (role
identification + activity probing) and generates tailored
learning paths from the Gemara tutorials. Users select or
describe their role, then describe their daily activities.
The system extracts domain keywords, maps them to Gemara
layers, and produces an ordered learning path where each step
includes "Why this matters," "How you will use this," and
"What you will learn" annotations tailored to the user's
stated activities.

This plan covers FR-001, FR-002, FR-003, FR-004, FR-007,
FR-011, FR-014, FR-015, FR-021, FR-022, FR-023, FR-024,
FR-025, and the US3 acceptance scenarios 1-10.

## Technical Context

**Language/Version**: Go 1.26.1
**Dependencies**: Existing `internal/session/` (schema version,
MCP status), `internal/consts/` (centralized constants),
`internal/cli/` (TUI styles, prompter interface),
`internal/fallback/` (lexicon for term consistency)
**Storage**: Local filesystem — role profiles stored as YAML in
`~/.config/pacman/roles/`; tutorial content read from
configurable directory (default: `~/github/openssf/gemara/
gemara/docs/tutorials`)
**Testing**: `go test ./...` via `make test`; TDD per
constitution
**Constraints**: Keyword-based matching (no ML); extensible
through configuration data; same role title with different
activities MUST produce different paths

## Constitution Check

| Principle | Status | Notes |
|:---|:---|:---|
| I. Schema Conformance | Pass | Learning paths reference tutorials by Gemara layer; no schema output produced |
| II. Gemara Layer Fidelity | Pass | Keyword-to-layer mapping respects the seven-layer model boundaries |
| III. TDD | Pass | Tests written before implementation per phase |
| IV. Tutorial-First Design | Pass | US3 is the tutorial routing engine itself |
| V. Incremental Delivery | Pass | US3 is independently usable after US1+US2 |
| VI. Decision Documentation | N/A | No new ADRs anticipated; existing patterns sufficient |
| VII. Centralized Constants | Pass | Roles, keywords, layer mappings in `internal/consts/` or config data |
| VIII. Composability | Pass | Role discovery is an independent CLI subcommand |
| IX. Convention Over Configuration | Pass | Predefined roles and keywords work with zero config |

## Source Code

```text
internal/
├── roles/
│   ├── roles.go              # Predefined role definitions,
│   │                         #   role matching, custom roles
│   ├── roles_test.go
│   ├── activities.go         # Activity probing: keyword
│   │                         #   extraction, layer mapping
│   ├── activities_test.go
│   ├── profiles.go           # Custom role profile CRUD
│   │                         #   (save, load, list, delete)
│   └── profiles_test.go
├── tutorials/
│   ├── loader.go             # Load tutorial metadata from
│   │                         #   the tutorials directory
│   ├── loader_test.go
│   ├── path.go               # Learning path generation:
│   │                         #   order, annotate, layer-map
│   └── path_test.go
├── cli/
│   ├── role_prompt.go        # CLI flow: role selection,
│   │                         #   activity probing, path
│   │                         #   display
│   └── role_prompt_test.go
├── consts/
│   └── consts.go             # Update: add predefined roles,
│                             #   keyword-to-layer map,
│                             #   tutorial dir, role config
│                             #   constants
└── session/
    └── session.go            # Update: add RoleProfile and
                              #   LearningPath fields
```

## Implementation Phases

### Phase 1: Role and Activity Data Model (FR-002, FR-022,
FR-025)

- Define the `Role` type: name, description, default activity
  keywords, default Gemara layer mappings ([]int, layers 1-7),
  source (predefined or custom).
- Define the `ActivityProfile` type: extracted keywords,
  matched activity categories, resolved layer mappings,
  user description (free-text), confidence indicators.
- Define the `KeywordMapping` type: maps domain keywords to
  Gemara layers, extensible through configuration.
- Add predefined roles to `internal/consts/consts.go`:
  Security Engineer, Compliance Officer, CISO/Security Leader,
  Developer, Platform Engineer, Policy Author, Auditor.
- Add the keyword-to-layer mapping per FR-022:
  - Layer 1 (Guidance): regulatory framework mapping, EU CRA,
    NIST, OWASP, HIPAA, GDPR, PCI, ISO, machine-readable
    format, standards, best practices, codify, formalize,
    evidence collection (when defining requirements).
  - Layer 2 (Threats & Controls): SDLC, threat modeling,
    penetration testing, secure architecture review, CI/CD,
    dependency management, upstream open-source, custom
    controls, OSPS Baseline, FINOS CCC.
  - Layer 3 (Risk & Policy): create policy, timeline for
    adherence, scope definition, audit interviews, assessment
    plans, adherence requirements, risk appetite,
    non-compliance handling, evidence collection (when
    operationalizing).
- Tests: Role type construction, keyword extraction from
  free-text, layer mapping resolution, partial role matching,
  ambiguous keywords spanning layers 1 and 3.

### Phase 2: Tutorial Loader (FR-004)

- `internal/tutorials/loader.go`: Scan the configurable
  tutorials directory, parse tutorial metadata (title, layer,
  sections, schema version references), return structured
  tutorial index.
- Handle edge cases: directory empty, directory inaccessible,
  tutorials referencing schemas unavailable in selected
  version.
- Tests: Load from valid directory, handle empty directory,
  handle missing directory, detect schema version mismatches.

### Phase 3: Role Identification Flow (FR-001, FR-002, FR-014,
FR-023)

- `internal/roles/roles.go`: Role matching logic.
  - `PredefinedRoles() []Role` — return the predefined list.
  - `MatchRole(input string) (*Role, MatchResult)` — extract
    keywords from free-text role input, identify partial
    matches against predefined roles, return match with
    confidence. Partial matches (e.g., "Product Security
    Engineer" contains "Security Engineer") MUST NOT assume
    the generic role — proceed to activity probing.
- `internal/cli/role_prompt.go`: Phase 1 of the CLI flow.
  - Present predefined role list + "My role isn't listed."
  - If predefined selected, proceed with that role.
  - If "My role isn't listed," accept free-text, run
    MatchRole, present partial matches for confirmation.
- Tests: Select predefined role, enter custom role with
  partial match, enter completely unknown role, keyword
  extraction accuracy.

### Phase 4: Activity Probing and Layer Resolution (FR-007,
FR-021, FR-022, FR-023)

- `internal/roles/activities.go`: Activity probing logic.
  - `ExtractKeywords(description string) []string` — extract
    domain keywords from free-text using the keyword-to-layer
    vocabulary.
  - `ResolveLayerMappings(role *Role, keywords []string)
    *ActivityProfile` — combine role defaults with extracted
    keywords to determine which Gemara layers are relevant.
    When keywords match both Layer 1 and Layer 3 (e.g.,
    "evidence collection"), ask a clarifying follow-up
    (FR-022).
  - `ClarificationNeeded(keywords []string) []string` —
    identify ambiguous keywords requiring follow-up.
- `internal/cli/role_prompt.go`: Phase 2 of the CLI flow.
  - Ask user to describe activities or select from categories.
  - Extract keywords, resolve layers, handle ambiguous
    keywords with clarifying questions.
  - If no recognizable keywords, present full activity
    category list for manual selection.
- Tests per acceptance scenarios 1, 2, 3, 4, 8, 9, 10:
  same role + different activities = different paths; unknown
  keywords trigger category selection; regulatory framework
  mapping routes to Layer 1; policy creation routes to
  Layer 3.

### Phase 5: Learning Path Generation (FR-003, FR-015)

- `internal/tutorials/path.go`: Build the learning path.
  - `GeneratePath(profile *ActivityProfile,
    tutorials []Tutorial, schemaVersion string) *LearningPath`
    — sequence tutorials by relevance to the user's resolved
    layers, prioritized by the user's stated activities.
  - Each `PathStep` includes: tutorial reference (file path),
    layer, "Why this matters for your role" annotation,
    "How you will use this" annotation, "What you will learn"
    annotation, prerequisites, completion status.
  - Annotations MUST be tailored to the user's stated
    activities, not generic role text.
  - Support non-linear navigation: any step accessible
    regardless of completion, with prerequisite notes
    (FR-015).
  - Detect schema version mismatches between tutorial content
    and selected version; flag discrepancies.
  - Handle missing tutorials for layers 6-7 (informative
    message + model documentation fallback).
- Tests per acceptance scenarios 5, 6:
  every path step has why/how/what annotations; non-linear
  navigation displays prerequisite notes.

### Phase 6: Custom Role Profiles (FR-024)

- `internal/roles/profiles.go`: CRUD for custom role profiles.
  - `SaveProfile(path string, profile *RoleProfile) error`
  - `LoadProfile(path string) (*RoleProfile, error)`
  - `ListProfiles(dir string) ([]RoleProfile, error)`
  - Custom profiles include: role name, activity keywords,
    layer mappings, optional description.
  - Saved profiles available in future sessions and in the
    predefined role selection list.
- Tests per acceptance scenario 7:
  save profile, load profile, profile appears in role list.

### Phase 7: CLI Integration and Polish

- Wire role prompt into the setup flow: after version
  selection (US2), before any further operations.
- Update `internal/cli/setup.go` to call the role discovery
  flow.
- Update `internal/session/session.go` to store the resolved
  `ActivityProfile` and `LearningPath`.
- Styled TUI output using existing lipgloss styles for the
  learning path display.
- Integration test: full flow from MCP setup -> version
  selection -> role discovery -> learning path output.
- Verify `make build`, `make test`, `make lint` pass.

## Dependencies & Execution Order

```text
Phase 1 (Data Model)
    │
    ├───────────────────┐
    ▼                   ▼
Phase 2 (Tutorial)   Phase 3 (Role ID)
    │                   │
    └───────┬───────────┘
            ▼
    Phase 4 (Activity Probing)
            │
            ▼
    Phase 5 (Learning Path)
            │
            ▼
    Phase 6 (Custom Profiles)
            │
            ▼
    Phase 7 (Integration)
```

- Phases 2 and 3 can run in parallel after Phase 1.
- Phase 4 depends on both Phase 2 (tutorial awareness) and
  Phase 3 (role identification).
- Phase 5 depends on Phase 4 (resolved activity profile).
- Phase 6 depends on Phase 5 (complete path to persist).
- Phase 7 integrates all preceding work.

## Complexity Tracking

No constitution violations identified. No complexity
justifications required.
