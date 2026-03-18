---
title: Tailored Policy Writing
layer: 3
schema_version: v0.20.0
sections:
  - Policy Scope Definition
  - Metadata and Naming Conventions
  - RACI Contacts Structure
  - Importing Control and Guidance Catalogs
  - Policy Criteria and Assessment Requirements
  - Implementation Plan
  - Evaluation and Enforcement Timelines
  - Adherence Configuration
  - Non-Compliance Handling
  - CUE Validation
  - Cross-References to Other Layers
---

# Tailored Policy Writing

A Policy is a clearly-scoped set of rules based on an
organization's Risk Appetite. It lives at Gemara Layer 3
(Risk & Policy) and connects upstream catalogs (Guidance
at Layer 1, Controls and Threats at Layer 2) to downstream
enforcement (Sensitive Activities at Layer 4, Evaluations
at Layer 5).

This tutorial walks you through authoring a complete Policy
artifact that conforms to the `#Policy` schema in
`layer-3.cue`. By the end, you will have a validated Policy
document that:

- Defines what is in and out of scope
- Imports Control and Guidance Catalogs by reference
- Establishes RACI contacts for accountability
- Sets evaluation and enforcement timelines
- Configures adherence methods and non-compliance handling

The examples build on the CI/CD Pipeline Security scenario
used throughout the Gemara tutorials, importing the
`ACME.SEC` Threat Catalog and `ACME.SEC.CTRL` Control
Catalog.

## Policy Scope Definition

Every Policy begins with scope. Scope determines which
systems, technologies, geographies, and users are governed
by this Policy — and equally important, what is excluded.

The `scope` section uses `in` and `out` blocks, each with
three dimensions:

| Dimension | Purpose |
|-----------|---------|
| `technologies` | Systems and platforms governed |
| `geopolitical` | Jurisdictions and regions |
| `users` | Teams, roles, or user populations |

Example scope for a CI/CD pipeline security policy:

```yaml
scope:
  in:
    technologies:
      - "CI/CD Pipelines"
      - "Container Registries"
      - "Cloud Compute (Kubernetes)"
    geopolitical:
      - "United States"
      - "European Union"
    users:
      - "platform-engineers"
      - "developers"
      - "release-managers"
  out:
    technologies:
      - "Legacy On-Premises Build Systems"
    geopolitical:
      - "Regions without cloud presence"
    users:
      - "external-contractors"
```

**Why scope matters**: A Policy without clear boundaries
creates ambiguity about where rules apply. The `out` block
is as important as `in` — it prevents scope creep and
clarifies which systems need separate policies.

**Tailoring guidance**: Start with a narrow scope covering
one system or pipeline. Expand after the first evaluation
cycle confirms the policy is enforceable. If your
organization operates across multiple jurisdictions, the
`geopolitical` dimension ensures that region-specific
compliance requirements (e.g., GDPR, SOC 2) are explicit.

## Metadata and Naming Conventions

Policy metadata follows the same structure as all Gemara
artifacts. The `id` field uses the hierarchical naming
convention `ORG.PROJECT.COMPONENT.POL##`:

```yaml
title: "CI/CD Pipeline Security Policy"

metadata:
  id: "ACME.SEC.POL01"
  type: Policy
  gemara-version: "0.20.0"
  description: >-
    Security policy governing the automated build,
    test, and deployment pipeline for microservices.
    Establishes requirements for pipeline integrity,
    artifact provenance, and deployment controls.
  version: "1.0.0"
  author:
    id: security-team
    name: "Security Engineering"
    type: Human
  mapping-references:
    - id: "ACME.SEC.CTRL"
      title: "CI/CD Pipeline Security Control Catalog"
      version: "1.0.0"
      description: >-
        Control catalog for pipeline security
    - id: "ACME.SEC"
      title: "CI/CD Pipeline Security Threat Catalog"
      version: "1.0.0"
      description: >-
        Threat catalog for the CI/CD pipeline
```

**Naming convention**: The `POL` prefix distinguishes
Policy artifacts from Threat Catalogs (`THR`), Control
Catalogs (`CTRL`), and Guidance Catalogs (`GDN`). The
two-digit suffix (`01`) allows multiple policies per
component.

