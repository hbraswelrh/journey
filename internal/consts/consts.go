// SPDX-License-Identifier: Apache-2.0

// Package consts defines centralized constants for the Gemara User Journey
// project. All magic strings, URLs, tool names, and
// configuration defaults are defined here and referenced by
// name throughout the codebase.
package consts

// GemaraRepoURL is the upstream Gemara schema repository.
const GemaraRepoURL = "https://github.com/gemaraproj/gemara"

// JourneyRepoURL is the Gemara User Journey tutorial engine repository.
const JourneyRepoURL = "https://github.com/hbraswelrh/journey"

// PacmanDiscussionsURL is the GitHub Discussions URL for the
// Gemara User Journey repository, where users share their Gemara journey.
const PacmanDiscussionsURL = JourneyRepoURL + "/discussions"

// PacmanNewDiscussionURL is the URL to create a new discussion
// using the Gemara Journey template.
const PacmanNewDiscussionURL = PacmanDiscussionsURL +
	"/new?category=general"

// GemaraMCPRepoURL is the Gemara MCP server repository (HTTPS).
const GemaraMCPRepoURL = "https://github.com/gemaraproj/gemara-mcp"

// GemaraMCPCloneSSH is the SSH clone URL for the Gemara MCP
// server repository.
const GemaraMCPCloneSSH = "git@github.com:gemaraproj/gemara-mcp.git"

// GemaraMCPCloneHTTPS is the HTTPS clone URL for the Gemara MCP
// server repository.
const GemaraMCPCloneHTTPS = "https://github.com/gemaraproj/gemara-mcp.git"

// MCPBinaryName is the expected binary name for the Gemara MCP
// server when built from source.
const MCPBinaryName = "gemara-mcp"

// MCPPodmanContainer is the expected Podman container name for
// the Gemara MCP server.
const MCPPodmanContainer = "gemara-mcp"

// MCPPodmanImage is the container image reference for the
// Gemara MCP server.
const MCPPodmanImage = "ghcr.io/gemaraproj/gemara-mcp:latest"

// MCPInstallDir is the subdirectory under ~/.local/share
// where the MCP server is installed from source.
const MCPInstallDir = "journey"

// InstalledReleaseFile is the filename for the installed
// release metadata, stored alongside the built binary.
const InstalledReleaseFile = "installed-release.json"

// OpenCodeConfigFile is the OpenCode configuration file where
// MCP server entries are registered.
const OpenCodeConfigFile = "opencode.json"

// MCPServerName is the name used for the gemara-mcp entry in
// the OpenCode MCP configuration.
const MCPServerName = "gemara-mcp"

// WizardThreatAssessment is the MCP prompt name for the
// threat assessment wizard.
const WizardThreatAssessment = "threat_assessment"

// WizardControlCatalog is the MCP prompt name for the
// control catalog wizard.
const WizardControlCatalog = "control_catalog"

// MCP server operating modes. Advisory is read-only analysis
// and validation; artifact adds guided creation wizards.
const (
	MCPModeAdvisory = "advisory"
	MCPModeArtifact = "artifact"
)

// MCPModeDefault is the default server mode when none is
// specified.
const MCPModeDefault = MCPModeArtifact

// MCPModeFlag is the command-line flag name for selecting
// the server mode.
const MCPModeFlag = "--mode"

// MCP tool names as defined by the Gemara MCP server.
// Only validate_gemara_artifact is a callable tool;
// lexicon and schema docs are accessed as MCP resources.
const (
	ToolValidateArtifact = "validate_gemara_artifact"
)

// MCP resource URIs as defined by the Gemara MCP server.
const (
	ResourceLexicon           = "gemara://lexicon"
	ResourceSchemaDefinitions = "gemara://schema/definitions"
)

// GemaraCloneHTTPS is the HTTPS clone URL for the upstream
// Gemara repository.
const GemaraCloneHTTPS = "https://github.com/gemaraproj/gemara.git"

// GemaraCloneSSH is the SSH clone URL for the upstream
// Gemara repository.
const GemaraCloneSSH = "git@github.com:gemaraproj/gemara.git"

// GemaraTutorialsSubdir is the subdirectory within the
// Gemara repository that contains the tutorials.
const GemaraTutorialsSubdir = "docs/tutorials"

