# Contributing to Pac-Man

Thank you for your interest in contributing to Pac-Man. This
document provides guidelines and workflow requirements for
contributors.

## Authoritative Source

The [Pac-Man Constitution](.specify/memory/constitution.md)
is the authoritative source for all project-level rules. Any
workflow rule in this document that conflicts with the
constitution MUST be corrected to match. When in doubt, the
constitution prevails.

## Prerequisites

Before contributing, ensure you have the following installed:

- **Go 1.26.1** or later
- **CUE**: `brew install cue-lang/tap/cue` (or
  [alternative methods](https://cuelang.org/docs/install/))
- **Gitleaks**: `brew install gitleaks` (required for
  pre-commit secret scanning)
- **OpenCode** (recommended): `brew install anomalyco/tap/opencode`
  — the preferred AI development harness for this project
- **goimports**: `go install golang.org/x/tools/cmd/goimports@latest`

## Development Setup

1. Fork and clone the repository.
2. Install pre-commit hooks:
   ```sh
   pre-commit install
   ```
3. Verify your setup:
   ```sh
   make build
   make test
   make lint
   ```

## Development Workflow

### Branching

Feature work MUST occur on a dedicated branch named with the
pattern `<issue-number>-<short-description>` (e.g.,
`12-add-evaluation-output`). Direct commits to `main` are
prohibited except for single-commit documentation fixes.

### Commits

- Each commit MUST compile and pass all existing tests.
- Commit messages MUST follow
  [Conventional Commits](https://www.conventionalcommits.org/)
  format: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`.
- All commits MUST include a DCO sign-off:
  ```sh
  git commit -s -S -m "feat: add new feature"
  ```
- All commits MUST be cryptographically signed (`-S`) using a
  GPG, SSH, or S/MIME key registered with your GitHub account.

### Pull Requests

- Every PR MUST address a single concern. Unrelated changes
  (incidental refactoring, formatting fixes, variable renames)
  MUST be submitted in a separate PR.
- PRs MUST include a description of what changed and why, and
  MUST reference the relevant spec or issue.
- Every PR MUST receive approval from at least two reviewers
  before merge.
- Sync your fork's `main` with upstream and rebase your feature
  branch before opening a PR.

### Build Verification

Before submitting a PR, ensure all checks pass:

```sh
make build    # Zero warnings
make test     # All tests pass
make lint     # Zero lint issues
```

## Coding Standards

### SPDX License Headers

Every source file MUST begin with:

```go
// SPDX-License-Identifier: Apache-2.0
```

### Formatting and Linting

- Go source files MUST be formatted with `goimports`.
- Lines MUST be limited to 99 characters.
- All files MUST end with a single newline.
- Code MUST have zero lint issues per `.golangci.yml`.

### No Magic Strings

All constants MUST be defined in `internal/consts/consts.go`
and referenced by name. No string or numeric literals should
appear inline in business logic.

### Test-Driven Development

Write failing tests before implementation. All new
functionality MUST have corresponding tests. Run tests with
the race detector:

```sh
make test
```

### Makefile as Single Entry Point

Use `make build`, `make test`, `make lint` — not raw `go`
commands. The Makefile is the single entry point for all
build operations.

## Secret Scanning

Gitleaks MUST be configured to run as a pre-commit hook. The
hook executes `gitleaks dir --staged` against staged changes
before each commit. Commits are blocked if secrets are
detected.

## AI-Assisted Development

[OpenCode](https://opencode.ai) is the recommended AI
development harness. Contributors SHOULD use OpenCode for
code generation, code review, and interactive development
sessions. OpenCode sessions are initialized with project
context via the `AGENTS.md` file at the repository root.

## Project Structure

```
cmd/pacman/         Binary entry point
internal/
  consts/           Centralized constants
  mcp/              MCP server detection, installation, client
  fallback/         Local fallback (bundled lexicon, CUE validation)
  session/          Session state management
  schema/           Schema release fetching and version selection
  roles/            Role identification and activity probing
  tutorials/        Tutorial loading and learning path generation
  blocks/           Content block extraction and drift detection
  team/             Team configuration and collaboration view
  authoring/        Guided Gemara content authoring
  cli/              CLI commands, setup flows, and TUI rendering
specs/              Feature specifications, plans, and task lists
docs/adrs/          Architecture Decision Records
```

## Questions

If you have questions about contributing, open an issue or
reach out through the project's communication channels.