**Mapping references**: List every catalog that this
Policy imports or references. This creates a traceable
chain from guidance through controls to policy — the
foundation of Gemara's layered model.

**Author types**: Use `Human` for policies authored by
people, `Software Assisted` for policies generated or
drafted with tooling assistance.

## RACI Contacts Structure

The `contacts` section establishes accountability using
the RACI model: Responsible, Accountable, Consulted,
Informed. Each role is an array of contact objects.

```yaml
contacts:
  responsible:
    - name: "Platform Engineering"
      affiliation: "Engineering"
      email: "platform-eng@example.com"
    - name: "DevSecOps Team"
      affiliation: "Security"
  accountable:
    - name: "VP of Engineering"
      affiliation: "Engineering"
    - name: "CISO"
      affiliation: "Security"
  consulted:
    - name: "Legal & Compliance"
      affiliation: "Legal"
    - name: "Cloud Architecture"
      affiliation: "Engineering"
  informed:
    - name: "All Engineering"
      affiliation: "Engineering"
    - name: "Product Management"
      affiliation: "Product"
```

| Role | Definition |
|------|-----------|
| **Responsible** | Teams that execute the policy day-to-day |
| **Accountable** | Individuals with final authority; usually one per area |
| **Consulted** | Subject matter experts providing input before decisions |
| **Informed** | Stakeholders kept aware of policy changes and outcomes |

**Tailoring guidance**: For a small organization, one
person may fill multiple RACI roles. For larger
organizations, align RACI contacts with existing
organizational structures. The `affiliation` field helps
auditors identify which business unit owns each
responsibility.

## Importing Control and Guidance Catalogs

The `imports` section references external catalogs that
the Policy depends on. These references connect Layer 3
(Policy) to Layer 2 (Controls and Threats) and Layer 1
(Guidance).

```yaml
imports:
  catalogs:
    - reference-id: "ACME.SEC.CTRL"
    - reference-id: "ACME.SEC"
```

When you need to override or tighten specific Assessment
Requirements from an imported catalog, use
`assessment-requirement-modifications`:

```yaml
imports:
  catalogs:
    - reference-id: "ACME.SEC.CTRL"
      assessment-requirement-modifications:
        - id: "strict-tls"
          target-id: "ACME.SEC.CTRL.CTL02.AR01"
          modification-type: Override
          modification-rationale: >-
            Organization requires TLS 1.3 minimum,
            stricter than the catalog default of
            TLS 1.2.
          text: >-
            MUST use TLS 1.3 with certificate
            pinning for all inter-service
            communication.
```

**Modification types**: `Override` replaces the original
Assessment Requirement text. This creates a traceable
deviation — auditors can see exactly what changed and why.

**Why imports matter**: Imports establish the relationship
between your Policy and the Controls it enforces. Without
imports, a Policy is a standalone document with no
connection to the operational controls that implement it.

## Policy Criteria and Assessment Requirements

Policy criteria define the specific, verifiable conditions
that must be met. Each criterion maps to one or more
Assessment Requirements from imported Control Catalogs.

Assessment Requirements are tightly scoped, verifiable
conditions that must be satisfied and confirmed by an
evaluator. They are the atomic units of compliance
verification.

When authoring policy criteria, ask:

1. **What rule must be followed?** — The criterion
   description
2. **How will compliance be verified?** — The Assessment
   Requirement from your imported catalog
3. **What evidence is needed?** — Evidence requirements
   for the assessment plan

These connections flow downstream into Layer 5
(Evaluation) where each Assessment Requirement is
individually assessed and recorded in an Evaluation Log.

## Implementation Plan

The implementation plan defines how and when the Policy
becomes active. It has three components:

| Component | Purpose |
|-----------|---------|
| `notification-process` | How stakeholders learn about the policy |
| `evaluation-timeline` | When compliance assessment begins and ends |
| `enforcement-timeline` | When violations have consequences |

