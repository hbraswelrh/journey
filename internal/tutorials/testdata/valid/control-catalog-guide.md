---
title: Control Catalog Guide
layer: 2
schema_version: v0.20.0
sections:
  - Control Catalog Structure
  - Custom Control Authoring
  - Importing External Catalogs
  - OSPS Baseline Integration
  - FINOS CCC Integration
---

# Control Catalog Guide

This tutorial walks through creating a Control Catalog
using the Gemara schema, building on the threats and
scope identified in the Threat Assessment Guide.

Controls are safeguards with a stated objective and
testable assessment requirements. Families group related
controls by domain (e.g., supply chain, access control).
Threats link each control to what it mitigates.

## Control Catalog Structure

A Control Catalog requires metadata, mapping references,
applicability categories, families, and controls.

Key metadata fields:

| Field | Purpose |
|-------|---------|
| `title` | Display name for the control catalog |
| `mapping-references` | Pointers to threat/control catalogs |
| `applicability-categories` | When controls apply |

Applicability categories define scope for assessment
requirements (e.g., production, CI/CD, untrusted
networks). Each assessment requirement references these
category ids.

Example metadata:

```yaml
title: Container Management Tool Control Catalog
metadata:
  id: SEC.SLAM.CM
  description: Control catalog for container security
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: SEC.SLAM.CM
      title: Container Management Threat Catalog
      version: "1.0.0"
      url: file://threats.yaml
      description: Threat IDs for threat-mappings
    - id: CCC
      title: Common Cloud Controls Core
      version: v2025.10
      url: https://github.com/finos/common-cloud-controls
      description: Reusable security controls by FINOS
  applicability-categories:
    - id: production
      title: Production
      description: Production workloads and clusters
    - id: all_deployments
      title: All Deployments
      description: All build, pull, or run environments
```

## Custom Control Authoring

For controls specific to your scope, define them with
these required fields:

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Pattern: ORG.PROJ.COMP.CTL## |
| `title` | Yes | Short name describing the control |
| `objective` | Yes | Risk-reduction statement |
| `family` | Yes | Family id from this catalog |
| `assessment-requirements` | Yes | Testable conditions |
| `threats` | No | Links to threat catalog IDs |
| `state` | Yes | Active, Draft, Deprecated, Retired |

Each **assessment requirement** must be testable — an
evaluator must determine pass or fail from the text
alone. Use the pattern: "When [condition], [subject]
MUST [observable action]."

Good: "When YAML is submitted, the server MUST reject
payloads exceeding a configured maximum size."

Bad: "User input MUST be validated or sanitized."

Example control with assessment requirements:

```yaml
controls:
  - id: SEC.SLAM.CM.CTL01
    title: Use Immutable Image References by Digest
    objective: |
      Require signature validation so only trusted
      images are accepted, then pin each image to an
      immutable digest after the check.
    threats:
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.THR01
          - reference-id: SEC.SLAM.CM.THR03
    family: SEC.SLAM.CM.FAM01
    state: Active
    assessment-requirements:
      - id: SEC.SLAM.CM.CTL01.AR01
        text: |
          The system MUST resolve image references to
          immutable digests before deployment.
        applicability: ["all_deployments"]
        state: Active
```

## Importing External Catalogs

Import controls from external catalogs (FINOS CCC, OSPS
Baseline) via `mapping-references` and `imports`:

```yaml
imports:
  controls:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.CTL42
        remarks: Image signing and verification
```

Ensure each `reference-id` appears in your
`metadata.mapping-references`. The imported catalog's
controls apply alongside your custom controls.

## OSPS Baseline Integration

The OpenSSF Open Source Project Security (OSPS) Baseline
defines foundational security requirements for open
source projects. Add it as a mapping reference and
import relevant controls:

```yaml
metadata:
  mapping-references:
    - id: OSPS
      title: OSPS Baseline
      version: "2025"
      url: https://baseline.openssf.org
      description: OpenSSF project security baseline
```

OSPS controls cover areas like dependency management,
CI/CD pipeline security, and contributor access control
that complement FINOS CCC's cloud-focused controls.

## FINOS CCC Integration

The FINOS Common Cloud Controls (CCC) Core catalog is
the recommended starting point. It provides pre-built
controls, families, and threat mappings you can import.

Catalog download:
https://github.com/finos/common-cloud-controls/releases

To integrate:
1. Add CCC as a `mapping-reference` in metadata
2. Import controls via `imports.controls`
3. Reference CCC threat IDs in control `threats` fields
4. Use CCC family IDs or define your own

Validate the complete catalog:

```bash
cue vet -c -d '#ControlCatalog' \
  github.com/gemaraproj/gemara@latest \
  your-controls.yaml
```

**What's next:** Build a Policy referencing this Control
Catalog (Layer 3 Policy schema), or generate Privateer
plugins from assessment requirements. See
https://gemara.openssf.org/schema/layer-2.html
