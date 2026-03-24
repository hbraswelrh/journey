// SPDX-License-Identifier: Apache-2.0

package roles

import (
	"sort"
	"strings"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// LayerKeywords maps domain keywords to Gemara layers per
// FR-022. This mapping is extensible through configuration.
// When a keyword maps to multiple layers, the system asks a
// clarifying follow-up.
var LayerKeywords = map[string][]int{
	// Layer 1 (Vectors & Guidance) keywords.
	"eu cra":                   {consts.LayerVectorsGuidance},
	"nist":                     {consts.LayerVectorsGuidance},
	"owasp":                    {consts.LayerVectorsGuidance},
	"hipaa":                    {consts.LayerVectorsGuidance},
	"gdpr":                     {consts.LayerVectorsGuidance},
	"pci":                      {consts.LayerVectorsGuidance},
	"iso":                      {consts.LayerVectorsGuidance},
	"best practices":           {consts.LayerVectorsGuidance},
	"machine-readable format":  {consts.LayerVectorsGuidance},
	"standards":                {consts.LayerVectorsGuidance},
	"codify":                   {consts.LayerVectorsGuidance},
	"formalize best practices": {consts.LayerVectorsGuidance},
	"internal use-case":        {consts.LayerVectorsGuidance},
	"regulatory":               {consts.LayerVectorsGuidance},
	"guidance":                 {consts.LayerVectorsGuidance},
	"attack vectors":           {consts.LayerVectorsGuidance},
	"vectors":                  {consts.LayerVectorsGuidance},
	"mitre att&ck":             {consts.LayerVectorsGuidance},
	"secure design principles": {consts.LayerVectorsGuidance},
	"design principles":        {consts.LayerVectorsGuidance},
	"principles":               {consts.LayerVectorsGuidance},
	"vector catalog":           {consts.LayerVectorsGuidance},
	"principle catalog":        {consts.LayerVectorsGuidance},
	"guidance catalog":         {consts.LayerVectorsGuidance},

	// Layer 2 (Threats & Controls) keywords.
	"sdlc": {consts.LayerThreatsControls},
	"threat modeling": {
		consts.LayerThreatsControls,
	},
	"penetration testing": {
		consts.LayerThreatsControls,
	},
	"secure architecture review": {
		consts.LayerThreatsControls,
	},
	"ci/cd": {consts.LayerThreatsControls},
	"dependency management": {
		consts.LayerThreatsControls,
	},
	"upstream open-source": {
		consts.LayerThreatsControls,
	},
	"custom controls": {
		consts.LayerThreatsControls,
	},
	"osps baseline": {consts.LayerThreatsControls},
	"finos ccc":     {consts.LayerThreatsControls},
	"control catalog": {
		consts.LayerThreatsControls,
	},
	"threat assessment": {
		consts.LayerThreatsControls,
	},
	"capability catalog": {
		consts.LayerThreatsControls,
	},
	"system capabilities": {
		consts.LayerThreatsControls,
	},

	// Layer 3 (Risk & Policy) keywords.
	"create policy": {consts.LayerRiskPolicy},
	"create a policy": {
		consts.LayerRiskPolicy,
	},
	"timeline for adherence": {
		consts.LayerRiskPolicy,
	},
	"scope definition": {consts.LayerRiskPolicy},
	"audit interviews": {consts.LayerRiskPolicy},
	"assessment plans": {consts.LayerRiskPolicy},
	"adherence requirements": {
		consts.LayerRiskPolicy,
	},
	"risk appetite": {consts.LayerRiskPolicy},
	"non-compliance handling": {
		consts.LayerRiskPolicy,
	},
	"compliance scope": {consts.LayerRiskPolicy},
	"risk catalog":     {consts.LayerRiskPolicy},
	"risk categories":  {consts.LayerRiskPolicy},
	"risk severity":    {consts.LayerRiskPolicy},

	// Ambiguous keywords (span Layers 1 and 3).
	"evidence collection": {
		consts.LayerVectorsGuidance,
		consts.LayerRiskPolicy,
	},
	"adherence": {
		consts.LayerVectorsGuidance,
		consts.LayerRiskPolicy,
	},

	// Layer 4 (Sensitive Activities) keywords.
	"pipeline security": {
		consts.LayerSensitiveActivity,
	},
	"deployment pipeline": {
		consts.LayerSensitiveActivity,
	},

	// Layer 5 (Intent & Behavior Evaluation) keywords.
	"evaluation":     {consts.LayerEvaluation},
	"assessment":     {consts.LayerEvaluation},
	"evaluation log": {consts.LayerEvaluation},
	"control evaluation": {
		consts.LayerEvaluation,
	},
	"intent evaluation": {
		consts.LayerEvaluation,
	},
	"behavior evaluation": {
		consts.LayerEvaluation,
	},

	// Layer 6 (Preventive & Remediative Enforcement)
	// keywords.
	"enforcement": {consts.LayerEnforcement},
	"enforcement log": {
		consts.LayerEnforcement,
	},
	"preventive enforcement": {
		consts.LayerEnforcement,
	},
	"remediative enforcement": {
		consts.LayerEnforcement,
	},
	"admission controller": {
		consts.LayerEnforcement,
	},

	// Layer 7 (Audit & Continuous Monitoring) keywords.
	"audit":     {consts.LayerAudit},
	"audit log": {consts.LayerAudit},
	"continuous monitoring": {
		consts.LayerAudit,
	},
	"audit results": {consts.LayerAudit},
}

// Confidence represents how strongly a keyword matched.
type Confidence int

const (
	// ConfidenceInferred means the keyword was inferred
	// from context rather than explicitly stated.
	ConfidenceInferred Confidence = iota
	// ConfidenceStrong means the keyword was explicitly
	// present in the user's description.
	ConfidenceStrong
)

// LayerMapping records a single layer association with
// confidence.
type LayerMapping struct {
	Layer      int
	Confidence Confidence
	// Keywords are the matched keywords that produced
	// this mapping.
	Keywords []string
}

// ActivityProfile is the result of activity probing for a
// specific user session.
type ActivityProfile struct {
	// ExtractedKeywords are domain terms found in the
	// user's description.
	ExtractedKeywords []string
	// MatchedCategories are the activity categories that
	// matched the keywords.
	MatchedCategories []string
	// ResolvedLayers are the Gemara layers determined from
	// the combination of role defaults and extracted
	// keywords.
	ResolvedLayers []LayerMapping
	// UserDescription is the original free-text input.
	UserDescription string
	// Role is the identified role for this profile.
	Role *Role
	// Recommendations are artifact types recommended for
	// this user based on their resolved layers. Populated
	// by calling ArtifactRecommendations after layer
	// resolution.
	Recommendations []ArtifactRecommendation
}

// ArtifactRecommendation represents a recommended artifact
// type for the user based on their resolved Gemara layers.
type ArtifactRecommendation struct {
	// ArtifactType is the artifact type identifier
	// (e.g., "ThreatCatalog").
	ArtifactType string
	// SchemaDef is the CUE schema definition
	// (e.g., "#ThreatCatalog").
	SchemaDef string
	// Description is a one-sentence user-facing
	// description of the artifact type.
	Description string
	// Layer is the Gemara layer number (1-7).
	Layer int
	// Confidence is Inferred or Strong, from the layer
	// mapping that produced this recommendation.
	Confidence Confidence
	// MCPWizard is the MCP wizard prompt name, or empty
	// if no wizard is available for this type.
	MCPWizard string
	// AuthoringApproach is "wizard" or "collaborative",
	// derived from MCPWizard presence.
	AuthoringApproach string
}

// ExtractKeywords identifies domain keywords from a free-text
// description using the LayerKeywords vocabulary. Multi-word
// terms are matched first (longest match wins).
func ExtractKeywords(description string) []string {
	if description == "" {
		return nil
	}

	lower := strings.ToLower(description)

	// Sort keywords by length descending so longer
	// multi-word terms match first.
	sortedKeys := make([]string, 0, len(LayerKeywords))
	for k := range LayerKeywords {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return len(sortedKeys[i]) > len(sortedKeys[j])
	})

	var found []string
	matched := make(map[string]bool)

	for _, keyword := range sortedKeys {
		if strings.Contains(lower, keyword) &&
			!matched[keyword] {
			found = append(found, keyword)
			matched[keyword] = true
		}
	}

	return found
}