```yaml
implementation-plan:
  notification-process: |
    1. Policy published to internal wiki
    2. Team leads briefed in architecture review
    3. All-hands announcement with Q&A session
    4. Per-team onboarding sessions scheduled
  evaluation-timeline:
    start: 2025-07-01T00:00:00Z
    end: 2025-10-01T00:00:00Z
    notes: >-
      Initial 90-day evaluation phase. Automated
      checks enabled in advisory mode (warn, do not
      block). Baseline metrics collected.
  enforcement-timeline:
    start: 2025-10-01T00:00:00Z
    notes: >-
      Enforcement begins after evaluation baseline
      established. Deployments that violate policy
      criteria will be blocked by admission
      controller.
```

**Tailoring guidance**: The notification process should
match your organization's communication channels. The
evaluation period should be long enough to establish
baselines but short enough to maintain urgency — 60 to
90 days is typical.

## Evaluation and Enforcement Timelines

The evaluation and enforcement timelines are the two
most critical dates in a Policy. They define the
transition from "we are measuring" to "we are
enforcing."

**Evaluation timeline**:

| Field | Required | Description |
|-------|----------|-------------|
| `start` | Yes | ISO 8601 timestamp — when measurement begins |
| `end` | No | When the evaluation-only phase ends |
| `notes` | No | Context for the evaluation approach |

During evaluation, automated Assessments are rolled out
and baselines established. This phase produces the
Evaluation data needed for enforcement decisions.
Assessment results during this phase are advisory — they
surface non-compliance without blocking workflows.

**Enforcement timeline**:

| Field | Required | Description |
|-------|----------|-------------|
| `start` | Yes | When enforcement begins |
| `notes` | No | Enforcement conditions and escalation |

Enforcement begins after the evaluation baseline is
established. The gap between evaluation `end` and
enforcement `start` should be zero or minimal to
prevent compliance drift.

**Common pattern**: Set evaluation `end` equal to
enforcement `start` so there is no gap where
non-compliance goes unaddressed.

## Adherence Configuration

Adherence defines how Compliance with the Policy is
evaluated and enforced at runtime. It is the operational
heart of the Policy.

```yaml
adherence:
  evaluation-methods:
    - type: Automated
      description: >-
        CI pipeline checks validate artifact
        signatures, dependency provenance, and
        container image scanning results.
      executor:
        id: github-actions
        name: "GitHub Actions"
        type: Software
    - type: Gate
      description: >-
        Kubernetes admission controller blocks
        deployments that fail policy checks.
      executor:
        id: opa-gatekeeper
        name: "OPA Gatekeeper"
        type: Software
    - type: Manual
      description: >-
        Quarterly security review of pipeline
        configuration and access controls.
      executor:
        id: security-eng
        name: "Security Engineering"
        type: Human
```

**Evaluation method types**:

| Type | Description | When to use |
|------|-------------|-------------|
| `Manual` | Human-driven review | Complex judgment calls, architecture review |
| `Behavioral` | Observable process compliance | Training completion, workflow adherence |
| `Automated` | Tool-driven verification | CI/CD checks, scanning, signatures |
| `Autoremediation` | Automatic fix on violation | Auto-rollback, config drift correction |
| `Gate` | Blocks progression until compliant | Admission control, merge checks |

The `executor` field requires an entity struct with `id`,
`name`, and `type` (one of `Human`, `Software`, or
`Software Assisted`).

**Assessment plans** connect specific Assessment
Requirements to evaluation schedules:

```yaml
  assessment-plans:
    - requirement-id: "ACME.SEC.CTRL.CTL01.AR01"
      frequency: "per-commit"
      evaluation-methods:
        - type: Automated
          description: "Provenance attestation check"
      evidence-requirements:
        - "SLSA provenance attestation"
        - "Signed commit verification log"
    - requirement-id: "ACME.SEC.CTRL.CTL02.AR01"
      frequency: "per-deployment"
      evaluation-methods:
        - type: Automated
          description: "TLS validation"
        - type: Gate
          description: "Admission control check"
      evidence-requirements:
        - "TLS certificate chain validation log"
        - "Admission controller decision log"
```

**Tailoring guidance**: Match `frequency` to the natural
cadence of the activity being assessed. Code-level
checks use `per-commit`; deployment checks use
`per-deployment`; architectural reviews use `quarterly`.

