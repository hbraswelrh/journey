# Tasks: US3 — Role and Activity Discovery with Tailored
# Learning Path (P3)

**Input**: `plan-us3.md`, `spec.md` (User Story 3)
**Prerequisites**: US1 completed (MCP setup), US2 completed
(schema version selection)

---

## Phase 1: Role and Activity Data Model (FR-002, FR-022,
FR-025)

### Tests

- [x] T201 [P] [US3] Write test
  `internal/roles/roles_test.go`: PredefinedRoles returns
  the required minimum list (Security Engineer, Compliance
  Officer, CISO/Security Leader, Developer, Platform
  Engineer, Policy Author, Auditor)
- [x] T202 [P] [US3] Write test
  `internal/roles/roles_test.go`: Role struct contains
  required fields (name, description, default keywords,
  default layer mappings, source)
- [x] T203 [P] [US3] Write test
  `internal/roles/activities_test.go`: ExtractKeywords
  extracts known domain terms from free-text description
  (e.g., "CI/CD pipeline management" yields "CI/CD")
- [x] T204 [P] [US3] Write test
  `internal/roles/activities_test.go`: ExtractKeywords
  returns empty slice for text with no recognizable
  domain keywords
- [x] T205 [P] [US3] Write test
  `internal/roles/activities_test.go`: KeywordMapping maps
  Layer 1 keywords (EU CRA, NIST, best practices,
  machine-readable format) correctly
- [x] T206 [P] [US3] Write test
  `internal/roles/activities_test.go`: KeywordMapping maps
  Layer 2 keywords (SDLC, threat modeling, CI/CD, OSPS
  Baseline, FINOS CCC) correctly
- [x] T207 [P] [US3] Write test
  `internal/roles/activities_test.go`: KeywordMapping maps
  Layer 3 keywords (create policy, timeline for adherence,
  risk appetite, audit interviews) correctly
- [x] T208 [P] [US3] Write test
  `internal/roles/activities_test.go`: Ambiguous keywords
  spanning Layers 1 and 3 (evidence collection) are
  identified by ClarificationNeeded

### Implementation

- [x] T209 [US3] Add predefined role constants and
  keyword-to-layer mapping to `internal/consts/consts.go`:
  role names, descriptions, default keywords per role,
  default layer mappings per role, domain keyword vocabulary
  with layer associations
- [x] T210 [US3] Implement `internal/roles/roles.go`:
  `Role` struct, `ActivityProfile` struct,
  `KeywordMapping` type, `PredefinedRoles() []Role` function
- [x] T211 [US3] Implement `internal/roles/activities.go`:
  `ExtractKeywords(description string) []string`,
  `ClarificationNeeded(keywords []string) []string`,
  `ResolveLayerMappings(role *Role, keywords []string)
  *ActivityProfile`

**Checkpoint**: Data model compiles. Role and keyword types are
defined. ExtractKeywords identifies domain terms from free-text.

---

## Phase 2: Tutorial Loader (FR-004)

### Tests

- [x] T212 [P] [US3] Write test
  `internal/tutorials/loader_test.go`: LoadTutorials from
  valid directory returns structured tutorial index with
  titles, layers, and section headings
- [x] T213 [P] [US3] Write test
  `internal/tutorials/loader_test.go`: LoadTutorials from
  empty directory returns empty list with no error
- [x] T214 [P] [US3] Write test
  `internal/tutorials/loader_test.go`: LoadTutorials from
  nonexistent directory returns informative error with
  expected path and resolution guidance
- [x] T215 [P] [US3] Write test
  `internal/tutorials/loader_test.go`: LoadTutorials detects
  tutorials referencing schemas unavailable in the selected
  version and flags them

### Implementation

- [x] T216 [US3] Create test fixtures in `testdata/tutorials/`:
  sample tutorial files with front matter (title, layer,
  schema version), valid structure, and an empty directory
  variant
