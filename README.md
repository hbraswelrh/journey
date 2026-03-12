# Pac-Man

Pac-Man is a role-based tutorial engine for the
[Gemara](https://github.com/gemaraproj/gemara) governance, risk,
and compliance (GRC) schema project. It processes a user's job
role and daily activities to generate tailored learning paths
through the Gemara tutorials, guides users through authoring
Gemara-conformant artifacts, and enables cross-functional teams
to understand how their work connects across Gemara's seven-layer
model.

Pac-Man answers three questions for every user:

- **Why** the Gemara project matters for their specific role
- **How** they will use Gemara concepts in their day-to-day work
- **What** each tutorial covers and what artifacts they will produce

## Problem Statement

Gemara's tutorials are organized by artifact type (Guidance
Catalogs, Control Catalogs, Policies), not by audience. A
Security Engineer, a Compliance Officer, and a Developer all
encounter the same documentation in the same order, even though
each role interacts with different layers of the model and
produces different artifacts. Standards, best practices, and
guidelines embedded in tutorials are not structured for reuse
or adaptation as the project evolves.

Pac-Man bridges this gap by routing users to the content they
need based on who they are and what they do.

## Capabilities

### Role and Activity Discovery

Pac-Man uses a two-phase discovery process to determine the
right learning path:

1. **Role Identification** — The user selects from a predefined
   list of common GRC roles (Security Engineer, Compliance
   Officer, CISO, Developer, Platform Engineer, Policy Author,
   Auditor) or enters a custom role via free text.

2. **Activity Probing** — The system asks the user to describe
   their daily activities or the problem they are trying to
   solve. It extracts domain keywords (e.g., "SDLC," "threat
   modeling," "evidence collection," "CI/CD pipeline," "EU CRA,"
   "create a policy") and maps them to specific Gemara layers.

The same job title with different activity descriptions produces
different learning paths. A Product Security Engineer focused on
audit interviews and evidence collection receives a path through
Layer 1 (Guidance) and Layer 3 (Policy). A Product Security
Engineer focused on CI/CD pipelines and dependency management
receives a path through Layer 2 (Threats and Controls).

### Gemara MCP Server Integration

