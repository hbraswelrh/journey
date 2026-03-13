# Constitution Rules for AI Agents

These rules encode the Pac-Man constitution
(`.specify/memory/constitution.md`) for automated enforcement
during AI-assisted development.

## Before Writing Code

1. Read `.golangci.yml` and `.pre-commit-config.yaml` to
   understand enforced lint and format rules.
2. Check `internal/consts/consts.go` for existing constants
   before introducing any string literal, URL, or numeric
   value that could be shared.
3. Verify the target file has an SPDX license header:
   `// SPDX-License-Identifier: Apache-2.0`

## Code Standards

- Lines must not exceed 99 characters.
- All Go files formatted with `goimports`.
- No magic strings or numbers inline — define in
  `internal/consts/consts.go`.
- Errors must always be checked and handled; never silently
  discarded.
- Test files live alongside source using `_test.go` convention.

## TDD Workflow

- Write a failing test before writing implementation.
- Use positive and negative fixtures for schema validation.
- Test files must have SPDX headers.

## Build and Validate

- Use `make build`, `make test`, `make lint` — never raw
  `go build` or `go test`.
- Schema validation uses `make schema-check`.
- All commands must pass with zero errors and zero warnings
  before a change is considered complete.

## Commit Standards

- Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`,
  `test:`
- Each commit must compile and pass all tests.
- DCO sign-off required (`git commit -s`).

## Gemara Lexicon

- Use Gemara lexicon terms in all user-facing output.
- Do not redefine or use alternate meanings for controlled
  vocabulary.
- When the Gemara MCP server is available, source lexicon from
  `get_lexicon`.
