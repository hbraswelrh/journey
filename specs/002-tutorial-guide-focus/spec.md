# Feature Specification: Refocus Gemara User Journey as Tutorial Guide

**Feature Branch**: `002-tutorial-guide-focus`  
**Created**: 2026-03-17  
**Status**: Draft  
**Input**: User description: "Pacman should not replicate what already exists as the Gemara MCP Server, but instead be used as a guiding tool that will allow users to identify their activities, desired outputs, and leverage the tutorials for a walkthrough of the authoring procedure prior to using the MCP Server for assisted authoring. The latest release should be used and the functionality to change the mcp server version should be removed. It can be kept as a planned implementation, but should not be prioritized."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Activity and Output Identification (Priority: P1)

A user opens Gemara User Journey and describes their role and daily activities. Gemara User Journey identifies which Gemara layers are relevant, what artifact types the user will likely need to produce, and presents a clear summary of recommended tutorials and expected outputs before the user begins any authoring work. The user understands what they will create and why before ever touching the MCP server.

**Why this priority**: This is the foundational capability that differentiates Gemara User Journey from the MCP server. Without the ability to guide users through self-identification of activities and desired outputs, Gemara User Journey has no distinct purpose. This story must work independently to deliver immediate value.

**Independent Test**: Can be fully tested by a user stating their role and activities and receiving a tailored summary of recommended tutorials and expected artifact outputs without any MCP server interaction.

**Acceptance Scenarios**:

1. **Given** a user has launched Gemara User Journey, **When** they describe their role (e.g., "I'm a Security Engineer working on CI/CD pipeline security"), **Then** Gemara User Journey identifies the relevant Gemara layers (e.g., L2: Threats & Controls, L4: Sensitive Activities) and lists the artifact types they are expected to produce (e.g., Threat Catalog, Control Catalog).
2. **Given** a user has described their activities, **When** Gemara User Journey resolves their activity profile, **Then** Gemara User Journey presents a recommended learning path showing which tutorials to follow and in what order, with a brief explanation of why each tutorial matters for their role.
3. **Given** a user has received their activity summary, **When** they review the recommended outputs, **Then** each recommended artifact type includes a one-sentence description of what it is and how it relates to the user's stated activities.
4. **Given** a user describes activities that span multiple Gemara layers, **When** Gemara User Journey resolves ambiguous keywords (e.g., "evidence collection" could be L1 or L3), **Then** Gemara User Journey asks the user a targeted clarification question to determine the correct layer before proceeding.

---

### User Story 2 - Tutorial Walkthrough Before Authoring (Priority: P2)

After identifying their activities and desired outputs, the user follows a guided tutorial walkthrough. Gemara User Journey presents tutorial content section by section, tailored to the user's role, highlighting sections most relevant to their stated activities. The walkthrough explains the authoring procedure conceptually so the user understands the structure and purpose of each artifact field before they begin authoring with the MCP server.

**Why this priority**: The tutorial walkthrough is the core educational experience. It bridges the gap between "I know what I need to create" (US1) and "I'm ready to create it with the MCP server." Without this, users would jump from identification directly to authoring without understanding the process.

**Independent Test**: Can be fully tested by navigating through a tutorial for a specific artifact type and verifying that each section includes role-tailored explanations (Why/How/What), that relevant sections are highlighted, and that the user emerges with a conceptual understanding of what fields and decisions the authoring process will require.

**Acceptance Scenarios**:

1. **Given** a user has completed activity identification (US1) and selected a tutorial, **When** Gemara User Journey presents the first section, **Then** the section includes a "Why this matters for you" annotation tailored to the user's role and a "What you will learn" preview.
2. **Given** a user is navigating tutorial sections, **When** a section matches one of their stated activity keywords, **Then** that section is visually highlighted as a focus area and presented with additional context about how it relates to their specific work.
3. **Given** a user has completed a tutorial walkthrough, **When** they reach the end, **Then** Gemara User Journey presents a summary of what artifact type they are now prepared to author, what key decisions they will need to make during authoring, and a clear prompt to proceed to the MCP server for assisted authoring.
4. **Given** a user is mid-tutorial, **When** they want to skip to a different section, **Then** Gemara User Journey allows non-linear navigation while warning if skipped sections contain prerequisite concepts.

---

### User Story 3 - Latest Release Auto-Selection (Priority: P3)

When Gemara User Journey sets up the Gemara MCP server connection, it automatically uses the latest available release without presenting the user with a version selection choice. The schema version prompt and version switching functionality are removed from the active user flow. The user no longer needs to understand or choose between Stable and Latest versions; Gemara User Journey defaults to the latest release.

**Why this priority**: This simplifies the setup experience by removing a decision point that creates friction and confusion for users who are focused on learning, not version management. It supports the feature's goal of making Gemara User Journey a guide rather than a configuration tool. However, it depends on US1 and US2 already working to have impact.

