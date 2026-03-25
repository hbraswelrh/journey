# Specification Quality Checklist: Role-Based Tutorial Engine

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-03-12
**Updated**: 2026-03-12 (post OpenCode harness integration amendment)
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- All 16/16 items passed (fifth iteration after OpenCode
  harness integration amendment).
- New US1 (Gemara MCP Server Setup) added as P1 with 6
  acceptance scenarios. All subsequent stories reprioritized
  (P2-P6).
- 5 new functional requirements added (FR-026 through FR-030):
  MCP installation prompt, binary/Podman installation methods,
  MCP-preferred data sourcing, local fallback behavior, and
  auto-detection of existing installations.
- FR-010 (validation) and FR-011 (lexicon) updated to reference
  MCP server tools as preferred sources with local fallbacks.
- 1 new entity added (MCP Server Connection).
- 1 new edge case added (MCP server mid-session disconnection).
- 2 new success criteria added (SC-013 installation within 5
  minutes, SC-014 MCP-preferred sourcing with fallback).
- Cross-references updated throughout (US numbering, US
  back-references in FR-024, FR-030, and narrative text).
- 2 new functional requirements added (FR-033, FR-034):
  OpenCode as preferred AI development harness and guided
  onboarding interface, OpenCode-specific project configuration
  for role-based flows.
- US1 narrative updated to reference OpenCode as the interface
  through which users launch Gemara User Journey.
- New assumption added: users interact with Gemara User Journey through
  OpenCode, with CLI fallback for users who choose not to use
  it.
- Constitution updated to v1.3.0 with OpenCode in Technology &
  Integration Constraints and Agent and Automation Awareness.
- FR-035 added: Homebrew as preferred installation method for
  CUE, Gitleaks, and OpenCode with alternative methods
  documented. Constitution updated with Tool Installation
  subsection in Technology & Integration Constraints.
- FR-027 rewritten: gemara-mcp installation changed from
  manual guidance to automated pipeline (resolve latest
  release, retrieve SHA256 commit digest, clone via SSH/HTTPS
  at pinned digest, `make build`, configure `opencode.json`
  with built binary path). SHA256 digest pinning prevents tag
  substitution attacks and ensures reproducible builds. US1
  acceptance scenarios 2-3 updated. Constitution MCP
  Integration bullet updated. `internal/mcp/config.go` added
  to plan for opencode.json management. Tasks renumbered
  (T001-T054).