// ClarificationNeeded returns keywords from the extracted set
// that map to multiple Gemara layers and therefore require a
// clarifying follow-up question (per FR-022: Layers 1 and 3
// combined for evidence collection and adherence).
func ClarificationNeeded(keywords []string) []string {
	var ambiguous []string
	for _, kw := range keywords {
		lower := strings.ToLower(kw)
		if layers, ok := LayerKeywords[lower]; ok {
			if len(layers) > 1 {
				ambiguous = append(ambiguous, kw)
			}
		}
	}
	return ambiguous
}

// ResolveLayerMappings combines the role's default layer
// mappings with keyword-extracted layers to produce a unified
// ActivityProfile. Keywords provide strong confidence;
// role defaults provide inferred confidence.
func ResolveLayerMappings(
	role *Role,
	keywords []string,
	description string,
) *ActivityProfile {
	layerMap := make(map[int]*LayerMapping)

	// Add role defaults as inferred confidence.
	if role != nil {
		for _, layer := range role.DefaultLayers {
			layerMap[layer] = &LayerMapping{
				Layer:      layer,
				Confidence: ConfidenceInferred,
			}
		}
	}

	// Add keyword-resolved layers as strong confidence.
	for _, kw := range keywords {
		lower := strings.ToLower(kw)
		if layers, ok := LayerKeywords[lower]; ok {
			for _, layer := range layers {
				if existing, found := layerMap[layer]; found {
					existing.Confidence = ConfidenceStrong
					existing.Keywords = append(
						existing.Keywords, kw,
					)
				} else {
					layerMap[layer] = &LayerMapping{
						Layer:      layer,
						Confidence: ConfidenceStrong,
						Keywords:   []string{kw},
					}
				}
			}
		}
	}

	// Convert map to sorted slice.
	layers := make([]LayerMapping, 0, len(layerMap))
	for _, lm := range layerMap {
		layers = append(layers, *lm)
	}
	sort.Slice(layers, func(i, j int) bool {
		// Strong confidence first, then by layer number.
		if layers[i].Confidence != layers[j].Confidence {
			return layers[i].Confidence >
				layers[j].Confidence
		}
		return layers[i].Layer < layers[j].Layer
	})

	categories := resolveCategories(keywords)

	return &ActivityProfile{
		ExtractedKeywords: keywords,
		MatchedCategories: categories,
		ResolvedLayers:    layers,
		UserDescription:   description,
		Role:              role,
	}
}

