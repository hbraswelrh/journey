## 1. Data Pipeline Extension

- [x] 1.1 Add playground example artifact constants to `internal/consts/consts.go` mapping artifact type identifiers to upstream test data file names (`good-ccc.yaml`, `good-aigf.yaml`, `good-threat-catalog.yaml`, `good-risk-catalog.yaml`, `good-policy.yaml`)
- [x] 1.2 Add schema field documentation constants to `internal/consts/consts.go` for the five supported artifact types (ControlCatalog, GuidanceCatalog, ThreatCatalog, RiskCatalog, Policy) with field names, types, required flags, and descriptions
- [x] 1.3 Add lexicon term constants to `internal/consts/consts.go` with term name, definition, and relevant artifact type tags
- [x] 1.4 Extend `cmd/genwebdata/main.go` to fetch example artifact YAML files from GitHub raw URLs during build and embed them as string constants in the generated TypeScript module under a `playgroundExamples` key
- [x] 1.5 Extend `cmd/genwebdata/main.go` to export `playgroundSchemas` and `playgroundLexicon` data into the generated TypeScript module
- [x] 1.6 Add build error handling: fail `make web-data` with clear error message if any upstream test data file cannot be fetched

## 2. Install Dependencies and Configure Routing

- [x] 2.1 Install `react-router-dom` v7 in `web/package.json`
- [x] 2.2 Install CodeMirror 6 packages: `@codemirror/view`, `@codemirror/state`, `@codemirror/lang-yaml`, `@codemirror/language`, `codemirror`, `@codemirror/theme-one-dark`
- [x] 2.3 Wrap the App component with `BrowserRouter` in `web/src/main.tsx`
- [x] 2.4 Add route definitions: `/` renders the existing wizard (App component), `/playground` renders the new Playground component via lazy import
- [x] 2.5 Configure Vite history fallback for SPA routing so direct navigation to `/playground` works

## 3. Playground Editor Component

- [x] 3.1 Create `web/src/components/playground/PlaygroundEditor.tsx` with CodeMirror 6 editor instance configured for YAML mode, line numbers, bracket matching, and auto-indentation
- [x] 3.2 Implement light/dark theme support for the editor using CSS custom properties and CodeMirror theme extensions, matching the existing app theme system (`prefers-color-scheme`)
- [x] 3.3 Create `web/src/components/playground/PlaygroundLayout.tsx` as the top-level playground component with sidebar + editor layout

## 4. Sidebar Components

- [x] 4.1 Create `web/src/components/playground/sidebar/ArtifactSelector.tsx` listing the five artifact types with click-to-select behavior and active state styling
- [x] 4.2 Create `web/src/components/playground/sidebar/SchemaPanel.tsx` rendering schema field documentation for the active artifact type from the generated data
- [x] 4.3 Create `web/src/components/playground/sidebar/LexiconPanel.tsx` rendering Gemara lexicon terms, sorting relevant terms first based on the active artifact type
- [x] 4.4 Create `web/src/components/playground/sidebar/Sidebar.tsx` with tabbed navigation between Artifacts, Schema, and Lexicon panels
- [x] 4.5 Implement sidebar collapse behavior for narrow viewports

## 5. Editor Toolbar and Actions

- [x] 5.1 Create `web/src/components/playground/EditorToolbar.tsx` with artifact type label, "Copy to Clipboard" button, and MCP status indicator
- [x] 5.2 Implement "Copy to Clipboard" functionality using the Clipboard API with success notification
- [x] 5.3 Implement confirmation dialog when switching artifact types with unsaved editor changes

## 6. Playground Routing and Query Parameters

- [x] 6.1 Create `web/src/components/playground/Playground.tsx` that reads the `type` query parameter and pre-selects the corresponding artifact type
- [x] 6.2 Handle invalid or missing `type` query parameter by falling back to the default (ControlCatalog)
- [x] 6.3 Verify code-splitting works: confirm the playground chunk is not loaded on the `/` route

## 7. Tutorial Playground Link

- [x] 7.1 Add artifact type mapping to the `UpstreamTutorial` type in the generated data (tutorial ID -> primary artifact type identifier)
- [x] 7.2 Add "Open Playground" button to `TutorialCard` in `web/src/components/TutorialSuggestions.tsx` that opens `/playground?type=<ArtifactType>` in a new tab
- [x] 7.3 Style the "Open Playground" button distinctly from the existing "Open Tutorial" link with an editor/code icon

## 8. MCP Integration (Isolated)

- [x] 8.1 Create `web/src/components/playground/mcp/MCPClient.ts` with a `PlaygroundMCPClient` class encapsulating connection and validation logic
- [x] 8.2 Create `web/src/components/playground/mcp/MCPBanner.tsx` with the non-blocking connection prompt, "Connect" button, and "Dismiss" option
- [x] 8.3 Create `web/src/components/playground/mcp/MCPStatus.tsx` connection status indicator component (Not Connected / Connected / Connection Failed states)
- [x] 8.4 Add "Validate" button to `EditorToolbar.tsx` that is disabled when MCP is not connected, and sends editor content to `validate_gemara_artifact` when clicked
- [x] 8.5 Create `web/src/components/playground/mcp/ValidationResults.tsx` panel to display validation success or error messages below the editor

## 9. Styling and Polish

- [x] 9.1 Add playground-specific CSS styles to `web/src/App.css` or a new `web/src/Playground.css` for the sidebar layout, editor area, toolbar, and panel styling
- [x] 9.2 Ensure consistent use of CSS custom properties from `web/src/index.css` for theme compatibility
- [x] 9.3 Add responsive breakpoints: sidebar collapses to icons on narrow viewports, editor takes full width on mobile

## 10. Verification

- [x] 10.1 Run `make web-data` and verify `playgroundExamples`, `playgroundSchemas`, and `playgroundLexicon` appear in the generated TypeScript output
- [x] 10.2 Run `make web-build` and verify the build succeeds with no TypeScript errors
- [x] 10.3 Run `make web-dev` and manually verify: wizard at `/` works unchanged, playground at `/playground` loads, artifact type switching works, example YAML loads, sidebar tabs work, "Copy to Clipboard" works, "Open Playground" button appears on tutorial cards
- [x] 10.4 Verify code-splitting: check that navigating to `/` does not download the playground chunk in the browser network tab
