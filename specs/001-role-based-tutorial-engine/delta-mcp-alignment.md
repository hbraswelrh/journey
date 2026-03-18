# Delta: Align Spec with Actual gemara-mcp Server

**Date**: 2026-03-16
**Source**: gemara-mcp README.md (gemaraproj/gemara-mcp main)
**Scope**: spec.md in specs/001-role-based-tutorial-engine/

This document catalogs every proposed change to `spec.md` to
align the user stories, functional requirements, key entities,
edge cases, and success criteria with the actual gemara-mcp
server architecture. No changes are applied until approved.

---

## Background: Actual gemara-mcp Architecture

The gemara-mcp server exposes three MCP protocol categories:

| Category | Name | Mode | Description |
|----------|------|------|-------------|
| **Tool** | `validate_gemara_artifact` | advisory, artifact | Validate YAML against CUE schema |
| **Resource** | `gemara://lexicon` | advisory, artifact | Term definitions for the Gemara model |
| **Resource** | `gemara://schema/definitions` | advisory, artifact | CUE schema definitions (latest) |
| **Resource** | `gemara://schema/definitions{?version}` | advisory, artifact | CUE schema definitions (specific version) |
| **Prompt** | `threat_assessment` | artifact only | Interactive wizard for Threat Catalogs |
| **Prompt** | `control_catalog` | artifact only | Interactive wizard for Control Catalogs |

Two server modes:
- `advisory` — Read-only analysis and validation
- `artifact` — All advisory capabilities plus guided wizards

Default mode: `artifact`. Selected via `--mode` flag:
```
gemara-mcp serve --mode advisory
gemara-mcp serve --mode artifact
```

### Key Terminology Correction

The MCP protocol distinguishes between:
- **Tools**: Callable functions (request-response, like API
  endpoints). `validate_gemara_artifact` is a tool.
- **Resources**: Static or parameterized data the server exposes
  (like REST resources). The lexicon and schema docs are
  resources, not tools.
- **Prompts**: Templated conversation starters for guided
  workflows. The wizards are prompts, not tools.

The current spec incorrectly classifies all three as "tools."

---

## Change 1: User Story 1 — MCP Server Setup (Lines 10-116)

### 1a. Replace "three tools" with accurate MCP categories

**Current** (lines 20-33):
> The MCP server provides direct access to three tools that
> augment Pac-Man's capabilities throughout all subsequent
> operations:
>
> - **get_lexicon**: Retrieve the upstream Gemara lexicon...
> - **validate_gemara_artifact**: Validate YAML artifacts...
> - **get_schema_docs**: Retrieve schema documentation...

**Proposed**:
> The MCP server provides direct access to MCP tools, resources,
> and prompts that augment Pac-Man's capabilities throughout all
> subsequent operations:
>
> - **Tool — `validate_gemara_artifact`**: Validate YAML
>   artifacts against Gemara schema definitions without
>   requiring the user to install CUE locally or run `cue vet`
>   manually.
> - **Resource — `gemara://lexicon`**: Retrieve the upstream
>   Gemara lexicon entries, ensuring all terminology used by
>   Pac-Man and the user's authored content aligns with the
>   canonical Gemara vocabulary.
> - **Resource — `gemara://schema/definitions`**: Retrieve
>   schema documentation for the Gemara CUE module, providing
>   contextual reference material during guided authoring and
>   learning paths. Supports a `version` parameter
>   (`gemara://schema/definitions{?version}`) for
>   version-specific documentation.
> - **Prompts — `threat_assessment`, `control_catalog`**
>   (artifact mode only): Interactive wizards that guide users
>   through creating Gemara-compatible Threat Catalogs and
>   Control Catalogs respectively.

**Rationale**: Aligns with MCP protocol categories. Users and
developers need to understand that lexicon and schema docs are
resources (read via resource URIs), not callable tools.

### 1b. Add server mode selection to US1

**Current**: No mention of modes. The config examples show
`"args": ["serve"]` without a `--mode` flag.

**Proposed**: Add a new paragraph after the installation
description (after line 46, before "Why this priority"):

