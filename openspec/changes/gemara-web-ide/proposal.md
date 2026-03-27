## Why

The Gemara User Journey web UI currently links users to upstream tutorials on `gemara.openssf.org` as external URLs, but provides no way for users to interactively explore or edit Gemara artifacts alongside those tutorials. Users must context-switch between the tutorial content and a separate editor to create or modify YAML artifacts (control catalogs, guidance catalogs, threat catalogs, risk catalogs, policies). A lightweight, in-browser IDE that loads alongside the tutorial step would let users learn-by-doing: reading the tutorial in one tab while editing real Gemara YAML in another, with schema awareness and live validation feedback.

## What Changes

- Add a new **Gemara Playground** web IDE component: a lightweight, browser-based YAML editor with a sidebar for selecting artifact types (Control Catalog, Guidance Catalog, Threat Catalog, Risk Catalog, Policy).
- Integrate a code editor library (CodeMirror 6) for syntax highlighting, YAML-aware editing, and schema-guided autocompletion hints.
- Pre-load example artifacts from the Gemara project test data (`gemaraproj/gemara/tree/main/test/test-data`) so users can start from working samples rather than blank files.
- Add a **sidebar** that lets users switch between artifact types, view schema documentation (field descriptions, required fields), and see the Gemara lexicon terms relevant to the current artifact.
- Add an **"Open Playground"** button on each tutorial card in the `TutorialSuggestions` step that opens the web IDE in a new browser tab, pre-configured with the artifact type that tutorial covers.
- The web IDE must be a **standalone route** (e.g., `/playground`) that works independently so users can follow the upstream tutorial in one tab and edit in the playground tab.
- Provide an **optional MCP integration prompt**: when a user opens the playground, offer a non-blocking prompt to connect the `gemara-mcp` server for live validation. If the MCP server is not connected, the playground still functions fully with client-side schema reference and example loading. MCP functionality (validation via `validate_gemara_artifact`) remains isolated behind a clear integration boundary.
- Extend the `cmd/genwebdata` pipeline to export schema documentation and example artifact data so the playground can consume them at build time.

## Capabilities

### New Capabilities
- `playground-editor`: The core web IDE component with CodeMirror 6 YAML editor, artifact type selector sidebar, schema documentation panel, and example artifact loading.
- `playground-routing`: Standalone `/playground` route using client-side routing so the IDE opens in a new tab with artifact-type query parameters.
- `tutorial-playground-link`: "Open Playground" button on tutorial cards that launches the playground pre-configured for that tutorial's artifact type.
- `playground-mcp-integration`: Optional, isolated MCP server connection for live artifact validation within the playground.
- `playground-data-pipeline`: Extension of `cmd/genwebdata` to export schema docs, example artifacts, and lexicon data for the playground.

### Modified Capabilities

## Impact

- **Web frontend (`web/`)**: New components, new route, new dependency (CodeMirror 6). The existing wizard flow is unchanged; the playground is additive.
- **`cmd/genwebdata/`**: Must generate additional data (schema field docs, example artifacts, lexicon terms) into `web/src/generated/`.
- **`internal/consts/`**: May need new constants for artifact schema field descriptions and example artifact URLs/content.
- **Build pipeline (`Makefile`)**: The `web-data` target must fetch or bundle example artifact YAML files from upstream test data.
- **Dependencies**: Adds `@codemirror/lang-yaml`, `@codemirror/view`, `@codemirror/state`, and related CodeMirror 6 packages to `web/package.json`. Adds `react-router-dom` for client-side routing.
- **Bundle size**: CodeMirror 6 is modular and tree-shakeable; expected addition ~150-200KB gzipped. No server-side changes required.