## Non-Compliance Handling

Non-compliance handling defines the escalation path when
a violation is detected. This section bridges the gap
between detection (Layer 5 Evaluation) and response
(Layer 6 Enforcement).

```yaml
  enforcement-methods:
    - "Deployment blocked by admission controller"
    - "Pull request blocked by CI gate"
  non-compliance-handling:
    - >-
      Automated: Deployment blocked immediately by
      admission controller. No manual intervention
      required.
    - >-
      Notification: Security team notified via PagerDuty
      within 15 minutes of gate block.
    - >-
      Remediation: Development team opens remediation
      issue within 24 hours. Issue tracked in project
      board with SLA timer.
    - >-
      Escalation: Unresolved violations escalated to
      VP Engineering after 72 hours. CISO briefed on
      pattern violations monthly.
```

**Tailoring guidance**: Define escalation tiers that
match your organization's incident response process.
Automated enforcement (Gates, Autoremediation) should
handle the majority of violations. Manual escalation
is reserved for exceptions and pattern analysis.

**Key principle**: Non-compliance handling should be
proportional. A first-time configuration drift gets
auto-remediated. A pattern of bypassed gates triggers
human review. Systemic non-compliance triggers policy
review.

## CUE Validation

Validate your completed Policy against the Gemara schema
before publishing:

```bash
cue vet -c -d '#Policy' \
  github.com/gemaraproj/gemara@latest \
  your-policy.yaml
```

If the Gemara MCP server is available, use the
`validate_gemara_artifact` tool for validation without
installing CUE locally:

```
Tool:       validate_gemara_artifact
Parameters:
  artifact_content: <your YAML content>
  definition:       #Policy
```

**Common validation errors**:

| Error | Cause | Fix |
|-------|-------|-----|
| `metadata.type: incomplete value` | Missing `type` field | Add `type: Policy` to metadata |
| `scope.in: incomplete value` | Missing scope dimensions | Add at least one dimension to `scope.in` |
| `contacts: incomplete value` | Missing RACI contacts | Add at least `responsible` and `accountable` |
| `implementation-plan: incomplete value` | Missing timelines | Add both `evaluation-timeline` and `enforcement-timeline` |
| `adherence: incomplete value` | Missing evaluation methods | Add at least one entry in `evaluation-methods` |

**Iterative validation**: Validate after completing each
major section rather than waiting until the end. This
catches structural issues early. The Pac-Man guided
authoring flow validates at each step automatically.

## Cross-References to Other Layers

A Policy does not exist in isolation. It connects to
every other layer in the Gemara model:

| Flow | Direction | Relationship |
|------|-----------|-------------|
| Layer 1 → Layer 3 | Upstream | Guidance Catalogs are referenced by Policy documents |
| Layer 2 → Layer 3 | Upstream | Control and Threat Catalogs feed Policy evaluation criteria |
| Layer 3 → Layer 4 | Downstream | Policy governs which Controls apply to Sensitive Activities |
| Layer 3 → Layer 5 | Downstream | Policy drives Evaluation Log Assessments |

**Upstream connections**: Your Policy's `imports` section
references Control Catalogs (Layer 2). The
`mapping-references` in metadata reference both Threat
and Control Catalogs. These imports create the
traceability chain that auditors follow.

**Downstream connections**: Once published, your Policy
becomes the input for:

1. **Sensitive Activity analysis** (Layer 4): Teams
   identify which activities are governed by this Policy
   and what Controls apply.
2. **Evaluation** (Layer 5): Evaluators assess each
   Assessment Requirement defined in the imported
   Controls, recording findings in Evaluation Logs.
3. **Enforcement** (Layer 6): Preventive and Remediative
   Enforcement actions implement the non-compliance
   handling defined in your Policy.
4. **Audit** (Layer 7): Auditors review the complete
   chain from Policy through Evaluation findings to
   enforcement records.

**What's next**: After publishing your Policy, use the
Evaluation Log schema (`#EvaluationLog`) at Layer 5 to
record Assessment results. Each Assessment Requirement
from your imported Control Catalog becomes an evaluation
target.