- [x] T217 [US3] Implement `internal/tutorials/loader.go`:
  `Tutorial` struct (title, file path, layer, sections,
  schema version references).
  `LoadTutorials(dir string) ([]Tutorial, error)` — scan
  directory, parse tutorial metadata from front matter or
  heading structure, return indexed list.
  `CheckVersionCompat(tutorials []Tutorial,
  selectedVersion string) []VersionMismatch` — identify
  tutorials whose schema references differ from the
  selected version

**Checkpoint**: Tutorial loader reads from a configurable
directory. Schema version mismatches are detected.

---

## Phase 3: Role Identification Flow (FR-001, FR-002, FR-014,
FR-023)

### Tests

- [x] T218 [P] [US3] Write test
  `internal/roles/roles_test.go`: MatchRole with exact
  predefined name returns the role with high confidence
- [x] T219 [P] [US3] Write test
  `internal/roles/roles_test.go`: MatchRole with partial
  match ("Product Security Engineer" contains "Security
  Engineer") returns partial match result — does NOT assume
  the generic role
- [x] T220 [P] [US3] Write test
  `internal/roles/roles_test.go`: MatchRole with completely
  unknown title returns no match
- [x] T221 [P] [US3] Write test
  `internal/cli/role_prompt_test.go`: Selecting a predefined
  role from the list proceeds to activity probing with that
  role
- [x] T222 [P] [US3] Write test
  `internal/cli/role_prompt_test.go`: Selecting "My role
  isn't listed" accepts free-text input and shows partial
  matches for confirmation
- [x] T223 [P] [US3] Write test
  `internal/cli/role_prompt_test.go`: Entering a custom role
  with no partial match proceeds to activity probing with
  extracted keywords only

### Implementation

- [x] T224 [US3] Implement `internal/roles/roles.go`:
  `MatchRole(input string) (*Role, MatchResult)` — extract
  keywords from free-text role input, compare against
  predefined role names and keywords, return match type
  (exact, partial, none) with confidence.
  `MatchResult` struct: matched role (if any), match type,
  overlapping keywords, confidence level
- [x] T225 [US3] Implement `internal/cli/role_prompt.go`:
  Phase 1 flow — present predefined role list plus "My role
  isn't listed" option. Handle free-text input, call
  MatchRole, present partial matches for user confirmation.
  Uses existing `UserPrompter` interface and lipgloss styles

**Checkpoint**: Users can select a predefined role or enter a
custom role. Partial matches are identified and presented.

---

## Phase 4: Activity Probing and Layer Resolution (FR-007,
FR-021, FR-022, FR-023)

### Tests

- [x] T226 [P] [US3] Write test
  `internal/roles/activities_test.go`: Security Engineer +
  "CI/CD pipeline management, dependency management, coding
  with upstream open-source components" resolves to Layer 2
  (Threats & Controls) emphasis
- [x] T227 [P] [US3] Write test
  `internal/roles/activities_test.go`: Security Engineer +
  "evidence collection, audit interviews, defining
  compliance scope" resolves to Layers 1 and 3 emphasis
- [x] T228 [P] [US3] Write test
  `internal/roles/activities_test.go`: Same role title with
  different activities produces different layer mappings
- [x] T229 [P] [US3] Write test
  `internal/roles/activities_test.go`: "map my best
  practices to the EU CRA" extracts keywords "best
  practices," "map," "EU CRA" and routes to Layer 1
- [x] T230 [P] [US3] Write test
  `internal/roles/activities_test.go`: "create a reusable
  machine-readable format for my internal standards" routes
  to Layer 1
- [x] T231 [P] [US3] Write test
  `internal/roles/activities_test.go`: "create a policy and
  define a timeline for adherence" routes to Layer 3
- [x] T232 [P] [US3] Write test
  `internal/roles/activities_test.go`: Ambiguous keyword
  "evidence collection" triggers clarifying follow-up
  between Layers 1 and 3
- [x] T233 [P] [US3] Write test
  `internal/cli/role_prompt_test.go`: No recognizable
  keywords in activity description presents full activity
  category list for manual selection
- [x] T234 [P] [US3] Write test
  `internal/cli/role_prompt_test.go`: "Secure Software
  Development professional" role extracts "SDLC" keyword
  and presents relevant activity categories

### Implementation

