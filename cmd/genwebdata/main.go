// SPDX-License-Identifier: Apache-2.0

// Command genwebdata generates a TypeScript module from the
// Go role definitions, layer keywords, artifact metadata,
// and MCP requirements. The output is written to the web/
// directory for use by the React frontend.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/roles"
)

// webRole is the JSON-serializable representation of a role.
type webRole struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	DefaultKeywords []string `json:"defaultKeywords"`
	DefaultLayers   []int    `json:"defaultLayers"`
}

// webActivityCategory is the JSON-serializable activity
// category.
type webActivityCategory struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Layers      []int    `json:"layers"`
	Keywords    []string `json:"keywords"`
}

// webLayer represents a Gemara layer for display.
type webLayer struct {
	Number      int      `json:"number"`
	Name        string   `json:"name"`
	Purpose     string   `json:"purpose"`
	ArtifactIDs []string `json:"artifactIds"`
}

// webArtifactType holds artifact metadata.
type webArtifactType struct {
	ID                string   `json:"id"`
	SchemaDef         string   `json:"schemaDef"`
	Description       string   `json:"description"`
	MCPWizard         string   `json:"mcpWizard"`
	AuthoringApproach string   `json:"authoringApproach"`
	Sections          []string `json:"sections"`
	Checklist         []string `json:"checklist"`
}

// webMCPRequirement represents a single MCP setup check.
type webMCPRequirement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	InstallCmd  string `json:"installCmd"`
	InstallURL  string `json:"installUrl"`
	Category    string `json:"category"`
}

// webMCPCapability describes an MCP server capability.
type webMCPCapability struct {
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// webLayerFlow describes a flow between layers.
type webLayerFlow struct {
	From        int    `json:"from"`
	To          int    `json:"to"`
	Description string `json:"description"`
}

// webUpstreamTutorial describes an upstream Gemara tutorial.
type webUpstreamTutorial struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	URL           string   `json:"url"`
	Layer         int      `json:"layer"`
	ArtifactTypes []string `json:"artifactTypes"`
	Prerequisites []string `json:"prerequisites"`
	Goals         []string `json:"goals"`
	Roles         []string `json:"roles"`
}

// webData is the top-level export structure.
type webData struct {
	Roles              []webRole             `json:"roles"`
	ActivityCategories []webActivityCategory `json:"activityCategories"`
	LayerKeywords      map[string][]int      `json:"layerKeywords"`
	Layers             []webLayer            `json:"layers"`
	ArtifactTypes      []webArtifactType     `json:"artifactTypes"`
	LayerFlows         []webLayerFlow        `json:"layerFlows"`
	UpstreamTutorials  []webUpstreamTutorial `json:"upstreamTutorials"`
	MCPRequirements    []webMCPRequirement   `json:"mcpRequirements"`
	MCPCapabilities    []webMCPCapability    `json:"mcpCapabilities"`
	MCPModes           []webMCPMode          `json:"mcpModes"`
	Config             webConfig             `json:"config"`
}

// webMCPMode describes an MCP server operating mode.
type webMCPMode struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Recommended bool   `json:"recommended"`
}

// webConfig holds configuration constants.
type webConfig struct {
	GemaraRepoURL      string `json:"gemaraRepoUrl"`
	GemaraMCPRepoURL   string `json:"gemaraMcpRepoUrl"`
	MCPBinaryName      string `json:"mcpBinaryName"`
	OpenCodeConfigFile string `json:"openCodeConfigFile"`
	DiscussionsURL     string `json:"discussionsUrl"`
	NewDiscussionURL   string `json:"newDiscussionUrl"`
}

func main() {
	outDir := "web/src/generated"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	data := buildWebData()

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	// Write JSON.
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(1)
	}

	jsonPath := filepath.Join(outDir, "journey-data.json")
	if err := os.WriteFile(
		jsonPath, jsonBytes, 0o644,
	); err != nil {
		fmt.Fprintf(os.Stderr, "write json: %v\n", err)
		os.Exit(1)
	}

	// Write TypeScript module.
	tsPath := filepath.Join(outDir, "journey-data.ts")
	tsContent := generateTypeScript(jsonBytes)
	if err := os.WriteFile(
		tsPath, []byte(tsContent), 0o644,
	); err != nil {
		fmt.Fprintf(os.Stderr, "write ts: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s\n", jsonPath)
	fmt.Printf("Generated %s\n", tsPath)
}

