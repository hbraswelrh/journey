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
	// Layer 1 (Guidance) keywords.
	"eu cra":                   {consts.LayerGuidance},
	"nist":                     {consts.LayerGuidance},
	"owasp":                    {consts.LayerGuidance},
	"hipaa":                    {consts.LayerGuidance},
	"gdpr":                     {consts.LayerGuidance},
	"pci":                      {consts.LayerGuidance},
	"iso":                      {consts.LayerGuidance},
	"best practices":           {consts.LayerGuidance},
	"machine-readable format":  {consts.LayerGuidance},
	"standards":                {consts.LayerGuidance},
	"codify":                   {consts.LayerGuidance},
	"formalize best practices": {consts.LayerGuidance},
	"internal use-case":        {consts.LayerGuidance},
	"regulatory":               {consts.LayerGuidance},
	"guidance":                 {consts.LayerGuidance},

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

	// Ambiguous keywords (span Layers 1 and 3).
	"evidence collection": {
		consts.LayerGuidance,
		consts.LayerRiskPolicy,
	},
	"adherence": {
		consts.LayerGuidance,
		consts.LayerRiskPolicy,
	},

	// Layer 4 (Sensitive Activities) keywords.
	"pipeline security": {
		consts.LayerSensitiveActivity,
	},
	"deployment pipeline": {
		consts.LayerSensitiveActivity,
	},

	// Layer 5 (Evaluation) keywords.
	"evaluation":     {consts.LayerEvaluation},
	"audit":          {consts.LayerEvaluation},
	"assessment":     {consts.LayerEvaluation},
	"evaluation log": {consts.LayerEvaluation},
	"control evaluation": {
		consts.LayerEvaluation,
	},
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
