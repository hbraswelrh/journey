# Project Structure

Gemara User Journey is organized as a Go CLI application with a React
web frontend. The Go backend contains all domain logic;
the web frontend mirrors the role discovery flow in a
browser.

```
journey/
  cmd/journey/
    main.go                # Application entry point
  cmd/genwebdata/
    main.go                # TypeScript data generator for web UI
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
  web/
    src/
      components/          # React components (RoleSelection,
                           #   ActivityProbe, Results, etc.)
      generated/           # Auto-generated TypeScript from Go
                           #   constants
      lib/                 # Client-side role matching logic
  go.mod                   # Go module definition
  Makefile                 # Single entry point for build/test/lint
  specs/                   # Feature specifications
  docs/
    adrs/                  # Architecture Decision Records
    tutorials/             # Tailored tutorial content
  .github/                 # Issue templates, PR template, CI
  .specify/
    memory/
      constitution.md      # Project constitution (authoritative)
    templates/             # Specification and planning templates
```

## Architecture Decision Records

Every non-trivial technical or process decision is recorded
as an Architecture Decision Record (ADR) in `docs/adrs/`.
ADRs follow the format: Title, Status (Proposed, Accepted,
Deprecated, Superseded), Context, Decision, Consequences.
When a decision is questioned, the relevant ADR provides
the authoritative rationale.

---

[Back to README](../README.md)