func buildWebData() *webData {
	return &webData{
		Roles:              buildRoles(),
		ActivityCategories: buildActivityCategories(),
		LayerKeywords:      buildLayerKeywords(),
		Layers:             buildLayers(),
		ArtifactTypes:      buildArtifactTypes(),
		LayerFlows:         buildLayerFlows(),
		UpstreamTutorials:  buildUpstreamTutorials(),
		MCPRequirements:    buildMCPRequirements(),
		MCPCapabilities:    buildMCPCapabilities(),
		MCPModes:           buildMCPModes(),
		Config:             buildConfig(),
	}
}

func buildRoles() []webRole {
	predefined := roles.PredefinedRoles()
	result := make([]webRole, len(predefined))
	for i, r := range predefined {
		result[i] = webRole{
			Name:            r.Name,
			Description:     r.Description,
			DefaultKeywords: r.DefaultKeywords,
			DefaultLayers:   r.DefaultLayers,
		}
	}
	return result
}

func buildActivityCategories() []webActivityCategory {
	cats := roles.ActivityCategories()
	result := make([]webActivityCategory, len(cats))
	for i, c := range cats {
		result[i] = webActivityCategory{
			Name:        c.Name,
			Description: c.Description,
			Layers:      c.Layers,
			Keywords:    c.Keywords,
		}
	}
	return result
}

func buildLayerKeywords() map[string][]int {
	result := make(map[string][]int)
	for kw, layers := range roles.LayerKeywords {
		result[kw] = layers
	}
	return result
}

// layerNames maps layer numbers to human-readable names.
var layerNames = map[int]string{
	consts.LayerVectorsGuidance:   "Vectors & Guidance",
	consts.LayerThreatsControls:   "Threats & Controls",
	consts.LayerRiskPolicy:        "Risk & Policy",
	consts.LayerSensitiveActivity: "Sensitive Activities",
	consts.LayerEvaluation:        "Intent & Behavior Evaluation",
	consts.LayerEnforcement:       "Preventive & Remediative Enforcement",
	consts.LayerAudit:             "Audit & Continuous Monitoring",
}

// layerPurposes maps layer numbers to descriptions.
var layerPurposes = map[int]string{
	consts.LayerVectorsGuidance: "Standards, best practices, " +
		"regulatory requirements, attack vectors, " +
		"secure design principles",
	consts.LayerThreatsControls: "Threat catalogs, " +
		"control catalogs, capability catalogs",
	consts.LayerRiskPolicy: "Organizational policy, " +
		"risk catalogs, assessment plans, adherence",
	consts.LayerSensitiveActivity: "Deployment pipelines, " +
		"CI/CD, operational activities",
	consts.LayerEvaluation: "Assessment logs, control " +
		"evaluations, evidence",
	consts.LayerEnforcement: "Corrective actions for " +
		"noncompliance",
	consts.LayerAudit: "Efficacy review of all " +
		"previous outputs",
}

func buildLayers() []webLayer {
	var result []webLayer
	for i := consts.LayerVectorsGuidance; i <= consts.LayerAudit; i++ {
		artifacts := consts.LayerArtifacts[i]
		if artifacts == nil {
			artifacts = []string{}
		}
		result = append(result, webLayer{
			Number:      i,
			Name:        layerNames[i],
			Purpose:     layerPurposes[i],
			ArtifactIDs: artifacts,
		})
	}
	return result
}

