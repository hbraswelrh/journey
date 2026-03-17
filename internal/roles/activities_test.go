// SPDX-License-Identifier: Apache-2.0

package roles

import (
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// T203: ExtractKeywords extracts known domain terms from
// free-text description.
func TestExtractKeywordsFindsKnownTerms(t *testing.T) {
	t.Parallel()

	keywords := ExtractKeywords(
		"CI/CD pipeline management and " +
			"dependency management",
	)

	if len(keywords) == 0 {
		t.Fatal("expected keywords to be extracted")
	}

	found := make(map[string]bool)
	for _, kw := range keywords {
		found[kw] = true
	}

	if !found["ci/cd"] {
		t.Error("expected 'ci/cd' to be extracted")
	}
	if !found["dependency management"] {
		t.Error(
			"expected 'dependency management' to be " +
				"extracted",
		)
	}
}

// T204: ExtractKeywords returns empty slice for text with no
// recognizable domain keywords.
func TestExtractKeywordsNoMatch(t *testing.T) {
	t.Parallel()

	keywords := ExtractKeywords(
		"I like to go hiking on weekends",
	)

	if len(keywords) != 0 {
		t.Errorf(
			"expected no keywords, got %v", keywords,
		)
	}
}

// T205: KeywordMapping maps Layer 1 keywords correctly.
func TestKeywordMappingLayer1(t *testing.T) {
	t.Parallel()

	layer1Keywords := []string{
		"eu cra", "nist", "best practices",
		"machine-readable format",
	}

	for _, kw := range layer1Keywords {
		layers, ok := LayerKeywords[kw]
		if !ok {
			t.Errorf(
				"keyword %q not found in LayerKeywords",
				kw,
			)
			continue
		}
		if !containsLayer(layers, consts.LayerGuidance) {
			t.Errorf(
				"keyword %q should map to Layer 1 "+
					"(Guidance), got %v",
				kw, layers,
			)
		}
	}
}

// T206: KeywordMapping maps Layer 2 keywords correctly.
func TestKeywordMappingLayer2(t *testing.T) {
	t.Parallel()

	layer2Keywords := []string{
		"sdlc", "threat modeling", "ci/cd",
		"osps baseline", "finos ccc",
	}

	for _, kw := range layer2Keywords {
		layers, ok := LayerKeywords[kw]
		if !ok {
			t.Errorf(
				"keyword %q not found in LayerKeywords",
				kw,
			)
			continue
		}
		hasL2 := containsLayer(
			layers, consts.LayerThreatsControls,
		)
		if !hasL2 {
			t.Errorf(
				"keyword %q should map to Layer 2 "+
					"(Threats & Controls), got %v",
				kw, layers,
			)
		}
	}
}

// T207: KeywordMapping maps Layer 3 keywords correctly.
func TestKeywordMappingLayer3(t *testing.T) {
	t.Parallel()

	layer3Keywords := []string{
		"create policy", "timeline for adherence",
		"risk appetite", "audit interviews",
	}

	for _, kw := range layer3Keywords {
		layers, ok := LayerKeywords[kw]
		if !ok {
			t.Errorf(
				"keyword %q not found in LayerKeywords",
				kw,
			)
			continue
		}
		hasL3 := containsLayer(
			layers, consts.LayerRiskPolicy,
		)
		if !hasL3 {
			t.Errorf(
				"keyword %q should map to Layer 3 "+
					"(Risk & Policy), got %v",
				kw, layers,
			)
		}
	}
}

// T208: Ambiguous keywords spanning Layers 1 and 3 are
// identified by ClarificationNeeded.
func TestClarificationNeededAmbiguousKeywords(t *testing.T) {
	t.Parallel()

	keywords := []string{
		"evidence collection", "ci/cd", "adherence",
	}

	ambiguous := ClarificationNeeded(keywords)

	if len(ambiguous) == 0 {
		t.Fatal(
			"expected ambiguous keywords to be identified",
		)
	}

	found := make(map[string]bool)
	for _, kw := range ambiguous {
		found[kw] = true
	}

	if !found["evidence collection"] {
		t.Error(
			"expected 'evidence collection' to be " +
				"ambiguous",
		)
	}
	if !found["adherence"] {
		t.Error(
			"expected 'adherence' to be ambiguous",
		)
	}
	if found["ci/cd"] {
		t.Error(
			"'ci/cd' should NOT be ambiguous " +
				"(maps to Layer 2 only)",
		)
	}
}

// T226: Security Engineer + CI/CD activities resolves to
// Layer 2 emphasis.
func TestResolveLayerMappingsCICDFocus(t *testing.T) {
	t.Parallel()

	roles := PredefinedRoles()
	var secEng *Role
	for i := range roles {
		if roles[i].Name == consts.RoleSecurityEngineer {
			secEng = &roles[i]
			break
		}
	}

	keywords := ExtractKeywords(
		"CI/CD pipeline management, dependency " +
			"management, coding with upstream " +
			"open-source components",
	)

	profile := ResolveLayerMappings(
		secEng, keywords,
		"CI/CD pipeline management, dependency "+
			"management, coding with upstream "+
			"open-source components",
	)

	if profile == nil {
		t.Fatal("expected non-nil profile")
	}

	layers := profile.UniqueLayerNumbers()
	if !containsLayer(layers, consts.LayerThreatsControls) {
		t.Errorf(
			"expected Layer 2 in resolved layers, "+
				"got %v",
			layers,
		)
	}

	// Layer 2 should be the primary (strong confidence)
	// layer.
	for _, lm := range profile.ResolvedLayers {
		if lm.Layer == consts.LayerThreatsControls {
			if lm.Confidence != ConfidenceStrong {
				t.Errorf(
					"expected ConfidenceStrong for "+
						"Layer 2, got %d",
					lm.Confidence,
				)
			}
			break
		}
	}
}

// T227: Security Engineer + audit activities resolves to
// Layers 1 and 3 emphasis.
func TestResolveLayerMappingsAuditFocus(t *testing.T) {
	t.Parallel()

	roles := PredefinedRoles()
	var secEng *Role
	for i := range roles {
		if roles[i].Name == consts.RoleSecurityEngineer {
			secEng = &roles[i]
			break
		}
	}

	keywords := ExtractKeywords(
		"evidence collection, audit interviews, " +
			"defining compliance scope",
	)

	profile := ResolveLayerMappings(
		secEng, keywords,
		"evidence collection, audit interviews, "+
			"defining compliance scope",
	)

	if profile == nil {
		t.Fatal("expected non-nil profile")
	}

	layers := profile.UniqueLayerNumbers()
	hasL1 := containsLayer(layers, consts.LayerGuidance)
	hasL3 := containsLayer(layers, consts.LayerRiskPolicy)

	if !hasL1 || !hasL3 {
		t.Errorf(
			"expected Layers 1 and 3 in resolved "+
				"layers, got %v",
			layers,
		)
	}
}

// T228: Same role title with different activities produces
// different layer mappings.
func TestSameRoleDifferentActivities(t *testing.T) {
	t.Parallel()

	roles := PredefinedRoles()
	var secEng *Role
	for i := range roles {
		if roles[i].Name == consts.RoleSecurityEngineer {
			secEng = &roles[i]
			break
		}
	}

	cicdKW := ExtractKeywords(
		"CI/CD pipeline management and " +
			"dependency management",
	)
	cicdProfile := ResolveLayerMappings(
		secEng, cicdKW,
		"CI/CD pipeline management",
	)

	auditKW := ExtractKeywords(
		"evidence collection and audit interviews",
	)
	auditProfile := ResolveLayerMappings(
		secEng, auditKW,
		"evidence collection and audit interviews",
	)

	// The two profiles should have different primary
	// layers.
	cicdLayers := cicdProfile.UniqueLayerNumbers()
	auditLayers := auditProfile.UniqueLayerNumbers()

	if layersEqual(cicdLayers, auditLayers) {
		t.Errorf(
			"same role with different activities "+
				"should produce different layers: "+
				"CI/CD=%v, Audit=%v",
			cicdLayers, auditLayers,
		)
	}
}

// T229: "map my best practices to the EU CRA" routes to
// Layer 1.
func TestExtractKeywordsBestPracticesEUCRA(t *testing.T) {
	t.Parallel()

	keywords := ExtractKeywords(
		"map my best practices to the EU CRA",
	)

	found := make(map[string]bool)
	for _, kw := range keywords {
		found[kw] = true
	}

	if !found["best practices"] {
		t.Error(
			"expected 'best practices' to be extracted",
		)
	}
	if !found["eu cra"] {
		t.Error("expected 'eu cra' to be extracted")
	}

	// Verify these route to Layer 1.
	profile := ResolveLayerMappings(
		nil, keywords,
		"map my best practices to the EU CRA",
	)
	layers := profile.UniqueLayerNumbers()
	if !containsLayer(layers, consts.LayerGuidance) {
		t.Errorf(
			"expected Layer 1 (Guidance), got %v",
			layers,
		)
	}
}

// T230: "create a reusable machine-readable format for my
// internal standards" routes to Layer 1.
func TestExtractKeywordsMachineReadable(t *testing.T) {
	t.Parallel()

	keywords := ExtractKeywords(
		"create a reusable machine-readable format " +
			"for my internal standards",
	)

	found := make(map[string]bool)
	for _, kw := range keywords {
		found[kw] = true
	}

	if !found["machine-readable format"] {
		t.Error(
			"expected 'machine-readable format' to " +
				"be extracted",
		)
	}

	profile := ResolveLayerMappings(
		nil, keywords,
		"create a reusable machine-readable format",
	)
	layers := profile.UniqueLayerNumbers()
	if !containsLayer(layers, consts.LayerGuidance) {
		t.Errorf(
			"expected Layer 1 (Guidance), got %v",
			layers,
		)
	}
}

// T231: "create a policy and define a timeline for adherence"
// routes to Layer 3.
func TestExtractKeywordsCreatePolicy(t *testing.T) {
	t.Parallel()

	keywords := ExtractKeywords(
		"create a policy and define a timeline " +
			"for adherence",
	)

	found := make(map[string]bool)
	for _, kw := range keywords {
		found[kw] = true
	}

	if !found["create a policy"] {
		t.Error(
			"expected 'create a policy' to be extracted",
		)
	}
	if !found["timeline for adherence"] {
		t.Error(
			"expected 'timeline for adherence' to " +
				"be extracted",
		)
	}

	profile := ResolveLayerMappings(
		nil, keywords,
		"create a policy and define a timeline "+
			"for adherence",
	)
	layers := profile.UniqueLayerNumbers()
	if !containsLayer(layers, consts.LayerRiskPolicy) {
		t.Errorf(
			"expected Layer 3 (Risk & Policy), got %v",
			layers,
		)
	}
}

// T232: Ambiguous keyword "evidence collection" triggers
// clarifying follow-up between Layers 1 and 3.
func TestClarificationNeededEvidenceCollection(
	t *testing.T,
) {
	t.Parallel()

	keywords := ExtractKeywords(
		"I do evidence collection as part of my " +
			"daily work",
	)

	if len(keywords) == 0 {
		t.Fatal(
			"expected 'evidence collection' to be " +
				"extracted",
		)
	}

	ambiguous := ClarificationNeeded(keywords)
	found := false
	for _, kw := range ambiguous {
		if kw == "evidence collection" {
			found = true
			break
		}
	}
	if !found {
		t.Error(
			"expected 'evidence collection' to need " +
				"clarification",
		)
	}
}

// ExtractKeywords returns nil for empty input.
func TestExtractKeywordsEmpty(t *testing.T) {
	t.Parallel()

	keywords := ExtractKeywords("")
	if keywords != nil {
		t.Errorf("expected nil, got %v", keywords)
	}
}

// ResolveLayerMappings with nil role uses only keywords.
func TestResolveLayerMappingsNilRole(t *testing.T) {
	t.Parallel()

	keywords := ExtractKeywords("threat modeling and SDLC")
	profile := ResolveLayerMappings(
		nil, keywords, "threat modeling and SDLC",
	)

	if profile == nil {
		t.Fatal("expected non-nil profile")
	}

	layers := profile.UniqueLayerNumbers()
	if !containsLayer(layers, consts.LayerThreatsControls) {
		t.Errorf(
			"expected Layer 2 from keywords, got %v",
			layers,
		)
	}
}

// Helper: containsLayer checks if a layer is in the list.
func containsLayer(layers []int, target int) bool {
	for _, l := range layers {
		if l == target {
			return true
		}
	}
	return false
}

// Helper: layersEqual checks if two layer slices are identical.
func layersEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// T005: ArtifactRecommendations returns recommendations for
// strong L2 layers.
func TestArtifactRecommendations_StrongL2(t *testing.T) {
	t.Parallel()

	profile := &ActivityProfile{
		ResolvedLayers: []LayerMapping{
			{
				Layer:      consts.LayerThreatsControls,
				Confidence: ConfidenceStrong,
				Keywords:   []string{"threat modeling"},
			},
		},
	}

	recs := ArtifactRecommendations(profile)

	if len(recs) != 2 {
		t.Fatalf(
			"expected 2 recommendations for L2, "+
				"got %d", len(recs),
		)
	}

	types := make(map[string]bool)
	for _, r := range recs {
		types[r.ArtifactType] = true
	}

	if !types[consts.ArtifactThreatCatalog] {
		t.Error("expected ThreatCatalog recommendation")
	}
	if !types[consts.ArtifactControlCatalog] {
		t.Error("expected ControlCatalog recommendation")
	}

	// Verify descriptions are populated.
	for _, r := range recs {
		if r.Description == "" {
			t.Errorf(
				"expected description for %s",
				r.ArtifactType,
			)
		}
		if r.SchemaDef == "" {
			t.Errorf(
				"expected schema def for %s",
				r.ArtifactType,
			)
		}
	}
}

// T005: ArtifactRecommendations returns recommendations for
// inferred L1 layers.
func TestArtifactRecommendations_InferredL1(t *testing.T) {
	t.Parallel()

	profile := &ActivityProfile{
		ResolvedLayers: []LayerMapping{
			{
				Layer:      consts.LayerGuidance,
				Confidence: ConfidenceInferred,
				Keywords:   []string{"nist"},
			},
		},
	}

	recs := ArtifactRecommendations(profile)

	if len(recs) != 1 {
		t.Fatalf(
			"expected 1 recommendation for L1, "+
				"got %d", len(recs),
		)
	}

	if recs[0].ArtifactType != consts.ArtifactGuidanceCatalog {
		t.Errorf(
			"expected GuidanceCatalog, got %s",
			recs[0].ArtifactType,
		)
	}
	if recs[0].Confidence != ConfidenceInferred {
		t.Error("expected inferred confidence")
	}
	if recs[0].MCPWizard != "" {
		t.Errorf(
			"expected no wizard for GuidanceCatalog, "+
				"got %s", recs[0].MCPWizard,
		)
	}
	if recs[0].AuthoringApproach != consts.ApproachCollaborative {
		t.Errorf(
			"expected collaborative approach, got %s",
			recs[0].AuthoringApproach,
		)
	}
}

// T005: ArtifactRecommendations returns empty for empty layers.
func TestArtifactRecommendations_EmptyLayers(t *testing.T) {
	t.Parallel()

	profile := &ActivityProfile{
		ResolvedLayers: []LayerMapping{},
	}

	recs := ArtifactRecommendations(profile)

	if len(recs) != 0 {
		t.Fatalf(
			"expected 0 recommendations for empty "+
				"layers, got %d", len(recs),
		)
	}
}

// T005: ArtifactRecommendations returns empty for L4 (no
// artifacts defined).
func TestArtifactRecommendations_L4NoArtifacts(
	t *testing.T,
) {
	t.Parallel()

	profile := &ActivityProfile{
		ResolvedLayers: []LayerMapping{
			{
				Layer:      consts.LayerSensitiveActivity,
				Confidence: ConfidenceStrong,
				Keywords:   []string{"pipeline security"},
			},
		},
	}

	recs := ArtifactRecommendations(profile)

	if len(recs) != 0 {
		t.Fatalf(
			"expected 0 recommendations for L4, "+
				"got %d", len(recs),
		)
	}
}

// T005: ArtifactRecommendations deduplicates across layers,
// keeping the highest confidence.
func TestArtifactRecommendations_Deduplication(
	t *testing.T,
) {
	t.Parallel()

	// L2 has ThreatCatalog+ControlCatalog. If same artifact
	// appeared in two layers, only one should appear.
	profile := &ActivityProfile{
		ResolvedLayers: []LayerMapping{
			{
				Layer:      consts.LayerThreatsControls,
				Confidence: ConfidenceStrong,
				Keywords:   []string{"threat modeling"},
			},
			{
				Layer:      consts.LayerGuidance,
				Confidence: ConfidenceInferred,
				Keywords:   []string{"nist"},
			},
		},
	}

	recs := ArtifactRecommendations(profile)

	// L2 = ThreatCatalog + ControlCatalog, L1 =
	// GuidanceCatalog → 3 unique types.
	if len(recs) != 3 {
		t.Fatalf(
			"expected 3 unique recommendations, "+
				"got %d", len(recs),
		)
	}

	// Verify strong confidence comes first.
	if recs[0].Confidence != ConfidenceStrong {
		t.Error(
			"expected strong confidence first in " +
				"sorted order",
		)
	}
}

// T005: ArtifactRecommendations sets MCPWizard for types
// that have wizards.
func TestArtifactRecommendations_WizardMapping(
	t *testing.T,
) {
	t.Parallel()

	profile := &ActivityProfile{
		ResolvedLayers: []LayerMapping{
			{
				Layer:      consts.LayerThreatsControls,
				Confidence: ConfidenceStrong,
				Keywords:   []string{"threat modeling"},
			},
		},
	}

	recs := ArtifactRecommendations(profile)

	for _, r := range recs {
		switch r.ArtifactType {
		case consts.ArtifactThreatCatalog:
			if r.MCPWizard != consts.WizardThreatAssessment {
				t.Errorf(
					"expected threat_assessment wizard, "+
						"got %s", r.MCPWizard,
				)
			}
			if r.AuthoringApproach != consts.ApproachWizard {
				t.Errorf(
					"expected wizard approach, got %s",
					r.AuthoringApproach,
				)
			}
		case consts.ArtifactControlCatalog:
			if r.MCPWizard != consts.WizardControlCatalog {
				t.Errorf(
					"expected control_catalog wizard, "+
						"got %s", r.MCPWizard,
				)
			}
		}
	}
}