// ActivityCategory represents a named grouping of related
// activities for manual selection.
type ActivityCategory struct {
	Name        string
	Description string
	Layers      []int
	Keywords    []string
}

// ActivityCategories returns the available categories for
// manual selection when keyword extraction yields no results.
func ActivityCategories() []ActivityCategory {
	return []ActivityCategory{
		{
			Name: "Regulatory Compliance",
			Description: "Mapping best practices to " +
				"regulatory frameworks, creating " +
				"machine-readable standards",
			Layers: []int{consts.LayerGuidance},
			Keywords: []string{
				"best practices", "regulatory",
				"standards", "guidance",
			},
		},
		{
			Name: "Threat & Control Authoring",
			Description: "Threat modeling, writing " +
				"controls, importing external catalogs",
			Layers: []int{consts.LayerThreatsControls},
			Keywords: []string{
				"threat modeling", "custom controls",
				"SDLC", "CI/CD",
			},
		},
		{
			Name: "Policy & Risk",
			Description: "Creating policies, defining " +
				"adherence timelines, scope definition",
			Layers: []int{consts.LayerRiskPolicy},
			Keywords: []string{
				"create policy",
				"timeline for adherence",
				"risk appetite",
			},
		},
		{
			Name: "Secure Development",
			Description: "CI/CD pipeline management, " +
				"dependency management, upstream " +
				"open-source integration",
			Layers: []int{
				consts.LayerThreatsControls,
				consts.LayerSensitiveActivity,
			},
			Keywords: []string{
				"CI/CD", "dependency management",
				"upstream open-source",
				"pipeline security",
			},
		},
		{
			Name: "Evaluation & Audit",
			Description: "Conducting assessments, " +
				"reviewing evidence, audit interviews",
			Layers: []int{consts.LayerEvaluation},
			Keywords: []string{
				"evaluation", "audit",
				"assessment", "evidence collection",
			},
		},
	}
}

