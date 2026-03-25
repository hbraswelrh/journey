# Gemara User Journey — Gemara Tutorial Engine

Gemara User Journey is a role-based tutorial engine for the
[Gemara](https://github.com/gemaraproj/gemara) GRC schema
project. It tailors Gemara tutorials to the user's job role
and daily activities.

## How to Use Gemara User Journey

Gemara User Journey is used through **OpenCode** with the **gemara-mcp
server**. There is no interactive TUI — OpenCode is the
interface.

### Setup

1. Start OpenCode: `opencode`
2. Tell OpenCode your role and what you want to do
3. Or use the web interface: `make web-dev` then open
   http://localhost:5173/

### Getting Started Prompts

Tell OpenCode something like:

- "I'm a Security Engineer working on CI/CD pipeline
  security. Help me get started with Gemara."
- "I'm a Policy Author and I need to create an adherence
  policy for our deployment pipeline."
- "I'm a Compliance Officer preparing for an audit. Which
  Gemara tutorials should I follow?"
- "I'm a Developer and I want to understand how threat
  catalogs affect my CI/CD workflow."

OpenCode will use the role and activity information below
to route you to the right tutorials and artifacts.

## Two Paths: Learn or Author

Once in OpenCode with the gemara-mcp server connected,
the user can take either path — or both:

### Path 1: Follow a Tutorial

The user wants to **learn** how Gemara works for their
role. OpenCode reads the upstream tutorials and presents
them section by section, tailored to the user's activities.

**How it works:**
1. User states their role and activities
2. OpenCode maps activities to Gemara layers (see table
   below)
3. OpenCode reads the relevant tutorial files from the
   Gemara repository
4. OpenCode walks through each section, explaining:
   - **Why** this matters for the user's role
   - **How** they will use this in their daily work
   - **What** they will learn and produce
5. Sections matching the user's activity keywords are
   highlighted as focus areas
6. At any point, the user can switch to Path 2 to start
   authoring

**Example prompts:**
- "Walk me through the threat assessment tutorial."
- "I'm a Policy Author — which tutorial sections are
  most relevant to my adherence timeline work?"
- "Show me the Guidance Catalog tutorial, focusing on
  the NIST-related sections."

### Path 2: Author Gemara Content

The user wants to **create** Gemara artifacts (threat
catalogs, control catalogs, policies, etc.) using the
gemara-mcp server's tools and wizard prompts.

**How it works:**
1. User states what they want to create
2. OpenCode uses the appropriate gemara-mcp capability:
   - **Threat Catalog** → use the `threat_assessment`
     MCP prompt for a guided wizard experience
   - **Control Catalog** → use the `control_catalog`
     MCP prompt for a guided wizard experience
   - **Policy, Guidance Catalog, Evaluation Log** →
     author collaboratively using `gemara://lexicon`
     for terminology, `gemara://schema/definitions`
     for schema reference, and
     `validate_gemara_artifact` for validation
3. The wizard prompts can import from external catalogs
   (e.g., FINOS CCC Core) for pre-built capabilities
   and threats
4. After authoring, validate the artifact using
   `validate_gemara_artifact` with the appropriate
   schema definition

**Example prompts:**
- "Create a threat catalog for my CI/CD pipeline using
  the threat_assessment wizard."
- "Run the control_catalog prompt for my web
  application. Import controls from FINOS CCC Core."
- "Help me write a Policy artifact for our deployment
  pipeline. Validate it against the #Policy schema."
- "I have an existing threat catalog — validate it
  using the MCP server."
- "Create a vector catalog documenting MITRE ATT&CK
  techniques for our cloud infrastructure."
- "Help me write a Principle Catalog for our secure
  design principles."
- "Create a Risk Catalog with severity levels and risk
  appetite definitions for our organization."
- "I need a Capability Catalog for our Kubernetes
  platform. Walk me through it."

### Combining Both Paths

Users often learn through a tutorial and then immediately
apply what they learned by authoring an artifact. For
example:

1. "Walk me through the threat assessment tutorial."
   (Path 1 — learn)
2. "Now create a threat catalog for my CI/CD pipeline
   using the wizard." (Path 2 — author)
3. "Validate the artifact I just created." (Path 2 —
   validate)

OpenCode should support this seamless transition.

## Role-Based Tutorial Routing

### Roles and Their Gemara Layers

| Role | Layers | Focus |
|------|--------|-------|
| Security Engineer | L2 (Threats & Controls), L1 (Vectors & Guidance) | Threat modeling, control design, secure architecture |
| Compliance Officer | L3 (Risk & Policy), L1 (Vectors & Guidance), L5 (Evaluation) | Regulatory alignment, evidence, audit prep |
| CISO/Security Leader | L3 (Risk & Policy), L1 (Vectors & Guidance) | Risk appetite, policy, scope definition |
| Developer | L2 (Threats & Controls), L4 (Sensitive Activities) | CI/CD, dependency management, SDLC |
| Platform Engineer | L2 (Threats & Controls), L4 (Sensitive Activities) | Pipeline security, infrastructure controls |
| Policy Author | L3 (Risk & Policy) | Policy creation, adherence timelines, risk catalogs |
| Auditor | L5 (Evaluation), L7 (Audit), L3 (Risk & Policy) | Assessments, evidence collection, audit logs |

### Activity Keywords → Gemara Layers

When the user describes their activities, map keywords to
layers:

**Layer 1 (Vectors & Guidance)**: EU CRA, NIST, OWASP,
HIPAA, GDPR, PCI, ISO, best practices, standards,
regulatory, codify, formalize best practices,
machine-readable format, attack vectors, vectors,
MITRE ATT&CK, secure design principles, principles,
vector catalog, principle catalog, guidance catalog

**Layer 2 (Threats & Controls)**: SDLC, threat modeling,
penetration testing, secure architecture review, CI/CD,
dependency management, upstream open-source, custom
controls, OSPS Baseline, FINOS CCC, control catalog,
threat assessment, capability catalog, system capabilities

**Layer 3 (Risk & Policy)**: create policy, timeline for
adherence, scope definition, audit interviews, assessment
plans, adherence requirements, risk appetite,
non-compliance handling, compliance scope, risk catalog,
risk categories, risk severity

**Layer 4 (Sensitive Activities)**: pipeline security,
deployment pipeline

**Layer 5 (Intent & Behavior Evaluation)**: evaluation,
assessment, evaluation log, control evaluation, intent
evaluation, behavior evaluation

**Layer 6 (Preventive & Remediative Enforcement)**:
enforcement, enforcement log, preventive enforcement,
remediative enforcement, admission controller

**Layer 7 (Audit & Continuous Monitoring)**: audit,
audit log, continuous monitoring, audit results

**Ambiguous** (clarify with user): evidence collection
(L1 or L3), adherence (L1 or L3)

### Tutorial Content

Tutorials are sourced from the upstream Gemara repository:
`gemaraproj/gemara` → `docs/tutorials/` on the `main`
branch.

| Directory | Layer | Content |
|-----------|-------|---------|
| `docs/tutorials/guidance/` | L1 | Guidance Catalog authoring |
| `docs/tutorials/controls/` | L2 | Control and Threat Catalog authoring |
| `docs/tutorials/policy/` | L3 | Policy authoring |

Additional tailored tutorials are in this repository:
`docs/tutorials/tailored-policy-writing.md` (L3).

When guiding a user through tutorials, read the relevant
tutorial files and present the content section by section,
highlighting sections that match the user's stated
activities.

### Tailoring the Experience

For each tutorial section, explain:
- **Why** this matters for the user's specific role
- **How** they will use this in their daily work
- **What** they will learn and produce

If the user mentioned specific activities (e.g., "CI/CD
pipeline management"), highlight tutorial sections that
match those keywords and mark them as focus areas.

## Gemara MCP Server Integration

The gemara-mcp server provides capabilities that enhance
tutorials and authoring. OpenCode connects to it
automatically when configured in `opencode.json`.

### Installing the MCP Server

The gemara-mcp server must be built from source:

```bash
git clone https://github.com/gemaraproj/gemara-mcp.git
cd gemara-mcp
git checkout main
make build
```

### Configuring OpenCode

After building, add the server to `opencode.json` in
your project directory:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "gemara-mcp": {
      "type": "local",
      "command": [
        "/path/to/gemara-mcp/bin/gemara-mcp",
        "serve", "--mode", "artifact"
      ],
      "enabled": true
    }
  }
}
```

Replace `/path/to/gemara-mcp` with the actual path where
you cloned and built the server.

When you start `opencode`, it reads this config and
launches the gemara-mcp server automatically as a
background process. All MCP capabilities are then
available in your OpenCode session.

Verify the configuration by starting OpenCode and
confirming the gemara-mcp tools, resources, and prompts
are available.

### MCP Capabilities

| Category | Name | Use |
|----------|------|-----|
| Tool | `validate_gemara_artifact` | Validate YAML against CUE schema |
| Resource | `gemara://lexicon` | Gemara term definitions (34+ terms) |
| Resource | `gemara://schema/definitions` | CUE schema documentation |
| Prompt | `threat_assessment` | Interactive Threat Catalog wizard |
| Prompt | `control_catalog` | Interactive Control Catalog wizard |