// DefaultGemaraDir is the default local directory where
// the Gemara repository is cloned for tutorial access.
const DefaultGemaraDir = ".local/share/journey/gemara"

// DefaultTutorialsDir is the default path to the Gemara
// tutorials directory, resolved at runtime from the home
// directory and DefaultGemaraDir.
const DefaultTutorialsDir = "~/github/openssf/gemara/gemara/docs/tutorials"

// DefaultOutputFormat is the default output format for
// structured data.
const DefaultOutputFormat = "yaml"

// Gemara schema definition names used for cue vet validation.
const (
	SchemaGuidanceCatalog   = "#GuidanceCatalog"
	SchemaVectorCatalog     = "#VectorCatalog"
	SchemaPrincipleCatalog  = "#PrincipleCatalog"
	SchemaControlCatalog    = "#ControlCatalog"
	SchemaThreatCatalog     = "#ThreatCatalog"
	SchemaCapabilityCatalog = "#CapabilityCatalog"
	SchemaPolicy            = "#Policy"
	SchemaRiskCatalog       = "#RiskCatalog"
	SchemaMappingDocument   = "#MappingDocument"
	SchemaEvaluationLog     = "#EvaluationLog"
	SchemaEnforcementLog    = "#EnforcementLog"
	SchemaAuditLog          = "#AuditLog"
)

// Gemara artifact type identifiers.
const (
	ArtifactGuidanceCatalog   = "GuidanceCatalog"
	ArtifactVectorCatalog     = "VectorCatalog"
	ArtifactPrincipleCatalog  = "PrincipleCatalog"
	ArtifactControlCatalog    = "ControlCatalog"
	ArtifactThreatCatalog     = "ThreatCatalog"
	ArtifactCapabilityCatalog = "CapabilityCatalog"
	ArtifactPolicy            = "Policy"
	ArtifactRiskCatalog       = "RiskCatalog"
	ArtifactMappingDocument   = "MappingDocument"
	ArtifactEvaluationLog     = "EvaluationLog"
	ArtifactEnforcementLog    = "EnforcementLog"
	ArtifactAuditLog          = "AuditLog"
)

// Gemara relationship type strings for MappingDocument entries.
const (
	RelImplements = "implements"
	RelEquivalent = "equivalent"
	RelSubsumes   = "subsumes"
)

// Homebrew installation commands for required and recommended
// tools. Homebrew is the preferred installation method on macOS
// and Linux.
const (
	BrewInstallCUE      = "brew install cue-lang/tap/cue"
	BrewInstallGitleaks = "brew install gitleaks"
	BrewInstallOpenCode = "brew install anomalyco/tap/opencode"
	BrewInstallPodman   = "brew install podman"
)

// GemaraReleasesAPI is the GitHub API endpoint for listing
// releases of the Gemara schema repository.
const GemaraReleasesAPI = "https://api.github.com/repos/gemaraproj/gemara/releases"

// SchemaStatusStable is the CUE status attribute value for
// stable schemas.
const SchemaStatusStable = "Stable"

// SchemaStatusExperimental is the CUE status attribute value
// for experimental schemas.
const SchemaStatusExperimental = "Experimental"

// CoreStableSchemas is the list of schema names that must be
// marked Stable for a release to qualify as the "Stable"
// version.
var CoreStableSchemas = []string{
	"base",
	"metadata",
	"mapping_inline",
}

// CacheDir is the subdirectory name under the user's config
// directory where Gemara User Journey stores cached data (lexicon, schema
// docs, version info).
const CacheDir = "journey"

// ReleaseCacheFile is the filename for cached release data.
const ReleaseCacheFile = "releases.json"

// SessionHealthCheckInterval is the interval in seconds between
// MCP server health checks during an active session.
const SessionHealthCheckInterval = 30

// Gemara layer numbers. The seven-layer model is Gemara's core
// organizing framework.
const (
	LayerVectorsGuidance   = 1
	LayerThreatsControls   = 2
	LayerRiskPolicy        = 3
	LayerSensitiveActivity = 4
	LayerEvaluation        = 5
	LayerEnforcement       = 6
	LayerAudit             = 7
)

// LayerGuidance is an alias for LayerVectorsGuidance,
// preserved for backward compatibility.
const LayerGuidance = LayerVectorsGuidance

// LayerDataCollection is an alias for LayerEnforcement,
// preserved for backward compatibility.
const LayerDataCollection = LayerEnforcement

