# Quickstart: Refocus Pac-Man as Tutorial Guide

**Branch**: `002-tutorial-guide-focus`
**Date**: 2026-03-17

## Overview

This guide walks through the new Pac-Man user experience
after implementing the tutorial guide refocus. The primary
change is that Pac-Man no longer asks users to select a
schema version or author artifacts directly. Instead, it
guides users from role identification through tutorial
walkthrough to a clear handoff to the MCP server.

## Prerequisites

- Go 1.26.1 installed
- OpenCode installed (`brew install anomalyco/tap/opencode`)
- CUE installed (`brew install cue-lang/tap/cue`)
- Pac-Man built (`make build`)
- Gemara MCP server built and configured (optional but
  recommended)

## The New Flow

### Step 1: Launch Pac-Man

```bash
opencode
```

Tell OpenCode your role:

> "I'm a Security Engineer working on CI/CD pipeline
> security. Help me get started with Gemara."

### Step 2: Automatic Version Selection

Pac-Man automatically resolves the latest Gemara release.
No prompt is displayed. You will see:

```text
Schema version: v0.20.0 (latest)
```

If experimental schemas exist, a note appears:

```text
i Note: The following schemas are experimental:
  sensitive_activity, data_collection
```

### Step 3: Activity Identification

Pac-Man identifies your relevant Gemara layers:

```text
Your Gemara Layers:
  L2: Threats & Controls (strong)
  L4: Sensitive Activities (strong)
  L1: Guidance (inferred from role)

Recommended Artifact Outputs:
  • Threat Catalog — A catalog of threats to a specific
    component, organized by capability, with severity
    and likelihood assessments.
    → MCP Wizard: threat_assessment

  • Control Catalog — A catalog of security controls that
    mitigate identified threats, with assessment
    requirements and evidence criteria.
    → MCP Wizard: control_catalog

  • Guidance Catalog — A structured catalog of standards,
    best practices, and regulatory requirements.
    → Collaborative authoring with MCP resources
```

### Step 4: Tutorial Walkthrough

Navigate through tutorials tailored to your role. Each
section includes:

- **Why this matters for you**: Role-specific relevance
- **How you will use this**: Application to your daily
  work
- **What you will learn**: Learning outcomes

Sections matching your activity keywords ("CI/CD",
"pipeline security") are highlighted as focus areas.

### Step 5: Handoff to OpenCode

After completing a tutorial, Pac-Man presents the handoff
summary directing you to OpenCode:

```text
────────────────────────────────────────────
┃ Ready to Author: Threat Catalog
┃
┃ Schema:  #ThreatCatalog
┃ Version: v0.20.0
┃ Wizard:  threat_assessment
┃
┃ Available in OpenCode:
┃   Tools:     validate_gemara_artifact
┃   Resources: gemara://lexicon
┃              gemara://schema/definitions
┃   Prompts:   threat_assessment
┃
┃ Key Decisions:
┃   1. Component scope and boundaries
┃   2. MITRE ATT&CK alignment
┃   3. Import source selection
┃
┃ Preparation Checklist:
┃   • Identify the component to assess
┃   • Determine scope boundaries
┃   • Decide on catalog import source
┃   • Consider MITRE ATT&CK alignment
┃
┃ Next: Open an OpenCode Session
┃
┃   Launch opencode and use the
┃   threat_assessment prompt with the
┃   gemara-mcp server to begin guided
┃   authoring.
────────────────────────────────────────────
```

### Step 6: Author in OpenCode with gemara-mcp

Launch OpenCode:

```bash
opencode
```

In your OpenCode session, tell the AI:

> "Run the threat_assessment wizard for my CI/CD pipeline."

The gemara-mcp server provides the tools and resources
for authoring:

- **Wizard prompts**: `threat_assessment` and
  `control_catalog` for guided artifact creation
- **Resources**: `gemara://lexicon` for terminology and
  `gemara://schema/definitions` for schema reference
- **Validation**: `validate_gemara_artifact` to check
  your artifact against the CUE schema

After authoring, validate the artifact:

> "Validate my artifact against the #ThreatCatalog schema."

## What Changed from Before

| Before | After |
|--------|-------|
| User prompted to choose Stable or Latest version | Latest version auto-selected |
| User could author artifacts directly in Pac-Man | Pac-Man directs to OpenCode + gemara-mcp |
| No post-tutorial summary | Handoff summary with MCP tools/resources list |
| Version switching available mid-session | Version switching deferred (planned future) |
| Authoring engine and wizards replicated in Pac-Man | Clear boundary: Pac-Man = learn, OpenCode = author |

## Verifying the Changes

### Run the doctor check

```bash
./pacman --doctor
```

Confirms environment setup including MCP server
availability.

### Run the test suite

```bash
make test
```

All tests pass including new tests for:
- `AutoSelectLatest` in `internal/schema/`
- `ArtifactRecommendations` in `internal/roles/`
- `BuildHandoffSummary` and `RenderHandoffSummary` in
  `internal/cli/`
- Setup flow bypass verification

### Verify no version prompt

Run the setup flow and confirm that no version selection
prompt appears. The session status should show the latest
version without user intervention.

### Verify doctor command unchanged

```bash
./pacman --doctor
```

Confirm all environment checks still pass and output is
unchanged from before the feature implementation.

### Verify handoff summary

Complete a tutorial and confirm that the handoff summary
appears with the correct artifact type, MCP prompt,
available MCP tools/resources, and preparation checklist.
Verify it directs the user to OpenCode with the
gemara-mcp server.