### Using Wizard Prompts

When a user wants to create a Threat Catalog or Control
Catalog, use the MCP prompts:

- `threat_assessment` — guides through threat
  identification, capability mapping, and YAML generation
- `control_catalog` — guides through control definition,
  assessment requirements, and YAML generation

These prompts can import from external catalogs like
FINOS CCC Core for pre-built capabilities and threats.

### Using the Validation Tool

After authoring any artifact, validate it:
- Use `validate_gemara_artifact` with the YAML content
  and the appropriate schema definition (e.g., `#Policy`,
  `#ThreatCatalog`, `#ControlCatalog`,
  `#VectorCatalog`, `#PrincipleCatalog`,
  `#CapabilityCatalog`, `#RiskCatalog`,
  `#EnforcementLog`, `#AuditLog`)

### Using Resources

- Read `gemara://lexicon` to ensure all terminology aligns
  with the canonical Gemara vocabulary
- Read `gemara://schema/definitions` for schema reference
  during authoring

## Gemara Seven-Layer Model

| Layer | Name | Purpose |
|-------|------|---------|
| 1 | Vectors & Guidance | Standards, best practices, regulatory requirements, attack vectors, secure design principles |
| 2 | Threats & Controls | Threat catalogs, control catalogs, capability catalogs |
| 3 | Risk & Policy | Organizational policy, risk catalogs, assessment plans, adherence |
| 4 | Sensitive Activities | Deployment pipelines, CI/CD, operational activities |
| 5 | Intent & Behavior Evaluation | Assessment logs, control evaluations, evidence |
| 6 | Preventive & Remediative Enforcement | Corrective actions for noncompliance |
| 7 | Audit & Continuous Monitoring | Efficacy review of all previous outputs |