// LayerReporting is an alias for LayerAudit,
// preserved for backward compatibility.
const LayerReporting = LayerAudit

// Predefined role names per FR-002. This list MUST be updated
// as research identifies new personas.
const (
	RoleSecurityEngineer  = "Security Engineer"
	RoleComplianceOfficer = "Compliance Officer"
	RoleCISO              = "CISO/Security Leader"
	RoleDeveloper         = "Developer"
	RolePlatformEngineer  = "Platform Engineer"
	RolePolicyAuthor      = "Policy Author"
	RoleAuditor           = "Auditor"
	RoleCustom            = "My role isn't listed"
)

// RoleProfileDir is the subdirectory under the user's config
// directory where custom role profiles are stored.
const RoleProfileDir = "journey/roles"

// RoleProfileExt is the file extension for saved role profiles.
const RoleProfileExt = ".yaml"

// BlockCacheDir is the subdirectory under the user's config
// directory where extracted content blocks are stored.
const BlockCacheDir = "journey/blocks"

// BlockManifestFile is the filename for the extraction
// manifest used by drift detection.
const BlockManifestFile = "manifest.yaml"

// Content block categories per FR-005.
const (
	CategoryPattern        = "pattern"
	CategoryValidationStep = "validation_step"
	CategoryNamingConv     = "naming_convention"
	CategorySchemaStruct   = "schema_structure"
	CategoryCrossRef       = "cross_reference"
)

// TeamConfigDir is the subdirectory under the user's config
// directory where team configurations are stored.
const TeamConfigDir = "journey/teams"

// TeamConfigExt is the file extension for saved team configs.
const TeamConfigExt = ".yaml"

// LayerArtifacts maps Gemara layer numbers to the artifact
// types primarily produced at that layer. MappingDocument is
// cross-layer (L1-L3) and is listed under each layer it
// spans.
var LayerArtifacts = map[int][]string{
	LayerVectorsGuidance: {
		ArtifactGuidanceCatalog,
		ArtifactVectorCatalog,
		ArtifactPrincipleCatalog,
	},
	LayerThreatsControls: {
		ArtifactThreatCatalog,
		ArtifactControlCatalog,
		ArtifactCapabilityCatalog,
	},
	LayerRiskPolicy: {
		ArtifactPolicy,
		ArtifactRiskCatalog,
	},
	LayerSensitiveActivity: {},
	LayerEvaluation:        {ArtifactEvaluationLog},
	LayerEnforcement:       {ArtifactEnforcementLog},
	LayerAudit:             {ArtifactAuditLog},
}

// ArtifactFlowDescriptions describes how artifacts flow
// between adjacent Gemara layers.
var ArtifactFlowDescriptions = map[[2]int]string{
	{LayerVectorsGuidance, LayerThreatsControls}: "" +
		"Guidance and vector catalogs inform threat " +
		"and control scope",
	{LayerVectorsGuidance, LayerRiskPolicy}: "Guidance " +
		"catalogs are referenced by policy documents",
	{LayerThreatsControls, LayerRiskPolicy}: "Control " +
		"and threat catalogs feed policy and risk " +
		"catalog evaluation criteria",
	{LayerThreatsControls, LayerSensitiveActivity}: "" +
		"Controls define requirements for sensitive " +
		"activities",
	{LayerRiskPolicy, LayerSensitiveActivity}: "Policy " +
		"governs which controls apply to sensitive " +
		"activities",
	{LayerRiskPolicy, LayerEvaluation}: "Policy drives " +
		"evaluation log assessments",
	{LayerEvaluation, LayerEnforcement}: "Evaluation " +
		"findings drive enforcement actions",
	{LayerEnforcement, LayerAudit}: "Enforcement logs " +
		"inform audit and continuous monitoring",
}

// AuthoringOutputDir is the default subdirectory for
// authored artifact output.
const AuthoringOutputDir = "artifacts"

// DefaultArtifactFormat is the default output format for
// authored artifacts.
const DefaultArtifactFormat = "yaml"

