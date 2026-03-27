## ADDED Requirements

### Requirement: Open Playground button on tutorial cards
The system SHALL display an "Open Playground" button on each tutorial card in the TutorialSuggestions step that opens the Gemara Playground in a new browser tab.

#### Scenario: Button appears on tutorial card
- **WHEN** the TutorialSuggestions step renders a tutorial card
- **THEN** an "Open Playground" button is visible on the card alongside the existing "Open Tutorial" link

#### Scenario: Button opens playground with correct artifact type
- **WHEN** the user clicks "Open Playground" on the Threat Assessment Guide tutorial card
- **THEN** a new browser tab opens at `/playground?type=ThreatCatalog`

#### Scenario: Button opens playground for guidance tutorial
- **WHEN** the user clicks "Open Playground" on the Guidance Catalog Guide tutorial card
- **THEN** a new browser tab opens at `/playground?type=GuidanceCatalog`

#### Scenario: Button opens playground for control catalog tutorial
- **WHEN** the user clicks "Open Playground" on the Control Catalog Guide tutorial card
- **THEN** a new browser tab opens at `/playground?type=ControlCatalog`

#### Scenario: Button opens playground for policy tutorial
- **WHEN** the user clicks "Open Playground" on the Organizational Risk & Policy Guide tutorial card
- **THEN** a new browser tab opens at `/playground?type=Policy`

### Requirement: Artifact type mapping from tutorials
The system SHALL map each upstream tutorial to its primary artifact type for the playground link. The mapping SHALL be: Guidance Catalog Guide -> GuidanceCatalog, Threat Assessment Guide -> ThreatCatalog, Control Catalog Guide -> ControlCatalog, Organizational Risk & Policy Guide -> Policy.

#### Scenario: Tutorial without a mapped artifact type
- **WHEN** a tutorial card has no artifact types defined
- **THEN** the "Open Playground" button opens `/playground` with the default artifact type

### Requirement: Visual distinction between tutorial and playground actions
The system SHALL visually distinguish the "Open Tutorial" link and the "Open Playground" button so users understand the difference: "Open Tutorial" goes to the external tutorial, "Open Playground" opens the local editor.

#### Scenario: Buttons have distinct styling
- **WHEN** a tutorial card renders both actions
- **THEN** "Open Tutorial" appears as a link/secondary action and "Open Playground" appears as a distinct button with a code/editor icon or label indicating it opens an editor