- [x] T235 [US3] Implement `internal/roles/activities.go`:
  complete activity probing logic — `ExtractKeywords`
  handles full sentences and multi-word domain terms,
  `ResolveLayerMappings` combines role defaults with
  keyword-resolved layers to produce a unified
  `ActivityProfile`, `ClarificationNeeded` returns
  ambiguous keywords (those matching multiple layers)
- [x] T236 [US3] Implement `internal/cli/role_prompt.go`:
  Phase 2 flow — ask user to describe activities or select
  from categories, call ExtractKeywords, handle ambiguous
  keywords with clarifying questions via the prompter,
  present resolved layer mappings for confirmation. If no
  keywords recognized, display full category list.
  Activity categories: Regulatory Compliance (L1), Threat
  & Control Authoring (L2), Policy & Risk (L3), Secure
  Development (L2/L4), Evaluation & Audit (L5)

**Checkpoint**: Activity probing correctly differentiates same
role with different activities. Ambiguous keywords handled.

---

## Phase 5: Learning Path Generation (FR-003, FR-015)

### Tests

- [x] T237 [P] [US3] Write test
  `internal/tutorials/path_test.go`: GeneratePath produces
  an ordered list of PathSteps based on the ActivityProfile
  layer mappings
- [x] T238 [P] [US3] Write test
  `internal/tutorials/path_test.go`: Every PathStep has
  non-empty why, how, and what annotations tailored to the
  user's stated activities (not generic role text)
- [x] T239 [P] [US3] Write test
  `internal/tutorials/path_test.go`: Learning path for
  Security Engineer (CI/CD focus) starts with Threat
  Assessment Guide and Control Catalog Guide (Layer 2)
- [x] T240 [P] [US3] Write test
  `internal/tutorials/path_test.go`: Learning path for
  Security Engineer (audit focus) starts with Guidance
  Catalog Guide (Layer 1) and Policy Guide (Layer 3)
- [x] T241 [P] [US3] Write test
  `internal/tutorials/path_test.go`: Non-linear navigation —
  accessing a later step without completing earlier ones
  shows prerequisite note
- [x] T242 [P] [US3] Write test
  `internal/tutorials/path_test.go`: Activities spanning
  multiple layers produce a combined learning path ordered
  by user-stated priority
- [x] T243 [P] [US3] Write test
  `internal/tutorials/path_test.go`: Layers with no
  tutorials (e.g., Layers 6-7) produce informative message
  with fallback to model documentation
- [x] T244 [P] [US3] Write test
  `internal/tutorials/path_test.go`: Schema version mismatch
  between tutorial and selected version is flagged in path
  step

### Implementation

- [x] T245 [US3] Implement `internal/tutorials/path.go`:
  `LearningPath` struct (target role, ordered steps,
  completion map).
  `PathStep` struct (tutorial reference, layer, why/how/what
  annotations, prerequisites, completion status).
  `GeneratePath(profile *ActivityProfile,
  tutorials []Tutorial, schemaVersion string) *LearningPath`
  — sequence tutorials by relevance to resolved layers,
  generate tailored annotations, detect schema mismatches,
  handle missing-tutorial layers.
  `StepStatus(path *LearningPath, stepIdx int)
  *StepNavInfo` — return step info including prerequisite
  warnings for non-linear navigation

**Checkpoint**: Learning paths are correctly differentiated by
activity. Annotations are tailored. Non-linear navigation
supported.

---

## Phase 6: Custom Role Profiles (FR-024)

### Tests

- [x] T246 [P] [US3] Write test
  `internal/roles/profiles_test.go`: SaveProfile writes a
  valid YAML file with role name, keywords, layer mappings,
  and description
- [x] T247 [P] [US3] Write test
  `internal/roles/profiles_test.go`: LoadProfile reads a
  saved profile and returns the correct RoleProfile struct
- [x] T248 [P] [US3] Write test
  `internal/roles/profiles_test.go`: ListProfiles returns
  all saved profiles from the profiles directory
- [x] T249 [P] [US3] Write test
  `internal/roles/profiles_test.go`: Saved custom profiles
  appear in the role selection list alongside predefined
  roles