// Artifact section name constants for guided authoring.
const (
	SectionMetadata       = "metadata"
	SectionScope          = "scope"
	SectionCapabilities   = "capabilities"
	SectionThreats        = "threats"
	SectionControls       = "controls"
	SectionGuidanceItems  = "guidance_items"
	SectionVectors        = "vectors"
	SectionPrinciples     = "principles"
	SectionRisks          = "risks"
	SectionPolicyCriteria = "policy_criteria"
	SectionMappings       = "mappings"
	SectionEvaluations    = "evaluations"
	SectionActions        = "actions"
	SectionResults        = "results"
)

// Validation status values for authored artifacts.
const (
	ValidationStatusNotValidated = "not_validated"
	ValidationStatusPartial      = "partial"
	ValidationStatusValid        = "valid"
	ValidationStatusInvalid      = "invalid"
)

// ArtifactTypeSections maps each artifact type to its
// ordered list of authoring sections.
var ArtifactTypeSections = map[string][]string{
	ArtifactGuidanceCatalog: {
		SectionMetadata,
		SectionScope,
		SectionGuidanceItems,
	},
	ArtifactVectorCatalog: {
		SectionMetadata,
		SectionScope,
		SectionVectors,
	},
	ArtifactPrincipleCatalog: {
		SectionMetadata,
		SectionScope,
		SectionPrinciples,
	},
	ArtifactControlCatalog: {
		SectionMetadata,
		SectionScope,
		SectionControls,
	},
	ArtifactThreatCatalog: {
		SectionMetadata,
		SectionScope,
		SectionCapabilities,
		SectionThreats,
	},
	ArtifactCapabilityCatalog: {
		SectionMetadata,
		SectionScope,
		SectionCapabilities,
	},
	ArtifactPolicy: {
		SectionMetadata,
		SectionScope,
		SectionPolicyCriteria,
	},
	ArtifactRiskCatalog: {
		SectionMetadata,
		SectionScope,
		SectionRisks,
	},
	ArtifactMappingDocument: {
		SectionMetadata,
		SectionMappings,
	},
	ArtifactEvaluationLog: {
		SectionMetadata,
		SectionScope,
		SectionEvaluations,
	},
	ArtifactEnforcementLog: {
		SectionMetadata,
		SectionScope,
		SectionActions,
	},
	ArtifactAuditLog: {
		SectionMetadata,
		SectionScope,
		SectionResults,
	},
}

// Authoring approach constants describe how a user creates
// an artifact after completing a tutorial. Wizard means an
// MCP wizard prompt guides them step-by-step; collaborative
// means they author with MCP resources (lexicon, schema
// docs) and validation support.
const (
	ApproachWizard        = "wizard"
	ApproachCollaborative = "collaborative"
)

// ArtifactDescriptions maps each artifact type to a
// one-sentence user-facing description suitable for
// display to all audiences including non-technical
// stakeholders.
var ArtifactDescriptions = map[string]string{
	ArtifactGuidanceCatalog: "A structured catalog of " +
		"standards, best practices, and regulatory " +
		"requirements that your organization follows.",
	ArtifactVectorCatalog: "A catalog of attack vectors " +
		"and techniques that document known methods of " +
		"compromise relevant to your domain.",
	ArtifactPrincipleCatalog: "A catalog of foundational " +
		"values and principles that guide governance, " +
		"design, and operational decisions.",
	ArtifactControlCatalog: "A catalog of security " +
		"controls that mitigate identified threats, " +
		"with assessment requirements and evidence " +
		"criteria.",
	ArtifactThreatCatalog: "A catalog of threats to a " +
		"specific component, organized by capability, " +
		"with severity and likelihood assessments.",
	ArtifactCapabilityCatalog: "A catalog of system " +
		"capabilities and features that can be " +
		"leveraged to implement security controls.",
	ArtifactPolicy: "An organizational policy document " +
		"defining adherence requirements, timelines, " +
		"and scope for a set of controls.",
	ArtifactRiskCatalog: "A structured collection of " +
		"documented risks with severity levels, risk " +
		"appetite definitions, and threat mappings.",
	ArtifactMappingDocument: "A cross-reference document " +
		"that maps controls to guidance items, " +
		"establishing traceability between layers.",
	ArtifactEvaluationLog: "An assessment log recording " +
		"control evaluations, evidence collected, and " +
		"compliance findings.",
	ArtifactEnforcementLog: "A log of enforcement " +
		"actions taken in response to noncompliance " +
		"findings from evaluations.",
	ArtifactAuditLog: "A formal audit record " +
		"documenting review results, evidence, and " +
		"recommendations for organizational " +
		"compliance posture.",
}

