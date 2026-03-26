## ADDED Requirements

### Requirement: MCP connection prompt
The system SHALL display a non-blocking prompt in the playground offering the user the option to connect the gemara-mcp server for live validation. The prompt SHALL be dismissible and SHALL NOT prevent any playground functionality.

#### Scenario: MCP prompt appears on first visit
- **WHEN** the user opens the playground for the first time
- **THEN** a non-modal banner or panel displays a message explaining that connecting the gemara-mcp server enables live artifact validation, with a "Connect" button and a "Dismiss" option

#### Scenario: Dismissing the prompt
- **WHEN** the user clicks "Dismiss" on the MCP connection prompt
- **THEN** the prompt is hidden and all playground functionality remains available

### Requirement: MCP connection status indicator
The system SHALL display a connection status indicator showing whether the gemara-mcp server is connected, disconnected, or unavailable.

#### Scenario: Status shows disconnected by default
- **WHEN** the playground loads and no MCP connection has been established
- **THEN** the status indicator shows "MCP: Not Connected" in a neutral style

#### Scenario: Status shows connected after successful connection
- **WHEN** the user successfully connects to the gemara-mcp server
- **THEN** the status indicator updates to "MCP: Connected" in a positive/green style

#### Scenario: Status shows error on connection failure
- **WHEN** the user attempts to connect but the MCP server is unavailable
- **THEN** the status indicator shows "MCP: Connection Failed" with an error message explaining the server may not be running

### Requirement: Validate with MCP button
The system SHALL provide a "Validate" button that sends the current editor content to the gemara-mcp server's `validate_gemara_artifact` tool for validation. This button SHALL only be active when MCP is connected.

#### Scenario: Validate button disabled without MCP
- **WHEN** the MCP server is not connected
- **THEN** the "Validate" button is disabled with a tooltip indicating MCP connection is required

#### Scenario: Successful validation
- **WHEN** the user clicks "Validate" with valid YAML content and the MCP server is connected
- **THEN** the validation result is displayed in a results panel showing success

#### Scenario: Validation with errors
- **WHEN** the user clicks "Validate" with invalid YAML content and the MCP server is connected
- **THEN** the validation errors are displayed in a results panel with error messages from the MCP server

### Requirement: MCP functionality isolation
The system SHALL keep all MCP communication logic in a dedicated module, separate from the core editor, sidebar, and routing components. No playground feature other than the "Validate" button SHALL depend on MCP connectivity.

#### Scenario: Playground functions without MCP
- **WHEN** the MCP server is not running or not connected
- **THEN** the editor, sidebar, artifact loading, schema docs, lexicon, and clipboard copy all function normally

#### Scenario: MCP module is independently importable
- **WHEN** the playground code is organized into modules
- **THEN** all MCP-related code resides in a dedicated file or directory (e.g., `playground/mcp/`) that is only imported by the MCP integration UI components