func buildArtifactTypes() []webArtifactType {
	allTypes := []string{
		consts.ArtifactGuidanceCatalog,
		consts.ArtifactVectorCatalog,
		consts.ArtifactPrincipleCatalog,
		consts.ArtifactControlCatalog,
		consts.ArtifactThreatCatalog,
		consts.ArtifactCapabilityCatalog,
		consts.ArtifactPolicy,
		consts.ArtifactRiskCatalog,
		consts.ArtifactMappingDocument,
		consts.ArtifactEvaluationLog,
		consts.ArtifactEnforcementLog,
		consts.ArtifactAuditLog,
	}

	schemaMap := map[string]string{
		consts.ArtifactGuidanceCatalog:   consts.SchemaGuidanceCatalog,
		consts.ArtifactVectorCatalog:     consts.SchemaVectorCatalog,
		consts.ArtifactPrincipleCatalog:  consts.SchemaPrincipleCatalog,
		consts.ArtifactControlCatalog:    consts.SchemaControlCatalog,
		consts.ArtifactThreatCatalog:     consts.SchemaThreatCatalog,
		consts.ArtifactCapabilityCatalog: consts.SchemaCapabilityCatalog,
		consts.ArtifactPolicy:            consts.SchemaPolicy,
		consts.ArtifactRiskCatalog:       consts.SchemaRiskCatalog,
		consts.ArtifactMappingDocument:   consts.SchemaMappingDocument,
		consts.ArtifactEvaluationLog:     consts.SchemaEvaluationLog,
		consts.ArtifactEnforcementLog:    consts.SchemaEnforcementLog,
		consts.ArtifactAuditLog:          consts.SchemaAuditLog,
	}

	result := make([]webArtifactType, len(allTypes))
	for i, t := range allTypes {
		wizard := consts.ArtifactWizards[t]
		approach := consts.ApproachCollaborative
		if wizard != "" {
			approach = consts.ApproachWizard
		}

		sections := consts.ArtifactTypeSections[t]
		if sections == nil {
			sections = []string{}
		}

		checklist := consts.DefaultPreparationChecklists[t]
		if checklist == nil {
			checklist = []string{}
		}

		result[i] = webArtifactType{
			ID:                t,
			SchemaDef:         schemaMap[t],
			Description:       consts.ArtifactDescriptions[t],
			MCPWizard:         wizard,
			AuthoringApproach: approach,
			Sections:          sections,
			Checklist:         checklist,
		}
	}
	return result
}

func buildLayerFlows() []webLayerFlow {
	var result []webLayerFlow

	// Sort keys for deterministic output.
	type flowKey = [2]int
	keys := make([]flowKey, 0,
		len(consts.ArtifactFlowDescriptions))
	for k := range consts.ArtifactFlowDescriptions {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})

	for _, k := range keys {
		result = append(result, webLayerFlow{
			From:        k[0],
			To:          k[1],
			Description: consts.ArtifactFlowDescriptions[k],
		})
	}
	return result
}

func buildUpstreamTutorials() []webUpstreamTutorial {
	result := make(
		[]webUpstreamTutorial,
		len(consts.UpstreamTutorials),
	)
	for i, t := range consts.UpstreamTutorials {
		artifactTypes := t.ArtifactTypes
		if artifactTypes == nil {
			artifactTypes = []string{}
		}
		prerequisites := t.Prerequisites
		if prerequisites == nil {
			prerequisites = []string{}
		}
		goals := t.Goals
		if goals == nil {
			goals = []string{}
		}
		roles := t.Roles
		if roles == nil {
			roles = []string{}
		}
		result[i] = webUpstreamTutorial{
			ID:            t.ID,
			Title:         t.Title,
			Description:   t.Description,
			URL:           t.URL,
			Layer:         t.Layer,
			ArtifactTypes: artifactTypes,
			Prerequisites: prerequisites,
			Goals:         goals,
			Roles:         roles,
		}
	}
	return result
}

