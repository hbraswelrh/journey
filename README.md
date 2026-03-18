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
which provides tools, resources, and prompts that enhance the
learning and authoring experience:

| Category | Name | Function |
|:---------|:-----|:---------|
| Tool | `validate_gemara_artifact` | Validate YAML artifacts against Gemara CUE schemas |
| Resource | `gemara://lexicon` | Retrieve the upstream Gemara lexicon (34+ terms) |
| Resource | `gemara://schema/definitions` | Retrieve schema documentation for the CUE module |
| Prompt | `threat_assessment` | Interactive wizard for Threat Catalog creation (artifact mode) |
| Prompt | `control_catalog` | Interactive wizard for Control Catalog creation (artifact mode) |

The MCP server operates in two modes:

- **Artifact mode** (default) — Full capabilities including
  guided creation wizards (`threat_assessment` and
  `control_catalog` prompts)
- **Advisory mode** — Read-only analysis and validation
  (tools and resources only, no wizard prompts)

MCP server installation is offered during first launch. All
capabilities remain functional without the MCP server using
local CUE tooling and bundled lexicon data.

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

- Linux or macOS (Windows: use WSL)
- [Go](https://go.dev/dl/) 1.21 or later
- [CUE](https://cuelang.org/docs/introduction/installation/)
  v0.15.1 or later (for schema validation)
- Git (with DCO sign-off and commit signing configured)
- [OpenCode](https://opencode.ai) (required — the AI coding
  agent that serves as the harness for tutorials and MCP
  interaction)

### macOS Setup

```bash
# Install Homebrew (if not already installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install required tools
brew install go
brew install cue-lang/tap/cue

# Install OpenCode (required harness)
brew install anomalyco/tap/opencode

# Install Gemara MCP server (recommended)
git clone https://github.com/gemaraproj/gemara-mcp.git
cd gemara-mcp
git checkout main
make build
cd ..
```

### Linux Setup

```bash
# Install Go (https://go.dev/dl/)
wget https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install CUE
go install cuelang.org/go/cmd/cue@latest

# Install OpenCode (required harness)
curl -fsSL https://opencode.ai/install | bash

# Install Gemara MCP server (recommended)
git clone https://github.com/gemaraproj/gemara-mcp.git
cd gemara-mcp
git checkout main
make build
cd ..
```

## Getting Started

### Step 1: Clone and Build Pac-Man

```bash
git clone https://github.com/hbraswelrh/pacman.git
cd pacman
make build
```

### Step 2: Configure the Gemara MCP Server

If you built gemara-mcp from source, configure it in
`opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "gemara-mcp": {
      "type": "local",
      "command": [
        "/path/to/gemara-mcp/bin/gemara-mcp",
        "serve", "--mode", "artifact"
      ],
      "enabled": true
    }
  }
}
```

Use `--mode artifact` (default) for full capabilities
including guided creation wizards, or `--mode advisory`
for read-only analysis and validation only.

If you already have gemara-mcp built elsewhere, point
the config to your existing binary path.

### Step 3: Launch OpenCode

OpenCode is the required harness for Pac-Man. It provides
the interactive terminal for tutorials and serves as the
MCP client for communicating with the Gemara MCP server.

```bash
cd pacman
opencode
```

OpenCode will:
- Read the project's `AGENTS.md` for context
- Start the gemara-mcp server automatically (from
  `opencode.json`)
- Provide access to MCP tools, resources, and prompts

### Step 4: Verify and Launch

```bash
# Verify your environment
./pacman --doctor

# Start OpenCode (the tutorial interface)
opencode
```

Pac-Man is a CLI tool for environment verification only.
The tutorial experience is delivered through OpenCode,
which reads `AGENTS.md` and connects to the gemara-mcp
server.

### Tutorials Source

Tutorials are sourced from the upstream Gemara repository
at [gemaraproj/gemara](https://github.com/gemaraproj/gemara)
(`docs/tutorials/` on the `main` branch).

On first tutorial launch, Pac-Man will automatically clone
the repository (shallow, single-branch) to
`~/.local/share/pacman/gemara/` if tutorials aren't found
locally. On subsequent launches, it pulls the latest
changes from `main`.

To override the tutorials directory:

```bash
./pacman --tutorials /path/to/gemara/docs/tutorials
```

### Using MCP Wizard Prompts

The `threat_assessment` and `control_catalog` prompts are
**MCP protocol messages**, not CLI commands. They require
an MCP client to invoke them. OpenCode is that client.

To use a wizard prompt in OpenCode:

1. Ensure `opencode.json` has gemara-mcp configured with
   `--mode artifact`
2. Start OpenCode: `opencode`
3. Ask OpenCode to run the wizard:
   - "Run the threat_assessment prompt"
   - "Help me create a control catalog using the wizard"
   - "Start the threat assessment wizard for my CI/CD
     pipeline"

OpenCode will invoke the MCP prompt, present the wizard's
guided questions, and produce a validated YAML artifact.

Alternatively, run `./pacman` and select "Launch a wizard"
from the main menu. Pac-Man will collect your context
(scope, component, role) and generate a ready-to-paste
command for OpenCode.

### Keeping gemara-mcp Up to Date

The gemara-mcp server is built from source against the
upstream [gemaraproj/gemara-mcp](https://github.com/gemaraproj/gemara-mcp)
repository. To ensure your build reflects the latest schema
support and bug fixes, sync with upstream regularly.

**If you cloned directly from gemaraproj (no fork):**

```bash
cd gemara-mcp
git fetch origin
git checkout main
git pull origin main
make build
```

**If you cloned from a personal fork:**

```bash
cd gemara-mcp

# Add upstream remote (one-time setup)
git remote add upstream \
  https://github.com/gemaraproj/gemara-mcp.git

# Fetch and merge upstream changes
git fetch upstream
git checkout main
git merge upstream/main

# Rebuild
make build
```

**Verify the build:**

```bash
# Check the binary runs
./bin/gemara-mcp --version

# Or use Pac-Man's doctor command
cd /path/to/pacman
./pacman --doctor
```

**When to sync:**

- Before starting a new tutorial or authoring session
- When `./pacman --doctor` reports a version mismatch
  between the MCP server and your selected Gemara schema
  version
- When new Gemara schema releases are published at
  [gemaraproj/gemara](https://github.com/gemaraproj/gemara/releases)

The gemara-mcp server's schema version must be compatible
with the Gemara schema version selected in your Pac-Man
session. If you select a newer schema version than your
MCP server was built against, Pac-Man will warn you during
schema version selection.

### First Launch

On first launch, Pac-Man will:

1. **Detect or configure the Gemara MCP server** — If not
   found, the system offers to build from source (clone
   via SSH or HTTPS from gemaraproj/gemara-mcp, checkout
   main, and build) or provide the path to an existing
   binary. After setup, select a server mode (artifact or
   advisory) and configure `opencode.json`.
2. **Prompt for schema version selection** — Choose between
   Stable (core schemas marked `@status(Stable)`) or Latest
   (most recent tagged release).
3. **Role and activity discovery** — Identify your role and
   describe your daily activities to receive a tailored
   learning path through the Gemara tutorials.

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
    schema/                # Schema release fetching and version
                           #   selection
    roles/                 # Role identification, activity probing,
                           #   custom profiles
    tutorials/             # Tutorial loading, learning path
                           #   generation
    blocks/                # Content block extraction, drift
                           #   detection, retrieval
    team/                  # Team configuration, handoff detection,
                           #   collaboration view
    authoring/             # Guided Gemara content authoring,
                           #   validation, YAML/JSON output
    cli/                   # CLI commands, setup flows, TUI
                           #   rendering
  go.mod                   # Go module definition
  Makefile                 # Single entry point for build/test/lint
  specs/                   # Feature specifications
    001-role-based-tutorial-engine/
      spec.md              # Role-based tutorial engine spec
      plan.md              # Implementation plan (US1)
      tasks.md             # Task breakdown (US1)
      plan-us2.md .. plan-us6.md   # Plans for US2-US6
      tasks-us2.md .. tasks-us6.md # Tasks for US2-US6
      checklists/          # Quality validation checklists
  docs/adrs/               # Architecture Decision Records
  .github/                 # Issue templates, PR template, CI
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
