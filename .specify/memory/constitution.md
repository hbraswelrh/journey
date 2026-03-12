<!--
  === Sync Impact Report ===
  Version change: 1.2.0 → 1.3.0 (MINOR — new subsection,
    existing section expanded)
  Modified principles:
    - I. Schema Conformance — unchanged
    - II. Gemara Layer Fidelity — unchanged
    - III. Test-Driven Development — unchanged
    - IV. Tutorial-First Design — unchanged
    - V. Incremental Delivery — unchanged
    - VI. Decision Documentation — unchanged
    - VII. Centralized Constants — unchanged
    - VIII. Composability — unchanged
    - IX. Convention Over Configuration — unchanged
  Added principles: none
  Added sections: none
  Modified sections:
    - Technology & Integration Constraints: added OpenCode as
      preferred AI development harness
    - Coding Standards > Agent and Automation Awareness:
      updated to reference OpenCode as the preferred agent
      harness
  Removed sections: none
  Templates requiring updates:
    - .specify/templates/plan-template.md        ✅ no updates needed
    - .specify/templates/spec-template.md         ✅ no updates needed
    - .specify/templates/tasks-template.md        ✅ no updates needed
    - .specify/templates/commands/*.md            ✅ no command files present
  Follow-up TODOs: none
-->

# Pac-Man Constitution

## Core Principles

### I. Schema Conformance

All documents produced or consumed by Pac-Man MUST conform to
the Gemara CUE schemas published at
`github.com/gemaraproj/gemara`. Specifically:

- Every output artifact (EvaluationLog, MappingDocument, or
  any other Gemara artifact type) MUST pass validation via
  `cue vet -c -d '#<SchemaType>'` against the pinned Gemara
  schema version without errors.
- Input documents MUST be validated against the relevant Gemara
  schema definition before processing. Invalid input MUST be
  rejected with a clear, actionable error message referencing
  the schema constraint that failed.
- Schema version pinning MUST be explicit in the project
  configuration. Upgrades to a new Gemara schema version MUST
  be a deliberate, tested change — never automatic.
- The Gemara lexicon (`docs/lexicon.yaml`, 34 defined terms)
  MUST be used consistently in all user-facing output, error
  messages, and tutorial content. Domain terms MUST NOT be
  redefined or used with alternate meanings.

**Rationale**: Pac-Man's value depends entirely on producing
artifacts that interoperate with the broader Gemara ecosystem.
Schema violations break downstream consumers silently.

### II. Gemara Layer Fidelity

Pac-Man MUST respect the boundaries and semantics of Gemara's
seven-layer model when generating, evaluating, or mapping
compliance artifacts:

- Evaluation output MUST target Layer 5 (`#EvaluationLog`,
  `#ControlEvaluation`, `#AssessmentLog`) and MUST NOT
  conflate evaluation results with Layer 2 control definitions
  or Layer 3 policy declarations.
- Mapping operations MUST use the `#MappingDocument` schema
  with correct `#RelationshipType` values (`implements`,
  `equivalent`, `subsumes`, etc.). Mapping entries MUST
  reference valid `#EntryReference` targets.
- Tutorial content MUST teach each layer's purpose in isolation
  before introducing cross-layer interactions, mirroring the
  Definition → Pivot → Measurement structure of the model.
- When Gemara marks a schema as `@status(Experimental)`,
  Pac-Man MUST document this status to users and MUST NOT
  present experimental schemas as stable API surfaces.

**Rationale**: The seven-layer model is Gemara's core
intellectual contribution. Tools that blur layer boundaries
undermine the model's ability to separate concerns and create
confusion for adopters.

### III. Test-Driven Development

Testable logic MUST follow a red-green-refactor cycle:

- Tests MUST be written and confirmed failing before the
  corresponding implementation is authored.
- Schema validation logic MUST be tested with both positive
  fixtures (valid Gemara documents) and negative fixtures
  (intentionally invalid documents that MUST be rejected),
  following the pattern established in
  `~/github/openssf/gemara/gemara/test/schema_test.go`.
- Mapping and evaluation functions MUST be unit-testable by
  accepting structured input and returning structured output,
  decoupled from I/O, MCP transport, and file system access.
- Integration tests MUST cover end-to-end workflows: document
  ingestion → processing → Gemara-conformant output.
- Test files MUST live alongside source files using Go's
  `_test.go` convention.

**Rationale**: GRC tooling operates on compliance-critical data.
Untested transformation logic can silently produce invalid
evaluations or incorrect mappings, with real regulatory
consequences for users.

### IV. Tutorial-First Design

Guided tutorials are a primary deliverable, not an afterthought.
All user-facing functionality MUST be designed with
learnability as a first-class requirement:

- Every capability (evaluation, mapping, conversion) MUST have
  an accompanying tutorial that walks a user from zero context
  to a working result using realistic example data.
- Tutorials MUST be executable: each step MUST include the
  exact command or API call, expected output, and explanation
  of what happened. Users MUST be able to copy-paste and
  reproduce results.
- Tutorial content MUST be validated in CI. Example commands
  MUST be extracted and run as part of the test suite to
  prevent documentation drift.
- The MCP server integration MUST surface tutorials
  contextually — when a user invokes a capability, the
  relevant tutorial MUST be discoverable through the MCP
  interface.
- When a user requests information about Git workflows,
  branching strategy, DCO sign-off, or upstream contribution
  procedures, the tool MUST provide accurate, sourced
  documentation. Explanations MUST reference the relevant
  upstream project documentation or Git official documentation
  rather than paraphrasing from memory. Invalid or unverifiable
  information MUST NOT be presented.

**Rationale**: Gemara's GRC model is conceptually dense, and
many contributors will be new to Git-based open source
workflows. Without high-quality guided learning paths covering
both the domain and the tooling, adoption will be blocked by
the steep onboarding curve regardless of technical merit.

### V. Incremental Delivery

Features MUST be delivered as independently usable increments:

- Each increment MUST result in a buildable, runnable binary
  (`go build ./...` MUST succeed with zero errors).
- Feature branches MUST represent a single coherent capability
  addition (e.g., "add EvaluationLog generation", "add
  control-to-guideline mapping", "add Layer 2 tutorial")
  rather than cross-cutting refactors bundled with new
  behavior.
- Each increment MUST be demonstrable: a user can invoke the
  tool, provide input, and observe correct Gemara-conformant
  output.
- Capabilities MUST be usable independently. A user who only
  needs evaluation MUST NOT be forced to configure mapping
  functionality, and vice versa.

**Rationale**: Incremental delivery surfaces integration issues
early and keeps the project in a perpetually shippable state,
which is critical for building trust with the OpenSSF
community and downstream Gemara adopters.

### VI. Decision Documentation

Every non-trivial technical or process decision MUST be
recorded as an Architecture Decision Record (ADR):

- ADRs MUST follow a consistent format: Title, Status
  (Proposed, Accepted, Deprecated, Superseded), Context,
  Decision, Consequences.
- ADRs MUST be stored in a `docs/adrs/` directory using
  sequential numbering (e.g., `ADR-0001-<slug>.md`).
- The Context section MUST describe the problem or question
  that prompted the decision. The Decision section MUST state
  the chosen option. The Consequences section MUST document
  known trade-offs, risks, and follow-up actions.
- When a user or reviewer asks why a particular approach was
  chosen, the answer MUST reference the relevant ADR. If no
  ADR exists for the decision in question, one MUST be created
  before the explanation is considered complete.
- Superseded ADRs MUST NOT be deleted. Their status MUST be
  updated to "Superseded by ADR-XXXX" with a link to the
  replacement.

**Rationale**: Decisions made without written rationale become
tribal knowledge. In an open source project with rotating
contributors, undocumented decisions are relitigated
repeatedly, wasting review cycles and risking inconsistency.

### VII. Centralized Constants

Values used in multiple places or that may change over time
MUST be centralized. Specifically:

- Magic strings (e.g., `"EvaluationLog"`,
  `"https://github.com/gemaraproj/..."`, schema definition
  names) and magic numbers (e.g., timeout values, retry
  counts) MUST NOT appear inline within logic. These values
  MUST be defined in dedicated constant or configuration files
  (e.g., `internal/consts/consts.go`).
- Gemara schema type identifiers, relationship type strings,
  and artifact type values MUST each be defined once and
  referenced by name throughout the codebase.
- Updating a centralized value MUST propagate the change to
  every consumer automatically. A single logical change MUST
  NOT require search-and-replace across multiple files.

**Rationale**: In a tool that interacts with structured schemas,
hardcoded schema names and field values are a primary source of
silent breakage when upstream schemas evolve. Centralizing them
ensures that a Gemara version upgrade requires changes in
exactly one location.

### VIII. Composability

Programs and functions MUST do one thing and do it well.
Pac-Man's capabilities MUST be designed to work together and
with external tools:

- Each capability (evaluation, mapping, conversion) MUST be
  invocable as an independent CLI subcommand that reads from
  stdin or file arguments and writes to stdout.
- Output from one capability MUST be consumable as input for
  another without intermediate transformation. For example,
  the output of an evaluation MUST be a valid Gemara document
  that a mapping tool can ingest directly.
- Error output MUST be written to stderr, never mixed with
  data output on stdout.
- All structured output MUST default to YAML for human
  consumption and support `--output json` for programmatic
  consumption.

**Rationale**: Modular tools that compose via standard streams
integrate naturally into shell pipelines, CI/CD workflows, and
MCP tool chains. Monolithic designs that require in-process
coupling limit adoption and reuse.

### IX. Convention Over Configuration

Pac-Man MUST minimize the number of decisions a user needs to
make to accomplish a task:

- Every capability MUST be usable with zero configuration for
  the most common use case. Users MUST only need to specify
  configuration when deviating from established defaults.
- Default schema version, output format, and validation
  strictness MUST be pre-set to production-safe values.
- When a default is overridable, the override mechanism MUST
  be consistent across capabilities (e.g., flags, environment
  variables, or config file — not a mix of all three without
  a clear precedence order).

**Rationale**: New users attempting tutorials or first-time
evaluations will abandon the tool if they must configure it
before they can run it. Sensible defaults lower the barrier to
entry without sacrificing flexibility for advanced users.

## Repository Standard Files

The repository MUST contain the following standard files in the
root directory:

| File | Description | Standard |
|:---|:---|:---|
| `README.md` | Project overview, installation, and usage. | Markdown |
| `LICENSE` | Legal terms of use. | Apache License 2.0 |
| `CONTRIBUTING.md` | Guidelines for contributors: branching, DCO, review process. | Markdown; references this constitution |
| `CODE_OF_CONDUCT.md` | Community standards for participant behavior. | Contributor Covenant or equivalent |
| `SECURITY.md` | Security policy and vulnerability reporting instructions. | Markdown |
| `.github/` | GitHub configuration: issue templates, PR templates, CI workflows. | GitHub-native formats |

These files MUST be kept current. `CONTRIBUTING.md` MUST
reference this constitution as the authoritative source for
workflow rules. Any workflow rule in `CONTRIBUTING.md` that
conflicts with this constitution MUST be corrected to match.

## Technology & Integration Constraints

- **Supported Platforms**: Linux and macOS only. All tooling,
  scripts, and build processes MUST work on both platforms.
  Windows is explicitly out of scope. Platform-specific code
  MUST NOT be introduced; all automation MUST use
  technology-agnostic methods (POSIX-compatible shell, Go
  standard library, Makefile targets) that function identically
  on both supported operating systems.
- **Language**: Go (version 1.21 or later as documented;
  current module targets 1.26.1).
- **Schema Language**: CUE. Pac-Man MUST use the CUE Go SDK
  (`cuelang.org/go`) for schema loading and validation.
  Hand-rolled schema validation logic is prohibited when CUE
  can perform the check natively.
- **Gemara Dependency**: The Gemara schema repository
  (`github.com/gemaraproj/gemara`) MUST be referenced at a
  pinned version. The Go SDK
  (`github.com/gemaraproj/go-gemara`) SHOULD be used for
  type-safe document construction when available.
- **MCP Integration**: Pac-Man MUST integrate with the Gemara
  MCP server. MCP tool definitions MUST follow the MCP
  protocol specification. All MCP-exposed capabilities MUST
  also be available via CLI for offline and CI use. MCP server
  installation MUST be automated: the system resolves the
  latest gemara-mcp release, retrieves the SHA256 commit
  digest for that release, clones the repository (via SSH or
  HTTPS per user preference), checks out the pinned commit by
  digest (not by mutable tag), runs `make build`, and
  configures the built binary path in the OpenCode MCP
  configuration (`opencode.json`) as a local MCP server entry.
  SHA256 digest pinning MUST be used to prevent tag
  substitution attacks and ensure reproducible builds.
- **Document Formats**: Input and output documents MUST support
  both YAML and JSON. YAML MUST be the default human-readable
  format; JSON MUST be supported for programmatic consumption.
- **Makefile**: A `Makefile` MUST be present at the repository
  root and MUST serve as the single entry point for all
  build, test, lint, format, and generation commands. Direct
  invocation of `go build`, `go test`, `cue vet`, etc. in
  documentation or CI MUST reference the corresponding
  Makefile target (e.g., `make build`, `make test`,
  `make schema-check`). Contributors MUST NOT need to
  memorize tool-specific flags — the Makefile encapsulates
  them.
- **Tool Installation**: Homebrew MUST be documented as the
  preferred installation method for required and recommended
  tools on macOS and Linux. Specifically:
  - **CUE**: `brew install cue-lang/tap/cue` (required for
    schema validation via local `cue vet` fallback).
  - **Gitleaks**: `brew install gitleaks` (required for
    pre-commit secret scanning).
  - **OpenCode**: `brew install anomalyco/tap/opencode`
    (recommended AI development harness).
  Alternative installation methods (binary releases, install
  scripts, package managers) MUST also be documented for each
  tool. Homebrew is preferred because it provides a consistent
  installation and upgrade experience across both supported
  platforms and simplifies onboarding for contributors who may
  be unfamiliar with manual binary installation.
- **Dependencies**: Third-party dependencies beyond the CUE SDK,
  Gemara, and standard library MUST be justified before
  addition. The dependency MUST be actively maintained and
  carry a compatible open-source license. Hard forks of
  upstream dependencies MUST NOT be created. If an upstream
  fix is needed, the fix MUST be contributed back to the
  upstream project.
- **AI Development Harness**:
  [OpenCode](https://opencode.ai) is the preferred AI coding
  agent for all Pac-Man development and user-facing guided
  workflows. OpenCode serves as the single entry point through
  which contributors and users — regardless of role — discover
  what tools they need, how to install them, and how to get
  started with the Pac-Man project. Specifically:
  - OpenCode MUST be the recommended harness in all onboarding
    documentation (`README.md`, `CONTRIBUTING.md`, tutorials).
    Contributors SHOULD use OpenCode for code generation, code
    review, and interactive development sessions.
  - OpenCode MUST be configured with project-specific rules
    (via `.opencode/rules/` or `AGENTS.md`) that encode this
    constitution's principles, coding standards, and workflow
    requirements so that AI-assisted development automatically
    conforms to project governance.
  - OpenCode's MCP server support MUST be used to connect to
    the Gemara MCP server (`gemara-mcp`) when available,
    enabling AI-assisted sessions to access `get_lexicon`,
    `validate_gemara_artifact`, and `get_schema_docs` tools
    directly within the development workflow.
  - For end users interacting with Pac-Man's role-based
    tutorial engine, OpenCode MUST guide them through the
    complete onboarding flow: role identification, activity
    probing, schema version selection, tool installation
    (including CUE and the Gemara MCP server), and learning
    path navigation — tailored to the user's stated role and
    activities.
  - OpenCode is open source and runs on Linux and macOS,
    consistent with the project's supported platforms. It
    MUST be installable via the methods documented at
    `https://opencode.ai/docs` (install script, npm, Homebrew,
    or binary release).
- **License**: Apache License 2.0. All contributed code MUST be
  compatible with this license.

## Coding Standards

### General

- **SPDX License Headers**: Every source file (`.go`, `.cue`,
  `.sh`, `.yaml` where supported) MUST begin with an SPDX
  license identifier comment:
  ```go
  // SPDX-License-Identifier: Apache-2.0
  ```
- **Line Length**: Lines MUST be limited to 99 characters unless
  exceeding the limit demonstrably improves readability (e.g.,
  a long URL in a comment). This applies to Go, CUE, shell
  scripts, and Markdown prose.
- **End of File**: All files MUST end with a single newline
  character. This ensures clean version control diffs and
  adheres to POSIX standards.
- **Lint-Zero Policy**: Code MUST have zero lint issues
  according to the lint configuration defined in the
  repository. No trailing whitespace.
- **Pre-commit Hooks**: The repository MUST configure
  pre-commit hooks (via [pre-commit](https://pre-commit.com/)
  or equivalent). Hooks MUST enforce at minimum: formatting
  (`gofmt`, `cue fmt`), linting, Gitleaks secret scanning,
  and DCO sign-off verification.
- **Readability**: Variable and function names MUST clearly
  describe their intent. Code MUST NOT use obscure one-liners
  or language tricks that require deep mental parsing. Comments
  MUST explain *why* (intent/rationale), not *what* (syntax).

### Go

- **File Naming**: File names MUST use lowercase letters and
  underscores (e.g., `evaluation_log.go`).
- **Package Names**: Package names MUST be short, concise, and
  lowercase. MUST NOT use underscores or mixed caps.
- **Error Handling**: Errors MUST always be checked and handled.
  Errors SHOULD be returned to the caller when the current
  function cannot resolve them. Errors MUST NOT be silently
  discarded.
- **Formatting**: All Go source files MUST be formatted with
  `goimports` (which subsumes `gofmt`). Repositories MUST
  define Go-specific lint rules via `.golangci.yml` and run
  them in CI and locally before submitting a PR.

### Agent and Automation Awareness

Code agents (AI assistants, MCP tools, automated generators)
are subject to the same standards as human contributors.
OpenCode is the preferred AI coding agent for this project
(see Technology & Integration Constraints). Additionally:

- Before generating or modifying code, agents MUST read the
  repository's lint and formatter configuration files
  (`.golangci.yml`, `.pre-commit-config.yaml`, `cue.mod/`,
  `Makefile`) to understand the enforced rules.
- All generated or modified code MUST conform to these
  configurations. Agents MUST NOT introduce lint violations,
  formatting deviations, or missing license headers.
- If no lint configuration is present for a given file type,
  agents MUST follow the language-specific defaults defined
  in this constitution.
- OpenCode sessions MUST be initialized with `/init` to
  generate an `AGENTS.md` file that encodes project-specific
  context. This file MUST be committed to version control
  and kept current as the project evolves.
- OpenCode's rules and custom commands SHOULD be used to
  codify recurring development patterns (e.g., schema
  validation checks, Gemara artifact scaffolding) so that
  all contributors — human and AI — follow consistent
  procedures.

## Development Workflow

- **Branching**: Feature work MUST occur on a dedicated branch
  named with the pattern `<issue-number>-<short-description>`
  (e.g., `12-add-evaluation-output`). Direct commits to `main`
  are prohibited except for single-commit documentation fixes.
- **Commits**: Each commit MUST compile and pass all existing
  tests. Commit messages MUST follow Conventional Commits
  format (e.g., `feat:`, `fix:`, `refactor:`, `docs:`,
  `test:`). All commits MUST include a
  `Signed-off-by: Name <email>` trailer (DCO sign-off via
  `git commit -s`). All commits MUST be cryptographically
  signed (`git commit -S`) using a GPG, SSH, or S/MIME key
  registered with the contributor's GitHub account.
- **Atomic Pull Requests**: Every PR MUST address a single
  concern. Changes unrelated to the PR's stated purpose
  (incidental refactoring, formatting fixes, variable
  renames) MUST be submitted in a separate PR. Large
  multi-concern PRs MUST be split into focused submissions.
  PRs MUST include a description of what changed and why, and
  MUST reference the relevant spec or issue.
- **Review Requirement**: Every PR MUST receive approval from
  at least two reviewers who are not the PR author before it
  is eligible for merge. A single approval is insufficient
  regardless of the change size.
- **Upstream Fork Synchronization**: Before opening a pull
  request against an upstream repository (e.g., Gemara), the
  contributor MUST sync their fork's `main` branch with the
  upstream `main` branch and rebase their feature branch onto
  the updated `main`. Stale forks MUST NOT be used as the
  basis for upstream PRs. The upstream repository is the
  source of truth for all shared code and schemas.
- **Build Verification**: `make build` and `make lint` MUST
  pass with zero warnings before a PR is eligible for merge.
  `make test` MUST pass.
- **Schema Verification**: Any change that affects document
  output MUST include a `make schema-check` (or equivalent
  Makefile target wrapping `cue vet`) validation step
  confirming the output conforms to the target Gemara schema.
- **Secret Scanning (Gitleaks)**: Contributors MUST have
  [Gitleaks](https://github.com/gitleaks/gitleaks) installed
  locally. For contributors unfamiliar with Git or secret
  scanning, the project onboarding tutorial MUST guide them
  through Gitleaks installation and configuration. Once
  installed, Gitleaks MUST be configured to run automatically
  as a pre-commit hook. The hook MUST execute
  `gitleaks dir --staged` against staged change directories
  before each commit. Commits MUST be blocked if Gitleaks
  detects secrets. CI MUST also run Gitleaks as a pipeline
  step to catch any bypass of the local hook.

## Governance

This constitution is the authoritative source of project-level
rules. It takes precedence over all other project documents,
including but not limited to: implementation plans, task lists,
feature specifications, README files, inline code comments, and
PR templates. In any conflict between this constitution and a
subordinate document, this constitution prevails and the
subordinate document MUST be amended to conform.

- **Amendments**: Any change to this constitution MUST be
  proposed as a pull request with a rationale. The version
  MUST be incremented per semantic versioning (MAJOR for
  principle removal/redefinition, MINOR for new
  principles/sections, PATCH for wording clarifications).
  Every amendment MUST be accompanied by an ADR documenting
  the change rationale (per Principle VI).
- **Compliance Review**: Each pull request review MUST verify
  that the change does not violate any principle listed above.
  Violations MUST be resolved before merge or explicitly
  granted an exception documented in the PR description with
  a plan to remediate.
- **Upstream Source of Truth**: The upstream repositories for
  Gemara (`github.com/gemaraproj/gemara`) and its Go SDK
  (`github.com/gemaraproj/go-gemara`) are the authoritative
  sources for schema definitions and SDK behavior. When
  discrepancies exist between Pac-Man's assumptions and
  upstream behavior, upstream is correct. Contributors MUST
  sync their fork's `main` branch with the upstream `main`
  before creating any pull request targeting the upstream
  repository. When the Gemara project releases a new schema
  version, this constitution MUST be reviewed for alignment.
  Schema upgrades that introduce new layers, artifact types,
  or breaking changes MUST trigger a constitution amendment.
- **Information Accuracy**: When explaining project decisions,
  Git workflows, or upstream contribution procedures to users,
  all information provided MUST be accurate and verifiable.
  Explanations MUST reference the relevant ADR, upstream
  documentation, or Git official documentation. Speculative
  or unverified information MUST NOT be presented as fact.
- **Guidance File**: Runtime development guidance (coding
  patterns, CUE idioms, Gemara integration patterns) lives in
  the project README and spec documents. This constitution
  governs process and non-negotiable constraints; guidance
  documents govern recommended patterns.

**Version**: 1.3.0 | **Ratified**: 2026-03-12 | **Last Amended**: 2026-03-12
