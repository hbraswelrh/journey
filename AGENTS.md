# Pac-Man Project Context

Pac-Man is a role-based tutorial engine for the Gemara GRC
schema project. It guides users through Gemara tutorials based
on their job role and daily activities.

## Project Structure

- `cmd/pacman/` — Binary entry point
- `internal/consts/` — Centralized constants (no magic strings)
- `internal/mcp/` — MCP server detection, installation,
  client, version compatibility, OpenCode config management
- `internal/fallback/` — Local fallback (bundled lexicon, local
  CUE validation, cached schema docs)
- `internal/session/` — Session state management
- `internal/schema/` — Schema release fetching and version
  selection
- `internal/roles/` — Role identification, activity probing,
  custom profiles
- `internal/tutorials/` — Tutorial loading, learning path
  generation
- `internal/blocks/` — Content block extraction, drift
  detection, retrieval
- `internal/team/` — Team configuration, handoff detection,
  collaboration view
- `internal/authoring/` — Guided Gemara content authoring,
  validation, YAML/JSON output
- `internal/cli/` — CLI commands, setup flows, TUI rendering
- `specs/` — Feature specifications, plans, and task lists
- `docs/adrs/` — Architecture Decision Records

## Governance

The authoritative source of project rules is the constitution
at `.specify/memory/constitution.md`. All code must conform to
it. Key rules:

- **Go 1.26.1**, formatted with `goimports`, linted with
  `golangci-lint` (`.golangci.yml`)
- **SPDX headers** on all source files:
  `// SPDX-License-Identifier: Apache-2.0`
- **Line length** limited to 99 characters
- **No magic strings** — all constants in
  `internal/consts/consts.go`
- **TDD** — write failing tests before implementation
- **Conventional Commits** — `feat:`, `fix:`, `docs:`, etc.
- **Makefile** is the single entry point — use `make build`,
  `make test`, `make lint`, not raw `go` commands
- **Gemara lexicon** terms must be used consistently in all
  user-facing output

## MCP Server Integration

When the Gemara MCP server (`gemara-mcp`) is available, use it
for lexicon lookups (`get_lexicon`), artifact validation
(`validate_gemara_artifact`), and schema documentation
(`get_schema_docs`). When unavailable, fall back to local
equivalents in `internal/fallback/`.

## Schema Validation

All output artifacts must pass `cue vet -c -d '#<SchemaType>'`
against the pinned Gemara schema version. Use the schema
definition constants from `internal/consts/consts.go`.

## When Searching for Code

- Constants and magic strings: `internal/consts/consts.go`
- MCP-related logic: `internal/mcp/`
- Fallback behavior: `internal/fallback/`
- Session state: `internal/session/`
- Schema version selection: `internal/schema/`
- Role and activity logic: `internal/roles/`
- Tutorial loading and paths: `internal/tutorials/`
- Content block extraction: `internal/blocks/`
- Team collaboration: `internal/team/`
- Guided authoring: `internal/authoring/`
- CLI flows: `internal/cli/`
