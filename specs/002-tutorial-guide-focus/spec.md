# Feature Specification: Refocus Pac-Man as Tutorial Guide

**Feature Branch**: `002-tutorial-guide-focus`  
**Created**: 2026-03-17  
**Status**: Draft  
**Input**: User description: "Pacman should not replicate what already exists as the Gemara MCP Server, but instead be used as a guiding tool that will allow users to identify their activities, desired outputs, and leverage the tutorials for a walkthrough of the authoring procedure prior to using the MCP Server for assisted authoring. The latest release should be used and the functionality to change the mcp server version should be removed. It can be kept as a planned implementation, but should not be prioritized."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Activity and Output Identification (Priority: P1)

A user opens Pac-Man and describes their role and daily activities. Pac-Man identifies which Gemara layers are relevant, what artifact types the user will likely need to produce, and presents a clear summary of recommended tutorials and expected outputs before the user begins any authoring work. The user understands what they will create and why before ever touching the MCP server.

**Why this priority**: This is the foundational capability that differentiates Pac-Man from the MCP server. Without the ability to guide users through self-identification of activities and desired outputs, Pac-Man has no distinct purpose. This story must work independently to deliver immediate value.

**Independent Test**: Can be fully tested by a user stating their role and activities and receiving a tailored summary of recommended tutorials and expected artifact outputs without any MCP server interaction.

**Acceptance Scenarios**:

1. **Given** a user has launched Pac-Man, **When** they describe their role (e.g., "I'm a Security Engineer working on CI/CD pipeline security"), **Then** Pac-Man identifies the relevant Gemara layers (e.g., L2: Threats & Controls, L4: Sensitive Activities) and lists the artifact types they are expected to produce (e.g., Threat Catalog, Control Catalog).
2. **Given** a user has described their activities, **When** Pac-Man resolves their activity profile, **Then** Pac-Man presents a recommended learning path showing which tutorials to follow and in what order, with a brief explanation of why each tutorial matters for their role.
3. **Given** a user has received their activity summary, **When** they review the recommended outputs, **Then** each recommended artifact type includes a one-sentence description of what it is and how it relates to the user's stated activities.
4. **Given** a user describes activities that span multiple Gemara layers, **When** Pac-Man resolves ambiguous keywords (e.g., "evidence collection" could be L1 or L3), **Then** Pac-Man asks the user a targeted clarification question to determine the correct layer before proceeding.

---

### User Story 2 - Tutorial Walkthrough Before Authoring (Priority: P2)

After identifying their activities and desired outputs, the user follows a guided tutorial walkthrough. Pac-Man presents tutorial content section by section, tailored to the user's role, highlighting sections most relevant to their stated activities. The walkthrough explains the authoring procedure conceptually so the user understands the structure and purpose of each artifact field before they begin authoring with the MCP server.

**Why this priority**: The tutorial walkthrough is the core educational experience. It bridges the gap between "I know what I need to create" (US1) and "I'm ready to create it with the MCP server." Without this, users would jump from identification directly to authoring without understanding the process.

**Independent Test**: Can be fully tested by navigating through a tutorial for a specific artifact type and verifying that each section includes role-tailored explanations (Why/How/What), that relevant sections are highlighted, and that the user emerges with a conceptual understanding of what fields and decisions the authoring process will require.

**Acceptance Scenarios**:

1. **Given** a user has completed activity identification (US1) and selected a tutorial, **When** Pac-Man presents the first section, **Then** the section includes a "Why this matters for you" annotation tailored to the user's role and a "What you will learn" preview.
2. **Given** a user is navigating tutorial sections, **When** a section matches one of their stated activity keywords, **Then** that section is visually highlighted as a focus area and presented with additional context about how it relates to their specific work.
3. **Given** a user has completed a tutorial walkthrough, **When** they reach the end, **Then** Pac-Man presents a summary of what artifact type they are now prepared to author, what key decisions they will need to make during authoring, and a clear prompt to proceed to the MCP server for assisted authoring.
4. **Given** a user is mid-tutorial, **When** they want to skip to a different section, **Then** Pac-Man allows non-linear navigation while warning if skipped sections contain prerequisite concepts.

---

### User Story 3 - Latest Release Auto-Selection (Priority: P3)

