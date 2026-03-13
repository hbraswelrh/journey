---
title: Threat Assessment Guide
layer: 2
schema_version: v0.20.0
sections:
  - Scope Definition
  - Capability Identification
  - Threat Identification
  - CUE Validation
---

# Threat Assessment Guide

This tutorial walks through creating a threat assessment
using the Gemara schema. Think of the component you are
assessing like a house: first, identify what it can do
(its capabilities), then identify what could go wrong
(threats to those capabilities).

In technical terms:
- **Capabilities** define what the technology can do.
  These form the **attack surface** because every
  intended function represents a potential path for
  unintended use.
- **Threats** define specific ways those capabilities
  could be misused or exploited.

## Scope Definition

Select a component or technology to assess — a service,
API, infrastructure component, or technology stack. Then
declare your scope and mapping references in the
metadata block.

**Leverage existing resources**: Gemara supports importing
threats and capabilities from external catalogs. The
FINOS Common Cloud Controls (CCC) Core catalog defines
well-vetted capabilities and threats that apply broadly
across cloud services:
https://github.com/finos/common-cloud-controls/releases

Key metadata fields:

| Field | Purpose |
|-------|---------|
| `title` | Display name for the threat catalog |
| `mapping-references` | Pointers to external catalogs (e.g., CCC) |
| `metadata.id` | Unique identifier (ORG.PROJ.COMP format) |

Example metadata YAML:

```yaml
title: Container Management Tool Threat Catalog
metadata:
  id: SEC.SLAM.CM
  description: Threat catalog for container management
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: CCC
      title: Common Cloud Controls Core
      version: v2025.10
      url: https://github.com/finos/common-cloud-controls
      description: Reusable security controls by FINOS
```

## Capability Identification

Capabilities are the core functions or features of the
component within your defined scope.

**Start with imported capabilities** from FINOS CCC.
Ask: "Which common cloud capabilities does this
technology have?" For example, a container management
tool that pulls images from registries matches CCC
capability CP29 (Active Ingestion). Image tags
functioning as version identifiers match CP18 (Resource
Versioning).

Import capabilities in YAML:

```yaml
imports:
  capabilities:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.CP29
        remarks: Active Ingestion
      - reference-id: CCC.Core.CP18
        remarks: Resource Versioning
```

**Then define custom capabilities** unique to your
target. Required fields:

| Field | Required | Description |
|-------|----------|-------------|
| Capability ID | Yes | Pattern: ORG.PROJ.COMP.CAP## |
| Title | Yes | Clear, concise name |
| Description | Yes | What this capability does |

Example custom capability:

```yaml
capabilities:
  - id: SEC.SLAM.CM.CAP01
    title: Image Retrieval by Tag
    description: |
      Ability to retrieve container images from
      registries using mutable tag names.
```

## Threat Identification

For each capability, identify potential threats — what
could go wrong?

**Check for imported threats first.** Review the CCC
Core catalog for threats linked to the capabilities you
imported. For example, CCC Core defines TH14 ("Older
Resource Versions are Used") linked to CP18. It applies
when mutable image tags let the tool resolve to a stale
or compromised version.

Import threats:

```yaml
imports:
  threats:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.TH14
```

**Define custom threats** with these required fields:

| Field | Required | Description |
|-------|----------|-------------|
| Threat ID | Yes | Pattern: ORG.PROJ.COMP.THR## |
| Title | Yes | Clear name describing the threat |
| Description | Yes | What goes wrong and why it matters |
| Capabilities | Yes | Links to the capabilities exploited |

Example custom threat:

```yaml
threats:
  - id: SEC.SLAM.CM.THR01
    title: Container Image Tampering or Poisoning
    description: |
      Attackers replace a legitimately published image
      tag with a malicious image by exploiting tag
      mutability in image registries.
    capabilities:
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.CP29
          - reference-id: CCC.Core.CP18
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.CAP01
```

Optional: Link threats to MITRE ATT&CK techniques via
`vectors` entries referencing the ATT&CK Enterprise
matrix for structured threat intelligence.

## CUE Validation

Assemble the complete threat catalog YAML and validate
against the Gemara schema:

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#ThreatCatalog' \
  github.com/gemaraproj/gemara@latest \
  your-threats.yaml
```

Fix any reported errors (missing required fields,
invalid references, malformed mappings) until the
catalog passes validation.

**What's next:** Create a Control Catalog that maps
security controls to the identified threats using the
Control Catalog Guide. See the Gemara Layer 2 schema
documentation at
https://gemara.openssf.org/schema/layer-2.html
