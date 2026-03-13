---
title: Guidance Catalog Guide
layer: 1
schema_version: v0.20.0
sections:
  - Creating a Guidance Catalog
  - Metadata Setup
  - Families and Groups
  - Cross-References
  - Mapping Documents
---

# Guidance Catalog Guide

This tutorial walks through creating a Guidance Catalog
using the Gemara schema. A Guidance Catalog is a
structured set of guidelines — recommendations,
requirements, or best practices — that help achieve
desired security outcomes.

Guidelines are grouped into families and can reference
each other within the same catalog. External mappings
(to OWASP, NIST, HIPAA, etc.) are handled through
separate Mapping Documents.

## Creating a Guidance Catalog

Start by choosing the scope and catalog type:

| Type | When to use |
|------|-------------|
| `Standard` | Formal, normative specs (ISO 27001, PCI-DSS) |
| `Regulation` | Legal requirements (HIPAA, GDPR, CRA) |
| `Best Practice` | Non-mandatory recommendations |
| `Framework` | High-level structure (NIST CSF) |

**Who writes guidance:** Internal teams (unique org
circumstances), industry groups (OWASP Top 10, PCI),
government agencies (NIST, HIPAA), or international
standards bodies (GDPR, CRA, ISO). Compliance
professionals can use Gemara as a logical model for
categorizing and mapping compliance activities.

## Metadata Setup

Declare the catalog with metadata and optional mapping
references for external standards:

| Field | Purpose |
|-------|---------|
| `title` | Display name for the catalog |
| `type` | Standard, Regulation, Best Practice, Framework |
| `metadata.id` | Unique identifier for referencing |
| `metadata.type` | Must be `GuidanceCatalog` |
| `metadata.gemara-version` | Gemara spec version |
| `metadata.applicability-categories` | Scope categories |

Example metadata:

```yaml
title: Secure Software Development Guidance
metadata:
  id: ORG.SSD.001
  type: GuidanceCatalog
  gemara-version: "0.20.0"
  description: Internal secure development guidelines
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: OWASP
      title: OWASP Top 10
      version: "2021"
      url: https://owasp.org/Top10
      description: OWASP Top 10 Security Risks
  applicability-categories:
    - id: containerized_workloads
      title: Containerized Workloads
      description: Container-based deployments
    - id: ci_cd
      title: CI/CD
      description: Integration and deployment pipelines
type: Best Practice
```

## Families and Groups

Families group guidelines by theme. The schema requires
at least one family when the catalog defines guidelines.
Each guideline's `family` field must match a family `id`.

Required fields per family:

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique identifier for controls to reference |
| `title` | Yes | Short name for the family |
| `description` | Yes | What guidelines in this family address |

Example:

```yaml
families:
  - id: ORG.SSD.FAM01
    title: Secure Dependencies and Supply Chain
    description: |
      Guidelines for selecting, updating, and
      verifying dependencies and images.
```

## Cross-References

Guidelines have required fields (id, title, objective,
family, state) and optional fields for linking:

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique guideline identifier |
| `title` | Yes | Short name |
| `objective` | Yes | Statement of intent |
| `family` | Yes | Family id reference |
| `state` | Yes | Active, Draft, Deprecated, Retired |
| `recommendations` | No | Actionable suggestions |
| `applicability` | No | Category ids from metadata |
| `see-also` | No | Cross-references to other guidelines |

Example with cross-references:

```yaml
guidelines:
  - id: ORG.SSD.GL01
    title: Prefer Immutable Image References
    objective: |
      Use digest-based references for container images
      to prevent tampering and ensure repeatable
      deployments.
    family: ORG.SSD.FAM01
    state: Active
    recommendations:
      - Prefer pull-by-digest over tags for production
      - Pin base image digests in Dockerfiles
    applicability:
      - containerized_workloads
      - ci_cd
    see-also:
      - ORG.SSD.GL02
      - ORG.SSD.GL03
```

## Mapping Documents

External mappings (to OWASP, NIST, etc.) are handled
through separate **Mapping Documents**, not inline in
the guidance catalog. Mapping Documents define
relationships between source and target artifacts:

- Source: your guidance catalog
- Target: external standard (declared in
  mapping-references)
- Relationship: implements, equivalent, or subsumes

Downstream Gemara layers (Layer 2 controls, Layer 3
policies) reference guidelines from guidance catalogs
via their `guidelines` field.

Validate your guidance catalog:

```bash
cue vet -c -d '#GuidanceCatalog' \
  github.com/gemaraproj/gemara@latest \
  your-guidance.yaml
```

**What's next:** Map guidelines to Layer 2 controls via
control catalogs, or reference this guidance from a
Policy. See the schema documentation at
https://gemara.openssf.org/schema/layer-1.html
