## ADDED Requirements

### Requirement: YAML editor with syntax highlighting
The system SHALL render a CodeMirror 6 editor instance configured for YAML syntax highlighting, bracket matching, line numbers, and indentation guides.

#### Scenario: Editor loads with YAML mode
- **WHEN** the playground page is opened
- **THEN** the editor area displays a CodeMirror 6 instance with YAML syntax highlighting enabled, line numbers visible, and bracket matching active

#### Scenario: Editor handles YAML indentation
- **WHEN** the user presses Enter after a YAML key ending with a colon
- **THEN** the editor auto-indents the next line by two spaces

### Requirement: Artifact type sidebar
The system SHALL display a sidebar with an "Artifacts" tab listing the five supported artifact types: Control Catalog, Guidance Catalog, Threat Catalog, Risk Catalog, and Policy.

#### Scenario: Sidebar displays artifact types
- **WHEN** the playground page is opened
- **THEN** the sidebar displays five selectable artifact type entries: Control Catalog, Guidance Catalog, Threat Catalog, Risk Catalog, and Policy

#### Scenario: User selects an artifact type
- **WHEN** the user clicks an artifact type in the sidebar
- **THEN** the editor loads the example artifact YAML for that type and the schema tab updates to show field documentation for that type

### Requirement: Example artifact loading
The system SHALL load pre-bundled example artifact YAML content from upstream Gemara test data when an artifact type is selected.

#### Scenario: Load example on type selection
- **WHEN** the user selects "Control Catalog" from the artifact type list
- **THEN** the editor content is replaced with the bundled example Control Catalog YAML (sourced from `gemaraproj/gemara/test/test-data/good-ccc.yaml`)

#### Scenario: Load example preserves previous edits warning
- **WHEN** the user has modified the editor content and selects a different artifact type
- **THEN** the system displays a confirmation prompt before replacing the editor content

### Requirement: Schema documentation panel
The system SHALL display a "Schema" tab in the sidebar showing field documentation for the currently selected artifact type, including field names, types, whether they are required, and brief descriptions.

#### Scenario: Schema tab shows field docs
- **WHEN** the user selects the "Schema" tab in the sidebar while "Threat Catalog" is the active artifact type
- **THEN** the panel displays the Threat Catalog schema fields with their names, types, required status, and descriptions

#### Scenario: Schema tab updates on type change
- **WHEN** the user switches the artifact type from "Threat Catalog" to "Policy"
- **THEN** the schema tab content updates to show Policy schema fields

### Requirement: Lexicon panel
The system SHALL display a "Lexicon" tab in the sidebar showing Gemara term definitions relevant to the current artifact type.

#### Scenario: Lexicon tab shows relevant terms
- **WHEN** the user selects the "Lexicon" tab while editing a Control Catalog artifact
- **THEN** the panel displays Gemara lexicon terms relevant to controls (e.g., "control", "assessment-requirement", "threat", "guideline")

#### Scenario: Lexicon tab shows all terms
- **WHEN** the user selects the "Lexicon" tab
- **THEN** all Gemara lexicon terms are accessible, with terms relevant to the current artifact type highlighted or sorted first

### Requirement: Copy to clipboard
The system SHALL provide a "Copy to Clipboard" button that copies the current editor content.

#### Scenario: Copy editor content
- **WHEN** the user clicks the "Copy to Clipboard" button
- **THEN** the full editor content is copied to the system clipboard and a brief success notification is displayed

### Requirement: Editor theme support
The system SHALL render the editor in a theme consistent with the existing web UI's light/dark mode support.

#### Scenario: Dark mode rendering
- **WHEN** the user's system preference is set to dark mode
- **THEN** the editor renders with a dark background and light text using a CodeMirror dark theme

#### Scenario: Light mode rendering
- **WHEN** the user's system preference is set to light mode
- **THEN** the editor renders with a light background and dark text using a CodeMirror light theme