When Pac-Man sets up the Gemara MCP server connection, it automatically uses the latest available release without presenting the user with a version selection choice. The schema version prompt and version switching functionality are removed from the active user flow. The user no longer needs to understand or choose between Stable and Latest versions; Pac-Man defaults to the latest release.

**Why this priority**: This simplifies the setup experience by removing a decision point that creates friction and confusion for users who are focused on learning, not version management. It supports the feature's goal of making Pac-Man a guide rather than a configuration tool. However, it depends on US1 and US2 already working to have impact.

**Independent Test**: Can be fully tested by running Pac-Man setup and verifying that the MCP server is configured with the latest release automatically, no version selection prompt is displayed, and the session proceeds directly to role and activity discovery.

**Acceptance Scenarios**:

1. **Given** a user launches Pac-Man for the first time, **When** the setup flow begins, **Then** the system automatically resolves and uses the latest Gemara release without presenting a version selection prompt.
2. **Given** the latest release has been resolved, **When** the MCP server is configured, **Then** the opencode.json configuration reflects the latest release and the session records the selected version without user intervention.
3. **Given** a user has previously completed setup with an older version, **When** they launch Pac-Man again and a newer release is available, **Then** the system automatically updates to the latest release and informs the user of the update.
4. **Given** the upstream release endpoint is unreachable, **When** the system cannot fetch release information, **Then** the system falls back to the most recently cached release and informs the user that offline mode is in effect.

---

### User Story 4 - Clear Handoff to MCP Server for Authoring (Priority: P4)

After completing the tutorial walkthrough in Pac-Man's terminal, the user transitions to OpenCode with the gemara-mcp server for assisted authoring. Pac-Man provides a clear handoff summary directing the user to open an OpenCode session and leverage the gemara-mcp server's tools (`validate_gemara_artifact`), resources (`gemara://lexicon`, `gemara://schema/definitions`), and wizard prompts (`threat_assessment`, `control_catalog`) for artifact creation. Pac-Man does not replicate these capabilities; it prepares the user with context so they arrive in OpenCode ready to author.

**Why this priority**: This story enforces the boundary between Pac-Man (terminal-based tutorial guide) and OpenCode + gemara-mcp (AI-assisted authoring). It prevents scope creep into authoring territory while ensuring users have a smooth, well-informed transition. It depends on US2 (tutorial walkthrough) having been experienced by the user.

**Independent Test**: Can be fully tested by completing a tutorial walkthrough and verifying that Pac-Man provides a handoff summary directing the user to OpenCode, naming the specific gemara-mcp prompts/tools/resources to use, the artifact type to create, and the key decisions the user should have pre-considered from the tutorial.

**Acceptance Scenarios**:

1. **Given** a user has completed a tutorial walkthrough for a specific artifact type, **When** they indicate readiness to begin authoring, **Then** Pac-Man presents a handoff summary directing them to open an OpenCode session, listing the specific gemara-mcp prompt to use (e.g., `threat_assessment` for Threat Catalogs), the schema definition for validation (e.g., `#ThreatCatalog`), available MCP resources (lexicon, schema docs), and key decisions the user should have answers for based on the tutorial.
2. **Given** a user is at the handoff point, **When** the gemara-mcp server is configured in `opencode.json`, **Then** Pac-Man confirms the configuration is present and instructs the user to launch `opencode` to begin authoring with full MCP tool and resource access.
3. **Given** a user is at the handoff point, **When** the gemara-mcp server is not configured, **Then** Pac-Man instructs the user to run `./pacman --doctor` to verify their environment, explains how to configure the MCP server in `opencode.json`, and provides the manual `cue vet` validation command as an alternative until the server is set up.

---

### User Story 5 - Deprecate Version Switching as Future Work (Priority: P5)

The existing schema version selection and mid-session version switching functionality is removed from the active user flow and documented as a planned future enhancement. The code is not deleted but is bypassed in the main flow. A clear marker in the codebase and documentation indicates this is intentional deferral, not an oversight.

**Why this priority**: This is a cleanup and documentation story. It has the lowest priority because the functional impact is achieved by US3 (auto-selecting latest). This story ensures the codebase is clean and the deferral is intentional and documented.

**Independent Test**: Can be fully tested by verifying that no version selection prompt appears during setup, that the version selection code is bypassed (not deleted), and that a decision record or code comment documents the intentional deferral with rationale.