func buildMCPRequirements() []webMCPRequirement {
	return []webMCPRequirement{
		{
			ID:   "opencode",
			Name: "OpenCode",
			Description: "AI-powered terminal IDE " +
				"that connects to the gemara-mcp " +
				"server.",
			Required:   true,
			InstallCmd: consts.BrewInstallOpenCode,
			Category:   "tools",
		},
		{
			ID:   "go",
			Name: "Go",
			Description: "Required to build " +
				"gemara-mcp from source.",
			Required:   true,
			InstallURL: "https://go.dev/dl/",
			Category:   "tools",
		},
		{
			ID:   "git",
			Name: "Git",
			Description: "Required to clone the " +
				"gemara-mcp repository.",
			Required:   true,
			InstallURL: "https://git-scm.com",
			Category:   "tools",
		},
		{
			ID:   "cue",
			Name: "CUE",
			Description: "Used for local schema " +
				"validation when the MCP server " +
				"is unavailable.",
			Required:   false,
			InstallCmd: consts.BrewInstallCUE,
			Category:   "tools",
		},
		{
			ID:   "gemara-mcp",
			Name: "Gemara MCP Server",
			Description: "The MCP server that " +
				"provides schema validation, " +
				"lexicon, and wizard prompts.",
			Required: true,
			InstallCmd: "git clone " +
				consts.GemaraMCPCloneHTTPS +
				" && cd gemara-mcp && " +
				"git checkout main && make build",
			Category: "server",
		},
		{
			ID:   "opencode-config",
			Name: "OpenCode Configuration",
			Description: "The opencode.json file " +
				"must contain a gemara-mcp entry " +
				"with the binary path and mode.",
			Required: true,
			Category: "config",
		},
	}
}

func buildMCPCapabilities() []webMCPCapability {
	return []webMCPCapability{
		{
			Category: "tool",
			Name:     consts.ToolValidateArtifact,
			Description: "Validate YAML content " +
				"against the Gemara CUE schema.",
		},
		{
			Category: "resource",
			Name:     consts.ResourceLexicon,
			Description: "Gemara term definitions " +
				"(34+ terms).",
		},
		{
			Category:    "resource",
			Name:        consts.ResourceSchemaDefinitions,
			Description: "CUE schema documentation.",
		},
		{
			Category: "prompt",
			Name:     consts.WizardThreatAssessment,
			Description: "Interactive Threat Catalog " +
				"creation wizard.",
		},
		{
			Category: "prompt",
			Name:     consts.WizardControlCatalog,
			Description: "Interactive Control Catalog " +
				"creation wizard.",
		},
	}
}

func buildMCPModes() []webMCPMode {
	return []webMCPMode{
		{
			ID:   consts.MCPModeArtifact,
			Name: "Artifact",
			Description: "Full capabilities: tools, " +
				"resources, and guided creation " +
				"wizard prompts.",
			Recommended: true,
		},
		{
			ID:   consts.MCPModeAdvisory,
			Name: "Advisory",
			Description: "Read-only analysis and " +
				"validation: tools and resources " +
				"only, no wizard prompts.",
			Recommended: false,
		},
	}
}

func buildConfig() webConfig {
	return webConfig{
		GemaraRepoURL:      consts.GemaraRepoURL,
		GemaraMCPRepoURL:   consts.GemaraMCPRepoURL,
		MCPBinaryName:      consts.MCPBinaryName,
		OpenCodeConfigFile: consts.OpenCodeConfigFile,
		DiscussionsURL:     consts.PacmanDiscussionsURL,
		NewDiscussionURL:   consts.PacmanNewDiscussionURL,
	}
}

func generateTypeScript(jsonData []byte) string {
	var sb strings.Builder

	sb.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	sb.WriteString("// AUTO-GENERATED by cmd/genwebdata — DO NOT EDIT\n")
	sb.WriteString("//\n")
	sb.WriteString("// Regenerate with: make web-data\n\n")

	sb.WriteString("export const journeyData = ")
	sb.Write(jsonData)
	sb.WriteString(" as const;\n\n")

	sb.WriteString("export type JourneyData = typeof journeyData;\n")
	sb.WriteString("export type Role = JourneyData['roles'][number];\n")
	sb.WriteString("export type ActivityCategory = JourneyData['activityCategories'][number];\n")
	sb.WriteString("export type Layer = JourneyData['layers'][number];\n")
	sb.WriteString("export type ArtifactType = JourneyData['artifactTypes'][number];\n")
	sb.WriteString("export type LayerFlow = JourneyData['layerFlows'][number];\n")
	sb.WriteString("export type MCPRequirement = JourneyData['mcpRequirements'][number];\n")
	sb.WriteString("export type MCPCapability = JourneyData['mcpCapabilities'][number];\n")
	sb.WriteString("export type MCPMode = JourneyData['mcpModes'][number];\n")
	sb.WriteString("export type UpstreamTutorial = JourneyData['upstreamTutorials'][number];\n")

	return sb.String()
}