// resolveCategories maps extracted keywords to activity
// category names.
func resolveCategories(keywords []string) []string {
	categories := ActivityCategories()
	matched := make(map[string]bool)
	var result []string

	for _, kw := range keywords {
		lower := strings.ToLower(kw)
		for _, cat := range categories {
			for _, catKW := range cat.Keywords {
				if strings.EqualFold(lower, catKW) &&
					!matched[cat.Name] {
					matched[cat.Name] = true
					result = append(result, cat.Name)
				}
			}
		}
	}

	return result
}

// UniqueLayerNumbers returns a deduplicated, sorted list of
// layer numbers from an ActivityProfile.
func (p *ActivityProfile) UniqueLayerNumbers() []int {
	seen := make(map[int]bool)
	var layers []int
	for _, lm := range p.ResolvedLayers {
		if !seen[lm.Layer] {
			seen[lm.Layer] = true
			layers = append(layers, lm.Layer)
		}
	}
	sort.Ints(layers)
	return layers
}

// AllKeywords returns all unique keywords from all resolved
// layers in the profile.
func (p *ActivityProfile) AllKeywords() []string {
	seen := make(map[string]bool)
	var all []string
	for _, lm := range p.ResolvedLayers {
		for _, kw := range lm.Keywords {
			if !seen[kw] {
				seen[kw] = true
				all = append(all, kw)
			}
		}
	}
	return all
}

// artifactSchemaMap maps artifact type identifiers to their
// CUE schema definition names. This duplicates the mapping
// in authoring/model.go to avoid a circular import.
var artifactSchemaMap = map[string]string{
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

// ArtifactRecommendations builds a list of recommended
// artifact types from an activity profile's resolved layers.
// Recommendations are ordered by confidence (Strong first),
// then by layer number. Duplicate artifact types across
// layers are deduplicated, keeping the highest confidence.
func ArtifactRecommendations(
	profile *ActivityProfile,
) []ArtifactRecommendation {
	if profile == nil || len(profile.ResolvedLayers) == 0 {
		return nil
	}

	// Collect recommendations, tracking seen types for
	// deduplication.
	seen := make(map[string]bool)
	var recs []ArtifactRecommendation

	// Process layers in their existing order (already
	// sorted by confidence then layer number from
	// ResolveLayerMappings).
	for _, lm := range profile.ResolvedLayers {
		artifacts := consts.LayerArtifacts[lm.Layer]
		for _, artType := range artifacts {
			if seen[artType] {
				continue
			}
			seen[artType] = true

			wizard := consts.ArtifactWizards[artType]
			approach := consts.ApproachCollaborative
			if wizard != "" {
				approach = consts.ApproachWizard
			}

			recs = append(recs, ArtifactRecommendation{
				ArtifactType:      artType,
				SchemaDef:         artifactSchemaMap[artType],
				Description:       consts.ArtifactDescriptions[artType],
				Layer:             lm.Layer,
				Confidence:        lm.Confidence,
				MCPWizard:         wizard,
				AuthoringApproach: approach,
			})
		}
	}

	return recs
}