**Independent Test**: Can be fully tested by running Gemara User Journey setup and verifying that the MCP server is configured with the latest release automatically, no version selection prompt is displayed, and the session proceeds directly to role and activity discovery.

**Acceptance Scenarios**:

1. **Given** a user launches Gemara User Journey for the first time, **When** the setup flow begins, **Then** the system automatically resolves and uses the latest Gemara release without presenting a version selection prompt.
2. **Given** the latest release has been resolved, **When** the MCP server is configured, **Then** the opencode.json configuration reflects the latest release and the session records the selected version without user intervention.
3. **Given** a user has previously completed setup with an older version, **When** they launch Gemara User Journey again and a newer release is available, **Then** the system automatically updates to the latest release and informs the user of the update.
4. **Given** the upstream release endpoint is unreachable, **When** the system cannot fetch release information, **Then** the system falls back to the most recently cached release and informs the user that offline mode is in effect.

---

### User Story 4 - Clear Handoff to MCP Server for Authoring (Priority: P4)

After completing the tutorial walkthrough in Gemara User Journey's terminal, the user transitions to OpenCode with the gemara-mcp server for assisted authoring. Gemara User Journey provides a clear handoff summary directing the user to open an OpenCode session and leverage the gemara-mcp server's tools (`validate_gemara_artifact`), resources (`gemara://lexicon`, `gemara://schema/definitions`), and wizard prompts (`threat_assessment`, `control_catalog`) for artifact creation. Gemara User Journey does not replicate these capabilities; it prepares the user with context so they arrive in OpenCode ready to author.

**Why this priority**: This story enforces the boundary between Gemara User Journey (terminal-based tutorial guide) and OpenCode + gemara-mcp (AI-assisted authoring). It prevents scope creep into authoring territory while ensuring users have a smooth, well-informed transition. It depends on US2 (tutorial walkthrough) having been experienced by the user.

**Independent Test**: Can be fully tested by completing a tutorial walkthrough and verifying that Gemara User Journey provides a handoff summary directing the user to OpenCode, naming the specific gemara-mcp prompts/tools/resources to use, the artifact type to create, and the key decisions the user should have pre-considered from the tutorial.

**Acceptance Scenarios**:

1. **Given** a user has completed a tutorial walkthrough for a specific artifact type, **When** they indicate readiness to begin authoring, **Then** Gemara User Journey presents a handoff summary directing them to open an OpenCode session, listing the specific gemara-mcp prompt to use (e.g., `threat_assessment` for Threat Catalogs), the schema definition for validation (e.g., `#ThreatCatalog`), available MCP resources (lexicon, schema docs), and key decisions the user should have answers for based on the tutorial.
2. **Given** a user is at the handoff point, **When** the gemara-mcp server is configured in `opencode.json`, **Then** Gemara User Journey confirms the configuration is present and instructs the user to launch `opencode` to begin authoring with full MCP tool and resource access.
3. **Given** a user is at the handoff point, **When** the gemara-mcp server is not configured, **Then** Gemara User Journey instructs the user to run `./journey --doctor` to verify their environment, explains how to configure the MCP server in `opencode.json`, and provides the manual `cue vet` validation command as an alternative until the server is set up.

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

### User Story 6 - GitHub README Documentation (Priority: P6)

The project's GitHub README.md is rewritten to serve as a concise, visually polished landing page for new users. It summarizes the project's goal, includes an image of the web UI, provides direct installation links for all dependencies, and narrates the user journey from role discovery through tutorial walkthrough to MCP-assisted authoring. The README enables a user to understand what Gemara User Journey does, install prerequisites, and begin the guided experience within minutes.

**Why this priority**: The README is the first touchpoint for users discovering the project on GitHub. Without a clear, concise README that shows the UI and explains the user journey end-to-end, users cannot self-serve onboarding. It depends on the core flows (US1-US4) being defined so the README accurately describes them.

**Independent Test**: Can be fully tested by having a new user read only the README, follow its dependency installation links, and confirm they can build and launch Gemara User Journey without consulting other documentation.

**Acceptance Scenarios**:

