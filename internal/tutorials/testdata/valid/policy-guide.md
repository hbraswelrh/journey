---
title: Policy Guide
layer: 3
schema_version: v0.20.0
sections:
  - Policy Structure
  - Implementation Plan
  - Evaluation Timeline
  - Enforcement Timeline
  - Adherence
---

# Policy Guide

This tutorial walks through creating an organizational
Policy Document using the Gemara schema. The document
conforms to the Policy schema in layer-3.cue.

A policy captures scope, imported controls and guidance,
RACI contacts, implementation timelines, and how
adherence is evaluated and enforced.

## Policy Structure

A Policy Document has these key sections:

| Section | Purpose |
|---------|---------|
| `metadata` | Identity, author, mapping references |
| `contacts` | RACI: responsible, accountable, consulted, informed |
| `scope` | What is in and out of scope |
| `imports` | External policies, control catalogs, guidance |
| `implementation-plan` | Evaluation and enforcement timelines |
| `adherence` | How compliance is evaluated and enforced |

Example metadata:

```yaml
title: "Information Security Policy"
metadata:
  id: "org-policy-001"
  type: Policy
  gemara-version: "0.20.0"
  description: "Policy for cloud and web application security"
  version: "1.0.0"
  author:
    id: security-team
    name: "Security Team"
    type: Human
  mapping-references:
    - id: "SEC.SLAM.CM"
      title: "Container Management Control Catalog"
      version: "1.0.0"
      description: "Control catalog for container security"
```

Define contacts with RACI roles:

```yaml
contacts:
  responsible:
    - name: "Platform Engineering"
      affiliation: "Engineering"
      email: "platform@example.com"
  accountable:
    - name: "CISO"
      affiliation: "Security"
  consulted:
    - name: "Legal"
      affiliation: "Legal"
  informed:
    - name: "All Engineering"
      affiliation: "Engineering"
```

## Implementation Plan

The implementation plan defines when the policy becomes
active with evaluation and enforcement timelines:

```yaml
implementation-plan:
  notification-process: |
    Policy communicated via internal wiki and
    team leads.
  evaluation-timeline:
    start: 2025-03-01T00:00:00Z
    end: 2025-06-01T00:00:00Z
    notes: Initial evaluation phase
  enforcement-timeline:
    start: 2025-06-01T00:00:00Z
    notes: Enforcement after evaluation baseline
```

The notification process describes how stakeholders
learn about the policy. Evaluation precedes enforcement.

## Evaluation Timeline

The evaluation timeline defines when compliance
assessment begins and ends:

| Field | Required | Description |
|-------|----------|-------------|
| `start` | Yes | ISO 8601 timestamp |
| `end` | No | End of evaluation phase |
| `notes` | No | Additional context |

During evaluation, automated checks are rolled out and
baselines established. This phase produces the data
needed for enforcement decisions.

## Enforcement Timeline

The enforcement timeline defines when violations have
consequences:

| Field | Required | Description |
|-------|----------|-------------|
| `start` | Yes | When enforcement begins |
| `notes` | No | Enforcement conditions |

Enforcement begins after the evaluation baseline is
established. The policy should define what happens on
non-compliance (see Adherence section).

## Adherence

Define how compliance is evaluated and enforced:

| Field | Purpose |
|-------|---------|
| `evaluation-methods` | How compliance is checked |
| `assessment-plans` | Specific assessment schedules |
| `enforcement-methods` | How violations are handled |
| `non-compliance-handling` | Escalation procedures |

Evaluation method types:

| Type | Description |
|------|-------------|
| `Manual` | Human-driven review |
| `Behavioral` | Observable process compliance |
| `Automated` | Tool-driven verification |
| `Autoremediation` | Automatic fix on violation |
| `Gate` | Blocks progression until compliant |

Example adherence:

```yaml
adherence:
  evaluation-methods:
    - type: Automated
      description: CI pipeline checks
      executor: "GitHub Actions"
    - type: Gate
      description: Deployment gate
      executor: "Admission Controller"
  assessment-plans:
    - requirement-id: "SEC.SLAM.CM.CTL02.AR01"
      frequency: "per-deployment"
      evaluation-methods:
        - Automated
      evidence-requirements:
        - "TLS certificate chain logs"
  non-compliance-handling:
    - "Deployment blocked by admission controller"
    - "Security team notified within 24 hours"
    - "Remediation tracked in issue tracker"
```

Define `scope.in` and `scope.out` with dimensions:

```yaml
scope:
  in:
    technologies: ["Cloud Computing", "Web Applications"]
    geopolitical: ["United States", "European Union"]
    users: ["developers", "platform-engineers"]
  out:
    technologies: ["Legacy On-Premises"]
```

Import control catalogs with optional assessment
requirement modifications:

```yaml
imports:
  catalogs:
    - reference-id: "SEC.SLAM.CM"
      assessment-requirement-modifications:
        - id: "CTL02-AR01-strict"
          target-id: "SEC.SLAM.CM.CTL02.AR01"
          modification-type: Override
          modification-rationale: "Stricter TLS requirements"
          text: "MUST use TLS 1.3 with certificate pinning"
```

Validate the complete policy:

```bash
cue vet -c -d '#Policy' \
  github.com/gemaraproj/gemara@latest \
  your-policy.yaml
```

**What's next:** Use the policy in Layer 5 evaluations
to record which assessment requirements have been
verified. See the schema documentation at
https://gemara.openssf.org/schema/layer-3.html