Pac-Man integrates with the
[Gemara MCP server](https://github.com/gemaraproj/gemara-mcp),
which provides three tools that enhance the learning and
authoring experience:

| MCP Tool                     | Function                                           |
|:-----------------------------|:---------------------------------------------------|
| `get_lexicon`                | Retrieve the upstream Gemara lexicon (34+ terms)   |
| `validate_gemara_artifact`   | Validate YAML artifacts against Gemara CUE schemas |
| `get_schema_docs`            | Retrieve schema documentation for the CUE module   |

MCP server installation is offered during first launch. All
capabilities remain functional without the MCP server using
local CUE tooling and bundled lexicon data. When using the
latest (non-stable) Gemara schema version, the installed
gemara-mcp version must be coordinated with the Gemara schema
version for accurate validation results.

### Schema Version Selection

On each session, Pac-Man fetches available Gemara releases from
the upstream repository and presents a choice:

- **Stable** — The most recent release where core schemas
  (`base`, `metadata`, `mapping_inline`) are marked
  `@status(Stable)`.
- **Latest** — The most recent tagged release, which may include
  schemas marked `@status(Experimental)`.

The selected version governs all validation, tutorial alignment,
and guided authoring for the session. Version data is cached
locally for offline use.

### Tailored Learning Paths

Each learning path is an ordered sequence of Gemara tutorials
annotated with role-specific context. Every step includes three
sections:

- **Why this matters for your role** — Connects the tutorial
  content to the user's stated activities.
- **How you will use this** — Describes practical application
  in the user's day-to-day work.
- **What you will learn** — Summarizes the tutorial's content
  and the artifacts it teaches the user to produce.

Users can navigate paths non-linearly. Skipped prerequisites
are flagged but not enforced.

### Reusable Content Blocks

Pac-Man extracts modular, reusable content blocks from Gemara
tutorials — patterns for scope definition, metadata setup,
CUE validation, naming conventions, and cross-referencing. Each
block is tagged with its source tutorial and Gemara schema
version. When upstream tutorials change, Pac-Man detects which
blocks are affected and flags them for review.

### Cross-Functional Collaboration View

Teams with mixed roles can generate a collaboration view that
maps each role to the Gemara layers they interact with, shows
artifact flows between roles, and identifies handoff points.
This view answers the question: "How does my work connect to
the rest of the team's work within the Gemara model?"

### Guided Content Authoring

Users can author Gemara artifacts (Guidance Catalogs, Control
Catalogs, Threat Catalogs, Policies, Mapping Documents,
Evaluation Logs) through a step-by-step guided process. The
system explains each field, suggests values based on the user's
scope, and validates the in-progress artifact against the Gemara
CUE schema at each step. Final output is a YAML document that
conforms to the selected Gemara schema version.

## Gemara Layer Reference

Pac-Man routes users to tutorials based on the Gemara
seven-layer model:

| Layer | Name                | Purpose                                              |
|:------|:--------------------|:-----------------------------------------------------|
| 1     | Guidance            | Standards, best practices, regulatory requirements   |
| 2     | Threats & Controls  | Threat catalogs, control catalogs, security measures  |
| 3     | Risk & Policy       | Organizational policy, assessment plans, adherence    |
| 4     | Sensitive Activities| Deployment pipelines, CI/CD, operational activities   |
| 5     | Evaluation          | Assessment logs, control evaluations, evidence        |
| 6     | Results             | Aggregated evaluation results                         |
| 7     | Communication       | Reporting and stakeholder communication               |

Tutorials currently exist for Layers 1 through 5. Layers 6 and
7 are covered by model documentation. The Gemara schema marks
`base`, `metadata`, and `mapping_inline` as Stable; all layer
schemas are currently Experimental.

## Prerequisites

- [Go](https://go.dev/dl/) 1.21 or later
- [CUE](https://cuelang.org/docs/introduction/installation/)
  v0.15.1 or later (for schema validation)
- Linux or macOS (Windows is not supported)
- Git (with DCO sign-off and commit signing configured)
- [Gitleaks](https://github.com/gitleaks/gitleaks) (for secret
  scanning; installation guidance is provided on first use)

### Recommended Installation via Homebrew

Homebrew is the preferred installation method for required and
recommended tools on macOS and Linux:

```bash
# CUE (required for schema validation)
brew install cue-lang/tap/cue

# Gitleaks (required for pre-commit secret scanning)
brew install gitleaks

# OpenCode (recommended AI development harness)
brew install anomalyco/tap/opencode

# Podman (for container-based MCP server installation)
brew install podman
```

Alternative installation methods (binary releases, install
scripts) are documented in each tool's upstream repository.

### Optional

- [Gemara MCP server](https://github.com/gemaraproj/gemara-mcp)
  (for enhanced lexicon, validation, and schema documentation;
  installable from source or via Podman during first launch)

## Getting Started

```bash
# Clone the repository
git clone https://github.com/hbraswelrh/pacman.git
cd pacman

# Build
make build

# Run
./pacman
```

On first launch, Pac-Man will offer to install the Gemara MCP
server, then prompt for schema version selection (Stable or
Latest), followed by role and activity discovery.

## Project Structure

```
pacman/
  cmd/pacman/
    main.go                # Application entry point
  internal/
    consts/                # Centralized constants (no magic strings)
    mcp/                   # MCP server detection, installation,
                           #   client, version compatibility,
                           #   OpenCode config management
    fallback/              # Local fallback (bundled lexicon,
                           #   local CUE validation, cached
                           #   schema docs)
    session/               # Session state management
    cli/                   # CLI commands and setup flows
  go.mod                   # Go module definition
  Makefile                 # Single entry point for build/test/lint
  specs/                   # Feature specifications
    001-role-based-tutorial-engine/
      spec.md              # Role-based tutorial engine spec
      plan.md              # Implementation plan (US1)
      tasks.md             # Task breakdown (US1)
      checklists/          # Quality validation checklists
  docs/adrs/               # Architecture Decision Records
  .specify/
    memory/
      constitution.md      # Project constitution (authoritative)
    templates/             # Specification and planning templates
```

## Contributing

All contributions require:

- A dedicated feature branch
  (`<issue-number>-<short-description>`)
- Conventional Commits format (`feat:`, `fix:`, `docs:`, etc.)
- DCO sign-off on every commit (`git commit -s`)
- Cryptographic commit signatures (`git commit -S`)
- Two non-author approvals on every pull request
- Gitleaks pre-commit hook passing with no findings
- `make build`, `make lint`, and `make test` passing with
  zero warnings

Upstream fork synchronization is required before opening pull
requests against upstream repositories. The upstream Gemara
repository is the source of truth for all schema definitions.

Full contribution guidelines are governed by the project
constitution at `.specify/memory/constitution.md`.

## Upstream Projects

Pac-Man depends on and integrates with the following upstream
projects:

| Project | Repository | Role |
|:--------|:-----------|:-----|
| Gemara | [gemaraproj/gemara](https://github.com/gemaraproj/gemara) | GRC schema definitions, tutorials, lexicon |
| Gemara MCP Server | [gemaraproj/gemara-mcp](https://github.com/gemaraproj/gemara-mcp) | Lexicon, validation, and schema documentation tools |

The upstream repositories are the authoritative sources for
schema definitions, lexicon terms, and tutorial content. When
discrepancies exist between Pac-Man's assumptions and upstream
behavior, upstream is correct.

## Architecture Decision Records

Every non-trivial technical or process decision is recorded as
an Architecture Decision Record (ADR) in `docs/adrs/`. ADRs
follow the format: Title, Status (Proposed, Accepted,
Deprecated, Superseded), Context, Decision, Consequences.
When a decision is questioned, the relevant ADR provides the
authoritative rationale.

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