1. **Given** a user visits the GitHub repository, **When** they view the README, **Then** they see a concise project summary explaining that Gemara User Journey is a role-based tutorial guide for the Gemara GRC schema project, distinguishing it from the MCP server (guide vs. authoring tool).
2. **Given** a user reads the README, **When** they look for installation instructions, **Then** the README provides direct hyperlinks to installation pages for each dependency (Go, CUE, OpenCode, Git) and instructions for building the Gemara MCP server from source.
3. **Given** a user reads the README, **When** they look for a visual preview of the project, **Then** the README includes a manually captured screenshot of the web UI (stored in `docs/images/`) showing the role discovery or tutorial suggestion interface.
4. **Given** a user reads the README, **When** they want to understand the user journey, **Then** the README describes the end-to-end flow: role and activity discovery in Gemara User Journey, tailored tutorial walkthrough, and handoff to OpenCode with the gemara-mcp server for artifact authoring using the Gemara schemas and MCP server tools.
5. **Given** a user reads the README, **When** they want to start using the project, **Then** the README provides a Getting Started section with no more than 4 steps to go from clone to first tutorial interaction.
6. **Given** a user wants detailed information (project structure, contributing guidelines, MCP update procedures, layer reference), **When** they look for it in the README, **Then** the README links to dedicated files in `docs/` rather than inlining the content.

---

### Edge Cases

- What happens when no tutorials exist for the user's identified layers? Gemara User Journey should inform the user which layers lack tutorials and suggest checking back as content is added, while still presenting any available tutorials.
- What happens when a user's described activities produce no keyword matches? Gemara User Journey should fall back to the user's role defaults and explain that no specific activity keywords were detected, offering the category-based selection as an alternative.
- What happens when the latest release has experimental or draft schemas? Gemara User Journey should proceed with the latest release but inform the user which schemas are experimental, noting that authoring for those schemas may produce artifacts that require updates when schemas stabilize.
- What happens when the user wants to author without completing a tutorial? Gemara User Journey should allow it but recommend completing the tutorial first, providing a brief explanation of what they may miss. If the user proceeds, direct them to OpenCode with the gemara-mcp server and the relevant MCP prompt name.
- What happens when the MCP server's built-in schema version differs from the auto-selected latest? Gemara User Journey should note the discrepancy in the handoff summary and recommend the user validate artifacts after authoring.

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
- **FR-010**: System MUST NOT replicate the MCP server's authoring wizards (threat_assessment, control_catalog prompts) or artifact validation within Gemara User Journey's guided flow.
- **FR-011**: System MUST provide a clear handoff point after tutorial completion that directs the user to open an OpenCode session with the gemara-mcp server, including the specific prompt name, schema definition, available MCP resources (lexicon, schema docs), and a preparation checklist.
- **FR-017**: The `--doctor` command MUST remain fully functional and unchanged. It continues to verify the user's environment (Go version, CUE installation, gemara-mcp server availability, opencode.json configuration) and report actionable status for each check.
- **FR-018**: All terminal output MUST be user-friendly and visually polished for all audiences, including non-technical stakeholders, security engineers, compliance officers, and developers. Output MUST use consistent styling (colors, spacing, icons, card layouts), avoid jargon without context, and present information in a scannable format with clear visual hierarchy.
- **FR-019**: The post-tutorial handoff MUST explicitly direct users to OpenCode as the authoring environment, referencing the gemara-mcp server's available tools (`validate_gemara_artifact`), resources (`gemara://lexicon`, `gemara://schema/definitions`), and wizard prompts (`threat_assessment`, `control_catalog`) by name so users know exactly what capabilities are available to them.
- **FR-012**: System MUST inform the user when no tutorials are available for their identified layers and suggest alternative actions.
- **FR-013**: System MUST handle ambiguous activity keywords by prompting the user with a targeted clarification question to resolve the correct layer mapping.
- **FR-014**: System MUST retain the existing version selection and switching code in the codebase, bypassed but not deleted, with documentation explaining the intentional deferral.
- **FR-015**: System MUST automatically update to the latest release when a newer version is available on subsequent launches, informing the user of the update.
- **FR-016**: System MUST warn the user when the MCP server's built-in schema version differs from the auto-selected latest release, noting the discrepancy in the handoff summary.
- **FR-020**: The GitHub README.md MUST include a concise project summary that explains Gemara User Journey's purpose as a role-based tutorial guide for Gemara and distinguishes it from the MCP server's authoring capabilities.
- **FR-021**: The GitHub README.md MUST provide direct hyperlinks to installation pages for all dependencies (Go, CUE, OpenCode, Git) and instructions for building the Gemara MCP server from source.
- **FR-022**: The GitHub README.md MUST include a manually captured screenshot of the web UI (role discovery or tutorial suggestion view), stored in `docs/images/` within the repository and referenced via a relative path.
- **FR-023**: The GitHub README.md MUST describe the end-to-end user journey: role and activity discovery, tailored tutorial walkthrough, and handoff to OpenCode with the gemara-mcp server for artifact authoring using the Gemara schemas and MCP server tools.
- **FR-024**: The GitHub README.md MUST be concise, targeting a focused landing page experience, with a clear Getting Started section of no more than 4 steps. Detailed content (project structure, contributing guidelines, MCP update procedures, layer reference) MUST be moved to separate files in `docs/` and linked from the README.

### Key Entities