// ArtifactWizards maps artifact types that have MCP wizard
// prompts to their prompt names. Artifact types not in this
// map use collaborative authoring with MCP resources.
var ArtifactWizards = map[string]string{
	ArtifactThreatCatalog:  WizardThreatAssessment,
	ArtifactControlCatalog: WizardControlCatalog,
}

// DefaultPreparationChecklists maps each artifact type to
// a list of items the user should have ready before
// beginning authoring in OpenCode with the gemara-mcp
// server.
var DefaultPreparationChecklists = map[string][]string{
	ArtifactGuidanceCatalog: {
		"Identify the standard, regulation, or best " +
			"practice to codify",
		"Determine scope and applicability",
		"Gather source material (regulatory text, " +
			"standard sections)",
	},
	ArtifactVectorCatalog: {
		"Identify the domain or technology to document " +
			"attack vectors for",
		"Determine scope and applicability contexts",
		"Gather known attack methods and exploitation " +
			"pathways (e.g., MITRE ATT&CK techniques)",
	},
	ArtifactPrincipleCatalog: {
		"Identify the governance or design area for " +
			"the principles",
		"Determine scope and applicability",
		"Gather foundational values and rationale from " +
			"organizational standards",
	},
	ArtifactThreatCatalog: {
		"Identify the component or system to assess",
		"Determine scope boundaries (what is in/out)",
		"Decide whether to import from an existing " +
			"catalog (e.g., FINOS CCC Core)",
		"Consider MITRE ATT&CK alignment preference",
	},
	ArtifactControlCatalog: {
		"Identify the component or system to protect",
		"Select the guideline framework(s) to align with",
		"Determine scope boundaries",
		"Decide whether to import from an existing catalog",
	},
	ArtifactCapabilityCatalog: {
		"Identify the system or technology to document " +
			"capabilities for",
		"Determine capability groupings and categories",
		"Gather feature and component descriptions",
	},
	ArtifactPolicy: {
		"Identify the controls this policy governs",
		"Define the adherence timeline",
		"Determine compliance scope (teams, systems, " +
			"regions)",
		"Establish non-compliance handling procedures",
	},
	ArtifactRiskCatalog: {
		"Identify organizational risk categories",
		"Determine risk appetite levels per category",
		"Map risks to known threats from Layer 2 " +
			"threat catalogs",
		"Define severity boundaries and RACI ownership",
	},
	ArtifactMappingDocument: {
		"Identify source and target catalogs to map",
		"Determine relationship types (implements, " +
			"equivalent, subsumes)",
		"Gather entry references for both catalogs",
	},
	ArtifactEvaluationLog: {
		"Identify the controls to evaluate",
		"Gather evidence and assessment materials",
		"Determine evaluation criteria and scoring",
	},
	ArtifactEnforcementLog: {
		"Identify the evaluation findings to respond to",
		"Determine enforcement disposition " +
			"(enforced, tolerated, clear)",
		"Document the enforcement method and steps",
		"Gather assessment findings for justification",
	},
	ArtifactAuditLog: {
		"Define the audit scope and criteria",
		"Identify the policies and controls to audit " +
			"against",
		"Gather evidence from evaluation and " +
			"enforcement logs",
		"Assign RACI ownership for the audit",
	},
}

// GemaraTutorialsBaseURL is the base URL for upstream Gemara
// tutorials on the official documentation site.
const GemaraTutorialsBaseURL = "https://gemara.openssf.org/tutorials"

// UpstreamTutorialID constants identify each upstream
// Gemara tutorial for programmatic reference.
const (
	TutorialThreatAssessment = "threat-assessment-guide"
	TutorialControlCatalog   = "control-catalog-guide"
	TutorialGuidanceCatalog  = "guidance-guide"
	TutorialPolicy           = "policy-guide"
)

