# ADR-0003: Version Selection Deferral

**Status**: Accepted
**Date**: 2026-03-17
**Deciders**: Project maintainers
**Feature**: 002-tutorial-guide-focus

## Context

Gemara User Journey's setup flow previously presented users with an
interactive prompt to choose between "Stable" and "Latest"
schema versions. This decision point served users who
needed to pin to a specific schema stability level.

However, this feature introduced friction during onboarding:

1. **Unnecessary complexity**: Users focused on learning
   Gemara through tutorials do not need to understand the
   difference between Stable and Latest schema versions
   before they begin.

2. **Premature decision**: The version choice is only
   meaningful when authoring artifacts. By the time users
   reach the authoring phase (in OpenCode with the
   gemara-mcp server), they have more context to make
   informed decisions about schema versions.

3. **Current state of upstream**: The Gemara repository may
   not yet have distinct Stable vs Latest versions that
   warrant user choice, making the prompt confusing when
   both options resolve to the same release.

4. **Role clarity**: Gemara User Journey is being refocused as a
   tutorial guide. Version management is a configuration
   concern better handled by the MCP server or OpenCode
   environment, not by a learning tool.

## Decision

### Auto-Select Latest

Replace the interactive `RunVersionSelection` prompt in
`RunSetup` (setup.go) with a non-interactive call to
`schema.AutoSelectLatest`, which:

1. Fetches or loads cached releases
2. Determines the latest version
3. Applies it to the session automatically
4. Returns experimental schema warnings for display

The user sees the selected version in the setup output
but is never asked to choose.

### Preserve Code for Re-Enablement

The `RunVersionSelection` function and
`VersionPromptConfig` struct in `version_prompt.go` are
retained intact. They are not deleted. A code comment
documents the bypass and provides re-enablement
instructions.

To re-enable interactive version selection, a developer
replaces the `AutoSelectLatest` call in `setup.go` with
a call to `RunVersionSelection` using the existing
`VersionPromptConfig`.

## Consequences

### Benefits

- Users start tutorials faster with one fewer decision
  point during setup.
- The setup flow is simpler and more predictable for
  non-technical users (compliance officers, CISOs, policy
  authors).
- Experimental schema warnings are still surfaced after
  auto-selection and in the handoff summary, so users are
  informed before they begin authoring.

### Risks and Trade-offs

- Users who specifically need a Stable version cannot
  select it during setup. They must rely on the latest
  release being the version they need, or manually
  configure the schema version through other means.
- If a breaking change is introduced in a new Latest
  release, users will not be warned before auto-selection.
  The post-selection experimental schema warning
  partially mitigates this.

### Re-Enablement Path

When the upstream Gemara repository has distinct Stable
and Latest releases with materially different schema
stability, re-enabling version selection requires:

1. Replace the `AutoSelectLatest` call in `setup.go`
   with `RunVersionSelection` using
   `VersionPromptConfig`.
2. Remove the bypass comment from `version_prompt.go`.
3. Update this ADR's status to "Superseded" with a
   reference to the new ADR.

Estimated effort: under 1 hour.
