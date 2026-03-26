## ADDED Requirements

### Requirement: Client-side routing with react-router-dom
The system SHALL use react-router-dom to provide client-side routing, mapping `/` to the existing wizard flow and `/playground` to the Gemara Playground IDE.

#### Scenario: Root route renders wizard
- **WHEN** a user navigates to `/`
- **THEN** the existing 5-step wizard flow renders unchanged

#### Scenario: Playground route renders IDE
- **WHEN** a user navigates to `/playground`
- **THEN** the Gemara Playground IDE component renders with the default artifact type selected

### Requirement: Artifact type via query parameter
The system SHALL accept a `type` query parameter on the `/playground` route that pre-selects the artifact type in the editor.

#### Scenario: Query parameter pre-selects type
- **WHEN** a user navigates to `/playground?type=ThreatCatalog`
- **THEN** the playground opens with "Threat Catalog" selected in the sidebar and the Threat Catalog example loaded in the editor

#### Scenario: Invalid query parameter falls back to default
- **WHEN** a user navigates to `/playground?type=InvalidType`
- **THEN** the playground opens with the default artifact type (Control Catalog) selected

#### Scenario: No query parameter uses default
- **WHEN** a user navigates to `/playground` without a `type` parameter
- **THEN** the playground opens with the default artifact type (Control Catalog) selected

### Requirement: Code-split playground bundle
The system SHALL lazy-load the playground route so that the CodeMirror editor and playground components are not included in the initial bundle loaded by the wizard flow.

#### Scenario: Wizard does not load playground code
- **WHEN** a user navigates to `/` and uses the wizard
- **THEN** no CodeMirror or playground component JavaScript is downloaded

#### Scenario: Playground loads on demand
- **WHEN** a user navigates to `/playground`
- **THEN** the playground chunk is loaded on demand via dynamic import

### Requirement: Vite history fallback for SPA routing
The system SHALL configure Vite to serve `index.html` for all unmatched routes so that direct navigation to `/playground` works correctly in both development and production.

#### Scenario: Direct navigation to playground
- **WHEN** a user enters `http://localhost:5173/playground` directly in the browser address bar
- **THEN** the SPA loads and the playground route renders (not a 404)