// UpstreamTutorial describes a tutorial published on the
// upstream Gemara documentation site.
type UpstreamTutorial struct {
	// ID is the unique tutorial identifier.
	ID string
	// Title is the human-readable tutorial title.
	Title string
	// Description explains what the user will learn.
	Description string
	// URL is the full URL to the tutorial page.
	URL string
	// Layer is the primary Gemara layer this tutorial
	// covers.
	Layer int
	// ArtifactTypes lists the artifact types produced
	// by completing this tutorial.
	ArtifactTypes []string
	// Prerequisites lists tutorial IDs that should be
	// completed before this one.
	Prerequisites []string
	// Goals describes user goals that map to this
	// tutorial, sourced from the upstream "Find Your
	// Tutorial" section.
	Goals []string
	// Roles lists the role names that benefit most from
	// this tutorial.
	Roles []string
}

// UpstreamTutorials is the authoritative list of tutorials
// published at gemara.openssf.org/tutorials/, ordered by
// recommended learning sequence.
var UpstreamTutorials = []UpstreamTutorial{
	{
		ID:    TutorialGuidanceCatalog,
		Title: "Guidance Catalog Guide",
		Description: "Create a structured set of " +
			"guidelines — recommendations, " +
			"requirements, or best practices — " +
			"grouped by theme with mapping " +
			"references to external standards.",
		URL:   GemaraTutorialsBaseURL + "/guidance/guidance-guide",
		Layer: LayerVectorsGuidance,
		ArtifactTypes: []string{
			ArtifactGuidanceCatalog,
		},
		Prerequisites: []string{},
		Goals: []string{
			"Creating a guidance catalog from " +
				"best practices",
			"Codifying standards, regulations, " +
				"or best practices into a " +
				"machine-readable format",
			"Understanding what guidance catalogs " +
				"are and how to structure them",
		},
		Roles: []string{
			RoleComplianceOfficer,
			RolePolicyAuthor,
			RoleCISO,
			RoleSecurityEngineer,
		},
	},
	{
		ID:    TutorialThreatAssessment,
		Title: "Threat Assessment Guide",
		Description: "Walk through a threat assessment " +
			"for a system or component — identify " +
			"capabilities, map threats to attack " +
			"surfaces, and import from external " +
			"catalogs like FINOS CCC Core.",
		URL:   GemaraTutorialsBaseURL + "/controls/threat-assessment-guide",
		Layer: LayerThreatsControls,
		ArtifactTypes: []string{
			ArtifactThreatCatalog,
		},
		Prerequisites: []string{},
		Goals: []string{
			"Performing a threat assessment for a " +
				"system or component",
			"Understanding what threats and " +
				"controls exist before writing " +
				"policy",
			"Understanding the security posture " +
				"of consumed software",
		},
		Roles: []string{
			RoleSecurityEngineer,
			RoleDeveloper,
			RolePlatformEngineer,
			RoleComplianceOfficer,
		},
	},
	{
		ID:    TutorialControlCatalog,
		Title: "Control Catalog Guide",
		Description: "Create a control catalog that " +
			"maps security controls to identified " +
			"threats — define objectives, " +
			"assessment requirements, and link " +
			"controls to threat catalogs.",
		URL:   GemaraTutorialsBaseURL + "/controls/control-catalog-guide",
		Layer: LayerThreatsControls,
		ArtifactTypes: []string{
			ArtifactControlCatalog,
		},
		Prerequisites: []string{
			TutorialThreatAssessment,
		},
		Goals: []string{
			"Defining security controls that " +
				"mitigate identified threats",
			"Reviewing the controls to reference " +
				"in a policy",
		},
		Roles: []string{
			RoleSecurityEngineer,
			RoleDeveloper,
			RolePlatformEngineer,
			RolePolicyAuthor,
		},
	},
	{
		ID:    TutorialPolicy,
		Title: "Organizational Risk & Policy Guide",
		Description: "Create a policy document that " +
			"translates risk appetite into mandatory " +
			"rules — define scope, imports, " +
			"adherence requirements, RACI contacts, " +
			"and implementation timelines.",
		URL:   GemaraTutorialsBaseURL + "/policy/policy-guide",
		Layer: LayerRiskPolicy,
		ArtifactTypes: []string{
			ArtifactPolicy,
		},
		Prerequisites: []string{
			TutorialThreatAssessment,
			TutorialControlCatalog,
		},
		Goals: []string{
			"Creating organizational policy",
			"Translating risk appetite into " +
				"mandatory rules",
			"Defining adherence timelines and " +
				"enforcement methods",
		},
		Roles: []string{
			RolePolicyAuthor,
			RoleCISO,
			RoleComplianceOfficer,
			RoleAuditor,
		},
	},
}