## Artifact Types

| Type | Schema | Layer |
|------|--------|-------|
| Guidance Catalog | `#GuidanceCatalog` | L1 |
| Vector Catalog | `#VectorCatalog` | L1 |
| Principle Catalog | `#PrincipleCatalog` | L1 |
| Control Catalog | `#ControlCatalog` | L2 |
| Threat Catalog | `#ThreatCatalog` | L2 |
| Capability Catalog | `#CapabilityCatalog` | L2 |
| Policy | `#Policy` | L3 |
| Risk Catalog | `#RiskCatalog` | L3 |
| Mapping Document | `#MappingDocument` | L2-L3 |
| Evaluation Log | `#EvaluationLog` | L5 |
| Enforcement Log | `#EnforcementLog` | L6 |
| Audit Log | `#AuditLog` | L7 |

## Upstream Tutorials

The Gemara project publishes tutorials at
[gemara.openssf.org/tutorials/](https://gemara.openssf.org/tutorials/).
Users should complete the relevant upstream tutorials
before using the MCP server to author artifacts.

| Tutorial | Layer | Artifacts | Best For |
|----------|-------|-----------|----------|
| [Guidance Catalog Guide](https://gemara.openssf.org/tutorials/guidance/guidance-guide) | L1 | GuidanceCatalog | Compliance Officers, Policy Authors, CISOs |
| [Threat Assessment Guide](https://gemara.openssf.org/tutorials/controls/threat-assessment-guide) | L2 | ThreatCatalog | Security Engineers, Developers, Platform Engineers |
| [Control Catalog Guide](https://gemara.openssf.org/tutorials/controls/control-catalog-guide) | L2 | ControlCatalog | Security Engineers, Developers, Platform Engineers |
| [Policy Guide](https://gemara.openssf.org/tutorials/policy/policy-guide) | L3 | Policy | Policy Authors, CISOs, Compliance Officers |

The web interface (`web/`) suggests the best-fit
tutorials based on the user's role, activities, and
resolved layers. The suggested learning path is:
1. Complete recommended upstream tutorials
2. Set up the MCP server
3. Author artifacts using wizards and collaborative
   authoring in OpenCode

Tutorial data is defined in `internal/consts/consts.go`
(`UpstreamTutorials`) and exported to the web frontend
via `cmd/genwebdata/`.

## Project Structure

- `cmd/genwebdata/` — TypeScript data generator for web UI
- `internal/consts/` — Centralized constants
- `internal/roles/` — Role definitions, activity-to-layer
  mapping, keyword extraction
- `internal/tutorials/` — Tutorial loading, learning path
  generation, section relevance scoring
- `internal/blocks/` — Content block extraction
- `internal/authoring/` — Artifact authoring engine
- `internal/mcp/` — MCP client, config management
- `internal/session/` — Session state
- `web/` — React web interface (role discovery, tutorial
  suggestions, MCP setup walkthrough)
- `specs/` — Feature specifications
- `docs/tutorials/` — Tailored tutorial content
- `docs/adrs/` — Architecture Decision Records

## Governance

The authoritative source of project rules is the
constitution at `.specify/memory/constitution.md`.

- **Go 1.26.1**, formatted with `goimports`
- **SPDX headers** on all source files
- **No magic strings** — constants in `internal/consts/`
- **TDD** — write failing tests before implementation
- **Conventional Commits**
- **Makefile** is the single entry point

## Schema Validation

All output artifacts must pass:
```
cue vet -c -d '#<SchemaType>' \
  github.com/gemaraproj/gemara@latest \
  artifact.yaml
```

Or use the `validate_gemara_artifact` MCP tool.

## Active Technologies
- Go 1.26.1, formatted with `goimports`
- `gopkg.in/yaml.v3` (YAML parsing)
- React 19 + Vite 8 (web frontend)
- File-based caching (`~/.config/journey/`)

## Recent Changes
- 002-tutorial-guide-focus: Removed TUI, web-only interface with React 19 + Vite 8