> The system MUST prompt the user to select a server mode:
>
> - **Advisory mode** (`--mode advisory`): Read-only analysis
>   and validation of existing artifacts. Provides the
>   `validate_gemara_artifact` tool and all resources but no
>   guided creation prompts. Suitable for users who want to
>   validate existing work or explore schemas without wizard
>   assistance.
> - **Artifact mode** (`--mode artifact`): All advisory
>   capabilities plus guided artifact creation wizards
>   (`threat_assessment` and `control_catalog` prompts).
>   Suitable for users who want full guided authoring support.
>
> The default mode is `artifact`. The selected mode determines
> which MCP capabilities are available and MUST be recorded in
> the session state. The `opencode.json` configuration MUST
> include the `--mode` flag in the server args:
> ```json
> {
>   "mcpServers": {
>     "gemara-mcp": {
>       "command": "/path/to/gemara-mcp",
>       "args": ["serve", "--mode", "artifact"]
>     }
>   }
> }
> ```

**Rationale**: Mode selection is a first-class installation
decision that affects available capabilities throughout the
session.

### 1c. Update Acceptance Scenario 2 config example

**Current** (lines 74-86): References SSH/HTTPS choice and
config but does not include `--mode` in args.

**Proposed**: Add to the "Then" clause:
> ...writes or updates the OpenCode MCP configuration
> (`opencode.json`) with the built binary path as a local MCP
> server entry including the selected `--mode` flag in the args
> array, and confirms the server responds to a health check.

### 1d. Update Acceptance Scenario 6

**Current** (lines 110-115):
> ...it queries `get_lexicon` to load the current upstream
> lexicon, `get_schema_docs` to cache schema documentation, and
> confirms `validate_gemara_artifact` is available for
> on-demand validation throughout the session.

**Proposed**:
> ...it reads the `gemara://lexicon` resource to load the
> current upstream lexicon, reads
> `gemara://schema/definitions` to cache schema documentation,
> confirms the `validate_gemara_artifact` tool is available for
> on-demand validation, and — if running in artifact mode —
> confirms the `threat_assessment` and `control_catalog`
> prompts are listed via the prompts endpoint.

### 1e. Add new Acceptance Scenario 7 for mode selection

**Proposed**:

> 7. **Given** a user chooses to install the MCP server,
>    **When** installation completes, **Then** the system
>    prompts the user to select a server mode: Advisory (read-
>    only analysis and validation) or Artifact (advisory plus
>    guided creation wizards). The selected mode is written to
>    the OpenCode MCP configuration and recorded in the session
>    state. If the user selects Advisory mode, the system
>    informs them that the `threat_assessment` and
>    `control_catalog` prompts will not be available.

---

## Change 2: User Story 2 — Schema Version Selection
(Lines 119-207)

### 2a. Replace `get_schema_docs` with resource URI

**Current** (lines 134-136):
> ...the system MAY use the MCP server's `get_schema_docs`
> tool to supplement version information with schema
> documentation for the selected version.

**Proposed**:
> ...the system MAY use the MCP server's
> `gemara://schema/definitions{?version}` resource to
> supplement version information with schema documentation
> for the selected version.

**Rationale**: Schema docs are exposed as a parameterized MCP
resource, not a tool. The `{?version}` parameter allows
requesting docs for a specific version, which directly supports
the version selection flow.

### 2b. Update Acceptance Scenario 3

**Current** (line 173-176): References "the Gemara MCP server
is installed" for compatibility checks.

**Proposed**: Add to the "Then" clause:
> ...the system also reads
> `gemara://schema/definitions{?version}` with the selected
> version to verify schema documentation is available for that
> version.

---

## Change 3: User Story 6 — Guided Authoring (Lines 456-498)

### 3a. Reference MCP prompts for wizard-assisted authoring

**Current**: US6 describes guided authoring as a Pac-Man built-in
feature. It does not reference the MCP prompts.

**Proposed**: Add a new paragraph after the first paragraph
(after line 465):

> When the Gemara MCP server is running in artifact mode, the
> system MAY delegate authoring of Threat Catalogs and Control
> Catalogs to the MCP server's interactive prompts
> (`threat_assessment` and `control_catalog`). These prompts
> provide structured wizard flows that mirror the Gemara
> tutorials' structure and produce validated artifacts. The
> system MUST present the user with a choice between using the
> MCP-assisted wizard (if available) or the built-in guided
> authoring flow. When the MCP server is running in advisory
> mode or is unavailable, only the built-in guided authoring
> flow is available.

### 3b. Add new Acceptance Scenario 4 for MCP-assisted authoring

**Proposed**:

> 4. **Given** the Gemara MCP server is running in artifact
>    mode and a Security Engineer wants to author a Threat
>    Catalog, **When** they begin guided authoring, **Then**
>    the system offers two options: (a) use the MCP server's
>    `threat_assessment` prompt for an interactive wizard
>    experience, or (b) use the built-in guided authoring
>    flow. If the user selects the MCP wizard, the system
>    delegates to the prompt and validates the final artifact
>    using `validate_gemara_artifact`.

### 3c. Add new Acceptance Scenario 5 for advisory mode

**Proposed**:

> 5. **Given** the Gemara MCP server is running in advisory
>    mode and a user wants to author a Control Catalog,
>    **When** they begin guided authoring, **Then** the system
>    uses only the built-in guided authoring flow (the
>    `control_catalog` prompt is not available in advisory
>    mode) and validates each step using the
>    `validate_gemara_artifact` tool, which IS available in
>    advisory mode.

---

## Change 4: Functional Requirements (Lines 576-837)

### 4a. FR-026 — Replace "three tools" with accurate categories

**Current** (lines 729-735):
> The system MUST explain the three tools the MCP server
> provides (`get_lexicon`, `validate_gemara_artifact`,
> `get_schema_docs`) and how each enhances the Pac-Man
> experience.

**Proposed**:
> The system MUST explain the MCP server's capabilities: the
> `validate_gemara_artifact` tool for schema validation, the
> `gemara://lexicon` resource for terminology alignment, the
> `gemara://schema/definitions` resource for schema
> documentation, and — when running in artifact mode — the
> `threat_assessment` and `control_catalog` prompts for
> guided artifact creation wizards. The system MUST explain
> how each enhances the Pac-Man experience.

### 4b. FR-027 — Add mode flag to installation config

**Current** (lines 737-758): Describes two installation methods
but does not mention the `--mode` flag in the config.

**Proposed**: After "ensuring the MCP server is available in
subsequent OpenCode sessions" (line 751), add:

> The `opencode.json` entry MUST include the `--mode` flag
> in the args array, defaulting to `artifact` unless the user
> explicitly selects advisory mode. The args MUST be
> `["serve", "--mode", "<selected-mode>"]`.

### 4c. FR-028 — Update to use resource URIs

**Current** (lines 759-764):
> ...the system MUST use it as the preferred source for
> lexicon data (`get_lexicon`), schema documentation
> (`get_schema_docs`), and artifact validation
> (`validate_gemara_artifact`).

**Proposed**:
> ...the system MUST use it as the preferred source for
> lexicon data (reading the `gemara://lexicon` resource),
> schema documentation (reading
> `gemara://schema/definitions` or
> `gemara://schema/definitions{?version}`), and artifact
> validation (calling the `validate_gemara_artifact` tool).
> When running in artifact mode, the system MUST also make
> the `threat_assessment` and `control_catalog` prompts
> available for guided authoring.

### 4d. FR-029 — Update fallback to mention mode-specific gaps

**Current** (lines 765-771): Describes local fallbacks.

**Proposed**: Add to the end of FR-029:
> When the MCP server is installed but running in advisory
> mode, the system MUST inform the user that guided creation
> prompts (`threat_assessment`, `control_catalog`) are
> unavailable in this mode and MUST offer to reconfigure the
> server to artifact mode if the user attempts wizard-based
> authoring.

### 4e. FR-030 — Detect mode along with server presence

**Current** (lines 772-777): Detects whether MCP is installed
and running.

**Proposed**: Add:
> The system MUST also determine the server's operating mode
> (advisory or artifact) at session start and record it in
> the session state, adjusting available capabilities
> accordingly.

### 4f. New FR-036 — Server mode selection

**Proposed**:

> - **FR-036**: The system MUST support two gemara-mcp server
>   operating modes as defined by the server:
>   - **Advisory mode** (`--mode advisory`): Read-only
>     analysis and validation. Provides the
>     `validate_gemara_artifact` tool and all resources
>     (`gemara://lexicon`,
>     `gemara://schema/definitions{?version}`) but no
>     prompts.
>   - **Artifact mode** (`--mode artifact`): All advisory
>     capabilities plus the `threat_assessment` and
>     `control_catalog` prompts for guided artifact creation.
>   The default mode MUST be `artifact`. The system MUST
>   allow the user to select a mode during installation and
>   MUST allow mode changes in subsequent sessions by
>   updating the `opencode.json` configuration. The selected
>   mode MUST be persisted in the session state and MUST
>   determine which capabilities are presented to the user.

### 4g. New FR-037 — MCP resource access

**Proposed**:

> - **FR-037**: When the Gemara MCP server is available, the
>   system MUST access lexicon data and schema documentation
>   via MCP resource URIs (`gemara://lexicon` and
>   `gemara://schema/definitions{?version}`) rather than MCP
>   tool calls. The system MUST distinguish between MCP tools
>   (callable functions like `validate_gemara_artifact`), MCP
>   resources (data endpoints like `gemara://lexicon`), and
>   MCP prompts (guided workflows like `threat_assessment`)
>   in all internal code and user-facing explanations.

### 4h. New FR-038 — MCP prompt-assisted authoring

**Proposed**:

> - **FR-038**: When the Gemara MCP server is running in
>   artifact mode, the system MUST present users with the
>   option to use MCP prompts (`threat_assessment`,
>   `control_catalog`) as an alternative to the built-in
>   guided authoring flow for the corresponding artifact
>   types (ThreatCatalog, ControlCatalog). The system MUST
>   NOT present MCP prompt options when the server is in
>   advisory mode or unavailable. For artifact types that do
>   not have corresponding MCP prompts (GuidanceCatalog,
>   Policy, MappingDocument, EvaluationLog), the built-in
>   guided authoring flow MUST always be used.

---

## Change 5: Key Entities (Lines 839-901)

### 5a. Update MCP Server Connection entity

**Current** (lines 894-901):
> - **MCP Server Connection**: ...Attributes: installation
>   method (binary or Podman), connection status (running,
>   stopped, not installed), server version, Gemara schema
>   version the server was built against, compatibility status
>   with the user's selected schema version (compatible,
>   mismatched, unknown), available tools (get_lexicon,
>   validate_gemara_artifact, get_schema_docs), last health
>   check timestamp.

**Proposed**:
> - **MCP Server Connection**: The Gemara MCP server instance
>   used by the current session. Attributes: installation
>   method (binary or Podman), connection status (running,
>   stopped, not installed), server mode (advisory or
>   artifact), server version, Gemara schema version the
>   server was built against, compatibility status with the
>   user's selected schema version (compatible, mismatched,
>   unknown), available tools (`validate_gemara_artifact`),
>   available resources (`gemara://lexicon`,
>   `gemara://schema/definitions`), available prompts
>   (`threat_assessment`, `control_catalog` — artifact mode
>   only), last health check timestamp.

---

## Change 6: Edge Cases (Lines 502-574)

### 6a. Add mode-specific edge case

**Proposed** (new edge case after line 574):

> - What happens when the user attempts to launch an MCP
>   wizard (e.g., `threat_assessment`) but the server is
>   running in advisory mode? The system MUST inform the user
>   that prompts are only available in artifact mode, offer to
>   reconfigure the server to artifact mode (updating the
>   `--mode` flag in `opencode.json` and restarting the
>   server), and fall back to the built-in guided authoring
>   flow if the user declines.

### 6b. Add resource vs tool failure edge case

**Proposed** (new edge case):

> - What happens when an MCP resource read fails (e.g.,
>   `gemara://lexicon` returns an error) but the MCP tool
>   (`validate_gemara_artifact`) still works? The system MUST
>   handle partial MCP availability gracefully: fall back to
>   bundled lexicon data for the failed resource while
>   continuing to use the functional tool for validation. The
>   system MUST inform the user which specific capabilities
>   have fallen back to local equivalents.

---

## Change 7: Success Criteria (Lines 950-1021)

### 7a. Update SC-013 to include mode selection

**Current** (lines 1011-1014):
> ...Users who install the Gemara MCP server can complete the
> installation and verify server connectivity within 5 minutes,
> with no more than 3 steps for either the binary or Podman
> installation method.

**Proposed**:
> ...Users who install the Gemara MCP server can complete the
> installation (including mode selection) and verify server
> connectivity within 5 minutes, with no more than 4 steps for
> either the binary or Podman installation method.

(Step count increases from 3 to 4 because mode selection is
an additional decision point.)

### 7b. Update SC-014 to distinguish tools, resources, prompts

**Current** (lines 1016-1021):
> When the MCP server is available, all lexicon lookups, schema
> documentation requests, and artifact validations use the MCP
> server's tools.

**Proposed**:
> When the MCP server is available, all lexicon lookups use the
> `gemara://lexicon` resource, all schema documentation
> requests use the `gemara://schema/definitions` resource, and
> all artifact validations use the `validate_gemara_artifact`
> tool. When running in artifact mode, the `threat_assessment`
> and `control_catalog` prompts are available for guided
> authoring. When the MCP server becomes unavailable, the
> system falls back to local equivalents within 5 seconds with
> zero data loss on in-progress work.

