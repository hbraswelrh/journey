## Context

The Gemara User Journey web UI is a static React 19 + Vite 8 SPA that guides users through a 5-step wizard: role selection, activity probing, results, tutorial suggestions, and MCP walkthrough. All data is baked in at build time via a Go-to-TypeScript codegen pipeline (`cmd/genwebdata` -> `web/src/generated/journey-data.ts`). The app currently has no client-side routing, no code editor components, and links to upstream tutorials at `gemara.openssf.org` as external URLs only.

Users who complete the wizard and reach the tutorial suggestions step have no way to practice creating or editing Gemara artifacts within the web experience. They must switch to a separate tool (text editor + CLI) to author YAML, then manually validate with `cue vet` or the gemara-mcp server. This friction reduces the learn-by-doing value of the journey.

The upstream Gemara project provides well-structured test data files (`test/test-data/good-*.yaml`) that serve as canonical examples of valid artifacts across all schema types.

## Goals / Non-Goals

**Goals:**
- Provide a lightweight, browser-based YAML editor (the "Gemara Playground") where users can view, edit, and experiment with Gemara artifacts.
- Pre-load example artifacts from upstream test data so users start from working samples.
- Include a sidebar with artifact type selection, schema field documentation, and relevant lexicon terms.
- Launch the playground from tutorial cards via a button that opens a new tab, pre-configured for the tutorial's artifact type.
- Offer optional MCP server integration for live validation, kept fully isolated so the playground works without it.
- Maintain the existing wizard flow unchanged; the playground is purely additive.

**Non-Goals:**
- Full IDE features (file system access, multi-file projects, git integration, terminal).
- Server-side processing or storage of user artifacts. Everything remains client-side.
- Replacing the gemara-mcp server's validation with a client-side CUE validator. Client-side provides schema reference only; real validation requires MCP.
- Rendering upstream tutorial content within the playground. Tutorials remain on `gemara.openssf.org`; the playground is a companion tool.
- Supporting artifact types beyond the five primary types (Control Catalog, Guidance Catalog, Threat Catalog, Risk Catalog, Policy) in the initial release.

## Decisions

### 1. CodeMirror 6 as the editor engine

**Choice**: CodeMirror 6 (`@codemirror/view`, `@codemirror/state`, `@codemirror/lang-yaml`)

**Rationale**: CodeMirror 6 is modular, tree-shakeable, and purpose-built for embedding in web apps. It has first-class YAML language support, accessibility features, and a small bundle footprint (~80-120KB gzipped for the modules we need). Monaco Editor was considered but rejected due to its ~2MB bundle size and being designed for full IDE scenarios. The playground needs to be lightweight -- users open it in a second tab alongside a tutorial.

**Alternatives considered**:
- Monaco Editor: Too heavy, designed for VS Code-level use cases.
- Ace Editor: Older API, less modular, weaker TypeScript support.
- Plain `<textarea>`: Insufficient -- no syntax highlighting, no bracket matching, poor UX for YAML indentation.

### 2. Client-side routing with react-router-dom

**Choice**: Add `react-router-dom` v7 for client-side routing.

**Rationale**: The playground must open in a new tab via a URL (`/playground?type=ControlCatalog`). The existing wizard flow uses state-driven step rendering with no URLs. Adding a router lets us assign `/` to the wizard and `/playground` to the IDE without breaking the existing flow. Query parameters pass the pre-selected artifact type.

**Alternatives considered**:
- Separate Vite entry point (`playground.html`): Would complicate the build pipeline and share no routing infrastructure.
- Hash-based routing without a library: Fragile, no parameter parsing, poor DX.

### 3. Build-time bundling of example artifacts and schema docs

**Choice**: Extend `cmd/genwebdata` to embed example artifact YAML content and schema field documentation into the generated TypeScript data module.

**Rationale**: The existing data pipeline already converts Go constants to TypeScript at build time. Adding example artifacts and schema docs to this pipeline keeps the architecture consistent -- no runtime fetching, no CORS issues, no external service dependencies. Example YAML files from `gemaraproj/gemara/test/test-data/` will be fetched during `make web-data` and inlined as string constants.

**Alternatives considered**:
- Fetching examples from GitHub raw URLs at runtime: Would break offline use, add latency, and create a dependency on GitHub availability.
- Committing copies of test data files into this repo: Would create drift if upstream changes. Build-time fetch is better.

### 4. Isolated MCP integration behind an explicit boundary

**Choice**: The playground renders a connection prompt and a "Validate with MCP" button. MCP communication uses a dedicated `PlaygroundMCPClient` class that is instantiated only when the user opts in. All playground features work without MCP.

**Rationale**: The AGENTS.md explicitly states "the functionality of the gemara-mcp server should remain isolated." The playground must be fully functional without the MCP server running. The MCP integration adds one capability: live validation via `validate_gemara_artifact`. This is surfaced as an optional "Validate" action, clearly separated from the core editor experience.

**Alternatives considered**:
- Always requiring MCP: Violates the isolation requirement and adds a hard dependency.
- No MCP integration at all: Misses an opportunity to show the value of the MCP server in context.

### 5. Sidebar with tab-based navigation

**Choice**: A left sidebar with three tabs: **Artifacts** (type selector + load example), **Schema** (field docs for current type), **Lexicon** (Gemara term definitions).

**Rationale**: Users need three types of reference while editing: what artifact types exist, what fields the current schema expects, and what Gemara terms mean. A tabbed sidebar keeps this accessible without cluttering the editor. The sidebar collapses on narrow viewports for mobile friendliness.

### 6. No persistent storage in initial release

**Choice**: Editor content lives in React state only. No localStorage, no IndexedDB, no file download. Users can copy content from the editor.

**Rationale**: The playground is for experimentation and learning, not production authoring. Adding persistence adds complexity (conflict resolution, storage limits, migration). A future enhancement can add localStorage or file download if users request it. The editor will include a "Copy to Clipboard" button as the primary export mechanism.

## Risks / Trade-offs

- **[Bundle size increase]** -> CodeMirror 6 adds ~100-150KB gzipped. Mitigated by code-splitting the playground route so the editor is only loaded when `/playground` is visited. The wizard flow bundle is unaffected.
- **[Example data staleness]** -> Test data fetched at build time may drift from upstream. Mitigated by documenting the source commit/version and providing a `make web-data-refresh` target.
- **[YAML validation accuracy without MCP]** -> Client-side schema docs show field names and types but cannot validate structural correctness. Mitigated by clearly labeling the schema panel as "reference" and prompting users to validate via MCP for authoritative results.
- **[react-router-dom adds routing complexity]** -> The wizard flow currently uses simple state. Mitigated by keeping the wizard at `/` with its existing state logic untouched; the router wraps it without changing its internals.
- **[MCP connection reliability]** -> The MCP server runs locally; connectivity issues may confuse users. Mitigated by clear status indicators and error messages in the MCP prompt UI.
