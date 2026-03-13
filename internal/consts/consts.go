// SPDX-License-Identifier: Apache-2.0

// Package consts defines centralized constants for the Pac-Man
// project. All magic strings, URLs, tool names, and
// configuration defaults are defined here and referenced by
// name throughout the codebase.
package consts

// GemaraRepoURL is the upstream Gemara schema repository.
const GemaraRepoURL = "https://github.com/gemaraproj/gemara"

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
const MCPInstallDir = "pacman"

// InstalledReleaseFile is the filename for the installed
// release metadata, stored alongside the built binary.
const InstalledReleaseFile = "installed-release.json"

// OpenCodeConfigFile is the OpenCode configuration file where
// MCP server entries are registered.
const OpenCodeConfigFile = "opencode.json"

// MCPServerName is the name used for the gemara-mcp entry in
// the OpenCode MCP configuration.
const MCPServerName = "gemara-mcp"

// MCP tool names as defined by the Gemara MCP server.
const (
	ToolGetLexicon       = "get_lexicon"
	ToolValidateArtifact = "validate_gemara_artifact"
	ToolGetSchemaDocs    = "get_schema_docs"
)

// DefaultTutorialsDir is the default path to the Gemara
// tutorials directory.
const DefaultTutorialsDir = "~/github/openssf/gemara/gemara/docs/tutorials"

// DefaultOutputFormat is the default output format for
// structured data.
const DefaultOutputFormat = "yaml"

// Gemara schema definition names used for cue vet validation.
const (
	SchemaGuidanceCatalog = "#GuidanceCatalog"
	SchemaControlCatalog  = "#ControlCatalog"
	SchemaThreatCatalog   = "#ThreatCatalog"
	SchemaPolicy          = "#Policy"
	SchemaMappingDocument = "#MappingDocument"
	SchemaEvaluationLog   = "#EvaluationLog"
)

// Gemara artifact type identifiers.
const (
	ArtifactGuidanceCatalog = "GuidanceCatalog"
	ArtifactControlCatalog  = "ControlCatalog"
	ArtifactThreatCatalog   = "ThreatCatalog"
	ArtifactPolicy          = "Policy"
	ArtifactMappingDocument = "MappingDocument"
	ArtifactEvaluationLog   = "EvaluationLog"
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
// directory where Pac-Man stores cached data (lexicon, schema
// docs, version info).
const CacheDir = "pacman"

// ReleaseCacheFile is the filename for cached release data.
const ReleaseCacheFile = "releases.json"

// SessionHealthCheckInterval is the interval in seconds between
// MCP server health checks during an active session.
const SessionHealthCheckInterval = 30

// Gemara layer numbers. The seven-layer model is Gemara's core
// organizing framework.
const (
	LayerGuidance          = 1
	LayerThreatsControls   = 2
	LayerRiskPolicy        = 3
	LayerSensitiveActivity = 4
	LayerEvaluation        = 5
	LayerDataCollection    = 6
	LayerReporting         = 7
)

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
const RoleProfileDir = "pacman/roles"

// RoleProfileExt is the file extension for saved role profiles.
const RoleProfileExt = ".yaml"

// BlockCacheDir is the subdirectory under the user's config
// directory where extracted content blocks are stored.
const BlockCacheDir = "pacman/blocks"

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
const TeamConfigDir = "pacman/teams"

// TeamConfigExt is the file extension for saved team configs.
const TeamConfigExt = ".yaml"

// LayerArtifacts maps Gemara layer numbers to the artifact
// types primarily produced at that layer. MappingDocument is
// cross-layer (L1-L3) and is listed under each layer it
// spans.
var LayerArtifacts = map[int][]string{
	LayerGuidance: {ArtifactGuidanceCatalog},
	LayerThreatsControls: {
		ArtifactThreatCatalog,
		ArtifactControlCatalog,
	},
	LayerRiskPolicy:        {ArtifactPolicy},
	LayerSensitiveActivity: {},
	LayerEvaluation:        {ArtifactEvaluationLog},
	LayerDataCollection:    {},
	LayerReporting:         {},
}

// ArtifactFlowDescriptions describes how artifacts flow
// between adjacent Gemara layers.
var ArtifactFlowDescriptions = map[[2]int]string{
	{LayerGuidance, LayerThreatsControls}: "Guidance " +
		"catalogs inform threat and control scope",
	{LayerGuidance, LayerRiskPolicy}: "Guidance " +
		"catalogs are referenced by policy documents",
	{LayerThreatsControls, LayerRiskPolicy}: "Control " +
		"and threat catalogs feed policy evaluation " +
		"criteria",
	{LayerThreatsControls, LayerSensitiveActivity}: "" +
		"Controls define requirements for sensitive " +
		"activities",
	{LayerRiskPolicy, LayerSensitiveActivity}: "Policy " +
		"governs which controls apply to sensitive " +
		"activities",
	{LayerRiskPolicy, LayerEvaluation}: "Policy drives " +
		"evaluation log assessments",
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
	SectionPolicyCriteria = "policy_criteria"
	SectionMappings       = "mappings"
	SectionEvaluations    = "evaluations"
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
	ArtifactPolicy: {
		SectionMetadata,
		SectionScope,
		SectionPolicyCriteria,
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
}