### 7c. New SC-015 for mode behavior

**Proposed**:

> - **SC-015**: Users running the MCP server in advisory mode
>   can access all validation and reference capabilities
>   (tool + resources) but are correctly informed that guided
>   creation prompts are unavailable. Switching from advisory
>   to artifact mode (or vice versa) takes effect within one
>   session restart.

---

## Change 8: Assumptions (Lines 903-948)

### 8a. Add MCP protocol assumption

**Proposed** (new assumption):

> - The Gemara MCP server follows the Model Context Protocol
>   specification, exposing tools (callable functions),
>   resources (data endpoints accessed by URI), and prompts
>   (guided conversation templates). Pac-Man's MCP client
>   MUST use the appropriate MCP protocol methods for each
>   category: tool calls for `validate_gemara_artifact`,
>   resource reads for `gemara://lexicon` and
>   `gemara://schema/definitions`, and prompt listing/get for
>   `threat_assessment` and `control_catalog`.

---

## Summary of All Changes

| # | Section | Type | Description |
|---|---------|------|-------------|
| 1a | US1 | Correction | Replace "three tools" with tools/resources/prompts |
| 1b | US1 | Addition | Add server mode selection paragraph |
| 1c | US1 AS2 | Update | Include `--mode` in config example |
| 1d | US1 AS6 | Update | Use resource URIs, add prompt check |
| 1e | US1 | Addition | New AS7 for mode selection |
| 2a | US2 | Correction | Replace `get_schema_docs` with resource URI |
| 2b | US2 AS3 | Update | Add version-parameterized resource read |
| 3a | US6 | Addition | Reference MCP prompts for wizard authoring |
| 3b | US6 | Addition | New AS4 for MCP-assisted authoring |
| 3c | US6 | Addition | New AS5 for advisory mode authoring |
| 4a | FR-026 | Correction | Replace "three tools" with categories |
| 4b | FR-027 | Update | Add `--mode` flag to config requirement |
| 4c | FR-028 | Correction | Use resource URIs instead of tool names |
| 4d | FR-029 | Addition | Mode-specific fallback for prompts |
| 4e | FR-030 | Addition | Detect mode at session start |
| 4f | New FR-036 | Addition | Server mode selection requirement |
| 4g | New FR-037 | Addition | MCP resource access protocol |
| 4h | New FR-038 | Addition | MCP prompt-assisted authoring |
| 5a | Key Entities | Update | Add mode, split tools/resources/prompts |
| 6a | Edge Cases | Addition | Advisory mode wizard attempt |
| 6b | Edge Cases | Addition | Partial MCP availability |
| 7a | SC-013 | Update | Include mode selection in install time |
| 7b | SC-014 | Update | Distinguish tools/resources/prompts |
| 7c | New SC-015 | Addition | Mode behavior verification |
| 8a | Assumptions | Addition | MCP protocol assumption |

---

## Code-Level Implications

These spec changes will require corresponding code updates.
The following are noted for planning but are out of scope for
this delta:

1. **`internal/consts/consts.go`**: Add constants for mode
   names (`MCPModeAdvisory`, `MCPModeArtifact`), resource
   URIs (`ResourceLexicon`, `ResourceSchemaDefinitions`),
   and the `--mode` flag name.

2. **`internal/mcp/client.go`**: The `GetLexicon()` and
   `GetSchemaDocs()` methods currently use `callTool()`. These
   should be refactored to use MCP resource read protocol
   methods instead of tool calls. A `ReadResource()` method
   should be added to the `Transport` interface.

3. **`internal/mcp/config.go`**: `EnsureMCPEntry()` and
   `EnsureMCPEntryPodman()` currently hardcode `"serve"` as
   the only arg. These must include `"--mode"` and the
   selected mode.

4. **`internal/session/session.go`**: The `AvailableTools`
   struct should be refactored to `AvailableCapabilities`
   with separate fields for tools, resources, and prompts.
   A `ServerMode` field should be added to `Session`.

5. **`internal/cli/wizard_prompt.go`**: Already correctly
   handles the case where MCP is unavailable. Needs to also
   check server mode — wizards should only be offered when
   mode is `artifact`.

6. **`internal/cli/setup.go`**: Must add mode selection prompt
   during installation flow.