- **Activity Profile**: Represents a user's resolved set of Gemara layers, extracted keywords, matched categories, and role context. Determines which tutorials and artifact types are recommended.
- **Learning Path**: An ordered sequence of tutorial steps tailored to the user's activity profile. Each step includes a tutorial reference, layer association, relevance annotations (Why/How/What), section relevance scores, and completion status.
- **Handoff Summary**: A structured transition point presented after tutorial completion. Directs the user to OpenCode with the gemara-mcp server, containing the target artifact type, the specific MCP prompt or tool to use, available MCP resources (lexicon, schema definitions), the schema definition for validation, and a list of key decisions the user should have pre-considered.
- **Release Resolution**: The process of automatically determining the latest Gemara release. Includes the resolved version tag, whether it was fetched live or from cache, and any experimental schema warnings.

## Assumptions

- OpenCode is the preferred AI development harness and the destination for all authoring work after Gemara User Journey tutorials. Users complete tutorials in the Gemara User Journey terminal, then switch to OpenCode where the gemara-mcp server provides tools, resources, and wizard prompts for artifact creation.
- The Gemara MCP server will continue to be the primary authoring tool and will maintain its current wizard prompts (`threat_assessment`, `control_catalog`), resources (`gemara://lexicon`, `gemara://schema/definitions`), and validation capabilities (`validate_gemara_artifact`). Gemara User Journey does not need to provide these.
- The upstream Gemara repository at `gemaraproj/gemara` will continue to host tutorial content in `docs/tutorials/` and publish releases via GitHub's releases API.
- "Latest release" means the most recent release by date from the Gemara repository, consistent with the existing `DetermineLatestVersion()` behavior. This includes prereleases if no stable release exists.
- The `--doctor` command is an existing, independent diagnostic tool that verifies the user's environment. It is not modified by this feature and continues to work as-is.
- The existing version selection code (`internal/schema/selector.go`, `internal/cli/version_prompt.go`) is mature enough to be preserved for future re-enablement without requiring reimplementation.
- The MCP server version (gemara-mcp binary release) is a separate concern from the Gemara schema version. This specification addresses schema version selection only; MCP server installation continues to use the latest release as it does today.
- Gemara User Journey's terminal output is consumed by all audiences (security engineers, compliance officers, CISOs, developers, policy authors, auditors). Output must be accessible and polished regardless of the user's technical depth.

## Clarifications

### Session 2026-03-25

- Q: Should README documentation be tracked as a new user story in this spec or handled separately? → A: Add a new User Story (US6) to this spec for README documentation with acceptance criteria.
- Q: How should the web UI screenshot be sourced and stored? → A: Manually capture a screenshot and commit to `docs/images/` in the repo.
- Q: How should the README be restructured for conciseness? → A: Reduce README to a focused landing page; move detailed content (project structure, contributing, MCP update procedures, layer reference) to separate files in `docs/` with links from the README.
- Q: Should the README reference a Gemara "SDK" alongside schemas and MCP tools? → A: No. The README should reference only Gemara schemas and MCP server tools. Do not include SDK references.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can identify their activities, see their recommended tutorials, and understand their expected artifact outputs within 5 minutes of launching Gemara User Journey.
- **SC-002**: 90% of users who complete a tutorial walkthrough can correctly name the MCP server prompt or tool they need to use for authoring without additional guidance.
- **SC-003**: The setup flow completes without presenting any version selection prompts, reducing the number of user decision points during setup by at least one compared to the current flow.
- **SC-004**: Users who complete the tutorial-to-authoring handoff report that they felt prepared to begin authoring, with at least 80% indicating they understood the key decisions required.
- **SC-005**: No Gemara User Journey flow duplicates functionality available in the MCP server's wizard prompts or validation tools, verified by review of Gemara User Journey's guided flow output against MCP server capabilities.
- **SC-006**: The version selection code remains functional in the codebase and can be re-enabled by a developer within 1 hour of effort, verified by the presence of bypass documentation and intact code.
- **SC-007**: All terminal output uses consistent visual styling (card layouts, color-coded labels, clear spacing) and is readable by users who are not software developers, verified by review of output screenshots across all flows (activity identification, tutorial navigation, handoff summary).
- **SC-008**: The post-tutorial handoff summary explicitly names OpenCode as the authoring environment and lists at least the relevant gemara-mcp prompt, available resources, and validation tool, verified by inspecting the rendered handoff output for each artifact type.
- **SC-009**: The `--doctor` command continues to function correctly after all changes, verified by running `./journey --doctor` and confirming all environment checks pass.
- **SC-010**: A new user reading only the GitHub README can identify the project's purpose, install all dependencies using the provided links, and reach a first tutorial interaction within 10 minutes, verified by walkthrough testing.
- **SC-011**: The README includes a visible web UI screenshot and renders correctly on GitHub with no broken images or links, verified by visual inspection of the rendered GitHub page.