**Acceptance Scenarios**:

1. **Given** the setup flow is executed, **When** the system reaches the point where version selection previously occurred, **Then** the system skips the version prompt entirely and proceeds with the latest release.
2. **Given** a developer reviews the codebase, **When** they find the version selection code, **Then** a clear comment or decision record explains that version switching is a planned future enhancement that has been intentionally deferred.
3. **Given** the existing version selection and switching functions still exist in the codebase, **When** a future developer wants to re-enable version switching, **Then** they can do so by removing the bypass without reimplementing the functionality.

---

### Edge Cases

- What happens when no tutorials exist for the user's identified layers? Pac-Man should inform the user which layers lack tutorials and suggest checking back as content is added, while still presenting any available tutorials.
- What happens when a user's described activities produce no keyword matches? Pac-Man should fall back to the user's role defaults and explain that no specific activity keywords were detected, offering the category-based selection as an alternative.
- What happens when the latest release has experimental or draft schemas? Pac-Man should proceed with the latest release but inform the user which schemas are experimental, noting that authoring for those schemas may produce artifacts that require updates when schemas stabilize.
- What happens when the user wants to author without completing a tutorial? Pac-Man should allow it but recommend completing the tutorial first, providing a brief explanation of what they may miss. If the user proceeds, direct them to OpenCode with the gemara-mcp server and the relevant MCP prompt name.
- What happens when the MCP server's built-in schema version differs from the auto-selected latest? Pac-Man should note the discrepancy in the handoff summary and recommend the user validate artifacts after authoring.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST present the user with an activity and output identification flow that maps their stated role and activities to specific Gemara layers, artifact types, and recommended tutorials.
- **FR-002**: System MUST generate a recommended learning path ordered by relevance to the user's identified layers, with each step annotated with Why (relevance to role), How (application to daily work), and What (learning outcomes and artifact outputs).
- **FR-003**: System MUST list the specific artifact types the user is expected to produce based on their activity profile, with a brief description of each artifact type and its relationship to the user's work.
- **FR-004**: System MUST present tutorial content section by section with role-tailored annotations and visual highlighting of sections that match the user's stated activity keywords.
- **FR-005**: System MUST support non-linear tutorial navigation with prerequisite warnings when users skip sections containing foundational concepts.
- **FR-006**: System MUST present a post-tutorial summary that lists the artifact type the user is prepared to author, key decisions they need to make, and the specific MCP server prompt or tool to use for authoring.
- **FR-007**: System MUST automatically resolve and use the latest Gemara release during setup without presenting a version selection prompt to the user.
- **FR-008**: System MUST fall back to the most recently cached release when the upstream release endpoint is unreachable, and inform the user that offline mode is in effect.
- **FR-009**: System MUST bypass the existing schema version selection prompt and mid-session version switching flows, proceeding directly from MCP setup to role and activity discovery.
- **FR-010**: System MUST NOT replicate the MCP server's authoring wizards (threat_assessment, control_catalog prompts) or artifact validation within Pac-Man's guided flow.
- **FR-011**: System MUST provide a clear handoff point after tutorial completion that directs the user to open an OpenCode session with the gemara-mcp server, including the specific prompt name, schema definition, available MCP resources (lexicon, schema docs), and a preparation checklist.
- **FR-017**: The `--doctor` command MUST remain fully functional and unchanged. It continues to verify the user's environment (Go version, CUE installation, gemara-mcp server availability, opencode.json configuration) and report actionable status for each check.
- **FR-018**: All terminal output MUST be user-friendly and visually polished for all audiences, including non-technical stakeholders, security engineers, compliance officers, and developers. Output MUST use consistent styling (colors, spacing, icons, card layouts), avoid jargon without context, and present information in a scannable format with clear visual hierarchy.
- **FR-019**: The post-tutorial handoff MUST explicitly direct users to OpenCode as the authoring environment, referencing the gemara-mcp server's available tools (`validate_gemara_artifact`), resources (`gemara://lexicon`, `gemara://schema/definitions`), and wizard prompts (`threat_assessment`, `control_catalog`) by name so users know exactly what capabilities are available to them.
- **FR-012**: System MUST inform the user when no tutorials are available for their identified layers and suggest alternative actions.
- **FR-013**: System MUST handle ambiguous activity keywords by prompting the user with a targeted clarification question to resolve the correct layer mapping.
- **FR-014**: System MUST retain the existing version selection and switching code in the codebase, bypassed but not deleted, with documentation explaining the intentional deferral.
- **FR-015**: System MUST automatically update to the latest release when a newer version is available on subsequent launches, informing the user of the update.
- **FR-016**: System MUST warn the user when the MCP server's built-in schema version differs from the auto-selected latest release, noting the discrepancy in the handoff summary.