- [x] T250 [P] [US3] Write test
  `internal/cli/role_prompt_test.go`: After completing
  discovery, user can save their profile as a custom role
  for future reuse

### Implementation

- [x] T251 [US3] Add profile directory and file constants to
  `internal/consts/consts.go`: profile directory path
  (`~/.config/gemara-user-journey/roles/`), profile file extension
  (`.yaml`)
- [x] T252 [US3] Implement `internal/roles/profiles.go`:
  `RoleProfile` struct (role name, activity keywords, layer
  mappings, description, created timestamp).
  `SaveProfile(dir string, profile *RoleProfile) error` —
  write YAML file to profiles directory.
  `LoadProfile(path string) (*RoleProfile, error)` — read
  and parse YAML profile.
  `ListProfiles(dir string) ([]RoleProfile, error)` — scan
  directory and return all saved profiles.
  `MergeWithPredefined(predefined []Role,
  custom []RoleProfile) []Role` — combine predefined and
  custom roles into a unified selection list
- [x] T253 [US3] Update `internal/cli/role_prompt.go`:
  after learning path generation, offer to save the profile.
  On role selection, include saved custom profiles in the
  list via MergeWithPredefined

**Checkpoint**: Custom profiles persist across sessions. Saved
profiles appear in the role selection list.

---

## Phase 7: CLI Integration and Polish

- [x] T254 [US3] Update `internal/session/session.go`:
  add `RoleProfile *ActivityProfile` and `LearningPath
  *LearningPath` fields to the `Session` struct with
  appropriate getter/setter methods
- [x] T255 [US3] Update `internal/cli/setup.go`:
  call role discovery flow after version selection completes,
  passing the session from SetupResult. Wire the full flow:
  MCP setup -> version selection -> role discovery
- [x] T256 [US3] Implement learning path TUI display in
  `internal/cli/role_prompt.go`: styled output using
  existing lipgloss styles — render the learning path with
  numbered steps, layer badges, and why/how/what sections.
  Use panelStyle for path overview and headingStyle for
  section headers
- [x] T257 [US3] Integration test: full flow from MCP setup
  through version selection through role discovery with
  "Security Engineer" + CI/CD activities — verify session
  has correct ActivityProfile and LearningPath targeting
  Layer 2
- [x] T258 [US3] Integration test: full flow with custom role
  "Product Security Engineer" + audit activities — verify
  partial match identification, activity probing, and
  Layer 1/3 learning path
- [x] T259 [US3] Verify all CLI help text and error messages
  use Gemara lexicon terms consistently (FR-011)
- [x] T260 [US3] Verify `make build`, `make test`, `make lint`
  pass with zero errors and zero warnings

**Checkpoint**: US3 is fully functional. A user can launch
Gemara User Journey, complete MCP setup, select a schema version, identify
their role and activities, and receive a tailored learning
path. Two users with the same title but different activities
receive different paths.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Data Model)**: No US3-internal dependencies —
  start immediately (depends on US1+US2 being complete)
- **Phase 2 (Tutorial Loader)**: Depends on Phase 1 (Tutorial
  struct uses layer types from data model)
- **Phase 3 (Role ID)**: Depends on Phase 1 (uses Role and
  MatchResult types). Can run in **parallel** with Phase 2
- **Phase 4 (Activity Probing)**: Depends on Phases 2 and 3
  (needs both tutorial awareness and role identification)
- **Phase 5 (Learning Path)**: Depends on Phase 4 (needs
  resolved ActivityProfile and loaded tutorials)
- **Phase 6 (Custom Profiles)**: Depends on Phase 5 (needs
  complete path to persist)
- **Phase 7 (Integration)**: Depends on all preceding phases

### Parallel Opportunities

```text
Phase 1 (Data Model)
    │
    ├───────────────────┐
    ▼                   ▼
Phase 2 (Tutorial)   Phase 3 (Role ID)    ← parallel
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
- US3 tasks numbered T201-T260 to avoid conflicts with US1
  (T001-T054) and US2 (T101-T127).
- Key invariant: same role title + different activities MUST
  produce different learning paths (SC-002).
