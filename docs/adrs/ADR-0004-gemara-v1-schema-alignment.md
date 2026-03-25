# ADR-0004: Gemara v1 Schema Alignment

**Status**: Accepted
**Date**: 2026-03-23
**Deciders**: Project maintainers
**Feature**: 002-tutorial-guide-focus

## Context

The upstream Gemara project (`gemaraproj/gemara`) has
evolved its schema catalog to include artifact types
across all seven layers of its model. Gemara User Journey's internal
constants, keyword mappings, authoring templates, and
documentation referenced only a subset of the available
schemas:

- **Layer 1**: GuidanceCatalog only
- **Layer 2**: ThreatCatalog, ControlCatalog
- **Layer 3**: Policy only
- **Layers 5-7**: EvaluationLog only; Layers 6 and 7 had
  no artifact types and used outdated names ("Data
  Collection", "Reporting")

The Gemara v1 model defines the following additional
artifact types with published CUE schemas:

| Artifact | Layer | Schema Status |
|----------|-------|---------------|
| VectorCatalog | L1 | Experimental |
| PrincipleCatalog | L1 | Experimental |
| CapabilityCatalog | L2 | Stable |
| RiskCatalog | L3 | Experimental |
| EnforcementLog | L6 | Experimental |
| AuditLog | L7 | Experimental |

Additionally, the canonical layer names in the Gemara
model differ from what Gemara User Journey was using:

| Layer | Gemara User Journey (before) | Gemara v1 (canonical) |
|-------|------------------|-----------------------|
| 1 | Guidance | Vectors & Guidance |
| 5 | Evaluation | Intent & Behavior Evaluation |
| 6 | Data Collection | Preventive & Remediative Enforcement |
| 7 | Reporting | Audit & Continuous Monitoring |

Without these updates, Gemara User Journey could not route users to
tutorials or recommend artifacts for activities involving
attack vectors, secure design principles, system
capabilities, organizational risk, enforcement actions,
or audit results.

## Decision

### Add All v1 Artifact Types

Expand the centralized constants, schema maps, authoring
templates, and keyword mappings to include all 12 artifact
types defined by the Gemara v1 schemas:

| Layer | Artifact Types |
|-------|---------------|
| L1 | GuidanceCatalog, VectorCatalog, PrincipleCatalog |
| L2 | ThreatCatalog, ControlCatalog, CapabilityCatalog |
| L3 | Policy, RiskCatalog |
| L4 | (none — sensitive activities, no schema) |
| L5 | EvaluationLog |
| L6 | EnforcementLog |
| L7 | AuditLog |
| Cross | MappingDocument |

### Update Layer Names

Rename layer constants and display strings to match the
canonical Gemara v1 model. Backward-compatible aliases
are provided so existing references continue to compile:

- `LayerVectorsGuidance = 1` (alias: `LayerGuidance`)
- `LayerEnforcement = 6` (alias: `LayerDataCollection`)
- `LayerAudit = 7` (alias: `LayerReporting`)

### Expand Activity Keywords

Add keywords for the new artifact types and layer
concepts so that user activity descriptions route
correctly:

- **L1**: attack vectors, vectors, MITRE ATT&CK, secure
  design principles, principles, vector catalog,
  principle catalog, guidance catalog
- **L2**: capability catalog, system capabilities
- **L3**: risk catalog, risk categories, risk severity
- **L5**: intent evaluation, behavior evaluation
- **L6**: enforcement, enforcement log, preventive
  enforcement, remediative enforcement, admission
  controller
- **L7**: audit, audit log, continuous monitoring, audit
  results

The keyword "audit" was moved from L5 to L7 to match the
Gemara model's distinction between evaluation (L5) and
audit (L7).

### Add Authoring Templates

Each new artifact type receives a guided authoring
template with step definitions, example values, and
section constants consistent with the CUE schema
structure:

- `vectorsStep()` — id, description, group
- `principlesStep()` — id, title, description, group
- `risksStep()` — id, description, severity, group
- `actionsStep()` — disposition, method ref, message
- `auditResultsStep()` — id, type, description

### No Functional Changes

This decision does not change runtime behavior, CLI
flow, or MCP integration. It expands the available
options so that users describing activities related to
vectors, principles, capabilities, risks, enforcement,
or audit are routed to the correct layers and offered
the correct artifact recommendations.

## Consequences

### Benefits

- Users describing activities like "document MITRE
  ATT&CK techniques" or "create a risk catalog" are now
  routed to the correct Gemara layer and offered the
  correct artifact type.
- The Auditor role now correctly maps to L7 (Audit) in
  addition to L5 (Evaluation) and L3 (Risk & Policy).
- Layer 6 and Layer 7 are no longer empty in the
  `LayerArtifacts` map, which means enforcement and
  audit activities produce artifact recommendations
  instead of silently returning nothing.
- Layer names in user-facing messages (e.g.,
  `MissingLayerMessage`) now match the canonical Gemara
  terminology.
- AGENTS.md documentation is aligned with the upstream
  Gemara model, so OpenCode sessions produce accurate
  guidance.

### Risks and Trade-offs

- The `SupportedArtifactTypes()` list grew from 6 to 12
  entries, which makes the artifact selection prompt
  longer. This is acceptable because the selection is
  filtered by role and activity context in practice.
- Several experimental schemas (VectorCatalog,
  PrincipleCatalog, RiskCatalog, EnforcementLog,
  AuditLog) may change before reaching Stable status.
  Users will see experimental schema warnings from the
  existing version selection logic when these are used.
- The keyword "audit" moving from L5 to L7 changes
  routing for users who previously typed "audit" and
  were routed to evaluation. This is intentional — the
  Gemara model distinguishes evaluation (L5, inspection
  of activities) from audit (L7, formal review of
  compliance posture).

### Files Changed

| File | Nature of Change |
|------|-----------------|
| `internal/consts/consts.go` | New constants, updated maps |
| `internal/roles/activities.go` | New keywords, updated schema map |
| `internal/authoring/model.go` | Expanded schema map and supported types |
| `internal/authoring/engine.go` | 6 new templates, 5 new step functions |
| `internal/tutorials/path.go` | Updated layer names |
| `AGENTS.md` | Full v1 alignment |
| `docs/tutorials/tailored-policy-writing.md` | Cross-references updated |
| Test files | Updated expected counts and selection indices |
