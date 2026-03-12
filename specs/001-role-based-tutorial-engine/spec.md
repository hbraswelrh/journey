# Feature Specification: Role-Based Tutorial Engine

**Feature Branch**: `001-role-based-tutorial-engine`
**Created**: 2026-03-12
**Status**: Draft
**Input**: User description: "Build an application that processes user input of job role for adapting and tailoring the tutorials in ~/github/openssf/gemara/gemara/docs/tutorials and will help users effectively author gemara content and understand the 'why' and 'how' alongside the 'what.' I want to build a technology-agnostic tool that can enable cross-functional teams to collaborate and learn from each other how to use the Gemara project, why they would use it, and what it can do for them. The goal is to allow for directed tailored tutorials to orchestrate communication that will immediately determine based on a job role or person what they need to know and what is the best path to move forward for the collaborative effort. Anything that is written as a standard, best practice, or guidance/guideline should be transformed into something that can be reused and adjusted as advancements occur."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Gemara MCP Server Setup (Priority: P1)

When the user first launches Pac-Man, the system offers the
option to install and configure the Gemara MCP server
([github.com/gemaraproj/gemara-mcp](https://github.com/gemaraproj/gemara-mcp))
as an enhanced authoring and learning companion. The MCP server
provides direct access to three tools that augment Pac-Man's
capabilities throughout all subsequent operations:

- **get_lexicon**: Retrieve the upstream Gemara lexicon entries,
  ensuring all terminology used by Pac-Man and the user's
  authored content aligns with the canonical Gemara vocabulary.
- **validate_gemara_artifact**: Validate YAML artifacts against
  Gemara schema definitions without requiring the user to
  install CUE locally or run `cue vet` manually.
- **get_schema_docs**: Retrieve schema documentation for the
  Gemara CUE module, providing contextual reference material
  during guided authoring and learning paths.

The user may install the MCP server via a pre-built binary or
Docker. If the user declines installation, Pac-Man MUST continue
to function using local CUE tooling for validation and bundled
lexicon data, but the system MUST inform the user which enhanced
capabilities are unavailable without the MCP server and offer
installation again at any point during the session.

**Why this priority**: The MCP server provides the upstream
lexicon, live schema validation, and schema documentation that
every other feature benefits from. Offering it first ensures that
users who install it get the richest experience from the start —
the lexicon feeds into learning paths (US3), the validator
supports guided authoring (US6), and the schema docs provide
contextual help throughout. Users who skip installation still
have a functional tool but with degraded capabilities.

**Independent Test**: Can be fully tested by launching Pac-Man
with the MCP server not installed, verifying the installation
prompt appears, completing installation via binary or Docker,
and confirming that the three MCP tools respond correctly. Also
testable by declining installation and verifying that all
features still function with local fallbacks.

**Acceptance Scenarios**:

1. **Given** a user launches Pac-Man for the first time and the
   Gemara MCP server is not detected, **When** the initial setup
   screen is presented, **Then** the system offers to install the
   Gemara MCP server with a brief explanation of the three tools
   it provides and how they enhance the Pac-Man experience.

2. **Given** a user chooses to install the MCP server via binary,
   **When** the system guides them through installation, **Then**
   it provides platform-appropriate instructions (Linux or
   macOS), verifies the binary is accessible, configures the MCP
   client connection, and confirms the server responds to a
   health check.

3. **Given** a user chooses to install the MCP server via Docker,
   **When** the system guides them through installation, **Then**
   it provides the Docker run configuration, verifies the
   container starts, and confirms the server responds to a
   health check.

4. **Given** a user declines MCP server installation, **When**
   they proceed to use Pac-Man, **Then** the system informs them
   that lexicon lookups will use bundled data (which may not
   reflect the latest upstream terms), validation will require
   local CUE tooling, and schema documentation will be limited
   to locally cached content. All features remain functional.

5. **Given** a user previously declined MCP server installation,
   **When** they later request enhanced capabilities (e.g.,
   attempt guided authoring or lexicon lookup), **Then** the
   system offers the MCP server installation option again with
   context on how it would improve the current operation.

6. **Given** the MCP server is installed and running, **When**
   the system initializes a session, **Then** it queries
   `get_lexicon` to load the current upstream lexicon,
   `get_schema_docs` to cache schema documentation, and
   confirms `validate_gemara_artifact` is available for
   on-demand validation throughout the session.

---

### User Story 2 - Upstream Schema Fetch and Version Selection (Priority: P2)

When the user launches Pac-Man or initiates any operation that
depends on the Gemara schemas, the system fetches the latest
schema information from the upstream Gemara repository
(`github.com/gemaraproj/gemara`). It retrieves the list of
available tagged releases and determines which schemas are marked
as Stable versus Experimental. The system then prompts the user
to choose whether to work against the latest stable schema
version (the most recent tagged release where the schemas they
need are marked `@status(Stable)`) or the latest overall schema
version (the most recent tagged release, which may include
Experimental schemas). The user's choice determines which schema
version is used for all validation, tutorial content alignment,
and guided authoring throughout the session. If the Gemara MCP
server is installed (per US1), the system MAY use the MCP
server's `get_schema_docs` tool to supplement version information
with schema documentation for the selected version.

**Why this priority**: Schema version selection is a prerequisite
for every feature after MCP setup. Learning paths, content
blocks, and authored artifacts all depend on knowing which schema
version the user is targeting. Without this, the tool cannot
guarantee that its guidance and validation are aligned with the
user's intended version.

**Independent Test**: Can be fully tested by running the system
with network access to the upstream repository and verifying that
it presents the correct stable and latest versions, records the
user's selection, and uses the selected version for subsequent
validation commands.

**Acceptance Scenarios**:

1. **Given** the upstream Gemara repository has releases from
   v0.1.0 through v0.20.0, **When** the user launches the tool,
   **Then** the system fetches the available releases and
   presents the user with a choice: "Stable" (identifying the
   most recent release where base, metadata, and mapping_inline
   schemas are `@status(Stable)`) or "Latest" (identifying the
   most recent release tag, e.g., v0.20.0).

2. **Given** the user selects "Stable," **When** the system
   configures the session, **Then** all subsequent schema
   validation uses the selected stable version, and the system
   displays a notice identifying which layer schemas are
   Experimental at that version and may behave differently from
   the latest release.

3. **Given** the user selects "Latest," **When** the system
   configures the session, **Then** the system displays a
   warning that some schemas at this version are marked
   Experimental and may have breaking changes in future
   releases, and proceeds to use the latest version for all
   operations. If the Gemara MCP server is installed, the
   system also checks whether the installed gemara-mcp version
   is compatible with the selected Gemara schema version and
   warns the user if a mismatch is detected (see FR-031).

4. **Given** the system has previously fetched and cached schema
   version information, **When** the user launches the tool
   again, **Then** the system checks whether new releases are
   available upstream and notifies the user if a newer version
   exists, while defaulting to the previously selected version.

5. **Given** the system cannot reach the upstream repository
   (e.g., no network access), **When** the user launches the
   tool, **Then** the system falls back to the most recently
   cached version information, informs the user that upstream
   could not be reached, and displays the date of the last
   successful fetch.

6. **Given** the Gemara MCP server is installed at a version
   built against Gemara schema v0.18.0, **When** the user
   selects "Latest" and the latest Gemara schema is v0.20.0,
   **Then** the system warns the user that the installed
   gemara-mcp version may produce inaccurate validation
   results or stale lexicon data because it was built against
   an older schema version. The system MUST recommend that the
   user either update their gemara-mcp installation to a
   version compatible with v0.20.0, or select the schema
   version that matches their installed gemara-mcp version.

7. **Given** the Gemara MCP server is installed and the user
   selects "Stable," **When** the stable Gemara schema version
   matches or is older than the version the installed
   gemara-mcp was built against, **Then** the system proceeds
   normally with no compatibility warning.

---

### User Story 3 - Role and Activity Discovery with Tailored Learning Path (Priority: P3)

A user launches Pac-Man and selects their schema version (per
US2). The system then walks them through a two-phase role
discovery process:

**Phase 1 — Role Identification**: The system presents a list of
common roles (e.g., Security Engineer, Compliance Officer, CISO,
Developer, Platform Engineer, Policy Author, Auditor). The list
also includes "My role isn't listed." If the user selects that
option, the system asks "What is your role?" and accepts free-text
input (e.g., "Product Security Engineer," "Compliance Manager,"
"Secure Software Development professional"). The system extracts
keywords from the response and maps them to the closest known
role profile. If the title partially matches an existing role
(e.g., "Product Security Engineer" contains "Security Engineer"),
the system recognizes the overlap but does not assume an exact
match — it proceeds to Phase 2 for refinement.

**Phase 2 — Activity Probing**: Because the same job title can
map to completely different Gemara layers depending on daily
activities, the system asks the user to describe what they
actually do or what problem they are trying to solve. The user
can provide a free-text description (e.g., "All things SDLC on
the Resilient Development team within Product Security") or
select from activity categories. The system extracts domain
keywords from the response — terms like "SDLC," "threat
modeling," "penetration testing," "evidence collection," "audit
interviews," "CI/CD pipeline," "dependency management," "secure
architecture review" — and uses them to determine which Gemara
layers and tutorials are most relevant.

For example:
- A Product Security Engineer who describes "evidence collection
  and audit interviews" and "doesn't typically interact with the
  terminal or leverage git" maps to Layer 3 (Policy — assessment
  plans, scope definition, adherence) and Layer 1 (Guidance —
  requirements that need to be met).
- A Product Security Engineer who describes "CI/CD pipeline
  management, dependency management, coding day-in and day-out,
  and leveraging upstream open-source components" maps to Layer 2
  (Threats and Controls — writing custom threat-informed,
  technology-specific controls, importing external catalogs like
  OSPS Baseline and FINOS CCC).

Based on the combined role + activity profile, the system
generates a tailored learning path that sequences the Gemara
tutorials in the most relevant order. Each step includes *why*
this tutorial matters for the user's specific activities, *how*
they will use the concepts, and *what* the tutorial covers. If
the selected schema version differs from the version assumed by
a tutorial, the system notes discrepancies.

Users can also define and save custom roles for reuse by
themselves or their team.

**Why this priority**: Without role and activity-based routing,
every user faces the same undifferentiated documentation. A
generic "Security Engineer" path is insufficient — the same
title covers fundamentally different work. Activity probing is
what makes the learning path actually useful rather than merely
approximate.

**Independent Test**: Can be tested by providing various
role + activity combinations (including custom roles and
free-text descriptions) and verifying that the system produces
correctly differentiated learning paths. Two users with the same
title but different activity descriptions MUST receive different
paths.

**Acceptance Scenarios**:

1. **Given** a user selects "Security Engineer" from the
   predefined list, **When** the system proceeds to activity
   probing and the user describes "CI/CD pipeline management,
   dependency management, and coding with upstream open-source
   components," **Then** the learning path starts with the
   Threat Assessment Guide (Layer 2) and the Control Catalog
   Guide (Layer 2), emphasizing custom control authoring and
   importing from OSPS Baseline and FINOS CCC catalogs.

2. **Given** a user selects "Security Engineer" from the
   predefined list, **When** the system proceeds to activity
   probing and the user describes "evidence collection, audit
   interviews, and defining compliance scope," **Then** the
   learning path starts with the Guidance Catalog Guide
   (Layer 1) and the Policy Guide (Layer 3), emphasizing
   assessment plans, scope definition, and adherence sections.

3. **Given** a user selects "My role isn't listed" and enters
   "Compliance Manager," **When** the system extracts the
   keyword "Compliance" and proceeds to activity probing,
   **Then** it maps the user to a compliance-oriented profile
   and asks about their specific activities before generating
   a learning path.

4. **Given** a user enters "Secure Software Development
   professional" as their role, **When** the system extracts
   the keyword "SDLC," **Then** it identifies activities
   associated with secure development (best practices, secure
   architecture review, threat modeling, penetration testing,
   pipeline security) and presents relevant activity categories
   for the user to confirm or refine.

5. **Given** a user selects a role and completes activity
   probing, **When** the learning path is displayed, **Then**
   every tutorial reference includes three clearly labeled
   sections: "Why this matters for your role," "How you will
   use this," and "What you will learn," each tailored to the
   user's stated activities rather than generic role text.

6. **Given** a user has been presented a learning path, **When**
   they choose to jump to a later step without completing
   earlier ones, **Then** the system allows navigation to any
   step and displays a note indicating prerequisite knowledge
   that may have been skipped.

7. **Given** a user completes role and activity discovery,
   **When** they choose to save their profile as a custom role,
   **Then** the system stores the role name, activity keywords,
   and layer mappings so the same profile can be reused in
   future sessions or assigned to team members.

8. **Given** a user describes their goal as "map my best
   practices to the EU CRA," **When** the system extracts the
   keywords "best practices," "map," and "EU CRA," **Then** the
   learning path routes to the Guidance Catalog Guide (Layer 1),
   emphasizing how to create a Guidance Catalog typed as
   "Standard" or "Regulation" and how to use Mapping Documents
   to align internal guidelines with external regulatory
   frameworks.

9. **Given** a user describes their goal as "create a reusable
   machine-readable format for my internal standards," **When**
   the system extracts the keywords "machine-readable format"
   and "standards," **Then** the learning path routes to the
   Guidance Catalog Guide (Layer 1), emphasizing how to
   structure guidelines as machine-readable Gemara artifacts
   with proper metadata, families, and cross-references.

10. **Given** a user describes their goal as "create a policy
    and define a timeline for adherence," **When** the system
    extracts the keywords "create a policy" and "timeline for
    adherence," **Then** the learning path routes to the Policy
    Guide (Layer 3), emphasizing the implementation plan
    section (evaluation timeline, enforcement timeline) and
    the adherence section (evaluation methods, enforcement
    methods, non-compliance handling).

---

### User Story 4 - Reusable Content Transformation (Priority: P4)

A user has standards, best practices, or guidelines (either from
the Gemara tutorials or from their own organization) that they
want to transform into reusable, structured content. The system
reads existing tutorial content from the Gemara tutorials
directory and extracts actionable patterns — naming conventions,
validation steps, schema structures, cross-referencing
techniques — into modular, reusable content blocks. These blocks
can be adapted as the Gemara project evolves without rewriting
the entire tutorial.

**Why this priority**: This directly addresses the user's
requirement that "anything written as a standard, best practice,
or guidance/guideline should be transformed into something that
can be reused and adjusted as advancements occur." It creates
the foundation for keeping tutorials evergreen.

**Independent Test**: Can be tested by pointing the system at the
existing Gemara tutorials directory, running the extraction, and
verifying that the output contains modular content blocks with
clear boundaries, metadata, and a mechanism for updating them
when source material changes.

**Acceptance Scenarios**:

1. **Given** the Gemara tutorials directory contains the Threat
   Assessment Guide, **When** the system processes it, **Then**
   it produces reusable content blocks for: scope definition
   pattern, capability identification pattern, threat
   identification pattern, and CUE validation pattern — each
   with metadata identifying the source tutorial and Gemara
   schema version.

2. **Given** a reusable content block was extracted from a
   tutorial at Gemara schema version v0.17.0, **When** the
   upstream tutorial is updated for a newer schema version,
   **Then** the system identifies which content blocks are
   affected and flags them for review.

3. **Given** a user wants to create their own guidance document,
   **When** they request relevant content blocks, **Then** the
   system returns blocks applicable to their stated goal along
   with instructions for adapting each block to their context.

---

### User Story 5 - Cross-Functional Collaboration View (Priority: P5)

A team with mixed roles (e.g., a Security Engineer, a CISO, and a
Developer) wants to understand how their individual contributions
connect within the Gemara model. The system provides a
collaboration view that maps each team member's role to the
Gemara layers they primarily interact with and shows the
handoff points between roles. For example: the Security Engineer
authors threat assessments and control catalogs (Layer 2), which
the CISO references when drafting organizational policy (Layer 3),
which the Developer consumes to understand what controls apply to
their deployment pipeline (Layer 4 sensitive activities).

**Why this priority**: This feature transforms Pac-Man from an
individual learning tool into a team coordination tool. It is
lower priority because it depends on the role definitions
established in User Story 3, and because individual learning
paths deliver standalone value even without the team view.

**Independent Test**: Can be tested by configuring a team with
3 or more different roles and verifying that the system produces
a visual or structured map showing which Gemara layers each role
owns, where handoffs occur, and what artifacts flow between
roles.

**Acceptance Scenarios**:

1. **Given** a team consists of a Security Engineer, a Compliance
   Officer, and a Developer, **When** the collaboration view is
   generated, **Then** it shows the Security Engineer mapped to
   Layers 1-2 (Guidance/Controls authoring), the Compliance
   Officer mapped to Layer 3 (Policy authoring) and Layer 5
   (Evaluation review), and the Developer mapped to Layer 4
   (Sensitive Activities) and Layer 5 (Evaluation execution).

2. **Given** a collaboration view is displayed, **When** a user
   selects a handoff point between two roles, **Then** the
   system displays the specific Gemara artifacts that flow
   across that boundary and links to the relevant tutorials
   for both the producing and consuming roles.

3. **Given** a team has been configured, **When** a new role is
   added to the team, **Then** the collaboration view updates
   to include the new role's layer mappings and any new handoff
   points that emerge.

---

### User Story 6 - Guided Gemara Content Authoring (Priority: P6)

A user wants to author a new Gemara artifact (e.g., a Guidance
Catalog, a Control Catalog, or a Policy document). Based on their
role and the artifact type, the system provides step-by-step
guided authoring that mirrors the structure of the corresponding
Gemara tutorial but is personalized to the user's context. The
system explains each field, suggests values based on the user's
stated scope, and validates the in-progress artifact against the
Gemara CUE schema at each step.

**Why this priority**: Content authoring is the ultimate
value-delivery action — it turns learning into production
artifacts. However, it depends on the learning paths (US3) and
reusable content blocks (US4) being in place to provide
contextual guidance during the authoring process.

**Independent Test**: Can be tested by selecting a role and
artifact type, walking through the guided authoring steps,
and verifying that the resulting document passes
`cue vet -c -d '#<SchemaType>'` validation against the Gemara
schema.

**Acceptance Scenarios**:

1. **Given** a Security Engineer wants to author a Threat
   Catalog, **When** they begin guided authoring, **Then** the
   system walks them through each required section (scope,
   metadata, capabilities, threats) with role-specific
   explanations and example values drawn from the Gemara
   tutorial examples.

2. **Given** a user is partway through authoring a Control
   Catalog, **When** they complete the metadata section, **Then**
   the system validates what has been entered so far against the
   Gemara `#ControlCatalog` schema and reports any validation
   errors with actionable fix suggestions.

3. **Given** a user has completed all authoring steps, **When**
   they request final output, **Then** the system produces a
   YAML document that passes full `cue vet` validation and
   follows the naming conventions documented in the Gemara
   tutorials (e.g., `ORG.PROJ.COMPONENT.THR##`).

---

### Edge Cases

- What happens when a user's activity profile maps to Gemara
  layers that have no tutorials currently available (e.g.,
  Layers 6-7)? The system MUST inform the user which layers
  lack tutorials, provide the model documentation for those
  layers as a fallback, and suggest the closest available
  tutorials that are relevant to their activities.

- What happens when the Gemara tutorials directory is empty or
  inaccessible? The system MUST report a clear error indicating
  the expected path and how to resolve the issue (e.g., clone
  the Gemara repository or update the configured path).

- What happens when a user provides a free-text role that
  partially matches a predefined role (e.g., "Product Security
  Engineer" contains "Security Engineer")? The system MUST
  recognize the partial match, inform the user of the overlap,
  and proceed to activity probing to determine which specific
  layer mappings apply rather than assuming the generic
  "Security Engineer" path.

- What happens when a user's activity description contains no
  recognizable domain keywords? The system MUST present the
  full list of activity categories for manual selection and
  explain each category's relationship to the Gemara layers,
  rather than guessing or defaulting silently.

- What happens when a user's activities span multiple layers
  that would normally be handled by different roles (e.g., a
  user who both authors controls at Layer 2 and defines policy
  at Layer 3)? The system MUST generate a combined learning
  path that covers all relevant layers, ordered by the user's
  stated priority among their activities.

- What happens when the Gemara schema version used by the
  tutorials changes? The system MUST detect version mismatches
  between its content blocks and the current upstream tutorials,
  and flag affected content for review before presenting it
  to users.

- What happens when the user selects a stable schema version but
  the tutorials reference schemas that are only available in a
  newer version? The system MUST identify which tutorial
  sections are incompatible with the selected version and
  clearly indicate that the user needs to either upgrade their
  schema selection or skip those sections.

- What happens when no tagged releases exist in the upstream
  repository (e.g., a fresh fork with no tags)? The system MUST
  fall back to reading the VERSION file and `cue.mod/module.cue`
  from the local repository clone and present whatever version
  information is available, with a warning that no official
  releases were found.

- What happens when the Gemara MCP server was installed but
  becomes unavailable mid-session (e.g., Docker container
  stops, binary crashes)? The system MUST detect the
  disconnection, notify the user, automatically fall back to
  local equivalents (bundled lexicon, local CUE validation,
  cached schema docs), and attempt to reconnect on subsequent
  operations. In-progress work MUST NOT be lost.

- What happens when the user selects "Latest" for the Gemara
  schema but the installed gemara-mcp was built against an
  older schema version? The system MUST warn the user that
  validation results, lexicon entries, and schema documentation
  from the MCP server may not reflect the latest schema
  changes. The system MUST NOT silently use mismatched
  versions — it MUST present the mismatch and let the user
  decide whether to proceed (accepting the risk), update their
  gemara-mcp, or switch to the schema version that matches
  their MCP server.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept a job role as input through a
  two-phase discovery process (role identification followed by
  activity probing) and produce a tailored learning path
  referencing specific Gemara tutorials based on the combined
  role + activity profile.
- **FR-002**: System MUST provide a predefined list of common
  roles (at minimum: Security Engineer, Compliance Officer,
  CISO/Security Leader, Developer, Platform Engineer, Policy
  Author, Auditor) plus a "My role isn't listed" option that
  accepts free-text input. This list MUST be updated as
  research is completed to include newly identified personas
  and job titles from the Gemara community.
- **FR-003**: System MUST annotate each tutorial reference in a
  learning path with three sections: "Why this matters for your
  role," "How you will use this," and "What you will learn."
- **FR-004**: System MUST read tutorial content from a
  configurable directory path, defaulting to the Gemara
  tutorials location at
  `~/github/openssf/gemara/gemara/docs/tutorials`.
- **FR-005**: System MUST extract reusable content blocks from
  Gemara tutorials, each tagged with source tutorial identity,
  Gemara schema version, and the Gemara layer it belongs to.
- **FR-006**: System MUST detect when upstream tutorial content
  has changed relative to previously extracted content blocks
  and flag affected blocks for review.
- **FR-007**: System MUST map each role to Gemara layers based
  on the combination of role title and stated activities, using
  the seven-layer model as the organizing framework. The same
  role title with different activity profiles MUST produce
  different layer mappings.
- **FR-008**: System MUST generate a collaboration view for a
  configured team showing role-to-layer mappings, artifact
  flows, and handoff points between roles.
- **FR-009**: System MUST provide guided authoring for Gemara
  artifact types that have published CUE schemas (currently:
  GuidanceCatalog, ControlCatalog, ThreatCatalog, Policy,
  MappingDocument, EvaluationLog).
- **FR-010**: System MUST validate authored artifacts against the
  pinned Gemara CUE schema at each authoring step and upon
  final output, reporting errors with actionable fix guidance.
  When the Gemara MCP server is available, validation MUST use
  the `validate_gemara_artifact` tool. When the MCP server is
  unavailable, the system MUST fall back to local `cue vet`.
- **FR-011**: System MUST use the Gemara lexicon terms
  consistently in all user-facing output and MUST NOT redefine
  or use alternate meanings for controlled vocabulary terms.
  When the Gemara MCP server is available, the system MUST
  source lexicon data from the `get_lexicon` tool to ensure
  alignment with the latest upstream vocabulary.
- **FR-012**: System MUST produce all structured output in both
  YAML and JSON formats, with YAML as the default for
  human-readable output.
- **FR-013**: System MUST function on Linux and macOS without
  requiring platform-specific installation steps or
  dependencies beyond the documented prerequisites.
- **FR-014**: When a user selects "My role isn't listed," the
  system MUST accept free-text role input, extract keywords
  from the response, identify partial matches against known
  role profiles, and present the mapping for user confirmation
  before proceeding to activity probing.
- **FR-015**: System MUST allow users to navigate non-linearly
  through a learning path, accessing any step regardless of
  completion status, while indicating prerequisite knowledge
  that may have been skipped.
- **FR-016**: System MUST fetch schema release information from
  the upstream Gemara repository
  (`github.com/gemaraproj/gemara`) by querying available tagged
  releases.
- **FR-017**: System MUST present the user with a choice between
  "Stable" (the most recent tagged release where the schemas
  the user needs are marked `@status(Stable)`) and "Latest"
  (the most recent tagged release overall). The system MUST
  display the version number for each option and indicate which
  schemas are Stable versus Experimental at each version.
- **FR-018**: System MUST cache fetched schema version
  information locally so that the tool remains functional when
  the upstream repository is unreachable. The cache MUST record
  the date of the last successful fetch.
- **FR-019**: System MUST use the user's selected schema version
  consistently for all validation, content block extraction,
  and guided authoring operations throughout the session.
  Switching versions mid-session MUST be supported but MUST
  require explicit user confirmation and MUST re-validate any
  in-progress work against the new version.
- **FR-020**: System MUST notify the user when a newer schema
  version is available upstream compared to their currently
  selected or cached version, without forcing an upgrade.
- **FR-021**: After role identification, the system MUST conduct
  activity probing by asking the user to describe what they do
  or what problem they are trying to solve. The system MUST
  accept free-text descriptions — including full sentences
  such as "map my best practices to the EU CRA" or "create a
  reusable machine-readable format for my standards" — and
  extract domain keywords (e.g., "SDLC," "threat modeling,"
  "evidence collection," "CI/CD pipeline," "dependency
  management," "audit interviews," "secure architecture
  review," "penetration testing," "create a policy," "define
  timeline for adherence," "EU CRA," "machine-readable
  format," "best practices") to refine layer mappings.
- **FR-022**: System MUST maintain a keyword-to-layer mapping
  that associates domain activity terms with specific Gemara
  layers and tutorial content. This mapping is illustrative,
  not exhaustive — the system MUST support extending it
  through configuration. The keyword and role lists MUST be
  updated as ongoing user research, persona studies, and
  community feedback identify new activity patterns, job
  titles, and domain terms that are not yet represented.
  At minimum the following MUST be included:
  - **Layer 1 (Guidance)**: mapping best practices to
    regulatory frameworks (e.g., "map my best practices to the
    EU CRA," "align with NIST," "OWASP," "HIPAA," "GDPR,"
    "PCI," "ISO"), creating reusable machine-readable formats
    for standards or internal best practices (e.g., "create a
    reusable machine-readable format for my standards,"
    "codify internal use-case," "formalize best practices"),
    evidence collection, and defining guidance requirements
    that need to be met.
  - **Layer 2 (Threats & Controls)**: SDLC, threat modeling,
    penetration testing, secure architecture review, CI/CD
    pipeline management, dependency management, upstream
    open-source component usage, writing custom controls,
    importing external catalogs (OSPS Baseline, FINOS CCC).
  - **Layer 3 (Risk & Policy)**: creating a policy, defining
    timeline for adherence to policy, scope definition, audit
    interviews, assessment plans, adherence requirements,
    risk appetite, non-compliance handling.
  - **Layers 1 and 3 combined**: evidence collection and
    adherence may span both layers depending on whether the
    user is defining guidance requirements (Layer 1) or
    operationalizing them in policy (Layer 3). The system
    MUST ask a clarifying follow-up when activity keywords
    match both layers.
- **FR-023**: When a role title partially matches a predefined
  role (e.g., "Product Security Engineer" contains "Security
  Engineer"), the system MUST NOT assume the generic role's
  layer mapping is correct. It MUST inform the user of the
  partial match and proceed to activity probing to determine
  the actual layer mappings.
- **FR-024**: System MUST allow users to create, save, and reuse
  custom role profiles. A custom role profile MUST include:
  role name, activity keywords, Gemara layer mappings, and
  an optional description. Saved profiles MUST be available
  for selection in future sessions and MUST be assignable to
  team members in the collaboration view (US5).
- **FR-025**: System MUST support adding new roles to the
  predefined list through configuration data (not code
  changes). New role definitions MUST specify: role name,
  description, default activity keywords, default layer
  mappings, and default tutorial ordering.
- **FR-026**: On first launch, the system MUST offer the user
  the option to install the Gemara MCP server
  ([github.com/gemaraproj/gemara-mcp](https://github.com/gemaraproj/gemara-mcp))
  before any other operation. The system MUST explain the
  three tools the MCP server provides (`get_lexicon`,
  `validate_gemara_artifact`, `get_schema_docs`) and how each
  enhances the Pac-Man experience.
- **FR-027**: System MUST support two MCP server installation
  methods: pre-built binary and Docker. Installation guidance
  MUST be platform-appropriate (Linux or macOS) and MUST
  include verification that the server is running and
  responsive.
- **FR-028**: When the Gemara MCP server is installed and
  running, the system MUST use it as the preferred source for
  lexicon data (`get_lexicon`), schema documentation
  (`get_schema_docs`), and artifact validation
  (`validate_gemara_artifact`). MCP-sourced data MUST take
  precedence over locally bundled or cached equivalents.
- **FR-029**: When the Gemara MCP server is not installed or
  not running, the system MUST fall back to local equivalents:
  bundled lexicon data for terminology, local CUE tooling for
  validation, and cached schema documentation for reference.
  The system MUST inform the user which capabilities are
  degraded and offer MCP server installation at any point
  during the session.
- **FR-030**: The system MUST detect whether the Gemara MCP
  server is already installed and running at the start of each
  session. If detected, the system MUST skip the installation
  prompt and proceed directly to schema version selection
  (US2), confirming MCP server availability in the session
  status.
- **FR-031**: When the user selects "Latest" as their Gemara
  schema version and the Gemara MCP server is installed, the
  system MUST verify that the installed gemara-mcp version is
  compatible with the selected Gemara schema version. The
  `gemaraproj/gemara` schema version and the
  `gemaraproj/gemara-mcp` server version MUST be coordinated
  for accurate validation results, lexicon data, and schema
  documentation. If a version mismatch is detected, the system
  MUST warn the user that results may be inaccurate and MUST
  recommend one of: (a) updating the gemara-mcp installation
  to a version built against the selected schema version, or
  (b) selecting the schema version that matches the installed
  gemara-mcp version.
- **FR-032**: The system MUST determine gemara-mcp version
  compatibility by querying the installed server for its
  version information and comparing the Gemara schema version
  it was built against with the user's selected schema version.
  If the server does not expose version metadata, the system
  MUST warn the user that compatibility cannot be verified and
  recommend updating to a gemara-mcp version that exposes this
  information.

### Key Entities

- **Role**: A job function that serves as the starting point for
  tutorial routing. Attributes: name, description, default
  activity keywords, default Gemara layer mappings (list of
  layer numbers 1-7), typical artifacts produced, typical
  artifacts consumed, source (predefined or custom), saved
  activity profile (if customized).

- **Activity Profile**: The result of activity probing for a
  specific user session. Attributes: extracted keywords (list
  of domain terms), matched activity categories, resolved
  Gemara layer mappings (may differ from role defaults),
  user-provided description (free-text), confidence indicators
  per layer mapping (strong match vs. inferred).

- **Learning Path**: An ordered sequence of tutorial references
  tailored to a specific role. Attributes: target role,
  ordered list of path steps, completion status per step.

- **Path Step**: A single item within a learning path.
  Attributes: tutorial reference (source file path), Gemara
  layer, why-annotation, how-annotation, what-annotation,
  prerequisites (list of other path step references),
  completion status.

- **Content Block**: A modular, reusable unit of knowledge
  extracted from a Gemara tutorial. Attributes: source tutorial
  identity, source section, Gemara schema version at time of
  extraction, Gemara layer, content category (pattern,
  validation step, naming convention, schema structure),
  content body, last-verified date.

- **Team Configuration**: A collection of roles representing a
  cross-functional team. Attributes: team name, list of roles,
  generated collaboration view.

- **Collaboration View**: A mapping of roles to Gemara layers
  showing artifact flows and handoff points. Attributes:
  role-to-layer mappings, handoff points (pairs of roles with
  the artifact type that flows between them), layer coverage
  gaps (layers with no assigned role).

- **Schema Version**: A specific tagged release of the Gemara
  schema repository. Attributes: version tag (e.g., v0.20.0),
  release date, schema status map (file name to Stable or
  Experimental), CUE language version, source (upstream fetch
  or local cache), cache timestamp.

- **Authored Artifact**: A Gemara-conformant document produced
  through guided authoring. Attributes: artifact type (from
  Gemara `#ArtifactType` enum), target schema definition,
  target schema version, validation status, content
  (YAML/JSON), authoring role.

- **MCP Server Connection**: The Gemara MCP server instance
  used by the current session. Attributes: installation method
  (binary or Docker), connection status (running, stopped, not
  installed), server version, Gemara schema version the server
  was built against, compatibility status with the user's
  selected schema version (compatible, mismatched, unknown),
  available tools (get_lexicon, validate_gemara_artifact,
  get_schema_docs), last health check timestamp.

### Assumptions

- The Gemara tutorials directory structure and file naming
  conventions will remain stable across minor versions. If
  the directory structure changes materially, content block
  extraction will need to be re-run.
- The predefined role list (FR-002) and keyword-to-layer
  mapping (FR-022) cover the most common personas and activity
  terms in the Gemara ecosystem but are explicitly not
  exhaustive. Both MUST be treated as living artifacts that
  are updated as user research, persona studies, and community
  feedback reveal new patterns. The system is designed to
  handle unlisted roles through free-text input and activity
  probing. New predefined roles and keywords can be added
  through configuration data without code changes (FR-025).
- Users have `cue` installed locally for schema validation, as
  documented in the Gemara project prerequisites.
- The Gemara schema version is available from
  `cue.mod/module.cue` in the Gemara repository and can be
  read programmatically. Tagged releases in the upstream
  repository follow semantic versioning and are published to
  the CUE Central Registry.
- Schema `@status()` attributes (Stable, Experimental,
  Deprecated) are embedded in the CUE source files and can be
  parsed programmatically to determine which schemas are stable
  at a given release.
- The upstream Gemara repository is publicly accessible and
  its tagged releases can be queried without authentication.
- Free-text role matching (FR-014) and activity keyword
  extraction (FR-021) use keyword-based matching against a
  maintained vocabulary of domain activity terms rather than
  requiring a machine learning model. The keyword-to-layer
  mapping (FR-022) is extensible through configuration.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user with no prior Gemara knowledge can complete
  role identification and activity probing and receive a
  tailored learning path within 60 seconds of starting the
  discovery process.

- **SC-002**: Two users with the same job title but different
  activity descriptions receive different learning paths that
  correctly reflect their stated activities (e.g., two Product
  Security Engineers — one audit-focused, one pipeline-focused
  — receive paths targeting different Gemara layers).

- **SC-011**: 90% of users following an activity-tailored
  learning path can correctly identify which Gemara layer their
  primary work maps to after completing the first two steps of
  their path.

- **SC-012**: Users who select "My role isn't listed" and provide
  a free-text role description complete the full discovery
  process (role entry, activity probing, path generation) with
  the same success rate as users who select a predefined role.

- **SC-003**: Reusable content blocks extracted from Gemara
  tutorials cover at least 80% of the actionable patterns in
  each processed tutorial (scope definition, metadata setup,
  entity definition, validation, cross-referencing).

- **SC-004**: When upstream Gemara tutorials change, the system
  detects and flags affected content blocks within one
  execution cycle — no manual tracking required.

- **SC-005**: Users completing the guided authoring flow produce
  a valid Gemara artifact that passes `cue vet` validation on
  the first attempt at least 85% of the time.

- **SC-006**: A cross-functional team of 3+ roles can generate
  a collaboration view that correctly maps all roles to their
  primary Gemara layers and identifies all handoff points in
  under 60 seconds.

- **SC-007**: All system output uses Gemara lexicon terms
  consistently — zero instances of redefined or alternate
  terminology in any user-facing text.

- **SC-008**: The tool runs identically on Linux and macOS with
  no platform-specific instructions or workarounds required.

- **SC-009**: The system fetches and presents available schema
  versions (stable vs. latest) within 10 seconds of launch
  when network access is available. When offline, the system
  falls back to cached version data and informs the user
  within 5 seconds.

- **SC-010**: 100% of validation operations, content block
  extractions, and guided authoring sessions use the schema
  version explicitly selected by the user — no operation
  silently defaults to an unselected version.

- **SC-013**: Users who install the Gemara MCP server can
  complete the installation and verify server connectivity
  within 5 minutes, with no more than 3 steps for either the
  binary or Docker installation method.

- **SC-014**: When the MCP server is available, all lexicon
  lookups, schema documentation requests, and artifact
  validations use the MCP server's tools. When the MCP server
  becomes unavailable, the system falls back to local
  equivalents within 5 seconds with zero data loss on
  in-progress work.
