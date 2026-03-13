# ADR-0001: OpenCode as Preferred AI Development Harness

**Status**: Accepted
**Date**: 2026-03-12
**Deciders**: Project maintainers
**Constitution Version**: 1.3.0

## Context

Pac-Man is a role-based tutorial engine that guides users —
across diverse job roles (Security Engineers, Compliance
Officers, Developers, CISOs, Auditors, and others) — through
the Gemara project's seven-layer GRC model. The tool must
onboard users who may have no prior knowledge of Gemara, CUE
schemas, Git workflows, or the specific tools required for
their role. At the same time, contributors developing Pac-Man
need an AI-assisted development workflow that automatically
conforms to the project's constitution, coding standards, and
Gemara schema constraints.

The project needed to decide:

1. Whether to adopt a specific AI coding agent as the preferred
   development harness, or leave the choice unspecified.
2. If adopting one, which agent best fits the project's
   requirements: open source, cross-platform (Linux and macOS),
   terminal-native, MCP server support for Gemara integration,
   and configurable project rules.

Alternatives considered:

- **No preferred agent**: Contributors use whatever AI tool they
  prefer. This creates inconsistency — different agents produce
  code with different formatting, structure, and adherence to
  project rules. No single configuration can enforce the
  constitution across all tools.
- **Custom agent/harness**: Build a project-specific AI
  integration. This would require significant development effort
  unrelated to Pac-Man's core mission and would need ongoing
  maintenance as AI tooling evolves.
- **Other existing agents**: Several commercial and open source
  AI coding agents exist. Most are either closed source,
  IDE-specific (not terminal-native), or lack MCP server
  support for connecting to the Gemara MCP server.

## Decision

Adopt [OpenCode](https://opencode.ai) as the preferred AI
development harness for both Pac-Man development and user-facing
guided workflows.

OpenCode is selected because it satisfies all project
requirements:

- **Open source**: Aligns with the project's open source values
  and Apache 2.0 licensing.
- **Terminal-native**: Runs in the terminal on Linux and macOS,
  matching the project's supported platforms. Also available as
  a desktop app and IDE extension.
- **MCP server support**: Can connect to the Gemara MCP server
  (`gemara-mcp`), enabling AI-assisted sessions to access
  `get_lexicon`, `validate_gemara_artifact`, and
  `get_schema_docs` tools directly.
- **Project-configurable**: Supports project-specific rules
  (`.opencode/rules/`), custom commands, and `AGENTS.md` files
  that can encode the constitution's principles, coding
  standards, and workflow requirements.
- **Role-agnostic onboarding**: Can guide any user — regardless
  of role — through tool installation, role identification,
  schema version selection, and learning path navigation.

OpenCode is integrated at two levels:

1. **Contributor-facing**: Recommended for all development
   activities (code generation, review, interactive sessions).
   Project rules encode the constitution automatically.
2. **User-facing**: Serves as the interface through which end
   users interact with Pac-Man's role-based tutorial engine,
   guided through onboarding without requiring prior knowledge
   of the project.

## Consequences

### Benefits

- Single entry point for all roles reduces onboarding friction.
- Project rules enforce constitution compliance automatically
  in AI-assisted development.
- MCP server integration provides direct access to Gemara tools
  within the development and user workflow.
- Open source and terminal-native aligns with project values
  and platform requirements.

### Risks and Trade-offs

- **Tool dependency**: The project's onboarding and guided
  experience is optimized for OpenCode. Users who decline
  OpenCode still have CLI access but lose the guided experience.
  This is acceptable because the CLI remains fully functional
  (per FR-033).
- **Maintenance burden**: OpenCode configuration files
  (`.opencode/rules/`, `AGENTS.md`, custom commands) must be
  maintained alongside the constitution. Changes to the
  constitution require corresponding updates to OpenCode
  configuration.
- **Evolution risk**: If OpenCode's development direction
  diverges from the project's needs, migration to an
  alternative would require updating onboarding documentation,
  project rules, and the constitution. This risk is mitigated
  by OpenCode being open source.

### Follow-up Actions

- Constitution amended to v1.3.0 (completed).
- Spec updated with FR-033 and FR-034 (completed).
- OpenCode project configuration (rules, commands, `AGENTS.md`)
  to be created during US1 implementation.
- `README.md` and `CONTRIBUTING.md` to reference OpenCode as
  the preferred harness during repository standard files setup.