### Key Entities

- **Activity Profile**: Represents a user's resolved set of Gemara layers, extracted keywords, matched categories, and role context. Determines which tutorials and artifact types are recommended.
- **Learning Path**: An ordered sequence of tutorial steps tailored to the user's activity profile. Each step includes a tutorial reference, layer association, relevance annotations (Why/How/What), section relevance scores, and completion status.
- **Handoff Summary**: A structured transition point presented after tutorial completion. Directs the user to OpenCode with the gemara-mcp server, containing the target artifact type, the specific MCP prompt or tool to use, available MCP resources (lexicon, schema definitions), the schema definition for validation, and a list of key decisions the user should have pre-considered.
- **Release Resolution**: The process of automatically determining the latest Gemara release. Includes the resolved version tag, whether it was fetched live or from cache, and any experimental schema warnings.

## Assumptions

- OpenCode is the preferred AI development harness and the destination for all authoring work after Pac-Man tutorials. Users complete tutorials in the Pac-Man terminal, then switch to OpenCode where the gemara-mcp server provides tools, resources, and wizard prompts for artifact creation.
- The Gemara MCP server will continue to be the primary authoring tool and will maintain its current wizard prompts (`threat_assessment`, `control_catalog`), resources (`gemara://lexicon`, `gemara://schema/definitions`), and validation capabilities (`validate_gemara_artifact`). Pac-Man does not need to provide these.
- The upstream Gemara repository at `gemaraproj/gemara` will continue to host tutorial content in `docs/tutorials/` and publish releases via GitHub's releases API.
- "Latest release" means the most recent release by date from the Gemara repository, consistent with the existing `DetermineLatestVersion()` behavior. This includes prereleases if no stable release exists.
- The `--doctor` command is an existing, independent diagnostic tool that verifies the user's environment. It is not modified by this feature and continues to work as-is.
- The existing version selection code (`internal/schema/selector.go`, `internal/cli/version_prompt.go`) is mature enough to be preserved for future re-enablement without requiring reimplementation.
- The MCP server version (gemara-mcp binary release) is a separate concern from the Gemara schema version. This specification addresses schema version selection only; MCP server installation continues to use the latest release as it does today.
- Pac-Man's terminal output is consumed by all audiences (security engineers, compliance officers, CISOs, developers, policy authors, auditors). Output must be accessible and polished regardless of the user's technical depth.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can identify their activities, see their recommended tutorials, and understand their expected artifact outputs within 5 minutes of launching Pac-Man.
- **SC-002**: 90% of users who complete a tutorial walkthrough can correctly name the MCP server prompt or tool they need to use for authoring without additional guidance.
- **SC-003**: The setup flow completes without presenting any version selection prompts, reducing the number of user decision points during setup by at least one compared to the current flow.
- **SC-004**: Users who complete the tutorial-to-authoring handoff report that they felt prepared to begin authoring, with at least 80% indicating they understood the key decisions required.
- **SC-005**: No Pac-Man flow duplicates functionality available in the MCP server's wizard prompts or validation tools, verified by review of Pac-Man's guided flow output against MCP server capabilities.
- **SC-006**: The version selection code remains functional in the codebase and can be re-enabled by a developer within 1 hour of effort, verified by the presence of bypass documentation and intact code.
- **SC-007**: All terminal output uses consistent visual styling (card layouts, color-coded labels, clear spacing) and is readable by users who are not software developers, verified by review of output screenshots across all flows (activity identification, tutorial navigation, handoff summary).
- **SC-008**: The post-tutorial handoff summary explicitly names OpenCode as the authoring environment and lists at least the relevant gemara-mcp prompt, available resources, and validation tool, verified by inspecting the rendered handoff output for each artifact type.
- **SC-009**: The `--doctor` command continues to function correctly after all changes, verified by running `./pacman --doctor` and confirming all environment checks pass.
